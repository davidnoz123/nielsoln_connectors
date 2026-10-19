package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestABareNameGoesSomewhereDurable is the defect this file fixed.
func TestABareNameGoesSomewhereDurable(t *testing.T) {
	got, err := profileDir("default")
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(got) {
		t.Errorf("got %q, which is not absolute", got)
	}
	// ⚠️ THE WHOLE POINT. The previous version returned
	// os.TempDir()/chrome_debug_profile, so a sign-in survived only until the
	// next disk cleanup -- and the symptom was a capture finding an empty
	// account, not an error.
	tmp, _ := filepath.Abs(os.TempDir())
	if rel, err := filepath.Rel(tmp, got); err == nil &&
		!strings.HasPrefix(rel, "..") {
		t.Errorf("⚠️ %q is INSIDE the temp directory %q. A profile there "+
			"loses its sign-in to a disk cleanup.", got, tmp)
	}
	if !strings.Contains(got, "chrome_ready") {
		t.Errorf("got %q, which does not namespace by the connector", got)
	}
}

// TestAPathIsUsedAsGiven is AGENTS' "declared, never allocated".
//
// ⚠️ THE PATH IS DERIVED FROM UserConfigDir, NOT FROM t.TempDir(). The first
// version of this test used filepath.Join(t.TempDir(), "..", "x") on the
// assumption that `..` escaped the temp tree. It does not: t.TempDir() is
// itself inside the system temp directory, so the parent is still in it and
// refuseTemp refused the path correctly. The test was wrong, not the code.
func TestAPathIsUsedAsGiven(t *testing.T) {
	base, err := os.UserConfigDir()
	if err != nil {
		t.Skipf("UNANSWERED: no user config directory on this machine, so a "+
			"non-temp path could not be constructed: %v", err)
	}
	want := filepath.Join(base, "chrome_ready_test_declared")
	got, err := profileDir(want)
	if err != nil {
		t.Fatalf("a caller-declared absolute path was refused: %v", err)
	}
	if got != want {
		t.Errorf("got %q, want the declared path %q.\n⚠️ A connector that "+
			"relocates a directory the caller named is allocating, not "+
			"accepting a declaration.", got, want)
	}
}

// TestTempIsRefusedNotWarnedAbout pins the refusal.
func TestTempIsRefusedNotWarnedAbout(t *testing.T) {
	inTemp := filepath.Join(os.TempDir(), "chrome_debug_profile")
	_, err := profileDir(inTemp)
	if err == nil {
		t.Fatal("⚠️ a profile path inside the temp directory was ACCEPTED. " +
			"That is the exact defect this file exists to prevent, and it " +
			"fails silently weeks later.")
	}
	for _, want := range []string{"REFUSED", "temp directory", "empty account"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not mention %q, so a caller cannot "+
				"tell why: %v", want, err)
		}
	}
}

// TestTheDefaultFlagValueIsAccepted guards the two halves agreeing.
//
// ⚠️ The default used to be "chrome_debug_profile" resolved under temp. If
// the flag default and profileDir ever disagree, the failure is that every
// run with no -profile is refused, which would be loud -- but this is cheap.
func TestTheDefaultFlagValueIsAccepted(t *testing.T) {
	if _, err := profileDir("default"); err != nil {
		t.Errorf("the flag's default value is refused by profileDir: %v", err)
	}
}

func TestAnEmptyProfileIsRefusedWithAReason(t *testing.T) {
	_, err := profileDir("")
	if err == nil {
		t.Fatal("an empty -profile was accepted")
	}
	if !strings.Contains(err.Error(), "-profile") {
		t.Errorf("the error does not name the flag: %v", err)
	}
}

// TestDirExistsDistinguishesFileFromDirectory matters because the result
// decides whether the run tells the user to sign in.
func TestDirExistsDistinguishesFileFromDirectory(t *testing.T) {
	dir := t.TempDir()
	if !dirExists(dir) {
		t.Error("an existing directory reported as absent")
	}
	f := filepath.Join(dir, "afile")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if dirExists(f) {
		t.Error("a FILE reported as an existing directory. The fresh-profile " +
			"message keys on this, so it would stay silent for a profile " +
			"that was never created.")
	}
	if dirExists(filepath.Join(dir, "nope")) {
		t.Error("a missing path reported as existing")
	}
}
