//go:build windows

// console_windows.go -- does this process have a console, and can it get one?
//
// WHY THIS EXISTS. s50 decided the console belongs to the keeper rather than
// to a capability, and the question put back was the sharper one: can a
// keeper start with NO console and allocate one only when it turns out to
// have something to show?
//
// ⚠️ The answer is yes, and it is `AllocConsole`. A process with no console
// can create one at any time, which makes a lazy console possible: no window
// for a click that has nothing to say, and a window the moment something
// does. That is strictly better than deciding in advance, and it is why s50's
// refusal of a descriptor field gets stronger rather than weaker.
//
// Raw kernel32 rather than anything else, for the reason the repo already
// records twice: `AppActivate` through PowerShell cost 330ms per call, and a
// dependency is the thing that stops a human auditing the trust root by
// reading it.
package main

import (
	"os"
	"syscall"
)

var (
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleWindow = kernel32.NewProc("GetConsoleWindow")
	procAllocConsole     = kernel32.NewProc("AllocConsole")
	// ⚠️ A fresh console inherits the system OEM codepage, not UTF-8, so
	// every symbol this program prints arrives mangled until this is called.
	// Measured: codepage 437 on the machine this was written for.
	procSetConsoleOutputCP = kernel32.NewProc("SetConsoleOutputCP")
	// ⚠️ How a GUI-subsystem binary speaks into the terminal it was
	// typed in, rather than a new window that flashes and vanishes.
	procAttachConsole = kernel32.NewProc("AttachConsole")

	user32                  = syscall.NewLazyDLL("user32.dll")
	procIsWindowVisible     = user32.NewProc("IsWindowVisible")
	procGetForegroundWindow = user32.NewProc("GetForegroundWindow")
)

// consoleState describes the console this process was GIVEN, if any.
//
// ⚠️ Measuring before deciding. s50 has to choose between a GUI-subsystem
// binary, which removes the window and lengthens the documented install line,
// and FreeConsole, which keeps the line and leaves a brief flash. Both assume
// the window is VISIBLE and that assumption was never checked: a URI
// activation is not a shell launch, so Windows may well create the console
// hidden, in which case there is nothing to fix and the row closes for free.
//
// Visible is the first question. Foreground is the sharper one, because a
// window stealing focus interrupts typing, which is worse than a flicker and
// is the category s10 exists for.
func consoleState() (present, visible, foreground bool) {
	h, _, _ := procGetConsoleWindow.Call()
	if h == 0 {
		return false, false, false
	}
	v, _, _ := procIsWindowVisible.Call(h)
	fg, _, _ := procGetForegroundWindow.Call()
	return true, v != 0, fg == h
}

// hasConsole reports whether this process is attached to one.
//
// GetConsoleWindow returns NULL when there is none, which is the only
// question being asked: a process launched from a GUI context may or may not
// have been given one, and nothing in the design should guess.
func hasConsole() bool {
	h, _, _ := procGetConsoleWindow.Call()
	return h != 0
}

