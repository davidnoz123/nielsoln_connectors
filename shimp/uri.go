// uri.go -- parse, VALIDATE and hash a shimp URI. No build tag, no OS.
//
// The grammar frozen in the slate:
//
//	shimp://v1/go/run/<module>@<40 hex>[/<arg>...][#<action>]
//
// ⚠️ THE CENTRAL RULE IS REFUSE, NEVER NORMALISE, which is s35. Nothing here
// lower-cases, percent-decodes, NFC-folds or strips a trailing slash in order
// to make a URI acceptable. Anything non-canonical is REJECTED.
//
// The reason is an asymmetry worth stating plainly, because the opposite
// choice looks more helpful: a normaliser that is subtly wrong maps two
// spellings onto ONE identity silently, which is the single thing a singleton
// exists to prevent. A validator that is subtly wrong rejects a GOOD link
// loudly. Only one of those failures announces itself.
//
// And it is that rule which makes s44's hashing correct. Identity is the
// URI's bytes taken VERBATIM, with no transformation at all, and that is only
// safe because by the time any byte is hashed every other spelling has
// already been refused.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// runners is s43's closed TABLE.
//
// ⚠️ A table rather than a parse, because a left-to-right parser cannot find
// where the runner ends: a module path has slashes too and nothing in the
// text marks the boundary. The table says how many segments each runner
// consumes, and an unrecognised one FAILS CLOSED, which is the rule already
// chosen for `version`.
//
// And s43 collapses s33's launch rule into this: the runner's segments ARE
// the leading argv, literally and in order. There is no implied verb to
// prepend and so nothing to get wrong.
var runners = map[string][]string{
	"go": {"go", "run"},
}

// runnerSegments is how many path segments a runner name occupies, including
// its own name. Only `go/run` today, so two.
var runnerSegments = map[string]int{
	"go": 2,
}

// shaLen is s37: exactly one legal length, and it is 40.
//
// ⚠️ Not "at least": EXACTLY. Under s35 identity is the hash of the URI
// bytes, so `@bdc56ca` and the full SHA would be two identities for one
// commit, hence two singletons. Reconciling them by expanding the
// abbreviation would require asking the module proxy, which makes identity
// depend on network state and on what the repo holds at the moment of the
// click.
//
// A named constant deliberately, because s37 is RAISED rather than DECIDED:
// it waits on one measurement, what the toolchain does with an AMBIGUOUS
// abbreviation. Permitting a shorter form later ADDS a legal spelling and
// breaks nothing; starting shorter is unrecoverable. So 40 is the only
// direction that can be revised.
const shaLen = 40

// grammarVersion is the only version this build interprets. s19: an unknown
// version fails closed rather than being interpreted hopefully.
const grammarVersion = "v1"

// URI is a parsed, validated shimp link.
type URI struct {
	// Raw is the identity-bearing prefix, VERBATIM: every byte from the
	// start up to but excluding the first "#". s44.
	Raw string
	// Version, Runner, Module, Sha are the decoded identity parts, for
	// humans and for building argv. They are never re-serialised into Raw.
	Version string
	Runner  string
	Module  string
	Sha     string
	// Args are the percent-DECODED argv items, in order. Decoding happens
	// here and never affects Raw.
	Args []string
	// Action is the fragment with its leading "#" removed, or empty when
	// there was no "#" at all. Not part of identity. s41.
	Action string
	// HasAction distinguishes "shimp://...#" from "shimp://..." . The first
	// is an empty action, the second is none at all.
	HasAction bool
}

// Identity is s44: sha256 of the identity bytes, hex.
//
// The test vector pinned in uri_test.go is the whole point of this being a
// method over Raw rather than over the parsed parts: it hashes what arrived.
func (u *URI) Identity() string {
	sum := sha256.Sum256([]byte(u.Raw))
	return hex.EncodeToString(sum[:])
}

