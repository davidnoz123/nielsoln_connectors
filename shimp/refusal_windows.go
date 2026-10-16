//go:build windows

// refusal_windows.go -- make a refused click VISIBLE, and hold it long enough
// to read.
//
// ⚠️ WHY THIS EXISTS, AND IT IS A MEASURED FAILURE OF THE PRODUCT RATHER THAN
// OF THE TOOLING.
//
// A human clicked eight links on 15 Oct 2026 and reported, twice: "nothing
// seems to be happening", then "all I see when I click is a command prompt
// flash". Both were correct and both were the same hole:
//
//	built -H=windowsgui          no window at all, total silence
//	built console-subsystem      a window appears and shows NOTHING
//
// The second is WORSE than the first. A window appearing implies something
// happened and then tells you nothing, because `_log` writes to the file and
// only falls back to stderr when it cannot.
//
// So a user could not distinguish a refused link from: the scheme not
// registered, the handler dying before `_log` existed, Chrome blocking the
// launch, or a malformed link. ⚠️ That is this repo's cardinal sin, a silent
// failure, in the thing being shipped.
//
// ⚠️ AND THE ANSWER WAS ALREADY DECIDED AND BUILT, which is the part worth
// recording. console_windows.go's header says it outright:
//
//	"can a keeper start with NO console and allocate one only when it turns
//	 out to have something to show? The answer is yes, and it is
//	 AllocConsole ... no window for a click that has nothing to say, and a
//	 window the moment something does."
//
// `allocConsole` was written for that and never called by anything. The
// design was reasoned, written down, implemented, and left unwired, which is
// the third instance of that exact pattern in two days.
//
// THE ASYMMETRY IS THE WHOLE DESIGN. A refusal is an error the user must
// see; a success has nothing to say. So s50's refusal of a console field
// gets stronger rather than weaker: nothing is declared in advance, and the
// window appears only where there is something to read.
package main

import (
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"
)

// showRefusal puts a refusal where a human will see it.
//
// Called on every refusal path and nowhere else. It never returns an error:
// a failure to display must not change what the handler did, and the log
// already holds the authoritative record either way.
func showRefusal(reason string) {
	// ⚠️ The log already holds the authoritative record, so declining to
	// DISPLAY a refusal loses nothing a harness needs. A human gets the
	// window; an unattended run gets the log.
	if os.Getenv(noConsoleEnv) == "1" {
		_log("click: %s is set, so the refusal was logged and not shown",
			noConsoleEnv)
		return
	}

	// A console only if there is something to say, which is now.
	if err := allocConsole(); err != nil {
		// Reported to the log rather than swallowed. If this fails the user
		// is back to the flash, and the log is the only way anybody finds
		// out why.
		_log("click: could not allocate a console to show the refusal: %v", err)
		return
	}

	fmt.Fprintf(os.Stdout, "\n  %s:// REFUSED\n\n  %s\n\n", scheme, wrap(reason, 72))
	fmt.Fprintf(os.Stdout, "  the full record is in %s\n",
		logPathForDisplay())

	if os.Getenv(noHoldEnv) == "1" {
		return
	}
	fmt.Fprintf(os.Stdout, "\n  press any key, or this closes in %ds\n",
		holdSeconds)
	waitForKeyOrTimeout(holdSeconds * time.Second)
}

// logPathForDisplay is the log's location, or a note saying why it is not
// known. ⚠️ Never an empty string: pointing a user at nothing is worse than
// telling them the path could not be determined.
func logPathForDisplay() string {
	dir, err := home()
	if err != nil {
		return fmt.Sprintf("(the log location could not be determined: %v)", err)
	}
	return dir + `\` + logName
}

var (
	procReadConsoleInput   = kernel32.NewProc("ReadConsoleInputW")
	procGetNumberOfConsole = kernel32.NewProc("GetNumberOfConsoleInputEvents")
	procGetStdHandle       = kernel32.NewProc("GetStdHandle")
	procFlushConsoleInput  = kernel32.NewProc("FlushConsoleInputBuffer")
)

const stdInputHandle = ^uintptr(10) + 1 // (DWORD)-10

// waitForKeyOrTimeout returns on the first key press or when d elapses.
//
// ⚠️ POLLED, not blocking. ReadConsoleInput blocks with no timeout, so a
// blocking read would ignore the deadline entirely and hang an unattended
// run that had not set SHIMP_NO_HOLD. Polling the event count costs nothing
// at 50ms and honours both exits.
//
// ⚠️ And the input buffer is FLUSHED first. A console freshly allocated can
// already hold events -- focus changes, mouse moves -- and treating one of
// those as the user's keypress would close the window before it was read,
// which is the bug this whole file exists to fix, reintroduced one layer
// down.
// ⚠️ Returns nothing on purpose. A caller that needs to know WHETHER a key
// was pressed must use waitForKeyPress, whose bool cannot be mistaken for
// anything else. Consent is decided on that bool; this wrapper is for
// holding a window open, where the distinction does not matter.
// ⚠️ AND IT HONOURS noHoldEnv ITSELF, rather than leaving each caller to
// remember. The check used to live only in showRefusal, so the ABORTED path
// in runLink called straight through and sat for 25 seconds: the test suite
// went from 3s to 61s and the env var that was supposed to prevent exactly
// that had no effect there. A guard every caller must remember is a guard
// one caller will forget.
func waitForKeyOrTimeout(d time.Duration) {
	if os.Getenv(noHoldEnv) == "1" {
		return
	}
	_ = waitForKeyPress(d)
}

// waitForKeyPress returns true if a key was pressed before d elapsed.
func waitForKeyPress(d time.Duration) bool {
	h, _, _ := procGetStdHandle.Call(stdInputHandle)
	if h == 0 || h == uintptr(syscall.InvalidHandle) {
		// ⚠️ No input handle means no key can ever arrive, so this must
		// report FALSE rather than sleeping and leaving the caller to guess.
		// Consent depends on it: returning true here would execute code
		// because nobody could be asked.
		time.Sleep(d)
		return false
	}
	procFlushConsoleInput.Call(h)

	deadline := time.Now().Add(d)
	// KEY_EVENT records are 20 bytes inside a 20-byte INPUT_RECORD union on
	// amd64; a buffer of 16 records is more than one poll can ever need.
	var buf [16 * 20]byte
	for time.Now().Before(deadline) {
		var n uint32
		if r, _, _ := procGetNumberOfConsole.Call(h,
			uintptr(unsafe.Pointer(&n))); r == 0 {
			time.Sleep(d)
			return false
		}
		if n > 0 {
			var read uint32
			procReadConsoleInput.Call(h, uintptr(unsafe.Pointer(&buf[0])),
				uintptr(n), uintptr(unsafe.Pointer(&read)))
			// EventType is the first uint16; 1 is KEY_EVENT. bKeyDown is the
			// first field of the record, 4 bytes in.
			for i := 0; i < int(read) && i < 16; i++ {
				rec := buf[i*20:]
				if rec[0] == 1 && rec[1] == 0 && rec[4] != 0 {
					return true
				}
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	return false
}
