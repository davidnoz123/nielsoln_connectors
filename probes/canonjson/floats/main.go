// floats -- can Go ORIGINATE a float the way Python renders it?
//
//	go run . ../floats.bits ../floats.out
//
// canonjson answered a narrower question: Go can reproduce Python's canonical
// JSON when it has Python's own number TEXT to echo, via json.Number. A Go
// peer writing a patch from its own store has no such text. It holds a
// float64 and must render the characters Python would have rendered, or the
// patch hash will not match and the patch is refused.
//
// ⚠️ BOTH SIDES START FROM HEX BIT PATTERNS, so this measures RENDERING only.
// Feeding decimal literals would tangle a parsing question into a formatting
// one, and parsing is not what is in doubt: IEEE-754 decimal conversion is
// specified and both languages round correctly.
package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: floats <floats.bits> <floats.out>")
		os.Exit(2)
	}
	bitsRaw, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot read bits: %v\n", err)
		os.Exit(1)
	}
	refRaw, err := os.ReadFile(os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot read the Python reference: %v\n", err)
		os.Exit(1)
	}

	want := map[string]string{}
	for _, line := range strings.Split(strings.ReplaceAll(string(refRaw), "\r\n", "\n"), "\n") {
		if parts := strings.SplitN(strings.TrimSpace(line), "\t", 2); len(parts) == 2 {
			want[parts[0]] = parts[1]
		}
	}

	var okG, okP int
	var n int
	fmt.Printf("%-18s %-26s %-26s %-4s %-4s\n",
		"BITS", "python", "go: pythonFloat()", "'g'", "ours")
	fmt.Println(strings.Repeat("-", 86))

	for _, h := range strings.Fields(string(bitsRaw)) {
		u, err := strconv.ParseUint(h, 16, 64)
		if err != nil {
			continue
		}
		f := math.Float64frombits(u)
		w, has := want[h]
		if !has {
			continue
		}
		n++

		// The naive attempt: Go's shortest round-trip 'g'.
		naive := strconv.FormatFloat(f, 'g', -1, 64)
		ours := pythonFloat(f)

		if naive == w {
			okG++
		}
		if ours == w {
			okP++
		}
		mark := func(s string) string {
			if s == w {
				return "ok"
			}
			return "NO"
		}
		fmt.Printf("%-18s %-26s %-26s %-4s %-4s\n",
			h, w, ours, mark(naive), mark(ours))
	}

	fmt.Printf("\nMEASURED over %d doubles: Go 'g' %d/%d, pythonFloat() %d/%d\n",
		n, okG, n, okP, n)
	switch {
	case okP == n:
		fmt.Println("\n✅ A Go peer CAN originate Python-identical numbers. The " +
			"rules are in pythonFloat() below and there are only three:\n" +
			"   whole floats take a trailing \".0\"; the exponent carries a " +
			"sign and at least two digits; and the switch to\n" +
			"   exponent form happens at Python's thresholds, not Go's.")
	default:
		fmt.Printf("\n⚠️ %d double(s) STILL DIVERGE. Until that is 0, a Go-written "+
			"patch cannot be hashed to match Python's, and the all-Go\n"+
			"   route needs either a different canonical form upstream or a "+
			"numeric type that is not float64.\n", n-okP)
	}
}

// pythonFloat renders a float64 the way CPython's json.dumps does.
//
// CPython calls repr(float), which is shortest-round-trip, then json emits it
// verbatim. Three rules separate that from Go's strconv:
//
//  1. ⚠️ A WHOLE FLOAT KEEPS ITS ".0". Python prints 1700000000.0 where Go's
//     'g' prints 1.7e+09 and Go's json prints 1700000000. This is the one that
//     would bite in practice, because create_time and update_time are REAL
//     columns and a timestamp with no fractional part is ordinary.
//  2. THE EXPONENT IS SIGNED AND AT LEAST TWO DIGITS: 1e-07, not 1e-7.
//  3. ⚠️ THE THRESHOLDS ARE PYTHON'S. repr switches to exponent form at 1e16
//     and below 1e-4, which is not where Go's 'g' switches.
func pythonFloat(f float64) string {
	switch {
	case math.IsInf(f, 1):
		return "Infinity"
	case math.IsInf(f, -1):
		return "-Infinity"
	case math.IsNaN(f):
		return "NaN"
	}

	abs := math.Abs(f)
	// Python's repr uses exponent form outside [1e-4, 1e16).
	if abs != 0 && (abs < 1e-4 || abs >= 1e16) {
		s := strconv.FormatFloat(f, 'e', -1, 64)
		return fixExponent(s)
	}
	s := strconv.FormatFloat(f, 'f', -1, 64)
	if !strings.ContainsAny(s, ".") {
		s += ".0" // rule 1
	}
	return s
}

// fixExponent turns Go's exponent spelling into Python's.
//
// Go gives 1e-07 as "1e-07" but 1e+21 as "1e+21" and 1e16 as "1e+16"; the
// mantissa may also carry a trailing ".0" Python does not want, or lack one
// Python does. Normalised here rather than assumed.
func fixExponent(s string) string {
	i := strings.IndexAny(s, "eE")
	if i < 0 {
		return s
	}
	mant, exp := s[:i], s[i+1:]
	sign := "+"
	if exp != "" && (exp[0] == '+' || exp[0] == '-') {
		sign = string(exp[0])
		exp = exp[1:]
	}
	for len(exp) < 2 {
		exp = "0" + exp // rule 2
	}
	// Python's repr never leaves a bare "1." mantissa and never pads it.
	mant = strings.TrimSuffix(mant, ".")
	return mant + "e" + sign + exp
}
