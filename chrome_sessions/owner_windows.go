//go:build windows

// owner_windows.go -- which process holds a listening port, through iphlpapi.
//
// WHY THIS IS WORTH A PLATFORM FILE. An answering port with an unexpected
// owner is a different problem from a dead port, and only the owner tells them
// apart. A CDP port answering from something that is not the browser you
// think it is would otherwise look exactly like success, which is the failure
// shape this whole repo is organised against.
//
// Raw syscall rather than shelling out to netstat, for the reason
// pdf_view_server/raise_windows.go records about PowerShell's AppActivate:
// 330ms per call, spawning a whole runtime, for something the platform API
// does in microseconds. And it keeps the repo's zero third-party dependencies.
package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	iphlpapi                = syscall.NewLazyDLL("iphlpapi.dll")
	procGetExtendedTcpTable = iphlpapi.NewProc("GetExtendedTcpTable")
)

const (
	// AF_INET, because a CDP port is bound to 127.0.0.1. A browser listening
	// only on ::1 would be missed, and that is not a configuration anything
	// here creates.
	afINET = 2
	// TCP_TABLE_OWNER_PID_LISTENER: listeners with their owning pid, which is
	// exactly the question and nothing more.
	tableOwnerPIDListener = 3
	// ERROR_INSUFFICIENT_BUFFER, the expected answer to the sizing call.
	errInsufficientBuffer = 122
)

// mibTCPRowOwnerPID mirrors MIB_TCPROW_OWNER_PID: six DWORDs, in this order.
//
// ⚠️ THE PORTS ARE NETWORK BYTE ORDER IN THE LOW 16 BITS. Reading dwLocalPort
// as a plain uint32 gives 30987 for 9222, which looks like a real port number
// and is wrong, so it would not announce itself as a bug.
type mibTCPRowOwnerPID struct {
	state      uint32
	localAddr  uint32
	localPort  uint32
	remoteAddr uint32
	remotePort uint32
	owningPID  uint32
}

func (r mibTCPRowOwnerPID) port() int {
	b := (*[4]byte)(unsafe.Pointer(&r.localPort))
	return int(b[0])<<8 | int(b[1])
}

// listeners reads the table once and returns {port: pid}.
//
// Sized first, then read: the table changes between the two calls on a busy
// machine, so a buffer that was big enough a moment ago can be too small, and
// the retry is the documented way round that rather than a guess.
func listeners() (map[int]int, error) {
	var size uint32
	ret, _, _ := procGetExtendedTcpTable.Call(0, uintptr(unsafe.Pointer(&size)),
		0, afINET, tableOwnerPIDListener, 0)
	if ret != errInsufficientBuffer && ret != 0 {
		return nil, fmt.Errorf("GetExtendedTcpTable sizing failed: %d", ret)
	}
	for attempt := 0; attempt < 4; attempt++ {
		buf := make([]byte, size)
		ret, _, _ = procGetExtendedTcpTable.Call(
			uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)),
			0, afINET, tableOwnerPIDListener, 0)
		if ret == errInsufficientBuffer {
			continue // the table grew between the two calls; size is updated
		}
		if ret != 0 {
			return nil, fmt.Errorf("GetExtendedTcpTable failed: %d", ret)
		}
		n := *(*uint32)(unsafe.Pointer(&buf[0]))
		out := make(map[int]int, n)
		// The rows follow the count. A 4-byte count then 24-byte rows, and on
		// amd64 the struct needs no padding because every field is a DWORD.
		const head = 4
		const row = int(unsafe.Sizeof(mibTCPRowOwnerPID{}))
		for i := 0; i < int(n); i++ {
			at := head + i*row
			if at+row > len(buf) {
				break
			}
			r := *(*mibTCPRowOwnerPID)(unsafe.Pointer(&buf[at]))
			out[r.port()] = int(r.owningPID)
		}
		return out, nil
	}
	return nil, fmt.Errorf("the TCP table kept growing between calls")
}

// ownerOfPort is 0 when nothing holds the port or the table cannot be read.
// Never raises: this is a diagnostic field on a report, and a report that
// fails because an aside failed is worse than a report with a blank in it.
func ownerOfPort(port int) int {
	got, err := listeners()
	if err != nil {
		return 0
	}
	return got[port]
}

// discoverPorts lists every listening port, for -discover.
//
// ⚠️ A DIAGNOSTIC, NOT A WAY TO CHOOSE A PORT. Ports are declared, because a
// capability's identity must be computable before its process runs. This
// exists for the other direction: a browser is up on a port somebody forgot,
// and finding it beats guessing.
func discoverPorts() ([]int, string) {
	got, err := listeners()
	if err != nil {
		return nil, err.Error()
	}
	var out []int
	for port := range got {
		out = append(out, port)
	}
	return out, ""
}
