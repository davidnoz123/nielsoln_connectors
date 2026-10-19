// canon.go -- Python's canonical JSON, in Go, byte for byte.
//
// This is `json.dumps(obj, ensure_ascii=False, sort_keys=True)` reproduced
// exactly, because patchbucket hashes a patch with that call and REFUSES any
// patch whose sha256 does not match. Anything less than byte-identical makes
// every Go-written patch `invalid_patch_hash`.
//
// ⚠️ EVERY RULE BELOW WAS MEASURED, NOT INFERRED, by probes/canonjson and
// probes/canonjson/floats on 19 Oct 2026. The scores are in probes/README.md.
// The obvious implementation -- json.Marshal of a decoded map with
// SetEscapeHTML(false) -- scored 0 out of 32. Not 31. Zero.
package main

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

// canonical renders v the way CPython's json.dumps does with
// ensure_ascii=False and sort_keys=True.
//
// ⚠️ FOUR RULES, AND NONE IS OPTIONAL. Dropping any one of them scored a
// mismatch in the probe:
//
//  1. numbers emitted from their ORIGINAL TEXT where there is one
//     (json.Number), because decoding into `any` turns every number into a
//     float64 and 1700000000.0 then renders as 1700000000.
//  2. keys sorted bytewise, which equals Python's code-point order for UTF-8.
//  3. Python's separators, ", " and ": ", WITH THE SPACES. Go's Encoder emits
//     "," and ":" and has no option to change them. This alone failed all 32.
//  4. Python's string escaping, which is not Go's: see canonString.
func canonical(v any) (string, error) {
	var sb strings.Builder
	if err := writeCanonical(&sb, v); err != nil {
		return "", err
	}
	return sb.String(), nil
}

func writeCanonical(sb *strings.Builder, v any) error {
	switch t := v.(type) {
	case nil:
		sb.WriteString("null")
	case bool:
		if t {
			sb.WriteString("true")
		} else {
			sb.WriteString("false")
		}
	case json.Number:
		// Rule 1. Verbatim, so a value that came from Python goes back
		// exactly as it arrived.
		sb.WriteString(t.String())
	case float64:
		// No original text, so it must be RENDERED as Python would. This is
		// the path a Go-originated patch takes.
		sb.WriteString(canonFloat(t))
	case int:
		sb.WriteString(strconv.Itoa(t))
	case int64:
		sb.WriteString(strconv.FormatInt(t, 10))
	case string:
		sb.WriteString(canonString(t))
	case []any:
		sb.WriteByte('[')
		for i, e := range t {
			if i > 0 {
				sb.WriteString(", ") // rule 3
			}
			if err := writeCanonical(sb, e); err != nil {
				return err
			}
		}
		sb.WriteByte(']')
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys) // rule 2
		sb.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(canonString(k))
			sb.WriteString(": ") // rule 3
			if err := writeCanonical(sb, t[k]); err != nil {
				return err
			}
		}
		sb.WriteByte('}')
	default:
		// ⚠️ REFUSED RATHER THAN BEST-EFFORTED. An unmodelled type silently
		// rendered some other way would produce a patch whose hash is wrong,
		// and the failure would surface on a peer's machine as
		// invalid_patch_hash with nothing to point at.
		return fmt.Errorf("canonical: no rule for %T. Add one, and add a "+
			"probe case pinning what Python does with it", v)
	}
	return nil
}

// canonString escapes a string the way CPython does with ensure_ascii=False.
//
// ⚠️ GO'S OWN ESCAPER CANNOT BE MADE TO DO THIS. encoding/json always escapes
// U+2028 and U+2029, with no option to stop it, and the reason is sound: a raw
// line separator inside a <script> block terminates a JavaScript string
// literal. Python emits them literally, so this is the one rule that forced a
// hand-written escaper rather than a configured one. It was strategy B's only
// failure in the probe: 31 of 32.
//
// CPython's rule, from json/encoder.py's ESCAPE_DCT: escape `"` and `\`, use
// the short forms for \b \t \n \f \r, use \uXXXX for any other control
// character below 0x20, and emit EVERYTHING ELSE literally -- including DEL,
// U+2028, and every non-ASCII rune.
func canonString(s string) string {
	var sb strings.Builder
	sb.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			sb.WriteString(`\"`)
		case '\\':
			sb.WriteString(`\\`)
		case '\b':
			sb.WriteString(`\b`)
		case '\t':
			sb.WriteString(`\t`)
		case '\n':
			sb.WriteString(`\n`)
		case '\f':
			sb.WriteString(`\f`)
		case '\r':
			sb.WriteString(`\r`)
		default:
			if r < 0x20 {
				fmt.Fprintf(&sb, `\u%04x`, r)
			} else {
				sb.WriteRune(r)
			}
		}
	}
	sb.WriteByte('"')
	return sb.String()
}

// canonFloat renders a float64 the way CPython's json.dumps does.
//
// ⚠️ MEASURED AT 27/27 by probes/canonjson/floats, against Go's own shortest
// round-trip 'g' at 18/27. Three rules:
//
//  1. ⚠️ A WHOLE FLOAT KEEPS ITS ".0". Python prints 1700000000.0 where Go's
//     'g' prints 1.7e+09 and Go's json prints 1700000000. This is the rule
//     that would bite in practice: create_time and update_time are REAL
//     columns, and a timestamp with no fractional part is ordinary.
//  2. The exponent carries a sign and AT LEAST TWO DIGITS: 1e-07, not 1e-7.
//  3. ⚠️ The thresholds are PYTHON's. repr switches to exponent form at 1e16
//     and below 1e-4, which is not where Go's 'g' switches.
func canonFloat(f float64) string {
	switch {
	case math.IsInf(f, 1):
		return "Infinity"
	case math.IsInf(f, -1):
		return "-Infinity"
	case math.IsNaN(f):
		return "NaN"
	}
	if abs := math.Abs(f); abs != 0 && (abs < 1e-4 || abs >= 1e16) {
		return canonExponent(strconv.FormatFloat(f, 'e', -1, 64))
	}
	s := strconv.FormatFloat(f, 'f', -1, 64)
	if !strings.Contains(s, ".") {
		s += ".0" // rule 1
	}
	return s
}

// canonExponent rewrites Go's exponent spelling as Python's.
func canonExponent(s string) string {
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
	return strings.TrimSuffix(mant, ".") + "e" + sign + exp
}
