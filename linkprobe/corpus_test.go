// corpus_test.go -- the checks on the verdict reader, which is where the one
// inverting mistake lives.
package main

import (
	"strings"
	"testing"
)

const validTail = "click: console present=true visible=true foreground=true\n" +
	"click: os.Args[1] is 118 bytes\n" +
	"click: shimp-x://v1/go/run/example.com/m/chrome_ready@abc\n" +
	"click: identity deadbeefdeadbeefdeadbeefdeadbeef over 118 bytes\n" +
	"click: would run [\"go\" \"run\" \"example.com/m/chrome_ready@abc\"]\n" +
	"click: no action, so this is ensure-running only\n" +
	"click REFUSED: the KEEPER is not built. The link is valid and its " +
	"identity is above\n"

// TestAValidLinkIsNotReadAsRejected is THE test on this file.
//
// ⚠️ A valid link ends in a refusal today, because the keeper is not built.
// Keying acceptance on the absence of the word REFUSED would classify every
// correct link as rejected and every rejection as correct -- a check that is
// not merely blind but exactly inverted.
func TestAValidLinkIsNotReadAsRejected(t *testing.T) {
	saw := readHandler(validTail)
	if saw.Verdict != Accepted {
		t.Fatalf("a link the handler PARSED was read as %s. Every corpus "+
			"result would be inverted.\nidentity=%q refusal=%q",
			saw.Verdict, saw.Identity, saw.Refusal)
	}
	if !saw.KeeperRefusal {
		t.Error("the keeper refusal was not recognised as such, so the day " +
			"the keeper lands nothing here will notice the change")
	}
	if saw.IdentityBytes != 118 {
		t.Errorf("identity bytes = %d, want 118", saw.IdentityBytes)
	}
	if !strings.Contains(saw.ArgvLog, "chrome_ready@abc") {
		t.Errorf("argv not captured: %q", saw.ArgvLog)
	}
	if !strings.Contains(saw.Raw, "://") {
		t.Errorf("raw URI not captured: %q", saw.Raw)
	}
}

func TestAParserRefusalIsReadAsRejected(t *testing.T) {
	body := "click: console present=true visible=true foreground=true\n" +
		"click: os.Args[1] is 117 bytes\n" +
		"click: shimp-x://v1/go/run/example.com/m/x@abc\n" +
		"click REFUSED: the commit is 39 characters and must be exactly 40\n"
	saw := readHandler(body)
	if saw.Verdict != HandlerRefused {
		t.Fatalf("a parser refusal was read as %s", saw.Verdict)
	}
	if saw.KeeperRefusal {
		t.Error("a parser refusal was mistaken for the keeper refusal, which " +
			"would let a rejected shape pass as accepted")
	}
	if saw.Identity != "" {
		t.Errorf("a refused link reported an identity %q, which it cannot "+
			"have: the parser never got that far", saw.Identity)
	}
}

func TestNothingInTheLogIsNotSeen(t *testing.T) {
	if saw := readHandler(""); saw.Verdict != NotSeen {
		t.Errorf("an empty log gave %s, want NOT-SEEN. Anything else would "+
			"credit the handler with a decision it never made", saw.Verdict)
	}
}

// TestJudgeCatchesAnAcceptedShapeThatWasRefused and its mirror are the
// can-fail companions: each check is shown failing on a synthetic case.
func TestJudgeCatchesAnAcceptedShapeThatWasRefused(t *testing.T) {
	s := Shape{ID: "x", Tail: "a/b", Expect: Accepted}
	res := judge(s, HandlerSaw{Verdict: HandlerRefused, Refusal: "nope"}, "shimp-x")
	if res.OK() {
		t.Fatal("a shape expecting ACCEPTED passed while being refused")
	}
}

func TestJudgeCatchesARefusalForTheWrongReason(t *testing.T) {
	s := Shape{ID: "x", Tail: "a/b", Expect: HandlerRefused,
		Reason: "must be exactly 40"}
	res := judge(s, HandlerSaw{Verdict: HandlerRefused,
		Refusal: "segment 9: some unrelated complaint"}, "shimp-x")
	if res.OK() {
		t.Fatal("a link refused for the WRONG reason passed. Asserting only " +
			"that a link was refused is a check that is green while blind")
	}
	if !strings.Contains(strings.Join(res.Problems, " "), "wrong reason") {
		t.Errorf("the problem does not say what disagreed: %v", res.Problems)
	}
}

// ⚠️ The subtlest one. A shape expecting rejection that instead PARSED will
// carry the keeper refusal, and keying on "was there a refusal" would pass it.
func TestJudgeCatchesARejectedShapeThatActuallyParsed(t *testing.T) {
	s := Shape{ID: "bad", Tail: "a/b", Expect: HandlerRefused, Reason: "40"}
	saw := readHandler(validTail)
	res := judge(s, saw, "shimp-x")
	if res.OK() {
		t.Fatal("a shape expecting rejection PARSED and still passed. The " +
			"keeper refusal was mistaken for a parser refusal, which is the " +
			"one mistake that makes this whole table meaningless")
	}
}

func TestJudgeCatchesAWrongIdentityByteCount(t *testing.T) {
	s := Shape{ID: "x", Tail: "a/b#frag", Expect: Accepted}
	// "shimp-x://a/b" is 13 bytes; the fragment must be excluded (s44).
	if want := s.ExpectedIdentityBytes("shimp-x"); want != 13 {
		t.Fatalf("ExpectedIdentityBytes = %d, want 13 (everything before #)", want)
	}
	res := judge(s, HandlerSaw{Verdict: Accepted, Identity: "d", IdentityBytes: 21,
		Raw: s.Link("shimp-x")}, "shimp-x")
	if res.OK() {
		t.Fatal("an identity computed over the WHOLE link, fragment included, " +
			"passed. s44 is then unchecked")
	}
}

func TestEveryShapeIsWellFormed(t *testing.T) {
	seen := map[string]bool{}
	for _, s := range shapes {
		if seen[s.ID] {
			t.Errorf("duplicate shape id %q", s.ID)
		}
		seen[s.ID] = true
		if s.Label == "" || s.Because == "" {
			t.Errorf("%s lacks a label or a reason for existing", s.ID)
		}
		// ⚠️ The rot-catcher, same discipline as routes.go: a shape
		// expecting rejection must say on what grounds, or it passes for a
		// link refused for an unrelated reason.
		if s.Expect == HandlerRefused && strings.TrimSpace(s.Reason) == "" {
			t.Errorf("%s expects rejection and names no reason", s.ID)
		}
		if s.Expect == Accepted && s.Argv == nil {
			t.Errorf("%s expects acceptance but declares no argv, so s43 "+
				"goes unchecked for it", s.ID)
		}
		if strings.Contains(s.Tail, "://") {
			t.Errorf("%s hard-codes a scheme in its tail, so it cannot be "+
				"run against an isolated installation", s.ID)
		}
	}
}

func TestTheCorpusCoversBothVerdicts(t *testing.T) {
	var acc, ref int
	for _, s := range shapes {
		switch s.Expect {
		case Accepted:
			acc++
		case HandlerRefused:
			ref++
		}
	}
	if acc == 0 || ref == 0 {
		t.Fatalf("the corpus has %d accepted and %d refused shapes. With "+
			"either at zero it measures only half of what the handler does, "+
			"which is how the parser rules went unchecked until a human "+
			"clicked them", acc, ref)
	}
}
