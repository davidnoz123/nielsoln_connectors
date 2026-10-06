// sessions.go -- which Chromes on this machine can be driven, and what they hold.
//
// WHY THIS EXISTS. Everything that reaches a remote account through a browser
// needs the SAME thing first: a Chrome that is already signed in to that
// account, reachable over CDP. `remote_chrome_cdp` deliberately refuses to be
// that browser, and says so in launch.go: it uses a dedicated profile, always,
// which means the Chrome it launches is signed OUT. That is right for sweeping
// a client's public pages and useless for anything behind a login.
//
// So this connector takes the opposite stance: ATTACH, and never launch a
// fresh anonymous profile. It reports what is there rather than arranging it.
//
// ⚠️ LIVENESS AND AUTHORITY ARE DIFFERENT QUESTIONS, and only the first is
// answered here.
//
//	is Chrome up on this port          generic, and /json/version answers it
//	                                   over plain HTTP with no websocket
//	is it signed in, as which account  site-specific, needs a websocket and
//	                                   Runtime.evaluate, NOT in this increment
//
// The second matters because it is the failure that looks like success: a
// signed-out browser yields an empty ChatGPT catalog, which reads as "nothing
// to capture". Until it is implemented, this program must not be read as
// saying a session has authority. `Tabs` is a hint about what a browser has
// open, never a claim about who it is logged in as.
//
// ⚠️ AND NO CDP CLIENT LIVES HERE YET, on purpose. `remote_chrome_cdp/cdp.go`
// and ws.go are 42KB of working CDP client, and the authority probe needs
// them. Copying them would put two implementations of subtle machinery in one
// repo, which on 5 Oct 2026 cost three separate incidents in the one area
// where two tools both reasoned about Excel sessions. They get EXTRACTED, in
// their own increment, and this file stays HTTP-only until then.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"net"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Where Chrome's debugging HTTP endpoints live. Loopback only: a CDP port
// bound anywhere reachable hands the user's signed-in browser to the network.
const host = "127.0.0.1"

// Nothing below this is a CDP endpoint, and discovery used to probe the lot.
//
// ⚠️ MEASURED ON THE FIRST REAL RUN: -discover sent an HTTP GET to 135, 139,
// 443 and 445, which is RPC and SMB, and printed sixteen lines of refusals
// before the two answers. Talking HTTP at a file-sharing port is impolite as
// well as useless, and one of them answered by forcibly closing the
// connection, which is the service objecting.
const firstUnreservedPort = 1024

// worthProbing drops what discovery should never have asked.
//
// Deliberately NOT a filter on the owning process name, which is what this
// flag's help text used to claim. Reading a process name needs a second
// Windows API, and promising a filter that does not exist is worse than
// filtering by something honest and saying so.
func worthProbing(port int) bool { return port >= firstUnreservedPort }

// Tab is one target Chrome has open. A HINT about what a browser is for, and
// never evidence that it is signed in to anything: a page can be open and
// logged out, and /json/list cannot tell the difference.
type Tab struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Type  string `json:"type"`
}

// Session is one Chrome that may or may not be answering on a port.
type Session struct {
	Port int `json:"port"`
	// Alive means /json/version answered. It does NOT mean the browser is
	// signed in to anything, which is the distinction this whole file turns
	// on.
	Alive bool `json:"alive"`
	// OwnerPID is the process listening on the port, or 0 where the platform
	// cannot say. Useful because an answering port and an unexpected owner is
	// a different problem from a dead port.
	OwnerPID     int    `json:"owner_pid"`
	Browser      string `json:"browser,omitempty"`
	UserAgent    string `json:"user_agent,omitempty"`
	WebSocketURL string `json:"websocket_url,omitempty"`
	Tabs         []Tab  `json:"tabs,omitempty"`
	// Why it is not alive, in the words of whatever refused. Kept because
	// "connection refused" and "timed out" mean different things: the first is
	// nothing listening, the second is something listening and wedged.
	Error string `json:"error,omitempty"`
}

// version is the shape of /json/version, with the fields worth keeping.
type version struct {
	Browser              string `json:"Browser"`
	UserAgent            string `json:"User-Agent"`
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
}

// probe asks one port whether it is a CDP endpoint. Never raises.
//
// Two requests, and the second is allowed to fail on its own: a browser that
// answers /json/version is reachable whether or not it will list targets, and
// reporting it as dead because the tab list failed would be a lie about the
// thing that matters.
func probe(port int, timeout time.Duration) Session {
	out := Session{Port: port, OwnerPID: ownerOfPort(port)}
	client := &http.Client{Timeout: timeout}

	var v version
	if err := getJSON(client, port, "/json/version", &v); err != nil {
		out.Error = err.Error()
		return out
	}
	out.Alive = true
	out.Browser = v.Browser
	out.UserAgent = v.UserAgent
	out.WebSocketURL = v.WebSocketDebuggerURL

	var tabs []Tab
	if err := getJSON(client, port, "/json/list", &tabs); err != nil {
		// Reported on the session rather than thrown away, and it does not
		// clear Alive. See the note above.
		out.Error = "listed version but not targets: " + err.Error()
		return out
	}
	for _, t := range tabs {
		if t.Type == "page" {
			out.Tabs = append(out.Tabs, t)
		}
	}
	return out
}

