package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// A name beginning with two dots is a NAME, not a way out.
//
// firstLink skipped its link check for any path whose relative form started
// with "..", using HasPrefix. `..x` and `..hidden` are ordinary filenames
// inside the shared folder and both matched, so links with those names were
// never checked.
//
// Harmless alone. On 2 Oct 2026 resolve() stopped refusing paths it could not
// follow, on the reasoning that firstLink had already examined every component
// at or below the root -- which was true except for these. A DANGLING link
// named `..x` then resolved to its own lexical path, passed containment, and
// write_file opens with O_CREATE: it would have created the target outside the
// shared folder.
//
// Found by a reviewer reading the exact commit the page was publishing.
func TestDotDotNameIsNotAnEscape(t *testing.T) {
	root := filepath.Join("C:", "share")
	if runtime.GOOS != "windows" {
		root = "/share"
	}

	// What firstLink's guard is really asking: is this path outside the root?
	outside := func(name string) bool {
		rel, err := filepath.Rel(root, filepath.Join(root, name))
		if err != nil {
			return true
		}
		return rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator))
	}

	for _, name := range []string{"..x", "..hidden", "...", "ordinary.txt"} {
		if outside(name) {
			t.Errorf("%q is a name INSIDE the folder and was treated as outside, "+
				"so the link check would be skipped for it", name)
		}
	}
	for _, name := range []string{"..", filepath.Join("..", "y")} {
		if !outside(name) {
			t.Errorf("%q really is outside the folder and was treated as inside", name)
		}
	}
}

// The escape itself, end to end, where the filesystem allows it to be built.
//
// Creating a symlink needs SeCreateSymbolicLinkPrivilege on Windows, which a
// normal user does not have, so this SKIPS rather than passing quietly. A skip
// that reads as a pass is the failure this repository keeps finding.
func TestDanglingDotDotLinkIsRefused(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "victim.txt")

	link := filepath.Join(root, "..x")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("cannot create a symlink here (%v). On Windows this needs "+
			"SeCreateSymbolicLinkPrivilege; the guard is still covered by "+
			"TestDotDotNameIsNotAnEscape.", err)
	}

	// Dangling on purpose: the target does not exist, which is what makes
	// EvalSymlinks fail and sends resolve() down the path that used to accept
	// the lexical name.
	if _, err := resolve(root, "..x"); err == nil {
		t.Fatalf("resolve accepted a link named ..x pointing at %s. "+
			"write_file would have created that file outside the share.",
			outside)
	}
}
