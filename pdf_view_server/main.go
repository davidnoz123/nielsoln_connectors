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
	"crypto/subtle"
	"encoding/json"
	"flag"
	"fmt"
	"io"
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
	port := flag.String("port", "8899", "port to bind; falls back to a free one")
	token := flag.String("token", "", "required on every request; empty disables the check")
	nav := flag.String("nav", navDefault, "file under -root that carries the current id")
	poll := flag.Duration("poll", 100*time.Millisecond, "how often to stat the nav file")
	open := flag.Bool("open", false, "open a browser at the root once listening")
	logPath := flag.String("log", "", "also append the log here (default <root>/review.log)")
	raise := flag.Bool("raise", true, "bring the viewer to the front on each selection")
	own := flag.Bool("own-window", true,
		"open the viewer in a dedicated Chrome window we launch and can raise")
	flag.Parse()

	abs, err := resolveRoot(*root)
	if err != nil {
		log.Fatalf("-root: %v", err)
	}
	if *token == "" {
		log.Printf("WARNING: no -token, so anything on this machine that can reach " +
			"the port can read -root. Pass one unless you mean this.")
	}

	// TEE THE LOG TO A FILE as well as the console.
	//
	// The console window exists, but the viewer window we open lands on top of
	// it, so somebody looking for the log finds Chrome. A file is readable
	// afterwards, over a phone call, and from the other end of an ssh session
	// -- none of which a console window is.
	lp := *logPath
	if lp == "" {
		lp = filepath.Join(abs, "review.log")
	}
	if f, err := os.OpenFile(lp, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644); err == nil {
		log.SetOutput(io.MultiWriter(os.Stderr, f))
		log.Printf("---- started ----")
	} else {
		log.Printf("could not open %s, so logging to this window only: %v", lp, err)
	}

	h := &hub{clients: map[chan string]bool{}, raise: *raise, root: abs}
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
	ln, err := listen(*host, *port, *token)
	if err != nil {
		log.Fatalf("%v", err)
	}
	if ln == nil {
		// ALREADY RUNNING IS NOT A FAILURE, and the first version treated it
		// like one: the launcher printed "nothing started / the viewer has
		// stopped" while a perfectly good server was serving. Somebody who
		// double-clicks twice -- which everybody does -- was told their
		// working viewer was broken.
		//
		// So open the browser at the one that exists and exit 3, which the
		// launcher reads as "fine, say something reassuring".
		existing := fmt.Sprintf("http://%s/", net.JoinHostPort(*host, *port))
		if *token != "" {
			existing += "?t=" + *token
		}
		if *open {
			openBrowser(existing)
		}
		os.Exit(3)
	}
	url := fmt.Sprintf("http://%s/", ln.Addr().String())
	if *token != "" {
		url += "?t=" + *token
	}
	log.Printf("serving %s", abs)
	log.Printf("nav file %s", navPath)
	log.Printf("open     %s", url)
	log.Printf("log      %s", lp)
	log.Printf("stop with Ctrl-C, or by closing this window")
	if *open {
		if *own {
			if pid := openOwnWindow(url); pid > 0 {
				h.mu.Lock()
				h.ownPID = pid
				h.mu.Unlock()
				log.Printf("viewer window opened as pid %d", pid)
			} else {
				log.Printf("no Chrome found, so using the default browser. " +
					"The viewer may end up as a background TAB, and no browser " +
					"can be told to switch tabs from outside -- put the two " +
					"windows side by side instead.")
				openBrowser(url)
			}
		} else {
			openBrowser(url)
		}
	}
	log.Fatal(http.Serve(ln, mux))
}

// listen binds the wanted port, or explains itself.
//
// A BUSY PORT IS NOT A FATAL ERROR, and treating it as one is how the launcher
// failed on 2 Oct 2026: the port was held by an unrelated tool on the same
// machine and the reviewer got Go error text ending "Only one usage of each
// socket address is normally permitted", which tells them nothing they can act
// on.
//
// Three outcomes instead:
//
//   - free: bind it
//   - held by ANOTHER COPY OF THIS SERVER: say where, and stop. Two servers
//     over one folder is never what anybody wanted
//   - held by something else: take any free port and carry on, saying so. The
//     reviewer never types a port, so a different number costs them nothing
func listen(host, port, token string) (net.Listener, error) {
	addr := net.JoinHostPort(host, port)
	ln, err := net.Listen("tcp", addr)
	if err == nil {
		return ln, nil
	}
	if isOurs(addr, token) {
		log.Printf("the viewer is ALREADY RUNNING on %s. Nothing to do.", addr)
		return nil, nil
	}
	log.Printf("port %s is in use by something else, so taking a free one instead.", port)
	free, err2 := net.Listen("tcp", net.JoinHostPort(host, "0"))
	if err2 != nil {
		return nil, fmt.Errorf("cannot listen on %s (%v), nor on any free port (%v)",
			addr, err, err2)
	}
	return free, nil
}

