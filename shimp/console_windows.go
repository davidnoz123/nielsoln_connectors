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
)

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
	out, err := os.OpenFile("CONOUT$", os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	os.Stdout, os.Stderr = out, out
	return nil
}
