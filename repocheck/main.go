// repocheck -- the two checks that cannot be made from inside one module.
//
//	go run github.com/davidnoz123/nielsoln_connectors/repocheck@<sha>
//
// # WHY IT EXISTS
//
// s55 settles that code may be shared across folders at independent commits,
// under two maxims: every module in the closure lives in this repo, and
// nothing stands between the declaration and the thing. ⚠️ Neither is
// checkable from inside a module, because both are statements about the WHOLE
// repo. A per-package test cannot see its siblings.
//
// And s60 records why the second check exists, which is the more interesting
// reason. The leaf candidates were reasoned about for a whole session, and
// when the duplication was finally MEASURED it was somewhere nobody had
// looked: `contained`, the path-containment check, byte-identical in two
// connectors, calling a `foldsCase` that had diverged, with a learned lesson
// stranded in one copy. Reasoning found cdpcore; measurement found path
// safety. So the measurement runs rather than the reasoning repeating.
//
// # WHAT IT CHECKS
//
//	provenance   every module named in any go.sum lives in this repo (s55)
//	overlap      no function name appears in two connectors unless the
//	             overlap is declared and acknowledged in known_overlaps.go
//
// # WHAT OVERLAP IS NOT
//
// ⚠️ Overlap is not a defect. `main` is in every connector, and two functions
// can share a name and do different jobs: `resolve` is 12 lines in one place
// and 92 in another. So this does not forbid duplication. It forbids
// UNNOTICED duplication, by failing on anything not already written down.
// The baseline is the point: a new overlap stops the build until somebody has
// looked at it and either shared the code or recorded why not.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// repoPrefix is s55's maxim 1: provenance, not shape.
const repoPrefix = "github.com/davidnoz123/nielsoln_connectors/"

var funcRe = regexp.MustCompile(`(?m)^func ([a-zA-Z_][a-zA-Z0-9_]*)\(`)

// modulesIn finds the sibling folders that are Go modules.
//
// Returns nothing when run outside a checkout, which is correct rather than a
// failure: a module fetched by `go run module@sha` has no siblings, and a
// check that cannot see the repo should say so rather than pass.
// modulesIn finds every code directory worth checking.
//
// ⚠️ IT USED TO MEAN "DIRECTORY WITH A go.mod", and that left two holes which
// appeared the moment `probes/` was added on 16 Oct 2026:
//
//  1. a PYTHON-ONLY directory has no go.mod, so none of it was discovered.
//     This is the third time the same blind spot has been found in two days:
//     SharedBodies could not see test files, then could not see Python, and
//     now the discovery step could not see a directory holding only Python.
//  2. a NESTED module one level down was missed, because the scan read a
//     single level. `probes/focus` is exactly that.
//
// So a directory counts if it holds a go.mod OR any .py file, and immediate
// subdirectories are checked for go.mod too. Names may therefore contain a
// separator, which the callers handle because they join rather than
// concatenate.
func modulesIn(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if isCodeDir(filepath.Join(root, e.Name())) {
			out = append(out, e.Name())
		}
		// One level deeper, for a module nested inside a code directory.
		// ⚠️ Deliberately NOT a full walk: a recursive scan would pick up
		// vendored trees and build output, and the repo's layout is two
		// levels by convention.
		subs, err := os.ReadDir(filepath.Join(root, e.Name()))
		if err != nil {
			continue
		}
		for _, se := range subs {
			if !se.IsDir() {
				continue
			}
			p := filepath.Join(root, e.Name(), se.Name())
			if _, err := os.Stat(filepath.Join(p, "go.mod")); err == nil {
				out = append(out, filepath.ToSlash(
					filepath.Join(e.Name(), se.Name())))
			}
		}
	}
	sort.Strings(out)
	return out, nil
}

// isCodeDir reports whether a directory holds code this repo should check.
func isCodeDir(dir string) bool {
	if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
		return true
	}
	py, err := filepath.Glob(filepath.Join(dir, "*.py"))
	return err == nil && len(py) > 0
}

// Overlap is one function name found in more than one module.
type Overlap struct {
	Name string
	// In maps module name to a short hash of that function's body, so
	// identical copies and diverged ones are distinguishable at a glance.
	In map[string]string
}

