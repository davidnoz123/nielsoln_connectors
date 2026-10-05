package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestWantsFocus pins the one rule that cost a human an afternoon.
//
// A server-wide -focus=false was the only way to keep a test run from
// stealing the desktop. One was left set, and every right-click afterwards
// navigated correctly in a BACKGROUND tab: no error, no log line saying
// anything was wrong, just focus never arriving and a page that looked like
// it had not been driven. Nothing in the system could tell the two apart,
// so nothing caught it.
//
// The rule is now per request, and this is what holds it there.
func TestWantsFocus(t *testing.T) {
	cases := []struct {
		name   string
		focus  bool
		client string
		query  string
		want   bool
	}{
		{"a human right-clicking is shown the page", true, "excel-vba", "", true},
		{"a test run leaves the desktop alone", true, "drive-test", "", false},

		// THE RULE INVERTED on 5 Oct, and these three are the reason.
		// It used to be `!= "drive-test"`, a deny-list with one entry, so
		// every caller that had not been taught to say that string raised
		// the window. A batch of 25 drives arrived from one that had not.
		//
		// The first case ASSERTED THE OPPOSITE until that day: a browser
		// click was shown the page. It is a real loss and it is kept
		// visible here rather than deleted, because a browser and a script
		// that forgot send exactly the same thing, which is nothing.
		{"a browser click no longer raises", true, "", "", false},
		{"a tool that forgot its name is quiet", true, "sweeper", "", false},
		{"and a typo in the name is quiet, not loud", true, "excel_vba", "", false},

		// The global switch still wins over everything: "leave my desktop
		// alone while I work" has to be absolute or it is not a switch.
		{"server off beats a human", false, "excel-vba", "", false},
		{"server off beats an explicit ask", false, "excel-vba", "focus=1", false},

		// An explicit ask beats the per-client default in both directions,
		// so anything can override what it would otherwise get.
		{"a test may ask for focus", true, "drive-test", "focus=1", true},
		{"a human may decline it", true, "excel-vba", "focus=0", false},
		// The way back for anything the allow-list leaves out, including a
		// browser: one request, said out loud, rather than a default.
		{"a browser may ask for it", true, "", "focus=1", true},
		{"so may an unnamed tool", true, "sweeper", "focus=1", true},
		{"false is spelled as well as numbered", true, "excel-vba", "focus=false", false},
		{"and case does not matter", true, "excel-vba", "focus=FALSE", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			url := "/show/tl-e7-09:before?t=tok"
			if c.query != "" {
				url += "&" + c.query
			}
			r := httptest.NewRequest(http.MethodGet, url, nil)
			if c.client != "" {
				r.Header.Set("X-Drive-Client", c.client)
			}
			s := &server{focus: c.focus}
			if got := s.wantsFocus(r); got != c.want {
				t.Fatalf("focus=%v client=%q query=%q: got %v, want %v",
					c.focus, c.client, c.query, got, c.want)
			}
		})
	}
}
