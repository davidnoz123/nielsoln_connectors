// pdf_view_server -- serve one folder over loopback, and tell the page which
// item to show.
//
//	go run github.com/davidnoz123/nielsoln_connectors/pdf_view_server@<sha> \
//	    -root . -host 127.0.0.1 -port 8790 -token K7M2XQ4P
//
// The problem it exists for: a reviewer works through rows in a spreadsheet and
// wants the matching passage shown, highlighted, in a browser. Making the
// spreadsheet open a URL gives a NEW TAB PER ROW, which is unusable by about
// the fifth row, and nothing in the hyperlink path can ask a browser to reuse
// one.
//
// So the browser is never driven. A page served from here holds an SSE stream
// open, and the spreadsheet pushes an id. The page moves itself.
//
// Three consequences follow, and they are the whole reason for this shape:
//
//   - ANY BROWSER WORKS. Safari, Chrome, Firefox. Nothing is automated, so
//     there is no AppleScript, no CDP, no "Excel wants to control Chrome"
//     permission dialog, and no browser-specific code to keep working.
//   - ONE TAB, FOR THE WHOLE SESSION.
//   - fetch() WORKS, because this is http:// rather than file://. That is what
//     lets a page use PDF.js on a real PDF instead of pre-rendered images, so
//     text stays selectable and zoom stays sharp.
//
// It knows nothing about PDFs. It serves bytes out of -root and relays ids. The
// viewer, the PDF and whatever maps an id to a page and a rectangle are all
// DATA, supplied by whoever runs it. That is deliberate: this repo is public and
// the documents people review with it are not.
//
// # Two ways in, because Excel has two
//
// `POST /select` is the obvious one and works from Windows VBA through
// MSXML2.XMLHTTP.
//
// Excel for Mac HAS NO HTTP CLIENT IN VBA. MSXML does not exist there, so the
// alternatives are AppleScriptTask shelling out to curl -- which drags in a
// script file in a sandboxed folder -- or writing a file, which VBA does
// identically on both platforms with no permissions at all.
//
// So the nav file is watched as well. `Open ... For Output` in two lines of
// VBA, the same on both platforms, and the gate around it stays identical.
//
// # Polling, not fsnotify
//
// Native watching means inotify, FSEvents and ReadDirectoryChangesW -- a
// dependency and three platform files, to watch ONE file. A 100ms stat is
// imperceptible to someone moving down spreadsheet rows and keeps go.mod empty
// and the build free of constraints.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const navDefault = ".nav"

func main() {
	root := flag.String("root", ".", "the only folder this server may read")
	host := flag.String("host", "127.0.0.1", "address to bind")
	port := flag.String("port", "8790", "port to bind")
	token := flag.String("token", "", "required on every request; empty disables the check")
	nav := flag.String("nav", navDefault, "file under -root that carries the current id")
	poll := flag.Duration("poll", 100*time.Millisecond, "how often to stat the nav file")
	open := flag.Bool("open", false, "open a browser at the root once listening")
	flag.Parse()

	abs, err := resolveRoot(*root)
	if err != nil {
		log.Fatalf("-root: %v", err)
	}
	if *token == "" {
		log.Printf("WARNING: no -token, so anything on this machine that can reach " +
			"the port can read -root. Pass one unless you mean this.")
	}

	h := &hub{clients: map[chan string]bool{}}
	navPath := filepath.Join(abs, filepath.Clean("/"+*nav))
	go watchNav(navPath, *poll, h)

	mux := http.NewServeMux()
	mux.HandleFunc("/events", guard(*token, h.events))
	mux.HandleFunc("/select", guard(*token, h.selectHandler))
	mux.HandleFunc("/state", guard(*token, h.stateHandler))
	mux.HandleFunc("/", guard(*token, fileHandler(abs)))

	// Bind BEFORE announcing, so the address printed is one that is actually
	// listening. A message followed by "address already in use" sends the
	// reader looking for the wrong problem.
	addr := net.JoinHostPort(*host, *port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("cannot listen on %s: %v", addr, err)
	}
	url := fmt.Sprintf("http://%s/", addr)
	if *token != "" {
		url += "?t=" + *token
	}
	log.Printf("serving %s", abs)
	log.Printf("nav file %s", navPath)
	log.Printf("open     %s", url)
	log.Printf("stop with Ctrl-C, or by closing this window")
	if *open {
		openBrowser(url)
	}
	log.Fatal(http.Serve(ln, mux))
}

