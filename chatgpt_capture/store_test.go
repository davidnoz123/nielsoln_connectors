package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func put(t *testing.T, s *handoffStore, body string, updateTime float64) {
	t.Helper()
	if err := s.Put(json.RawMessage(body)); err != nil {
		t.Fatal(err)
	}
	s.Observe(updateTime)
}

func TestWatermarkSaysWHENItCannotSay(t *testing.T) {
	out := filepath.Join(t.TempDir(), "c.jsonl")
	s, err := newHandoffStore(out)
	if err != nil {
		t.Fatal(err)
	}
	wm, how := s.Watermark()
	if wm != 0 {
		t.Errorf("a fresh store reported watermark %v, want 0", wm)
	}
	// ⚠️ THE REASON IS THE POINT. A caller that cannot tell "nothing stored
	// yet" from "this store cannot say" would do a full sweep and report it
	// as incremental.
	if !strings.Contains(how, "FULL sweep") {
		t.Errorf("the reason does not warn it is a full sweep: %q", how)
	}
}

func TestStdoutHasNoSidecarAndAdmitsIt(t *testing.T) {
	s, err := newHandoffStore("-")
	if err != nil {
		t.Fatal(err)
	}
	_, how := s.Watermark()
	if !strings.Contains(how, "stdout") || !strings.Contains(how, "FULL sweep") {
		t.Errorf("stdout should say it has nowhere to keep state: %q", how)
	}
	put(t, s, `{"a":1}`, 500)
	if _, err := s.Commit(); err != nil {
		t.Fatal(err)
	}
	// ⓘ Nothing to assert about a file, which is the point: inventing a
	// sidecar location would put a file somewhere the caller did not name.
}

func TestTheSidecarCarriesTheWatermarkForward(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "c.jsonl")

	s, err := newHandoffStore(out)
	if err != nil {
		t.Fatal(err)
	}
	put(t, s, `{"id":"a"}`, 1000)
	put(t, s, `{"id":"b"}`, 2000)
	if n, err := s.Commit(); err != nil || n != 2 {
		t.Fatalf("got %d, %v; want 2, nil", n, err)
	}

	// A second store on the same path must find the watermark.
	s2, err := newHandoffStore(out)
	if err != nil {
		t.Fatal(err)
	}
	wm, how := s2.Watermark()
	if wm != 2000 {
		t.Errorf("got watermark %v, want 2000 (the newest Observed)", wm)
	}
	if !strings.Contains(how, "state.json") {
		t.Errorf("the reason does not name where it came from: %q", how)
	}
}

// TestTheWatermarkNEVERGOESBACKWARDS is the one that would lose data.
func TestTheWatermarkNEVERGOESBACKWARDS(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "c.jsonl")

	s, err := newHandoffStore(out)
	if err != nil {
		t.Fatal(err)
	}
	put(t, s, `{"id":"new"}`, 9000)
	if _, err := s.Commit(); err != nil {
		t.Fatal(err)
	}

	// A later run that happens to capture only OLDER conversations -- a
	// -since 0 sweep, say -- must not rewind the mark.
	s2, err := newHandoffStore(out)
	if err != nil {
		t.Fatal(err)
	}
	put(t, s2, `{"id":"old"}`, 100)
	if _, err := s2.Commit(); err != nil {
		t.Fatal(err)
	}

	s3, err := newHandoffStore(out)
	if err != nil {
		t.Fatal(err)
	}
	if wm, _ := s3.Watermark(); wm != 9000 {
		t.Errorf("⚠️ the watermark REWOUND to %v from 9000. The next run would "+
			"re-fetch everything in between, and on a store that advanced "+
			"wrongly it would SKIP what never landed.", wm)
	}
}

func TestNoSidecarOnAnEmptyRun(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "c.jsonl")
	s, err := newHandoffStore(out)
	if err != nil {
		t.Fatal(err)
	}
	if n, err := s.Commit(); err != nil || n != 0 {
		t.Fatalf("got %d, %v", n, err)
	}
	// ⚠️ Writing a watermark for zero conversations would claim ground
	// nothing covered.
	if _, err := os.Stat(out + ".state.json"); err == nil {
		t.Error("an empty run wrote a sidecar, claiming a watermark it did " +
			"not earn")
	}
}

// TestTheSidecarWarnsTheReaderInTheFILE pins the self-describing property.
func TestTheSidecarWarnsTheReaderInTheFILE(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "c.jsonl")
	s, err := newHandoffStore(out)
	if err != nil {
		t.Fatal(err)
	}
	put(t, s, `{"id":"a"}`, 42)
	if _, err := s.Commit(); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(out + ".state.json")
	if err != nil {
		t.Fatal(err)
	}
	var st handoffState
	if err := json.Unmarshal(raw, &st); err != nil {
		t.Fatal(err)
	}
	// ⚠️ The hazard travels WITH the file. Somebody finding a stale watermark
	// should not have to find the source to understand it.
	for _, want := range []string{"CAPTURED", "INGESTED", "delete this file"} {
		if !strings.Contains(st.Warning, want) {
			t.Errorf("the sidecar's own warning omits %q: %q", want, st.Warning)
		}
	}
	if st.Watermark != 42 || st.Count != 1 {
		t.Errorf("got watermark %v count %d, want 42 and 1", st.Watermark, st.Count)
	}
}

// TestPutDoesNotInterpret is the module's property, now via the store.
func TestPutDoesNotInterpret(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "c.jsonl")
	s, err := newHandoffStore(out)
	if err != nil {
		t.Fatal(err)
	}
	// Non-alphabetical keys and a number a float round-trip would change.
	const odd = `{"zzz":1,"id":"x","create_time":1730000000.123456}`
	put(t, s, odd, 1)
	if _, err := s.Commit(); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(got)) != odd {
		t.Errorf("the store altered the bytes.\n got %s\nwant %s",
			strings.TrimSpace(string(got)), odd)
	}
}
