// repocheck_test.go -- the checks, run as a test so they are in the loop.
//
// A program nobody runs is a program nobody runs. `AGENTS.project.md` puts
// `go vet ./... && go test ./...` in every connector's loop, so the checks
// live in a test and the `main` above is for reading the report by hand.
//
// ⚠️ Every check here has a companion proving it CAN fail. Three times in this
// repo a safety check has passed because it found nothing: the caveats walk
// missed its package's primary reach, then matched per-file instead of
// per-reach, then could not see inside an import. A check that is green
// because it is blind is worse than no check, so each one is shown failing on
// a synthetic case.
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const root = ".."

func skipIfNoCheckout(t *testing.T) []string {
	t.Helper()
	mods, err := modulesIn(root)
	if err != nil || len(mods) == 0 {
		t.Skipf("UNANSWERED: no sibling modules under %s, so the repo-wide "+
			"checks cannot run. A module fetched by `go run module@sha` has "+
			"no siblings.", root)
	}
	return mods
}

func TestProvenanceEveryRequiredModuleIsOurs(t *testing.T) {
	mods := skipIfNoCheckout(t)
	sums, err := SumEntries(root, mods)
	if err != nil {
		t.Fatal(err)
	}
	if len(sums) == 0 {
		t.Log("nothing is required anywhere yet, so this passes trivially. " +
			"TestTheProvenanceCheckCanFail is why that means something.")
	}
	for path, by := range sums {
		if !strings.HasPrefix(path, repoPrefix) {
			t.Errorf("FOREIGN MODULE %s, required by %s. s55 maxim 1: every "+
				"module in the closure lives in this repo.",
				path, strings.Join(by, " "))
		}
	}
}

func TestEveryOverlapIsDeclared(t *testing.T) {
	mods := skipIfNoCheckout(t)
	overlaps, err := Overlaps(root, mods)
	if err != nil {
		t.Fatal(err)
	}
	for _, o := range overlaps {
		if !Acknowledged(o.Name) {
			t.Errorf("NEW OVERLAP: %s\n  Not a defect, but undeclared. "+
				"Either share the code or add it to known_overlaps.go with "+
				"the reason it is tolerated.", o)
			continue
		}
		t.Logf("declared: %s", o)
	}
}

// TestBaselineHasNoStaleEntries keeps the allowlist honest in BOTH directions.
//
// ⚠️ An allowlist that accumulates entries for things that stopped being
// overlaps is clutter that implies duplication where there is none, and the
// next reader cannot tell which entries are live. The first version of
// known_overlaps.go had seven such entries, listed from a grep that had not
// distinguished a cross-connector overlap from a build-tag pair inside one
// module. This test found them.
func TestBaselineHasNoStaleEntries(t *testing.T) {
	mods := skipIfNoCheckout(t)
	overlaps, err := Overlaps(root, mods)
	if err != nil {
		t.Fatal(err)
	}
	live := map[string]bool{}
	for _, o := range overlaps {
		live[o.Name] = true
	}
	for name := range acknowledged {
		if !live[name] {
			t.Errorf("STALE BASELINE ENTRY: %q is acknowledged in "+
				"known_overlaps.go and is not an overlap. It implies "+
				"duplication that does not exist.", name)
		}
	}
}

func TestTheOverlapCheckCanFail(t *testing.T) {
	dir := t.TempDir()
	for _, m := range []string{"alpha", "beta"} {
		md := filepath.Join(dir, m)
		if err := os.MkdirAll(md, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(md, "go.mod"),
			[]byte("module x/"+m+"\n\ngo 1.27\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(md, "a.go"),
			[]byte("package main\n\nfunc sharedThing() {\n}\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mods, err := modulesIn(dir)
	if err != nil {
		t.Fatal(err)
	}
	overlaps, err := Overlaps(dir, mods)
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, o := range overlaps {
		if o.Name == "sharedThing" && len(o.In) == 2 {
			found = true
			if !o.Identical() {
				t.Error("two identical bodies reported as DIVERGED")
			}
		}
	}
	if !found {
		t.Fatalf("the overlap check did not see a function present in two "+
			"modules, so every green result above is meaningless: %v", overlaps)
	}
}

func TestTheProvenanceCheckCanFail(t *testing.T) {
	dir := t.TempDir()
	md := filepath.Join(dir, "alpha")
	if err := os.MkdirAll(md, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(md, "go.mod"),
		[]byte("module x/alpha\n\ngo 1.27\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(md, "go.sum"),
		[]byte("github.com/someone/else v1.2.3 h1:abc=\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mods, err := modulesIn(dir)
	if err != nil {
		t.Fatal(err)
	}
	sums, err := SumEntries(dir, mods)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := sums["github.com/someone/else"]; !ok {
		t.Fatalf("the provenance check did not see a foreign module in a "+
			"go.sum, so it would never catch one: %v", sums)
	}
	t.Log("the check can see a foreign module, so a green result above " +
		"means something")
}

// TestDivergenceIsDistinguishedFromIdentity pins the thing the report exists
// to show: `contained` is byte-identical in two connectors while `foldsCase`,
// which it calls, is not. That pair is why this program was written.
func TestDivergenceIsDistinguishedFromIdentity(t *testing.T) {
	mods := skipIfNoCheckout(t)
	overlaps, err := Overlaps(root, mods)
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]Overlap{}
	for _, o := range overlaps {
		byName[o.Name] = o
	}
	if o, ok := byName["contained"]; ok && !o.Identical() {
		t.Errorf("`contained` is no longer identical across connectors: %s\n"+
			"  s60 records it as identical. One copy has changed, which is "+
			"either the fix or a new divergence, and either way the row is "+
			"now wrong.", o)
	}
	if o, ok := byName["foldsCase"]; ok && o.Identical() {
		t.Logf("`foldsCase` is now identical: %s\n  s60's divergence appears "+
			"to have been resolved. Update the row.", o)
	}
}