func getJSON(client *http.Client, port int, path string, into any) error {
	url := fmt.Sprintf("http://%s:%d%s", host, port, path)
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s said %s", path, resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(into)
}

// parsePorts reads a comma-separated list, refusing anything that is not a
// usable port rather than silently dropping it.
//
// ⚠️ DECLARED, NEVER ALLOCATED, which is a settled rule for a different
// reason: a capability's identity has to be computable before its process
// runs, so "pick a free port" cannot be part of it. This is the declaring
// side of that rule.
func parsePorts(spec string) ([]int, error) {
	var out []int
	seen := map[int]bool{}
	for _, raw := range strings.Split(spec, ",") {
		got := strings.TrimSpace(raw)
		if got == "" {
			continue
		}
		n, err := strconv.Atoi(got)
		if err != nil {
			return nil, fmt.Errorf("%q is not a port number", got)
		}
		if n < 1 || n > 65535 {
			return nil, fmt.Errorf("%d is not a port in 1..65535", n)
		}
		if !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no ports given")
	}
	return out, nil
}

// listening reports whether anything holds the port, without speaking HTTP.
//
// Cheap, and it separates "nothing is there" from "something is there and will
// not answer", which are different problems with different fixes. A wedged
// browser is the second, and a report that called it dead would send somebody
// looking for the wrong thing.
func listening(port int, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp",
		net.JoinHostPort(host, strconv.Itoa(port)), timeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func render(sessions []Session, asJSON bool) string {
	if asJSON {
		raw, err := json.MarshalIndent(sessions, "", "  ")
		if err != nil {
			return fmt.Sprintf("{\"error\": %q}\n", err.Error())
		}
		return string(raw) + "\n"
	}
	var b strings.Builder
	for _, s := range sessions {
		state := "DEAD"
		if s.Alive {
			state = "alive"
		}
		owner := "owner unknown"
		if s.OwnerPID > 0 {
			owner = fmt.Sprintf("pid %d", s.OwnerPID)
		}
		fmt.Fprintf(&b, "  %-6d %-5s %-16s %s\n", s.Port, state, owner,
			s.Browser)
		if s.Error != "" {
			fmt.Fprintf(&b, "         %s\n", s.Error)
		}
		for _, t := range s.Tabs {
			// UNESCAPED FOR A HUMAN, and only here. The first real run showed
			// `Edit &quot;Home&quot; with Elementor`, which is what the page
			// reported. The JSON form keeps Chrome's own string, because a
			// caller parsing it should get the value rather than my idea of
			// the value.
			title := html.UnescapeString(t.Title)
			if len(title) > 48 {
				title = title[:48]
			}
			fmt.Fprintf(&b, "         %-50s %s\n", title, t.URL)
		}
	}
	if len(sessions) == 0 {
		b.WriteString("  nothing to report\n")
	}
	return b.String()
}

func main() {
	ports := flag.String("ports", "9222",
		"comma-separated CDP ports to probe. Declared, never allocated: see "+
			"the note on parsePorts")
	discover := flag.Bool("discover", false,
		"also probe every listening port at or above 1024, and report only "+
			"the ones that answer. A diagnostic for finding a browser on a "+
			"port somebody forgot, never a way to choose a port")
	asJSON := flag.Bool("json", false,
		"emit JSON, so a Python caller can use this before shimp exists")
	timeoutMS := flag.Int("timeout-ms", 1500,
		"per-request budget. A browser that cannot answer /json/version in "+
			"this long is reported as unreachable rather than waited for")
	flag.Parse()

	want, err := parsePorts(*ports)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  REFUSING: %v\n", err)
		os.Exit(2)
	}
	declared := map[int]bool{}
	for _, p := range want {
		declared[p] = true
	}
	tried := 0
	if *discover {
		found, why := discoverPorts()
		if why != "" {
			fmt.Fprintf(os.Stderr, "  discovery said: %s\n", why)
		}
		for _, p := range found {
			if !declared[p] && worthProbing(p) {
				want = append(want, p)
				tried++
			}
		}
		sort.Ints(want)
	}

	timeout := time.Duration(*timeoutMS) * time.Millisecond
	var out []Session
	alive := 0
	for _, p := range want {
		s := probe(p, timeout)
		if !s.Alive && !listening(p, timeout) {
			s.Error = "nothing is listening on this port"
		}
		if s.Alive {
			alive++
		}
		// A DECLARED PORT IS REPORTED EITHER WAY, because somebody asked
		// about it and "it is not there" is the answer. A DISCOVERED one is
		// reported only if it answered: the rest are services that were never
		// the question, and sixteen lines of them buried the two that were.
		if declared[p] || s.Alive {
			out = append(out, s)
		}
	}
	fmt.Print(render(out, *asJSON))
	if tried > 0 {
		fmt.Fprintf(os.Stderr, "  probed %d discovered port(s) besides the "+
			"declared ones; only those that answered are listed above\n", tried)
	}

	// ⚠️ EXIT 0 EVEN WITH NOTHING ALIVE. This is a report, and a caller asking
	// "what is there" gets an answer rather than a failure. A caller that
	// NEEDS a session should read the output and refuse for itself, loudly,
	// because the refusal belongs where the requirement is.
	if alive == 0 {
		fmt.Fprintf(os.Stderr,
			"  no CDP session answered on %v. Nothing here launches one: "+
				"this connector attaches, and a signed-in browser has to be "+
				"one you started yourself.\n", want)
	}
}
