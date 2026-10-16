// corpus_run.go -- the two modes that drive corpus.go.
//
//	-corpus   dispatch every shape down the baseline route, judge each
//	-serve    render the same shapes as a page for a human to click
//
// ⚠️ ONE TABLE, TWO CONSUMERS, which is the point. A throwaway page of
// hand-written links was built once and had drifted from what the harness
// checked inside a single turn. The served page now renders `shapes`, so a
// human clicking and a machine dispatching are looking at the same thing by
// construction.
package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// buildAndInstall puts a real handler on the machine and returns its path.
//
// ⚠️ -ldflags is a PARAMETER here, not a constant, and that is a finding
// rather than a convenience. linkprobe had only ever built the handler
// console-subsystem, so it had only ever measured the noisy case. A human
// clicking a windowsgui handler got total silence, and the harness could
// not have told him that because it reads a log rather than a screen.
func buildAndInstall(root, work, scheme string, windowsGUI bool) (string, error) {
	exe := filepath.Join(work, "shimp.exe")
	args := []string{"build"}
	if windowsGUI {
		args = append(args, "-ldflags", "-H=windowsgui")
	}
	args = append(args, "-o", exe, ".")
	cmd := exec.Command("go", args...)
	cmd.Dir = filepath.Join(root, "shimp")
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("building shimp failed: %v\n%s", err, out)
	}
	if out, err := exec.Command(exe, "install", "-scheme", scheme).
		CombinedOutput(); err != nil {
		return "", fmt.Errorf("install -scheme %s failed: %v\n%s", scheme, err, out)
	}
	return exe, nil
}

// doPurge is the recovery path for an installation nothing cleaned up.
//
// ⚠️ IT EXISTS BECAUSE A HARD KILL LEAVES ONE BEHIND, measured: `-serve`
// purges on Ctrl-C through signal.Notify, but nothing survives being killed
// outright, and then the temp build directory is gone too. So the user has a
// registered scheme and no obvious binary to withdraw it with.
//
// ⚠️ And the binary install PLACED in the home cannot do it either: Windows
// will not delete a running executable, so `uninstall -purge` run from there
// withdraws the registry claim and then refuses to remove its own directory,
// which is exactly what shimp's own test pins. Building a fresh copy in a
// temp directory sidesteps that, which is the only reason this is a mode
// rather than a line of documentation.
func doPurge() int {
	root, err := repoRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	work, err := os.MkdirTemp("", "linkprobe-purge-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer os.RemoveAll(work)

	exe := filepath.Join(work, "shimp.exe")
	cmd := exec.Command("go", "build", "-o", exe, ".")
	cmd.Dir = filepath.Join(root, "shimp")
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "building shimp failed: %v\n%s", err, out)
		return 1
	}
	out, err := exec.Command(exe, "uninstall", "-purge", "-scheme", testScheme).
		CombinedOutput()
	fmt.Fprintf(progress, "%s", out)
	if err != nil {
		// Reported rather than swallowed, and NOT treated as fatal: there
		// may have been nothing to purge, which is the common case and is
		// the outcome the user wanted.
		fmt.Fprintf(progress, "\n(exit %v -- if nothing was installed, that "+
			"is the expected result)\n", err)
	}
	return 0
}

func purge(exe, scheme string) {
	out, err := exec.Command(exe, "uninstall", "-purge", "-scheme", scheme).
		CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"\n⚠️ PURGE FAILED, so this machine still carries a %s registry "+
				"claim: %v\n%s\n", scheme, err, out)
		return
	}
	fmt.Fprintf(progress, "\npurged %s\n", scheme)
}