// isOurs reports whether whatever holds addr answers /state the way we do.
func isOurs(addr, token string) bool {
	c := http.Client{Timeout: 700 * time.Millisecond}
	u := fmt.Sprintf("http://%s/state", addr)
	if token != "" {
		u += "?t=" + token
	}
	resp, err := c.Get(u)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}
	var out map[string]any
	if json.NewDecoder(resp.Body).Decode(&out) != nil {
		return false
	}
	_, hasClients := out["clients"]
	_, hasID := out["id"]
	return hasClients && hasID
}

// -- auth -------------------------------------------------------------------

const cookieName = "pvs_token"

// guard checks the token on every request.
//
// THREE PLACES, and each earns its keep:
//
//   - a HEADER, for anything scripted
//   - a QUERY PARAMETER, because EventSource cannot set headers. A design that
//     is header-only cannot open an SSE stream from a browser at all
//   - a COOKIE, set the first time a valid query parameter arrives
//
// The cookie is not a convenience. Measured 2 Oct 2026 in a real browser: the
// PAGE carries ?t= in its URL, but every sub-resource it then requests -- every
// <img>, and the PDF that PDF.js would fetch -- does NOT. Those came back 403,
// and a 403 body decodes as an image about as well as it sounds, so the page
// rendered its header and caption over a blank space with img.complete true and
// naturalWidth 0. Nothing in the log said "forbidden" because nothing was
// looking.
//
// A smoke-test page with no sub-resources cannot find this. It took a real one.
func guard(token string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if token == "" {
			next(w, r)
			return
		}
		got, fromQuery := r.Header.Get("X-Token"), false
		if got == "" {
			if q := r.URL.Query().Get("t"); q != "" {
				got, fromQuery = q, true
			}
		}
		if got == "" {
			if c, err := r.Cookie(cookieName); err == nil {
				got = c.Value
			}
		}
		// Constant time, so a caller cannot learn the token one byte at a time
		// from how long the refusal takes. Cheap here and awkward to add later.
		if subtle.ConstantTimeCompare([]byte(got), []byte(token)) != 1 {
			// No detail. A refusal that distinguishes "wrong token" from "no
			// token" is a probing oracle, and there is nothing a legitimate
			// caller learns from the difference.
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if fromQuery {
			http.SetCookie(w, &http.Cookie{
				Name: cookieName, Value: token, Path: "/",
				HttpOnly: true, SameSite: http.SameSiteStrictMode,
			})
		}
		next(w, r)
	}
}

// -- the hub ----------------------------------------------------------------

type hub struct {
	mu       sync.Mutex
	clients  map[chan string]bool
	last     string
	raise    bool
	browser  string // inferred from the page's User-Agent
	ownPID   int    // the dedicated window WE launched, if we launched one
	title    string // the page's own <title>, which is the window's title too
	root     string // so the title can be written where the spreadsheet reads it
	lastRise time.Time
}

func (h *hub) broadcast(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if id == h.last {
		return
	}
	h.last = id
	// LOG EVERY SELECTION. The startup banner is not what somebody watching
	// the window wants to see -- they want to know whether their click arrived,
	// and from which route. Without this the log says nothing during the only
	// part anybody watches.
	log.Printf("selected %s  (%d page%s listening)", id, len(h.clients),
		map[bool]string{true: "", false: "s"}[len(h.clients) == 1])
	// Bring the browser forward. Without this the page updates behind whatever
	// the reviewer is working in, so a click appears to do nothing at all --
	// which is exactly how it read on 2 Oct 2026.
	//
	// The APPLICATION is activated, never the URL re-opened: re-opening is what
	// produces a second tab, and avoiding that is why this server exists.
	if h.raise && time.Since(h.lastRise) > 300*time.Millisecond {
		h.lastRise = time.Now()
		// OUR OWN WINDOW BY PID where we have one. A browser the reviewer
		// already had open may hold the viewer in a BACKGROUND TAB, and
		// nothing outside a browser can change which tab is in front -- so
		// raising that window shows the wrong page, which is worse than
		// doing nothing. A window we launched ourselves has no other tabs.
		go activate(h.ownPID, h.browser, h.title)
	}
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
	// Which browser is actually showing the page. Asked of the connection
	// rather than assumed, because the default browser and the one holding the
	// viewer need not be the same.
	if b := browserFrom(r.UserAgent()); b != "" {
		h.browser = b
	}
	// The page tells us its title when it connects, because that is how the
	// window is found. Asking the page beats guessing: it is the same string
	// the window manager shows.
	if t := r.URL.Query().Get("title"); t != "" {
		h.title = t
		// WRITE IT DOWN WHERE THE SPREADSHEET CAN READ IT.
		//
		// Windows will not let a BACKGROUND process take the foreground; only
		// the process that currently holds it may give it away. The server is
		// always the background one, so its own SetForegroundWindow is denied
		// the instant a human has just clicked in Excel -- which is every time
		// that matters. Excel, being the foreground process, is allowed.
		//
		// So the title goes in a file beside the nav file and the VBA raises
		// the window itself with one AppActivate.
		if h.root != "" {
			_ = os.WriteFile(filepath.Join(h.root, "viewer.title"),
				[]byte(t), 0o644)
		}
	}
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
					if id := cleanID(b); id != "" {
						h.broadcast(id)
					}
				}
			}
		}
		time.Sleep(every)
	}
}

