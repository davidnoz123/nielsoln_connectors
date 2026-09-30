//go:build !windows

// proctree_other.go -- the same question everywhere that is not Windows.
//
// A process group is the Unix answer to what a Job Object does on Windows:
// give the shell its own group, then signal the group rather than the one
// process, so the shell's children go with it.
//
// Setpgid has to be asked for BEFORE the command starts, which is why
// prepareTree exists at all: superviseTree alone would be too late.
package main

import (
	"os/exec"
	"syscall"
)

// prepareTree asks for a new process group. Called before Start.
func prepareTree(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setpgid = true
}

// superviseTree returns the function that ends the command and its children.
// Negating the pid signals the whole group.
func superviseTree(cmd *exec.Cmd) func() {
	if cmd.Process == nil {
		return func() {}
	}
	pid := cmd.Process.Pid
	killed := false
	return func() {
		if killed {
			return
		}
		killed = true
		if err := syscall.Kill(-pid, syscall.SIGKILL); err != nil {
			// The group has gone but this one may not have. Worth trying, and
			// worth not pretending the group call succeeded.
			_ = cmd.Process.Kill()
		}
	}
}
