// corpus.go -- link SHAPES and what the handler should make of each.
//
// ⚠️ THE SECOND AXIS, AND THE HARNESS HAD ONLY ONE.
//
// routes.go asks "does this way in deliver the bytes?" and sends exactly one
// link shape down fifteen routes. That measures TRANSPORT. It says nothing
// about what the handler then does, so s37's forty-character rule and s45's
// one-spelling rule had only ever been unit-tested -- until a human clicked
// six hand-written links and found both working, which the harness could not
// have told him.
//
// This table is the other axis: many shapes, one route, asserting on the
// handler's VERDICT rather than on bytes arriving.
//
// ⚠️ ONE ROUTE, NOT FIFTEEN, and the reason matters. The handler's verdict
// depends only on the bytes it receives. routes.go already establishes which
// routes deliver bytes faithfully, so crossing the two tables would run
// ninety combinations to learn nothing the two axes do not already say
// separately. Six dispatches down the baseline is the whole cost.
//
// ⚠️ AND THE SERVED PAGE READS THIS SAME TABLE. `-serve` renders these shapes
// as clickable links for a human. A throwaway page was written by hand once,
// and within one turn it had drifted from what the harness checked. One
// table, two consumers, so the human view and the machine view cannot
// disagree.
package main

import (
	"fmt"
	"strings"
)

// Verdict is what the HANDLER made of a link, which is a different question
// from whether a route delivered it.
type Verdict string

const (
	// Accepted: parsed. The handler logged an identity and the argv it
	// derived.
	//
	// ⚠️ An accepted link STILL ends in a refusal today, because the keeper
	// is not built. So "accepted" is decided by the presence of an identity
	// line, never by the absence of the word REFUSED. Getting that backwards
	// would classify every valid link as rejected.
	Accepted Verdict = "ACCEPTED"

	// HandlerRefused: the parser said no, with a reason. s35's "refuse,
	// never normalise" is this outcome being reached loudly.
	HandlerRefused Verdict = "HANDLER-REFUSED"

	// NotSeen: nothing reached the handler at all.
	NotSeen Verdict = "NOT-SEEN"
)

// Shape is one link and what should become of it.
type Shape struct {
	ID string

	// Label is the human-facing name, used by -serve.
	Label string

	// Tail is everything after "<scheme>://", so the scheme can be injected
	// and no shape hard-codes which installation it is testing.
	Tail string

	Expect Verdict

	// Reason is a substring that must appear in the refusal. ⚠️ Required for
	// HandlerRefused: asserting only THAT a link was refused would pass for
	// a link refused for the wrong reason, which is the shape of a check
	// that is green while blind.
	Reason string

	// IdentityOver is the byte count the identity must be computed over, or
	// zero to skip. This is s44 checked end to end: everything from the
	// scheme to the byte before the first `#`.
	IdentityOver int

	// Argv is the argument vector the handler must derive, or nil to skip.
	// This is s43's runner table checked end to end: the runner consumes two
	// segments and the rest ARE the leading arguments.
	Argv []string

	// Because is why this shape is in the corpus.
	Because string
}

const corpusSha = "0123456789abcdef0123456789abcdef01234567"
const corpusMod = "v1/go/run/github.com/davidnoz123/nielsoln_connectors"

// Link renders the shape for a given scheme.
func (s Shape) Link(scheme string) string { return scheme + "://" + s.Tail }