// Identical reports whether every copy hashes the same.
func (o Overlap) Identical() bool {
	var first string
	for _, h := range o.In {
		if first == "" {
			first = h
		} else if h != first {
			return false
		}
	}
	return true
}

func (o Overlap) String() string {
	mods := make([]string, 0, len(o.In))
	for m := range o.In {
		mods = append(mods, m)
	}
	sort.Strings(mods)
	parts := make([]string, 0, len(mods))
	for _, m := range mods {
		parts = append(parts, m+":"+o.In[m])
	}
	kind := "DIVERGED"
	if o.Identical() {
		kind = "identical"
	}
	return fmt.Sprintf("%-16s %-9s %s", o.Name, kind, strings.Join(parts, "  "))
}

// bodyOf extracts a function's source from the start of its declaration to
// the first line that is exactly "}".
//
// ⚠️ Deliberately crude, and it has to be said rather than hidden: a brace on
// its own line is how gofmt ends a top-level function, and gofmt is mandatory
// here, so this is reliable for formatted source and would not be for
// arbitrary Go. The hash is a divergence SIGNAL, not an identity proof.
func bodyOf(src, name string) string {
	loc := regexp.MustCompile(`(?m)^func ` + regexp.QuoteMeta(name) + `\(`).
		FindStringIndex(src)
	if loc == nil {
		return ""
	}
	rest := src[loc[0]:]
	if end := strings.Index(rest, "\n}\n"); end >= 0 {
		return rest[:end+3]
	}
	return rest
}

func shortHash(s string) string {
	// Blank lines dropped so that formatting-only differences do not read as
	// divergence.
	var kept []string
	for _, line := range strings.Split(s, "\n") {
		if strings.TrimSpace(line) != "" {
			kept = append(kept, line)
		}
	}
	sum := sha256.Sum256([]byte(strings.Join(kept, "\n")))
	return hex.EncodeToString(sum[:])[:12]
}

