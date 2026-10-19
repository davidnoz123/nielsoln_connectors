//go:build darwin

// The macOS boundary, checked on a Mac rather than reasoned about on Windows.
//
//	go test ./...
//
// This is the probe that was run by hand on 20 Oct 2026 before any of
// cage_darwin.go was written, turned into something that runs on every
// build. The hand run is in the file comment there; keeping it only as a
// comment would mean the next macOS release silently removes the boundary
// and nothing says so.
//
// THE CONTROL IS THE POINT, as it was in the probe. A profile that denied
// everything would pass a test that only looked for denials, and so would a
// filename that does not exist. Every denial check here is paired with
// something that must SUCCEED.
package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSandboxExecIsOnThisMac(t *testing.T) {
	// If this ever fails the connector is not broken: cageSupported() answers
	// false, applyCage refuses, exec is not offered and files still work.
	// It fails loudly anyway, because that is a real change in what a
	// participant on a Mac can do.
	if !cageSupported() {
		t.Fatalf("%s is not on this machine, so no command can be confined",
			sandboxExec)
	}
}

func TestTheProfileRefusesTheHomeDirectoryAndAllowsTheWorkspace(t *testing.T) {
	ws := t.TempDir()
	profile, err := profileFor(ws)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(filepath.Dir(profile))

	inside := filepath.Join(ws, "inside.txt")
	if err := os.WriteFile(inside, []byte("secret-inside\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(os.Getenv("HOME"), ".nielsoln-cage-test.txt")
	if err := os.WriteFile(outside, []byte("secret-outside\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(outside)

	// 1. The shared folder must stay usable. A boundary participants have to
	//    abandon is not a boundary.
	if err := exec.Command(sandboxExec, "-f", profile,
		"/bin/cat", inside).Run(); err != nil {
		t.Errorf("the profile refused the shared folder itself: %v", err)
	}
	// 2. The participant's own files must be refused. This is the promise.
	if err := exec.Command(sandboxExec, "-f", profile,
		"/bin/cat", outside).Run(); err == nil {
		t.Error("a sandboxed command read a file in the home directory, so " +
			"the boundary is not holding")
	}
	// 3. THE CONTROL. Without the sandbox the same read must succeed, or
	//    check 2 proved only that the file was missing or /bin/cat was.
	if err := exec.Command("/bin/cat", outside).Run(); err != nil {
		t.Fatalf("the control failed (%v), so check 2 measured nothing", err)
	}
}

func TestTheProfileIsNotWritableByWhatItConfines(t *testing.T) {
	// A profile kept in the shared folder could be rewritten by the very
	// commands it confines, and the next command would run under rules a
	// previous one chose. So it must land outside the workspace.
	ws := t.TempDir()
	profile, err := profileFor(ws)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(filepath.Dir(profile))

	abs, err := filepath.Abs(ws)
	if err != nil {
		t.Fatal(err)
	}
	rel, err := filepath.Rel(abs, profile)
	if err != nil {
		// Different volumes: it is certainly not inside the workspace.
		return
	}
	if rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		t.Errorf("the profile is inside the workspace at %s, where a confined "+
			"command could rewrite it", profile)
	}
}

func TestExecIsShutUntilConfinementHasBeenPROVEN(t *testing.T) {
	// childrenConfinable() must be the RESULT of a measurement, never a
	// statement that macOS has a sandbox. applyCage has not run in this test
	// process, so nothing has been proven and the gate must be shut, on a
	// machine where sandbox-exec plainly does exist.
	if childrenConfinable() {
		t.Fatal("exec was offered without applyCage having proven anything, " +
			"so the gate is trusting the platform instead of measuring it")
	}
	if cageWrap() != nil {
		t.Error("cageWrap handed back a prefix before the profile was proven")
	}
}

func TestApplyCageProvesTheBoundaryAndThenOpensTheGate(t *testing.T) {
	ws := t.TempDir()
	if _, relaunched, err := applyCage(ws); err != nil {
		t.Fatalf("applyCage could not establish a boundary: %v", err)
	} else if relaunched {
		t.Fatal("applyCage asked for a relaunch, which is the Windows route; " +
			"macOS wraps each child instead and this process carries on")
	}
	if !childrenConfinable() {
		t.Fatal("applyCage returned success and the gate is still shut")
	}
	// inCage() answers a DIFFERENT question and must stay false: this
	// process's own reads are checked by resolve() in our code, not by the
	// kernel, and main.go uses this to decide what the participant is told.
	if inCage() {
		t.Error("inCage() says this process is kernel-confined on macOS, " +
			"which would put an untrue sentence on the one line a " +
			"participant is asked to look for")
	}
	w := cageWrap()
	if len(w) != 3 || w[0] != sandboxExec || w[1] != "-f" {
		t.Fatalf("the wrapper is not a sandbox-exec invocation: %v", w)
	}
}

func TestTheProofAndTheProfileNameTheSameHome(t *testing.T) {
	// MEASURED THE HARD WAY, 20 Oct 2026. The profile denied
	// os.UserHomeDir() and the proof read os.Getenv("HOME"). They agree on a
	// normal Mac, so nothing was wrong, and nothing made them agree either: a
	// proof that reads a path the profile never mentions is refused for the
	// wrong reason and opens the gate anyway.
	home, err := deniedHome()
	if err != nil {
		t.Fatal(err)
	}
	if home == "" || home == "/" {
		t.Fatalf("the denied home is %q, which would deny nothing or "+
			"everything", home)
	}
	ws := t.TempDir()
	profile, err := profileFor(ws)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(filepath.Dir(profile))
	body, err := os.ReadFile(profile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), home) {
		t.Errorf("the profile does not mention %s, so proving a refusal "+
			"there proves nothing about it: %s", home, string(body))
	}
}