// shapes is the corpus.
//
// ⚠️ PROVENANCE PER SHAPE, because "measured" and "predicted" are different
// claims and an earlier draft of this comment asserted the stronger one for
// all of them.
//
// Six were clicked BY HAND in a real browser on 15 Oct 2026 before being
// written down: plain, argv-and-action, action-with-arguments, long,
// bad-sha-39, double-encoded-colon. Those expectations are recorded
// observations.
//
// Two were WRITTEN AS PREDICTIONS and then confirmed by the first -corpus
// run: control-byte and unknown-runner. Weaker provenance, and worth
// distinguishing: nobody has clicked those in a browser, so they are known
// to hold through ShellExecuteW and nowhere else.
var shapes = []Shape{
	{
		ID:     "plain",
		Label:  "the plain case",
		Tail:   corpusMod + "/chrome_ready@" + corpusSha,
		Expect: Accepted,
		// 118 bytes with scheme "shimp-try". Computed rather than hard-coded
		// below, because the scheme length varies with the installation.
		IdentityOver: -1, // -1 means "the whole link", there being no fragment
		Argv: []string{"go", "run",
			"github.com/davidnoz123/nielsoln_connectors/chrome_ready@" + corpusSha},
		Because: "no action at all, so this is ensure-running only. The " +
			"simplest thing that must work.",
	},
	{
		ID:    "argv-and-action",
		Label: "the ChatGPT case: ensure a CDP Chrome on a named profile",
		Tail: corpusMod + "/chrome_ready@" + corpusSha +
			"/-port/9222/-profile/chrome_debug_profile#ensure",
		Expect:       Accepted,
		IdentityOver: -1,
		Argv: []string{"go", "run",
			"github.com/davidnoz123/nielsoln_connectors/chrome_ready@" + corpusSha,
			"-port", "9222", "-profile", "chrome_debug_profile"},
		Because: "⚠️ the shape the whole project is for. Measured: the argv " +
			"is derived exactly, and the action is logged as NOT part of " +
			"identity (s44).",
	},
	{
		ID:    "action-with-arguments",
		Label: "with action arguments",
		Tail: corpusMod + "/chrome_ready@" + corpusSha +
			"/-port/9222#ensure?wait=30&raise=true",
		Expect:       Accepted,
		IdentityOver: -1,
		Argv: []string{"go", "run",
			"github.com/davidnoz123/nielsoln_connectors/chrome_ready@" + corpusSha,
			"-port", "9222"},
		Because: "⚠️ the `?` case. explorer.exe DROPS this silently; measured " +
			"15 Oct, a browser click delivers it whole. The action arguments " +
			"are outside the identity, so a 26-byte action changes nothing " +
			"about which code runs.",
	},
	{
		ID:    "long",
		Label: "a long one, past Excel's formula threshold",
		Tail: corpusMod + "/chrome_ready@" + corpusSha +
			"/-port/9222/-pad/" + strings.Repeat("p", 190) + "#ensure",
		Expect:       Accepted,
		IdentityOver: -1,
		Argv: []string{"go", "run",
			"github.com/davidnoz123/nielsoln_connectors/chrome_ready@" + corpusSha,
			"-port", "9222", "-pad", strings.Repeat("p", 190)},
		Because: "332 bytes: past =HYPERLINK()'s 255 and Excel's 2084 object " +
			"cap is still far above. Clicked twice by hand and both gave the " +
			"same identity, which is the only evidence that it is a function " +
			"of the bytes and nothing else.",
	},
	{
		ID:     "bad-sha-39",
		Label:  "a BAD commit, 39 hex",
		Tail:   corpusMod + "/chrome_ready@" + corpusSha[:len(corpusSha)-1],
		Expect: HandlerRefused,
		Reason: "must be exactly 40",
		Because: "⚠️ s37, and the harness could not express this at all " +
			"before. An abbreviation would be re-resolved on every click for " +
			"the life of the link.",
	},
	{
		ID:     "double-encoded-colon",
		Label:  "a needlessly escaped character",
		Tail:   corpusMod + "/chrome_ready@" + corpusSha + "/-port/%3A9222",
		Expect: HandlerRefused,
		Reason: "second spelling of the same value",
		Because: "⚠️ s45, and this is legalRaw: twelve characters have two " +
			"legal spellings, and accepting both would give one link two " +
			"identities. Found by writing a real link, not by deriving " +
			"refusal cases.",
	},
	{
		ID:     "control-byte",
		Label:  "a control byte in a segment",
		Tail:   corpusMod + "/chrome_ready@" + corpusSha + "/-port/%091",
		Expect: HandlerRefused,
		Reason: "control",
		Because: "a tab, percent-encoded. ⚠️ Control bytes have bitten this " +
			"repo four times through heredocs, and a link carrying one must " +
			"be refused rather than passed to a shell.",
	},
	{
		ID:     "unknown-runner",
		Label:  "an unknown runner",
		Tail:   "v1/rust/run/github.com/davidnoz123/nielsoln_connectors/x@" + corpusSha,
		Expect: HandlerRefused,
		Reason: "runner",
		Because: "s43: the runner table must FAIL CLOSED. A runner nobody " +
			"implemented must be refused by name rather than ignored, or the " +
			"grammar silently grows.",
	},
}

// ExpectedIdentityBytes is how many bytes the identity should cover for a
// shape under a given scheme.
//
// ⚠️ COMPUTED, NOT TABULATED. Writing 159 into the table would pin the
// scheme's length as well as the rule, so renaming the test scheme would
// break every row for a reason having nothing to do with shimp. The rule is
// "everything before the first #", so that is what is expressed.
func (s Shape) ExpectedIdentityBytes(scheme string) int {
	if s.IdentityOver > 0 {
		return s.IdentityOver
	}
	link := s.Link(scheme)
	if i := strings.Index(link, "#"); i >= 0 {
		return i
	}
	return len(link)
}

// ExpectedArgvLog is the argv formatted exactly as the handler logs it, so
// the comparison is against the real line rather than a re-parse of it.
func (s Shape) ExpectedArgvLog() string {
	if s.Argv == nil {
		return ""
	}
	return fmt.Sprintf("%q", s.Argv)
}
