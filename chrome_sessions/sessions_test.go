// sessions_test.go -- what can be checked without a Chrome.
//
// The probe speaks plain HTTP, so a fake endpoint tests it exactly: the thing
// under test is "does it read /json/version and /json/list correctly, and does
// it describe a refusal honestly", and none of that needs a browser.
//
// ⚠️ WHAT THESE TESTS CANNOT SHOW is that a real Chrome answers in this shape.
// That is the fake-versus-real split the house rules are blunt about: a fake
// is a crystallised model of the dependency and can only show the code matches
// the model. The real run is a separate exercise against a browser, and
// saying "tests pass" on the strength of this file alone would be a false
// statement about what was verified.
package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

// fakeChrome serves the two endpoints a probe reads, and nothing else.
//
// Strict on purpose: an unexpected path is a 404 with a named path rather than
// a plausible empty answer, so code reaching for surface this fake never
// modelled fails loudly instead of quietly passing.
func fakeChrome(t *testing.T, tabs []Tab, breakList bool) (int, func()) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/json/version", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(version{
			Browser:              "Chrome/141.0.0.0",
			UserAgent:            "Mozilla/5.0 (fake)",
			WebSocketDebuggerURL: "ws://127.0.0.1/devtools/browser/abc",
		})
	})
	mux.HandleFunc("/json/list", func(w http.ResponseWriter, _ *http.Request) {
		if breakList {
			http.Error(w, "no", http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(tabs)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "fake chrome does not model "+r.URL.Path,
			http.StatusNotFound)
	})
	srv := httptest.NewServer(mux)
	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parsing the fake's own URL: %v", err)
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		t.Fatalf("the fake's port is not a number: %v", err)
	}
	return port, srv.Close
}

func TestProbeReadsAnAnsweringEndpoint(t *testing.T) {
	want := []Tab{
		{Title: "ChatGPT", URL: "https://chatgpt.com/", Type: "page"},
		{Title: "a worker", URL: "https://example.com/sw.js", Type: "worker"},
	}
	port, stop := fakeChrome(t, want, false)
	defer stop()

	got := probe(port, 2*time.Second)
	if !got.Alive {
		t.Fatalf("an answering endpoint read as dead: %+v", got)
	}
	if got.Browser != "Chrome/141.0.0.0" {
		t.Errorf("browser = %q", got.Browser)
	}
	if got.WebSocketURL == "" {
		t.Error("the websocket URL was dropped, and the authority probe will " +
			"need it")
	}
	// Only pages. A worker is a target and not somewhere a person is signed
	// in, so including it would pad the hint with noise.
	if len(got.Tabs) != 1 || got.Tabs[0].Title != "ChatGPT" {
		t.Errorf("tabs = %+v, wanted the page only", got.Tabs)
	}
	if got.Error != "" {
		t.Errorf("a clean probe reported an error: %q", got.Error)
	}
}

func TestVersionWithoutTargetsIsStillAlive(t *testing.T) {
	port, stop := fakeChrome(t, nil, true)
	defer stop()

	got := probe(port, 2*time.Second)
	// THE POINT OF THE CASE: a browser that answers /json/version is
	// reachable whether or not it will list targets, and calling it dead
	// because an aside failed would be a lie about the thing that matters.
	if !got.Alive {
		t.Fatalf("a browser that listed its version read as dead: %+v", got)
	}
	if !strings.Contains(got.Error, "not targets") {
		t.Errorf("the tab-list failure was not reported: %q", got.Error)
	}
}

func TestNothingListeningSaysSo(t *testing.T) {
	// A port nothing is on. Taken by binding and releasing, so the number is
	// known to be free rather than assumed.
	srv := httptest.NewServer(http.NewServeMux())
	u, _ := url.Parse(srv.URL)
	port, _ := strconv.Atoi(u.Port())
	srv.Close()

	got := probe(port, 500*time.Millisecond)
	if got.Alive {
		t.Fatalf("a closed port read as alive: %+v", got)
	}
	if got.Error == "" {
		t.Error("a refusal with no reason given, which is the thing this " +
			"field exists to prevent")
	}
	if listening(port, 300*time.Millisecond) {
		t.Error("listening() says something holds a port that was just closed")
	}
}