// -- auth -------------------------------------------------------------------

// guard checks the token on every request.
//
// A QUERY PARAMETER as well as a header, because EventSource cannot set
// headers. Any design that is header-only cannot open an SSE stream from a
// browser, which is the one thing this server exists to do.
func guard(token string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if token != "" {
			got := r.Header.Get("X-Token")
			if got == "" {
				got = r.URL.Query().Get("t")
			}
			if got != token {
				// No detail. A refusal that distinguishes "wrong token" from
				// "no token" is a probing oracle, and there is nothing a
				// legitimate caller learns from the difference.
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
		}
		next(w, r)
	}
}

// -- the hub ----------------------------------------------------------------

type hub struct {
	mu      sync.Mutex
	clients map[chan string]bool
	last    string
}

func (h *hub) broadcast(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if id == h.last {
		return
	}
	h.last = id
	for c := range h.clients {
		// Non-blocking. A page that has stopped reading must not wedge the
		// watcher goroutine, and the next event carries the current state
		// anyway, so a dropped one costs nothing.
		select {
		case c <- id:
		default:
		}
	}
}

func (h *hub) events(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Flush the headers NOW, before anything else can block.
	//
	// Found by TestSelectReachesAConnectedPage: net/http does not send headers
	// until the first write, so with no selection yet the first write was the
	// 25s keepalive -- and a browser's EventSource.onopen did not fire until
	// then. The stream looked dead for twenty-five seconds every time somebody
	// opened the page before touching the spreadsheet, which is the normal
	// order of events.
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	ch := make(chan string, 8)
	h.mu.Lock()
	h.clients[ch] = true
	current := h.last
	h.mu.Unlock()
	defer func() {
		h.mu.Lock()
		delete(h.clients, ch)
		h.mu.Unlock()
	}()

	// Send the CURRENT state immediately. A page that connects after a
	// selection was made would otherwise sit blank until the next one, which
	// reads as the server being broken.
	if current != "" {
		fmt.Fprintf(w, "data: %s\n\n", current)
		flusher.Flush()
	}
	// A comment line every 25s. Some proxies and some browsers quietly drop an
	// idle stream, and a reviewer can easily look at one row for a minute.
	tick := time.NewTicker(25 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case id := <-ch:
			fmt.Fprintf(w, "data: %s\n\n", id)
			flusher.Flush()
		case <-tick.C:
			fmt.Fprint(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}

// selectHandler is the Windows route. Accepts the id as a query parameter or a
// plain body, because VBA's MSXML can send either and neither deserves JSON.
func (h *hub) selectHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		buf := make([]byte, 4096)
		n, _ := r.Body.Read(buf)
		id = strings.TrimSpace(string(buf[:n]))
	}
	if id == "" {
		http.Error(w, "no id", http.StatusBadRequest)
		return
	}
	h.broadcast(id)
	w.WriteHeader(http.StatusNoContent)
}

func (h *hub) stateHandler(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	out := map[string]any{"id": h.last, "clients": len(h.clients)}
	h.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// -- the nav file -----------------------------------------------------------

// watchNav polls one file and broadcasts what it says.
//
// Size AND modification time, because a spreadsheet rewriting the same-length
// id within one filesystem timestamp tick is exactly the case here -- ids are
// a fixed width and a reviewer can move rows faster than a 1s mtime
// granularity. Reading on every tick would also work; this just avoids the
// read when nothing moved.
func watchNav(path string, every time.Duration, h *hub) {
	var lastMod time.Time
	var lastSize int64 = -1
	for {
		if st, err := os.Stat(path); err == nil && !st.IsDir() {
			if st.ModTime() != lastMod || st.Size() != lastSize {
				lastMod, lastSize = st.ModTime(), st.Size()
				if b, err := os.ReadFile(path); err == nil {
					if id := strings.TrimSpace(string(b)); id != "" {
						h.broadcast(id)
					}
				}
			}
		}
		time.Sleep(every)
	}
}

// -- files ------------------------------------------------------------------

func fileHandler(root string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, err := resolve(root, r.URL.Path)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		st, err := os.Stat(p)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if st.IsDir() {
			// index.html or nothing. A directory listing of -root would expose
			// the nav file and anything else the operator put beside the
			// viewer, and nobody reviewing a document needs one.
			idx := filepath.Join(p, "index.html")
			if _, err := os.Stat(idx); err != nil {
				http.Error(w, "no index.html here", http.StatusNotFound)
				return
			}
			p = idx
		}
		http.ServeFile(w, r, p)
	}
}

