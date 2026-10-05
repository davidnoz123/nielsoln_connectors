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
	"os"
	"strconv"
	"strings"
	"sync"
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
	out := raiseWindow(hwnd)
	if strings.HasPrefix(out, "raised") {
		rememberWindow(hwnd)
	}
	return out
}

// The window we last raised, so the next request need not earn the right to
// know which one it is.
//
// WHY A CACHE AT ALL. Raising needs a window handle, and the only way this
// program had of finding one was to match a TAB TITLE, which does not exist
// until the page has loaded. So focus waited on a navigation it has nothing
// to do with: on a drive that opened its own tab, Excel's
// AllowSetForegroundWindow grant was already seconds stale by the time we
// spent it and Windows refused. Every "focus: Windows refused to raise it"
// on an own-tab drive is that, and the complaint was simply that the window
// should come up first.
//
// A handle is stable for the life of a window, so the FIRST raise of a
// session learns it the slow way and every one after is immediate.
// PERSISTED, because this server is restarted constantly and a handle
// outlives it. Without the file, the first right-click after every restart
// paid the slow path, and that is the click most likely to be the one
// somebody is watching. The drive Chrome has been up since the day before
// this was written, so the number in the file is usually still good.
//
// Safe to be wrong. A handle read from disk is checked for class and
// visibility like any other, and a stale one just falls back to the title
// route, so the worst case is the behaviour we had before the file existed.
var (
	lastRaised   uintptr
	lastRaisedMu sync.Mutex
	handleFile   string
)

// LoadWindowHandle reads the remembered handle, if there is one to read.
func LoadWindowHandle(path string) string {
	handleFile = path
	if path == "" {
		return "not remembered between runs"
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return "no handle remembered yet"
	}
	n, err := strconv.ParseUint(strings.TrimSpace(string(raw)), 10, 64)
	if err != nil || n == 0 {
		return "the remembered handle is unreadable"
	}
	hwnd := uintptr(n)
	if vis, _, _ := procIsWindowVisible.Call(hwnd); vis == 0 {
		return "the remembered window is gone"
	}
	if windowClass(hwnd) != chromeWndClass {
		return "the remembered handle is not a Chrome window now"
	}
	lastRaised = hwnd
	return fmt.Sprintf("remembered window %d, %q", hwnd,
		trim(windowText(hwnd), 40))
}

func rememberWindow(hwnd uintptr) {
	lastRaisedMu.Lock()
	defer lastRaisedMu.Unlock()
	if lastRaised == hwnd {
		return
	}
	lastRaised = hwnd
	if handleFile == "" || hwnd == 0 {
		return
	}
	if err := os.WriteFile(handleFile,
		[]byte(strconv.FormatUint(uint64(hwnd), 10)), 0o644); err != nil {
		// Said, not fatal: losing the file costs a slow first raise next
		// time and nothing else.
		fmt.Printf("  could not remember the window handle: %v\n", err)
	}
}

// RaiseRemembered raises the window we last raised, if it is still a visible
// Chrome window. Returns "" when there is nothing usable, so the caller can
// fall back to the title route rather than report a failure.
//
// VALIDATED BEFORE USE, because a handle outlives nothing: close the window
// and the number is stale or, worse, belongs to something else. Class and
// visibility are two cheap questions, and raising the wrong window is worse
// than raising none because it looks like the tool working while showing the
// human something else.
//
// ⚠️ ONE WINDOW. With two Chrome windows open on the driven profile this can
// raise the one the tab is NOT in. The late raise by title still runs
// whenever this did not succeed, so the failure mode is a slower correct
// raise rather than a wrong one; a second window would want the handle keyed
// by target.
func RaiseRemembered() string {
	lastRaisedMu.Lock()
	hwnd := lastRaised
	lastRaisedMu.Unlock()
	if hwnd == 0 {
		return ""
	}
	if vis, _, _ := procIsWindowVisible.Call(hwnd); vis == 0 {
		rememberWindow(0)
		return ""
	}
	if windowClass(hwnd) != chromeWndClass {
		rememberWindow(0)
		return ""
	}
	return raiseWindow(hwnd)
}
