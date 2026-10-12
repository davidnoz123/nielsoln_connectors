// uri_test.go -- the grammar, pinned.
//
// ⚠️ THE FIRST TEST IS s44's TEST VECTOR AND IT IS THE IMPORTANT ONE. The
// identity rule is the single decision in this design with no migration path:
// a version segment lets the grammar change, and nothing lets the hashing
// change, because every id that ever existed was computed under it. The row
// carries a worked example precisely so a first implementation can prove
// itself, and this is that proof.
//
// The rest of the file is one case per refusal, because under s35 the
// refusals ARE the grammar. A validator that accepts a second spelling is not
// a lenient validator, it is a second identity for one capability.
package main

import (
	"strings"
	"testing"
)

// -- the vector ---------------------------------------------------------------

func TestIdentityMatchesTheSlatesTestVector(t *testing.T) {
	// Copied from s44, which computed it on 10 Oct 2026.
	const (
		link = "shimp://v1/go/run/github.com/davidnoz123/nielsoln_connectors/" +
			"editor_open@1173327f58213898a33698949e49c5aa6d9a83e2/-editor/code"
		wantBytes = 126
		wantHash  = "7564ba535f3a1d65f96c219b29112e5195b4cd7c3c4f9c007e3b84fc5e34d9bc"
	)
	u, err := ParseURI(link, "shimp")
	if err != nil {
		t.Fatalf("the slate's own test vector does not parse: %v", err)
	}
	if len(u.Raw) != wantBytes {
		t.Fatalf("identity is %d bytes, s44 says %d", len(u.Raw), wantBytes)
	}
	if got := u.Identity(); got != wantHash {
		t.Fatalf("identity hash mismatch.\n  got  %s\n  want %s\n"+
			"s44 is the one decision with no migration path, so this is "+
			"wrong rather than the vector", got, wantHash)
	}
	// And the parts, because a hash matching by luck would still be wrong.
	if u.Module != "github.com/davidnoz123/nielsoln_connectors/editor_open" {
		t.Fatalf("module parsed as %q", u.Module)
	}
	if u.Sha != "1173327f58213898a33698949e49c5aa6d9a83e2" {
		t.Fatalf("sha parsed as %q", u.Sha)
	}
	if len(u.Args) != 2 || u.Args[0] != "-editor" || u.Args[1] != "code" {
		t.Fatalf("args parsed as %q", u.Args)
	}
	if u.HasAction {
		t.Fatal("a link with no '#' reported an action")
	}
}

// -- identity is independent of the action, which is the whole design --------

func TestTheActionIsNotPartOfIdentity(t *testing.T) {
	base := "shimp://v1/go/run/example.com/x/y@" + strings.Repeat("a", 40) +
		"/-p/1"
	a, err := ParseURI(base+"#select/page/17", "shimp")
	if err != nil {
		t.Fatal(err)
	}
	b, err := ParseURI(base+"#select/page/23", "shimp")
	if err != nil {
		t.Fatal(err)
	}
	none, err := ParseURI(base, "shimp")
	if err != nil {
		t.Fatal(err)
	}
	if a.Identity() != b.Identity() || a.Identity() != none.Identity() {
		t.Fatal("two links differing only after the '#' produced different " +
			"identities, so they would start two instances. s19 and s41: the " +
			"action addresses an instance, it does not define one.")
	}
	if a.Action != "select/page/17" || b.Action != "select/page/23" {
		t.Fatalf("the actions were not carried: %q and %q", a.Action, b.Action)
	}
	// An empty fragment is an action, and no fragment is not. The difference
	// is "activate" against "ensure running only".
	empty, err := ParseURI(base+"#", "shimp")
	if err != nil {
		t.Fatal(err)
	}
	if !empty.HasAction || empty.Action != "" {
		t.Fatal(`"#" with nothing after it should be an EMPTY action, not an absent one`)
	}
	if none.HasAction {
		t.Fatal("a link with no '#' should have no action at all")
	}
}

// -- s43's launch rule -------------------------------------------------------

func TestArgvIsTheRunnerSegmentsThenTargetThenArgs(t *testing.T) {
	u, err := ParseURI("shimp://v1/go/run/example.com/x@"+
		strings.Repeat("b", 40)+"/-root/%2Ftmp%2Fwork/-n/3", "shimp")
	if err != nil {
		t.Fatal(err)
	}
	got := u.Argv()
	want := []string{"go", "run", "example.com/x@" + strings.Repeat("b", 40),
		"-root", "/tmp/work", "-n", "3"}
	if len(got) != len(want) {
		t.Fatalf("argv is %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("argv[%d] is %q, want %q (full: %q)", i, got[i], want[i], got)
		}
	}
	// ⚠️ The encoded slashes decoded into ONE argument rather than splitting
	// it. That is the whole reason s45 requires %2F: a raw slash would have
	// become three arguments.
	if u.Args[1] != "/tmp/work" {
		t.Fatalf("the encoded path became %q", u.Args[1])
	}
}

// -- the refusals, which ARE the grammar -------------------------------------

