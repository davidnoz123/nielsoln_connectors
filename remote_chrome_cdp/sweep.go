// sweep.go -- check every anchor, one page load per page, lanes in parallel.
//
// # WHY A SECOND PATH AT ALL
//
// /show/ exists for a human clicking a cell. It reloads on every click, on
// purpose: the reason somebody clicks the same row twice is usually to see
// whether the change landed, and a cached DOM would show the old value with
// a fresh highlight on it -- the one failure that looks exactly like
// success.
//
// A sweep is the opposite situation. Nobody is watching, nothing has changed
// between the first anchor and the twelfth, and 27 anchors spread over 10
// pages were costing 27 page loads. Loading each page once and running every
// anchor that lives on it is worth roughly a 2.5x cut, and it is only safe
// BECAUSE no human is in the loop.
//
// # WHY MULTIPLE ANCHORS PER LOAD IS SAFE
//
// markJS concatenates every text node before searching, so the <mark> it
// inserts splits a text node without changing the string being searched.
// Anchor twelve therefore sees the same text anchor one did. Marks are
// still cleared between anchors, so that a tab left open afterwards shows
// the LAST thing asked about rather than a dozen overlapping highlights
// nobody can attribute.
//
// # THE LANES ARE BOUNDED, AND EACH OWNS ITS CONNECTION
//
// Page loads are network-bound and mostly idle, so running several fills a
// pipe one cannot. They are bounded rather than unlimited because the far
// end is one WordPress site: twenty parallel view-source fetches is a small
// denial of service against the customer we are about to go and help.
//
// Each lane dials its own CDP connection. One shared websocket would
// interleave replies from concurrent commands and hand a lane somebody
// else's answer.
//
// # IT NEVER TAKES FOCUS
//
// A sweep raising ten windows in four seconds makes the machine unusable.
// /show/ raises deliberately; this never does.
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// clearMarksJS undoes markJS: unwraps the <mark> elements and drops the
// inline styling used when a match crossed element boundaries.
//
// normalize() rejoins the text nodes that unwrapping leaves behind. Without
// it a page swept many times accumulates thousands of adjacent fragments
// and the tree walk gets measurably slower each pass.
const clearMarksJS = `(function () {
  var n = 0;
  document.querySelectorAll('mark').forEach(function (m) {
    var parent = m.parentNode;
    if (!parent) return;
    while (m.firstChild) parent.insertBefore(m.firstChild, m);
    parent.removeChild(m);
    parent.normalize();
    n++;
  });
  Array.from(document.querySelectorAll('[style]')).forEach(function (e) {
    if (e.style && e.style.background === 'rgb(255, 225, 77)') {
      e.style.background = '';
      e.style.outline = '';
      n++;
    }
  });
  return {cleared: n};
})()`

// sweepResult is one anchor's verdict. Deliberately NOT a pass or a fail:
// whether "absent" is correct depends on the slate's expectation, and that
// judgement belongs to the caller that holds the slate.
type sweepResult struct {
	ID     string `json:"id"`
	Page   string `json:"page"`
	Anchor string `json:"anchor"`
	Expect string `json:"expect"`
	Found  bool   `json:"found"`
	Spans  int    `json:"spans"`
	MS     int64  `json:"ms"`
	Note   string `json:"note"`
}

type sweepJob struct {
	id   string
	task Task
}

// groupByPage returns the page keys in first-seen order plus their jobs.
//
// Order is kept so that a report reads in the same sequence as the slate,
// which is how anybody reading it expects to scan.
func (s *server) groupByPage(ids []string) ([]string, map[string][]sweepJob) {
	order := []string{}
	groups := map[string][]sweepJob{}
	for _, id := range ids {
		task, ok := s.manifest.Lookup(id)
		if !ok {
			continue
		}
		key := task.Page
		if task.Source {
			key = "view-source:" + task.Page
		}
		if _, seen := groups[key]; !seen {
			order = append(order, key)
		}
		groups[key] = append(groups[key], sweepJob{id: id, task: task})
	}
	return order, groups
}

