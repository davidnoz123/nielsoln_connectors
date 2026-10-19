// canonjson -- can Go reproduce Python's canonical JSON, byte for byte?
//
//	go run . records.jsonl python.out
//
// # WHY THIS NUMBER MATTERS
//
// patchbucket computes a patch's integrity hash as sha256 over
// `json.dumps(patch, ensure_ascii=False, sort_keys=True)` and REFUSES any
// patch whose hash does not match. So a Go peer that cannot reproduce those
// exact bytes cannot write a patch the Python side will accept -- every one
// comes back `invalid_patch_hash`. That is a loud failure rather than a silent
// one, which is good, but it blocks the whole all-Go route.
//
// ⚠️ MEASURED BEFORE BUILDING ANYTHING ON IT. The alternative was to write a
// Go store and a Go row-patch writer first and discover the mismatch at the
// end. A fifty-line probe answers it now.
//
// # THE TWO STRATEGIES, AND WHY BOTH ARE MEASURED
//
//	A  json.Marshal of a decoded map[string]any, SetEscapeHTML(false).
//	   The obvious approach. Go sorts map keys for free.
//	B  decode with UseNumber() and emit each number's ORIGINAL TEXT, with a
//	   hand-written encoder that sorts keys.
//
// They fail differently, and knowing which cases break which is the whole
// output. A loses the int/float distinction, because every JSON number becomes
// a float64. B preserves the input text, which is only right where Python's
// output happens to equal its input.
package main

import (
	"bytes"

	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr,
			"usage: canonjson <records.jsonl> <python-reference-output>\n"+
				"Produce the reference with: python canon.py records.jsonl")
		os.Exit(2)
	}
	records, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot read records: %v\n", err)
		os.Exit(1)
	}
	reference, err := os.ReadFile(os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot read the Python reference: %v\n", err)
		os.Exit(1)
	}

	refLines := splitLines(string(reference))
	recLines := splitLines(string(records))
	if len(refLines) != len(recLines) {
		fmt.Fprintf(os.Stderr,
			"⚠️ %d records but %d reference lines. The reference is stale; "+
				"re-run canon.py.\n", len(recLines), len(refLines))
		os.Exit(1)
	}

	var okA, okB, okC int
	var failed []string
	fmt.Printf("%-24s %-3s %-3s %-3s %s\n", "CASE", "A", "B", "C", "divergence")
	fmt.Println(strings.Repeat("-", 78))

	for i, rec := range recLines {
		want := refLines[i]
		if tab := strings.IndexByte(want, '\t'); tab >= 0 {
			want = want[tab+1:]
		}

		gotA, errA := strategyA([]byte(rec))
		gotB, errB := strategyB([]byte(rec))
		gotC, errC := strategyC([]byte(rec))

		matchA := errA == nil && gotA == want
		matchB := errB == nil && gotB == want
		matchC := errC == nil && gotC == want
		if matchA {
			okA++
		}
		if matchB {
			okB++
		}
		if matchC {
			okC++
		}

		name := caseName(rec)
		note := ""
		// ⚠️ Only C's divergences are reported in full. A and B are known
		// broken and listing all 32 of A's failures would bury the one case
		// that still needs a decision.
		if !matchC {
			note = "C differs: " + firstDiff(want, gotC)
			failed = append(failed, fmt.Sprintf("  %s\n      python: %s\n"+
				"      go  C : %s", name, want, gotC))
		}
		fmt.Printf("%-24s %-3s %-3s %-3s %s\n", truncate(name, 24),
			tick(matchA), tick(matchB), tick(matchC), note)
	}

	n := len(recLines)
	fmt.Printf("\nMEASURED over %d records: A %d/%d, B %d/%d, C %d/%d\n",
		n, okA, n, okB, n, okC, n)
	// ⚠️ The hash is over the WHOLE patch, so one divergent field in one row
	// breaks the entire patch. A per-case score is diagnostic; the verdict is
	// all-or-nothing.
	fmt.Println("\n⚠️ VERDICT, and it is all-or-nothing: patch_hash covers the " +
		"whole envelope, so ONE divergent value anywhere makes the patch " +
		"unacceptable. A strategy is only usable at 100%.")
	switch {
	case okC == n:
		fmt.Println("   ✅ STRATEGY C reproduces Python EXACTLY on this corpus: " +
			"json.Number for verbatim number text, hand-sorted keys, Python's\n" +
			"      \", \" and \": \" separators, and a hand-written string " +
			"escaper. All four are required; none is optional.")
		fmt.Println("\n   ⚠️ WHAT THIS DOES NOT ESTABLISH. C round-trips bytes " +
			"that PYTHON wrote. A Go peer ORIGINATING a patch from its own\n" +
			"      store has no original text to echo, so it must format " +
			"numbers as Python's repr() would -- `.0` on whole floats, a\n" +
			"      two-digit exponent, and Python's own threshold for " +
			"switching to exponent form. That is mechanical but unwritten,\n" +
			"      and it is the next thing to measure, not something to " +
			"assume from this result.")
	case okB == n:
		fmt.Println("   strategy B reproduces Python exactly on this corpus.")
	case okA == n:
		fmt.Println("   strategy A reproduces Python exactly on this corpus.")
	default:
		fmt.Printf("   NO strategy is usable as written. A is %d short, "+
			"B is %d short, C is %d short.\n", n-okA, n-okB, n-okC)
	}
	if len(failed) > 0 {
		fmt.Println("\nDIVERGENT CASES IN FULL")
		for _, f := range failed {
			fmt.Println(f)
		}
	}
}

