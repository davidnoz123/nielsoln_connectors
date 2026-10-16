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
	"strconv"
	"strings"
	"time"
)

// consentWindow is how long to wait, honouring the override.
func consentWindow() time.Duration {
	if v := os.Getenv(consentSecondsEnv); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			// ⚠️ Floored at one second rather than allowing zero. At zero the
			// prompt would be printed and the abort logged in the same
			// instant, so a human who happened to be watching could not read
			// what had been declined.
			if n < 1 {
				n = 1
			}
			if n > 600 {
				n = 600
			}
			_log("click: consent window shortened to %ds by %s", n,
				consentSecondsEnv)
			return time.Duration(n) * time.Second
		}
		_log("click: %s=%q is not a number, using the default %ds",
			consentSecondsEnv, v, consentSeconds)
	}
	return consentSeconds * time.Second
}

// runLink asks, then runs. It returns the process exit code to use.
func runLink(u *URI) int {
	argv := u.Argv()
	if len(argv) == 0 {
		_log("click REFUSED: the link produced an empty argv, which should be " +
			"impossible after parsing. Refusing rather than guessing.")
		showRefusalFn("the link produced no command to run")
		return 1
	}

	// ⚠️ COMPUTED ONCE. consentWindow() logs when the override is in play,
	// and calling it for the prompt text and again for the wait logged the
	// same line twice, which reads like two clicks in a log whose whole job
	// is to say what happened.
	window := consentWindow()

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
		// ⚠️ THE REAL WINDOW, not the default. With the override set this
		// printed "wait 30s" while aborting after one second: a prompt lying
		// about its own deadline, which is the same class of defect as the
		// stale errno the focus probe used to print.
		fmt.Fprintf(os.Stdout,
			"\n  press any key to RUN it, or wait %ds to abort\n",
			int(window.Seconds()))
	}

	if os.Getenv(assumeConsentEnv) == "1" {
		if err := allocConsole(); err != nil {
			_log("click: no console for the assumed-consent path: %v", err)
		}
		show()
		_log("click: consent ASSUMED via %s, which a test harness sets and "+
			"nothing else should", assumeConsentEnv)
		fmt.Fprintf(os.Stdout, "\n  consent assumed via %s\n", assumeConsentEnv)
	} else if !askConsentFn(show, window) {
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
