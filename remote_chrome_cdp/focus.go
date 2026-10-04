// focus.go -- raise the Chrome window we are driving.
//
// # WHY THIS CAN WORK HERE AT ALL
//
// Windows normally refuses SetForegroundWindow from a process the user is
// not interacting with: it flashes the taskbar button instead. Measured
// 2 Oct 2026 -- the same call succeeded from a script when nobody had
// touched the keyboard and did nothing after a real right-click.
//
// The exception is AllowSetForegroundWindow. The process that HOLDS the
// foreground may hand that right to another, and Excel holds it at the
// instant of a right-click. The VBA calls AllowSetForegroundWindow(ASFW_ANY)
// immediately before its request, so by the time we are asked, the right has
// already been granted to us. We were simply not using it.
//
// # WHY NOT AppActivate IN VBA
//
// That was the first attempt, copied from document_reviewer. It works, but
// it matches windows by title prefix, which means the server has to hand
// the title back, the VBA has to match on it, and a page with an unexpected
// title quietly raises nothing. This program already knows exactly which
// window it is driving. Matching a string to find something we have a
// handle to is a worse answer than keeping the handle.
//
// # CDP CANNOT DO IT
//
// Page.bringToFront makes a tab the active one INSIDE Chrome, which is
// necessary and not sufficient: it does not raise the window. Both steps are
// needed and they are different mechanisms.
package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	user32              = syscall.NewLazyDLL("user32.dll")
	procEnumWindows     = user32.NewProc("EnumWindows")
	procGetWindowTextW  = user32.NewProc("GetWindowTextW")
	procGetClassNameW   = user32.NewProc("GetClassNameW")
	procIsWindowVisible = user32.NewProc("IsWindowVisible")
	procSetForeground   = user32.NewProc("SetForegroundWindow")
	procShowWindow      = user32.NewProc("ShowWindow")
	procSetWindowPos    = user32.NewProc("SetWindowPos")
)

const (
	swRestore      = 9
	hwndTopmost    = ^uintptr(0) // (HWND)-1
	hwndNoTopmost  = ^uintptr(1) // (HWND)-2
	swpNoMove      = 0x0002
	swpNoSize      = 0x0001
	swpShowWindow  = 0x0040
	chromeWndClass = "Chrome_WidgetWin_1"
)

func windowText(hwnd uintptr) string {
	buf := make([]uint16, 512)
	n, _, _ := procGetWindowTextW.Call(hwnd,
		uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return syscall.UTF16ToString(buf[:n])
}

func windowClass(hwnd uintptr) string {
	buf := make([]uint16, 256)
	n, _, _ := procGetClassNameW.Call(hwnd,
		uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	return syscall.UTF16ToString(buf[:n])
}

// findChromeWindow returns the visible Chrome window whose title starts with
// `title`, or 0.
//
// The CLASS is checked as well as the title. A title alone would match a
// Word document or an editor that happens to have the page's name in its
// caption, and raising the wrong window is worse than raising none: it
// looks like the tool working and shows the human something else.
func findChromeWindow(title string) uintptr {
	var found uintptr
	cb := syscall.NewCallback(func(hwnd uintptr, _ uintptr) uintptr {
		if found != 0 {
			return 0
		}
		if vis, _, _ := procIsWindowVisible.Call(hwnd); vis == 0 {
			return 1
		}
		if windowClass(hwnd) != chromeWndClass {
			return 1
		}
		text := windowText(hwnd)
		if text == "" {
			return 1
		}
		if len(title) > 0 && len(text) >= len(title) && text[:len(title)] == title {
			found = hwnd
			return 0
		}
		return 1
	})
	procEnumWindows.Call(cb, 0)
	return found
}

// raiseWindow brings hwnd to the front. Returns what happened, for the log.
//
// Two steps, because a minimised window cannot be foregrounded: restore it
// first, then ask. The topmost nudge is the documented fallback for when
// SetForegroundWindow is refused despite the grant, and it is undone at once
// so the window does not stay pinned above everything the human owns.
func raiseWindow(hwnd uintptr) string {
	procShowWindow.Call(hwnd, swRestore)
	if ok, _, _ := procSetForeground.Call(hwnd); ok != 0 {
		return "raised"
	}
	procSetWindowPos.Call(hwnd, hwndTopmost, 0, 0, 0, 0,
		swpNoMove|swpNoSize|swpShowWindow)
	procSetWindowPos.Call(hwnd, hwndNoTopmost, 0, 0, 0, 0,
		swpNoMove|swpNoSize|swpShowWindow)
	if ok, _, _ := procSetForeground.Call(hwnd); ok != 0 {
		return "raised after a topmost nudge"
	}
	return "Windows refused to raise it; the taskbar button will be flashing"
}

// RaiseChromeTitled finds and raises the Chrome window showing `title`.
func RaiseChromeTitled(title string) string {
	if title == "" {
		return "no title to look for"
	}
	hwnd := findChromeWindow(title)
	if hwnd == 0 {
		return fmt.Sprintf("no visible Chrome window titled %q", trim(title, 50))
	}
	return raiseWindow(hwnd)
}
