// Containment must not fold case on a filesystem that does not.
//
// contained() lowercased both sides on Windows and macOS, on the reasoning
// that "folding when the filesystem does not is harmless here". It is not.
// With the share at ...\me\x and a genuinely separate ...\me\X beside it,
// filepath.Rel correctly answers ..\X\secret.txt, and the folding branch then
// overwrites that with secret.txt and calls it contained.
//
// Both defaults really are only defaults: NTFS carries a per-directory case
// sensitivity flag, and `fsutil file setCaseSensitiveInfo` sets it WITHOUT
// elevation. WSL sets it on directories it creates, which is not an exotic
// configuration on a machine anyone would run this on.
//
// So the question is answered by asking the filesystem rather than the OS
// name, and this is the measurement.
//
//	go test ./...
package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// caseSensitiveDir returns a directory whose filesystem does NOT fold case,
// or "" with a reason when this machine cannot make one.
func caseSensitiveDir(t *testing.T) string {
	t.Helper()
	base := t.TempDir()
	if runtime.GOOS == "windows" {
		out, err := exec.Command("fsutil.exe", "file",
			"setCaseSensitiveInfo", base, "enable").CombinedOutput()
		if err != nil {
			t.Skipf("no case-sensitive directory here: %v: %s", err, out)
		}
	}
	// Prove it, rather than trusting the flag we just set. Two names differing
	// only in case must be two directories.
	lower := filepath.Join(base, "casetest")
	upper := filepath.Join(base, "CASETEST")
	if err := os.Mkdir(lower, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(upper, 0o755); err != nil {
		t.Skipf("this filesystem folds case, so the escape cannot exist: %v", err)
	}
	return base
}

func TestContainmentDoesNotFoldCaseOnCaseSensitiveFS(t *testing.T) {
	base := caseSensitiveDir(t)

	// The share, and a SEPARATE folder beside it differing only in case.
	root := filepath.Join(base, "share")
	outside := filepath.Join(base, "SHARE")
	for _, d := range []string{root, outside} {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	secret := filepath.Join(outside, "secret.txt")
	if err := os.WriteFile(secret, []byte("NOT SHARED\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Sanity: these really are two different directories on this filesystem.
	if err := os.WriteFile(filepath.Join(root, "secret.txt"),
		[]byte("shared\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(secret); string(b) != "NOT SHARED\n" {
		t.Skip("the two folders are the same one here; nothing to test")
	}

	t.Logf("root      = %s", root)
	t.Logf("flipped   = %s", flipASCIICase(root))
	if st, err := os.Stat(flipASCIICase(root)); err != nil {
		t.Logf("flipped stat: %v (so probeFold says not folding)", err)
	} else {
		here, _ := os.Stat(root)
		t.Logf("flipped stat: ok, SameFile=%v", os.SameFile(here, st))
	}
	t.Logf("foldsCase = %v  (must be false here)", foldsCase(root))

	got, err := resolve(root, filepath.Join("..", "SHARE", "secret.txt"))
	if err == nil {
		t.Fatalf("containment allowed a path outside the share: %s\n"+
			"root was %s, and SHARE is a different directory on this "+
			"filesystem", got, root)
	}
}
