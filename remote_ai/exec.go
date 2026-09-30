// exec.go -- one op standing in for Claude Code's Bash, BashOutput and
// KillShell.
//
// THIS FILE ADDS os/exec TO THIS PROGRAM. Every review of the published source
// so far has swept for it and reported it absent, and that sentence is no
// longer true. It is the whole reason the AppContainer had to land first: a
// process this connector starts is not bound by resolve(), firstLink() or any
// other path check in main.go, so without the cage this file would not weaken
// containment, it would delete it. `exec` is therefore offered ONLY when
// inCage() reports that the kernel is holding the boundary, and inCage() asks
// the process token rather than anything this program set for itself.
//
// Shape taken from Claude Code's own tools, because that is what the remote
// end already knows how to drive: a command through a shell, a timeout, an
// option to run in the background, a way to read what a background command has
// said so far, and a way to kill it. slate_surface.py has mapped Bash,
// BashOutput and KillShell onto one connector op called `exec` since before
// this was written, so they are actions here rather than three ops.
//
// Three constraints come from PROTOCOL.md rather than from taste:
//
//   - One reply per request. Streaming is therefore background-plus-poll, not
//     many responses to one request.
//   - A repeated request id returns the cached response and does NOT
//     re-execute. The bridge re-sends anything unanswered when a socket drops,
//     and a command that ran twice because a laptop lost wifi is the worst bug
//     this protocol could have. Handled by the existing `done` cache in
//     main.go, which is why nothing here is idempotent by itself.
//   - Output is coalesced, capped, and dropped WITH A VISIBLE MARKER rather
//     than buffered without limit. On 12 Sep a dev server streaming into a
//     stalled reader took the room down: ttyd is single-threaded, the browser
//     stopped draining, the event loop blocked and the relay's proxy thread
//     hung on it. "Build the ceiling in before the feature that needs it."
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	// Claude Code's own defaults, so a command that works there behaves the
	// same here rather than dying for a reason the model cannot see.
	execDefaultTimeout = 120 * time.Second
	execMaxTimeout     = 600 * time.Second

	// What one process may hold in memory. Beyond this the OLDEST output is
	// dropped and the count is reported, because the end of a build is what
	// says whether it worked.
	execBufferBytes = 256 * 1024

	// What one reply may carry, matching read_file's ceiling.
	execResponseBytes = 64 * 1024

	// Matching the protocol's unacknowledged-requests limit. A session that
	// wants a ninth concurrent process has lost track of the eight it has.
	execMaxProcesses = 8
)

var (
	execMu      sync.Mutex
	execRunning = map[string]*execProc{}
	execCounter int
)

type execProc struct {
	id  string
	cmd *exec.Cmd
	// Ends the command AND everything it started. Not cmd.Process.Kill():
	// that kills the shell and leaves its children running, which a 1.5s
	// timeout taking 29 seconds is how we found out. See proctree_*.go.
	kill    func()
	mu      sync.Mutex
	buf     []byte
	dropped int
	done    bool
	exit    int
	err     string
}

// Write is where the command's stdout and stderr both go. Bounded: when it is
// full the OLDEST bytes go and the tally rises, because the end of a build is
// what says whether it worked.
func (p *execProc) Write(b []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.buf = append(p.buf, b...)
	if over := len(p.buf) - execBufferBytes; over > 0 {
		p.buf = p.buf[over:]
		p.dropped += over
	}
	return len(b), nil
}

// take returns what has accumulated and empties it. Destructive because the
// caller is polling: leaving it would send the same build log again next time.
func (p *execProc) take() (string, int, bool, int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := p.buf
	dropped := p.dropped
	p.buf, p.dropped = nil, 0
	if len(out) > execResponseBytes {
		dropped += len(out) - execResponseBytes
		out = out[len(out)-execResponseBytes:]
	}
	return string(out), dropped, p.done, p.exit
}

func (p *execProc) finish(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.done = true
	if ee, ok := err.(*exec.ExitError); ok {
		p.exit = ee.ExitCode()
	} else if err != nil {
		p.err = err.Error()
		p.exit = -1
	}
}

