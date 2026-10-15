// linkprobe_test.go -- the checks that do not need a machine.
//
// ⚠️ Every check here has a companion proving it CAN fail. Three times in
// this repo a safety check has passed because it found nothing: the caveats
// walk missed its package's primary reach, then matched per-file instead of
// per-reach, then could not see inside an import. A check that is green
// because it is blind is worse than no check.
//
// The actual probing needs a real registry claim and a real Excel, so it is
// not here. `go run . -run` does that, deliberately, because a test that
// installs a protocol handler as a side effect of `go test ./...` is a test
// that modifies the machine of anyone running the repo's loop.
package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEveryRouteIsWellFormed(t *testing.T) {
	seen := map[string]bool{}
	for _, r := range routes {
		if r.ID == "" {
			t.Errorf("a route has no id: %+v", r)
			continue
		}
		if seen[r.ID] {
			t.Errorf("duplicate route id %q, so one of them is unreachable "+
				"by -only and the report has two rows with one name", r.ID)
		}
		seen[r.ID] = true

		if r.What == "" {
			t.Errorf("%s has no prose. A matrix of outcomes with no "+
				"description is unreadable six months later", r.ID)
		}
		if r.Kind == Auto && len(r.Argv) == 0 {
			t.Errorf("%s is auto and has no argv, so it can never run", r.ID)
		}
		if r.Kind == NeedsHuman && len(r.Argv) != 0 {
			t.Errorf("%s needs a human and also carries argv, which is a "+
				"contradiction: one of the two is wrong", r.ID)
		}
	}
}

// TestEveryNonVerbatimExpectationIsJustified is the rot-catcher.
//
// ⚠️ "This one is supposed to fail" is the single easiest claim to leave
// behind after the reason stops being true. An expectation of anything other
// than VERBATIM is a standing assertion about the world, so it must carry
// why, in the table, next to it.
func TestEveryNonVerbatimExpectationIsJustified(t *testing.T) {
	for _, r := range routes {
		if r.Expect == Verbatim {
			continue
		}
		if strings.TrimSpace(r.Because) == "" {
			t.Errorf("%s expects %s and says nothing about why. An "+
				"unexplained expected failure becomes an accepted failure",
				r.ID, r.Expect)
		}
	}
}

func TestTheBaselineRouteExistsAndExpectsVerbatim(t *testing.T) {
	var found bool
	for _, r := range routes {
		if r.ID != "shellexecute" {
			continue
		}
		found = true
		if r.Expect != Verbatim {
			t.Errorf("the baseline expects %s. If the OS hop itself is not "+
				"expected to deliver the bytes, every other row in the table "+
				"is measuring something else", r.Expect)
		}
		if r.Kind != Auto {
			t.Errorf("the baseline is %s, so an unattended run has nothing "+
				"to compare the other routes against", r.Kind)
		}
	}
	if !found {
		t.Fatal("no shellexecute route, so there is no baseline and a " +
			"MANGLED result cannot be distinguished from a broken install")
	}
}

// TestTheLinkCarriesTheRouteOutsideTheFragment pins the design decision in
// linkFor, because getting it wrong produces a plausible-looking matrix.
//
// ⚠️ If the route id lived ONLY in the fragment, a route that strips
// fragments would deliver a URI the probe could not attribute, and would be
// reported as ABSENT. A silent-loss verdict on a route that actually
// delivered is the worst single error this program could make.
func TestTheLinkCarriesTheRouteOutsideTheFragment(t *testing.T) {
	uri := linkFor("some-route")
	head, frag, ok := strings.Cut(uri, "#")
	if !ok {
		t.Fatal("linkFor produced no fragment, so fragment-stripping " +
			"cannot be detected at all")
	}
	if !strings.Contains(head, "some-route") {
		t.Errorf("the route id is not in the identity-bearing part of the "+
			"link:\n  %s\nA route that strips the fragment would be "+
			"reported as ABSENT rather than MANGLED.", head)
	}
	if !strings.Contains(frag, "some-route") {
		t.Errorf("the route id is not in the fragment either, so fragment "+
			"loss produces no visible difference: %s", frag)
	}
}

func TestFillSubstitutesEveryPlaceholder(t *testing.T) {
	r := Route{Argv: []string{"{{shimp}}", "dispatch", "{{uri}}", "{{python}}"}}
	got := r.Fill(map[string]string{
		"{{uri}}":    "shimp://x",
		"{{shimp}}":  "S.exe",
		"{{python}}": "P.exe",
	})
	want := []string{"S.exe", "dispatch", "shimp://x", "P.exe"}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("argv[%d] = %q, want %q", i, got[i], want[i])
		}
	}
	for _, a := range got {
		if strings.Contains(a, "{{") {
			t.Errorf("unsubstituted placeholder survived into argv: %q. It "+
				"would be passed to the OS literally and the route would "+
				"fail for a reason that looks like a shimp bug", a)
		}
	}
}

