//go:build windows

// raise_windows.go -- bring a window to the front through user32, not a shell.
//
// The first version shelled out to PowerShell's AppActivate. Measured on
// 2 Oct 2026: 330ms per call, spawning an entire PowerShell runtime, on EVERY
// selection. It ran in a goroutine so it did not delay the push, but it is a
// third of a second of work for something user32 does in microseconds.
//
// This is a genuine platform API rather than different arguments to one
// command, which is what earns a build constraint here when `activate` and
// `openBrowser` do not need one.
package main

import (
	"syscall"
	"unsafe"
)

var (
	user32                  = syscall.NewLazyDLL("user32.dll")
	procEnumWindows         = user32.NewProc("EnumWindows")
	procGetWindowThreadPID  = user32.NewProc("GetWindowThreadProcessId")
	procIsWindowVisible     = user32.NewProc("IsWindowVisible")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
	procShowWindow          = user32.NewProc("ShowWindow")
	procAttachThreadInput   = user32.NewProc("AttachThreadInput")
	procGetForegroundWindow = user32.NewProc("GetForegroundWindow")
	procIsIconic            = user32.NewProc("IsIconic")
	procGetWindowTextW      = user32.NewProc("GetWindowTextW")
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	procGetCurrentThreadID  = kernel32.NewProc("GetCurrentThreadId")
)

const swRestore = 9

// windowTitled finds a visible top-level window with exactly this title.
//
// BY TITLE, not by pid, and that is the whole lesson here. Launching
// chrome.exe against an existing --user-data-dir makes the new process hand
// off to the running instance and EXIT, so the pid we recorded was already
// dead and raisePID found nothing. A Chrome `--app` window's title is exactly
// the page's <title> -- no " - Google Chrome" suffix -- which makes it both
// distinctive and stable.
func windowTitled(title string) uintptr {
	var found uintptr
	buf := make([]uint16, 512)
	cb := syscall.NewCallback(func(hwnd uintptr, _ uintptr) uintptr {
		if vis, _, _ := procIsWindowVisible.Call(hwnd); vis == 0 {
			return 1
		}
		n, _, _ := procGetWindowTextW.Call(hwnd,
			uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
		if n == 0 {
			return 1
		}
		if syscall.UTF16ToString(buf[:n]) == title {
			found = hwnd
			return 0
		}
		return 1
	})
	procEnumWindows.Call(cb, 0)
	return found
}

// windowOf finds a visible top-level window belonging to pid.
func windowOf(pid int) uintptr {
	var found uintptr
	cb := syscall.NewCallback(func(hwnd uintptr, _ uintptr) uintptr {
		var owner uint32
		procGetWindowThreadPID.Call(hwnd, uintptr(unsafe.Pointer(&owner)))
		if int(owner) != pid {
			return 1 // keep going
		}
		if vis, _, _ := procIsWindowVisible.Call(hwnd); vis == 0 {
			return 1
		}
		found = hwnd
		return 0 // stop
	})
	procEnumWindows.Call(cb, 0)
	return found
}

// raisePID brings the window of pid to the front. Reports whether it tried.
//
// SetForegroundWindow alone is refused when the caller is not already the
// foreground process -- Windows protects against focus theft. Attaching to the
// foreground thread's input queue for the duration is the documented way round
// it, and it is detached again immediately.
func raiseWindow(pid int, title string) bool {
	hwnd := uintptr(0)
	if title != "" {
		hwnd = windowTitled(title)
	}
	if hwnd == 0 && pid > 0 {
		hwnd = windowOf(pid)
	}
	if hwnd == 0 {
		return false
	}
	if iconic, _, _ := procIsIconic.Call(hwnd); iconic != 0 {
		procShowWindow.Call(hwnd, swRestore)
	}
	fg, _, _ := procGetForegroundWindow.Call()
	var fgPID uint32
	fgThread, _, _ := procGetWindowThreadPID.Call(fg, uintptr(unsafe.Pointer(&fgPID)))
	me, _, _ := procGetCurrentThreadID.Call()
	if fgThread != 0 && fgThread != me {
		procAttachThreadInput.Call(me, fgThread, 1)
		defer procAttachThreadInput.Call(me, fgThread, 0)
	}
	procSetForegroundWindow.Call(hwnd)
	return true
}