func TestRefusals(t *testing.T) {
	sha := strings.Repeat("a", 40)
	base := "shimp://v1/go/run/example.com/x@" + sha

	cases := []struct{ name, link, because string }{
		{"wrong scheme", "other://v1/go/run/example.com/x@" + sha,
			"the scheme is in the identity bytes, so it is a different protocol"},
		{"no scheme", "v1/go/run/example.com/x@" + sha, ""},
		{"unknown grammar version",
			"shimp://v2/go/run/example.com/x@" + sha,
			"an unknown version must fail closed, not be read as v1"},
		{"unknown runner",
			"shimp://v1/rust/run/example.com/x@" + sha,
			"an unknown runner must fail closed"},
		{"runner second segment wrong",
			"shimp://v1/go/walk/example.com/x@" + sha,
			"go/walk must not pass as go/run"},
		{"abbreviated sha",
			"shimp://v1/go/run/example.com/x@1173327",
			"an abbreviation is re-resolved on every click for the life of the link"},
		{"41-char sha",
			"shimp://v1/go/run/example.com/x@" + sha + "a", ""},
		{"upper-case sha",
			"shimp://v1/go/run/example.com/x@" + strings.ToUpper(sha),
			"upper case is a second spelling of one commit, so a second identity"},
		{"no @ at all",
			"shimp://v1/go/run/example.com/x", "nothing pins a commit"},
		{"two @ in one segment",
			"shimp://v1/go/run/example.com/x@y@" + sha, ""},
		{"module without a host",
			"shimp://v1/go/run/notahost@" + sha,
			"a module path needs a host and a host has a dot"},
		{"query before the hash", base + "?a=1",
			"a query is reorderable, so it cannot sit in the hashed bytes"},
		{"two fragments", base + "#a#b",
			"RFC 3986 would fold the second in, giving two spellings"},
		{"trailing slash", base + "/",
			"one spelling only, refused rather than trimmed"},
		{"empty segment", "shimp://v1/go//run/example.com/x@" + sha, ""},
		{"lower-case escape", base + "/-root/%2ftmp",
			"lower-case hex is a second spelling"},
		{"unnecessary escape", base + "/%2Droot",
			"%2D encodes '-', which never needs encoding, so it is a second spelling"},
		{"truncated escape", base + "/-root/%2", ""},
		{"raw space", base + "/-root/a b", ""},
		{"dot-segment argument", base + "/%2E%2E",
			"a decoded dot segment is removed by path normalisation"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			u, err := ParseURI(c.link, "shimp")
			if err == nil {
				msg := "accepted a link it must refuse"
				if c.because != "" {
					msg += ": " + c.because
				}
				t.Fatalf("%s\n  link: %s\n  parsed as: %+v", msg, c.link, u)
			}
			// The message has to be actionable, which for this program means
			// naming a reason rather than a position. Checked loosely: the
			// exact wording is not the contract.
			if len(err.Error()) < 30 {
				t.Fatalf("the refusal is too terse to act on: %q", err)
			}
		})
	}
}

// -- the query that IS allowed, because the fragment is not hashed -----------

func TestAQueryIsLegalAfterTheHash(t *testing.T) {
	// s40's first capability needs exactly this, and s39's check 6 was
	// corrected for it.
	link := "shimp://v1/go/run/example.com/editor_open@" +
		strings.Repeat("c", 40) + "/-editor/code" +
		"#open?file=slate_tools%2Fslatex.py&line=1735"
	u, err := ParseURI(link, "shimp")
	if err != nil {
		t.Fatalf("a query in the FRAGMENT was refused, which would refuse the "+
			"first capability's own links (s40, s41): %v", err)
	}
	if u.Action != "open?file=slate_tools%2Fslatex.py&line=1735" {
		t.Fatalf("the action was altered: %q", u.Action)
	}
	// ⚠️ And the action is NOT decoded. It is forwarded as an opaque string,
	// per s20: what the target makes of it is the target's business.
	if strings.Contains(u.Action, "/slatex.py") {
		t.Fatal("the action was percent-decoded. It must be forwarded " +
			"verbatim as an opaque message (s20).")
	}
}

// -- a different installation's link is recognised, not mis-parsed ------------

func TestALinkForAnotherSchemeIsRefusedByName(t *testing.T) {
	link := "shimp-selftest://v1/go/run/example.com/x@" + strings.Repeat("a", 40)
	if _, err := ParseURI(link, "shimp"); err == nil {
		t.Fatal("a link for another installation was accepted")
	}
	// And it parses fine for the installation it belongs to, which is what
	// makes s48's isolated test scheme work at all.
	u, err := ParseURI(link, "shimp-selftest")
	if err != nil {
		t.Fatalf("the link was refused by its own scheme: %v", err)
	}
	// ⚠️ Different scheme, therefore different identity bytes, therefore a
	// different singleton. s44 and s48: a test click can never be answered
	// by a production instance.
	prod, err := ParseURI("shimp://v1/go/run/example.com/x@"+
		strings.Repeat("a", 40), "shimp")
	if err != nil {
		t.Fatal(err)
	}
	if u.Identity() == prod.Identity() {
		t.Fatal("two schemes produced the same identity, so a test " +
			"installation could be answered by a production instance")
	}
}