func TestParsePortsRefusesRatherThanDropping(t *testing.T) {
	ok, err := parsePorts(" 9222, 9333 ,9222 ")
	if err != nil {
		t.Fatalf("a valid list was refused: %v", err)
	}
	// Deduplicated, and in the order given: probing one port twice would
	// report it twice, which reads as two browsers.
	if len(ok) != 2 || ok[0] != 9222 || ok[1] != 9333 {
		t.Errorf("ports = %v", ok)
	}
	for _, bad := range []string{"", "   ", "nine", "0", "65536", "9222,x"} {
		if _, err := parsePorts(bad); err == nil {
			t.Errorf("%q was accepted, and a silently dropped port is a port "+
				"nobody probes", bad)
		}
	}
}

func TestRenderSaysDeadOutLoud(t *testing.T) {
	text := render([]Session{
		{Port: 9222, Alive: true, Browser: "Chrome/141", OwnerPID: 1234},
		{Port: 9333, Error: "nothing is listening on this port"},
	}, false)
	if !strings.Contains(text, "alive") || !strings.Contains(text, "DEAD") {
		t.Errorf("the human report hides a state:\n%s", text)
	}
	if !strings.Contains(text, "pid 1234") {
		t.Errorf("the owner was not reported:\n%s", text)
	}
	if !strings.Contains(text, "owner unknown") {
		t.Errorf("an unknown owner should say so rather than print 0:\n%s",
			text)
	}

	var back []Session
	raw := render([]Session{{Port: 9222, Alive: true}}, true)
	if err := json.Unmarshal([]byte(raw), &back); err != nil {
		t.Fatalf("the JSON form does not parse, and a Python caller is the "+
			"whole reason it exists: %v", err)
	}
	if len(back) != 1 || back[0].Port != 9222 || !back[0].Alive {
		t.Errorf("round trip lost something: %+v", back)
	}
}

func TestEmptyReportSaysNothingToReport(t *testing.T) {
	// A blank page reads as a crash. Something has to be said.
	if !strings.Contains(render(nil, false), "nothing to report") {
		t.Error("an empty human report is silent")
	}
}

func TestDiscoveryLeavesTheReservedRangeAlone(t *testing.T) {
	// MEASURED ON THE FIRST REAL RUN: -discover sent an HTTP GET to 135, 139,
	// 443 and 445, which is RPC and SMB. One of them answered by forcibly
	// closing the connection, which is the service objecting.
	for _, port := range []int{135, 139, 443, 445, 1023} {
		if worthProbing(port) {
			t.Errorf("%d would be probed, and nothing below 1024 is a CDP "+
				"endpoint", port)
		}
	}
	for _, port := range []int{1024, 8000, 9222, 9333, 65535} {
		if !worthProbing(port) {
			t.Errorf("%d would be skipped, and a browser can sit there", port)
		}
	}
}

func TestHumanViewUnescapesTitlesAndJSONDoesNot(t *testing.T) {
	escaped := "Edit &quot;Home&quot; with Elementor"
	one := []Session{{Port: 9222, Alive: true, Tabs: []Tab{
		{Title: escaped, URL: "https://example.com/", Type: "page"}}}}

	text := render(one, false)
	if strings.Contains(text, "&quot;") {
		t.Errorf("the human view still shows an entity:\n%s", text)
	}
	if !strings.Contains(text, `Edit "Home"`) {
		t.Errorf("the title was not decoded:\n%s", text)
	}

	// And the JSON keeps what Chrome said, because a caller parsing it should
	// get the value rather than this program's idea of the value.
	var back []Session
	if err := json.Unmarshal([]byte(render(one, true)), &back); err != nil {
		t.Fatalf("JSON did not parse: %v", err)
	}
	if back[0].Tabs[0].Title != escaped {
		t.Errorf("JSON title = %q, wanted Chrome's own string",
			back[0].Tabs[0].Title)
	}
}
