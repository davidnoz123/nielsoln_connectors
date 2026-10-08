package main

import "testing"

// A seat name reaches a URL path, so what it may contain is the question.
//
// -seat is how a connector says WHICH bridge it wants, before the socket is
// open and before anything is authenticated. That makes it routing
// information rather than a credential, which is why it may be guessable.
// It also makes it a string a proxy acts on, which is why `../bridge` has to
// be refused at the edge rather than hoped about later.
func TestSeatNameRefusesAnythingThatCouldRoute(t *testing.T) {
	for _, name := range []string{
		"seat1", "seat12", "alice", "a-b", "a1",
	} {
		if !seatName.MatchString(name) {
			t.Errorf("%q is an ordinary seat name and was refused", name)
		}
	}
	for _, name := range []string{
		"../bridge",   // the whole reason this is checked at all
		"a/b",         // a path of its own
		"a..b",        // dot segments, which a proxy may normalise
		".",           // "
		"",            // the empty case is handled before the regexp
		"A",           // a Linux account may not have it, nor may this
		"1seat",       // useradd refuses a leading digit
		"a_b",         // not in the manifest's charset either
		"a b",         // a space ends an argument somewhere
		"a?b",         // a query string
		"a%2fb",       // an encoded slash, decoded by somebody downstream
		"ことり",         // not ASCII
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", // 40, over the limit
	} {
		if seatName.MatchString(name) {
			t.Errorf("%q would reach a URL path and was accepted", name)
		}
	}
}

// The path is built from the seat, and an empty seat is the single-seat
// server as it stands today: /bridge and nothing appended.
func TestPathIsPerSeatAndUnchangedWithoutOne(t *testing.T) {
	if got := (&connector{}).path(); got != "/bridge" {
		t.Errorf("no seat should dial /bridge, got %q", got)
	}
	if got := (&connector{seat: "seat2"}).path(); got != "/bridge/seat2" {
		t.Errorf("seat2 should dial /bridge/seat2, got %q", got)
	}
}
