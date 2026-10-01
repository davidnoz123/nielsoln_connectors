// go test ./...
package main

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// The share, and a sibling OUTSIDE it holding something that must stay
// unreachable. Both are needed: a traversal test against a path that does not
// exist passes for the wrong reason.
func shareAndSecret(t *testing.T) (root, secret string) {
	t.Helper()
	base := t.TempDir()
	root = filepath.Join(base, "share")
	writeFile(t, filepath.Join(root, "index.html"), "<p>ok</p>")
	secret = filepath.Join(base, "outside", "secret.txt")
	writeFile(t, secret, "NOT FOR SERVING")
	real, err := resolveRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	return real, secret
}

func TestResolveServesWhatIsInside(t *testing.T) {
	root, _ := shareAndSecret(t)
	got, err := resolve(root, "/index.html")
	if err != nil {
		t.Fatalf("refused a file inside the share: %v", err)
	}
	if !strings.HasSuffix(got, "index.html") {
		t.Fatalf("resolved to %q", got)
	}
}

func TestResolveRefusesTraversal(t *testing.T) {
	root, _ := shareAndSecret(t)
	for _, p := range []string{
		"/../outside/secret.txt",
		"/../../outside/secret.txt",
		"/foo/../../outside/secret.txt",
		"/..%2foutside/secret.txt",
	} {
		if _, err := resolve(root, p); err == nil {
			t.Errorf("%q was ALLOWED out of the share", p)
		}
	}
}

// A link INSIDE the share pointing OUT of it is the escape a textual check
// cannot see, which is why resolve re-tests the path after EvalSymlinks.
func TestResolveRefusesSymlinkEscape(t *testing.T) {
	root, secret := shareAndSecret(t)
	link := filepath.Join(root, "escape.txt")
	if err := os.Symlink(secret, link); err != nil {
		t.Skipf("cannot create a symlink here: %v", err)
	}
	if _, err := resolve(root, "/escape.txt"); err == nil {
		t.Fatal("a symlink out of the share was ALLOWED")
	}
}

// contained must not fold case unless the filesystem does. Folding wrongly
// admits a path that is outside the share; this is remote_ai's case_test.go
// lesson, kept here because the code was copied with it.
func TestContainedDoesNotFoldWhenFilesystemDoesNot(t *testing.T) {
	root, _ := shareAndSecret(t)
	sibling := filepath.Join(filepath.Dir(root), strings.ToUpper(filepath.Base(root)))
	if err := os.Mkdir(sibling, 0o755); err != nil {
		t.Skipf("filesystem folds case, so the two cannot coexist: %v", err)
	}
	// They coexist, so this filesystem does not fold -- and a path in the
	// UPPERCASE sibling must not read as contained by the lowercase share.
	if contained(root, filepath.Join(sibling, "secret.txt")) {
		t.Fatal("a path outside the share read as inside it")
	}
}

func TestGuard(t *testing.T) {
	hit := false
	h := guard("SECRET", func(w http.ResponseWriter, r *http.Request) { hit = true })
	for _, tc := range []struct {
		name, url, hdr string
		want           int
	}{
		{"no token", "/x", "", http.StatusForbidden},
		{"wrong query", "/x?t=nope", "", http.StatusForbidden},
		{"right query", "/x?t=SECRET", "", http.StatusOK},
		{"right header", "/x", "SECRET", http.StatusOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hit = false
			r := httptest.NewRequest("GET", tc.url, nil)
			if tc.hdr != "" {
				r.Header.Set("X-Token", tc.hdr)
			}
			w := httptest.NewRecorder()
			h(w, r)
			if tc.want == http.StatusOK && !hit {
				t.Fatal("handler was not reached")
			}
			if tc.want != http.StatusOK && w.Code != tc.want {
				t.Fatalf("got %d, want %d", w.Code, tc.want)
			}
		})
	}
}

