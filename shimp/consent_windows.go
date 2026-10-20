//go:build windows

// consent_windows.go -- a keypress that is actually the USER'S keypress.
//
// ⚠️ WHY THIS FILE REPLACED A SIMPLER ONE, AND IT IS THE WORST DEFECT THIS
// PROGRAM HAS HAD.
//
// The first consent gate called allocConsole, then polled STD_INPUT_HANDLE
// for a key event. Tested on 15 Oct 2026 with NO key pressed, it logged
// "consent given by keypress" and executed the program, IN THE SAME SECOND
// as the dispatch. It then launched a Chrome nobody had agreed to.
//
// Two faults, and the first is structural:
//
//  1. ⚠️ allocConsole RETURNS EARLY IF A CONSOLE ALREADY EXISTS. A handler
//     launched by ShellExecuteW from a shell INHERITS that shell's console,
//     input buffer and all. So consent was read from a buffer holding the
//     Enter key that had started the command. The user's answer was never
//     involved.
//
//  2. STD_INPUT_HANDLE follows whatever the process inherited, so even on a
//     fresh console it is not reliably the console's own input.
//
// The fix is not a tweak to the polling. A gate that authorises code
// execution must read from input it OWNS:
//
//	FreeConsole   detach from anything inherited
//	AllocConsole  create one that is ours, with an empty input buffer
//	CONIN$        open the console's own input directly, never a std handle
//	settle+flush  discard the arrival burst a new console generates
//	minSettle     and ignore any key arriving implausibly fast
//
// ⚠️ ANY FAILURE MEANS NO CONSENT MEANS NO RUN. Failing closed is the only
// safe direction: a handler that executed because it could not ask would be
// worse than one that does nothing.
//
// THE ASYMMETRY WITH refusal_windows.go IS DELIBERATE. That one still uses
// the inherited console, because a spurious keypress there merely closes a
// window that has already done its job. Here a spurious keypress runs a
// program. Same mechanism, opposite consequence, so opposite strictness.
package main

import (
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"
)

var (
	procFreeConsole        = kernel32.NewProc("FreeConsole")
	procReadConsoleInputC  = kernel32.NewProc("ReadConsoleInputW")
	procNumConsoleEventsC  = kernel32.NewProc("GetNumberOfConsoleInputEvents")
	procFlushConsoleInputC = kernel32.NewProc("FlushConsoleInputBuffer")
)

// minSettle is how long after the prompt a keypress must arrive to count.
//
// ⚠️ Defence in depth, not the main fix. A key already sitting in a buffer
// arrives at roughly zero milliseconds; a human reading two lines and
// deciding cannot. The owned console should make this unnecessary, and it is
// here because the first version of this gate was defeated by exactly the
// thing it guards against.
const minSettle = 250 * time.Millisecond

// settleWindow is how long the fresh console is drained before the prompt is
// shown, to absorb the focus and buffer-size events a new console emits.
const settleWindow = 200 * time.Millisecond

// ownConsole detaches from any inherited console and creates one we own.
//
// Returns the CONIN$ handle, which is the console's OWN input rather than
// whatever a std handle happens to point at.
func ownConsole() (syscall.Handle, error) {
	// ⚠️ The error is deliberately ignored: there may be no console attached,
	// which is the -H=windowsgui case and is correct rather than a failure.
	procFreeConsole.Call()

	if r, _, err := procAllocConsole.Call(); r == 0 {
		return 0, fmt.Errorf("AllocConsole: %w", err)
	}

	// ⚠️ UTF-8 ON THE NEW CONSOLE, OR EVERY SYMBOL WE PRINT IS MOJIBAKE.
	// A fresh console inherits the system OEM codepage -- measured 437 on the
	// machine this was written for -- and Go writes UTF-8. So the first real
	// click produced:
	//
	//	Γôÿ the Go toolchain fetches and compiles first ...
	//	ΓÜá∩╕Å UNVERIFIED, and accepted only because ...
	//
	// ⚠️ THAT IS NOT COSMETIC. The ⚠️ and ⓘ markers are how this repo
	// distinguishes a warning from an aside, and the refusal messages R2 and
	// R6 produce are the ones a user most needs to read. Mangling them makes
	// the careful wording worse than plain text would have been.
	//
	// The error is ignored on purpose: a console that will not take UTF-8 is
	// still a usable console, and failing the consent prompt over a codepage
	// would trade a cosmetic fault for a functional one.
	procSetConsoleOutputCP.Call(uintptr(cpUTF8))

	// Output rebound to the NEW console. AllocConsole alone leaves a window
	// that stays empty, which console_windows.go already records.
	co, err := os.OpenFile("CONOUT$", os.O_WRONLY, 0)
	if err != nil {
		return 0, fmt.Errorf("CONOUT$: %w", err)
	}
	os.Stdout, os.Stderr = co, co

	// ⚠️ RDWR, not RDONLY. ReadConsoleInput and FlushConsoleInputBuffer both
	// require write access to the input handle, and opening read-only yields
	// a handle that fails both with ERROR_ACCESS_DENIED.
	ci, err := os.OpenFile("CONIN$", os.O_RDWR, 0)
	if err != nil {
		return 0, fmt.Errorf("CONIN$: %w", err)
	}
	return syscall.Handle(ci.Fd()), nil
}