type execArgs struct {
	Action     string `json:"action"`
	Command    string `json:"command"`
	Background bool   `json:"background"`
	TimeoutMS  int    `json:"timeout_ms"`
	ProcessID  string `json:"process_id"`
}

type execResult struct {
	ProcessID string `json:"process_id,omitempty"`
	Output    string `json:"output"`
	Done      bool   `json:"done"`
	ExitCode  *int   `json:"exit_code,omitempty"`
	// Named rather than silent. A build whose middle vanished must say so, or
	// the session reasons about a log that is not what happened.
	DroppedBytes int    `json:"dropped_bytes,omitempty"`
	Error        string `json:"error,omitempty"`
}

// shellFor returns the shell and the flag that makes it run a string.
func shellFor() (string, string) {
	if runtime.GOOS == "windows" {
		comspec := os.Getenv("COMSPEC")
		if comspec == "" {
			comspec = "cmd.exe"
		}
		return comspec, "/c"
	}
	return "/bin/sh", "-c"
}

// childEnv builds the environment a command runs with, rather than handing it
// this process's own.
//
// Inheriting would give every command the participant's whole environment,
// which on a developer's machine routinely carries GITHUB_TOKEN, AWS_ keys and
// proxy credentials. The cage is a filesystem boundary and does nothing about
// a secret that is already in the process.
//
// The redirects do a second job. Inside the cage %USERPROFILE% and %APPDATA%
// still NAME the participant's real folders, and those are now denied, so git
// and npm would fail with errors about paths nobody asked them to touch.
// Pointing them into the workspace turns that into a clean, empty, disposable
// environment: git never reads the real .gitconfig and its credential helper,
// npm never reads the real .npmrc and its registry token. Measured 1 Oct 2026:
// the cage already redirects TEMP and LOCALAPPDATA to a private writable area
// of its own, but leaves USERPROFILE and APPDATA pointing at denied paths.
func childEnv(root string) []string {
	home := filepath.Join(root, ".session-home")
	tmp := filepath.Join(root, ".session-tmp")
	os.MkdirAll(home, 0o755)
	os.MkdirAll(tmp, 0o755)

	// An allowlist, so adding a variable is a decision somebody made.
	env := []string{}
	for _, k := range []string{
		"PATH", "PATHEXT", "COMSPEC", "SYSTEMROOT", "WINDIR", "SYSTEMDRIVE",
		"NUMBER_OF_PROCESSORS", "PROCESSOR_ARCHITECTURE", "OS", "LANG", "LC_ALL",
	} {
		if v := os.Getenv(k); v != "" {
			env = append(env, k+"="+v)
		}
	}
	return append(env,
		"HOME="+home,
		"USERPROFILE="+home,
		"APPDATA="+home,
		"LOCALAPPDATA="+home,
		"XDG_CACHE_HOME="+filepath.Join(home, ".cache"),
		"GIT_CONFIG_GLOBAL="+filepath.Join(home, ".gitconfig"),
		"GIT_CONFIG_SYSTEM="+filepath.Join(home, ".gitconfig-system"),
		"npm_config_cache="+filepath.Join(home, ".npm"),
		"npm_config_userconfig="+filepath.Join(home, ".npmrc"),
		"PIP_CACHE_DIR="+filepath.Join(home, ".pip"),
		"TEMP="+tmp, "TMP="+tmp, "TMPDIR="+tmp,
	)
}

func opExec(root string, raw json.RawMessage) (any, error) {
	// The gate. Asked of the kernel on every call rather than cached at
	// startup, because a capability this program grants itself on its own
	// say-so is not a capability anybody should trust.
	if !inCage() {
		return nil, refuse("not_allowed",
			"Running commands is only offered when this session is confined "+
				"by the operating system, and this one is not. The files in "+
				"the shared folder can still be read and written.")
	}

	var a execArgs
	if err := json.Unmarshal(raw, &a); err != nil {
		return nil, refuse("bad_request", "That request could not be read.")
	}

	switch a.Action {
	case "", "run":
		return execRun(root, a)
	case "output":
		return execOutput(a)
	case "kill":
		return execKill(a)
	default:
		return nil, refuse("bad_request",
			"%s is not something exec does. It runs a command, reads what a "+
				"background one has said, or kills it.", a.Action)
	}
}

