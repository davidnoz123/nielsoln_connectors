package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "tasks.json")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// The bug this file exists for.
//
// Lookup upper-cased the id before looking it up. That was invisible while
// the keys were task ids like "E11" and broke every single one the moment
// they became row ids like "tl-e3-02": the server listed fifteen tasks in
// its startup log and then answered "unknown task" for all fifteen. It took
// an end-to-end smoke test through Excel, a browser and Chrome to find
// something a table test finds in a second.
func TestLookupIsCaseInsensitiveBothWays(t *testing.T) {
	path := write(t, `{"tasks":{
		"tl-e3-02":{"page":"https://example.test/","find":"x"},
		"E11":{"page":"https://example.test/","find":"y"}}}`)
	m, err := LoadManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{
		"tl-e3-02", "TL-E3-02", "Tl-E3-02", " tl-e3-02 ",
		"E11", "e11", "  e11",
	} {
		if _, ok := m.Lookup(id); !ok {
			t.Errorf("Lookup(%q) said no; the manifest holds it under a "+
				"different case and a link is not going to match the file's "+
				"spelling by luck", id)
		}
	}
	if _, ok := m.Lookup("tl-e3-03"); ok {
		t.Error("Lookup invented a task that is not in the manifest")
	}
}

// A manifest that lists nothing would serve nothing while looking perfectly
// healthy in the log, which is the failure mode this whole program is
// written against.
func TestEmptyManifestIsRefused(t *testing.T) {
	for _, body := range []string{`{}`, `{"tasks":{}}`} {
		if _, err := LoadManifest(write(t, body)); err == nil {
			t.Errorf("%s was accepted; a server with no tasks refuses every "+
				"request and says nothing about why", body)
		}
	}
}

// A page is the one field with nothing sensible to fall back on. Catching it
// at startup rather than at request time means it is found while somebody is
// editing, not halfway through a call.
func TestTaskWithNoPageIsRefusedAtStartup(t *testing.T) {
	_, err := LoadManifest(write(t, `{"tasks":{
		"ok":{"page":"https://example.test/","find":"x"},
		"bad":{"page":"  ","find":"y"}}}`))
	if err == nil {
		t.Fatal("a task with a blank page was accepted")
	}
	if !strings.Contains(err.Error(), "bad") {
		t.Errorf("the error does not name the offending task: %v", err)
	}
}

func TestBadJSONIsRefusedWithoutPanicking(t *testing.T) {
	if _, err := LoadManifest(write(t, `{"tasks":`)); err == nil {
		t.Error("truncated JSON was accepted")
	}
	if _, err := LoadManifest(filepath.Join(t.TempDir(), "absent.json")); err == nil {
		t.Error("a missing manifest was accepted")
	}
}

func TestIDsAreSortedSoTheStartupLogIsReadable(t *testing.T) {
	m, err := LoadManifest(write(t, `{"tasks":{
		"tl-e7-11":{"page":"https://example.test/"},
		"tl-e3-02":{"page":"https://example.test/"},
		"tl-e11-02":{"page":"https://example.test/"}}}`))
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Join(m.IDs(), ",")
	want := "tl-e11-02,tl-e3-02,tl-e7-11"
	if got != want {
		t.Errorf("IDs() = %q, want %q", got, want)
	}
}
