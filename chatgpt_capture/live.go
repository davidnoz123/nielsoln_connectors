// live.go -- `-port`: capture from a live, signed-in ChatGPT session over CDP.
//
// # WHAT IT DOES
//
// Drives an already-authenticated browser (the one `chrome_ready` ensures),
// asks ChatGPT's own API which conversations have changed, fetches the changed
// ones, and writes them with the SAME output path the zip half uses. The API
// returns the same conversation shape the export does, which is not an
// assumption: chatgpt_tools feeds both into one `upsert_conversation`,
// differing only in `source="network"` against `source="zip"`.
//
// # ⚠️ THE ACCESS TOKEN NEVER ENTERS THIS PROCESS
//
// `/backend-api/*` needs `Authorization: Bearer <accessToken>`, and that token
// is a bearer credential: the response carrying it says, in its own payload,
// "DO NOT SHARE ANY PART OF THE INFORMATION YOU SEE HERE ... SHARING THIS
// INFORMATION IS LIKE SHARING YOUR PASSWORD".
//
// So the token is fetched AND USED entirely inside the page. Every expression
// below reads the session, builds the header, makes the call and returns only
// the conversation JSON. The token is never returned across the websocket,
// never held in a Go variable and never written anywhere.
//
// ⚠️ THIS IS A FIX FOR A MISTAKE MADE EARLIER THE SAME DAY. A probe written to
// answer "is this browser signed in" parsed the CDP reply one level too deep,
// fell through to a diagnostic that printed the raw response, and put a token
// prefix into a transcript. The lesson was not "be careful with the output":
// it was to reduce inside the page so there is nothing to leak.
//
// # WHY A DIRECT FETCH AND NOT PASSIVE INTERCEPTION
//
// chatgpt_tools captures by scrolling the sidebar and harvesting ChatGPT's own
// XHRs out of a network log, which is why it needs DOM scrolling, an
// IntersectionObserver nudge and a plateau detector. ⓘ A direct fetch from the
// page needs none of that, and the same file already does exactly this in
// op_fetch_turn_counts -- `Runtime.evaluate` with `await fetch(...)`. So the
// simpler route is the established one here, not a new idea.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// cdpTarget is one entry from /json/list.
type cdpTarget struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	URL   string `json:"url"`
	Title string `json:"title"`
	WSURL string `json:"webSocketDebuggerUrl"`
}

// chatgptPage finds a page target already on chatgpt.com.
//
// ⚠️ IT DOES NOT NAVIGATE OR OPEN A TAB. A capture that moved the user's tabs
// would be doing something nobody asked for, and a relative fetch only works
// from an origin that is already chatgpt.com. If no such tab exists, that is
// reported as a REASON rather than fixed silently.
func chatgptPage(port int) (cdpTarget, error) {
	var out cdpTarget
	c := &http.Client{Timeout: 10 * time.Second}
	resp, err := c.Get("http://127.0.0.1:" + strconv.Itoa(port) + "/json/list")
	if err != nil {
		return out, fmt.Errorf("no CDP on 127.0.0.1:%d. Run chrome_ready "+
			"first: %w", port, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<22))
	if err != nil {
		return out, err
	}
	var targets []cdpTarget
	if err := json.Unmarshal(raw, &targets); err != nil {
		return out, fmt.Errorf("%d answered but not with a CDP target list: %w",
			port, err)
	}
	for _, t := range targets {
		if t.Type == "page" && t.WSURL != "" &&
			(strings.Contains(t.URL, "chatgpt.com") ||
				strings.Contains(t.URL, "chat.openai.com")) {
			return t, nil
		}
	}
	return out, fmt.Errorf(
		"no chatgpt.com tab is open on port %d, and this program will not "+
			"open one: a relative fetch only works from that origin, and "+
			"moving somebody's tabs is not this program's business. Open "+
			"https://chatgpt.com/ in that browser and run this again",
		port)
}

// session is a CDP connection to one page, with an id counter.
type session struct {
	ws *wsConn
	id int
}

func dialPage(t cdpTarget) (*session, error) {
	ws, err := wsDial(t.WSURL, 20*time.Second)
	if err != nil {
		return nil, err
	}
	return &session{ws: ws}, nil
}

func (s *session) Close() { _ = s.ws.Close() }

