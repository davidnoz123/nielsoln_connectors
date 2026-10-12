package main

import (
	"strings"
	"testing"
)

// TestOneSpellingOnlyForSubDelims pins the twelve characters that had two.
//
// WARN FOUND BY WRITING A REAL LINK, not by review. `-listen 127.0.0.1:8080`
// is an ordinary argument, and the raw ":" and "%3A" were both accepted while
// hashing differently. Two spellings is two identities for one capability,
// which is the one thing s35 exists to prevent.
//
// The cause: the escape check refused an escape only when it encoded an
// UNRESERVED character, when it had to refuse any escape of a byte that is
// legal RAW in a segment. RFC 3986's pchar includes sub-delims and ":" and
// "@", which is twelve characters wider than `unreserved`.
func TestOneSpellingOnlyForSubDelims(t *testing.T) {
	sha := strings.Repeat("a", 40)
	base := "shimp://v1/go/run/example.com/x@" + sha + "/-listen/"
	// Every character that is LEGAL RAW in a path segment. If its escape is
	// also accepted, that is two spellings and therefore two identities.
	for _, c := range []struct{ raw, esc string }{
		{":", "%3A"}, {"@", "%40"}, {"$", "%24"}, {"&", "%26"},
		{"+", "%2B"}, {",", "%2C"}, {";", "%3B"}, {"=", "%3D"},
		{"!", "%21"}, {"(", "%28"}, {"*", "%2A"}, {"'", "%27"},
	} {
		// The raw form must be ACCEPTED: these are legal in a segment.
		if _, err := ParseURI(base+c.raw, "shimp"); err != nil {
			t.Errorf("a raw %q was refused, and it is legal in a path "+
				"segment: %v", c.raw, err)
		}
		// And the escaped form must be REFUSED, because it is a second
		// spelling of the same argument.
		if _, err := ParseURI(base+c.esc, "shimp"); err == nil {
			a, _ := ParseURI(base+c.raw, "shimp")
			b, _ := ParseURI(base+c.esc, "shimp")
			t.Errorf("%q and %q are BOTH accepted and hash differently "+
				"(%s vs %s): two spellings, two identities, one capability",
				c.raw, c.esc, a.Identity()[:12], b.Identity()[:12])
		}
	}
	// And the escapes that ARE required stay required.
	for _, esc := range []string{"%2F", "%23", "%3F", "%5B", "%5D", "%20", "%25"} {
		if _, err := ParseURI(base+"a"+esc+"b", "shimp"); err != nil {
			t.Errorf("the REQUIRED escape %s was refused: %v", esc, err)
		}
	}
}
