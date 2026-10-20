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
	"io"
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
	// ⚠️ SAY WHAT IS HAPPENING AND HOW LONG IT MAY BE SILENT. `running...`
	// alone was followed by up to 27 seconds of nothing: s49 measured
	// `go run module@sha` at 24-27s cold and 1.1-4.8s warm, and `go` speaks
	// only during the DOWNLOAD half ("go: downloading ...") -- the compile
	// and link that follow are silent.
	//
	// The capability cannot help here, because its main() does not exist yet.
	// This is the last thing that runs before the toolchain takes over, so it
	// is the only place the wait can be explained.
	//
	// ⓘ AND A HEARTBEAT NOW DOES EXIST, below. This comment said there was
	// deliberately none, on the grounds that a tick could interleave with the
	// child's output. That objection was real and piping removed it; the
	// comment is corrected rather than left to contradict the code.
	fmt.Fprintf(os.Stdout, "\n  running  %s\n", strings.Join(argv, " "))
	fmt.Fprintf(os.Stdout,
		"  ⓘ the Go toolchain fetches and compiles first. A first run of a "+
			"given SHA was measured at 24-27s and\n"+
			"    a repeat at about 1-5s. It is silent while compiling, so a "+
			"pause here is expected, not a hang.\n\n")

	cmd := exec.Command(argv[0], argv[1:]...)
	// ⚠️ PIPED AND RELAYED RATHER THAN INHERITED, so a heartbeat is possible.
	//
	// s49 chose inheritance because capturing output "hides a long fetch", and
	// that reasoning was right about CAPTURING. Relaying line by line hides
	// nothing: every byte still reaches the console, in order, as it arrives.
	//
	// What inheritance made impossible was knowing WHEN the child starts to
	// speak -- and without that, a ticking progress line could land in the
	// middle of a line the capability was writing. I used that as the reason
	// not to have one. ⚠️ The owner's first real click was silent for 25
	// seconds and he said so, which is better evidence than my argument: the
	// objection was to interleaving, and piping removes it rather than
	// trading it away.
	//
	// Safe for a Go child specifically: writes to os.Stdout are unbuffered,
	// so relaying does not delay anything. A child that block-buffers when
	// its stdout is not a terminal would appear later than it does now, and
	// that is a real limit of this approach rather than a hidden one.
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return 1
	}
	cmd.Stderr = cmd.Stdout
	if err := cmd.Start(); err != nil {
		_log("click: cannot start: %v", err)
		fmt.Fprintf(os.Stdout, "\n  FAILED to start: %v\n", err)
		return 1
	}

	spoke := make(chan struct{})
	go heartbeat(spoke)
	relay(stdout, spoke)
	err = cmd.Wait()

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

// heartbeat prints elapsed seconds until the child says something.
//
// ⚠️ IT STOPS AT THE FIRST BYTE, which is the whole design. Ticking alongside
// the capability's own output would be noise competing with the thing the user
// is trying to read; ticking only into silence is information.
//
// The first tick waits `firstTick` rather than firing immediately, so a warm
// run -- measured at 1 to 5 seconds for a SHA already in the module cache --
// usually prints nothing at all. A progress line on a fast run is clutter.
func heartbeat(spoke <-chan struct{}) {
	start := time.Now()
	t := time.NewTicker(tickEvery)
	defer t.Stop()
	first := time.NewTimer(firstTick)
	defer first.Stop()

	select {
	case <-spoke:
		return
	case <-first.C:
	}
	fmt.Fprintf(os.Stdout, "  ... still working, %ds elapsed. The Go "+
		"toolchain is fetching or compiling.\n", int(time.Since(start).Seconds()))
	for {
		select {
		case <-spoke:
			return
		case <-t.C:
			fmt.Fprintf(os.Stdout, "  ... %ds\n",
				int(time.Since(start).Seconds()))
		}
	}
}

// relay copies the child's output through, closing spoke on the first byte.
//
// ⚠️ Copied in SMALL READS rather than with io.Copy, because io.Copy would
// not tell us when the first byte arrived without wrapping the reader anyway,
// and the wrapper is the only thing this function is for.
func relay(r io.Reader, spoke chan struct{}) {
	buf := make([]byte, 4096)
	closed := false
	for {
		n, err := r.Read(buf)
		if n > 0 {
			if !closed {
				// Stop the heartbeat BEFORE writing, so a tick cannot land
				// between here and the first line of real output.
				close(spoke)
				closed = true
			}
			os.Stdout.Write(buf[:n])
		}
		if err != nil {
			if !closed {
				close(spoke)
			}
			return
		}
	}
}

// firstTick and tickEvery pace the heartbeat.
//
// ⚠️ MEASURED, not chosen for feel. s49 put `go run module@sha` at 24-27s
// cold and 1.1-4.8s warm, so a first tick at 4s stays silent through almost
// every warm run and speaks early in a cold one.
const (
	firstTick = 4 * time.Second
	tickEvery = 5 * time.Second
)