// TestArrivalIsFoundInTheLog drives the classifier over a synthetic log.
func TestArrivalIsFoundInTheLog(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, logName)

	older := "click: console present=true visible=true foreground=true\n" +
		"click: os.Args[1] is 9 bytes\n" +
		"click: shimp://old\n"
	if err := os.WriteFile(p, []byte(older), 0o644); err != nil {
		t.Fatal(err)
	}
	off := logSize(p)

	// ⚠️ Nothing new yet, so nothing must be found. A classifier that
	// returned the PREVIOUS route's arrival would report every route as
	// VERBATIM, which is the exact shape of a check that is green by being
	// blind.
	if got, ok := arrivalSince(p, off); ok {
		t.Fatalf("found an arrival %q with nothing appended since the "+
			"offset, so every route would inherit the previous one's result",
			got)
	}

	appended := "click: os.Args[1] is 10 bytes\nclick: shimp://new\n"
	if err := os.WriteFile(p, []byte(older+appended), 0o644); err != nil {
		t.Fatal(err)
	}
	got, ok := arrivalSince(p, off)
	if !ok {
		t.Fatal("an appended click was not seen, so every route would " +
			"report ABSENT")
	}
	if got != "shimp://new" {
		t.Errorf("got %q, want the newly appended URI. The byte-count line "+
			"also starts with the click prefix, and picking it instead "+
			"would make every comparison report MANGLED", got)
	}
}

func TestArrivalIgnoresALogThatDoesNotExist(t *testing.T) {
	if _, ok := arrivalSince(filepath.Join(t.TempDir(), "nope.log"), 0); ok {
		t.Error("claimed an arrival from a log that does not exist")
	}
	if n := logSize(filepath.Join(t.TempDir(), "nope.log")); n != 0 {
		t.Errorf("logSize of a missing file = %d, want 0. A non-zero offset "+
			"would hide the first real arrival", n)
	}
}

// TestTheProbeRefusesTheProductionScheme is the s48 guard, asserted by name.
func TestTheProbeRefusesTheProductionScheme(t *testing.T) {
	if testScheme == "shimp" {
		t.Fatal("testScheme is the production scheme. This program installs " +
			"and purges a real registry claim, so it would uninstall the " +
			"user's own handler")
	}
	if !strings.HasPrefix(testScheme, "shimp-") {
		t.Errorf("testScheme %q does not announce what it belongs to, so a "+
			"leftover registry key would be unattributable", testScheme)
	}
}

