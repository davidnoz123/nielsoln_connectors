// shared_bodies.go -- the same CODE under two different names.
//
// ⚠️ WHY THIS EXISTS, AND IT IS THE SECOND TIME THE GAP HAS BITTEN.
//
// Overlaps() keys on the function NAME and compares bodies, so it finds a
// name present in two modules and tells you whether the implementations
// agree. s60 wrote down its blind spot at the time: name-matching
// UNDERSTATES duplication, because the two websocket clients in this repo
// never appeared in the overlap count at all -- they use different function
// names, and the only reason anybody knew about them was a conversation.
//
// Nothing was done about it, and then it bit again: `shimp/live_test.go` and
// `linkprobe/main.go` now both drive the same five-step cycle, and the
// overlap report shows only `main` and `probe` between those modules.
//
// So this check inverts the key. Group by BODY, report when one body appears
// in two or more modules, and name the functions it is hiding behind.
//
// ⚠️ WHAT IT CATCHES AND WHAT IT DOES NOT, stated plainly rather than
// discovered later:
//
//	CATCHES    copy-paste-and-rename. Byte-identical logic, new name.
//	MISSES     the same RECIPE written twice in different code.
//
// The live_test/linkprobe cycle is the second kind, and no cheap mechanical
// check finds it, because the two are not copies: they are the same sequence
// of steps expressed differently. known_recipes.go is the honest answer to
// that one -- a declared inventory with a staleness check -- because for
// semantic duplication the only mechanism that works is writing it down.
//
// ⚠️ AND THIS ONE READS _test.go FILES, which Overlaps() deliberately does
// not. A copied helper is far likelier to live in a test than in shipped
// code, and excluding tests is part of why the cycle above went unseen.
package main

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// minBodyLines is the floor below which a shared body means nothing.
//
// ⚠️ Without it this check is pure noise: `func (x T) Close() error { ... }`
// one-liners, two-line error wrappers and empty stubs collide constantly
// across modules and none of it is duplication worth a human's attention.
// Measured on this repo: at 3 lines the report is unreadable, at 5 it is
// short enough to act on.
const minBodyLines = 5

// SharedBody is one body found under more than one name or module.
type SharedBody struct {
	Hash string

	// Where maps "module.funcName" to nothing; it is a set, kept sorted on
	// read so the report is stable.
	Where []string

	// Lines is the body's size, reported because a 6-line collision and a
	// 60-line one deserve different reactions.
	Lines int
}

func (s SharedBody) String() string {
	return strings.Join(s.Where, "  ") + "  (" + strconv.Itoa(s.Lines) +
		" lines, " + s.Hash + ")"
}

// Names returns the distinct function names this body hides behind.
func (s SharedBody) Names() []string {
	seen := map[string]bool{}
	var out []string
	for _, w := range s.Where {
		if i := strings.LastIndex(w, "."); i >= 0 {
			n := w[i+1:]
			if !seen[n] {
				seen[n] = true
				out = append(out, n)
			}
		}
	}
	sort.Strings(out)
	return out
}

// Modules returns the distinct modules this body appears in.
func (s SharedBody) Modules() []string {
	seen := map[string]bool{}
	var out []string
	for _, w := range s.Where {
		if i := strings.Index(w, "."); i >= 0 {
			m := w[:i]
			if !seen[m] {
				seen[m] = true
				out = append(out, m)
			}
		}
	}
	sort.Strings(out)
	return out
}

// SharedBodies finds bodies of at least minBodyLines that appear in more
// than one module, regardless of what each copy is called.
func SharedBodies(root string, modules []string) ([]SharedBody, error) {
	type hit struct {
		where string
		lines int
	}
	byHash := map[string][]hit{}

	for _, m := range modules {
		files, err := filepath.Glob(filepath.Join(root, m, "*.go"))
		if err != nil {
			return nil, err
		}
		for _, f := range files {
			raw, err := os.ReadFile(f)
			if err != nil {
				return nil, err
			}
			src := string(raw)
			// Per module, the first copy of a given name wins, for
			// Overlaps()'s reason: a _windows/_other build-tag pair inside
			// ONE module is not duplication.
			seenHere := map[string]bool{}
			for _, mt := range funcRe.FindAllStringSubmatch(src, -1) {
				name := mt[1]
				if seenHere[name] {
					continue
				}
				seenHere[name] = true
				body := bodyOf(src, name)
				n := countSignificant(body)
				if n < minBodyLines {
					continue
				}
				h := shortHash(afterSignature(body))
				byHash[h] = append(byHash[h], hit{m + "." + name, n})
			}
		}
	}

	var out []SharedBody
	for h, hits := range byHash {
		mods := map[string]bool{}
		var where []string
		for _, it := range hits {
			where = append(where, it.where)
			if i := strings.Index(it.where, "."); i >= 0 {
				mods[it.where[:i]] = true
			}
		}
		// Only CROSS-MODULE sharing is reported. Two identical bodies inside
		// one module are that module's own business and are frequently
		// legitimate, such as a build-tag pair.
		if len(mods) < 2 {
			continue
		}
		sort.Strings(where)
		out = append(out, SharedBody{Hash: h, Where: where, Lines: hits[0].lines})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Lines != out[j].Lines {
			return out[i].Lines > out[j].Lines
		}
		return out[i].Hash < out[j].Hash
	})
	return out, nil
}

// countSignificant counts non-blank, non-comment lines in a body.
//
// ⚠️ Comments excluded deliberately. This repo's functions carry long
// explanatory comments, and counting them would push trivial code over the
// threshold on the strength of its documentation, which is the opposite of
// what the floor is for.
func countSignificant(body string) int {
	n := 0
	for _, line := range strings.Split(body, "\n") {
		t := strings.TrimSpace(line)
		if t == "" || strings.HasPrefix(t, "//") {
			continue
		}
		n++
	}
	return n
}

// afterSignature strips the declaration line, so a RENAMED copy hashes the
// same as its original.
//
// ⚠️ THE CHECK DID NOT WORK WITHOUT THIS, and its own companion test is what
// found out. bodyOf() starts at `func <name>(`, so the name sat inside the
// hashed text and two differently-named copies hashed differently. A check
// built to find renamed copies could not find one, and it would have passed
// every run, green, for exactly the reason three earlier checks in this repo
// passed: it found nothing because it could not see.
//
// ⓘ So the hashes here differ from the ones Overlaps reports for the same
// function. That is deliberate: Overlaps compares implementations under one
// name, where a changed SIGNATURE is itself a divergence worth seeing, and
// this compares code across names, where the signature is the thing being
// ignored.
func afterSignature(body string) string {
	if i := strings.Index(body, "{\n"); i >= 0 {
		return body[i+2:]
	}
	return body
}
