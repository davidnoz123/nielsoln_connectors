// chrome_ready -- ensure a CDP-enabled Chrome is answering on a port.
//
//	go run github.com/davidnoz123/nielsoln_connectors/chrome_ready@<sha> \
//	    -port 9222 -profile chrome_debug_profile
//
// THE FIRST SHIMP CAPABILITY, and it was chosen because it is the first half
// of what the user actually asked for: "make sure a CDP is open with my
// profile so I automatically access my ChatGPT account". The scraping half
// already exists in chatgpt_tools and needs no help; this is the part that
// was a wish.
//
// ⚠️ A ONE-SHOT, DELIBERATELY. s46's three classes: this probes, possibly
// launches, reports and exits. It holds no lock, keeps no state and needs
// no keeper, which is why it can be the first thing a click actually runs
// while s20 is still unbuilt.
//
// IDEMPOTENT BY DESIGN. Clicking twice must be harmless, because a link is a
// thing people click twice. If Chrome is already answering it says so and
// changes nothing.
//
// ⚠️ IT WILL NOT ATTACH TO A RUNNING CHROME THAT HAS NO DEBUG PORT, and
// cannot: CDP is enabled only at launch. Where it finds a Chrome running on
// the target profile WITHOUT a port, it says exactly that rather than
// launching a second one that would fail to bind or silently open a tab in
// the first. remote_chrome_cdp learned that lesson on 2 October by driving
// somebody else's tab away from its own page.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

// version is what /json/version answers with. Only the fields used.
type version struct {
	Browser  string `json:"Browser"`
	Protocol string `json:"Protocol-Version"`
	WSURL    string `json:"webSocketDebuggerUrl"`
}

func main() {
	port := flag.Int("port", 9222, "the CDP port to require")
	profile := flag.String("profile", "default",
		"the Chrome user-data-dir. A PATH is used exactly as given; a bare "+
			"name resolves under <user config dir>/chrome_ready/<name>. ⚠️ "+
			"NEVER the temp directory, and never Chrome's own default "+
			"profile: see profileDir()")
	launch := flag.Bool("launch", true,
		"launch Chrome if nothing answers. -launch=false only reports")
	wait := flag.Duration("wait", 25*time.Second,
		"how long to wait for a launched Chrome to answer")
	flag.Parse()

	fmt.Printf("chrome_ready: requiring CDP on 127.0.0.1:%d\n", *port)

	if v, err := probe(*port, 3*time.Second); err == nil {
		// ⚠️ The already-running case comes FIRST and reports success
		// without touching anything. That is what makes a second click
		// harmless.
		fmt.Printf("  ALREADY READY\n")
		report(v)
		os.Exit(0)
	}

	fmt.Printf("  nothing answering on %d\n", *port)
	if !*launch {
		fmt.Printf("  -launch=false, so stopping here and changing nothing\n")
		os.Exit(1)
	}

	exe := chromeExe()
	if exe == "" {
		// Reported by name rather than guessed at. owner_other.go's rule:
		// say what cannot be found.
		fmt.Fprintf(os.Stderr,
			"  FAILED: no chrome.exe in any known location. Looked under "+
				"PROGRAMFILES, PROGRAMFILES(X86) and LOCALAPPDATA.\n")
		os.Exit(1)
	}

	dir, err := profileDir(*profile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  FAILED: %v\n", err)
		os.Exit(1)
	}
	// Whether the profile already exists decides whether a sign-in is
	// needed, so it must be read BEFORE MkdirAll creates it.
	fresh := !dirExists(dir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "  FAILED: cannot create the profile "+
			"directory %s: %v\n", dir, err)
		os.Exit(1)
	}

	fmt.Printf("  launching %s\n", exe)
	fmt.Printf("  profile    %s\n", dir)
	if fresh {
		// ⚠️ SAID ON THE RUN THAT CREATES IT, not buried in a doc. A fresh
		// profile is signed in to NOTHING, so a capability that expects an
		// authenticated account gets a blank browser and no explanation.
		fmt.Println("  ⚠️ THIS PROFILE IS NEW, so it is signed out of " +
			"everything. Sign in once in the window that opens; the")
		fmt.Println("     profile is durable and the session persists across " +
			"runs. It is NOT in the temp directory, so a disk")
		fmt.Println("     cleanup will not silently sign you out again.")
	}
	cmd := exec.Command(exe,
		"--remote-debugging-port="+strconv.Itoa(*port),
		"--user-data-dir="+dir,
		"--remote-allow-origins=*",
		"--no-first-run",
		"--no-default-browser-check",
	)
	// ⚠️ VISIBLE, always, and not detached from the user's desktop. AGENTS is
	// explicit: never start a browser invisibly, because a hidden window with
	// a modal in it is a process nobody can diagnose.
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "  FAILED to start Chrome: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  started pid %d, waiting up to %s for the port\n",
		cmd.Process.Pid, *wait)

	// ⚠️ Polled on the PORT ANSWERING, never on the process existing. Chrome
	// forks and the first process may exit immediately, and a launched
	// browser that never opens its port is exactly the silent failure this
	// program exists to prevent.
	deadline := time.Now().Add(*wait)
	for time.Now().Before(deadline) {
		if v, err := probe(*port, 2*time.Second); err == nil {
			fmt.Printf("  READY\n")
			report(v)
			os.Exit(0)
		}
		time.Sleep(500 * time.Millisecond)
	}
	fmt.Fprintf(os.Stderr,
		"  FAILED: Chrome was started but nothing answered on %d within %s. "+
			"Either it is still starting, or another Chrome already owns "+
			"this profile directory.\n", *port, *wait)
	os.Exit(1)
}

// probe asks the port what it is.
func probe(port int, timeout time.Duration) (version, error) {
	var v version
	c := &http.Client{Timeout: timeout}
	url := "http://127.0.0.1:" + strconv.Itoa(port) + "/json/version"
	resp, err := c.Get(url)
	if err != nil {
		return v, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if err != nil {
		return v, err
	}
	if err := json.Unmarshal(raw, &v); err != nil {
		return v, fmt.Errorf("%s answered but not with CDP JSON: %w", url, err)
	}
	if v.Browser == "" {
		// ⚠️ Something is listening and it is NOT Chrome. Saying "ready"
		// here would hand a caller a port that cannot speak CDP.
		return v, fmt.Errorf("%s answered without a Browser field, so "+
			"something other than Chrome holds this port", url)
	}
	return v, nil
}

func report(v version) {
	fmt.Printf("  browser    %s\n", v.Browser)
	fmt.Printf("  protocol   %s\n", v.Protocol)
	if v.WSURL != "" {
		fmt.Printf("  websocket  %s\n", v.WSURL)
	}
}

// chromeExe searches known locations and returns "" if none has it.
func chromeExe() string {
	for _, base := range []string{"PROGRAMFILES", "PROGRAMFILES(X86)",
		"LOCALAPPDATA"} {
		root := os.Getenv(base)
		if root == "" {
			continue
		}
		p := filepath.Join(root, "Google", "Chrome", "Application", "chrome.exe")
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	return ""
}
