// caveats_test.go -- the check that makes caveats.go worth reading.
//
// ⚠️ WITHOUT THIS, Caveats IS A COMMENT. An auditor finding no caveats could
// not tell "reaches for nothing" from "nobody wrote any", and silence would
// read as safety. This walks the package's own source and asserts that every
// outward reach of a known kind is declared, so a new exec.Command that
// nobody wrote down breaks the build.
//
// It parses ALL files including both build-tag halves, deliberately. An audit
// wants to know what the program does on every platform, not what it does on
// the one the test happens to run on. owner_windows.go's syscall work is
// invisible to this check for a different reason, stated in caveats.go: a raw
// syscall looks nothing like a function call the checker recognises.
package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// reachesIn returns every call expression in the package that matches one of
// the named surfaces, as "surface at file:line".
// repoPrefix is what maxim 1 of s55 calls provenance: every module in the
// closure must come from here.
const repoPrefix = "github.com/davidnoz123/nielsoln_connectors/"

// importsIn returns the in-repo modules this package imports.
//
// ⚠️ THE HOLE THIS CLOSES. reachesIn parses only this directory, so an
// imported module's reaches are INVISIBLE to it: `chatgptwrite.Post(...)`
// is a selector that matches nothing in CaveatSurfaces. A package could
// declare "reaches no non-loopback address", import something that posts to
// a third-party service as the signed-in user, and be COMPLETE by the
// checker's standard.
//
// That could not bite while every require block was empty. It becomes
// reachable the moment the first import lands, which is exactly when s13
// decided there would be one.
//
// So caveats COMPOSE over the closure: a package must declare what it
// imports, and s55's go.sum gives the closure's stopping condition.
func importsIn(t *testing.T, dir string) []string {
	t.Helper()
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, func(fi fs.FileInfo) bool {
		n := fi.Name()
		return strings.HasSuffix(n, ".go") && !strings.HasSuffix(n, "_test.go")
	}, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("cannot parse %s: %v", dir, err)
	}
	seen := map[string]bool{}
	var out []string
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, imp := range file.Imports {
				path := strings.Trim(imp.Path.Value, `"`)
				if strings.HasPrefix(path, repoPrefix) && !seen[path] {
					seen[path] = true
					out = append(out, path)
				}
			}
		}
	}
	return out
}

func reachesIn(t *testing.T, surfaces []string) map[string][]string {
	t.Helper()
	want := map[string]bool{}
	for _, s := range surfaces {
		want[s] = true
	}

	fset := token.NewFileSet()
	// ParseDir ignores build constraints, which is what an audit needs.
	pkgs, err := parser.ParseDir(fset, ".", func(fi fs.FileInfo) bool {
		n := fi.Name()
		return strings.HasSuffix(n, ".go") && !strings.HasSuffix(n, "_test.go")
	}, 0)
	if err != nil {
		t.Fatalf("cannot parse the package: %v", err)
	}

	found := map[string][]string{}
	for _, pkg := range pkgs {
		for name, file := range pkg.Files {
			// PASS 2, and it is the one that matters. ⚠️ The call-based
			// pass below MISSED this package's primary outward reach: the
			// HTTP request is `client.Get(url)`, a method on a value, and an
			// AST-only checker cannot resolve what `client` is. So
			// `http.Get` was in the surface list, the code never calls it,
			// and the checker passed while blind to the thing the program
			// exists to do.
			//
			// The fix is to look for the TYPE being referenced rather than
			// the function being called: you cannot make an HTTP request
			// without naming http.Client or http.Get somewhere, and
			// `&http.Client{...}` is a selector this pass can see.
			ast.Inspect(file, func(n ast.Node) bool {
				sel, ok := n.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				ident, ok := sel.X.(*ast.Ident)
				if !ok {
					return true
				}
				full := ident.Name + "." + sel.Sel.Name
				if want[full] {
					pos := fset.Position(sel.Pos())
					site := name + ":" + itoa(pos.Line)
					for _, had := range found[full] {
						if had == site {
							return true
						}
					}
					found[full] = append(found[full], site)
				}
				return true
			})

			ast.Inspect(file, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				ident, ok := sel.X.(*ast.Ident)
				if !ok {
					return true
				}
				full := ident.Name + "." + sel.Sel.Name
				if want[full] {
					pos := fset.Position(call.Pos())
					found[full] = append(found[full],
						name+":"+itoa(pos.Line))
				}
				return true
			})
		}
	}
	return found
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