// sweepPage loads one page once and checks every anchor on it.
func (s *server) sweepPage(key string, jobs []sweepJob) []sweepResult {
	out := make([]sweepResult, 0, len(jobs))
	fail := func(note string) []sweepResult {
		for _, j := range jobs {
			out = append(out, sweepResult{ID: j.id, Page: key,
				Anchor: j.task.Find, Expect: j.task.Expect, Note: note})
		}
		return out
	}

	sess, err := dialCDP(s.cdpPort)
	if err != nil {
		return fail("no chrome: " + err.Error())
	}
	defer sess.Close()

	// Same reclaim rule as a click: a tab left by an earlier run is closed
	// rather than inherited, so the count stays flat and the page is clean.
	sess.closeTabsAt(key)
	targetID, sessionID, err := sess.openTab()
	if err != nil {
		return fail("could not open a tab: " + err.Error())
	}
	loaded := time.Now()
	sess.send(sessionID, "Page.enable", nil)
	if _, err := sess.send(sessionID, "Page.navigate",
		map[string]any{"url": key}); err != nil {
		sess.closeTab(targetID)
		return fail("could not navigate: " + err.Error())
	}
	sess.waitEvent("Page.loadEventFired", 20*time.Second)
	loadMS := time.Since(loaded).Milliseconds()

	for i, j := range jobs {
		started := time.Now()
		if i > 0 {
			// Only between anchors: the first one inherits a fresh page.
			sess.eval(sessionID, clearMarksJS)
		}
		res := sweepResult{ID: j.id, Page: key, Anchor: j.task.Find,
			Expect: strings.ToLower(j.task.Expect)}
		if j.task.Find == "" {
			res.Note = "no anchor to look for"
			res.MS = time.Since(started).Milliseconds()
			out = append(out, res)
			continue
		}
		quoted, qerr := json.Marshal(j.task.Find)
		if qerr != nil {
			res.Note = "unquotable anchor: " + qerr.Error()
			out = append(out, res)
			continue
		}
		// Same budgets as a click, for the same reason: an anchor the slate
		// says is not there yet should cost a glance, not a wait.
		budget := 8 * time.Second
		if res.Expect == "after" {
			budget = s.absentBudget
		}
		var got struct {
			Found bool `json:"found"`
			Spans int  `json:"spans"`
		}
		probe := time.Now().Add(budget)
		for time.Now().Before(probe) {
			val, evErr := sess.eval(sessionID,
				fmt.Sprintf(markJS, string(quoted)))
			if evErr == nil && val != nil &&
				json.Unmarshal(val, &got) == nil && got.Found {
				break
			}
			time.Sleep(150 * time.Millisecond)
		}
		res.Found, res.Spans = got.Found, got.Spans
		res.MS = time.Since(started).Milliseconds()
		if i == 0 {
			res.Note = fmt.Sprintf("page loaded in %dms", loadMS)
		}
		out = append(out, res)
	}
	return out
}

// sweep checks every id, grouped by page, `lanes` pages at a time.
func (s *server) sweep(ids []string, lanes int) []sweepResult {
	// Excludes clicks for the duration: a human raising a window mid-sweep
	// would have it yanked away by the next lane.
	s.mu.Lock()
	defer s.mu.Unlock()

	order, groups := s.groupByPage(ids)
	if lanes < 1 {
		lanes = 1
	}
	if lanes > len(order) {
		lanes = len(order)
	}

	var mu sync.Mutex
	collected := map[string][]sweepResult{}
	sem := make(chan struct{}, lanes)
	var wg sync.WaitGroup
	for _, key := range order {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			res := s.sweepPage(key, groups[key])
			mu.Lock()
			collected[key] = res
			mu.Unlock()
		}(key)
	}
	wg.Wait()

	// Back into the order the caller asked in, not the order they finished.
	rank := map[string]int{}
	for i, id := range ids {
		rank[id] = i
	}
	flat := []sweepResult{}
	for _, key := range order {
		flat = append(flat, collected[key]...)
	}
	sort.SliceStable(flat, func(a, b int) bool {
		return rank[flat[a].ID] < rank[flat[b].ID]
	})
	return flat
}

// sweepHandler answers POST/GET /sweep?t=token&ids=a,b,c&lanes=4
//
// `ids` empty means every task in the manifest, which is the usual call.
func (s *server) sweepHandler(w http.ResponseWriter, r *http.Request) {
	if !s.guard(w, r) {
		return
	}
	ids := []string{}
	for _, raw := range strings.Split(r.URL.Query().Get("ids"), ",") {
		if id := strings.TrimSpace(raw); id != "" {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		ids = s.manifest.IDs()
	}
	lanes := s.sweepLanes
	if v := r.URL.Query().Get("lanes"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			lanes = n
		}
	}
	started := time.Now()
	results := s.sweep(ids, lanes)
	pages := map[string]bool{}
	for _, res := range results {
		pages[res.Page] = true
	}
	elapsed := time.Since(started)
	s.log("sweep", fmt.Sprintf("%d anchor(s) over %d page(s), %d lane(s), %s",
		len(results), len(pages), lanes, took(started)))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"anchors": len(results),
		"pages":   len(pages),
		"lanes":   lanes,
		"ms":      elapsed.Milliseconds(),
		"results": results,
	})
}
