// What exec must do, and what it must refuse.
//
// The gate and the machinery are deliberately in different functions so this
// can test both: opExec holds the inCage() check, execRun holds the running.
// A test process is never confined, so the gate test gets the real answer and
// the machinery tests can still run.
//
//	go test ./...
package main

import (
	"encoding/json"
	"runtime"
	"strings"
	"testing"
	"time"
)

// echoCmd returns a command that prints *what* on this platform.
func echoCmd(what string) string {
	if runtime.GOOS == "windows" {
		return "echo " + what
	}
	return "printf '%s\\n' " + what
}

func TestExecIsRefusedWhenNotConfined(t *testing.T) {
	// The whole reason this file is allowed to exist. A test process is not in
	// an AppContainer, so this is the real gate answering, not a stub.
	if inCage() {
		t.Skip("this test process is confined, so the refusal cannot be seen")
	}
	raw, _ := json.Marshal(execArgs{Command: echoCmd("hello")})
	if _, err := opExec(t.TempDir(), raw); err == nil {
		t.Fatal("exec ran a command on an unconfined connector, which is the " +
			"one thing the cage exists to prevent")
	}
}

func TestExecRunsAndReportsWhatHappened(t *testing.T) {
	got, err := execRun(t.TempDir(), execArgs{Command: echoCmd("marker42")})
	if err != nil {
		t.Fatal(err)
	}
	r := got.(execResult)
	if !strings.Contains(r.Output, "marker42") {
		t.Errorf("the command's output did not come back: %q", r.Output)
	}
	if !r.Done || r.ExitCode == nil || *r.ExitCode != 0 {
		t.Errorf("a command that succeeded did not say so: %+v", r)
	}
}

func TestANonZeroExitIsReportedRatherThanHidden(t *testing.T) {
	// A failing build must arrive as a failing build. Reporting an error
	// instead would make the session guess.
	cmd := "exit 3"
	got, err := execRun(t.TempDir(), execArgs{Command: cmd})
	if err != nil {
		t.Fatal(err)
	}
	r := got.(execResult)
	if r.ExitCode == nil || *r.ExitCode != 3 {
		t.Errorf("exit 3 did not come back as 3: %+v", r)
	}
}

func TestATimeoutStopsTheCommandAndSaysSo(t *testing.T) {
	slow := "sleep 30"
	if runtime.GOOS == "windows" {
		slow = "ping -n 30 127.0.0.1"
	}
	start := time.Now()
	got, err := execRun(t.TempDir(), execArgs{Command: slow, TimeoutMS: 1500})
	if err != nil {
		t.Fatal(err)
	}
	if took := time.Since(start); took > 20*time.Second {
		t.Fatalf("the timeout did not stop it: took %s", took)
	}
	if r := got.(execResult); !strings.Contains(r.Error, "stopped after") {
		t.Errorf("a command that was stopped did not say so: %+v", r)
	}
}

func TestBackgroundOutputIsCollectedAndThenForgotten(t *testing.T) {
	got, err := execRun(t.TempDir(), execArgs{
		Command: echoCmd("background42"), Background: true})
	if err != nil {
		t.Fatal(err)
	}
	id := got.(execResult).ProcessID
	if id == "" {
		t.Fatal("a background command came back with no id to ask about")
	}

	var last execResult
	for i := 0; i < 50; i++ {
		out, err := execOutput(execArgs{ProcessID: id})
		if err != nil {
			t.Fatal(err)
		}
		last = out.(execResult)
		if last.Output != "" || last.Done {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !strings.Contains(last.Output, "background42") {
		t.Errorf("the background command's output never arrived: %+v", last)
	}

	// Collected, so forgotten. Asking again must say so rather than answering
	// emptily, which would read as a command that produced nothing.
	if last.Done {
		if _, err := execOutput(execArgs{ProcessID: id}); err == nil {
			t.Error("a finished and collected command is still being tracked")
		}
	}
}

func TestOutputBeyondTheCeilingIsDroppedAndCounted(t *testing.T) {
	// The 12 Sep rule: dropped with a visible marker, never buffered without
	// limit. A silent truncation would have the session reasoning about a log
	// that is not what happened.
	p := &execProc{}
	p.Write(make([]byte, execBufferBytes))
	p.Write([]byte("the end, which is the part that matters"))

	out, dropped, _, _ := p.take()
	if dropped == 0 {
		t.Fatal("output was discarded without saying how much")
	}
	if !strings.HasSuffix(out, "the part that matters") {
		t.Error("the OLDEST output should go, not the newest: the end of a " +
			"build is what says whether it worked")
	}
	if len(out) > execResponseBytes {
		t.Errorf("one reply carried %d bytes, over the %d ceiling",
			len(out), execResponseBytes)
	}
}

func TestTheChildEnvironmentIsBuiltRatherThanInherited(t *testing.T) {
	// A command must not be handed the participant's whole environment. The
	// cage is a filesystem boundary and does nothing about a token that is
	// already in the process.
	t.Setenv("GITHUB_TOKEN", "ghp_secret_value")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "aws_secret_value")

	root := t.TempDir()
	for _, kv := range childEnv(root) {
		if strings.Contains(kv, "secret_value") {
			t.Fatalf("a secret from this process reached the command: %s", kv)
		}
	}

	// And the redirects must actually point into the workspace, or git and npm
	// reach for the participant's real config and are denied by the cage.
	want := map[string]bool{"HOME": false, "USERPROFILE": false, "APPDATA": false}
	for _, kv := range childEnv(root) {
		k, v, _ := strings.Cut(kv, "=")
		if _, ok := want[k]; ok {
			if !strings.HasPrefix(v, root) {
				t.Errorf("%s points outside the workspace: %s", k, v)
			}
			want[k] = true
		}
	}
	for k, seen := range want {
		if !seen {
			t.Errorf("%s was not redirected into the workspace at all", k)
		}
	}
}

func TestKillingSomethingThatIsNotThereSaysSo(t *testing.T) {
	if _, err := execKill(execArgs{ProcessID: "proc-never"}); err == nil {
		t.Error("killing a command that does not exist reported success")
	}
}
