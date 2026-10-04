package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The guards are the whole security story, so they are tested without a
// browser anywhere near them. Every case here is reachable from a web page
// the user happens to have open, which is the threat this program has to
// survive: a hyperlink in a document is clicked by a browser, and any other
// tab in that browser can issue the same request.
func testServer(t *testing.T) *server {
	t.Helper()
	path := filepath.Join(t.TempDir(), "tasks.json")
	body := `{"tasks":{"tl-e11-02":{"page":"https://example.test/","find":"x"}}}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := LoadManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	return &server{manifest: m, token: "SECRET", cdpPort: 1, focus: false,
		tabs: map[string]string{}}
}

func TestGuardRefusesWithoutTheToken(t *testing.T) {
	s := testServer(t)
	for _, url := range []string{
		"/show/tl-e11-02",
		"/show/tl-e11-02?t=",
		"/show/tl-e11-02?t=wrong",
		"/show/tl-e11-02?token=SECRET", // right value, wrong parameter
	} {
		w := httptest.NewRecorder()
		if s.guard(w, httptest.NewRequest(http.MethodGet, url, nil)) {
			t.Errorf("guard allowed %s", url)
		}
		if w.Code != http.StatusForbidden {
			t.Errorf("%s gave %d, want 403", url, w.Code)
		}
	}
}

// An Origin header means a web page made the request. A link clicked from
// Excel, Word or the address bar never carries one, so this check costs a
// legitimate caller nothing and removes the whole class.
func TestGuardRefusesAnythingSentByAWebPage(t *testing.T) {
	s := testServer(t)
	for _, origin := range []string{
		"https://evil.example",
		"http://localhost:3000",
		"null",
	} {
		req := httptest.NewRequest(http.MethodGet, "/show/tl-e11-02?t=SECRET", nil)
		req.Header.Set("Origin", origin)
		w := httptest.NewRecorder()
		if s.guard(w, req) {
			t.Errorf("guard allowed a request with Origin %q, and the token "+
				"being correct is not a defence: the token travels in a "+
				"document and is not a secret", origin)
		}
		if w.Code != http.StatusForbidden {
			t.Errorf("Origin %q gave %d, want 403", origin, w.Code)
		}
	}
}

func TestGuardAllowsAPlainClick(t *testing.T) {
	s := testServer(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/show/tl-e11-02?t=SECRET", nil)
	if !s.guard(w, req) {
		t.Fatalf("guard refused a plain link click: %d %s", w.Code, w.Body)
	}
}

// An unknown id must be refused BEFORE any Chrome is touched, and the
// refusal must name what is known: "no" with no list is the message that
// sends somebody to read the source mid-call.
func TestUnknownTaskIsRefusedAndSaysWhatIsKnown(t *testing.T) {
	s := testServer(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/show/tl-nope?t=SECRET", nil)
	s.showHandler(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("unknown task gave %d, want 404", w.Code)
	}
	if !strings.Contains(w.Body.String(), "tl-e11-02") {
		t.Errorf("the refusal does not list the known tasks: %s", w.Body)
	}
}

// The wire carries an id. If a URL in the path could reach Chrome, any page
// could use this program to drive the user's browser anywhere.
func TestAUrlOnTheWireIsJustAnUnknownTask(t *testing.T) {
	s := testServer(t)
	for _, path := range []string{
		"/show/https://evil.example",
		"/show/../../etc/passwd",
		"/show/view-source:https://evil.example",
	} {
		w := httptest.NewRecorder()
		s.showHandler(w, httptest.NewRequest(http.MethodGet, path+"?t=SECRET", nil))
		if w.Code == http.StatusOK {
			t.Errorf("%s was accepted", path)
		}
	}
}
