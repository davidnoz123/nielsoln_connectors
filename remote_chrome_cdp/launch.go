// launch.go -- start a Chrome to drive, when there is not one.
//
// WHY THIS EXISTS. The debugging Chrome died three times in one afternoon,
// and each death turned every link in the spreadsheet into an error page.
// Nothing recovers a browser somebody closed, so the server recovering it
// is the only thing that makes the links dependable.
//
// ⚠️ A DEDICATED PROFILE, ALWAYS.
//
// Launching with the default profile while an ordinary Chrome is running
// does NOT give a debuggable browser: Chrome hands the request to the
// instance already running and the debugging flag appears to do nothing.
// That failure looks exactly like success, which is why the profile
// directory is not optional here.
//
// The cost is that this Chrome is signed out. For the pages in the manifest
// that is right: they are public pages of the client's own site. It is NOT
// enough for a console screen behind a Google login, and the honest answer
// there is to point -chrome-port at a browser you have signed in yourself.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

func chromeCandidates() []string {
	if runtime.GOOS != "windows" {
		return []string{"google-chrome", "chromium", "chrome"}
	}
	var out []string
	for _, env := range []string{"ProgramFiles", "ProgramFiles(x86)", "LocalAppData"} {
		if base := os.Getenv(env); base != "" {
			out = append(out, filepath.Join(base,
				"Google", "Chrome", "Application", "chrome.exe"))
		}
	}
	return out
}

func findChrome(explicit string) (string, error) {
	if explicit != "" {
		if _, err := os.Stat(explicit); err != nil {
			return "", fmt.Errorf("no chrome at %s: %w", explicit, err)
		}
		return explicit, nil
	}
	for _, c := range chromeCandidates() {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}
	if p, err := exec.LookPath("chrome"); err == nil {
		return p, nil
	}
	return "", fmt.Errorf("could not find chrome; pass -chrome-exe")
}

// launchChrome starts one and waits for it to answer. Visible, never
// headless: a browser the human cannot see is one they cannot check, and
// the whole point of this program is showing somebody something.
func launchChrome(exe, profile string, port int, wait time.Duration) error {
	path, err := findChrome(exe)
	if err != nil {
		return err
	}
	if profile == "" {
		return fmt.Errorf(
			"no -chrome-profile given. Refusing to invent one: the default " +
				"used to be a throwaway directory under TEMP, which is a " +
				"SIGNED OUT browser. Every Google link then stops at the " +
				"sign-in gate, and nothing says why, because a signed-out " +
				"Chrome looks exactly like a signed-in one from here")
	}
	fresh := false
	if _, err := os.Stat(profile); os.IsNotExist(err) {
		fresh = true
	}
	if err := os.MkdirAll(profile, 0o755); err != nil {
		return err
	}
	if fresh {
		fmt.Println("  NEW PROFILE at " + profile + ": this Chrome is " +
			"signed out of everything. Sign in once in the window that " +
			"opens, and the profile keeps it.")
	}
	cmd := exec.Command(path,
		fmt.Sprintf("--remote-debugging-port=%d", port),
		"--remote-allow-origins=*",
		fmt.Sprintf("--user-data-dir=%s", profile),
		"--no-first-run",
		"--no-default-browser-check",
		"about:blank")
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("could not start %s: %w", path, err)
	}
	// Released rather than waited on: this browser outlives the request
	// that needed it, and often the server too.
	go func() { _ = cmd.Wait() }()

	deadline := time.Now().Add(wait)
	for time.Now().Before(deadline) {
		if _, err := browserWS(port); err == nil {
			return nil
		}
		time.Sleep(300 * time.Millisecond)
	}
	return fmt.Errorf("started %s but nothing answered on port %d within %s",
		path, port, wait)
}
