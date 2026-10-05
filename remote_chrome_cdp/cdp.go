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
  var w = document.createTreeWalker(document.body, NodeFilter.SHOW_TEXT);
  var nodes = [], text = "", n;
  while ((n = w.nextNode())) {
    if (!n.nodeValue) continue;
    nodes.push({node: n, at: text.length, len: n.nodeValue.length});
    text += n.nodeValue;
  }
  var i = text.indexOf(needle);
  if (i < 0) return {found: false};
  function locate(offset) {
    for (var k = 0; k < nodes.length; k++) {
      if (offset < nodes[k].at + nodes[k].len) {
        return {node: nodes[k].node, off: offset - nodes[k].at};
      }
    }
    var last = nodes[nodes.length - 1];
    return {node: last.node, off: last.len};
  }
  var a = locate(i), b = locate(i + needle.length - 1);
  var rg = document.createRange();
  rg.setStart(a.node, a.off);
  rg.setEnd(b.node, b.off + 1);
  var marked = false;
  try {
    var mark = document.createElement("mark");
    mark.style.background = "#ffe14d";
    mark.style.outline = "3px solid #d08700";
    rg.surroundContents(mark);
    mark.scrollIntoView({block: "center", inline: "center"});
    marked = true;
  } catch (e) {
    // surroundContents refuses a range that crosses element boundaries,
    // which is exactly the attribute case. So EVERY node the match covers
    // is highlighted, not just the first: styling only the start node marks
    // the word "property" and leaves the value it is pointing at unmarked,
    // which reads as the tool finding the wrong thing.
    var from = i, to = i + needle.length;
    for (var k = 0; k < nodes.length; k++) {
      var s0 = nodes[k].at, s1 = s0 + nodes[k].len;
      if (s1 <= from || s0 >= to) continue;
      var host = nodes[k].node.parentElement;
      if (!host) continue;
      host.style.background = "#ffe14d";
      host.style.outline = "2px solid #d08700";
      if (!marked) {
        host.scrollIntoView({block: "center", inline: "center"});
      }
      marked = true;
    }
  }
  return {found: marked, spans: a.node === b.node ? 1 : 2};
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
	if _, err := s.send(sessionID, "Page.navigate", map[string]any{"url": target}); err != nil {
		if !reused {
			s.closeTab(targetID)
		}
		return "", reused, err
	}
	s.waitEvent("Page.loadEventFired", 12*time.Second)

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
		budget := 8 * time.Second
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
			if focus {
				s.send(sessionID, "Page.bringToFront", nil)
			}
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
	// The ONLY deliberate focus change, and only once the page is ready.
	// -focus=false is for driving this while working on something else.
	//
	// This is now purely about what the human looks at. Making the page
	// RENDER is done above, unconditionally, because the two were one call
	// and a quiet drive silently got neither.
	if focus {
		s.send(sessionID, "Page.bringToFront", nil)
	}
	return targetID, reused, nil
}

func trim(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
