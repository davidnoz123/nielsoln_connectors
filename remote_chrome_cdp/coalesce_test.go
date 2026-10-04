package main

import (
	"sync"
	"testing"
)

// TestClaim pins what a double-click costs, which was thirteen seconds.
//
// Two right-clicks two seconds apart on one cell ran two complete drives,
// 5.3s and then 10.8s, because the supersede check only drops a request
// still QUEUED when a newer one arrives: the first click was already past
// that check and running. The log showed two "begin" lines for one id with
// the focus and done lines of both interleaved, which is what prompted the
// question "what happened here".
func TestClaim(t *testing.T) {
	s := &server{flying: map[string]bool{}}
	const id = "tl-e10-01:edit"

	if !s.claim(id) {
		t.Fatal("the first click must drive")
	}
	if s.claim(id) {
		t.Fatal("the second click, two seconds later, must be ignored")
	}

	// A DIFFERENT CELL IS A DIFFERENT INTENTION and is never ignored. This
	// is the case supersede was built for -- four rows clicked in a row,
	// where the last one is the one wanted -- and it must still get there.
	if !s.claim("tl-e10-01:before") {
		t.Fatal("another cell must drive while this one is opening")
	}

	// THE IGNORED CLICK MUST NOT HAVE TAKEN THE CLAIM AWAY. If it had, the
	// drive still running would release a claim it no longer owns and the
	// cell would accept a third click mid-drive: the original fault, one
	// click later and harder to see.
	s.release(id)
	if !s.claim(id) {
		t.Fatal("a click after the drive finished must drive")
	}

	// Releasing twice is a drive ending after its id was already forgotten.
	// Harmless by construction, asserted because the defer that calls this
	// runs on every exit path including the ones that never claimed.
	s.release(id)
	s.release(id)
	if !s.claim(id) {
		t.Fatal("release must be safe to repeat")
	}
}

// TestClaimUnderRace proves exactly one of many simultaneous clicks drives.
//
// Two clicks 2s apart are serialised by the network anyway; the reason this
// needs a mutex rather than a plain map read is that the handler runs in one
// goroutine per request, and a map written from two of them at once is a
// crash rather than a wrong answer. Run with -race to mean it.
func TestClaimUnderRace(t *testing.T) {
	s := &server{flying: map[string]bool{}}
	var wg sync.WaitGroup
	won := make(chan bool, 64)
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			won <- s.claim("tl-e10-01:edit")
		}()
	}
	wg.Wait()
	close(won)
	drives := 0
	for w := range won {
		if w {
			drives++
		}
	}
	if drives != 1 {
		t.Fatalf("64 simultaneous clicks produced %d drives, want 1", drives)
	}
}