// Argv is s43's launch rule: the runner's segments, then the target, then the
// descriptor's argv, verbatim and in order. Nothing is quoted or joined, and
// there is no shell anywhere.
func (u *URI) Argv() []string {
	out := append([]string{}, runners[u.Runner]...)
	out = append(out, u.Module+"@"+u.Sha)
	return append(out, u.Args...)
}

// ParseURI parses and validates, or refuses with a reason a human can act on.
//
// `want` is the scheme this handler is installed as. It is a parameter rather
// than a read of the package variable so that a test can parse a link for
// another installation without mutating global state.
func ParseURI(raw, want string) (*URI, error) {
	// -- the scheme, exactly, lower-case
	prefix := want + "://"
	if !strings.HasPrefix(raw, prefix) {
		return nil, fmt.Errorf("not a %s:// URI. The scheme is part of the "+
			"identity bytes (s44), so a different one is a different "+
			"protocol rather than a typo", want)
	}

	// -- split identity from action at the FIRST "#", and refuse a second
	//
	// ⚠️ The split is on the first "#" and the fragment keeps everything
	// after it, including any "?" (s41). A SECOND "#" is refused: RFC 3986
	// would fold it into the fragment, so two spellings would differ in
	// bytes while meaning the same action, and the identity prefix is what
	// gets hashed.
	ident, action := raw, ""
	hasAction := false
	if i := strings.Index(raw, "#"); i >= 0 {
		ident, action, hasAction = raw[:i], raw[i+1:], true
		if strings.Contains(action, "#") {
			return nil, fmt.Errorf("more than one '#'. The first starts " +
				"the action, and a second would be folded into the fragment " +
				"by RFC 3986, so two spellings would differ in bytes while " +
				"meaning the same action")
		}
	}

	// -- s41: no query before the "#", any number after it
	if strings.Contains(ident, "?") {
		return nil, fmt.Errorf("a '?' appears before the '#'. A query is " +
			"reorderable and its repeated-key order is parser-dependent, so " +
			"it cannot sit in the bytes that are hashed. After the '#' it is " +
			"fine, because the action is not part of identity (s41)")
	}

	u := &URI{Raw: ident, Action: action, HasAction: hasAction}

	// -- segments
	rest := strings.TrimPrefix(ident, prefix)
	if strings.HasSuffix(rest, "/") {
		return nil, fmt.Errorf("a trailing '/' before the '#' or the end. " +
			"One spelling only, so it is refused rather than trimmed (s35)")
	}
	segs := strings.Split(rest, "/")
	for i, s := range segs {
		if s == "" {
			return nil, fmt.Errorf("segment %d is empty, so the path "+
				"contains %q. One spelling only (s39, s45)", i+1, "//")
		}
		if err := checkEncoding(s); err != nil {
			return nil, fmt.Errorf("segment %d (%q): %w", i+1, s, err)
		}
	}

	// -- the version, which must come first and must be known
	if len(segs) < 1 {
		return nil, fmt.Errorf("no grammar version segment")
	}
	u.Version = segs[0]
	if u.Version != grammarVersion {
		// ⚠️ FAILS CLOSED, s19. Not "assume v1": a future grammar may give
		// the same bytes a different meaning, and guessing is how a link
		// written for v2 gets run as v1.
		return nil, fmt.Errorf("grammar version %q is not one this handler "+
			"interprets (it knows %q). Unknown versions fail closed rather "+
			"than being interpreted hopefully (s19)", u.Version, grammarVersion)
	}
	segs = segs[1:]

	// -- the runner, from s43's table
	if len(segs) == 0 {
		return nil, fmt.Errorf("no runner segment after the version")
	}
	n, ok := runnerSegments[segs[0]]
	if !ok {
		return nil, fmt.Errorf("runner %q is not one this handler knows. "+
			"Unknown runners fail closed, which is why s3 insisted the "+
			"field exist before the first link was ever written", segs[0])
	}
	if len(segs) < n {
		return nil, fmt.Errorf("runner %q occupies %d segments and only %d "+
			"are present", segs[0], n, len(segs))
	}
	// Every segment the runner claims must match the table, so `go/walk`
	// cannot pass as `go/run`.
	expect := runners[segs[0]]
	for i := 0; i < n; i++ {
		if segs[i] != expect[i] {
			return nil, fmt.Errorf("runner segment %d is %q, and %q expects "+
				"%q there", i+1, segs[i], segs[0], expect[i])
		}
	}
	u.Runner = segs[0]
	segs = segs[n:]

	// -- the target: everything up to and including the segment holding "@"
	if len(segs) == 0 {
		return nil, fmt.Errorf("no target module after the runner")
	}
	at := -1
	for i, s := range segs {
		if strings.Contains(s, "@") {
			at = i
			break
		}
	}
	if at < 0 {
		return nil, fmt.Errorf("no %q in the path, so no commit is pinned. A "+
			"branch or a tag would make one link mean different code on "+
			"different days", "@")
	}
	last := segs[at]
	if strings.Count(last, "@") != 1 {
		return nil, fmt.Errorf("segment %q contains more than one %q",
			last, "@")
	}
	cut := strings.IndexByte(last, '@')
	modTail, sha := last[:cut], last[cut+1:]
	if modTail == "" {
		return nil, fmt.Errorf("nothing before the %q, so the module path is "+
			"incomplete", "@")
	}
	u.Module = strings.Join(append(append([]string{}, segs[:at]...), modTail), "/")
	// A module path needs a host, and a host has a dot. Cheap, local, and it
	// catches a pasted path that lost its prefix.
	if !strings.Contains(strings.Split(u.Module, "/")[0], ".") {
		return nil, fmt.Errorf("module %q does not start with a host, so it "+
			"is not a module path go can resolve", u.Module)
	}
	if err := checkSha(sha); err != nil {
		return nil, err
	}
	u.Sha = sha
	segs = segs[at+1:]

	// -- argv, percent-decoded. s45.
	for i, s := range segs {
		dec, err := decodeSegment(s)
		if err != nil {
			return nil, fmt.Errorf("argument %d (%q): %w", i+1, s, err)
		}
		u.Args = append(u.Args, dec)
	}
	return u, nil
}

