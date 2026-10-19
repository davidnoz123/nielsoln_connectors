// verify.go -- R2: the port answering must be the PROFILE that was asked for.
//
// # THE DEFECT, MEASURED
//
// ⚠️ This program used to probe the PORT and never consult `-profile`. Run as
// `-port 9333 -profile chatgpt` it reported ALREADY READY against
// `C:\Temp\chrome_checks_profile`, a completely different browser. On the
// machine where that was found, port 9222 held live dataannotation, ClickUp
// and Slack sessions, so "attaches to whatever is already there" is not a
// hypothetical: a link could drive the recipient's working browser.
//
// s73 records this as the first of four deployment surprises, and R2 as the
// rule that prevents it.
//
// # WHY A MARKER FILE AND NOT THE PROCESS TABLE
//
// The obvious implementation is port -> owning PID -> command line ->
// --user-data-dir. Two routes, both rejected:
//
//   - CDP cannot answer it. `Browser.getBrowserCommandLine` returns
//     "Command line not returned because --enable-automation not set", and
//     ⚠️ passing --enable-automation makes the browser announce itself as
//     automated through navigator.webdriver and the infobar, which a site
//     like ChatGPT may act on. Wrong trade for a verification step.
//   - The TCP table route works and `chrome_sessions/owner_windows.go`
//     already does it, but this module cannot import it: the repo has no leaf
//     modules yet (s13). Copying it would duplicate ~100 lines of syscall
//     structs, and AGENTS is explicit that the struct layouts are the
//     dangerous part -- dwLocalPort is network byte order in the low 16 bits,
//     and reading it as a plain uint32 turns 9222 into 30987, which looks
//     like a real port and so does not announce itself as a bug.
//
// So verification is keyed on something only one browser can produce: the
// BROWSER GUID in its own webSocketDebuggerUrl. Measured distinct across
// three live instances. chrome_ready writes the GUID into the profile it
// launched, and a later run matches the live GUID against that file.
//
// ⚠️ WHAT THIS CANNOT DO, STATED RATHER THAN PAPERED OVER: it cannot verify a
// browser chrome_ready did not start, because there is no marker to compare.
// That case is REFUSED rather than assumed good, which is the safe direction
// but means a hand-started browser needs -accept-unverified.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// markerName is the file chrome_ready writes into a profile it launched.
//
// Inside the profile directory on purpose: a marker elsewhere could outlive
// the profile, or describe one that had been deleted and recreated.
const markerName = ".chrome_ready.json"

// marker is what we record about a browser we started.
type marker struct {
	Port      int    `json:"port"`
	BrowserID string `json:"browser_id"`
	PID       int    `json:"pid"`
	// ⚠️ NO TIMESTAMP, deliberately. A time would invite an expiry rule, and
	// an expiry rule would make a long-running browser fail verification for
	// no reason. The GUID is the whole proof: it changes when the browser
	// restarts, which is exactly when the marker should stop matching.
}

// browserID extracts the unique browser GUID from a webSocketDebuggerUrl.
//
// The URL is ws://127.0.0.1:<port>/devtools/browser/<guid>. Measured distinct
// across three live instances on one machine.
func browserID(wsURL string) string {
	const sep = "/devtools/browser/"
	if i := strings.Index(wsURL, sep); i >= 0 {
		return wsURL[i+len(sep):]
	}
	return ""
}

// writeMarker records the browser we just launched, inside its own profile.
//
// A failure here is reported by the caller and is NOT fatal: the browser is
// up and usable, and the only loss is that the next run cannot verify it.
func writeMarker(profileDir string, m marker) error {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(profileDir, markerName), append(b, '\n'),
		0o644)
}

func readMarker(profileDir string) (marker, error) {
	var m marker
	b, err := os.ReadFile(filepath.Join(profileDir, markerName))
	if err != nil {
		return m, err
	}
	if err := json.Unmarshal(b, &m); err != nil {
		return m, fmt.Errorf("%s is not readable as a marker: %w",
			markerName, err)
	}
	return m, nil
}

// Verdict is the answer to "is the browser on this port the one asked for".
type Verdict int

const (
	// Verified: the live browser's GUID matches the marker in the requested
	// profile. This is the only case that may proceed silently.
	Verified Verdict = iota
	// NoMarker: something answers, but this profile has no marker, so the
	// browser was started by something other than chrome_ready.
	NoMarker
	// Mismatch: this profile HAS a marker and the live GUID is different, so
	// the port belongs to a different browser than the one asked for.
	Mismatch
)

// verify compares the live browser against the marker in the requested profile.
func verify(profileDir, liveWS string, livePort int) (Verdict, marker) {
	m, err := readMarker(profileDir)
	if err != nil {
		return NoMarker, marker{}
	}
	live := browserID(liveWS)
	if live != "" && live == m.BrowserID && m.Port == livePort {
		return Verified, m
	}
	return Mismatch, m
}

// explain renders a refusal a reader can act on.
//
// ⚠️ IT NAMES BOTH SIDES. "Wrong profile" is useless; "the port answers a
// browser whose id is X, and the profile you named recorded Y" tells the
// reader which of the two to change.
func explain(v Verdict, profileDir string, liveWS string, livePort int,
	m marker) string {
	switch v {
	case NoMarker:
		return fmt.Sprintf(
			"⚠️ REFUSED: something is already answering on port %d, but the "+
				"profile you asked for has no record of starting it.\n"+
				"  profile   %s\n"+
				"  live id   %s\n"+
				"  So this port belongs to a browser chrome_ready did not "+
				"start, and using it would drive a browser that may be "+
				"somebody's\n  working session rather than this capability's. "+
				"Choose a free port, or pass -accept-unverified if you know "+
				"the browser on\n  %d is the right one.",
			livePort, profileDir, browserID(liveWS), livePort)
	case Mismatch:
		return fmt.Sprintf(
			"⚠️ REFUSED: port %d answers a DIFFERENT browser from the one "+
				"this profile recorded.\n"+
				"  profile    %s\n"+
				"  recorded   browser %s on port %d\n"+
				"  live       browser %s on port %d\n"+
				"  The profile's browser has been replaced, or another "+
				"browser has taken the port. Using it would silently drive "+
				"the wrong one.",
			livePort, profileDir, m.BrowserID, m.Port,
			browserID(liveWS), livePort)
	}
	return ""
}
