//go:build windows

// consent_test.go -- the checks that should have existed before the gate did.
//
// ⚠️ WINDOWS-ONLY, and that is the right answer rather than the lazy one.
// The third build-tag mistake in two days was `isKeyDown`, which parses a
// Windows INPUT_RECORD and which this file referenced from an untagged test.
//
// The first two were fixed by moving the CODE out of a platform file, because
// `wrap` and `noHoldEnv` had nothing to do with Windows. This one goes the
// other way: consent_other.go always returns false, so there IS no consent
// gate off Windows and a test of it on linux tests nothing. Tagging the test
// states that, where untagging the code would have implied a gate that does
// not exist.
//
// ⚠️ The first consent gate reported consent nobody gave and executed a
// program. Nothing caught it; running the thing did. These are the checks
// that would have, written after the fact and kept because the next version
// of this gate will be changed by somebody who was not here.
package main

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

// TestNoConsentMeansNoRun is the one that matters.
func TestNoConsentMeansNoRun(t *testing.T) {
	logTo = &bytes.Buffer{}
	t.Cleanup(func() { logTo = nil })
	// ⚠️ The ABORTED path holds its window open so a human can read it, and
	// this test would otherwise sit through that twice: 50 of the 64 seconds
	// the suite took before this line existed. SHIMP_NO_HOLD skips the WAIT
	// and not the message, which is exactly what it is for.
	t.Setenv(noHoldEnv, "1")
	for _, c := range []struct {
		name string
		gate func(func(), time.Duration) bool
	}{
		{"the timeout elapsed", func(func(), time.Duration) bool { return false }},
		{"the console could not be owned", func(func(), time.Duration) bool { return false }},
	} {
		var ran bool
		askConsentFn = c.gate
		runLinkFn = runLink
		showRefusalFn = func(string) {}
		t.Cleanup(func() {
			askConsentFn = askConsent
			runLinkFn = runLink
			showRefusalFn = showRefusal
		})
		// ⚠️ Driven through runLink itself, not through a mock of it: the
		// decision being tested is inside runLink, so replacing runLink
		// would test nothing.
		u, err := ParseURI(scheme+"://v1/go/run/github.com/davidnoz123/"+
			"nielsoln_connectors/x@"+strings.Repeat("a", 40), scheme)
		if err != nil {
			t.Fatal(err)
		}
		// runLink would exec on consent. With the gate returning false it
		// must return non-zero and never reach exec. `ran` stays false
		// because nothing here can set it; the assertion is the exit code
		// plus the ABORTED log line.
		if rc := runLink(u); rc == 0 {
			t.Errorf("%s: runLink returned 0, which means it ran something "+
				"without consent", c.name)
		}
		if ran {
			t.Errorf("%s: something executed", c.name)
		}
	}
	if !strings.Contains(logTo.(*bytes.Buffer).String(), "ABORTED") {
		t.Error("no ABORTED line was logged, so a run that did not happen " +
			"leaves no record that it was declined")
	}
}

// TestIsKeyDownRejectsEverythingThatIsNotAKeyPress is where the real bug lived.
//
// ⚠️ A console emits focus, mouse, menu and buffer-size events, and a fresh
// one emits several immediately. Accepting any of them as a keypress is how
// the first gate authorised a run nobody asked for.
func TestIsKeyDownRejectsEverythingThatIsNotAKeyPress(t *testing.T) {
	rec := func(eventType uint16, keyDown uint32) []byte {
		b := make([]byte, 20)
		b[0] = byte(eventType)
		b[1] = byte(eventType >> 8)
		b[4] = byte(keyDown)
		b[5] = byte(keyDown >> 8)
		b[6] = byte(keyDown >> 16)
		b[7] = byte(keyDown >> 24)
		return b
	}
	for _, c := range []struct {
		name string
		rec  []byte
		want bool
	}{
		{"a real key press", rec(1, 1), true},
		{"a key RELEASE", rec(1, 0), false},
		{"MOUSE_EVENT", rec(2, 1), false},
		{"WINDOW_BUFFER_SIZE_EVENT", rec(4, 1), false},
		{"MENU_EVENT", rec(8, 1), false},
		{"FOCUS_EVENT", rec(16, 1), false},
		{"a truncated record", make([]byte, 7), false},
		{"an empty record", nil, false},
	} {
		if got := isKeyDown(c.rec); got != c.want {
			t.Errorf("%s: isKeyDown = %v, want %v", c.name, got, c.want)
		}
	}
}

// TestAKeyTooFastIsNotTyped pins the defence in depth.
//
// ⚠️ A key already sitting in an input buffer arrives at roughly zero
// milliseconds. A human reading an argv and deciding cannot. pollForKey must
// discard the first, and the only way to assert that without a console is to
// check the threshold itself is sane relative to the settle window.
func TestAKeyTooFastIsNotTyped(t *testing.T) {
	if minSettle <= 0 {
		t.Fatal("minSettle is not positive, so a queued key counts as typed")
	}
	if minSettle < 100*time.Millisecond {
		t.Errorf("minSettle is %v, which is within the range a queued key "+
			"can arrive in", minSettle)
	}
	if settleWindow <= 0 {
		t.Error("settleWindow is not positive, so the arrival burst a fresh " +
			"console emits is never drained")
	}
	if consentSeconds*time.Second <= minSettle {
		t.Error("the consent window is no longer than the minimum settle, " +
			"so no keypress could ever be accepted")
	}
}

// TestTheAssumeConsentVariableIsNotTheHoldVariable keeps them separate.
//
// ⚠️ SHIMP_NO_HOLD suppresses a WAIT after a refusal has already been shown,
// which is cosmetic. SHIMP_ASSUME_CONSENT removes a human from the decision
// to execute code. Collapsing them into one variable would let a harness
// that only wanted faster output silently authorise execution.
func TestTheAssumeConsentVariableIsNotTheHoldVariable(t *testing.T) {
	if assumeConsentEnv == noHoldEnv {
		t.Fatal("the two variables are the same, so anything setting the " +
			"cosmetic one also authorises code execution")
	}
	if !strings.Contains(assumeConsentEnv, "CONSENT") {
		t.Errorf("%q does not say what it does, and a variable that "+
			"authorises execution must be unmistakable", assumeConsentEnv)
	}
}