// strategyA re-encodes a decoded map with HTML escaping disabled.
func strategyA(rec []byte) (string, error) {
	var v any
	if err := json.Unmarshal(rec, &v); err != nil {
		return "", err
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	// ⚠️ WITHOUT THIS, `<` `>` and `&` become < > & and nothing
	// Python writes will ever match. It is the first thing to get right and
	// the easiest to forget, since json.Marshal has no such option.
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return "", err
	}
	// Encode appends a newline. Encoder is still the right tool, because
	// SetEscapeHTML exists only on it.
	return strings.TrimSuffix(buf.String(), "\n"), nil
}

// strategyB emits each number's original text and sorts keys by hand.
func strategyB(rec []byte) (string, error) {
	dec := json.NewDecoder(bytes.NewReader(rec))
	// ⚠️ UseNumber is the whole point: without it every number becomes a
	// float64 and the integer 1700000000 is indistinguishable from the float
	// 1700000000.0, which Python renders differently.
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return "", err
	}
	var sb strings.Builder
	if err := writeCanon(&sb, v); err != nil {
		return "", err
	}
	return sb.String(), nil
}

// strategyC is strategyB with the Python-matching string escaper.
func strategyC(rec []byte) (string, error) {
	dec := json.NewDecoder(bytes.NewReader(rec))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return "", err
	}
	var sb strings.Builder
	if err := writeCanonPy(&sb, v); err != nil {
		return "", err
	}
	return sb.String(), nil
}

func writeCanon(sb *strings.Builder, v any) error {
	return canon(sb, v, encodeStringOrPanic)
}

func writeCanonPy(sb *strings.Builder, v any) error {
	return canon(sb, v, pythonString)
}

func encodeStringOrPanic(s string) string {
	out, err := encodeString(s)
	if err != nil {
		return ""
	}
	return out
}

func canon(sb *strings.Builder, v any, esc func(string) string) error {
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
		sb.WriteString(t.String()) // verbatim, as it appeared in the input
	case string:
		sb.WriteString(esc(t))
	case []any:
		sb.WriteByte('[')
		for i, e := range t {
			if i > 0 {
				sb.WriteString(", ") // Python's default separator
			}
			if err := canon(sb, e, esc); err != nil {
				return err
			}
		}
		sb.WriteByte(']')
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		// Byte-wise sort. For UTF-8 this equals Python's code-point order,
		// which is what sort_keys=True gives.
		sort.Strings(keys)
		sb.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(esc(k))
			sb.WriteString(": ")
			if err := canon(sb, t[k], esc); err != nil {
				return err
			}
		}
		sb.WriteByte('}')
	default:
		return fmt.Errorf("unexpected type %T", v)
	}
	return nil
}

// encodeString escapes one string the way Go does with HTML escaping off.
//
// ⚠️ DELIBERATELY GO'S OWN ESCAPER rather than a hand-rolled one. The question
// this probe answers is whether GO can match Python, so substituting my own
// idea of correct escaping would measure my idea instead of Go's.
func encodeString(s string) (string, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(s); err != nil {
		return "", err
	}
	return strings.TrimSuffix(buf.String(), "\n"), nil
}

// pythonString escapes one string the way CPython's json does with
// ensure_ascii=False, which Go's own escaper cannot be configured to do.
//
// ⚠️ THE ONE RULE GO GETS WRONG, and it is the only reason strategy B misses.
// encoding/json ALWAYS escapes U+2028 and U+2029 -- a deliberate choice, since
// a raw line separator inside a <script> tag terminates a JavaScript string
// literal. Python emits them literally. There is no option to turn it off, so
// matching Python means writing the escaper.
//
// CPython's rule, from json/encoder.py's ESCAPE_DCT: escape `"` and `\`,
// escape control characters below 0x20 using the short forms \b \t \n \f \r
// where they exist and \uXXXX otherwise, and emit EVERYTHING ELSE literally --
// including DEL, U+2028 and every non-ASCII rune.
func pythonString(s string) string {
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

func splitLines(s string) []string {
	var out []string
	for _, l := range strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n") {
		if strings.TrimSpace(l) != "" {
			out = append(out, l)
		}
	}
	return out
}

// caseName is the first key of a record, used as its label.
func caseName(rec string) string {
	dec := json.NewDecoder(strings.NewReader(rec))
	if _, err := dec.Token(); err != nil {
		return "?"
	}
	tok, err := dec.Token()
	if err != nil {
		return "?"
	}
	if s, ok := tok.(string); ok {
		return s
	}
	return "?"
}

// firstDiff shows the first point of divergence, with a little context.
func firstDiff(want, got string) string {
	i := 0
	for i < len(want) && i < len(got) && want[i] == got[i] {
		i++
	}
	from := i - 6
	if from < 0 {
		from = 0
	}
	return fmt.Sprintf("at %d py=%q go=%q", i,
		truncate(want[from:], 22), truncate(got[from:], 22))
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func tick(ok bool) string {
	if ok {
		return "ok"
	}
	return "NO"
}
