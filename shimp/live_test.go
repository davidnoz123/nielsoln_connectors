// live_test.go -- the install cycle against the real machine, end to end.
//
// This is the second driver the house rules insist on. The fake in
// shimp_test.go says our logic matches a model of the registry; only this
// says an install actually installs, that Windows actually routes a click to
// it, and that uninstall actually leaves nothing behind.
//
// ⚠️ IT NEVER TOUCHES THE PRODUCTION SCHEME. Not "defaults away from it":
// REFUSES it by name, before anything is touched. A default can be overridden
// by a careless invocation, and the thing on the other side of a careless
// invocation is somebody's working installation.
//
// ⚠️ AND IT BUILDS A REAL BINARY RATHER THAN INSTALLING ITSELF. The first
// version of this file called cmdInstall in process, which copies
// os.Executable() -- and under `go test` that is the TEST binary. Windows
// would then have invoked a test binary as the handler, so the click would
// never have reached onClick, and the test would have reported on something
// it had not exercised. Driving the built CLI as a user would is the only
// version of this test that means anything.
package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// testScheme is deliberately NOT derived from `scheme`, so a future change to
// the default cannot quietly point this file at production.
const testScheme = "shimp-selftest"

// run executes the built CLI and reports its combined output.
func run(t *testing.T, exe string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command(exe, args...)
	out, err := cmd.CombinedOutput()
	t.Logf("$ %s %s\n%s", filepath.Base(exe), strings.Join(args, " "), out)
	return string(out), err
}