// -- containment ------------------------------------------------------------

func resolveRoot(root string) (string, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	real, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", fmt.Errorf("%s: %w", abs, err)
	}
	st, err := os.Stat(real)
	if err != nil {
		return "", err
	}
	if !st.IsDir() {
		return "", fmt.Errorf("%s is not a directory", real)
	}
	return real, nil
}

// resolve turns a URL path into a real path inside root, or refuses.
//
// EvalSymlinks AFTER joining, then contained() again: a link INSIDE root that
// points outside it is the escape that a textual check cannot see, and it is
// the whole reason the resolved path is re-tested rather than the requested
// one.
func resolve(root, urlPath string) (string, error) {
	clean := filepath.Clean("/" + strings.TrimPrefix(urlPath, "/"))
	p := filepath.Join(root, filepath.FromSlash(clean))
	real, err := filepath.EvalSymlinks(p)
	if err != nil {
		return "", err
	}
	if !contained(root, real) {
		return "", fmt.Errorf("outside root")
	}
	return real, nil
}

// contained reports whether real is root or sits beneath it.
//
// NOT filepath.Rel, and this is carried over from remote_ai's case_test.go
// rather than rediscovered: on Windows, Rel compares components with
// strings.EqualFold, so it folds case whatever it is handed. Folding
// unconditionally is worse than useless -- with a share at ...\me\x and a
// genuinely separate ...\me\X beside it, the folded answer calls a path
// outside the share contained.
//
// So fold only when the filesystem MEASURABLY folds, asked of the filesystem
// rather than inferred from runtime.GOOS. NTFS carries a per-directory case
// sensitivity flag that `fsutil` sets without elevation, and WSL sets it on
// directories it creates, so the OS name is not the answer.
func contained(root, real string) bool {
	a, b := root, real
	if foldsCase(root) {
		a, b = strings.ToLower(a), strings.ToLower(b)
	}
	if a == b {
		return true
	}
	if !strings.HasSuffix(a, string(filepath.Separator)) {
		a += string(filepath.Separator)
	}
	return strings.HasPrefix(b, a)
}

var (
	foldMu    sync.Mutex
	foldCache = map[string]bool{}
)

// foldsCase asks the filesystem whether dir and a case-flipped spelling of it
// are the same place.
//
// Anything but a clear yes answers NO. The two errors are not symmetrical:
// failing to fold refuses paths that are legitimately inside root, which is
// visible and annoying, while folding wrongly admits paths that are outside
// it. Only one of those is safe to guess at.
func foldsCase(dir string) bool {
	foldMu.Lock()
	defer foldMu.Unlock()
	if known, ok := foldCache[dir]; ok {
		return known
	}
	st, err := os.Stat(dir)
	if err != nil {
		return false
	}
	flipped := flipCase(dir)
	if flipped == dir {
		return false // nothing to flip; cannot be asked
	}
	other, err := os.Stat(flipped)
	if err != nil {
		foldCache[dir] = false
		return false
	}
	folds := os.SameFile(st, other)
	foldCache[dir] = folds
	return folds
}

func flipCase(s string) string {
	r := []rune(s)
	for i, c := range r {
		switch {
		case c >= 'a' && c <= 'z':
			r[i] = c - 32
		case c >= 'A' && c <= 'Z':
			r[i] = c + 32
		}
	}
	return string(r)
}

// -- convenience ------------------------------------------------------------

// openBrowser is the only place the operating system is named, and it needs no
// build constraint: these are different ARGUMENTS to the same call, not
// different APIs, so one file compiles everywhere.
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		log.Printf("could not open a browser (%v). Open %s yourself.", err, url)
	}
}