// allocConsole creates a console for a process that has none, and rebinds the
// standard streams onto it.
//
// ⚠️ THE REBIND IS NOT OPTIONAL. AllocConsole gives the process a console,
// and it does NOT redirect the existing os.Stdout and os.Stderr, which were
// bound to whatever they pointed at before: usually nothing. So a window
// appears, the program writes to its old handles, and the window stays empty.
// That is a console which looks like a working one, which is the failure
// shape this repo is organised against. CONOUT$ is the console's own device
// and opening it is what makes the window show anything.
func allocConsole() error {
	if hasConsole() {
		return nil
	}
	if r, _, err := procAllocConsole.Call(); r == 0 {
		return err
	}
	// ⚠️ THE SAME UTF-8 CALL AS consent_windows.go, and it is needed in BOTH
	// because there are two places that allocate a console. Fixing only the
	// consent path would leave the refusal messages mangled, which are the
	// ones a reader most needs. Error ignored for the reason given there.
	procSetConsoleOutputCP.Call(uintptr(cpUTF8))
	out, err := os.OpenFile("CONOUT$", os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	os.Stdout, os.Stderr = out, out
	return nil
}

// cpUTF8 is CP_UTF8, the codepage identifier for UTF-8.
//
// ⚠️ Named rather than inlined, because 65001 on its own in a syscall reads
// like a handle or a flag. A console allocated without it shows this repo's
// ⚠️ and ⓘ markers as mojibake, measured on a codepage-437 console.
const cpUTF8 = 65001

// attachParentConsole is -1, ATTACH_PARENT_PROCESS.
//
// ⚠️ Named because -1 passed to a syscall reads like an error return.
const attachParentConsole = ^uintptr(0) // (DWORD)-1

// speakToTerminal makes a GUI-subsystem process able to print.
//
// # ⚠️ THE DEFECT THIS FIXES, AND IT WAS FOUND THE HARD WAY
//
// The handler is installed with -ldflags "-H=windowsgui" so that a CLICK does
// not raise a console window. That is right for clicks and WRONG for the
// verbs a human types: `install`, `verify`, `uninstall`. A GUI-subsystem
// process has no stdout, so those verbs were MUTE.
//
// Measured 20 Oct 2026: an install that refused -- correctly, because
// replacing the trust root needs -replace -- printed a careful message naming
// both versions and the exact flag to pass, and the user saw only:
//
//	exit status 1
//
// ⚠️ The message was good. It was invisible. And `install` is the one command
// in this program always run from a terminal, so it is the worst possible
// place to have no output.
//
// # WHY ATTACH BEFORE ALLOCATE
//
// AllocConsole would also make the process able to print, but into a NEW
// WINDOW that appears and vanishes -- the flash problem this repo keeps
// fixing. AttachConsole(ATTACH_PARENT_PROCESS) instead borrows the terminal
// the user actually typed in, so output lands where they are looking.
//
// ⓘ Falls back to AllocConsole for the no-parent-console case (launched from
// Explorer, say), because printing into a new window beats not printing.
// Returns whether a NEW console was created, so a caller can decide to hold
// it open rather than let the window vanish with the process.
func speakToTerminal() (allocated bool) {
	if hasConsole() {
		return false // console-subsystem build, or already attached
	}
	if r, _, _ := procAttachConsole.Call(attachParentConsole); r != 0 {
		bindStdio()
		return false
	}
	if err := allocConsole(); err != nil {
		return false
	}
	return true
}

// bindStdio points os.Stdout and os.Stderr at the console device.
//
// ⚠️ NOT OPTIONAL AFTER ATTACHING, for the same reason allocConsole says it is
// not optional after allocating: attaching gives the process a console and
// leaves the existing handles bound to whatever they pointed at before, which
// for a GUI-subsystem binary is nothing. The result would be a program that
// looks like it printed and did not.
func bindStdio() {
	procSetConsoleOutputCP.Call(uintptr(cpUTF8))

	// ⚠️ PREFER THE INHERITED HANDLE OVER CONOUT$, AND THE ORDER MATTERS.
	//
	// CONOUT$ is the console DEVICE. Writing there reaches the console window
	// and BYPASSES any redirection, so `shimp verify > log.txt` would leave
	// an empty file while the text scrolled past in the terminal. Measured
	// while testing this very fix: piping the output to `head` showed
	// nothing, because there was nothing in the pipe.
	//
	// A GUI-subsystem process gets no standard handles of its own, but cmd.exe
	// DOES pass them when the caller redirects. So: if a real handle was
	// inherited, use it -- that makes `>` and `|` work. Only when there is
	// none (a bare console invocation) fall back to the device.
	if h, _, _ := procGetStdHandle.Call(uintptr(stdOutputHandle)); h != 0 &&
		h != invalidHandle {
		f := os.NewFile(h, "stdout")
		os.Stdout, os.Stderr = f, f
		return
	}
	if out, err := os.OpenFile("CONOUT$", os.O_WRONLY, 0); err == nil {
		os.Stdout, os.Stderr = out, out
	}
}

// stdOutputHandle is STD_OUTPUT_HANDLE, invalidHandle is INVALID_HANDLE_VALUE.
//
// ⓘ Both named, because -11 and -1 in a syscall are indistinguishable from
// mistakes at a glance.
const (
	stdOutputHandle = ^uintptr(10) // (DWORD)-11
	invalidHandle   = ^uintptr(0)  // (HANDLE)-1
)