// checkSha is s37.
func checkSha(sha string) error {
	if len(sha) != shaLen {
		return fmt.Errorf("the commit is %d characters and must be exactly "+
			"%d. An abbreviation would be re-resolved by the toolchain on "+
			"every click for the life of the link, so a commit pushed next "+
			"year that shares a prefix would change what this link means "+
			"(s37)", len(sha), shaLen)
	}
	for i := 0; i < len(sha); i++ {
		c := sha[i]
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return fmt.Errorf("the commit contains %q at position %d. "+
				"Lower-case hex only: upper case is a second spelling of the "+
				"same commit and so a second identity (s35)", string(c), i)
		}
	}
	return nil
}

// unreserved is RFC 3986's set, which never needs encoding.
func unreserved(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') ||
		(c >= '0' && c <= '9') || c == '-' || c == '.' || c == '_' || c == '~'
}

// legalRaw reports whether a byte may appear UNENCODED in a path segment.
//
// RFC 3986: segment = *pchar, and pchar = unreserved / pct-encoded /
// sub-delims / ":" / "@". So the bytes that MUST be escaped are only these:
// "/" "?" "#" "[" "]" "%", space, the control range, and anything >= 0x80.
//
// ⚠️ THIS IS WIDER THAN `unreserved` AND THAT WAS A REAL DEFECT. The escape
// check originally refused an escape only when it encoded an UNRESERVED
// character, so `%3A` was accepted alongside a raw ":" -- and twelve
// characters had two legal spellings each: : @ $ & + , ; = ! ( * '
//
// Two spellings is two identities for one capability, which is the single
// thing s35 exists to prevent, and it was found by trying to write a real
// link: `-listen 127.0.0.1:8080` is an ordinary argument and it hit it
// immediately. A test now pins every one of the twelve.
func legalRaw(c byte) bool {
	if unreserved(c) {
		return true
	}
	switch c {
	// sub-delims
	case '!', '$', '&', '\'', '(', ')', '*', '+', ',', ';', '=':
		return true
	// explicitly permitted in a segment
	case ':', '@':
		return true
	}
	return false
}