// eval runs an async expression in the page and returns its string result.
//
// ⚠️ EVERY CALLER RETURNS A STRING IT BUILT INSIDE THE PAGE. The reply is
// never dumped on an unexpected shape: a miss returns an error naming the
// keys, because the one time a diagnostic printed a raw CDP reply it printed
// a credential with it.
func (s *session) eval(expr string, timeout time.Duration) (string, error) {
	s.id++
	req, err := json.Marshal(map[string]any{
		"id":     s.id,
		"method": "Runtime.evaluate",
		"params": map[string]any{
			"expression":    expr,
			"awaitPromise":  true,
			"returnByValue": true,
			"timeout":       timeout.Milliseconds(),
		},
	})
	if err != nil {
		return "", err
	}
	if err := s.ws.writeText(req); err != nil {
		return "", err
	}
	deadline := time.Now().Add(timeout + 10*time.Second)
	for {
		raw, err := s.ws.readText(deadline)
		if err != nil {
			return "", err
		}
		var msg struct {
			ID     int `json:"id"`
			Result struct {
				Result struct {
					Value string `json:"value"`
				} `json:"result"`
				Exception *struct {
					Text string `json:"text"`
				} `json:"exceptionDetails"`
			} `json:"result"`
			Error *struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(raw, &msg); err != nil {
			return "", err
		}
		if msg.ID != s.id {
			continue // an event, or a reply to something else
		}
		if msg.Error != nil {
			return "", fmt.Errorf("CDP refused the evaluate: %s", msg.Error.Message)
		}
		if msg.Result.Exception != nil {
			return "", fmt.Errorf("the page threw: %s", msg.Result.Exception.Text)
		}
		if msg.Result.Result.Value == "" {
			return "", fmt.Errorf(
				"the evaluate returned no string value. ⚠️ The reply is NOT " +
					"printed here: an earlier diagnostic that dumped a raw " +
					"CDP reply printed an access token with it")
		}
		return msg.Result.Result.Value, nil
	}
}

// jsBearerFetch is the expression template every call uses.
//
// ⚠️ THE TOKEN IS READ AND SPENT IN THE SAME EXPRESSION. It is never part of
// the returned string. %s is the path to fetch.
const jsBearerFetch = `(async () => {
  const s = await (await fetch('/api/auth/session', {credentials:'include'})).json();
  if (!s || !s.accessToken) return JSON.stringify({error:'not_signed_in'});
  const r = await fetch(%q, {headers:{Authorization:'Bearer ' + s.accessToken}});
  if (!r.ok) return JSON.stringify({error:'http_' + r.status});
  return JSON.stringify(await r.json());
})()`

// apiError is how the page reports a failure without leaking anything.
type apiError struct {
	Error string `json:"error"`
}

func (s *session) apiGet(path string, timeout time.Duration) (json.RawMessage, error) {
	out, err := s.eval(fmt.Sprintf(jsBearerFetch, path), timeout)
	if err != nil {
		return nil, err
	}
	var e apiError
	if json.Unmarshal([]byte(out), &e) == nil && e.Error != "" {
		switch e.Error {
		case "not_signed_in":
			return nil, fmt.Errorf(
				"the browser on that port is NOT SIGNED IN to ChatGPT. " +
					"/api/auth/session returned no accessToken. Sign in in " +
					"that window once; the profile is durable")
		default:
			return nil, fmt.Errorf("ChatGPT's API answered %s for %s",
				strings.TrimPrefix(e.Error, "http_"), path)
		}
	}
	return json.RawMessage(out), nil
}

// convSummary is the only part of the list this program interprets.
//
// ⚠️ TWO FIELDS, AND NOTHING ELSE. The design property of this module is that
// it does not interpret conversations; the list is the one exception, because
// deciding WHICH to fetch requires reading an id and a time. The conversation
// bodies still pass through untouched.
type convSummary struct {
	ID         string  `json:"id"`
	UpdateTime float64 `json:"update_time"`
}

// listChanged pages the conversation list and returns those newer than since.
//
// ⚠️ IT STOPS AT THE FIRST PAGE WHOSE ENTRIES ARE ALL OLDER, because the list
// is ordered by update_time descending. That is what makes this INCREMENTAL
// rather than a full sweep every run, and it is why `-since` matters.
func (s *session) listChanged(since float64, pageSize, max int) ([]convSummary, error) {
	var out []convSummary
	for offset := 0; ; offset += pageSize {
		path := fmt.Sprintf(
			"/backend-api/conversations?offset=%d&limit=%d&order=updated",
			offset, pageSize)
		raw, err := s.apiGet(path, 30*time.Second)
		if err != nil {
			return out, err
		}
		var page struct {
			Items []convSummary `json:"items"`
			Total int           `json:"total"`
		}
		if err := json.Unmarshal(raw, &page); err != nil {
			return out, fmt.Errorf("the conversation list did not parse: %w", err)
		}
		if len(page.Items) == 0 {
			return out, nil
		}
		stop := false
		for _, it := range page.Items {
			if it.UpdateTime > since {
				out = append(out, it)
				if max > 0 && len(out) >= max {
					return out, nil
				}
			} else {
				// Ordered descending, so everything after this is older too.
				stop = true
			}
		}
		if stop {
			return out, nil
		}
	}
}

// runLive captures changed conversations from a signed-in tab.
//
// Returns how many were written and how many the list said had changed. The
// two differ when -max cuts in, which main() then reports.
// runLive captures into a Store. `since` of -1 means "ask the store".
//
// ⚠️ THE STORE OWNS THE WATERMARK, which is the point of the seam. Before
// this, -since had to be supplied by the caller, so being incremental meant
// somebody else remembering a number. A store that knows what it already has
// makes the capture incremental without any external bookkeeping.
func runLive(port int, since float64, st Store, max int) (int, int, error) {
	if since < 0 {
		var how string
		since, how = st.Watermark()
		fmt.Fprintf(os.Stderr, "  since %.0f -- %s\n", since, how)
	}

	t, err := chatgptPage(port)
	if err != nil {
		return 0, 0, err
	}
	fmt.Fprintf(os.Stderr, "  using the tab at %s\n", t.URL)

	s, err := dialPage(t)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot open a CDP websocket to that tab: %w", err)
	}
	defer s.Close()

	changed, err := s.listChanged(since, listPageSize, max)
	if err != nil {
		return 0, 0, err
	}
	fmt.Fprintf(os.Stderr, "  %d conversation(s) newer than %.0f\n",
		len(changed), since)

	n := 0
	for i, c := range changed {
		if max > 0 && n >= max {
			break
		}
		raw, err := s.apiGet("/backend-api/conversation/"+c.ID, 60*time.Second)
		if err != nil {
			// ⚠️ PARTIAL RESULTS ARE KEPT AND COMMITTED. Losing 200 fetched
			// conversations because the 201st failed would make a transient
			// error cost a whole run, and the watermark then covers only what
			// actually landed.
			kept, _ := st.Commit()
			return kept, len(changed), fmt.Errorf(
				"after %d of %d, conversation %s: %w", n, len(changed), c.ID, err)
		}
		// ⚠️ HANDED TO THE STORE UNINTERPRETED. The store compacts and
		// writes; nothing here reads a field of the conversation.
		if err := st.Put(raw); err != nil {
			kept, _ := st.Commit()
			return kept, len(changed), fmt.Errorf("conversation %s: %w", c.ID, err)
		}
		// The ONE field the watermark needs, and it comes from the LIST that
		// was already parsed for exactly two fields -- not from the body.
		if o, ok := st.(interface{ Observe(float64) }); ok {
			o.Observe(c.UpdateTime)
		}
		n++
		if (i+1)%25 == 0 {
			// Progress, because a few hundred fetches is minutes and silence
			// reads as a hang.
			fmt.Fprintf(os.Stderr, "  %d/%d\n", n, len(changed))
		}
		time.Sleep(fetchDelay)
	}
	kept, err := st.Commit()
	return kept, len(changed), err
}

// listPageSize is what the API is asked for per list call.
const listPageSize = 100

// fetchDelay spaces the per-conversation calls.
//
// ⚠️ NOT TUNING, A MEASURED CONSTRAINT FROM THE PYTHON: its fetch_turn_counts
// uses 1.0s between /backend-api/conversation/<id> calls and records that 0.5
// TRIGGERS RATE LIMITS. Going faster here would rediscover that against the
// user's own account.
const fetchDelay = time.Second