// The page carries ?t= but its sub-resources do not, so a token scheme with no
// cookie serves the page and 403s every image in it. That is what happened on
// 2 Oct 2026, and it presented as a blank viewer rather than as an auth error.
func TestGuardSetsACookieSoSubresourcesWork(t *testing.T) {
	h := guard("SECRET", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// The page, with the token in the URL.
	w := httptest.NewRecorder()
	h(w, httptest.NewRequest("GET", "/?t=SECRET", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("the page itself was refused: %d", w.Code)
	}
	var jar *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == cookieName {
			jar = c
		}
	}
	if jar == nil {
		t.Fatal("no cookie was set, so every sub-resource will 403")
	}

	// An image, requested the way a browser requests it: no query, no header.
	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest("GET", "/pages/x.jpg", nil)
	r2.AddCookie(jar)
	h(w2, r2)
	if w2.Code != http.StatusOK {
		t.Fatalf("a sub-resource carrying the cookie was refused: %d", w2.Code)
	}

	// And without it, still refused.
	w3 := httptest.NewRecorder()
	h(w3, httptest.NewRequest("GET", "/pages/x.jpg", nil))
	if w3.Code != http.StatusForbidden {
		t.Fatalf("a sub-resource with no credential returned %d", w3.Code)
	}
}

// An empty -token means no check, and that has to stay true rather than
// accidentally refusing everything.
func TestGuardWithoutTokenAllows(t *testing.T) {
	hit := false
	guard("", func(w http.ResponseWriter, r *http.Request) { hit = true })(
		httptest.NewRecorder(), httptest.NewRequest("GET", "/x", nil))
	if !hit {
		t.Fatal("an empty token refused a request")
	}
}

// A page that connects AFTER a selection must be told the current one, or it
// sits blank until the reviewer moves again and reads as broken.
func TestEventsSendsCurrentStateOnConnect(t *testing.T) {
	h := &hub{clients: map[chan string]bool{}}
	h.broadcast("F-000123")

	srv := httptest.NewServer(http.HandlerFunc(h.events))
	defer srv.Close()
	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	line, err := bufio.NewReader(resp.Body).ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(line) != "data: F-000123" {
		t.Fatalf("got %q", line)
	}
}

func TestSelectReachesAConnectedPage(t *testing.T) {
	h := &hub{clients: map[chan string]bool{}}
	srv := httptest.NewServer(http.HandlerFunc(h.events))
	defer srv.Close()
	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	rd := bufio.NewReader(resp.Body)

	// Wait for the stream to be registered, otherwise the push races it.
	deadline := time.Now().Add(2 * time.Second)
	for {
		h.mu.Lock()
		n := len(h.clients)
		h.mu.Unlock()
		if n > 0 || time.Now().After(deadline) {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	w := httptest.NewRecorder()
	h.selectHandler(w, httptest.NewRequest("POST", "/select?id=F-000777", nil))
	if w.Code != http.StatusNoContent {
		t.Fatalf("select returned %d", w.Code)
	}
	line, err := rd.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(line) != "data: F-000777" {
		t.Fatalf("got %q", line)
	}
}

// The Mac route: Excel writes a file because Mac VBA has no HTTP client.
func TestNavFileReachesAConnectedPage(t *testing.T) {
	root := t.TempDir()
	navPath := filepath.Join(root, ".nav")
	h := &hub{clients: map[chan string]bool{}}
	go watchNav(navPath, 10*time.Millisecond, h)

	writeFile(t, navPath, "F-004242\n")
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		h.mu.Lock()
		got := h.last
		h.mu.Unlock()
		if got == "F-004242" {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("the nav file was never picked up")
}

// Ids are a fixed width here, and a reviewer moves faster than mtime
// granularity on some filesystems. Rewriting the same length must still be
// noticed.
func TestNavFileNoticesASameLengthRewrite(t *testing.T) {
	root := t.TempDir()
	navPath := filepath.Join(root, ".nav")
	h := &hub{clients: map[chan string]bool{}}
	writeFile(t, navPath, "F-000001")
	go watchNav(navPath, 10*time.Millisecond, h)

	waitFor := func(want string) bool {
		deadline := time.Now().Add(3 * time.Second)
		for time.Now().Before(deadline) {
			h.mu.Lock()
			got := h.last
			h.mu.Unlock()
			if got == want {
				return true
			}
			time.Sleep(10 * time.Millisecond)
		}
		return false
	}
	if !waitFor("F-000001") {
		t.Fatal("first id missed")
	}
	time.Sleep(30 * time.Millisecond)
	writeFile(t, navPath, "F-000002")
	if !waitFor("F-000002") {
		t.Fatal("a same-length rewrite was missed")
	}
}

func TestCleanIDStripsABOM(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"Pt2-0043", "Pt2-0043"},
		{"Pt2-0043\n", "Pt2-0043"},
		{"\ufeffPt2-0043\r\n", "Pt2-0043"},
		{"  \ufeff Pt2-0043  ", "Pt2-0043"},
		{"\ufeff", ""},
	} {
		if got := cleanID([]byte(tc.in)); got != tc.want {
			t.Errorf("cleanID(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestDirectoryListingIsRefused(t *testing.T) {
	root, _ := shareAndSecret(t)
	writeFile(t, filepath.Join(root, "sub", "a.txt"), "x")
	w := httptest.NewRecorder()
	fileHandler(root)(w, httptest.NewRequest("GET", "/sub/", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("a directory without index.html returned %d, not 404", w.Code)
	}
}
