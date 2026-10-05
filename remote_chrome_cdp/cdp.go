// cdp.go -- the Chrome half: open a tab we own, show one thing, close it.
//
// THE OP VOCABULARY IS CLOSED, AND Runtime.evaluate IS NOT IN IT.
//
// Everything a caller can ask for is in this file, parameterised by a URL
// and a string to look for. The JavaScript is ours. If the wire could carry
// JavaScript instead, this would be `exec` with extra steps: Runtime.evaluate
// reads any page the browser is attached to, so a caller who can send script
// can read anything that browser is signed into. That is the whole reason
// remote_ai gates `exec` behind a kernel cage, and the reason this program
// does not offer it at all.
//
// WE OPEN OUR OWN TAB, ALWAYS.
//
// Attaching to whatever is already open was tried, in the Python tooling this
// replaces, on 2 October. It drove a tab belonging to another program away
// from its own page and then reported a confident result measured on the
// wrong document, which is worse than failing. A tab we created is a tab
// nobody else is using, and it is the only kind we touch.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

const cdpTimeout = 30 * time.Second

type cdpSession struct {
	ws   *wsConn
	next int64
}

type cdpReply struct {
	ID     int64           `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Message string `json:"message"`
	} `json:"error"`
	Method string `json:"method"`
}

// browserWS asks Chrome where its browser-level endpoint is.
//
// /json/version rather than /json/list: we want the BROWSER target so we can
// create our own tab, not one of the pages somebody else is reading.
func browserWS(port int) (string, error) {
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/json/version", port))
	if err != nil {
		return "", fmt.Errorf("no Chrome answering on 127.0.0.1:%d. Start one "+
			"with --remote-debugging-port=%d: %w", port, port, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var v struct {
		WS string `json:"webSocketDebuggerUrl"`
	}
	if err := json.Unmarshal(body, &v); err != nil {
		return "", err
	}
	if v.WS == "" {
		return "", fmt.Errorf("Chrome on port %d gave no webSocketDebuggerUrl", port)
	}
	return v.WS, nil
}

func dialCDP(port int) (*cdpSession, error) {
	endpoint, err := browserWS(port)
	if err != nil {
		return nil, err
	}
	ws, err := wsDial(endpoint, 15*time.Second)
	if err != nil {
		return nil, err
	}
	return &cdpSession{ws: ws}, nil
}

func (s *cdpSession) Close() error { return s.ws.Close() }

// send issues one command and waits for the reply with the matching id.
//
// Events arrive on the same socket and are discarded here. That is right for
// a program whose whole job is request and reply: a reader that treated an
// event as its answer would return the wrong thing rather than nothing.
func (s *cdpSession) send(sessionID, method string, params map[string]any) (json.RawMessage, error) {
	id := atomic.AddInt64(&s.next, 1)
	msg := map[string]any{"id": id, "method": method}
	if params != nil {
		msg["params"] = params
	}
	if sessionID != "" {
		msg["sessionId"] = sessionID
	}
	raw, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	if err := s.ws.writeText(raw); err != nil {
		return nil, err
	}
	deadline := time.Now().Add(cdpTimeout)
	for {
		data, err := s.ws.readText(deadline)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", method, err)
		}
		var reply cdpReply
		if err := json.Unmarshal(data, &reply); err != nil {
			continue
		}
		if reply.ID != id {
			continue // an event, or another command's reply
		}
		if reply.Error != nil {
			return nil, fmt.Errorf("%s: %s", method, reply.Error.Message)
		}
		return reply.Result, nil
	}
}

// waitEvent reads frames until `method` arrives, or the deadline passes.
//
// Replaces a four second sleep. The sleep was not a measurement of anything:
// it was a number that worked, paid in full on every single click whether
// the page took 200ms or three seconds. A load event is the page itself
// saying when it is ready.
//
// Returns false on timeout rather than erroring, because a missing load
// event is not a failure: view-source and some single-page apps do not fire
// a useful one, and the caller then falls through to looking for the text,
// which is the real test anyway.
func (s *cdpSession) waitEvent(method string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		data, err := s.ws.readText(deadline)
		if err != nil {
			return false
		}
		var frame struct {
			Method string `json:"method"`
		}
		if json.Unmarshal(data, &frame) == nil && frame.Method == method {
			return true
		}
	}
	return false
}

// openTab creates a tab and attaches to it, returning (targetID, sessionID).
func (s *cdpSession) openTab() (string, string, error) {
	// background:true, so creating the tab does not yank the desktop away
	// from whatever the human is doing. The tab is raised ONCE, at the end,
	// when there is something on it worth looking at. Without this the
	// sequence steals focus twice: first for a blank tab, then again after
	// it has loaded.
	res, err := s.send("", "Target.createTarget", map[string]any{
		"url": "about:blank", "background": true})
	if err != nil {
		return "", "", err
	}
	var created struct {
		TargetID string `json:"targetId"`
	}
	if err := json.Unmarshal(res, &created); err != nil {
		return "", "", err
	}
	res, err = s.send("", "Target.attachToTarget", map[string]any{
		"targetId": created.TargetID, "flatten": true})
	if err != nil {
		return created.TargetID, "", err
	}
	var attached struct {
		SessionID string `json:"sessionId"`
	}
	if err := json.Unmarshal(res, &attached); err != nil {
		return created.TargetID, "", err
	}
	return created.TargetID, attached.SessionID, nil
}

// attachTo re-attaches to a tab we opened earlier, or reports that it is
// gone. The human is free to close our tab and usually will, so its absence
// is an ordinary outcome rather than an error.
func (s *cdpSession) attachTo(targetID string) (string, bool) {
	res, err := s.send("", "Target.getTargets", nil)
	if err != nil {
		return "", false
	}
	var list struct {
		TargetInfos []struct {
			TargetID string `json:"targetId"`
			Type     string `json:"type"`
		} `json:"targetInfos"`
	}
	if json.Unmarshal(res, &list) != nil {
		return "", false
	}
	found := false
	for _, t := range list.TargetInfos {
		if t.TargetID == targetID && t.Type == "page" {
			found = true
			break
		}
	}
	if !found {
		return "", false
	}
	res, err = s.send("", "Target.attachToTarget", map[string]any{
		"targetId": targetID, "flatten": true})
	if err != nil {
		return "", false
	}
	var attached struct {
		SessionID string `json:"sessionId"`
	}
	if json.Unmarshal(res, &attached) != nil || attached.SessionID == "" {
		return "", false
	}
	return attached.SessionID, true
}

// findTabAt returns the target id of a tab sitting on exactly `url`.
//
// This is how the tab the human CLICKED gets taken over rather than
// orphaned. It is on our own /show/ address, it exists only because they
// clicked our link, and turning it into the page they asked for is what
// they expected the link to do. Matching the FULL url, task id and token
// included, keeps two quick clicks from stealing each other's tab.
//
// Only works when the browser that handled the click is the Chrome we
// drive. When it is not, there is no such tab and the caller opens its own.
func (s *cdpSession) findTabAt(url string) (string, bool) {
	res, err := s.send("", "Target.getTargets", nil)
	if err != nil {
		return "", false
	}
	var list struct {
		TargetInfos []struct {
			TargetID string `json:"targetId"`
			Type     string `json:"type"`
			URL      string `json:"url"`
		} `json:"targetInfos"`
	}
	if json.Unmarshal(res, &list) != nil {
		return "", false
	}
	for _, t := range list.TargetInfos {
		if t.Type == "page" && t.URL == url {
			return t.TargetID, true
		}
	}
	return "", false
}

// closeTabsAt closes every tab sitting on exactly `url`, returning the count.
//
// WHY CLOSE RATHER THAN REUSE. A tab left by an EARLIER server process looks
// reusable and is not: the first navigation of one reliably ends with the
// page loaded and the highlight missing, having cost 30 seconds to get
// there. Measured 4 Oct 2026 -- adopting orphans failed on the first touch
// of each one and then worked forever after, which is the signature of the
// tab, not of the page.
//
// Reusing a tab inside ONE process is a different thing and still happens:
// that is how a row's before and after links share a tab. This is only for
// tabs whose creator is gone, and for those, a clean tab is worth more than
// a saved tab. Closing one and opening one keeps the count flat, which was
// the whole point.
func (s *cdpSession) closeTabsAt(url string) int {
	closed := 0
	// Re-asked each time: the ids shift as tabs close.
	for attempt := 0; attempt < 12; attempt++ {
		tid, ok := s.findTabAt(url)
		if !ok {
			break
		}
		if s.closeTab(tid) != nil {
			break
		}
		closed++
	}
	return closed
}

// closeOurOwnPages closes tabs sitting on this server's /show/ URLs.
//
// Clicking a link opens a browser tab to ASK us, and that tab then sits
// there showing our reply. When the Chrome we drive is the same one the
// click landed in, the result is two tabs per click: the one the human
// wanted and the one that asked for it.
//
// Matched on OUR OWN address and only the /show/ path, so the worst case is
// closing a page of ours that somebody opened deliberately. Any other tab
// is untouched.
func (s *cdpSession) closeOurOwnPages(prefix string) int {
	res, err := s.send("", "Target.getTargets", nil)
	if err != nil {
		return 0
	}
	var list struct {
		TargetInfos []struct {
			TargetID string `json:"targetId"`
			Type     string `json:"type"`
			URL      string `json:"url"`
		} `json:"targetInfos"`
	}
	if json.Unmarshal(res, &list) != nil {
		return 0
	}
	closed := 0
	for _, t := range list.TargetInfos {
		if t.Type == "page" && strings.HasPrefix(t.URL, prefix) {
			if s.closeTab(t.TargetID) == nil {
				closed++
			}
		}
	}
	return closed
}

// titleOf returns the tab's title, which is what AppActivate needs.
//
// Chrome's WINDOW title is "<tab title> - Google Chrome", and VBA's
// AppActivate matches any window whose title BEGINS with what it is given,
// so the tab title alone is enough and we never have to guess at the suffix
// or enumerate windows.
func (s *cdpSession) titleOf(targetID string) string {
	res, err := s.send("", "Target.getTargetInfo",
		map[string]any{"targetId": targetID})
	if err != nil {
		return ""
	}
	var info struct {
		TargetInfo struct {
			Title string `json:"title"`
		} `json:"targetInfo"`
	}
	if json.Unmarshal(res, &info) != nil {
		return ""
	}
	return info.TargetInfo.Title
}

func (s *cdpSession) closeTab(targetID string) error {
	_, err := s.send("", "Target.closeTarget", map[string]any{"targetId": targetID})
	return err
}

func (s *cdpSession) eval(sessionID, expr string) (json.RawMessage, error) {
	res, err := s.send(sessionID, "Runtime.evaluate", map[string]any{
		"expression": expr, "returnByValue": true, "awaitPromise": true})
	if err != nil {
		return nil, err
	}
	var out struct {
		Result struct {
			Value json.RawMessage `json:"value"`
		} `json:"result"`
	}
	if err := json.Unmarshal(res, &out); err != nil {
		return nil, err
	}
	return out.Result.Value, nil
}

// markJS finds `needle`, scrolls it into the middle of the screen and wraps
// it in a yellow <mark>. Returns {found, spans}.
//
// ⚠️ IT SEARCHES ACROSS TEXT NODES, and it has to.
//
// Chrome's view-source syntax-highlights HTML attributes, so
// `property="og:locale" content="en_US"` is split over several spans and NO
// single text node contains it. A per-node search finds strings inside a
// <script> block, where view-source emits one plain run, and silently fails
// on every meta tag. That is the worst possible split, because the schema
// operations would have worked and the title and locale ones would not,
// which looks like the tool working.
//
// So the text nodes are concatenated, the needle is found in the whole, and
// the offset is mapped back to the node it started in.
//
// A RANGE, NOT AN ELEMENT, for the scroll: on view-source the element
// holding a match is one minified line that can be a hundred thousand
// characters wide, so scrolling to the element puts the match off-screen.
const markJS = `(function (needle) {
  // THE TOP DOCUMENT, THEN ANY SAME-ORIGIN FRAME INSIDE IT.
  //
  // Elementor renders the page being edited in an iframe and keeps only its
  // own panel in the top document, so "Affordable family entertainment" was
  // not findable on the screen that edits it: the anchor search walked one
  // document and the words were in another. Same for the block editor, which
  // moved post content into a frame.
  //
  // Searched one document at a time, top first, rather than by concatenating
  // them. That keeps every offset inside the document it came from, so the
  // range arithmetic below is unchanged and a match in the top document
  // behaves exactly as it did before this existed.
  //
  // Same-origin only, and not by choice: reading contentDocument across
  // origins throws, and a cross-origin frame is one we are not allowed to
  // look into. Wrapped in try so a Google font frame or an ad cannot stop
  // the search.
  function documents() {
    var out = [document];
    for (var pass = 0; pass < out.length && out.length < 12; pass++) {
      var frames = out[pass].querySelectorAll
        ? out[pass].querySelectorAll("iframe,frame") : [];
      for (var k = 0; k < frames.length; k++) {
        try {
          var d = frames[k].contentDocument;
          if (d && d.body && out.indexOf(d) < 0) out.push(d);
        } catch (e) { /* cross-origin: not ours to read */ }
      }
    }
    return out;
  }

  function controlAround(el) {
    var sel = "select,[role=combobox],[role=listbox],input,textarea," +
              "button,mat-select";
    var p = el;
    for (var k = 0; k < 7 && p; k++) {
      if (p.matches && p.matches(sel)) return p;
      p = p.parentElement;
    }
    return null;
  }

  function reveal(el) {
    var ctl = controlAround(el);
    if (ctl) {
      ctl.style.outline = "3px solid #d08700";
      ctl.style.outlineOffset = "2px";
    }
    (ctl || el).scrollIntoView({block: "center", inline: "center"});
    return ctl ? ctl.tagName.toLowerCase() : "";
  }

  // markIn is the whole of what this function used to be, scoped to one
  // document so it can be tried against several.
  function markIn(doc, which) {
    var w = doc.createTreeWalker(doc.body, NodeFilter.SHOW_TEXT);
    var nodes = [], text = "", n;
    while ((n = w.nextNode())) {
      if (!n.nodeValue) continue;
      nodes.push({node: n, at: text.length, len: n.nodeValue.length});
      text += n.nodeValue;
    }
    // A MATCH NOBODY CAN SEE IS WORSE THAN NO MATCH. markJS walks every
    // text node, displayed or not, so a string sitting in a hidden template
    // counts: Elementor keeps a copy of the page's words in its own panel
    // markup, and the first occurrence of the homepage headline was one of
    // those. It marked something invisible, scrolled nowhere, and reported
    // success.
    //
    // So occurrences are tried in order and the first VISIBLE one wins.
    // Checked per occurrence rather than per node on purpose: a view-source
    // page has thousands of text nodes and measuring every one of them to
    // find one string would cost more than the search.
    function visibleAt(offset) {
      var hit = locate(offset);
      var el = hit.node.parentElement;
      if (!el) return false;
      if (el.getClientRects && el.getClientRects().length === 0) return false;
      var box = el.getBoundingClientRect();
      return box.width > 0 && box.height > 0;
    }

    var i = -1, len = needle.length;
    for (var at = text.indexOf(needle); at >= 0;
         at = text.indexOf(needle, at + 1)) {
      if (visibleAt(at)) { i = at; break; }
      if (i < 0) { i = -2; }          // seen, but not on screen
    }
    if (i === -2) {
      // Every occurrence was hidden. Say so rather than marking one: the
      // caller's "could not find it" is then true of the screen, which is
      // what the human is looking at.
      return null;
    }
    if (i < 0) {
      // A SECOND LOOK, WITH WHITESPACE TREATED AS WHITESPACE.
      //
      // The text above is raw nodeValues concatenated and nothing collapses
      // the gaps. Google Analytics renders a dropdown's value as a number in
      // one element and its unit in another, so the retention control reads
      // as 17 spaces, "2", a newline, 22 spaces and "months" where innerText
      // says "2 months". indexOf could never match that, and the drive
      // reported the anchor missing on a screen showing it in large type.
      // Any anchor whose words cross an element boundary was in the same
      // position: unmatchable, and indistinguishable from wrong.
      //
      // Searched against the RAW text rather than a normalised copy, because
      // the match index has to stay valid for locate() below, and remapping
      // offsets through a collapse is the kind of arithmetic that is wrong
      // once and then wrong for ever.
      var rx;
      try {
        rx = new RegExp(needle.trim().split(/\s+/).map(function (x) {
          return x.replace(/[-.*+?^${}()|[\]\\]/g, "\\$&");
        }).join("\\s+"));
      } catch (e) { return null; }
      var m = rx.exec(text);
      if (!m) return null;
      i = m.index;
      len = m[0].length;
    }

    function locate(offset) {
      for (var k = 0; k < nodes.length; k++) {
        if (offset < nodes[k].at + nodes[k].len) {
          return {node: nodes[k].node, off: offset - nodes[k].at};
        }
      }
      var last = nodes[nodes.length - 1];
      return {node: last.node, off: last.len};
    }

    var a = locate(i), b = locate(i + len - 1);
    var rg = doc.createRange();
    rg.setStart(a.node, a.off);
    rg.setEnd(b.node, b.off + 1);
    var marked = false, control = "";
    try {
      var mark = doc.createElement("mark");
      mark.style.background = "#ffe14d";
      mark.style.outline = "3px solid #d08700";
      rg.surroundContents(mark);
      control = reveal(mark);
      marked = true;
    } catch (e) {
      // surroundContents refuses a range that crosses element boundaries,
      // which is exactly the attribute case. So EVERY node the match covers
      // is highlighted, not just the first: styling only the start node
      // marks the word "property" and leaves the value it is pointing at
      // unmarked, which reads as the tool finding the wrong thing.
      var from = i, to = i + len;
      for (var k2 = 0; k2 < nodes.length; k2++) {
        var s0 = nodes[k2].at, s1 = s0 + nodes[k2].len;
        if (s1 <= from || s0 >= to) continue;
        var host = nodes[k2].node.parentElement;
        if (!host) continue;
        host.style.background = "#ffe14d";
        host.style.outline = "2px solid #d08700";
        if (!marked) { control = reveal(host); }
        marked = true;
      }
    }
    if (!marked) return null;
    return {found: true, spans: a.node === b.node ? 1 : 2,
            control: control, where: which};
  }

  var docs = documents();
  for (var d = 0; d < docs.length; d++) {
    try {
      var got = markIn(docs[d], d === 0 ? "page" : "frame " + d);
      if (got) return got;
    } catch (e) { /* a frame that went away mid-search */ }
  }
  return {found: false, documents: docs.length};
})(%s)`

// lineJS highlights one source line of a view-source page and scrolls to it.
//
// Chrome renders view-source as a table with one row per source line, so a
// line number is directly addressable and needs no parsing here. The parsing
// already happened in Python, against the same bytes Chrome is showing.
const lineJS = `(function (pat) {
  var rx;
  try { rx = new RegExp(pat, "i"); } catch (e) { return {found: false}; }
  var rows = document.querySelectorAll('table tr');
  for (var i = 0; i < rows.length; i++) {
    var text = rows[i].textContent || "";
    if (!rx.test(text)) continue;
    var tr = rows[i];
    tr.style.background = "#ffe14d";
    tr.style.outline = "3px solid #d08700";
    tr.scrollIntoView({block: "center", inline: "nearest"});
    return {found: true, line: i + 1, text: text.trim().slice(0, 90)};
  }
  return {found: false, rows: rows.length};
})(%s)`

// show opens or REUSES a tab for this page, marks the string and optionally
// raises the window. Returns the target id so the caller can offer it back.
//
// REUSE IS BY PAGE, NOT BY TASK, and that is the useful way round: two tasks
// often examine the same page. E11 and E12 both live in the homepage's head,
// so keyed by task they would fill the screen with identical view-source
// tabs, and keyed by page they share one that gets re-marked.
//
// A REUSED TAB IS STILL RELOADED. The whole point of clicking a second time
// is usually to see what changed, and a cached DOM would faithfully show the
// old value with a fresh highlight on it. That is the one failure here that
// would look exactly like success.
func (s *cdpSession) show(pageURL, find string, viewSource bool, focus bool,
	reuse string, expect string, absent time.Duration,
	locatorRE string) (string, bool, error) {
	var targetID, sessionID string
	var err error
	reused := false
	if reuse != "" {
		if sid, ok := s.attachTo(reuse); ok {
			targetID, sessionID, reused = reuse, sid, true
		}
	}
	if !reused {
		targetID, sessionID, err = s.openTab()
		if err != nil {
			return "", false, err
		}
	}
	target := pageURL
	if viewSource {
		target = "view-source:" + pageURL
	}
	// Page.enable BEFORE navigating, or the load event for this navigation
	// is never delivered.
	s.send(sessionID, "Page.enable", nil)

	// A BACKGROUND TAB DOES NOT ALWAYS PAINT, and a heavy single-page app
	// is the case that bites. Chrome throttles and freezes pages it thinks
	// nobody is looking at, so the Analytics admin screen loads its shell,
	// never renders the panel, and the anchor search times out on a tab
	// that is open and on exactly the right address.
	//
	// Page.bringToFront cures it, which is why this only ever failed with
	// focus off, but bringToFront is the wrong tool for the cure: it
	// changes which tab the human is looking at, and that is the one thing
	// a quiet drive must not do. setWebLifecycleState tells the RENDERER
	// the page is active without touching tab order or window order, which
	// is precisely the half that was needed. Unconditional, because
	// telling an already-active page it is active costs nothing.
	//
	// Measured 5 Oct: tl-e1-01:after failed after 31s with no focus and
	// passed in 12.4s with it. It had been passing for days on nothing
	// better than that Chrome window happening to be visible, which is a
	// suite that reports on the desktop as much as on the site.
	s.send(sessionID, "Page.setWebLifecycleState",
		map[string]any{"state": "active"})

	// THE TAB COMES TO FRONT NOW, WHILE IT LOADS, not once it has finished.
	//
	// Two different things were both called focus and both left until last.
	// Raising the WINDOW is an OS call and now happens the moment a request
	// arrives. Activating the TAB is a Chrome call and was still at the end
	// of this function, behind the navigation and the anchor search, so the
	// window came forward showing whichever tab had been active before and
	// sat there for seconds before switching. That reads exactly like a
	// tool struggling to find the right tab, and the tab was never the
	// problem: the log says finding it costs between 1 and 27 milliseconds.
	//
	// Nothing is protected by waiting. A tab brought forward before it
	// loads shows the page being built, which is what a browser does every
	// time anybody clicks a link, and Chrome holds the previous paint until
	// the new document commits rather than flashing white. Being shown the
	// work is better than being shown the wrong tab.
	//
	// Gated on `focus` like the raise it belongs with: a quiet drive moves
	// nothing, which is what -focus=false and a test both rely on.
	if focus {
		s.send(sessionID, "Page.bringToFront", nil)
	}

	navRes, err := s.send(sessionID, "Page.navigate",
		map[string]any{"url": target})
	if err != nil {
		if !reused {
			s.closeTab(targetID)
		}
		return "", reused, err
	}
	// WAIT ONLY FOR A LOAD EVENT THAT CAN ARRIVE.
	//
	// Navigating a tab to the URL it is already on, when that URL is a
	// hash route, is a SAME-DOCUMENT navigation: Chrome changes nothing,
	// fires no load event, and this waited the full twelve seconds every
	// time. It is why reusing a tab was SLOWER than opening one, which is
	// backwards and was the clue: tl-e1-01 cost 7.3s on a fresh tab and
	// 12.2s on a reused one, and 12.2 is this timeout plus the search.
	//
	// Page.navigate answers with a loaderId for a real navigation and
	// omits it for a same-document one, so the reply says which kind
	// happened. Measured against the live Analytics screen: loaderId came
	// back empty and no load event arrived in thirteen seconds.
	//
	// Nothing is lost by skipping the wait. A same-document navigation
	// means the DOM is already there, which is exactly when the anchor
	// search can start immediately.
	var nav struct {
		LoaderID string `json:"loaderId"`
	}
	sameDoc := json.Unmarshal(navRes, &nav) != nil || nav.LoaderID == ""
	if !sameDoc {
		s.waitEvent("Page.loadEventFired", 12*time.Second)
	}

	if find != "" {
		quoted, err := json.Marshal(find)
		if err != nil {
			return targetID, reused, err
		}
		// POLLED, not slept. Two seconds between tries meant a page that
		// was ready at 2.1s cost four. A single-page app can take a moment
		// to put the text in the DOM after its load event, so this keeps
		// asking cheaply rather than waiting expensively.
		var found struct {
			Found bool `json:"found"`
		}
		// A short look when the slate says it should not be there yet.
		// Long enough to catch the change having already been made, short
		// enough that confirming the expected costs nothing.
		//
		// 25s, AND THAT IS A MEASUREMENT. It was 8s, and all four Elementor
		// edit drives failed with "tab is open but X is NOT highlighted" on
		// anchors a manual search found instantly, which looked like the
		// anchors being wrong. Timed on 5 Oct: the homepage headline appears
		// in Elementor's panel markup at 1.8s, HIDDEN, and in the visible
		// preview frame at 16.8 seconds. Elementor loads the editor shell,
		// then the page, then paints it, and only the last of those puts the
		// words where somebody can see them.
		//
		// Paying for it only costs a genuine failure, because the loop
		// breaks the moment the anchor is found. The absent case has its own
		// budget below and is untouched, which is what keeps a full run from
		// getting slower.
		budget := 25 * time.Second
		if expect == "after" {
			// An anchor the slate says is not there yet can only be
			// confirmed by a budget expiring, so this is paid in full on
			// every such row: ten of them was 15s of a 47s test. Tunable
			// because the safe floor is a measurement, not a guess.
			budget = absent
		}
		probe := time.Now().Add(budget)
		for time.Now().Before(probe) {
			val, err := s.eval(sessionID, fmt.Sprintf(markJS, string(quoted)))
			if err == nil && val != nil && json.Unmarshal(val, &found) == nil && found.Found {
				break
			}
			time.Sleep(150 * time.Millisecond)
		}
		if !found.Found {
			// The tab stays open on purpose. The page is still the right
			// page, and a human can search it; closing it would leave them
			// with an error message and nothing to look at.
			//
			// And before they are left looking, try the structural
			// fallback: the exact text is gone or not written yet, but the
			// LINE it lives on is known independently of what it says. For
			// an anchor that has gone stale this is the difference between
			// the right line and an unmarked page; for one the slate says
			// is not there yet, it points at the line about to change.
			atLine := ""
			if locatorRE != "" {
				var hit struct {
					Found bool   `json:"found"`
					Line  int    `json:"line"`
					Text  string `json:"text"`
				}
				quotedRE, _ := json.Marshal(locatorRE)
				val, lerr := s.eval(sessionID,
					fmt.Sprintf(lineJS, string(quotedRE)))
				if lerr == nil && val != nil &&
					json.Unmarshal(val, &hit) == nil && hit.Found {
					atLine = fmt.Sprintf(" Showing line %d instead: %s",
						hit.Line, trim(hit.Text, 60))
				}
			}
			// The tab was brought to front before the navigation, so there
			// is nothing to do here. Kept as a comment rather than deleted
			// because this branch is the one where a reader will wonder:
			// the anchor was not found, and it is reasonable to ask whether
			// the human was shown the page at all. They were, from the
			// moment it started loading.
			if expect == "after" {
				// NOT an error in the sense that matters. The page is open
				// and the absence is the point: this is what she is being
				// shown before the change.
				return targetID, reused, fmt.Errorf(
					"not there yet, as expected: %q is what %s will "+
						"say AFTER the change.%s",
					trim(find, 50), trim(target, 50), atLine)
			}
			return targetID, reused, fmt.Errorf(
				"opened %s but could not find %q on it.%s",
				trim(target, 70), trim(find, 60), atLine)
		}
	}
	// NO bringToFront HERE ANY MORE. It was "the ONLY deliberate focus
	// change, and only once the page is ready", and being last was the
	// fault: the window arrived showing the previous tab and switched
	// seconds later. It is done before the navigation now, so by this point
	// the human has been watching the page load.
	//
	// Three calls where one will do, if this were left in: the early one,
	// this, and the not-found branch above. Each is a round trip, and two
	// of them are asking for something already true.
	return targetID, reused, nil
}

func trim(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
