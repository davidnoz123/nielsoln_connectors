//go:build windows

// run_windows.go -- the moment shimp stops being inert and runs something.
//
// ⚠️ THE MOST CONSEQUENTIAL FILE IN THIS PROGRAM. Everything before it parsed,
// hashed and refused. This executes code named by a link that arrived from
// outside, which is the whole point of the protocol and also the whole risk.
//
// ⚠️ AND s64 IS WHY CONSENT IS HERE RATHER THAN LATER. Measured 15 Oct 2026:
// once a user grants Chrome's per-origin permission, a web page can fire a
// shimp:// link with NO GESTURE AT ALL -- a meta refresh, no JavaScript, no
// click, reproducible three runs of three. While nothing ran, that was
// harmless. The moment this file exists, a visited page could make the
// machine run `go run <module>@<sha>`.
//
// So the gate is not an afterthought: a click shows WHAT IT WILL RUN and
// waits for a human keypress. Nothing runs without one.
//
// ⚠️ THE TIMEOUT ABORTS, IT DOES NOT PROCEED. This is the single most
// important line in the file. A consent prompt that runs the thing when
// nobody answers is not consent; it is a delay. The refusal display in
// refusal_windows.go times out by CLOSING, which is safe because it has
// already done its job. This one times out by DOING NOTHING.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// consentSeconds is how long a click waits for a human.
//
// Long enough to read an argv and decide, short enough that an unattended
// machine does not sit with a window open for ever. It aborts at the end,
// so erring long costs nothing but a window.
const consentSeconds = 30

// assumeConsentEnv skips the prompt. ⚠️ TEST HARNESSES ONLY.
//
// Named so that nobody can set it believing it makes anything safer, and
// deliberately NOT the same variable as SHIMP_NO_HOLD. That one suppresses a
// WAIT after a refusal has already been shown, which is cosmetic. This one
// removes a human from the decision to execute code, which is not.
//
// It is honest about its own weakness: anybody who can set this can already
// run programs. It exists so linkprobe can exercise the run path, and the
// run log records when it was used.
const assumeConsentEnv = "SHIMP_ASSUME_CONSENT"

// runLink asks, then runs. It returns the process exit code to use.
func runLink(u *URI) int {
	argv := u.Argv()
	if len(argv) == 0 {
		_log("click REFUSED: the link produced an empty argv, which should be " +
			"impossible after parsing. Refusing rather than guessing.")
		showRefusalFn("the link produced no command to run")
		return 1
	}

	// ⚠️ THE PROMPT IS A CLOSURE, and that is not a style choice. askConsent
	// detaches from any inherited console and allocates its own, rebinding
	// os.Stdout to it. Anything printed BEFORE that call goes to the console
	// being abandoned, so the user would see an empty window and the prompt
	// would land where nobody is looking.
	show := func() {
		fmt.Fprintf(os.Stdout, "\n  %s:// wants to run\n\n", scheme)
		fmt.Fprintf(os.Stdout, "    %s\n\n", strings.Join(argv, " "))
		fmt.Fprintf(os.Stdout, "  identity  %s\n", u.Identity())
		fmt.Fprintf(os.Stdout, "  over      %d bytes\n", len(u.Raw))
		if u.HasAction {
			fmt.Fprintf(os.Stdout, "  action    %s   (NOT part of the identity)\n",
				u.Action)
		}
		fmt.Fprintf(os.Stdout,
			"\n  press any key to RUN it, or wait %ds to abort\n", consentSeconds)
	}

	if os.Getenv(assumeConsentEnv) == "1" {
		if err := allocConsole(); err != nil {
			_log("click: no console for the assumed-consent path: %v", err)
		}
		show()
		_log("click: consent ASSUMED via %s, which a test harness sets and "+
			"nothing else should", assumeConsentEnv)
		fmt.Fprintf(os.Stdout, "\n  consent assumed via %s\n", assumeConsentEnv)
	} else if !askConsentFn(show, consentSeconds*time.Second) {
		// ⚠️ EVERY failure lands here and they all mean the same thing: no
		// console of our own, input that cannot be read, a key too fast to
		// have been typed, or the timeout. None of them means "run it".
		_log("click ABORTED: consent was not given, so nothing ran")
		fmt.Fprintf(os.Stdout, "\n  ABORTED. Nothing ran.\n")
		waitForKeyOrTimeout(holdSeconds * time.Second)
		return 1
	} else {
		_log("click: consent given by a keypress on a console we own")
	}

	_log("click: running %q", argv)
	fmt.Fprintf(os.Stdout, "\n  running...\n\n")

	cmd := exec.Command(argv[0], argv[1:]...)
	// ⚠️ INHERITED, so the user watches it work. The alternative is capturing
	// and replaying, which hides a long fetch: s49 measured `go run
	// module@sha` at 27.4s cold, and a window showing nothing for half a
	// minute is the flash problem again at a larger scale.
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	code := 0
	if err != nil {
		code = 1
		if ee, ok := err.(*exec.ExitError); ok {
			code = ee.ExitCode()
		}
	}
	_log("click: exited %d", code)
	fmt.Fprintf(os.Stdout, "\n  exited %d\n", code)

	// Held open regardless of outcome, because the output is the entire
	// point and the window closes with the process.
	fmt.Fprintf(os.Stdout, "\n  press any key to close, or %ds\n", holdSeconds)
	waitForKeyOrTimeout(holdSeconds * time.Second)
	return code
}
