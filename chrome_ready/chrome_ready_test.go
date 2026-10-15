// chrome_ready_test.go -- what can be checked without a browser.
//
// ⚠️ Launching Chrome is not tested here, deliberately. It would start a real
// browser on whoever runs `go test ./...`, which is the machine-modifying
// side effect this repo keeps out of the loop. The probe half IS testable,
// and it is where the interesting mistakes live.
package main

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

// serveOn starts a server on a free port and returns the port.
func serveOn(t *testing.T, h http.Handler) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := &httptest.Server{Listener: ln, Config: &http.Server{Handler: h}}
	srv.Start()
	t.Cleanup(srv.Close)
	_, p, _ := net.SplitHostPort(ln.Addr().String())
	n, _ := strconv.Atoi(p)
	return n
}

func TestProbeReadsARealCDPAnswer(t *testing.T) {
	port := serveOn(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/json/version" {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte(`{"Browser":"Chrome/155.0.0.0",` +
			`"Protocol-Version":"1.3",` +
			`"webSocketDebuggerUrl":"ws://127.0.0.1/devtools/browser/x"}`))
	}))
	v, err := probe(port, 3*time.Second)
	if err != nil {
		t.Fatalf("a valid CDP answer was rejected: %v", err)
	}
	if v.Browser != "Chrome/155.0.0.0" || v.Protocol != "1.3" {
		t.Errorf("fields not read: %+v", v)
	}
}

// TestSomethingElseHoldingThePortIsNotReady is the one that matters.
//
// ⚠️ A port answering HTTP is not a port speaking CDP. Reporting READY here
// would hand a caller a port that cannot do what they asked, and the failure
// would surface later in somebody else's websocket code with no clue why.
func TestSomethingElseHoldingThePortIsNotReady(t *testing.T) {
	for _, c := range []struct{ name, body string }{
		{"a plain web server", `<html>hello</html>`},
		{"valid JSON that is not CDP", `{"status":"ok"}`},
		{"CDP-shaped but no Browser", `{"Protocol-Version":"1.3"}`},
	} {
		port := serveOn(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(c.body))
		}))
		if _, err := probe(port, 2*time.Second); err == nil {
			t.Errorf("%s was accepted as a ready Chrome", c.name)
		}
	}
}

// TestNothingListeningIsAnError, with TestProbeReadsARealCDPAnswer as its
// companion: that one shows probe SUCCEEDING, so a failure here means
// something rather than being the only outcome probe can produce.
//
// ⚠️ An earlier version had its own listening() helper, written purely so
// this test could assert a second thing. repocheck flagged it as duplicating
// chrome_sessions, and it was unused by main(). A function that exists only
// to be tested is one worth deleting.
func TestNothingListeningIsAnError(t *testing.T) {
	// Port 1 on loopback: reserved and never listening.
	if _, err := probe(1, 500*time.Millisecond); err == nil {
		t.Error("probe succeeded against a port with nothing on it")
	}
}

// TestChromeExeSearchesRatherThanGuessing pins the reporting rule.
func TestChromeExeSearchesRatherThanGuessing(t *testing.T) {
	got := chromeExe()
	if got == "" {
		t.Skip("UNANSWERED: no Chrome on this machine, so the search cannot " +
			"be shown finding one. It correctly returned empty rather than " +
			"a guessed path.")
	}
	if !strings.HasSuffix(strings.ToLower(got), "chrome.exe") {
		t.Errorf("chromeExe returned %q, which is not a chrome.exe", got)
	}
}
