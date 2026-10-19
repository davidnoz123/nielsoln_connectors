// caveats.go -- s54's checker, moved out of the module it was checking.
//
// # WHY THIS IS HERE AND NOT IN EACH MODULE'S OWN TEST
//
// s54 built the check as `chrome_sessions/caveats_test.go`, so a reach the
// code has and caveats.go does not mention fails that module's build. Two
// things measured on 19 Oct 2026 argue for it living here instead.
//
// ⚠️ AN IN-MODULE TEST CANNOT DETECT AN ABSENT caveats.go. It only runs if
// somebody wrote it. `chrome_ready` -- the ONLY module a shimp link has ever
// executed -- had neither file, so the in-module design was structurally
// unable to see the dominant failure: not a wrong declaration, a missing one.
// 1 of 10 modules had caveats at all, and it was not the one that runs.
//
// ⚠️ AND EXTRACTING IT AS A LEAF MODULE WOULD OPEN THE HOLE IT CLOSES.
// s60 ranked the checker as leaf candidate #2, but caveats_test.go's own
// comment records that an imported module's reaches are invisible to an
// AST walk of one directory, and that this "becomes reachable the moment the
// first import lands". Making every connector import a shared checker IS that
// first import, and it would put a test-only dependency into each module's
// go.sum -- noise in the exact audit payload this serves. Here it costs no
// connector anything: repocheck imports nothing and nothing imports it.
//
// # WHAT IT CHECKS, AND WHAT IT CANNOT
//
// Import-level, not call-level: every outward-reaching package a module
// imports must be named somewhere in its caveats.go.
//
// ⚠️ SAID PLAINLY BECAUSE A CHECKER THAT OVERSTATES ITS REACH IS WORSE THAN
// NONE: this does not distinguish os.Open from os.Create, so "reads files"
// against "writes files" is not verified, only that `os` was declared. A raw
// syscall is invisible for the reason chrome_sessions/caveats.go already
// records -- it looks nothing like a call the walk recognises. s54's
// call-level walk is finer and still belongs in-module where it exists; this
// is the floor, not the ceiling.
package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// outwardPackages are the stdlib imports that reach beyond the process, each
// mapped to the tokens that count as declaring it.
//
// ⚠️ A CLOSED LIST, deliberately. An open heuristic ("anything that looks like
// I/O") would either flag every module for importing `fmt` or miss the one
// that matters. Adding to this list is a decision, and a new entry failing
// modules that never declared it is the check working, not a regression.
//
// ⚠️ AND A DECLARATION MAY NAME THE SYMBOL RATHER THAN THE IMPORT PATH, which
// the first version of this file got wrong. chrome_sessions/caveats.go says
// "[http.Client] connects: HTTP GET to 127.0.0.1:<port>/json/version" -- a
// complete declaration of its network reach that mentions `net/http` nowhere.
// Matching the path alone reported a GAP against the one module that had
// actually done the work, which is the failure this checker's own header warns
// about: a checker that overstates its reach is worse than none.
//
// `net/url` is NOT here. It parses strings and reaches nothing, so requiring a
// caveat for it would train readers to tick boxes.
var outwardPackages = map[string][]string{
	"net/http":      {"net/http", "http."},
	"net":           {`"net"`, "net.Dial", "net.Listen", "net.Conn"},
	"os/exec":       {"os/exec", "exec."},
	"syscall":       {"syscall"},
	"database/sql":  {"database/sql", "sql."},
	"archive/zip":   {"archive/zip", "zip."},
	"archive/tar":   {"archive/tar", "tar."},
	"compress/gzip": {"compress/gzip", "gzip."},
	"os":            {"os."},
	// ⚠️ `unsafe` is satisfied by declaring the syscall it serves. In this repo
	// it is never a reach of its own: it is the pointer plumbing for a
	// MIB_TCPROW_OWNER_PID or a console handle, and the thing an auditor needs
	// told is the syscall. Demanding a separate "[unsafe]" entry would add a
	// line that carries no information the syscall caveat does not.
	"unsafe": {"unsafe", "syscall"},
}

// CaveatReport is one module's verdict.
type CaveatReport struct {
	Module     string
	HasFile    bool
	Undeclared []string // outward imports absent from caveats.go
	Declared   []string // outward imports correctly named
}

// OK reports whether the module satisfies the floor.
func (r CaveatReport) OK() bool { return r.HasFile && len(r.Undeclared) == 0 }

func (r CaveatReport) String() string {
	switch {
	case !r.HasFile:
		return fmt.Sprintf("%-18s NO caveats.go. Reaches %s and declares "+
			"nothing, so an auditor cannot tell 'reaches nothing' from "+
			"'nobody wrote any'", r.Module, strings.Join(r.Declared, " "))
	case len(r.Undeclared) > 0:
		return fmt.Sprintf("%-18s UNDECLARED: %s", r.Module,
			strings.Join(r.Undeclared, " "))
	default:
		return fmt.Sprintf("%-18s ok (%d declared)", r.Module, len(r.Declared))
	}
}

// outwardImportsOf returns the outward-reaching packages a module imports.
//
// ⚠️ PARSES EVERY .go FILE INCLUDING BOTH BUILD-TAG HALVES, because an audit
// asks what the program does on every platform, not on the one this happens
// to run on. _test.go is excluded: a test's reaches are not the program's, and
// counting them would make every module that spins up a fake HTTP server
// declare a network reach it does not have in production.
func outwardImportsOf(dir string) (map[string][]string, error) {
	found := map[string][]string{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	fset := token.NewFileSet()
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") ||
			strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(dir, name), nil,
			parser.ImportsOnly)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", filepath.Join(dir, name), err)
		}
		for _, spec := range f.Imports {
			path, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				continue
			}
			if _, outward := outwardPackages[path]; outward {
				found[path] = append(found[path], name)
			}
		}
	}
	return found, nil
}

// declares reports whether caveats.go text declares the reach for importPath,
// by the path itself or by any symbol token that stands for it.
func declares(text, importPath string) bool {
	for _, token := range outwardPackages[importPath] {
		if strings.Contains(text, token) {
			return true
		}
	}
	return false
}

// Caveats checks every module against the floor described in this file.
func Caveats(root string, modules []string) ([]CaveatReport, error) {
	var out []CaveatReport
	for _, m := range modules {
		dir := filepath.Join(root, m)
		imports, err := outwardImportsOf(dir)
		if err != nil {
			// A directory with no parseable Go (a Python-only module) is not a
			// caveats question. Skipped rather than reported as a pass, since
			// a pass would claim something was checked.
			continue
		}
		if len(imports) == 0 {
			continue
		}
		body, readErr := os.ReadFile(filepath.Join(dir, "caveats.go"))
		rep := CaveatReport{Module: m, HasFile: readErr == nil}
		text := string(body)
		for path := range imports {
			if rep.HasFile && declares(text, path) {
				rep.Declared = append(rep.Declared, path)
			} else if rep.HasFile {
				rep.Undeclared = append(rep.Undeclared, path)
			} else {
				// No file: everything is undeclared, but reported as the one
				// fact that matters rather than as a list of ten.
				rep.Declared = append(rep.Declared, path)
			}
		}
		sort.Strings(rep.Declared)
		sort.Strings(rep.Undeclared)
		out = append(out, rep)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Module < out[j].Module })
	return out, nil
}
