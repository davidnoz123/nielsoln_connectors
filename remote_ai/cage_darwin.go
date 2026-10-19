//go:build darwin

// cage_darwin.go -- a boundary on macOS, around the CHILD rather than around
// this process.
//
// Windows and Linux both confine the connector itself and let children
// inherit: AppContainer via a relaunch, Landlock via a syscall on this
// process. macOS cannot be done that way from these builds. Apple's
// in-process route is sandbox_init(3), which lives in libSystem, and these
// binaries are built CGO_ENABLED=0 and CROSS-COMPILED ON A LINUX VM by
// GoReleaser. Calling a C library under those constraints means a macOS
// build host, which would change how every release is made for one
// platform.
//
// So the boundary goes around each command instead, through
// /usr/bin/sandbox-exec and an SBPL profile. Same enforcement, same kernel,
// different place to stand.
//
// MEASURED BEFORE IT WAS WRITTEN, on a real Mac, 20 Oct 2026:
//
//	inside  the workspace, sandboxed      ok      as expected
//	outside the workspace, sandboxed      denied  as expected
//	outside, NOT sandboxed (the control)  ok      as expected
//	python3 3.9.6 under the profile: read inside ok, read outside denied
//
// The third line is why the other two mean anything: a profile that denied
// everything would pass a test that only looked for denials.
//
// WHY inCage() STILL ANSWERS FALSE HERE, which looks wrong and is not.
// inCage() answers one question -- "is THIS process confined" -- and main.go
// uses it for two things: gating exec, and telling the participant
// "<authority> is holding this boundary: your other files are refused, not
// merely unasked for." On macOS the second sentence would be a LIE. This
// connector's own reads are guarded by resolve() and firstLink() in our
// code, not by the kernel. Children are genuinely kernel-confined; the
// connector is not. So the two questions are separate now, and this file
// answers the second one honestly rather than borrowing the first one's
// answer to unlock a capability.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

const sandboxExec = "/usr/bin/sandbox-exec"

// Set once by applyCage, read by childrenConfinable. Guarded because the
// connector answers requests on more than one goroutine.
var (
	cageMu      sync.RWMutex
	cageProfile string
	cageProven  bool
)

// inCage reports whether THIS PROCESS is confined. On macOS it never is:
// see the file comment. Answering true would unlock exec AND make main.go
// tell the participant the kernel is refusing access to their other files,
// which here it is not.
func inCage() bool { return false }

// cageSupported reports whether a cage can be built at all. False would mean
// the connector behaves exactly as it did before: file operations, their
// path checking, and no exec.
func cageSupported() bool {
	st, err := os.Stat(sandboxExec)
	return err == nil && !st.IsDir()
}

// childrenConfinable reports whether a command this connector starts will be
// held by the kernel. This is what gates exec on macOS.
//
// IT IS NOT A FLAG WE SET. applyCage proves it by running sandbox-exec
// against a file it must be refused, and this returns the result of that
// proof. If sandbox-exec is removed, if the profile stops being accepted, or
// if Apple changes what SBPL means, the proof fails and exec is not offered
// -- the same outcome as today, reached by measurement rather than by trust.
func childrenConfinable() bool {
	cageMu.RLock()
	defer cageMu.RUnlock()
	return cageProven && cageProfile != ""
}

// cageWrap is the argv prefix that confines a child, or nil where the
// process is already confined and children inherit.
func cageWrap() []string {
	cageMu.RLock()
	defer cageMu.RUnlock()
	if !cageProven || cageProfile == "" {
		return nil
	}
	return []string{sandboxExec, "-f", cageProfile}
}

// profileFor writes the SBPL profile and returns its path.
//
// OUTSIDE THE WORKSPACE, DELIBERATELY. A profile stored in the shared folder
// could be rewritten by the very commands it confines, and the next command
// would then run under rules a previous one had chosen. It goes in this
// process's own temp directory, which the profile itself does not grant the
// child write access to.
//
// The shape is permissive-base-then-subtract rather than (deny default).
// Deny-default is stronger and stops Python starting at all -- it needs
// /usr/lib, /System and a temp directory before it prints anything -- and a
// boundary that forces participants to abandon it is not a boundary. What
// this denies is the thing that matters: everything under the participant's
// home, with the shared folder allowed back.
func profileFor(workspace string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("no home directory to deny: %v", err)
	}
	dir, err := os.MkdirTemp("", "nielsoln-cage-")
	if err != nil {
		return "", err
	}
	// SBPL takes literal strings; a quote or backslash in a path would end
	// the string early and silently widen the profile.
	q := func(s string) string {
		return strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(s)
	}
	body := fmt.Sprintf(`(version 1)
(allow default)
(deny file-read* file-write* (subpath "%s"))
(allow file-read* file-write* (subpath "%s"))
`, q(home), q(workspace))

	path := filepath.Join(dir, "cage.sb")
	if err := os.WriteFile(path, []byte(body), 0o400); err != nil {
		return "", err
	}
	return path, nil
}

// applyCage prepares the boundary and PROVES it, returning (0, false, nil) on
// success: nothing is relaunched and this process carries on, as on Linux.
//
// The proof is the whole point. Writing a profile establishes nothing; what
// establishes something is a child that was refused.
func applyCage(workspace string) (int, bool, error) {
	if !cageSupported() {
		return 0, false, fmt.Errorf("%s is not on this machine", sandboxExec)
	}
	abs, err := filepath.Abs(workspace)
	if err != nil {
		return 0, false, err
	}
	path, err := profileFor(abs)
	if err != nil {
		return 0, false, err
	}

	// THE PROOF. A file certain to exist, certain to be outside the
	// workspace, and certain to be readable without a sandbox. If reading it
	// under the profile SUCCEEDS, the profile is not holding and exec must
	// stay shut.
	witness := filepath.Join(os.Getenv("HOME"), ".CFUserTextEncoding")
	if _, err := os.Stat(witness); err != nil {
		// No witness, no proof. Refusing to guess is the point.
		return 0, false, fmt.Errorf(
			"could not find a file outside the workspace to test the "+
				"boundary against (%v), so confinement is unproven", err)
	}
	if err := exec.Command(sandboxExec, "-f", path,
		"/bin/cat", witness).Run(); err == nil {
		return 0, false, fmt.Errorf(
			"sandbox-exec ran but did NOT refuse a file outside %s, so the "+
				"profile is not holding", abs)
	}
	// And the other direction, or "denied everything" would read as success.
	inside := filepath.Join(abs, ".cage-witness")
	if err := os.WriteFile(inside, []byte("ok\n"), 0o600); err != nil {
		return 0, false, fmt.Errorf("could not write inside the workspace: %v", err)
	}
	defer os.Remove(inside)
	if err := exec.Command(sandboxExec, "-f", path,
		"/bin/cat", inside).Run(); err != nil {
		return 0, false, fmt.Errorf(
			"the profile also refused the shared folder itself (%v), which "+
				"would leave nothing usable", err)
	}

	cageMu.Lock()
	cageProfile, cageProven = path, true
	cageMu.Unlock()
	return 0, false, nil
}

// cageBlocksLoopback: the profile denies files, not sockets, so a confined
// child can still reach a bridge on this machine. Unlike the Windows
// AppContainer, which cannot.
func cageBlocksLoopback() bool { return false }

// isLoopback is unused here, since cageBlocksLoopback is false, but main.go
// refers to it on every platform.
func isLoopback(host string) bool { return false }

// cageAuthority names what is enforcing the boundary, for the one line a
// participant is asked to look for.
func cageAuthority() string { return "macOS Seatbelt" }