// doCorpus sends every shape down the baseline route and judges each.
//
// ⚠️ THE BASELINE ROUTE ONLY, deliberately. The handler's verdict is a
// function of the bytes it receives, and routes.go already establishes which
// routes deliver bytes faithfully. Crossing the two tables would run ninety
// combinations to learn nothing the axes do not say separately.
func doCorpus(only string, keep, noGUI bool) int {
	if testScheme == "shimp" {
		fmt.Fprintln(os.Stderr, "REFUSED: testScheme is the production scheme.")
		return 1
	}
	root, err := repoRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	work, err := os.MkdirTemp("", "linkprobe-corpus-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer os.RemoveAll(work)

	// ⚠️ WINDOWSGUI BY DEFAULT HERE, which -gui=false turns off.
	//
	// SHIMP_NO_CONSOLE stops the handler ALLOCATING a console. It cannot
	// stop Windows GRANTING one: a console-subsystem binary launched through
	// the registry is given a window before any of its own code runs, and
	// that grant is what a human sees flash and lose focus to.
	//
	// Only the subsystem decides that, so the corpus builds the handler
	// windowsgui. Between the two, Windows grants nothing and shimp
	// allocates nothing, which is the only combination that produces zero
	// windows.
	//
	// ⚠️ And it costs something, stated rather than hidden: s66 noted the
	// harness had only ever measured the console-subsystem build. With no
	// console wanted either way the subsystem makes no difference to the
	// gate, and -gui=false restores the old behaviour for anyone who wants
	// to watch.
	exe, err := buildAndInstall(root, work, testScheme, !noGUI)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if !keep {
		defer purge(exe, testScheme)
	}
	logPath, err := probeLogPath()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	sub := "windowsgui + SHIMP_NO_CONSOLE, so NO window is created"
	if noGUI {
		sub = "console-subsystem, so Windows grants a console window per click"
	}
	fmt.Fprintf(progress, "%d shapes, handler built %s\n\n", len(shapes), sub)

	var results []CorpusResult
	for _, s := range shapes {
		if only != "" && s.ID != only {
			continue
		}
		fmt.Fprintf(progress, "  %-24s ", s.ID)
		before := logSize(logPath)
		link := s.Link(testScheme)

		if err := dispatchViaBaseline(exe, link); err != nil {
			// ⚠️ A dispatch failure is NOT a verdict. ShellExecuteW refusing
			// the URI means the measurement was never taken, and reporting
			// that as a rejection would credit the handler with a decision
			// it never made.
			fmt.Fprintf(progress, "DISPATCH FAILED  %v\n", err)
			results = append(results, CorpusResult{Shape: s,
				Problems: []string{"the OS refused to route the URI: " + err.Error()}})
			continue
		}

		saw := waitForVerdict(logPath, before)
		res := judge(s, saw, testScheme)
		results = append(results, res)
		mark := "as declared"
		if !res.OK() {
			mark = "⚠️ FINDING"
		}
		fmt.Fprintf(progress, "%-16s %s\n", saw.Verdict, mark)
	}
	reportCorpus(results)
	for _, r := range results {
		if !r.OK() {
			return 1
		}
	}
	return 0
}

// dispatchViaBaseline uses `shimp dispatch`, which is ShellExecuteW: the same
// path a click takes.
func dispatchViaBaseline(exe, link string) error {
	c := exec.Command(exe, "dispatch", link)
	// ⚠️ SUPPRESS THE HOLD, NOT THE MESSAGE. shimp now shows a refusal in a
	// console and waits up to 25s for a keypress, which is the fix for a
	// human seeing an empty flash. Eight shapes would add over three
	// minutes to an unattended run.
	//
	// ⚠️ And it only works where the environment PROPAGATES. ShellExecuteW
	// passes the caller's environment to the launched handler, so the
	// baseline route inherits it. A BROWSER click does not: that handler is
	// a child of Chrome, which never saw this variable. So -run's chrome
	// routes still hold, and that is measured rather than assumed.
	c.Env = quietEnv()
	out, err := c.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// waitForVerdict polls until the handler has finished writing about a click.
//
// ⚠️ It waits for a TERMINAL line rather than for any line. The handler
// writes several, and reading after the first would see an arrival with no
// identity yet and report a refusal that had not happened.
func waitForVerdict(logPath string, off int64) HandlerSaw {
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		raw, err := os.ReadFile(logPath)
		if err == nil && int64(len(raw)) > off {
			body := string(raw[off:])
			// ⚠️ ANY terminal line, not just a refusal. s67 gave a
			// one-shot two more endings -- ABORTED and exited -- and
			// waiting only for REFUSED meant four shapes polled for the
			// full 20s before falling through. Measured: the corpus went
			// from 28s to 82s and still passed, which is how a harness
			// drifts without failing.
			if strings.Contains(body, "click REFUSED: ") ||
				strings.Contains(body, "click ABORTED: ") ||
				strings.Contains(body, "click: exited ") {
				return readHandler(body)
			}
		}
		time.Sleep(150 * time.Millisecond)
	}
	raw, _ := os.ReadFile(logPath)
	if int64(len(raw)) > off {
		return readHandler(string(raw[off:]))
	}
	return HandlerSaw{Verdict: NotSeen}
}

func reportCorpus(results []CorpusResult) {
	fmt.Fprintf(progress, "\n%s\n", strings.Repeat("-", 72))
	bad := 0
	for _, r := range results {
		if r.OK() {
			continue
		}
		bad++
		fmt.Fprintf(progress, "⚠️ %s  (%s)\n", r.Shape.ID, r.Shape.Label)
		for _, p := range r.Problems {
			fmt.Fprintf(progress, "   %s\n", p)
		}
	}
	// What the accepted shapes actually produced, printed whether or not
	// they matched, for the Excel-formula reason: a verdict that hides its
	// own measurement is a verdict nobody can check.
	fmt.Fprintf(progress, "\nwhat the handler derived:\n")
	for _, r := range results {
		if r.Saw.Verdict != Accepted {
			continue
		}
		fmt.Fprintf(progress, "  %-24s identity %s over %d bytes\n",
			r.Shape.ID, short(r.Saw.Identity), r.Saw.IdentityBytes)
		if r.Saw.Action != "" {
			fmt.Fprintf(progress, "  %-24s %s\n", "", r.Saw.Action)
		}
	}
	for _, r := range results {
		if r.Saw.Verdict != HandlerRefused {
			continue
		}
		fmt.Fprintf(progress, "  %-24s refused: %s\n",
			r.Shape.ID, firstLine(r.Saw.Refusal))
	}
	fmt.Fprintf(progress, "\n%d shape(s), %d finding(s).\n", len(results), bad)
}

func short(s string) string {
	if len(s) > 16 {
		return s[:16] + "..."
	}
	return s
}

// doServe renders the corpus as a page and waits for a human.
//
// ⚠️ RUN THIS IN A VISIBLE TERMINAL. NEVER BACKGROUND IT.
//
// It installs a real registry claim, serves a page for a PERSON to click,
// and purges on THEIR Ctrl-C. Backgrounding it defeats all three: measured
// 15 Oct 2026, an agent started it from a hidden process and the user could
// not see it, could not stop it, and could not tell whether it was alive.
//
// AGENTS states the rule plainly -- no hidden windows, because a process the
// user cannot see is a process they cannot diagnose. That it was broken by
// the very tool built to find invisible failures is worth recording rather
// than quietly fixing.
//
// ⚠️ Served over http on 127.0.0.1, never written as a file:// page. s64
// measured why: Chrome keys its external-protocol permission on ORIGIN, and
// a file:// page has an opaque origin that cannot hold one, so a file page
// can never be granted "always allow" and its links stay unclickable.
func doServe(port int, windowsGUI bool) int {
	root, err := repoRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	work, err := os.MkdirTemp("", "linkprobe-serve-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer os.RemoveAll(work)

	exe, err := buildAndInstall(root, work, testScheme, windowsGUI)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer purge(exe, testScheme)

	ln, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot listen on 127.0.0.1:%d: %v\n", port, err)
		return 1
	}
	// Reported AFTER the socket is bound, which is the house rule: a server
	// that announces a port before binding names one it may not have.
	addr := "http://" + ln.Addr().String() + "/"
	logPath, _ := probeLogPath()
	fmt.Fprintf(progress, "serving %s\n", addr)
	fmt.Fprintf(progress, "scheme  %s\n", testScheme)
	fmt.Fprintf(progress, "log     %s\n", logPath)
	if windowsGUI {
		fmt.Fprintln(progress, "\n⚠️ built windowsgui, so a click shows NOTHING "+
			"on screen. Watch the log.")
	}
	fmt.Fprintln(progress, "\nCtrl-C to stop and purge.")

	srv := &http.Server{Handler: http.HandlerFunc(servePage)}
	go func() { _ = srv.Serve(ln) }()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	fmt.Fprintln(progress, "\nstopping")
	_ = srv.Close()
	return 0
}

func servePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	var b strings.Builder
	b.WriteString(`<!doctype html><meta charset="utf-8">
<title>shimp:// corpus</title>
<style>
 :root{color-scheme:light dark}
 body{font:15px/1.55 system-ui,sans-serif;max-width:54rem;margin:2.5rem auto;padding:0 1.2rem}
 h1{font-size:1.4rem;margin-bottom:.2rem}
 .sub{opacity:.7;margin-bottom:1.6rem}
 .warn{border-left:3px solid #d08;padding:.6rem 0 .6rem .9rem;margin:1.4rem 0}
 .row{border-top:1px solid rgba(128,128,128,.3);padding:1rem 0}
 .label{font-weight:600}
 .exp{font-size:12px;opacity:.65;font-family:ui-monospace,monospace}
 a.try{display:inline-block;margin:.4rem 0;font-family:ui-monospace,monospace;
       font-size:12px;word-break:break-all}
 .note{opacity:.75;font-size:13.5px}
 code{font-family:ui-monospace,monospace;font-size:12.5px}
</style>
<h1>shimp:// &mdash; the corpus, clickable</h1>
`)
	fmt.Fprintf(&b, `<div class="sub">scheme <code>%s</code>. These are the same
shapes <code>-corpus</code> dispatches, read from one table so the page and the
check cannot drift.</div>`, testScheme)
	b.WriteString(`<div class="warn"><strong>Every valid link ends in a
refusal</strong>, because the keeper is not built. The handler logs the identity
and the argv it derived, then declines. Two of these should be refused by the
<em>parser</em> instead, with a reason naming the rule.<br><br>
Chrome asks the first time. Ticking <strong>always allow</strong> grants exactly
the permission s64 measured, after which a page can fire these with no click at
all.</div>`)
	for _, s := range shapes {
		link := s.Link(testScheme)
		fmt.Fprintf(&b, `<div class="row"><div class="label">%s</div>
<div class="exp">expects %s`, s.Label, s.Expect)
		if s.Reason != "" {
			fmt.Fprintf(&b, ", reason containing %q", s.Reason)
		}
		if s.Argv != nil {
			fmt.Fprintf(&b, ", identity over %d bytes",
				s.ExpectedIdentityBytes(testScheme))
		}
		fmt.Fprintf(&b, `</div>
<a class="try" href="%s">%s</a>
<div class="note">%d bytes. %s</div></div>`, link, link, len(link), s.Because)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(b.String()))
}
