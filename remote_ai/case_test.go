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

// An empty share cannot be asked whether it folds, and the answer must not be
// remembered. The lesson tells participants to start in a NEW EMPTY FOLDER, so
// this is the ordinary case rather than a corner: a wrong answer cached at
// startup would be wrong for the whole session.
//
// This also stands in for a thing no test can assert directly. The probe used
// to fall back to stat-ing a case-flipped spelling of the share's OWN name,
// which is a sibling of the share and therefore outside it. There is no seam
// to observe that call through, so what is pinned here is the behaviour that
// replaced it.
func TestAnEmptyShareIsNotAnsweredFromMemory(t *testing.T) {
	root := t.TempDir()

	if foldsCase(root) {
		t.Fatal("an empty share cannot have been determined to fold")
	}

	// Now it can be asked. If the empty answer had been cached, this would
	// still say false on a filesystem that plainly folds.
	if err := os.WriteFile(filepath.Join(root, "Report.txt"),
		[]byte("hello\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !foldsCase(root) {
		if _, err := os.Stat(filepath.Join(root, "REPORT.TXT")); err == nil {
			t.Fatal("REPORT.TXT and Report.txt are the same file here, so " +
				"this filesystem folds and the empty answer was cached")
		}
		t.Skip("this filesystem does not fold, so there is nothing to check")
	}
}

// The other direction, which is the one that breaks the product rather than
// the security claim: on an ordinary folding filesystem the share must still
// accept a path spelled with different capitals. Without this, a fix for the
// escape above could disable folding everywhere and refuse paths participants
// can plainly see in their own folder.
func TestContainmentStillFoldsOnAFoldingFS(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "Report.txt"),
		[]byte("hello\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !foldsCase(root) {
		t.Skipf("%s does not fold case, so there is nothing to check", root)
	}
	if _, err := resolve(root, "REPORT.TXT"); err != nil {
		t.Fatalf("a folding filesystem refused a path inside the share: %v", err)
	}
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

	if foldsCase(root) {
		t.Fatalf("foldsCase said this filesystem folds, but %s and %s are "+
			"two different directories on it", root, outside)
	}

	got, err := resolve(root, filepath.Join("..", "SHARE", "secret.txt"))
	if err == nil {
		t.Fatalf("containment allowed a path outside the share: %s\n"+
			"root was %s, and SHARE is a different directory on this "+
			"filesystem", got, root)
	}
}
