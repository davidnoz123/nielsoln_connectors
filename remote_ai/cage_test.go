// The cage is switched off by one boolean, so that boolean is worth testing.
//
// isLoopback decides whether to confine the session. If it ever answers true
// for a real bridge, the cage silently never engages, every other check stays
// green, and the only symptom is a promise on a web page that is no longer
// true. That is the exact shape of defect this repo keeps finding, so the
// switch gets a test even though it is four lines.
//
//	go test ./...
package main

import "testing"

func TestLoopbackIsRecognisedByAddressNotBySpelling(t *testing.T) {
	// Every one of these is this machine. 127.0.0.1 is not the only loopback
	// address, and a check that compared strings would miss most of them and
	// confine a session that then could not reach its bridge.
	for _, host := range []string{
		"", "localhost", "127.0.0.1", "127.0.0.2", "127.1.2.3", "::1",
	} {
		if !isLoopback(host) {
			t.Errorf("%q is this machine, and a confined session cannot "+
				"reach it", host)
		}
	}
}

func TestARealBridgeIsNotMistakenForThisMachine(t *testing.T) {
	// The consequence of a false positive here is the quiet one: the cage is
	// skipped, nothing fails, and the session runs unconfined while the page
	// says otherwise.
	for _, host := range []string{
		"bridge.nielsoln.com", "example.com", "10.0.0.1", "192.168.1.10",
		"8.8.8.8", "2001:4860:4860::8888",
	} {
		if isLoopback(host) {
			t.Errorf("%q is not this machine, and treating it as one would "+
				"skip the cage without saying so", host)
		}
	}
}

func TestATestProcessIsNotInsideTheCage(t *testing.T) {
	// Pins the direction of the gate. inCage() reports what the KERNEL says
	// about this token, so an ordinary `go test` process must be outside; a
	// version that answered true by default would open the exec gate on every
	// machine, confined or not.
	if inCage() {
		t.Fatal("a plain test process reported itself confined, so inCage() " +
			"is not asking the kernel")
	}
}