// askConsent shows prompt on a console we own and returns true only if the
// user pressed a key there.
func askConsent(show func(), d time.Duration) bool {
	// ⚠️ NO CONSOLE MEANS NO CONSENT MEANS NO RUN, and that is the whole
	// mechanism rather than a limitation. A caller that wants no windows
	// gets an abort, which is what it wanted anyway.
	if windowMode() == WindowNone {
		_log("click ABORTED: %s=none, so no console was created, so consent "+
			"could not be asked and nothing ran", windowEnv)
		return false
	}

	// ⚠️ BEFORE the console exists, or the previous holder is unknowable.
	// A no-op unless the mode is background.
	keeper := rememberFocus()
	h, err := ownConsole()
	if err != nil {
		_log("click REFUSED: no console of our own could be obtained, so "+
			"consent cannot be asked for and nothing will run: %v", err)
		return false
	}

	keeper.restore()

	// Drain the arrival burst BEFORE the prompt is shown, then flush. A new
	// console reports focus and buffer-size events immediately, and one of
	// those arriving after a flush would otherwise be the first thing the
	// poll sees.
	time.Sleep(settleWindow)
	procFlushConsoleInputC.Call(uintptr(h))

	show()
	shownAt := time.Now()
	return pollForKey(h, shownAt, d)
}

// pollForKey returns true on a genuine key-down arriving after minSettle.
func pollForKey(h syscall.Handle, shownAt time.Time, d time.Duration) bool {
	deadline := shownAt.Add(d)
	var buf [16 * 20]byte
	for time.Now().Before(deadline) {
		var n uint32
		if r, _, _ := procNumConsoleEventsC.Call(uintptr(h),
			uintptr(unsafe.Pointer(&n))); r == 0 {
			// ⚠️ Cannot read the input at all, so no key can ever be
			// confirmed. FALSE, which aborts. Returning true here is the
			// bug this file exists to prevent.
			_log("click: cannot read the console input buffer, so consent " +
				"cannot be confirmed")
			return false
		}
		if n > 0 {
			var read uint32
			if r, _, _ := procReadConsoleInputC.Call(uintptr(h),
				uintptr(unsafe.Pointer(&buf[0])), uintptr(n),
				uintptr(unsafe.Pointer(&read))); r != 0 {
				for i := 0; i < int(read) && i < 16; i++ {
					if !isKeyDown(buf[i*20:]) {
						continue
					}
					if time.Since(shownAt) < minSettle {
						// A key this fast was queued, not typed.
						_log("click: a key event arrived %v after the prompt, "+
							"under the %v minimum, so it was discarded as "+
							"queued rather than typed",
							time.Since(shownAt), minSettle)
						continue
					}
					return true
				}
			}
		}
		time.Sleep(40 * time.Millisecond)
	}
	return false
}

// isKeyDown reports whether a 20-byte INPUT_RECORD is a key being pressed.
//
// INPUT_RECORD is EventType (WORD) at 0, two bytes of padding, then the
// union at 4. KEY_EVENT is 1, and KEY_EVENT_RECORD.bKeyDown is the BOOL at
// offset 4. Everything else -- mouse, focus, menu, buffer-size -- has a
// different EventType and must not count.
func isKeyDown(rec []byte) bool {
	if len(rec) < 20 {
		return false
	}
	eventType := uint16(rec[0]) | uint16(rec[1])<<8
	if eventType != 1 {
		return false
	}
	down := uint32(rec[4]) | uint32(rec[5])<<8 | uint32(rec[6])<<16 |
		uint32(rec[7])<<24
	return down != 0
}
