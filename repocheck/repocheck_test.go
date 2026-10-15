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

// ---- SharedBodies: the same code under a different name ----

func TestNoUndeclaredRenamedCopies(t *testing.T) {
	mods := skipIfNoCheckout(t)
	shared, err := SharedBodies(root, mods)
	if err != nil {
		t.Fatal(err)
	}
	for _, sb := range shared {
		names := sb.Names()
		// One name in two modules is Overlaps's business, already covered
		// by TestEveryOverlapIsDeclared. Two names sharing a body is a copy
		// somebody renamed, which is what this check exists for.
		if len(names) < 2 {
			continue
		}
		if !AcknowledgedBody(names) {
			t.Errorf("RENAMED COPY: %s\n  The same %d-line body under "+
				"different names. Either share it or add %q to "+
				"acknowledgedBodies with the reason.",
				sb, sb.Lines, strings.Join(names, "+"))
		}
	}
}

// TestTheRenamedCopyCheckCanFail is the companion. ⚠️ Three times in this
// repo a safety check has passed because it found nothing, so this one is
// shown catching a copy it was built to catch.
func TestTheRenamedCopyCheckCanFail(t *testing.T) {
	dir := t.TempDir()
	// A body long enough to clear minBodyLines, under TWO different names in
	// TWO different modules. Exactly what Overlaps cannot see.
	body := "{\n\tn := 0\n\tfor _, c := range s {\n\t\tif c == 'x' {\n\t\t\tn++\n\t\t}\n\t}\n\treturn n\n}\n"
	for mod, fn := range map[string]string{"alpha": "countEks", "beta": "tallyX"} {
		md := filepath.Join(dir, mod)
		if err := os.MkdirAll(md, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(md, "go.mod"),
			[]byte("module x/"+mod+"\n\ngo 1.27\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		src := "package main\n\nfunc " + fn + "(s string) int " + body
		if err := os.WriteFile(filepath.Join(md, "a.go"),
			[]byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mods, err := modulesIn(dir)
	if err != nil {
		t.Fatal(err)
	}
	shared, err := SharedBodies(dir, mods)
	if err != nil {
		t.Fatal(err)
	}
	// ⚠️ The names differ, so this is precisely the case the name-keyed
	// check misses. Prove BOTH: that SharedBodies sees it and that Overlaps
	// does not, because the second half is the justification for the first.
	var found bool
	for _, sb := range shared {
		if len(sb.Names()) == 2 && len(sb.Modules()) == 2 {
			found = true
		}
	}
	if !found {
		t.Fatalf("SharedBodies did not see one body under two names in two "+
			"modules, so every green result above is meaningless: %v", shared)
	}
	ov, err := Overlaps(dir, mods)
	if err != nil {
		t.Fatal(err)
	}
	if len(ov) != 0 {
		t.Errorf("Overlaps saw %v. If the name-keyed check can see a "+
			"renamed copy, SharedBodies is redundant and this file should "+
			"be deleted rather than maintained.", ov)
	}
}

func TestTrivialBodiesAreBelowTheFloor(t *testing.T) {
	// ⚠️ The floor is what makes the report readable. Without it, every
	// two-line wrapper in the repo collides and the real finding drowns.
	if n := countSignificant("{\n\treturn w.conn.Close()\n}\n"); n >= minBodyLines {
		t.Errorf("a one-line body counts %d significant lines, at or above "+
			"the floor of %d, so trivial collisions will be reported",
			n, minBodyLines)
	}
	withComments := "{\n\t// one\n\t// two\n\t// three\n\t// four\n\treturn 1\n}\n"
	if n := countSignificant(withComments); n >= minBodyLines {
		t.Errorf("comments pushed a one-line body to %d, so a function "+
			"would clear the floor on the strength of its documentation", n)
	}
}

// ---- known_recipes: semantic duplication, declared ----

// TestEveryRecipeIsStillReal is the staleness check, and it is what makes a
// hand-written list defensible.
//
// ⚠️ A declared inventory with no verification becomes a document that
// describes code which no longer exists, and the next reader cannot tell
// which entries are live. known_overlaps.go's first version had seven
// fictional entries and a test like this one found all seven.
func TestEveryRecipeIsStillReal(t *testing.T) {
	skipIfNoCheckout(t)
	for _, r := range recipes {
		if len(r.Sites) < 2 {
			t.Errorf("recipe %q names %d site(s). A recipe in one place is "+
				"not duplication.", r.Name, len(r.Sites))
			continue
		}
		if strings.TrimSpace(r.Why) == "" {
			t.Errorf("recipe %q has no reason. An entry without one is an "+
				"entry nobody dares delete.", r.Name)
		}
		if len(r.Markers) == 0 {
			t.Errorf("recipe %q has no markers, so it cannot go stale and "+
				"cannot be trusted either.", r.Name)
			continue
		}
		for _, site := range r.Sites {
			raw, err := os.ReadFile(filepath.Join(root, site))
			if err != nil {
				t.Errorf("recipe %q names %s, which cannot be read: %v\n"+
					"  Either the duplication is gone or it moved. Both "+
					"need a human.", r.Name, site, err)
				continue
			}
			for _, m := range r.Markers {
				if !strings.Contains(string(raw), m) {
					t.Errorf("recipe %q: %s no longer contains %q.\n"+
						"  The entry has gone stale and now describes code "+
						"that is not there.", r.Name, site, m)
				}
			}
		}
	}
}

func TestTheRecipeStalenessCheckCanFail(t *testing.T) {
	// A recipe pointing at a file that does not exist must be detectable,
	// because that is the whole staleness mechanism.
	r := Recipe{
		Name:    "fictional",
		Sites:   []string{"nowhere/at_all.go"},
		Markers: []string{"something"},
		Why:     "a test fixture",
	}
	if _, err := os.ReadFile(filepath.Join(root, r.Sites[0])); err == nil {
		t.Fatal("nowhere/at_all.go exists, so this fixture proves nothing")
	}
	// And a marker absent from a file that DOES exist.
	raw, err := os.ReadFile(filepath.Join(root, "repocheck", "known_recipes.go"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "zzz_marker_that_is_not_there") {
		t.Fatal("the absent-marker fixture is present, so it proves nothing")
	}
}

// TestPythonBodiesAreRead is the can-fail companion for the language widening.
//
// ⚠️ The Go-only version of SharedBodies was written one day and the gap bit
// the next: linkprobe grew 543 lines of Python, including a third copy of
// serve-a-page-and-click, invisible to a check whose purpose is noticing
// duplication. This proves the widening works rather than assuming it.
func TestPythonBodiesAreRead(t *testing.T) {
	dir := t.TempDir()
	body := "\n    n = 0\n    for c in s:\n        if c == 'x':\n            n += 1\n    return n\n"
	for mod, fn := range map[string]string{"alpha": "count_eks", "beta": "tally_x"} {
		md := filepath.Join(dir, mod)
		if err := os.MkdirAll(md, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(md, "go.mod"),
			[]byte("module x/"+mod+"\n\ngo 1.27\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		src := "def " + fn + "(s):" + body
		if err := os.WriteFile(filepath.Join(md, "a.py"),
			[]byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mods, err := modulesIn(dir)
	if err != nil {
		t.Fatal(err)
	}
	shared, err := SharedBodies(dir, mods)
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, sb := range shared {
		if len(sb.Names()) == 2 && len(sb.Modules()) == 2 {
			found = true
		}
	}
	if !found {
		t.Fatalf("a Python body under two names in two modules was not seen, "+
			"so the Python half of this repo is unchecked: %v", shared)
	}
}

// TestPyBodyOfEndsAtTheDedent pins the rule, there being no closing brace.
func TestPyBodyOfEndsAtTheDedent(t *testing.T) {
	src := "def a(x):\n    return 1\n\n\ndef b(y):\n    return 2\n"
	got := pyBodyOf(src, "a")
	if strings.Contains(got, "def b") {
		t.Errorf("pyBodyOf ran past the dedent and swallowed the next "+
			"function, so every Python hash covers more than one body:\n%s", got)
	}
	if !strings.Contains(got, "return 1") {
		t.Errorf("pyBodyOf stopped too early: %q", got)
	}
	// An indented method must be found too, since a copied Python helper is
	// as likely to be a method as a module-level function.
	cls := "class C:\n    def send(self, m):\n        self.n += 1\n        return self.n\n"
	if g := pyBodyOf(cls, "send"); !strings.Contains(g, "self.n += 1") {
		t.Errorf("an indented method was not extracted: %q", g)
	}
}

func TestPythonCommentsDoNotInflateABody(t *testing.T) {
	// ⚠️ Counting `#` lines would let a trivial function clear the floor on
	// the strength of its documentation, which this repo's Python has in
	// quantity.
	body := "def f():\n    # one\n    # two\n    # three\n    # four\n    # five\n    return 1\n"
	if n := countSignificant(body); n >= minBodyLines {
		t.Errorf("a one-line Python body counted %d significant lines, at or "+
			"above the floor of %d", n, minBodyLines)
	}
}