func execRun(root string, a execArgs) (any, error) {
	if strings.TrimSpace(a.Command) == "" {
		return nil, refuse("bad_request", "There was no command to run.")
	}
	timeout := execDefaultTimeout
	if a.TimeoutMS > 0 {
		timeout = time.Duration(a.TimeoutMS) * time.Millisecond
		if timeout > execMaxTimeout {
			timeout = execMaxTimeout
		}
	}

	shell, flag := shellFor()
	cmd := exec.Command(shell, flag, a.Command)
	cmd.Dir = root
	cmd.Env = childEnv(root)

	p := &execProc{cmd: cmd}
	cmd.Stdout = p
	cmd.Stderr = p

	// The participant watches this window, and PROTOCOL.md leaves "whether the
	// participant sees a console" open. Naming the command is the least that
	// question can be answered with: a session running something on somebody's
	// machine should be visible to them.
	logf("running: %s", a.Command)

	prepareTree(cmd)
	if err := cmd.Start(); err != nil {
		return nil, refuse("not_allowed", "That command could not be started.")
	}
	p.kill = superviseTree(cmd)

	if a.Background {
		execMu.Lock()
		if len(execRunning) >= execMaxProcesses {
			execMu.Unlock()
			p.kill()
			return nil, refuse("not_allowed",
				"There are already %d commands running in the background, "+
					"which is the most this session keeps track of.",
				execMaxProcesses)
		}
		execCounter++
		p.id = fmt.Sprintf("proc-%d", execCounter)
		execRunning[p.id] = p
		execMu.Unlock()

		go func() { p.finish(cmd.Wait()) }()
		return execResult{ProcessID: p.id, Done: false}, nil
	}

	waited := make(chan error, 1)
	go func() { waited <- cmd.Wait() }()
	select {
	case err := <-waited:
		p.finish(err)
	case <-time.After(timeout):
		p.kill()
		<-waited
		p.finish(nil)
		out, dropped, _, _ := p.take()
		return execResult{
			Output: out, Done: true, DroppedBytes: dropped,
			Error: fmt.Sprintf("stopped after %s", timeout),
		}, nil
	}
	out, dropped, done, exit := p.take()
	return execResult{Output: out, Done: done, ExitCode: &exit,
		DroppedBytes: dropped, Error: p.err}, nil
}

func execOutput(a execArgs) (any, error) {
	execMu.Lock()
	p, ok := execRunning[a.ProcessID]
	execMu.Unlock()
	if !ok {
		return nil, refuse("not_found",
			"There is no background command called %s in this session.",
			a.ProcessID)
	}
	out, dropped, done, exit := p.take()
	r := execResult{ProcessID: p.id, Output: out, Done: done,
		DroppedBytes: dropped, Error: p.err}
	if done {
		r.ExitCode = &exit
		// Forgotten only once it has been collected, so the last of a build's
		// output is never lost to tidying up.
		execMu.Lock()
		delete(execRunning, p.id)
		execMu.Unlock()
	}
	return r, nil
}

func execKill(a execArgs) (any, error) {
	execMu.Lock()
	p, ok := execRunning[a.ProcessID]
	execMu.Unlock()
	if !ok {
		return nil, refuse("not_found",
			"There is no background command called %s in this session.",
			a.ProcessID)
	}
	if p.kill != nil {
		p.kill()
	}
	logf("killed: %s", p.id)
	out, dropped, _, exit := p.take()
	execMu.Lock()
	delete(execRunning, p.id)
	execMu.Unlock()
	return execResult{ProcessID: p.id, Output: out, Done: true,
		ExitCode: &exit, DroppedBytes: dropped}, nil
}
