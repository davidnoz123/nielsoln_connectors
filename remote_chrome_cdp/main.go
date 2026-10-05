// remote_chrome_cdp -- a document's links, driving a Chrome you can see.
//
//	go run github.com/davidnoz123/nielsoln_connectors/remote_chrome_cdp@<sha> \
//	    -manifest tasks.json -listen 127.0.0.1:8790 -token K7M2XQ4P
//
// # WHAT IT IS FOR
//
// A playbook says "we are changing the logo URL in the page's code". Nothing
// on the rendered page changes, so the sentence has to be taken on trust. A
// link that opens the page's HTML, scrolled to that exact string and
// highlighted, replaces trust with looking.
//
// # WHAT IT CAN DO, EXACTLY
//
// Open a tab it created, go to one of the addresses listed in the manifest,
// highlight one string, and bring the window forward. That is all. It cannot
// be asked to go anywhere else, it cannot run script supplied by a caller,
// and it reads nothing back out of the page.
//
// # WHAT IT CANNOT DO
//
// It never touches a tab it did not open. It has no file access, no exec and
// no way to be given a URL over the wire. If the manifest does not list an
// id, the answer is no.
//
// # WHY THE TOKEN IS NOT THE POINT
//
// The token stops other pages in the same browser quietly calling this, and
// it is worth having. But it travels in a document, so it is not a secret.
// The thing that makes this safe to run is that the vocabulary is closed:
// the worst a caller can achieve is a page the manifest already allowed.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// took renders an elapsed time the way a human reads it. Every begin has a
// done with one of these: "it felt slow" is not something anybody can act
// on, and without a number nobody could have told that a four second sleep
// was the cost rather than Chrome being slow.
func tookFrom(a, b time.Time) string {
	d := b.Sub(a)
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

func took(from time.Time) string {
	d := time.Since(from)
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

const (
	defaultListen = "127.0.0.1:8790"
	defaultCDP    = 9222
)

type server struct {
	manifest      *Manifest
	token         string
	cdpPort       int
	record        string
	focus         bool
	listen        string
	launch        bool
	chromeExe     string
	chromeProfile string
	sweepLanes    int
	absentBudget  time.Duration

	mu sync.Mutex // one Chrome conversation at a time, see drive()

	// Bumped by every request. A drive() whose number is no longer the
	// newest gives up: when somebody clicks four rows in a row they want
	// the LAST one, and a queue hands it to them after three they have
	// stopped caring about. Measured on 3 Oct: a navigation cost four
	// seconds, so the fourth click landed sixteen seconds late.
	gen int64

	// The tasks already on their way, so a second click on the SAME cell
	// while the first is still opening does nothing. A DOUBLE-CLICK IS ONE
	// INTENTION, and `gen` above cannot see that: it drops a request still
	// QUEUED when a newer one arrives, and the first of a double-click is
	// not queued, it is past that check and running.
	//
	// Measured 5 Oct: two clicks two seconds apart on tl-e10-01:edit ran
	// both drives, 5.3s and then 10.8s, and the window came forward
	// thirteen seconds after the first click. Both drives were correct.
	// Neither was wanted.
	//
	// A set rather than one id, because a QUEUED drive is on its way too
	// and a click on that one is the same repeat.
	//
	// THE TIME MATTERS, not just the fact. This was a bare set and the
	// window was therefore "however long the drive takes", which on a slow
	// row was twelve seconds: a human who saw nothing happen, clicked
	// again after eleven, and had it silently swallowed as a double-click.
	// A click two seconds in is one intention expressed twice; a click
	// eleven seconds in is somebody who thinks it is broken, and the two
	// must not be treated alike.
	flying map[string]time.Time
	flyMu  sync.Mutex

	// The tab we last opened for each PAGE, so a second click re-uses it
	// instead of adding another. Keyed by page rather than by task because
	// several tasks examine the same page: E11 and E12 are both in the
	// homepage's head.
	tabs map[string]string
	// Those same keys, oldest first, so the cap closes the tab nobody has
	// looked at for longest rather than an arbitrary one. A slice because
	// the cap is single digits and finding a key in eight strings is not
	// worth a second map.
	tabAge  []string
	maxTabs int
}

// coalesceWindow is how long after a click a second click on the same cell
// is treated as the same click. A double-click is two events a quarter of a
// second apart; three seconds is generous for that and far short of the
// patience of somebody waiting for a window to appear.
const coalesceWindow = 3 * time.Second

func main() {
	listen := flag.String("listen", defaultListen,
		"address to serve on; loopback only unless you mean otherwise")
	manifestPath := flag.String("manifest", "tasks.json",
		"the list of tasks this program may show")
	token := flag.String("token", "", "required in every request as ?t=")
	cdpPort := flag.Int("chrome-port", defaultCDP,
		"the --remote-debugging-port of the Chrome to drive")
	record := flag.String("record", "",
		"append every request here as JSON lines")
	chromeExe := flag.String("chrome-exe", "",
		"path to chrome.exe; found automatically when empty")
	chromeProfile := flag.String("chrome-profile", "",
		"user-data-dir for a Chrome this program starts. REQUIRED with "+
			"-launch: a profile it invents is signed out, and the Google "+
			"screens then show a sign-in page instead of the setting")
	launch := flag.Bool("launch", true,
		"start a Chrome when none is listening, rather than refusing")
	focus := flag.Bool("focus", true,
		"raise the Chrome window once the page is ready; -focus=false "+
			"leaves the desktop alone, which is what you want while working")
	sweepLanes := flag.Int("sweep-lanes", 4,
		"pages loaded at once during /sweep. Bounded on purpose: the far "+
			"end is one customer site, and twenty parallel fetches of it "+
			"is a small denial of service against the people we are "+
			"helping")
	absentMS := flag.Int("absent-budget-ms", 1500,
		"how long to keep looking for an anchor the slate says is NOT "+
			"there yet. Paid in full on every such row, so it is the "+
			"largest single cost in a full test")
	maxTabs := flag.Int("max-tabs", 8,
		"how many tabs this process keeps open before closing its own "+
			"oldest. A full test run touches 21 pages and left 41 tabs, "+
			"95 Chrome processes and 8GB of working set. Tabs the human "+
			"opened are never counted and never closed; 0 disables the cap")
	windowCache := flag.String("window-cache", "",
		"a file to remember the Chrome window handle in, so the first "+
			"right-click after a restart raises the window as fast as the "+
			"rest. A stale handle is checked and discarded, so the worst "+
			"case is the behaviour without it")
	logPath := flag.String("log", "",
		"also append everything printed here to this file. The console is "+
			"the only place that says what each click did, which is no use "+
			"to anyone reading it later or to a tool checking it")
	flag.Parse()

	// Set up BEFORE the startup banner, so the file records which Chrome,
	// which profile and which warnings this process started with. Those
	// lines are the ones worth having when something turns out to have been
	// pointed at the wrong browser all along.
	if *logPath != "" {
		f, err := os.OpenFile(*logPath,
			os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			log.Printf("  WARNING: cannot write -log %s: %v", *logPath, err)
		} else {
			defer f.Close()
			// MultiWriter, not a redirect: the console stays live. A
			// windowless server hides exactly the output that answers
			// "what just happened".
			log.SetOutput(io.MultiWriter(os.Stderr, f))
			log.Printf("  logging to %s", *logPath)
		}
	}

	if *token == "" {
		log.Fatal("  -token is required. Without one, any page open in the " +
			"same browser can call this program.")
	}
	m, err := LoadManifest(*manifestPath)
	if err != nil {
		log.Fatalf("  %v", err)
	}
	// Said out loud at startup, because "which Chrome" is the question this
	// program is most likely to get wrong in a way nobody notices: a second
	// Chrome on another port looks identical and is not signed into
	// anything.
	if _, err := browserWS(*cdpPort); err != nil {
		log.Printf("  WARNING: %v", err)
		log.Printf("  Serving anyway. Requests will fail until a Chrome is " +
			"listening, and will say so.")
	}
	if host, _, err := net.SplitHostPort(*listen); err == nil &&
		host != "127.0.0.1" && host != "localhost" && host != "::1" {
		log.Printf("  WARNING: listening on %s, which is not loopback. "+
			"Anything that can reach this address can drive that Chrome.", host)
	}

	s := &server{manifest: m, token: *token, cdpPort: *cdpPort,
		record: *record, focus: *focus, listen: *listen,
		launch: *launch, chromeExe: *chromeExe,
		chromeProfile: *chromeProfile, tabs: map[string]string{},
		flying:       map[string]time.Time{},
		sweepLanes:   *sweepLanes,
		maxTabs:      *maxTabs,
		absentBudget: time.Duration(*absentMS) * time.Millisecond}
	// Before the first request, so the log says whether the window is
	// already known or has to be learned.
	log.Printf("  focus: %s", LoadWindowHandle(*windowCache))

	mux := http.NewServeMux()
	mux.HandleFunc("/show/", s.showHandler)
	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("/sweep", s.sweepHandler)

	log.Printf("  remote_chrome_cdp on http://%s", *listen)
	log.Printf("  %d task(s): %s", len(m.Tasks), strings.Join(m.IDs(), ", "))
	log.Printf("  driving Chrome on 127.0.0.1:%d", *cdpPort)
	// Said every time, because "which Chrome" decides whether a Google
	// screen shows the setting or the sign-in page, and the two are
	// indistinguishable from in here.
	if *launch {
		if *chromeProfile == "" {
			log.Printf("  -launch is on with NO -chrome-profile: if the " +
				"Chrome on this port goes away, nothing will be started " +
				"and requests will say so")
		} else {
			log.Printf("  if it has to start one, profile: %s", *chromeProfile)
		}
	}
	log.Printf("  a link looks like  http://%s/show/E11?t=%s", *listen, *token)

	srv := &http.Server{
		Addr:              *listen,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}

// guard applies the two checks that stand between this program and any web
// page the user happens to have open.
func (s *server) guard(w http.ResponseWriter, r *http.Request) bool {
	// An Origin header means a web page made this request, not a link click
	// from Excel, Word or the address bar. There is no legitimate caller of
	// this program that sends one, so the check costs nothing and removes
	// the whole class.
	if r.Header.Get("Origin") != "" {
		s.deny(w, http.StatusForbidden,
			"This request came from a web page. This program only answers "+
				"links clicked from a document or typed into the address bar.")
		return false
	}
	if r.URL.Query().Get("t") != s.token {
		s.deny(w, http.StatusForbidden, "Wrong or missing token.")
		return false
	}
	return true
}

func (s *server) healthHandler(w http.ResponseWriter, r *http.Request) {
	if !s.guard(w, r) {
		return
	}
	_, err := browserWS(s.cdpPort)
	out := map[string]any{
		"tasks":  s.manifest.IDs(),
		"chrome": err == nil,
	}
	if err != nil {
		out["chrome_error"] = err.Error()
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func (s *server) showHandler(w http.ResponseWriter, r *http.Request) {
	if !s.guard(w, r) {
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/show/")
	task, ok := s.manifest.Lookup(id)
	if !ok {
		s.deny(w, http.StatusNotFound, fmt.Sprintf(
			"No task %q in the manifest. Known: %s",
			id, strings.Join(s.manifest.IDs(), ", ")))
		s.log(id, "unknown task")
		return
	}

	// WHO ASKED, computed once. It used to be three separate wantsFocus
	// calls and nothing wrote the answer down, so the log could say a drive
	// began and not whether it was about to take the desktop. Answering
	// "what raised my window at 13:14" then meant reading process
	// command lines an hour later instead of grepping one file.
	client := r.Header.Get("X-Drive-Client")
	if client == "" {
		// NOT "browser". A missing header is a browser or a script that
		// forgot, and the log must not guess which.
		client = "unnamed"
	}
	wantFocus := s.wantsFocus(r)
	raising := "no focus"
	if wantFocus {
		raising = "will raise"
	}

	// ⚠️ THE WINDOW COMES UP BEFORE ANYTHING ELSE HAPPENS.
	//
	// This is the first work the process does for a request, ahead of the
	// Chrome health check, the reply, the lock and CDP. It has to be,
	// because the thing that makes raising possible at all is a grant Excel
	// hands over at the instant of the right-click, and a grant is spent
	// best while it is fresh.
	//
	// Before this, the only way to find the window was to match a tab
	// TITLE, which does not exist until the page has loaded, so focus
	// queued behind a navigation it has nothing to do with. On a drive that
	// opened its own tab that was seconds, and Windows refused the raise
	// about as often as it allowed it. The complaint was exactly this:
	// focus should be the first thing, not the last.
	//
	// A handle is what makes it possible, so the first raise of a session
	// still goes the slow way and teaches us the number. Everything after
	// is a validity check and a SetForegroundWindow.
	//
	// ONLY when this request wanted focus, which by the allow-list means a
	// human right-clicked a cell. A test raises nothing.
	raisedEarly := false
	if wantFocus {
		if out := RaiseRemembered(); out != "" {
			raisedEarly = strings.HasPrefix(out, "raised")
			s.log(id, "focus on arrival: "+out)
		}
	}

	// ⚠️ THE REPLY GOES FIRST, and that is the whole trick.
	//
	// The tab the human clicked is sitting on this very URL, in the Chrome
	// we drive, and turning THAT tab into the page they asked for is what
	// they expected a link to do. But it only exists once we have answered
	// it: before that the browser is still waiting and the tab has no URL
	// to find it by. So we answer, then go and take it over.
	//
	// A 302 would be simpler and cannot work here: Chrome refuses
	// view-source: from a redirect exactly as it does from a link. CDP
	// navigation is not subject to that, which is why this route exists at
	// all.
	//
	// The reply is also the fallback. If the click landed in a different
	// browser from the one we drive, there is no tab to take over, we open
	// our own, and what they are left reading is a page naming the task and
	// offering an address that opens.
	// ⚠️ ASKED BEFORE WE ANSWER, because the answer goes out before the
	// work is done and therefore cannot report on it. The first version
	// replied "It is open in another Chrome tab, highlighted" and then went
	// looking for a Chrome: when there was none, that sentence stayed on
	// screen asserting something that had not happened and was not going
	// to. A page that claims success it cannot know is worse than an error.
	if _, err := browserWS(s.cdpPort); err != nil {
		// Start one rather than refuse. A browser somebody closed is the
		// commonest reason a link in the sheet stops working, and it
		// recurred three times in one afternoon: nothing else recovers it.
		if !s.launch {
			s.noChrome(w, id, task, err)
			s.log(id, "no chrome: "+err.Error())
			return
		}
		log.Printf("  no Chrome on %d: starting one", s.cdpPort)
		if lerr := launchChrome(s.chromeExe, s.chromeProfile, s.cdpPort,
			25*time.Second); lerr != nil {
			s.noChrome(w, id, task, lerr)
			s.log(id, "could not start chrome: "+lerr.Error())
			return
		}
		log.Printf("  started a Chrome on %d", s.cdpPort)
		s.log(id, "started a Chrome")
	}

	// Refused BEFORE the generation is bumped and before "begin" is
	// logged, so an ignored click neither supersedes a genuinely queued
	// drive nor leaves a second "begin" in the log for somebody to puzzle
	// over an hour later.
	//
	// ONLY THE VBA CALLER, for two reasons that happen to agree. A browser
	// click has opened a tab at this address and drive() takes it over;
	// ignoring it would abandon that tab on a reply page, which is the
	// leak that took an evening to find. WinHttp opens nothing, so there
	// is nothing to abandon. It is also the only caller a human can
	// double-click, which is where this was measured.
	//
	// A deliberate second click is NOT this. The reason to click a row
	// twice is usually to see whether a change landed, and that still
	// drives, reloading the tab, exactly as before: the two are told apart
	// by whether the first drive is still running, which is the difference
	// between a double-click and a decision.
	//
	// Nothing is lost by doing nothing. Excel calls
	// AllowSetForegroundWindow before it calls us, so the ignored click
	// has already granted the right to raise the window and the drive that
	// is running may spend it.
	if !s.claim(id) && client == "excel-vba" {
		s.log(id, "clicked again while it is still opening, ignored")
		s.page(w, id, task, "Already opening it", "")
		return
	}

	mine := atomic.AddInt64(&s.gen, 1)
	s.log(id, fmt.Sprintf("begin (%s, %s)", client, raising))
	clicked := "http://" + s.listen + r.URL.RequestURI()

	// ⚠️ ONLY drive_test WAITS. Everybody else is answered first.
	//
	// The VBA caller used to wait here, and that froze Excel for the whole
	// drive: one to three seconds of a dead spreadsheet on every
	// right-click, which is the single thing that made this feel broken.
	//
	// It waited for a reason that stopped being true. The plan was for VBA
	// to raise the window itself with AppActivate, which needs the page
	// TITLE, which does not exist until the page has loaded -- so the reply
	// carried the title and had to come last. Then focus moved in here,
	// where the handle is already known and no title is needed, and the
	// VBA stopped reading X-Chrome-Title. The wait outlived its purpose and
	// nothing noticed, because a thing that is merely slow still works.
	//
	// drive_test keeps the synchronous path under its own name: a reply
	// arriving IS its completion signal, and without it the test goes back
	// to polling a record file, which was the largest cost in a run.
	if client == "drive-test" {
		title := s.driveSync(id, task, "", mine, wantFocus, raisedEarly)
		if title != "" {
			w.Header().Set("X-Chrome-Title", title)
		}
		s.page(w, id, task, "Shown in Chrome", "")
		return
	}
	if client == "excel-vba" {
		// Answered at once, so Excel is free again before the human has
		// let go of the mouse button. `clicked` is empty: WinHttp opened no
		// tab, and hunting for one cost 3.2 of a 3.9 second request.
		s.page(w, id, task, "Opening it", "")
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		go s.drive(id, task, "", mine, wantFocus, true, raisedEarly)
		return
	}

	// A browser asked, so it gets the reply first and the tab it opened
	// becomes the page.
	s.page(w, id, task, "Opening it", "")
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
	go s.drive(id, task, clicked, mine, wantFocus, true, raisedEarly)
}

// driveSync runs the work and returns the tab title to raise, or "".
//
// `clicked` is deliberately EMPTY from here. A VBA caller is WinHttp, not a
// browser, so no tab was ever opened at our address and hunting for one
// costs the full search budget before giving up. Measured 3 Oct: 3.2 of a
// 3.9 second request was spent looking for a tab that cannot exist.
func (s *server) driveSync(id string, task Task, clicked string,
	mine int64, focus bool, raised bool) string {
	done := make(chan string, 1)
	// A test is never superseded: see drive().
	go func() { done <- s.drive(id, task, clicked, mine, focus, false, raised) }()
	select {
	case title := <-done:
		return title
	case <-time.After(30 * time.Second):
		// The caller is Excel and Excel is frozen while it waits, so there
		// is a hard ceiling on how long it may be made to wait.
		s.log(id, "gave up waiting to answer the VBA caller")
		return ""
	}
}

// claim marks a task as on its way, and reports whether it already was.
//
// NOT GUARDED BY s.mu, which is held for the whole of a drive: asking it
// anything from a handler means waiting for that drive to finish, and a
// caller that is about to be told "already opening" is precisely the one
// that must not wait for it.
// WITHIN coalesceWindow ONLY. The first version refused for as long as the
// drive ran, which made the window as long as the slowest page: a human
// clicked tl-e1-01, saw nothing for eleven seconds, clicked again and had it
// swallowed. Past the window the newer click is allowed through, because at
// that point it is not a repeat of the first, it is a verdict on it.
func (s *server) claim(id string) bool {
	s.flyMu.Lock()
	defer s.flyMu.Unlock()
	if at, ok := s.flying[id]; ok && time.Since(at) < coalesceWindow {
		return false
	}
	s.flying[id] = time.Now()
	return true
}

// release forgets a claim.
//
// Deferred by drive() rather than written at the end of it, because drive
// gives up in several places before it reaches Chrome: no browser,
// superseded, dial failed. A claim that outlived its drive would wedge that
// one cell for the rest of the session, and the symptom would be a cell
// that has simply stopped responding to clicks -- no error, nothing in the
// log but a line saying the click was ignored.
func (s *server) release(id string) {
	s.flyMu.Lock()
	defer s.flyMu.Unlock()
	delete(s.flying, id)
}

// remember records the tab we are using for a page and marks it newest.
//
// Called under s.mu, like every other registry touch, so the slice and the
// map cannot disagree about which keys exist.
func (s *server) remember(key, targetID string) {
	s.tabs[key] = targetID
	for i, k := range s.tabAge {
		if k == key {
			s.tabAge = append(s.tabAge[:i], s.tabAge[i+1:]...)
			break
		}
	}
	s.tabAge = append(s.tabAge, key)
}

func (s *server) forget(key string) {
	delete(s.tabs, key)
	for i, k := range s.tabAge {
		if k == key {
			s.tabAge = append(s.tabAge[:i], s.tabAge[i+1:]...)
			return
		}
	}
}

// trimTabs closes our oldest tabs until the registry is back under the cap.
//
// OURS ONLY. The registry holds tabs this process opened, so a tab the human
// opened is not a candidate however old it is, and that is the whole safety
// argument: the alternative, closing by age across the browser, would one day
// close the thing somebody was reading.
//
// The tab just used is the newest and therefore always the last to go.
func (s *server) trimTabs(sess *cdpSession, id string) int {
	if s.maxTabs <= 0 {
		return 0
	}
	closed := 0
	for len(s.tabAge) > s.maxTabs {
		key := s.tabAge[0]
		if tid := s.tabs[key]; tid != "" {
			sess.closeTab(tid)
			closed++
		}
		s.forget(key)
	}
	if closed > 0 {
		s.log(id, fmt.Sprintf("closed %d tab(s) over the cap of %d",
			closed, s.maxTabs))
	}
	return closed
}

// current reports whether this request is still the newest one.
func (s *server) current(mine int64) bool {
	return atomic.LoadInt64(&s.gen) == mine
}

// wantsFocus decides whether THIS request should raise Chrome.
//
// PER REQUEST, not per server, and that distinction is the whole point. A
// server-wide -focus=false was the only way to stop a test run stealing the
// desktop, and leaving one set is exactly what broke a human's right-click
// for an afternoon: the drives worked, in a background tab, and the only
// symptom was focus never arriving. There is no mode to leave set now.
//
// The caller states its intent and the default follows from who is asking.
// A test should not interrupt anything; a human who just right-clicked a
// cell is waiting to be shown something.
//
// AN ALLOW-LIST, NOT A DENY-LIST, and the deny-list was the fault. The rule
// read `!= "drive-test"`, which is a deny-list with exactly one entry: any
// caller that did not send that precise string was treated as a human and
// raised the window. It did not protect the desktop, it protected it from
// the one tool that remembered to say its name, and every other tool in the
// repo that fires a drive had to remember too or be loud by accident.
//
// So the raise is EARNED by saying who you are. A caller that forgets to
// identify itself is quiet rather than loud, which is the right way round:
// a missed raise is a window you have to click on, while an unwanted one
// takes the desktop away from whatever you were typing into.
//
// THE PRICE, stated because it is a real loss. A plain browser click no
// longer raises Chrome, and focus_test asserted that it should until this
// changed it. It is not recoverable by being cleverer, either: a browser
// sends no X-Drive-Client and neither does a script that forgot, so the two
// are indistinguishable at the only point where the decision is made.
// Anybody who wants it back for one request can say &focus=1. It is
// tolerable because a browser that has just opened a tab is usually in
// front already.
//
// Precedence, narrowest last:
//
//	-focus=false        the whole server stays out of the way. Still
//	                    honoured, because "leave my desktop alone while I
//	                    work" is a real thing to want.
//	?focus=0 / ?focus=1 whatever the caller says, for anything that needs
//	                    to override its own default.
//	X-Drive-Client      on the allow-list below, or no raise.
func (s *server) wantsFocus(r *http.Request) bool {
	if !s.focus {
		return false
	}
	if v := r.URL.Query().Get("focus"); v != "" {
		return v != "0" && !strings.EqualFold(v, "false")
	}
	return raisesByDefault[r.Header.Get("X-Drive-Client")]
}

// raisesByDefault names every caller whose drives raise the window without
// being asked. ONE ENTRY, and adding a second should take an argument: the
// only thing that belongs here is a caller a human is waiting on.
var raisesByDefault = map[string]bool{
	"excel-vba": true, // a human just right-clicked a cell
}

// raiseEarly brings the tab we are about to drive to the front NOW.
//
// Focus used to arrive at the END of a drive, one to three seconds after
// the click, because the window is found by page title and there is no new
// title until the navigation has finished. But the tab we are about to
// reuse already HAS a title, and its window is the window we want. So the
// raise can happen before the navigation rather than after it, and the
// human sees Chrome come forward immediately and the page fill in a moment
// later -- which is the right way round.
//
// Reports whether it raised, so the caller can skip the late raise rather
// than yanking the desktop twice.
func (s *server) raiseEarly(sess *cdpSession, id, targetID string,
	focus bool) bool {
	if !focus || targetID == "" {
		// No tab yet means a brand new one, whose title is "about:blank"
		// and matches no window worth raising. Those keep the late raise.
		return false
	}
	if sid, ok := sess.attachTo(targetID); ok {
		// Inside Chrome only. Necessary and not sufficient: it selects the
		// tab and does not raise the window.
		sess.send(sid, "Page.bringToFront", nil)
	}
	title := sess.titleOf(targetID)
	if title == "" {
		return false
	}
	outcome := RaiseChromeTitled(title)
	s.log(id, "focus at start: "+outcome)
	return strings.HasPrefix(outcome, "raised")
}

// drive does the Chrome work after the reply has gone out.
func (s *server) drive(id string, task Task, clicked string, mine int64,
	focus bool, supersedable bool, raisedEarly bool) string {
	started := time.Now()
	defer s.release(id)
	// ONE AT A TIME. Two clicks in quick succession would otherwise race to
	// bring their own tab forward and the human would see whichever won.
	s.mu.Lock()
	defer s.mu.Unlock()

	// Checked AFTER the wait for the lock, which is exactly where a stale
	// request has been sitting while newer clicks arrived.
	//
	// NOT FOR A TEST. "They want the LAST one" is true of a human clicking
	// four rows and false of a suite enumerating every link in the sheet: a
	// test asks for all of them deliberately and a dropped one is a link
	// left unproved. drive_test runs concurrently, so a queued request
	// routinely had its generation overtaken, and the suite then waited out
	// its full 30 second timeout looking for a tab no drive had opened.
	// tl-e3-01:edit failed exactly that way and the message blamed the tab.
	if supersedable && !s.current(mine) {
		s.log(id, fmt.Sprintf("superseded by a newer click after %s",
			took(started)))
		return ""
	}

	sess, err := dialCDP(s.cdpPort)
	if err != nil {
		s.log(id, "no chrome: "+err.Error())
		return ""
	}
	defer sess.Close()

	key := task.Page
	if task.Source {
		key = "view-source:" + task.Page
	}

	// Give the browser a moment to finish loading our reply, or there is no
	// tab at that address yet to find.
	take := ""
	if clicked != "" {
		// Only a browser caller can have left a tab at our address.
		for attempt := 0; attempt < 8; attempt++ {
			if tid, ok := sess.findTabAt(clicked); ok {
				take = tid
				break
			}
			time.Sleep(400 * time.Millisecond)
		}
	}
	tFind := time.Now()
	how := "took over the clicked tab"
	if take == "" {
		take = s.tabs[key]
		how = "own tab"
	}
	// Already up, raised on arrival before this goroutine existed, so
	// neither the early nor the late raise below has anything to do.
	early := raisedEarly
	if take == "" {
		// RECLAIM what an earlier run left, rather than adding to it. The
		// registry above lives in this process only, so every restart
		// forgot every tab it had opened: the next click opened a second
		// tab for the same page and orphaned the first. Four restarts in
		// an evening left four copies of each page and eight of the home
		// page, which is how the leak was noticed.
		//
		// Chrome is the durable record of what is already open, so ask it
		// before opening anything -- then close what it finds. Inheriting
		// such a tab was tried first and is worse than useless: see
		// closeTabsAt. One closed, one opened, count unchanged.
		// Raised on the DOOMED tab, before it closes. It sits in the
		// window we want, and once it is gone there is no title to match
		// until the replacement page has loaded -- which is the whole
		// three seconds this is here to remove. The first click on a page
		// after a restart is the common case, not a rare one.
		// NOT reassigned when `early` is already true. The window went up on
		// arrival, and this would overwrite that with whatever a second,
		// later raise returned, which on a doomed tab is often a refusal:
		// asking twice and then believing the worse answer would report
		// "refused" about a window already in front of the human.
		if tid, ok := sess.findTabAt(key); ok && !early {
			early = s.raiseEarly(sess, id, tid, focus)
		}
		if n := sess.closeTabsAt(key); n > 0 {
			how = fmt.Sprintf("replaced %d tab(s) left by an earlier run", n)
		}
	}

	if !early {
		early = s.raiseEarly(sess, id, take, focus)
	}
	targetID, reused, err := sess.show(task.Page, task.Find, task.Source,
		focus, take, strings.ToLower(task.Expect), s.absentBudget,
		task.LocatorRE)
	// STAMPED HERE, BEFORE THE HOUSEKEEPING. This used to be taken after
	// the registry work below, so closing tabs was billed to `show` and the
	// log read "show 3.7s" for a drive that had spent part of that on
	// tidying. The one question somebody asked of this line was whether the
	// delay was the tab cap, and the line had been built so that it could
	// not answer.
	tShow := time.Now()
	title := sess.titleOf(targetID)

	// THE WINDOW BEFORE THE TIDYING. Raising it used to come after the cap
	// closed tabs, so housekeeping sat between the human and the thing they
	// had asked to look at. It only cost a fraction of a second, which is
	// not the point: nothing whose purpose is to save memory later belongs
	// in front of the one action somebody is waiting on.
	//
	// Still only when the early raise could not run or was refused. Raising
	// twice is a second yank for no gain.
	if focus && title != "" && !early {
		s.log(id, "focus: "+RaiseChromeTitled(title))
	}
	tFocus := time.Now()

	if targetID != "" {
		s.remember(key, targetID)
		// LRU, AND THE REASON IT BECAME NECESSARY TODAY.
		//
		// Driving 21 rows with four drives each leaves a tab per page and
		// nothing ever closed one: 41 page tabs, 95 Chrome processes and
		// 8GB of working set, reported as "sluggish" before anybody
		// counted it. Reclaiming on restart bounded the leak ACROSS runs
		// and never within one.
		//
		// It got worse this afternoon by my own hand.
		// Page.setWebLifecycleState active now goes out before every
		// navigation, so every tab we drive is told it is being looked at
		// and keeps running. The freezing that made a pile of background
		// tabs cheap is precisely what that defeats, so the throttling fix
		// and this cap are two halves of one decision.
		//
		// Only tabs in OUR registry are closed, so a tab the human opened
		// is never a candidate however old it is.
		if closed := s.trimTabs(sess, id); closed > 0 {
			how += fmt.Sprintf(", closed %d old tab(s)", closed)
		}
	} else {
		s.forget(key)
	}
	if reused && how == "own tab" {
		how = "reused our tab"
	}
	if err != nil {
		s.log(id, fmt.Sprintf("done in %s, %s, NO HIGHLIGHT: %s",
			took(started), how, err.Error()))
		// The page IS open even when the anchor was not found, so it is
		// still worth raising: she is looking at the right page.
		return title
	}
	// Phase timings, because "it feels slow" cannot be acted on and this
	// is how the 3.2s tab hunt was found.
	//
	// FOUR PHASES NOW, NOT TWO. "show 3.7s" was asked to account for a
	// three second delay and could not, because it was the only phase wide
	// enough to hide anything: raising the window and closing tabs over the
	// cap were both inside it. A reader could not tell a slow page from
	// slow housekeeping, and reasonably guessed the housekeeping.
	s.log(id, fmt.Sprintf("done in %s (tab %s, show %s, focus %s, tidy %s), %s",
		took(started), tookFrom(started, tFind),
		tookFrom(tFind, tShow), tookFrom(tShow, tFocus),
		tookFrom(tFocus, time.Now()), how))
	// Any OTHER tab still sitting on one of our replies is tidied away. The
	// one we took over is no longer at that address, so it is not caught.
	if n := sess.closeOurOwnPages("http://" + s.listen + "/show/"); n > 0 {
		log.Printf("  closed %d leftover reply tab(s)", n)
	}
	return title
}

// page is what the clicking browser gets. NOT 204: a blank tab reads as a
// broken link, and the one thing the human needs is to know it worked and
// where to look.
func (s *server) page(w http.ResponseWriter, id string, task Task, headline, detail string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// ⚠️ "another TAB", not "the other WINDOW". The window wording was
	// written when this drove a second Chrome on another port. Driving the
	// Chrome the click landed in is the better arrangement, and it made the
	// sentence wrong: there is no other window.
	//
	// And the fallback is Fallback, not Page. Page is usually
	// view-source:..., which Chrome refuses to open from a click, so
	// printing it was handing somebody a string that cannot work at the
	// moment they are already lost.
	fallback := task.Fallback
	note := "If you cannot see it, open this instead:"
	if fallback == "" {
		fallback = task.Page
		note = "If you cannot see it, the address is:"
	}
	body := fmt.Sprintf(`<!doctype html><meta charset="utf-8">
<title>%s</title>
<style>body{font:16px/1.5 system-ui,sans-serif;margin:3rem auto;max-width:42rem;color:#222}
a{color:#0b5}.d{color:#a33}small{color:#666}</style>
<h1>%s</h1>
<p><strong>%s</strong> %s</p>
<p class="d">%s</p>
<p><strong>Working.</strong> Nothing has happened yet: this tab is the
request, and it becomes the page you asked for when Chrome has it.</p>
<p>%s<br><a href="%s">%s</a></p>
<p><small>If this page is still here after a few seconds, the click landed
in a browser that is not the one being driven, or the page had nothing
matching on it. The console window running remote_chrome_cdp says which.
The link above goes to the same place either way.</small></p>`,
		html.EscapeString(id), html.EscapeString(headline),
		html.EscapeString(id), html.EscapeString(task.What),
		html.EscapeString(detail), html.EscapeString(note),
		html.EscapeString(fallback), html.EscapeString(fallback))
	fmt.Fprint(w, body)
}

// noChrome is the honest answer when there is nothing to drive.
//
// It says what is missing, how to fix it, and gives an address that works
// without any of this, because somebody mid-call needs the page more than
// they need the tooling.
func (s *server) noChrome(w http.ResponseWriter, id string, task Task, err error) {
	fallback := task.Fallback
	if fallback == "" {
		fallback = task.Page
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusServiceUnavailable)
	fmt.Fprintf(w, `<!doctype html><meta charset="utf-8">
<title>%s: no Chrome to drive</title>
<style>body{font:16px/1.5 system-ui,sans-serif;margin:3rem auto;max-width:42rem;color:#222}
a{color:#0b5}code{background:#f4f4f4;padding:.1rem .3rem}small{color:#666}</style>
<h1>Nothing to drive</h1>
<p><strong>%s</strong> %s</p>
<p>No Chrome is listening on port %d, so this could not open anything. An
ordinary Chrome window will not do: it has to be started with the debugging
port, which is what lets anything drive it.</p>
<p><code>chrome.exe --remote-debugging-port=%d</code></p>
<p>Meanwhile this goes to the same place without any tooling:<br>
<a href="%s">%s</a></p>
<p><small>%s</small></p>`,
		html.EscapeString(id), html.EscapeString(id),
		html.EscapeString(task.What), s.cdpPort, s.cdpPort,
		html.EscapeString(fallback), html.EscapeString(fallback),
		html.EscapeString(err.Error()))
}

func (s *server) deny(w http.ResponseWriter, code int, why string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)
	fmt.Fprintf(w, `<!doctype html><meta charset="utf-8">
<style>body{font:16px/1.5 system-ui,sans-serif;margin:3rem auto;max-width:42rem}</style>
<h1>No</h1><p>%s</p>`, html.EscapeString(why))
}

// log appends one line per request. For a program that drives a browser over
// somebody's live website, what was asked for and when is worth keeping
// whether or not anybody ever reads it.
func (s *server) log(id, outcome string) {
	line := fmt.Sprintf("%s  %-8s %s", time.Now().Format(time.RFC3339), id, outcome)
	log.Print("  " + line)
	if s.record == "" {
		return
	}
	f, err := os.OpenFile(s.record, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.Printf("  could not write the record: %v", err)
		return
	}
	defer f.Close()
	entry, _ := json.Marshal(map[string]string{
		"at": time.Now().UTC().Format(time.RFC3339), "task": id, "outcome": outcome})
	f.Write(append(entry, '\n'))
}