// checkEncoding enforces s45's canonical percent-encoding on one segment.
//
// ⚠️ Two rules, and the second is the one that matters for identity:
// uppercase hex only, and NO UNNECESSARY ESCAPE. "%2D" for "-" is a second
// spelling of the same argument, so under s44 it is a second identity, which
// is precisely what refuse-never-normalise exists to prevent.
//
// This is also the check s39 deliberately did NOT delegate to slate_tools,
// because knowing what MUST be encoded is knowing this grammar, and asking
// another repo for it would make that repo a second implementation of it.
func checkEncoding(seg string) error {
	for i := 0; i < len(seg); i++ {
		c := seg[i]
		if c != '%' {
			// A raw character that should have been encoded. "/" cannot
			// appear, having been consumed by the split, and "#" and "?" are
			// refused earlier, so this catches space and the rest.
			if c <= 0x20 || c == 0x7f {
				return fmt.Errorf("contains a raw control or space character "+
					"at %d, which must be percent-encoded", i)
			}
			continue
		}
		if i+2 >= len(seg) {
			return fmt.Errorf("a %q at %d with fewer than two hex digits "+
				"after it", "%", i)
		}
		hi, lo := seg[i+1], seg[i+2]
		if !isUpperHex(hi) || !isUpperHex(lo) {
			return fmt.Errorf("the escape %q at %d is not two UPPER-CASE hex "+
				"digits. Lower case is a second spelling and so a second "+
				"identity (s35, s45)", seg[i:i+3], i)
		}
		v := unhex(hi)<<4 | unhex(lo)
		if legalRaw(v) {
			return fmt.Errorf("the escape %q at %d encodes %q, which is legal "+
				"unencoded in a path segment. An unnecessary escape is a "+
				"second spelling of the same value, so it is refused rather "+
				"than accepted (s45)", seg[i:i+3], i, string(v))
		}
		i += 2
	}
	return nil
}

func isUpperHex(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'A' && c <= 'F')
}

func unhex(c byte) byte {
	if c >= '0' && c <= '9' {
		return c - '0'
	}
	return c - 'A' + 10
}

// decodeSegment percent-decodes one already-validated segment.
//
// ⚠️ Decoding happens AFTER validation and never feeds back into Raw, which
// is what keeps s44's verbatim hashing honest. And it refuses an empty
// result: s45 records that an empty-string argument cannot be expressed,
// because a path segment cannot be empty, and that is a refusal rather than
// a gap so the first person who needs one gets a reason instead of a
// silently dropped argument.
func decodeSegment(seg string) (string, error) {
	var b strings.Builder
	for i := 0; i < len(seg); i++ {
		if seg[i] != '%' {
			b.WriteByte(seg[i])
			continue
		}
		b.WriteByte(unhex(seg[i+1])<<4 | unhex(seg[i+2]))
		i += 2
	}
	out := b.String()
	if out == "" {
		return "", fmt.Errorf("decodes to the empty string, which a path " +
			"segment cannot express (s45)")
	}
	// ⚠️ Dot segments are refused DECODED as well as raw. A literal ".." is
	// subject to RFC 3986 dot-segment removal by anything that touches the
	// URI, so an un-encoded relative path can arrive silently different from
	// the one written. s45 records that as the correctness cost, larger than
	// the length one.
	if out == "." || out == ".." {
		return "", fmt.Errorf("is a dot segment. A literal one is removed by " +
			"RFC 3986 path normalisation, so the link would arrive different " +
			"from the one written (s45)")
	}
	return out, nil
}
