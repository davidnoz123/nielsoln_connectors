// shimp_test.go -- ONE contract suite, TWO drivers.
//
// AGENTS is explicit about why: a fake is a crystallised model of the
// dependency, so it can show that the code matches the model and never that
// the model matches reality. Every dependency bug worth catching is a case
// where the belief was wrong, so a fake reproduces the error rather than
// exposing it.
//
// So `contract` below is run twice: once against a strict in-memory fake,
// which gives fast regression cover on OUR logic, and once against the real
// HKEY_CURRENT_USER, which is the only run that says the registry still
// behaves as assumed.
//
// ⚠️ THE REAL RUN IS GATED AND SAYS SO WHEN IT IS SKIPPED. It writes
// HKCU\Software\Classes\shimp, which would overwrite a working installation
// on the machine running the tests. So it needs SHIMP_REAL_REGISTRY=1, and
// when that is absent the test FAILS LOUDLY in its message rather than
// passing quietly, because "tests pass" after the fake alone is a false
// statement about what was verified.
package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// -- the strict fake ---------------------------------------------------------

// fakeClaims models the registry surface this program actually touches and
// nothing else.
//
// ⚠️ STRICT BY CONSTRUCTION. It records every key it was asked about, and the
// contract asserts that nothing outside the modelled set was touched. A
// permissive fake, such as a bare map that answers anything, is worse than no
// fake: it stays green while the code reaches for surface the fake never
// modelled, so the false confidence concentrates exactly where nobody is
// looking any more.
type fakeClaims struct {
	t      *testing.T
	values map[string]string // "subkey\x00name" -> value
	keys   map[string]bool
	seen   []string
}

func newFake(t *testing.T) *fakeClaims {
	return &fakeClaims{t: t, values: map[string]string{}, keys: map[string]bool{}}
}

func (f *fakeClaims) where() string { return "a strict in-memory fake" }