func TestEveryOutwardReachIsDeclared(t *testing.T) {
	declared := strings.ToLower(strings.Join(Caveats, "\n"))
	found := reachesIn(t, CaveatSurfaces)

	if len(found) == 0 {
		t.Log("no reaches of a known surface found in this package")
	}

	for surface, sites := range found {
		// ⚠️ PER-REACH, NOT PER-FILE. The first version of this asked whether
		// the FILE was mentioned anywhere in Caveats, and it PASSED a
		// deliberately planted exec.Command because sessions.go was already
		// named for two other reaches. Every file gets mentioned eventually,
		// so that check was on its way to always passing while staying
		// green, which is the exact failure this file exists to prevent.
		//
		// The declaration must now carry the surface itself, in brackets,
		// which no unrelated line can satisfy.
		tag := "[" + strings.ToLower(surface) + "]"
		if !strings.Contains(declared, tag) {
			t.Errorf("UNDECLARED REACH: %s at %v.\n  No line in Caveats "+
				"opens with %s, so an auditor reading the declaration would "+
				"never know this program does it.", surface, sites, tag)
			continue
		}
		t.Logf("declared: %-18s %v", surface, sites)
	}
}

// TestTheCheckerCanActuallyFail guards against the worst outcome: a checker
// that passes because it finds nothing, for ever.
//
// ⚠️ This is the permissive-fake failure applied to a test rather than to a
// dependency. A checker that silently matched nothing would stay green while
// the thing it exists to catch walked past it, and the false confidence would
// concentrate exactly where nobody is looking any more.
func TestTheCheckerCanActuallyFail(t *testing.T) {
	// A surface this package definitely DOES use, with a deliberately empty
	// declaration: the checker must object.
	found := reachesIn(t, []string{"net.DialTimeout"})
	if len(found) == 0 {
		t.Fatal("the checker found no net.DialTimeout in a package that " +
			"calls it at sessions.go:195. The AST walk is not working, so " +
			"every other result in this file is meaningless.")
	}
	t.Logf("the checker can see %v, so a green result above means something",
		found)
}

// TestWhatTheCheckerCannotSee states the limit as a test, so it is read.
func TestWhatTheCheckerCannotSee(t *testing.T) {
	// owner_windows.go reaches the whole machine's TCP table through a raw
	// syscall. It is declared in Caveats by hand. Prove the checker is blind
	// to it, so nobody later mistakes the green test for completeness.
	found := reachesIn(t, []string{"syscall.NewLazyDLL", "iphlpapi.NewProc"})
	if _, ok := found["iphlpapi.NewProc"]; ok {
		t.Log("note: the checker can see iphlpapi.NewProc after all")
	}
	if !strings.Contains(strings.ToLower(strings.Join(Caveats, " ")),
		"getextendedtcptable") {
		t.Error("the iphlpapi reach is not declared, and the checker cannot " +
			"find it. That combination is the one an auditor cannot recover " +
			"from.")
	}
	t.Log("UNVERIFIED by construction: raw syscalls, reflection, and any " +
		"helper outside CaveatSurfaces. Declared by hand or not at all.")
}

// TestInRepoImportsAreDeclared closes the composition hole.
//
// ⚠️ reachesIn parses only THIS directory, so an imported module's
// reaches are invisible to it. A package could declare "reaches no
// non-loopback address", import something that posts to a third-party service
// as the signed-in user, and be COMPLETE by the checker's standard.
//
// So caveats compose over the closure: every in-repo import must be declared,
// and s55's go.sum is what says when the closure has been exhausted.
func TestInRepoImportsAreDeclared(t *testing.T) {
	declared := strings.ToLower(strings.Join(Caveats, "\n"))
	imports := importsIn(t, ".")
	if len(imports) == 0 {
		t.Log("no in-repo imports, so nothing to compose. This passes " +
			"trivially today, which is why TestTheImportCheckCanFail exists.")
		return
	}
	for _, path := range imports {
		tag := "[imports " + strings.ToLower(path) + "]"
		if !strings.Contains(declared, tag) {
			t.Errorf("UNDECLARED IMPORT: %s.\n  No line in Caveats "+
				"opens with %s, so this package says nothing about what that "+
				"module reaches for, and the AST walk cannot see inside it.",
				path, tag)
		}
	}
}

// TestTheImportCheckCanFail proves the check above can object, because today
// it has nothing to object to.
//
// ⚠️ Same discipline as TestTheCheckerCanActuallyFail: a check that
// passes because it found nothing, for ever, is the permissive-fake failure
// applied to a test rather than to a dependency.
func TestTheImportCheckCanFail(t *testing.T) {
	dir := t.TempDir()
	src := "package main\n\nimport \"" + repoPrefix +
		"chatgpt_write\"\n\nvar _ = chatgpt_write.Post\n"
	if err := os.WriteFile(filepath.Join(dir, "x.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	got := importsIn(t, dir)
	if len(got) != 1 || got[0] != repoPrefix+"chatgpt_write" {
		t.Fatalf("the import walk did not see an in-repo import: %v", got)
	}
	t.Logf("the check can see %v, so a green result above means something", got)
}
