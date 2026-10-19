package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const wsFor = "ws://127.0.0.1:9656/devtools/browser/"

func TestBrowserIDIsExtracted(t *testing.T) {
	const guid = "7e6c8b2a-a6b7-4b8f-9d69-eb0b1fb145b2"
	if got := browserID(wsFor + guid); got != guid {
		t.Errorf("got %q, want %q", got, guid)
	}
	// ⚠️ A URL with no browser segment must yield "", not a partial match. An
	// empty id can never equal a recorded one, so the failure direction is
	// refusal rather than false verification.
	for _, bad := range []string{
		"ws://127.0.0.1:9656/devtools/page/ABC", "", "not a url",
	} {
		if got := browserID(bad); got != "" {
			t.Errorf("browserID(%q) = %q, want empty", bad, got)
		}
	}
}

func TestVerifiedOnlyWhenGUIDAndPortMatch(t *testing.T) {
	dir := t.TempDir()
	const guid = "aaaa-bbbb"
	if err := writeMarker(dir, marker{Port: 9656, BrowserID: guid, PID: 1}); err != nil {
		t.Fatal(err)
	}

	if v, _ := verify(dir, wsFor+guid, 9656); v != Verified {
		t.Errorf("matching guid and port gave %v, want Verified", v)
	}
	// ⚠️ A DIFFERENT BROWSER ON THE SAME PORT is the case that silently
	// succeeded before R2 existed.
	if v, _ := verify(dir, wsFor+"cccc-dddd", 9656); v != Mismatch {
		t.Errorf("a different browser gave %v, want Mismatch", v)
	}
	// The same browser reported on a different port is also a mismatch: the
	// marker records the pair, and half a match is not a match.
	if v, _ := verify(dir, wsFor+guid, 9222); v != Mismatch {
		t.Errorf("same guid on another port gave %v, want Mismatch", v)
	}
}

func TestNoMarkerIsRefusedNotAssumedGood(t *testing.T) {
	dir := t.TempDir() // no marker written
	v, _ := verify(dir, wsFor+"anything", 9656)
	if v != NoMarker {
		t.Fatalf("an unmarked profile gave %v, want NoMarker", v)
	}
	// ⚠️ THE WHOLE POINT OF R2. Before this, an answering port was accepted
	// whatever was behind it, so `-port 9333 -profile chatgpt` succeeded
	// against chrome_checks_profile. Default to refusal.
	msg := explain(v, dir, wsFor+"anything", 9656, marker{})
	for _, want := range []string{"REFUSED", "did not start", "accept-unverified"} {
		if !strings.Contains(msg, want) {
			t.Errorf("the refusal does not mention %q:\n%s", want, msg)
		}
	}
}

// TestTheRefusalNamesBOTHSides is what makes it actionable.
func TestTheRefusalNamesBOTHSides(t *testing.T) {
	dir := t.TempDir()
	if err := writeMarker(dir, marker{Port: 9222, BrowserID: "recorded-id", PID: 7}); err != nil {
		t.Fatal(err)
	}
	v, m := verify(dir, wsFor+"live-id", 9656)
	msg := explain(v, dir, wsFor+"live-id", 9656, m)
	// "Wrong profile" is useless. The reader must be able to tell which of
	// the two to change.
	for _, want := range []string{"recorded-id", "live-id", "9222", "9656"} {
		if !strings.Contains(msg, want) {
			t.Errorf("the mismatch message omits %q, so a reader cannot tell "+
				"which side is wrong:\n%s", want, msg)
		}
	}
}

func TestACorruptMarkerIsTreatedAsAbsent(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, markerName),
		[]byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	// ⚠️ Absent, not Verified. A marker that cannot be read proves nothing,
	// and the safe reading of "proves nothing" is refusal.
	if v, _ := verify(dir, wsFor+"x", 9656); v != NoMarker {
		t.Errorf("a corrupt marker gave %v, want NoMarker", v)
	}
}

func TestMarkerRoundTrips(t *testing.T) {
	dir := t.TempDir()
	want := marker{Port: 9655, BrowserID: "guid-here", PID: 4242}
	if err := writeMarker(dir, want); err != nil {
		t.Fatal(err)
	}
	got, err := readMarker(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
	// It lives INSIDE the profile, so deleting the profile deletes the claim.
	if _, err := os.Stat(filepath.Join(dir, markerName)); err != nil {
		t.Errorf("the marker is not inside the profile directory: %v", err)
	}
}