// Overlaps finds every function name present in two or more modules.
func Overlaps(root string, modules []string) ([]Overlap, error) {
	seen := map[string]map[string]string{}
	for _, m := range modules {
		// ⚠️ GO ONLY, AND THAT IS A DECISION RATHER THAN AN OVERSIGHT.
		//
		// This was widened to read Python on 16 Oct 2026 and immediately
		// reported ELEVEN new overlaps, of which one was real. The rest were
		// names Python does not let you choose:
		//
		//	do_GET, log_message, handle_one_request   required overrides of
		//	                                          BaseHTTPRequestHandler
		//	_log, safe_local_imports                  mandated by AGENTS
		//	__init__, send, close, drain              ordinary method names
		//
		// A NAME-keyed check assumes the name was chosen. In Python it is
		// frequently dictated, by a framework, by a convention or by the
		// language. So the check that UNDERSTATES duplication in Go
		// massively OVERSTATES it in Python, and an allowlist of eleven
		// forced names is exactly the rubber-stamp this file's own header
		// warns about.
		//
		// Python is covered by SharedBodies instead, which keys on the BODY
		// and does not care what anything is called. That is the right key
		// for a language where the name is not a choice -- and it found the
		// one true duplicate among the eleven.
		var files []string
		found, err := filepath.Glob(filepath.Join(root, m, "*.go"))
		if err != nil {
			return nil, err
		}
		files = append(files, found...)
		for _, f := range files {
			if strings.HasSuffix(f, "_test.go") {
				continue
			}
			raw, err := os.ReadFile(f)
			if err != nil {
				return nil, err
			}
			src := string(raw)
			re := funcRe
			if strings.HasSuffix(f, ".py") {
				re = pyFuncRe
			}
			for _, hit := range re.FindAllStringSubmatch(src, -1) {
				name := hit[1]
				if seen[name] == nil {
					seen[name] = map[string]string{}
				}
				// ⚠️ A build-tag pair inside ONE module is not an overlap, so
				// the first body found per module wins. Without this every
				// _windows/_other pair would report as duplication and the
				// signal would drown.
				if _, had := seen[name][m]; !had {
					seen[name][m] = shortHash(bodyOfIn(f, src, name))
				}
			}
		}
	}
	var out []Overlap
	for name, in := range seen {
		if len(in) > 1 {
			out = append(out, Overlap{Name: name, In: in})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// SumEntries returns every module path named in any go.sum in the repo.
func SumEntries(root string, modules []string) (map[string][]string, error) {
	out := map[string][]string{}
	for _, m := range modules {
		raw, err := os.ReadFile(filepath.Join(root, m, "go.sum"))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		for _, line := range strings.Split(string(raw), "\n") {
			fields := strings.Fields(line)
			if len(fields) < 1 {
				continue
			}
			out[fields[0]] = append(out[fields[0]], m)
		}
	}
	return out, nil
}

func main() {
	root := ".."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	modules, err := modulesIn(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot read %s: %v\n", root, err)
		os.Exit(1)
	}
	if len(modules) == 0 {
		fmt.Fprintf(os.Stderr, "no Go modules under %s. Run this from a "+
			"checkout: a module fetched by `go run module@sha` has no "+
			"siblings to check.\n", root)
		os.Exit(1)
	}
	fmt.Printf("modules: %s\n\n", strings.Join(modules, " "))

	sums, err := SumEntries(root, modules)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	bad := 0
	fmt.Println("PROVENANCE (s55 maxim 1): every module in any go.sum is ours")
	if len(sums) == 0 {
		fmt.Println("  nothing required anywhere yet")
	}
	for path, by := range sums {
		if !strings.HasPrefix(path, repoPrefix) {
			fmt.Printf("  FOREIGN  %s  required by %s\n", path, strings.Join(by, " "))
			bad++
		}
	}

	fmt.Println("\nOVERLAP: the same function name in two or more connectors")
	overlaps, err := Overlaps(root, modules)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	for _, o := range overlaps {
		mark := "    "
		if !Acknowledged(o.Name) {
			mark = "NEW "
			bad++
		}
		fmt.Printf("  %s%s\n", mark, o)
	}
	// ⚠️ SAME CODE, DIFFERENT NAME. The check above keys on the function
	// name, and s60 recorded its blind spot: name-matching UNDERSTATES
	// duplication. This inverts the key.
	fmt.Println("\nSHARED BODIES: the same code in two modules, whatever it is called")
	shared, err := SharedBodies(root, modules)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	if len(shared) == 0 {
		fmt.Printf("  none over %d significant lines\n", minBodyLines)
	}
	for _, sb := range shared {
		names := sb.Names()
		mark := "    "
		// ⚠️ EVERY undeclared shared body is flagged, whatever it is called.
		//
		// This used to skip single-name cases on the grounds that Overlaps
		// covered them. It does not: Overlaps reads Go only, deliberately,
		// so a Python body shared under one name was reported here and
		// flagged nowhere. Tolerated-by-blindness is the state the baseline
		// exists to prevent.
		// ⚠️ And not flagged twice for ONE fact. `contained` is already
		// declared by NAME in known_overlaps.go, and demanding a second
		// declaration of the same duplication in acknowledgedBodies is how
		// an allowlist becomes a formality people tick without reading.
		if !allNamesAcknowledged(names) && !AcknowledgedBody(names) {
			mark = "NEW "
			bad++
		}
		fmt.Printf("  %s%s\n", mark, sb)
	}

	fmt.Println("\nDECLARED RECIPES: the same sequence of steps, written twice")
	fmt.Println("  (nothing finds these mechanically, so they are written down)")
	for _, r := range recipes {
		fmt.Printf("    %s\n      %s\n", r.Name, strings.Join(r.Sites, "  "))
	}

	if bad > 0 {
		fmt.Printf("\n%d thing(s) need a look. A NEW overlap is not a defect; "+
			"it is undeclared.\n", bad)
		os.Exit(1)
	}
	fmt.Println("\nall declared")
}

// allNamesAcknowledged reports whether every name a shared body hides behind
// is already declared in the name baseline.
func allNamesAcknowledged(names []string) bool {
	if len(names) == 0 {
		return false
	}
	for _, n := range names {
		if !Acknowledged(n) {
			return false
		}
	}
	return true
}
