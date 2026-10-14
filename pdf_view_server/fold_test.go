package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestFoldsCaseCachesOnlyDeterminedAnswers pins the fix propagated from
// remote_ai on 14 Oct 2026. s60.
//
// ⚠️ What is directly testable and what is not, said plainly. The three
// not-cached paths are observable: a missing directory, a directory with no
// letters to flip, and a flipped stat failing for a reason other than
// non-existence. The last one cannot be induced portably -- a permission
// error or a share hiccup is not something a test can arrange -- so this
// covers the two it can reach and the cache's emptiness afterwards, and the
// third is argued in the comment rather than proved here.
func TestFoldsCaseCachesOnlyDeterminedAnswers(t *testing.T) {
	reset := func() {
		foldMu.Lock()
		foldCache = map[string]bool{}
		foldMu.Unlock()
	}
	cached := func(k string) (bool, bool) {
		foldMu.Lock()
		defer foldMu.Unlock()
		v, ok := foldCache[k]
		return v, ok
	}

	// A directory that does not exist: the first stat fails, and nothing is
	// learned, so nothing may be remembered.
	reset()
	missing := filepath.Join(t.TempDir(), "no-such-dir")
	if foldsCase(missing) {
		t.Error("a missing directory reported as folding")
	}
	if _, ok := cached(missing); ok {
		t.Error("a missing directory was CACHED. The next call would answer " +
			"from a moment when the question was unanswerable.")
	}

	// ⚠️ THE "NOTHING TO FLIP" BRANCH CANNOT FIRE ON WINDOWS, and an
	// earlier version of this test asserted that it did.
	//
	// flipCase operates on the WHOLE PATH, not the basename, and every
	// absolute Windows path carries a drive letter. So `flipped == dir` is
	// unreachable here: C:\123 flips to c:\123. The branch is live only
	// where a path can contain no letters at all, which on Unix means
	// something like /123.
	//
	// The property is still worth pinning, so it is pinned on flipCase
	// directly rather than through a foldsCase call that cannot reach it.
	if flipCase("123") != "123" {
		t.Error("flipCase changed a string with no letters")
	}
	if flipCase("C:\123") == "C:\123" {
		t.Error("flipCase left a drive letter alone, so the unreachability " +
			"argument above is wrong and foldsCase CAN take that branch")
	}

	// A real directory with letters: whatever the answer, it IS determined,
	// so it must be remembered.
	reset()
	real := filepath.Join(t.TempDir(), "Letters")
	if err := os.Mkdir(real, 0o755); err != nil {
		t.Fatal(err)
	}
	got := foldsCase(real)
	if _, ok := cached(real); !ok {
		t.Errorf("a determined answer (%v) was not cached, so the stat pair "+
			"runs on every containment check", got)
	}
	t.Logf("this filesystem folds case for %s: %v", real, got)
}