func TestFromLocalsReadsAKeyAndIgnoresComments(t *testing.T) {
	dir := t.TempDir()
	body := "# a comment\nVENV_PYTHON = \"C:/somewhere/python.exe\"\nOTHER=x\n"
	if err := os.WriteFile(filepath.Join(dir, "locals.txt"),
		[]byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := fromLocals(dir, "VENV_PYTHON"); got != "C:/somewhere/python.exe" {
		t.Errorf("fromLocals = %q, want the quoted value unwrapped", got)
	}
	if got := fromLocals(dir, "ABSENT_KEY"); got != "" {
		t.Errorf("fromLocals of a missing key = %q, want empty", got)
	}
	if got := fromLocals(t.TempDir(), "VENV_PYTHON"); got != "" {
		t.Errorf("fromLocals with no locals.txt = %q, want empty. ⚠️ This is "+
			"the normal case for anyone who has not set one up, and it must "+
			"fall through to the PATH search rather than fail", got)
	}
}

// TestNoAbsolutePathsInSource is AGENTS.md's rule, checked rather than
// remembered. The Python driver documents the venv as <venv>, so a real
// drive-letter path appearing here is a regression.
func TestNoAbsolutePathsInSource(t *testing.T) {
	for _, name := range []string{"main.go", "routes.go", "excel_route.py",
		"linkprobe_test.go"} {
		raw, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		// ⚠️ The needles are ASSEMBLED rather than written out, because the
		// first version of this check failed on its own pattern literals.
		// That is the check working, but a self-match makes it useless.
		var needles []string
		for _, drive := range []rune{'C', 'D', 'E'} {
			for _, sep := range []rune{'\\', '/'} {
				needles = append(needles, string(drive)+":"+string(sep))
			}
		}
		for i, line := range strings.Split(string(raw), "\n") {
			for _, d := range needles {
				// TestFromLocals deliberately holds a fake path as DATA.
				if strings.Contains(line, d) && !strings.Contains(line, "somewhere") {
					t.Errorf("%s:%d holds an absolute path, which breaks on "+
						"every other machine: %s", name, i+1, strings.TrimSpace(line))
				}
			}
		}
	}
}

func TestSourceHasNoStrayControlCharacters(t *testing.T) {
	// ⚠️ Measured, four times: a heredoc substitution turned \n into a real
	// newline and once embedded a chr(14) inside a registry key path, which
	// compiled silently because it sat in a raw string. This check is cheap
	// and the failure it catches is invisible.
	for _, name := range []string{"main.go", "routes.go", "excel_route.py",
		"linkprobe_test.go"} {
		raw, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		for i, b := range raw {
			if b < 0x20 && b != '\n' && b != '\r' && b != '\t' {
				t.Errorf("%s byte %d is control character 0x%02x", name, i, b)
			}
		}
	}
}

// TestEmitJSONIsStableAndNarrow pins the published surface.
//
// ⚠️ A test suite asserts on this shape, so it is a contract rather than a
// convenience. The fields are deliberately NOT Result's, because Result
// embeds the whole Route including its prose, and a consumer must not break
// when somebody rewords a comment.
func TestEmitJSONIsStableAndNarrow(t *testing.T) {
	results := []Result{
		{
			Route: Route{ID: "a", Kind: Auto, Expect: Verbatim},
			Got:   Verbatim, Sent: "shimp://x", Received: "shimp://x",
		},
		{
			Route: Route{ID: "b", Kind: Auto, Expect: Verbatim},
			Got:   Mangled, Sent: "shimp://y", Received: "shimp://y/",
		},
		{
			Route: Route{ID: "c", Kind: NeedsHuman, Expect: Manual},
			Got:   Manual, Sent: "shimp://z",
		},
	}
	var buf bytes.Buffer
	if err := emitJSON(&buf, results); err != nil {
		t.Fatal(err)
	}
	var got jsonReport
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("our own output does not parse, so no consumer could use "+
			"it: %v\n%s", err, buf.String())
	}
	if got.Scheme != testScheme {
		t.Errorf("scheme = %q, want %q", got.Scheme, testScheme)
	}
	// ⚠️ MANUAL IS NOT COUNTED AS RUN, and is not a finding. A matrix that
	// counted unanswered routes as passing would read as complete when it
	// is not, which is the single most misleading thing this report could
	// do.
	if got.Run != 2 {
		t.Errorf("run = %d, want 2: the MANUAL route must not count as run", got.Run)
	}
	if got.Manual != 1 {
		t.Errorf("manual = %d, want 1", got.Manual)
	}
	if got.Findings != 1 {
		t.Errorf("findings = %d, want 1 (only the MANGLED route)", got.Findings)
	}
	if len(got.Results) != 3 {
		t.Fatalf("results = %d, want 3", len(got.Results))
	}
	if got.Results[1].Received != "shimp://y/" {
		t.Errorf("the mangled route lost its received bytes, which are the "+
			"whole point of reporting it: %+v", got.Results[1])
	}
	if got.Results[0].SentBytes != len("shimp://x") {
		t.Errorf("sent_bytes = %d, want %d",
			got.Results[0].SentBytes, len("shimp://x"))
	}
}

// TestJSONFieldNamesAreTheContract fails loudly if a field is renamed.
//
// ⚠️ Renaming a JSON field silently breaks every consumer, and nothing else
// in this repo would notice. Listed explicitly so the break is here, in a
// test with a message, rather than in somebody's CI a week later.
func TestJSONFieldNamesAreTheContract(t *testing.T) {
	var buf bytes.Buffer
	if err := emitJSON(&buf, []Result{{
		Route: Route{ID: "a", Kind: Auto, Expect: Verbatim},
		Got:   Verbatim, Sent: "s", Received: "s", Detail: "d",
	}}); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`"scheme"`, `"routes"`, `"run"`, `"findings"`, `"manual"`,
		`"results"`, `"id"`, `"kind"`, `"expect"`, `"got"`,
		`"as_expected"`, `"sent_bytes"`, `"sent"`, `"received"`, `"detail"`,
	} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("field %s is missing from the output. If it was renamed "+
				"on purpose, every consumer needs telling; this test is the "+
				"telling.\n%s", want, buf.String())
		}
	}
}
