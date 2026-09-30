//go:build windows

// proctree_windows.go -- stopping a command means stopping what it started.
//
// Found by a test rather than by reasoning: a 1.5 second timeout on
// `ping -n 30` took 29 seconds. Two separate reasons, and both are the same
// mistake, which is treating a command as one process.
//
// cmd.Process.Kill() kills the SHELL. The shell's own child, the thing that is
// actually doing the work, carries on. And cmd.Wait() does not return when the
// shell dies, because it waits for the output pipes to close, and the
// grandchild still holds the write end of them. So a timeout killed the wrong
// process and then blocked waiting for the right one.
//
// A Job Object is the Windows answer. Every process in the job dies together,
// including ones started later, so `npm install` spawning node spawning a
// hundred things is still one thing to stop.
//
// There is a race: the child could spawn before it is assigned to the job.
// Closing it properly needs CREATE_SUSPENDED and a manual ResumeThread, which
// os/exec does not expose. The window is microseconds between Start and
// Assign, and the consequence is an orphan rather than an escape, so it is
// documented rather than fixed.
package main

import (
	"os/exec"
	"syscall"
	"unsafe"
)

var (
	procCreateJobObject          = kernel32DLL.NewProc("CreateJobObjectW")
	procAssignProcessToJobObject = kernel32DLL.NewProc("AssignProcessToJobObject")
	procTerminateJobObject       = kernel32DLL.NewProc("TerminateJobObject")
	procSetInformationJobObject  = kernel32DLL.NewProc("SetInformationJobObject")
	procOpenProcess              = kernel32DLL.NewProc("OpenProcess")
)

const (
	jobObjectExtendedLimitInformation = 9
	jobObjectLimitKillOnJobClose      = 0x00002000

	processTerminate = 0x0001
	processSetQuota  = 0x0100
)

// The tail of JOBOBJECT_EXTENDED_LIMIT_INFORMATION that matters. The leading
// fields are laid out exactly as Windows expects; only LimitFlags is set.
type jobObjectBasicLimitInformation struct {
	PerProcessUserTimeLimit int64
	PerJobUserTimeLimit     int64
	LimitFlags              uint32
	MinimumWorkingSetSize   uintptr
	MaximumWorkingSetSize   uintptr
	ActiveProcessLimit      uint32
	Affinity                uintptr
	PriorityClass           uint32
	SchedulingClass         uint32
}

type ioCounters struct {
	ReadOperationCount  uint64
	WriteOperationCount uint64
	OtherOperationCount uint64
	ReadTransferCount   uint64
	WriteTransferCount  uint64
	OtherTransferCount  uint64
}

type jobObjectExtendedLimitInformationStruct struct {
	BasicLimitInformation jobObjectBasicLimitInformation
	IoInfo                ioCounters
	ProcessMemoryLimit    uintptr
	JobMemoryLimit        uintptr
	PeakProcessMemoryUsed uintptr
	PeakJobMemoryUsed     uintptr
}

// prepareTree is where Unix asks for a new process group. Windows needs
// nothing before the start: the job is made and assigned afterwards.
func prepareTree(cmd *exec.Cmd) {}

// superviseTree puts a started command and everything it goes on to start into
// one job, and returns the function that ends all of it.
//
// The returned func is safe to call more than once and safe to call on a
// command that has already finished, because both happen: a timeout and a kill
// can race, and a background process can end while the session is asking.
func superviseTree(cmd *exec.Cmd) func() {
	if cmd.Process == nil {
		return func() {}
	}
	job, _, _ := procCreateJobObject.Call(0, 0)
	if job == 0 {
		// No job: fall back to killing the one process we know about. Worse,
		// and honest about being worse, rather than silently doing nothing.
		return func() { _ = cmd.Process.Kill() }
	}

	// Kill everything in the job when the last handle to it closes, so a
	// connector that dies does not leave a build running on somebody's laptop.
	var info jobObjectExtendedLimitInformationStruct
	info.BasicLimitInformation.LimitFlags = jobObjectLimitKillOnJobClose
	procSetInformationJobObject.Call(job, jobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&info)), unsafe.Sizeof(info))

	handle, _, _ := procOpenProcess.Call(processTerminate|processSetQuota, 0,
		uintptr(cmd.Process.Pid))
	if handle == 0 {
		syscall.CloseHandle(syscall.Handle(job))
		return func() { _ = cmd.Process.Kill() }
	}
	procAssignProcessToJobObject.Call(job, handle)
	syscall.CloseHandle(syscall.Handle(handle))

	killed := false
	return func() {
		if killed {
			return
		}
		killed = true
		procTerminateJobObject.Call(job, 1)
		syscall.CloseHandle(syscall.Handle(job))
	}
}
