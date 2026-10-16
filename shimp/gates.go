// gates.go -- the numbers and names that govern a click's two gates.
//
// ⚠️ PLATFORM-INDEPENDENT, AND THAT IS THE POINT OF THE FILE.
//
// These lived in refusal_windows.go and run_windows.go, behind `//go:build
// windows`, and inherited that tag by proximity rather than by need. Nothing
// about a duration or an environment-variable name is Windows-specific, and
// `GOOS=linux go vet` failed twice on platform-independent tests referencing
// them: once for `wrap`, which moved to text.go, and once for `noHoldEnv`.
//
// Twice is a pattern. A value that is not platform-specific does not live in
// a platform file, however naturally it was written next to the code using
// it.
//
// # THE TWO GATES ARE NOT THE SAME KIND OF THING
//
//	the REFUSAL hold   a window stays readable after a decision was made.
//	                   Cosmetic. Shortening or skipping it changes nothing
//	                   about what happened.
//
//	the CONSENT window a human is asked whether to EXECUTE. Shortening it is
//	                   safe because it still ABORTS; skipping it is not,
//	                   because the only other answer is to run.
//
// That asymmetry is why there are three variables rather than one, and why
// the one that can authorise execution is named so nobody sets it by
// accident.
package main

import "os"

// holdSeconds is how long a refusal or a finished run stays readable.
const holdSeconds = 25

// noHoldEnv skips the WAIT after something has already been shown.
//
// ⚠️ It never suppresses the MESSAGE. A variable that hid a reason would be
// the opposite of the purpose: the refusal is still written to the console
// and still written to the log.
const noHoldEnv = "SHIMP_NO_HOLD"

// consentSeconds is how long a click waits for a human, by default.
//
// Long enough to read an argv and decide, short enough that an unattended
// machine does not sit with a window open for ever. It aborts at the end, so
// erring long costs nothing but a window.
const consentSeconds = 30

// consentSecondsEnv shortens the consent WAIT. It cannot grant consent.
//
// ⚠️ WHY THIS EXISTS, and the user spotted it rather than the harness.
// linkprobe's corpus dispatches eight shapes, four of which are one-shots
// that reach the gate. Nobody is at the keyboard, so each sat for the full
// thirty seconds: the corpus went from 28s to 82s, and the slow part was the
// HANDLER waiting rather than the harness polling.
//
// ⚠️ DELIBERATELY NOT assumeConsentEnv. That one skips the gate and then
// EXECUTES, which an unattended run must never do. This shortens the window
// and still ABORTS at the end, so the real branch runs and nothing else does.
const consentSecondsEnv = "SHIMP_CONSENT_SECONDS"

// assumeConsentEnv skips the prompt entirely. ⚠️ TEST HARNESSES ONLY.
//
// Named so nobody sets it believing it makes anything safer. It removes a
// human from the decision to execute code. Honest about its own weakness:
// anybody who can set this can already run programs. Every use is logged.
const assumeConsentEnv = "SHIMP_ASSUME_CONSENT"

// windowEnv chooses what a click's window does. One setting, three values,
// because "no window", "a window that does not grab focus" and "a window
// that does" are three different wants and were collapsed into one.
//
//	focus       (default) a console appears and TAKES the foreground.
//	            What a real click should do: the user clicked, so the thing
//	            they asked for should come to them.
//
//	background  a console appears and does NOT take the foreground.
//	            ⚠️ BEST EFFORT, and measured as such: AllocConsole activates
//	            the window before anything can be told otherwise, so this
//	            pushes it back afterwards. 5 of 8 dispatches kept the focus
//	            where it was; the other 3 needed the minimise fallback.
//
//	none        no console at all. Consent therefore CANNOT be asked and the
//	            click ABORTS, which is what an unattended run wants anyway.
//	            A refusal is logged rather than shown.
//
// ⚠️ NONE OF THE THREE CAN CAUSE ANYTHING TO RUN. `none` fails closed by
// construction. The setting that can is assumeConsentEnv, which is separate
// and named so nobody reaches for it by accident.
//
// ⚠️ It replaced SHIMP_NO_CONSOLE, which was a boolean that could only say
// "window" or "no window" and so had no way to express the middle case --
// the one a human watching a test run actually wants.
const windowEnv = "SHIMP_WINDOW"

// WindowMode is what the click's window should do.
type WindowMode int

const (
	WindowFocus WindowMode = iota
	WindowBackground
	WindowNone
)

func (w WindowMode) String() string {
	switch w {
	case WindowBackground:
		return "background"
	case WindowNone:
		return "none"
	}
	return "focus"
}

// windowMode reads the setting. ⚠️ An unrecognised value falls back to
// `focus` and SAYS SO, rather than silently choosing the quietest option: a
// typo must not turn a visible click into an invisible one.
func windowMode() WindowMode {
	switch v := os.Getenv(windowEnv); v {
	case "", "focus":
		return WindowFocus
	case "background":
		return WindowBackground
	case "none":
		return WindowNone
	default:
		_log("click: %s=%q is not focus, background or none. Using focus.",
			windowEnv, v)
		return WindowFocus
	}
}