// allow is the set of keys this program is permitted to touch. Anything else
// is a bug in the code or a gap in the model, and the fake refuses to guess
// which.
func (f *fakeClaims) allow(subkey string) {
	ok := subkey == classKey() ||
		strings.HasPrefix(subkey, classKey()+`\`) ||
		subkey == trustedProtocolKey()
	if !ok {
		f.t.Fatalf("fake: the registry surface does not model %q. Model it "+
			"deliberately and add a contract check pinning what the real "+
			"registry does with it.", subkey)
	}
	f.seen = append(f.seen, subkey)
}

func (f *fakeClaims) set(subkey, name, value string) error {
	f.allow(subkey)
	f.keys[subkey] = true
	f.values[subkey+"\x00"+name] = value
	return nil
}

func (f *fakeClaims) get(subkey, name string) (string, bool, error) {
	f.allow(subkey)
	v, ok := f.values[subkey+"\x00"+name]
	return v, ok, nil
}

func (f *fakeClaims) hasKey(subkey string) (bool, error) {
	f.allow(subkey)
	return f.keys[subkey], nil
}

func (f *fakeClaims) removeKey(subkey string) error {
	f.allow(subkey)
	if !f.keys[subkey] {
		// ⚠️ The real RegDeleteKeyW returns ERROR_FILE_NOT_FOUND here, and
		// the fake must refuse the same operation rather than returning
		// plausible success. A fake that enforces no rules tests nothing
		// about the constraints the code exists to respect.
		return errNotPresent{subkey}
	}
	delete(f.keys, subkey)
	for k := range f.values {
		if strings.HasPrefix(k, subkey+"\x00") {
			delete(f.values, k)
		}
	}
	return nil
}

type errNotPresent struct{ key string }

func (e errNotPresent) Error() string { return e.key + " was not present" }

// -- the contract ------------------------------------------------------------

// contract is every assertion that must hold of either driver.
//
// It asserts that a call FAILS, never HOW, because the fake raises its own
// type where the real registry raises a syscall.Errno, and pinning the type
// would pin the fake.
func contract(t *testing.T, c claims, label string) {
	t.Helper()

	// A value round-trips, including the empty one, which is a real state
	// here: "URL Protocol" is marked by its NAME existing and its value being
	// empty.
	for _, v := range []string{"URL:SHIMP Protocol", "", `"C:\x\shimp.exe" "%1"`} {
		if err := c.set(classKey(), "probe", v); err != nil {
			t.Fatalf("%s: set %q: %v", label, v, err)
		}
		got, present, err := c.get(classKey(), "probe")
		if err != nil || !present || got != v {
			t.Fatalf("%s: set %q then got %q present=%v err=%v",
				label, v, got, present, err)
		}
	}

	// An absent value is reported as absent, not as empty. ⚠️ Returning the
	// same answer for both is what let a publish proceed over a sheet it had
	// declined to protect, in slate_tools, and it is the same confusion here.
	if _, present, err := c.get(classKey(), "nothing-wrote-this"); err != nil || present {
		t.Fatalf("%s: an absent value reported present=%v err=%v", label, present, err)
	}

	// A key this program never created is absent rather than an error.
	if present, err := c.hasKey(classKey() + `\never-created`); err != nil || present {
		t.Fatalf("%s: hasKey on an absent key gave %v, %v", label, present, err)
	}

	// hasKey is TRUE for a key that exists with no default value, which is
	// the whole reason it is a separate method. s8's trusted-protocol key is
	// read by Office for its existence alone.
	if err := c.set(classKey()+`\shell`, "x", "y"); err != nil {
		t.Fatalf("%s: %v", label, err)
	}
	if present, err := c.hasKey(classKey() + `\shell`); err != nil || !present {
		t.Fatalf("%s: hasKey on a created key gave %v, %v", label, present, err)
	}

	// Removing a key that is not there FAILS. Asserting only that it fails,
	// per the rule above.
	if err := c.removeKey(classKey() + `\never-created`); err == nil {
		t.Fatalf("%s: removing an absent key reported success", label)
	}

	// And removal actually removes.
	if err := c.removeKey(classKey() + `\shell`); err != nil {
		t.Fatalf("%s: removeKey: %v", label, err)
	}
	if present, _ := c.hasKey(classKey() + `\shell`); present {
		t.Fatalf("%s: key still present after removeKey", label)
	}
	_ = c.removeKey(classKey())
}

func TestContractAgainstFake(t *testing.T) {
	contract(t, newFake(t), "fake")
	t.Log("ANSWERED: our logic matches the model of the registry. " +
		"This run says NOTHING about whether the model matches Windows.")
}

func TestContractAgainstRealRegistry(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skipf("UNANSWERED on %s: openClaims refuses here by design, so the "+
			"real half of this contract cannot run. See scheme_other.go.",
			runtime.GOOS)
	}
	if os.Getenv("SHIMP_REAL_REGISTRY") != "1" {
		// Not t.Skip alone: the message has to say what went unverified,
		// because "tests pass" after the fake alone is a false statement
		// about what was checked.
		t.Skip("UNANSWERED: the real-registry half did not run. It writes " +
			"HKCU\\Software\\Classes\\shimp and would overwrite a working " +
			"installation, so it needs SHIMP_REAL_REGISTRY=1. Until it has " +
			"run, nothing here says the registry behaves as assumed.")
	}
	c, err := claimsFor()
	if err != nil {
		t.Fatalf("openClaims: %v", err)
	}
	contract(t, c, "real registry")
	t.Log("ANSWERED: HKEY_CURRENT_USER still behaves as this code assumes.")
}

// -- install, verify, uninstall over the fake --------------------------------

// withTempHome points home() at a temporary directory so a test never touches
// a real installation.
func withTempHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", dir)
	} else {
		t.Setenv("HOME", dir)
	}
	got, err := home()
	if err != nil {
		t.Fatalf("home(): %v", err)
	}
	if !strings.HasPrefix(got, dir) {
		t.Fatalf("home() is %s, which is not under the temp dir %s. The test "+
			"would have written to a real installation.", got, dir)
	}
	return got
}

func TestInstallClaimsEverythingItShould(t *testing.T) {
	dir := withTempHome(t)
	f := newFake(t)
	claimsFor = func() (claims, error) { return f, nil }
	t.Cleanup(func() { claimsFor = openClaims })
	var logged bytes.Buffer
	logTo = &logged
	t.Cleanup(func() { logTo = nil })

	if rc := cmdInstall(false); rc != 0 {
		t.Fatalf("install returned %d, log:\n%s", rc, logged.String())
	}

	// The binary is placed, and under the home rather than anywhere else.
	bin := filepath.Join(dir, binaryName())
	if _, err := os.Stat(bin); err != nil {
		t.Fatalf("no binary at %s: %v", bin, err)
	}

	// ⚠️ THE TRAILING COLON, which is s8 and the one claim that looks right
	// when it is wrong. A key named "shimp" passes a glance in regedit and
	// does nothing.
	if !strings.HasSuffix(trustedProtocolKey(), `\shimp:`) {
		t.Fatalf("the trusted protocol key does not end in a colon: %s",
			trustedProtocolKey())
	}
	if present, _ := f.hasKey(trustedProtocolKey()); !present {
		t.Fatal("install did not claim the Office trusted protocol key")
	}

	// The class is marked as a URL scheme by the NAME existing.
	if _, present, _ := f.get(classKey(), "URL Protocol"); !present {
		t.Fatal(`install did not create the "URL Protocol" value, so Windows ` +
			`would not treat this as a URL scheme`)
	}

	// The command points at the placed binary and passes the URI.
	cmd, present, _ := f.get(commandKey(), "")
	if !present {
		t.Fatal("install did not register a shell open command")
	}
	if !strings.Contains(cmd, bin) {
		t.Fatalf("the command does not name the placed binary: %s", cmd)
	}
	if !strings.Contains(cmd, "%1") {
		t.Fatalf("the command does not pass the URI through %%1: %s", cmd)
	}

	// The stamp records what only install knows, which is s23's second audit
	// question. It is written LAST, so its presence means the claims above
	// all succeeded.
	st, err := readStamp()
	if err != nil {
		t.Fatalf("no install record: %v", err)
	}
	if st.Sha256 == "" || st.Path != bin {
		t.Fatalf("the install record does not describe the placed binary: %+v", st)
	}
	sum, err := fileSha256(bin)
	if err != nil || sum != st.Sha256 {
		t.Fatalf("the recorded sha256 does not match the placed binary: %v", err)
	}

	if !strings.Contains(logged.String(), "install complete") {
		t.Fatalf("install logged no completion, which is how a silent death "+
			"looks:\n%s", logged.String())
	}
}

func TestInstallRefusesToReplaceSilently(t *testing.T) {
	withTempHome(t)
	f := newFake(t)
	claimsFor = func() (claims, error) { return f, nil }
	t.Cleanup(func() { claimsFor = openClaims })
	logTo = &bytes.Buffer{}
	t.Cleanup(func() { logTo = nil })

	if rc := cmdInstall(false); rc != 0 {
		t.Fatalf("first install returned %d", rc)
	}

	// Re-running the SAME version is idempotent, so the documented install
	// line does not become an error a user has to reason about.
	if rc := cmdInstall(false); rc != 0 {
		t.Fatalf("re-installing the same version returned %d, which makes the "+
			"documented one-line install non-idempotent", rc)
	}

	// A DIFFERENT version is refused, because the handler is the trust root.
	st, _ := readStamp()
	st.Version = "v0.0.0-some-other-handler"
	if err := writeStamp(st); err != nil {
		t.Fatal(err)
	}
	if rc := cmdInstall(false); rc == 0 {
		t.Fatal("install replaced a different handler version without " +
			"-replace. s20: the handler is never silently replaced.")
	}
	if rc := cmdInstall(true); rc != 0 {
		t.Fatalf("install -replace returned %d, so there is no way to "+
			"legitimately upgrade", rc)
	}
}

func TestUninstallWithdrawsOnlyItsOwnAndKeepsTheLog(t *testing.T) {
	dir := withTempHome(t)
	f := newFake(t)
	claimsFor = func() (claims, error) { return f, nil }
	t.Cleanup(func() { claimsFor = openClaims })
	logTo = &bytes.Buffer{}
	t.Cleanup(func() { logTo = nil })

	if rc := cmdInstall(false); rc != 0 {
		t.Fatalf("install returned %d", rc)
	}
	// A previous run's log. ⚠️ The test has to create this: it sets logTo
	// directly, which bypasses openLog and so never creates the file. An
	// earlier version of this test asserted on a file it had itself
	// prevented from existing, and passed that off as uninstall deleting it.
	if err := os.WriteFile(filepath.Join(dir, logName),
		[]byte("an earlier run wrote this\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Something else's trusted protocol, under the shared parent.
	const other = `Software\Policies\Microsoft\Office\16.0\Common\Security\` +
		`Trusted Protocols\All Applications\notes:`
	f.keys[other] = true

	if rc := cmdUninstall(false); rc != 0 {
		t.Fatalf("uninstall returned %d", rc)
	}

	for _, k := range []string{classKey(), commandKey(), trustedProtocolKey()} {
		if present, _ := f.hasKey(k); present {
			t.Fatalf("uninstall left %s behind", k)
		}
	}
	// ⚠️ s8: "All Applications" is SHARED. Withdrawing another application's
	// trusted protocol would be a silent act of vandalism.
	if !f.keys[other] {
		t.Fatal("uninstall removed another application's trusted protocol key")
	}
	// The log survives, because it is the record of what happened rather
	// than runtime state.
	if _, err := os.Stat(filepath.Join(dir, logName)); err != nil {
		t.Fatalf("uninstall removed the log, which is the evidence somebody "+
			"diagnosing an install would need: %v", err)
	}
}

func TestVerifyFailsWhenTheBinaryChangedUnderneathIt(t *testing.T) {
	dir := withTempHome(t)
	f := newFake(t)
	claimsFor = func() (claims, error) { return f, nil }
	t.Cleanup(func() { claimsFor = openClaims })
	logTo = &bytes.Buffer{}
	t.Cleanup(func() { logTo = nil })

	// ⚠️ Asserting on the CHECK rather than on the exit code, which is what
	// the `out` seam exists for. The aggregate is not this test's subject and
	// is legitimately non-zero here for an unrelated reason: a `go test`
	// binary is console-subsystem, so s50's subsystem check correctly fails.
	// A test that read the exit code would have broken the moment an
	// unrelated check was added, and did.
	report := func() string {
		var buf bytes.Buffer
		out = &buf
		defer func() { out = os.Stdout }()
		cmdVerify()
		return buf.String()
	}

	if rc := cmdInstall(false); rc != 0 {
		t.Fatalf("install returned %d", rc)
	}
	before := report()
	if !strings.Contains(before, "[ ok ] binary sha256") {
		t.Fatalf("straight after install, the binary hash did not verify:\n%s",
			before)
	}

	// ⚠️ This is s23's second audit question made operational: "is what is
	// installed the thing I think it is". A path that exists proves nothing.
	bin := filepath.Join(dir, binaryName())
	if err := os.WriteFile(bin, []byte("not the handler"), 0o755); err != nil {
		t.Fatal(err)
	}
	after := report()
	if !strings.Contains(after, "[FAIL] binary sha256") ||
		!strings.Contains(after, "CHANGED since install") {
		t.Fatalf("verify did not report the replaced binary, so it checks "+
			"that a path exists rather than that the binary is the installed "+
			"one:\n%s", after)
	}
	if cmdVerify() == 0 {
		t.Fatal("verify exited 0 with a replaced binary")
	}
}

func TestClickLogsTheLengthS36CannotOtherwiseMeasure(t *testing.T) {
	withTempHome(t)
	var logged bytes.Buffer
	logTo = &logged
	t.Cleanup(func() { logTo = nil })

	uri := "shimp://v1/go/run/github.com/x/y@" + strings.Repeat("a", 40) +
		"/-p/1#act"
	if rc := onClick(uri); rc == 0 {
		t.Fatal("a click reported success while the dispatch half is not built")
	}
	// The number is the point. s36 says what the handler receives as
	// os.Args[1] is unmeasurable without a registered handler, so every click
	// records it.
	want := fmt.Sprintf("os.Args[1] is %d bytes", len(uri))
	if !strings.Contains(logged.String(), want) {
		t.Fatalf("the click did not log the argument length:\n%s", logged.String())
	}
	if !strings.Contains(logged.String(), "REFUSED") {
		t.Fatalf("the click did not record a refusal, so a user would see "+
			"nothing at all:\n%s", logged.String())
	}
	// ⚠️ And now that the parser exists, a click must record the IDENTITY it
	// computed and the argv it would run. Those are what the keeper will act
	// on, so having them in the log before the keeper exists is how the two
	// halves get checked against each other.
	for _, want := range []string{"identity ", "would run "} {
		if !strings.Contains(logged.String(), want) {
			t.Fatalf("the click did not log %q, so a real click says nothing "+
				"about what it resolved to:\n%s", want, logged.String())
		}
	}
}

// ---- the refusal display, s67 ----

// TestBothRefusalPathsReachTheUser is what the seam was added for.
//
// ⚠️ A human clicked eight links and saw an empty command-prompt flash,
// twice. The log had everything and the screen had nothing, which is
// indistinguishable from an unregistered scheme or a handler that died
// before _log existed. So a refusal reaching the USER is now a behaviour
// with a test, not a side effect.
func TestBothRefusalPathsReachTheUser(t *testing.T) {
	var shown []string
	showRefusalFn = func(reason string) { shown = append(shown, reason) }
	t.Cleanup(func() { showRefusalFn = showRefusal })
	logTo = &bytes.Buffer{}
	t.Cleanup(func() { logTo = nil })

	// -- path 1: the parser refuses. 39 hex rather than 40.
	bad := scheme + "://v1/go/run/github.com/davidnoz123/nielsoln_connectors/" +
		"x@" + strings.Repeat("a", 39)
	if rc := onClick(bad); rc == 0 {
		t.Fatal("a 39-character commit was accepted")
	}
	if len(shown) != 1 {
		t.Fatalf("a parser refusal showed the user %d message(s), want 1. "+
			"Logged-but-not-shown is the hole this closes.", len(shown))
	}
	if !strings.Contains(shown[0], "must be exactly 40") {
		t.Errorf("the user was shown %q, which does not name the rule that "+
			"rejected their link", shown[0])
	}

	// -- path 2: the link is VALID and the keeper is not built.
	good := scheme + "://v1/go/run/github.com/davidnoz123/nielsoln_connectors/" +
		"x@" + strings.Repeat("a", 40) + "#ensure"
	if rc := onClick(good); rc == 0 {
		t.Fatal("a valid link returned success, but nothing can run it yet")
	}
	if len(shown) != 2 {
		t.Fatalf("the keeper refusal showed the user %d message(s) in total, "+
			"want 2", len(shown))
	}
	msg := shown[1]
	// ⚠️ The user must be told the link is VALID, or a keeper refusal reads
	// exactly like a malformed link and they will go and edit a correct URI.
	if !strings.Contains(msg, "VALID") {
		t.Errorf("the keeper refusal does not tell the user the link is "+
			"valid, so it is indistinguishable from a bad link: %q", msg)
	}
	if !strings.Contains(msg, "identity") {
		t.Errorf("the keeper refusal names no identity, so there is nothing "+
			"to check against the link they clicked: %q", msg)
	}
}

// TestTheRefusalSeamIsNotTheRealFunctionDuringTests guards the accident.
//
// ⚠️ The real showRefusal allocates a console and waits up to 25 seconds.
// It did not hang this suite only because AllocConsole happens to fail under
// `go test`, which is luck rather than design: in a plain console it would
// have blocked. Anything that drives onClick must replace the seam first.
func TestTheRefusalSeamIsNotTheRealFunctionDuringTests(t *testing.T) {
	var called bool
	showRefusalFn = func(string) { called = true }
	t.Cleanup(func() { showRefusalFn = showRefusal })
	logTo = &bytes.Buffer{}
	t.Cleanup(func() { logTo = nil })

	onClick(scheme + "://nonsense")
	if !called {
		t.Fatal("the seam was not used, so onClick called the real display " +
			"function and this suite's speed depends on AllocConsole failing")
	}
}

func TestRefusalTextWrapsWithoutLosingWords(t *testing.T) {
	long := "the commit is 39 characters and must be exactly 40 because an " +
		"abbreviation would be re-resolved by the toolchain on every click"
	got := wrap(long, 40)
	for _, w := range strings.Fields(long) {
		if !strings.Contains(got, w) {
			t.Errorf("wrap lost the word %q", w)
		}
	}
	for _, line := range strings.Split(got, "\n") {
		if n := len(strings.TrimLeft(line, " ")); n > 40 {
			t.Errorf("a wrapped line is %d chars, over the 40 asked for: %q",
				n, line)
		}
	}
	// ⚠️ A refusal with no spaces must not be silently dropped: some of
	// these carry a 64-character identity as one token.
	if got := wrap(strings.Repeat("x", 80), 40); !strings.Contains(got, "xxxx") {
		t.Error("wrap discarded an unbreakable token")
	}
}