func TestLiveInstallCycle(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skipf("UNANSWERED on %s: scheme_other.go refuses to install here "+
			"by design, so the live cycle cannot run. s22 is the build order.",
			runtime.GOOS)
	}
	if os.Getenv("SHIMP_LIVE_INSTALL") != "1" {
		t.Skip("UNANSWERED: the live install cycle did not run. It writes a " +
			"real registry class, a real Office trusted-protocol key and a " +
			"real binary, then removes them, so it needs " +
			"SHIMP_LIVE_INSTALL=1. Until it has run, nothing here says an " +
			"install installs, that a click reaches the handler, or that " +
			"uninstall leaves nothing behind.")
	}

	// ⚠️ THE REFUSAL, checked before anything is touched.
	if testScheme == "shimp" {
		t.Fatal("the self-test is pointed at the PRODUCTION scheme. Refusing " +
			"rather than running: this test installs and uninstalls, and on " +
			"the production scheme that would destroy a working handler.")
	}

	// The scheme is switched in THIS process only so that home(), classKey()
	// and the rest name the test installation. The installed handler recovers
	// the same scheme from the stamp beside its own binary.
	if err := useScheme(testScheme); err != nil {
		t.Fatalf("the test scheme is not usable: %v", err)
	}
	t.Cleanup(func() { scheme = "shimp" })

	c, err := claimsFor()
	if err != nil {
		t.Fatalf("openClaims: %v", err)
	}
	dir, err := home()
	if err != nil {
		t.Fatal(err)
	}

	// The machine must be clean BEFORE, or this test cannot tell its own
	// leavings from somebody else's. Refusing beats adopting.
	for _, k := range []string{classKey(), trustedProtocolKey()} {
		if present, err := c.hasKey(k); err != nil {
			t.Fatalf("cannot read %s: %v", k, err)
		} else if present {
			t.Fatalf("%s already exists before the test ran. Refusing rather "+
				"than adopting state this test did not create: run "+
				"`shimp uninstall -purge -scheme %s` first.", k, testScheme)
		}
	}
	if _, err := os.Stat(dir); err == nil {
		t.Fatalf("%s already exists before the test ran; purge it first", dir)
	}

	// Whatever happens below, the machine goes back. A failed assertion must
	// not leave a registered scheme behind.
	t.Cleanup(func() {
		for _, k := range []string{trustedProtocolKey(), commandKey(),
			classKey() + `\shell\open`, classKey() + `\shell`, classKey()} {
			_ = c.removeKey(k)
		}
		_ = os.RemoveAll(dir)
	})

	// ⚠️ The production home must not exist either, because the whole point
	// of a test scheme is that a test cycle cannot touch production. An
	// earlier version opened the log BEFORE parsing -scheme, so every
	// invocation created ~/.shimp regardless, and no -purge could clean up
	// after it. That leak was found by looking at the machine rather than by
	// any assertion, so here is the assertion.
	prodHome := strings.TrimSuffix(dir, "-"+strings.TrimPrefix(testScheme, "shimp-"))
	prodExisted := false
	if _, err := os.Stat(prodHome); err == nil {
		prodExisted = true
		t.Logf("note: %s already exists, so this test can only check that it "+
			"is not CREATED, not that it stays absent", prodHome)
	}
	t.Cleanup(func() {
		if prodExisted {
			return
		}
		if _, err := os.Stat(prodHome); err == nil {
			t.Errorf("a test-scheme cycle created the PRODUCTION home %s. "+
				"Something is deriving a path before -scheme is applied.",
				prodHome)
		}
	})

	// -- build the real artifact
	exe := filepath.Join(t.TempDir(), "shimp.exe")
	if out, err := exec.Command("go", "build", "-o", exe, ".").CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}

	// -- install, driven as a user would
	if _, err := run(t, exe, "reinstall", "-scheme", testScheme); err != nil {
		t.Fatalf("live install failed: %v", err)
	}
	st, err := readStamp()
	if err != nil {
		t.Fatalf("no install record: %v", err)
	}
	if st.Scheme != testScheme {
		t.Fatalf("the stamp records scheme %q, not %q, so the installed "+
			"handler could not recover which scheme it serves",
			st.Scheme, testScheme)
	}
	if !strings.HasPrefix(st.Path, dir) {
		t.Fatalf("the installed binary is at %s, which is outside the test "+
			"home %s", st.Path, dir)
	}

	// -- verify, against the real registry
	if _, err := run(t, exe, "verify", "-scheme", testScheme); err != nil {
		t.Fatalf("verify failed against the real registry straight after a "+
			"successful install: %v", err)
	}

	// -- the click. ⚠️ Through ShellExecuteW, so WINDOWS does the routing,
	// which is the only way to take s36's measurement.
	uri := testScheme + "://v1/go/run/github.com/davidnoz123/" +
		"nielsoln_connectors/editor_open@" + strings.Repeat("a", 40) +
		"/-editor/code#open?file=x.go&line=1735"
	if err := dispatchURI(uri); err != nil {
		t.Fatalf("the OS refused to route the URI, so the registration is "+
			"not doing what verify claimed: %v", err)
	}

	// The handler is a separate process the OS started, so its log is the
	// only channel. Polled rather than slept: a fixed sleep is either flaky
	// or slow, and usually both.
	logPath := filepath.Join(dir, logName)
	var body string
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		if raw, err := os.ReadFile(logPath); err == nil {
			body = string(raw)
			if strings.Contains(body, "os.Args[1] is ") {
				break
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	if !strings.Contains(body, "os.Args[1] is ") {
		t.Fatalf("20s after ShellExecuteW the handler had not logged a "+
			"click. Either Windows did not route it, or the handler died "+
			"before it could write. Log so far:\n%s", body)
	}

	// REPORTED rather than asserted: the point of a measurement is to learn
	// the number, and asserting a guess would only pin the guess.
	for _, line := range strings.Split(body, "\n") {
		if strings.Contains(line, "os.Args[1] is ") {
			t.Logf("MEASURED (s36, OS hop only): %s", strings.TrimSpace(line))
		}
	}
	t.Logf("  the URI as sent was %d bytes", len(uri))
	t.Log("UNANSWERED: what Excel or a browser does to a URI BEFORE handing " +
		"it to the OS. This measured the registry hop only.")

	// -- uninstall -purge, and the machine must be as we found it
	if _, err := run(t, exe, "uninstall", "-purge", "-scheme", testScheme); err != nil {
		t.Fatalf("uninstall -purge failed: %v", err)
	}
	for _, k := range []string{classKey(), commandKey(), trustedProtocolKey()} {
		if present, err := c.hasKey(k); err != nil {
			t.Fatalf("cannot read %s after uninstall: %v", k, err)
		} else if present {
			t.Fatalf("uninstall -purge left %s behind, so a test cycle "+
				"accumulates state on the machine", k)
		}
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("uninstall -purge left %s behind: %v", dir, err)
	}
	t.Log("ANSWERED: an install installs, Windows routes a click to the " +
		"installed handler, and uninstall -purge leaves the machine as it " +
		"was found.")
}

// TestPurgeCannotDeleteItselfAndSaysSo pins the one case the cycle above
// structurally cannot reach.
//
// ⚠️ TestLiveInstallCycle drives an EXTERNALLY built binary, so its purge
// deletes a file it is not running from and always succeeds. Invoked as the
// INSTALLED binary, purge cannot remove itself: Windows refuses to delete a
// running executable. Measured 12 Oct 2026 after a real one-line install.
//
// A harness hides exactly the case it does not model, so this is the case,
// modelled.
func TestPurgeCannotDeleteItselfAndSaysSo(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skipf("UNANSWERED on %s: a running executable is deletable here, so "+
			"this failure does not arise", runtime.GOOS)
	}
	if os.Getenv("SHIMP_LIVE_INSTALL") != "1" {
		t.Skip("UNANSWERED: needs SHIMP_LIVE_INSTALL=1, because it installs " +
			"for real in order to then run the installed binary.")
	}
	const sch = "shimp-selfpurge"
	if err := useScheme(sch); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { scheme = "shimp" })

	c, err := claimsFor()
	if err != nil {
		t.Fatal(err)
	}
	dir, err := home()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		for _, k := range []string{trustedProtocolKey(), commandKey(),
			classKey() + `\shell\open`, classKey() + `\shell`, classKey()} {
			_ = c.removeKey(k)
		}
		_ = os.RemoveAll(dir)
	})

	exe := filepath.Join(t.TempDir(), "shimp.exe")
	if out, err := exec.Command("go", "build", "-o", exe, ".").CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	if _, err := run(t, exe, "reinstall", "-scheme", sch); err != nil {
		t.Fatalf("install: %v", err)
	}

	// Now purge using the INSTALLED binary, which is the case that cannot
	// succeed.
	installed := filepath.Join(dir, sch+".exe")
	out, err := run(t, installed, "uninstall", "-purge", "-scheme", sch)
	if err == nil {
		t.Fatal("purge reported success while deleting its own running " +
			"executable, which Windows does not permit. A success here means " +
			"the failure is being swallowed.")
	}
	if !strings.Contains(out, "could not fully purge") {
		t.Fatalf("the failure was not reported in terms a user could act "+
			"on:\n%s", out)
	}
	if !strings.Contains(out, "running executable") {
		t.Fatalf("the report does not say WHY, so it reads as a mystery:\n%s", out)
	}
	// And the claims must still be gone: the binary surviving is acceptable,
	// a registered scheme surviving is not.
	for _, k := range []string{classKey(), commandKey(), trustedProtocolKey()} {
		if present, _ := c.hasKey(k); present {
			t.Fatalf("the registry claim %s survived a failed purge, so a "+
				"click would still reach a handler the user uninstalled", k)
		}
	}
	t.Log("ANSWERED: purge cannot delete its own running binary, says so " +
		"precisely, and withdraws every registry claim regardless.")
}
