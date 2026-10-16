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

// noConsoleEnv stops a click creating any window at all.
//
// ⚠️ WHY THIS REPLACED A FOCUS DANCE, which is worth recording because the
// first attempt looked reasonable and did not work.
//
// A clicked console-subsystem handler is GIVEN a console by Windows that is
// visible and foreground, so a corpus run stole the keyboard eight times.
// The first fix remembered the previous foreground window, then called
// ShowWindow(SW_SHOWNOACTIVATE) and handed the focus back.
//
// It half worked: 5 of 8 dispatches left the focus alone and 3 logged a
// refusal. ⚠️ AllocConsole ACTIVATES the window before anything can be told
// not to, and showing-without-activating does not deactivate something that
// is already active. There was no ordering that fixed it, because the window
// exists before our first instruction about it.
//
// So the console is not created. Consent therefore CANNOT be asked, and
// consent_windows.go already fails closed: no console means no consent means
// no run. A refusal is logged and not displayed.
//
// ⚠️ SAFE BECAUSE IT FAILS CLOSED. The only behaviour it can produce is
// "nothing ran", which is what an unattended run must do anyway. Contrast
// assumeConsentEnv, whose whole effect is to let something run.
//
// ⚠️ AND IT CHANGES WHICH BRANCH IS EXERCISED, stated rather than hidden:
// the corpus now reaches the no-console abort rather than the key-polling
// one. That is a real cost. The key-polling branch is covered by
// TestAKeyTooFastIsNotTyped and by the real click a human made on 15 Oct,
// which is weaker than a harness exercising it and is not nothing.
const noConsoleEnv = "SHIMP_NO_CONSOLE"