// cleanID turns the nav file's bytes into an id.

// A UTF-8 BOM IS STRIPPED, and that is not theoretical: PowerShell's
// Set-Content -Encoding utf8 writes one, and on 2 Oct 2026 an id arrived with
// one on the front. It matched nothing in the map and the failure was SILENT --
// the page simply did not move, with nothing anywhere saying why.
//
// VBA's Print # writes no BOM, so the reviewer would not have hit it. Anything
// else that ever writes that file might, and a mismatch nobody can see is worth
// three lines to prevent.
func cleanID(b []byte) string {
	s := strings.TrimSpace(string(b))
	s = strings.TrimPrefix(s, "\ufeff")
	return strings.TrimSpace(s)
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

// browserFrom names the application behind a User-Agent, or "" if it is not one
// we know how to activate. Order matters: Edge and Opera both claim Chrome, and
// Chrome claims Safari.
func browserFrom(ua string) string {
	switch {
	case strings.Contains(ua, "Edg/"):
		return "Microsoft Edge"
	case strings.Contains(ua, "OPR/"):
		return "Opera"
	case strings.Contains(ua, "Firefox/"):
		return "Firefox"
	case strings.Contains(ua, "Chrome/"):
		return "Google Chrome"
	case strings.Contains(ua, "Safari/"):
		return "Safari"
	}
	return ""
}

// activate brings an application to the front WITHOUT handing it a URL.
//
// `open -a Name` on macOS is LaunchServices, not AppleScript, so it needs no
// "wants to control" Automation grant -- which is the whole reason it is
// preferred over `osascript ... to activate`.
//
// One file, no build constraints: these are different ARGUMENTS to exec.Command
// rather than different APIs.
func activate(ownPID int, app, title string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		// user32 directly, in microseconds. There is nothing to fall back to:
		// without a pid of our own there is no window we can be sure is the
		// viewer, and raising the wrong one is worse than raising none.
		raiseWindow(ownPID, title)
		return
	case "darwin":
		// `open -a` is LaunchServices, so no Automation grant is needed. It
		// activates the APPLICATION; with a dedicated --app window that is
		// almost always the right window, but macOS offers no
		// permission-free way to raise a SPECIFIC window, so this is the one
		// place the behaviour is best-effort. Verify it on the real machine.
		if app == "" {
			app = "Google Chrome"
		}
		cmd = exec.Command("open", "-a", app)
	default:
		return // X11 and Wayland disagree; not worth guessing
	}
	_ = cmd.Run()
}

// chromePaths lists where Chrome lives, most likely first.
func chromePaths() []string {
	switch runtime.GOOS {
	case "windows":
		var out []string
		for _, base := range []string{os.Getenv("ProgramFiles"),
			os.Getenv("ProgramFiles(x86)"), os.Getenv("LocalAppData")} {
			if base != "" {
				out = append(out, filepath.Join(base,
					"Google", "Chrome", "Application", "chrome.exe"))
			}
		}
		return out
	case "darwin":
		return []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			os.Getenv("HOME") + "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		}
	default:
		return []string{"/usr/bin/google-chrome", "/usr/bin/chromium"}
	}
}

// openOwnWindow launches a dedicated viewer window and returns its pid, or 0.
//
// `--app=` gives a window with NO TABS and no address bar, so the viewer cannot
// become a background tab -- which is the thing that made raising useless. Its
// own `--user-data-dir` keeps it out of the reviewer's session entirely: no
// extensions, nothing signed in, and closing it disturbs nothing.
func openOwnWindow(url string) int {
	var exe string
	for _, p := range chromePaths() {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			exe = p
			break
		}
	}
	if exe == "" {
		return 0
	}
	profile := filepath.Join(os.TempDir(), "pdf_view_server_profile")
	cmd := exec.Command(exe, "--app="+url, "--user-data-dir="+profile,
		"--no-first-run", "--no-default-browser-check")
	if err := cmd.Start(); err != nil {
		return 0
	}
	go func() { _ = cmd.Wait() }() // reap, so it does not linger as a zombie
	return cmd.Process.Pid
}

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
