//go:build windows

package main

import (
	"os"
	"strings"
	"testing"
	"time"
)

// capture swaps os.Stdout for a pipe and returns what was written to it.
func capture(t *testing.T, f func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	done := make(chan string, 1)
	go func() {
		buf := make([]byte, 1<<16)
		n, _ := r.Read(buf)
		done <- string(buf[:n])
	}()
	f()
	w.Close()
	os.Stdout = old
	select {
	case s := <-done:
		return s
	case <-time.After(2 * time.Second):
		return ""
	}
}

// TestTheHeartbeatStaysQUIETOnAFastChild is the one that matters.
//
// ⚠️ A progress line on a warm run is clutter, and s49 measured warm runs at
// 1.1 to 4.8 seconds. firstTick is 4s so almost every warm run prints
// nothing. A heartbeat that chattered would have made the complaint it was
// written to answer worse.
func TestTheHeartbeatStaysQUIETOnAFastChild(t *testing.T) {
	got := capture(t, func() {
		spoke := make(chan struct{})
		go heartbeat(spoke)
		time.Sleep(200 * time.Millisecond) // a child that speaks at once
		close(spoke)
		time.Sleep(200 * time.Millisecond)
	})
	if strings.Contains(got, "still working") || strings.Contains(got, "...") {
		t.Errorf("the heartbeat spoke on a fast child, which is the clutter "+
			"it exists to avoid:\n%q", got)
	}
}

// TestTheHeartbeatSpeaksIntoSilence is the other half.
func TestTheHeartbeatSpeaksIntoSilence(t *testing.T) {
	if testing.Short() {
		t.Skip("needs to outlast firstTick")
	}
	got := capture(t, func() {
		spoke := make(chan struct{})
		go heartbeat(spoke)
		time.Sleep(firstTick + 700*time.Millisecond)
		close(spoke)
		time.Sleep(200 * time.Millisecond)
	})
	if !strings.Contains(got, "still working") {
		t.Errorf("the heartbeat said nothing after %s of silence, which is "+
			"the 25 seconds of nothing it exists to fix. Got: %q",
			firstTick, got)
	}
	if !strings.Contains(got, "Go") {
		t.Errorf("the tick does not say WHAT is working, so it reads as a "+
			"spinner rather than information: %q", got)
	}
}

// TestRelayPassesEverythingThrough pins that piping did not become capturing.
//
// ⚠️ s49's objection to capturing was that it hides a long fetch. Relaying
// must therefore lose nothing, in order.
func TestRelayPassesEverythingThrough(t *testing.T) {
	const payload = "line one\nline two\n⚠️ and a UTF-8 marker\n"
	got := capture(t, func() {
		spoke := make(chan struct{})
		relay(strings.NewReader(payload), spoke)
	})
	if got != payload {
		t.Errorf("the relay altered the child's output.\n got %q\nwant %q",
			got, payload)
	}
}

// TestRelayClosesSpokeEvenWithNoOutput stops a silent child hanging the tick.
func TestRelayClosesSpokeEvenWithNoOutput(t *testing.T) {
	spoke := make(chan struct{})
	go relay(strings.NewReader(""), spoke)
	select {
	case <-spoke:
	case <-time.After(2 * time.Second):
		t.Error("a child that produced NO output left the heartbeat ticking " +
			"for ever, because spoke was never closed")
	}
}
