// linkprobe -- trigger one shimp:// link down every way in, and report what
// each route did to it.
//
//	go run . -list
//	go run . -run
//	go run . -run -only excel-hyperlink-object
//
// ⚠️ WHAT THIS IS FOR, in one sentence: s44 makes a link's identity the
// sha256 of its bytes verbatim, so a route that changes one byte on the way
// in does not fail cleanly, and until this program existed only ONE of the
// fourteen ways in had ever been run.
//
// ⚠️ AND WHY IT IS A SEPARATE MODULE. Same reason as repocheck: it measures
// shimp from outside, it installs and uninstalls a real registry claim, and
// it shells out to Python for the Excel routes. None of that belongs inside
// the thing under test. See s48 for the isolated-installation rule it rests
// on, which is what makes running this safe at all.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// testScheme is NOT "shimp", and the check below is by name rather than by
// comment. s48's whole point: an installer exercised against production
// leaves a machine that needs hand-repair, and the live install test found
// exactly that happening before the scheme was threaded through every verb.
const testScheme = "shimp-linkprobe"

const logName = "shimp.log"

// arrivalPrefix is what the handler writes for a click. ⚠️ It logs the raw
// URI BEFORE parsing it, which is what makes a mangled arrival visible at
// all: a link that the parser refuses still appears here, verbatim, so the
// probe can see what was actually delivered rather than only whether it
// was accepted.
const arrivalPrefix = "click: "

// exitCannotComplete is what a driver returns when it measured what it
// could and the remainder genuinely needs a human. See excel_route.py.
const exitCannotComplete = 3

// progress is where the human-readable report goes.
//
// ⚠️ It becomes STDERR under -json, so stdout carries exactly one
// JSON object and nothing else. A consumer piping stdout into a parser
// must not have to strip progress lines out of it, and a harness that
// mixed the two would be unusable from a test for a reason that looks
// like a parse bug.
var progress io.Writer = os.Stdout

// Result is one route's run.
type Result struct {
	Route    Route
	Got      Outcome
	Sent     string
	Received string

	// Detail carries the route's own message, which for a Refused outcome is
	// the interesting part: "Protocol \"shimp-linkprobe\" not supported" is a
	// correct and informative refusal, and worth printing rather than
	// flattening to a boolean.
	Detail string
}

// AsExpected reports whether the route did what the table said it would.
func (r Result) AsExpected() bool { return r.Got == r.Route.Expect }

// jsonResult is the machine-readable shape of one route's run.
//
// ⚠️ A SEPARATE TYPE, not json tags on Result. Result embeds the whole Route
// including its prose and its Because, and a consumer asserting on this must
// not break when somebody rewords a comment. This is the published surface
// and it is deliberately narrow.
type jsonResult struct {
	ID         string `json:"id"`
	Kind       string `json:"kind"`
	Expect     string `json:"expect"`
	Got        string `json:"got"`
	AsExpected bool   `json:"as_expected"`
	SentBytes  int    `json:"sent_bytes"`
	Sent       string `json:"sent"`
	Received   string `json:"received,omitempty"`
	Detail     string `json:"detail,omitempty"`
}

// jsonReport is what -json writes to stdout.
type jsonReport struct {
	Scheme   string       `json:"scheme"`
	Routes   int          `json:"routes"`
	Run      int          `json:"run"`
	Findings int          `json:"findings"`
	Manual   int          `json:"manual"`
	Results  []jsonResult `json:"results"`
}

func main() {
	list := flag.Bool("list", false, "print the route table and exit")
	run := flag.Bool("run", false, "install an isolated scheme, run the routes, purge")
	only := flag.String("only", "", "run a single route by id")
	keep := flag.Bool("keep", false,
		"do NOT purge afterwards, for inspecting the log by hand. ⚠️ leaves a "+
			"real registry claim in place")
	asJSON := flag.Bool("json", false,
		"write one JSON object to stdout instead of the human report, so a "+
			"test suite can assert on the result. Progress goes to stderr")
	flag.Parse()

	switch {
	case *list:
		printTable()
	case *run:
		if code := doRun(*only, *keep, *asJSON); code != 0 {
			os.Exit(code)
		}
	default:
		flag.Usage()
		fmt.Fprintf(os.Stderr, "\n%d routes. -list to read them, -run to run them.\n",
			len(routes))
		os.Exit(2)
	}
}

func printTable() {
	fmt.Fprintf(progress, "%d ways into a shimp:// link.\n\n", len(routes))
	for _, r := range routes {
		fmt.Fprintf(progress, "%-26s %-6s expect %-10s %s\n", r.ID, r.Kind, r.Expect, r.What)
		if r.Expect != Verbatim {
			fmt.Fprintf(progress, "%-26s        because %s\n", "", r.Because)
		}
	}
}

// doRun is the whole cycle: build, install, probe each route, report, purge.
func doRun(only string, keep bool, asJSON bool) int {
	if asJSON {
		progress = os.Stderr
	}

	if testScheme == "shimp" {
		fmt.Fprintln(os.Stderr,
			"REFUSED: testScheme is the production scheme. This program "+
				"installs and purges a real registry claim.")
		return 1
	}

	root, err := repoRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	work, err := os.MkdirTemp("", "linkprobe-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer os.RemoveAll(work)

	// -- build the real CLI. ⚠️ NOT this program, and not a test binary: the
	// live install test's own history is the reason. It once installed the
	// `go test` binary as the handler, which "worked" and measured nothing,
	// because a test binary handed a URI argument does not run the handler.
	shimpExe := filepath.Join(work, "shimp.exe")
	cmd := exec.Command("go", "build", "-o", shimpExe, ".")
	cmd.Dir = filepath.Join(root, "shimp")
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "building shimp failed: %v\n%s", err, out)
		return 1
	}

	// -- install the isolated scheme
	if out, err := exec.Command(shimpExe, "install", "-scheme", testScheme).CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "install -scheme %s failed: %v\n%s",
			testScheme, err, out)
		return 1
	}
	if !keep {
		defer func() {
			out, err := exec.Command(shimpExe, "uninstall", "-purge",
				"-scheme", testScheme).CombinedOutput()
			if err != nil {
				fmt.Fprintf(os.Stderr,
					"\n⚠️ PURGE FAILED, so this machine still carries a %s "+
						"registry claim: %v\n%s\n", testScheme, err, out)
				return
			}
			fmt.Fprintf(progress, "\npurged %s\n", testScheme)
		}()
	}

	logPath, err := probeLogPath()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	python := findPython(root)
	page := filepath.Join(work, "click_me.html")

	var results []Result
	for _, r := range routes {
		if only != "" && r.ID != only {
			continue
		}
		fmt.Fprintf(progress, "  %-26s ", r.ID)
		res := probe(r, shimpExe, python, logPath, root)
		results = append(results, res)
		verdict := "as expected"
		if !res.AsExpected() {
			verdict = "⚠️ FINDING"
		}
		fmt.Fprintf(progress, "%-10s %s\n", res.Got, verdict)
	}

	if err := writeClickPage(page, linkFor("manual-click")); err == nil {
		fmt.Fprintf(progress, "\nfor the human routes: %s\n", page)
	}

	report(results, logPath, keep)

	// ⚠️ EMITTED AFTER the human report, and on stdout while that went to
	// stderr. Both are produced, always: a consumer gets the object and a
	// human watching the same run still sees why.
	if asJSON {
		if err := emitJSON(os.Stdout, results); err != nil {
			fmt.Fprintf(os.Stderr, "writing JSON failed: %v\n", err)
			return 1
		}
	}
	return 0
}

// emitJSON writes the machine-readable report.
//
// ⚠️ EXIT CODE 0 EVEN WITH FINDINGS, deliberately, and this is the decision
// a consumer most needs to know. A finding is a measurement, not a crash:
// `explorer` is EXPECTED to drop the link and `chrome-no-gesture` is
// expected to deliver it, so "did the probe run" and "did the routes behave"
// are different questions. Non-zero is reserved for the probe failing to
// take a measurement at all -- no checkout, build failure, install refused.
//
// So a test suite asserts on `findings` or on a named route's `got`, never on
// the process exit code.
func emitJSON(w io.Writer, results []Result) error {
	rep := jsonReport{Scheme: testScheme, Routes: len(routes)}
	for _, r := range results {
		if r.Got == Manual {
			rep.Manual++
		} else {
			rep.Run++
			if !r.AsExpected() {
				rep.Findings++
			}
		}
		rep.Results = append(rep.Results, jsonResult{
			ID:         r.Route.ID,
			Kind:       string(r.Route.Kind),
			Expect:     string(r.Route.Expect),
			Got:        string(r.Got),
			AsExpected: r.AsExpected(),
			SentBytes:  len(r.Sent),
			Sent:       r.Sent,
			Received:   r.Received,
			Detail:     r.Detail,
		})
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(rep)
}

// probe runs one route and classifies what happened.
//
// ⚠️ ARRIVAL IS DECIDED BY THE LOG, NEVER BY AN EXIT CODE. explorer.exe exits
// non-zero on success; `cmd /c start` exits zero whether or not it opened
// anything. An exit code here reports the launcher's opinion of itself, which
// is not the question.
func probe(r Route, shimpExe, python, logPath, root string) Result {
	res := Result{Route: r, Sent: linkForPadded(r.ID, r.Pad)}

	if r.Kind == NeedsHuman {
		res.Got = Manual
		res.Detail = r.Because
		return res
	}

	// Where the log is NOW, so that only lines written after this point are
	// considered. A route that mangles the nonce beyond recognition is still
	// detectable, because its arrival is new even if it is unrecognisable.
	before := logSize(logPath)

	var c *exec.Cmd
	if r.Kind == NeedsExcel {
		if python == "" {
			res.Got = Unrunnable
			res.Detail = "no Python interpreter found, and the Excel routes are driven through COM"
			return res
		}
		// ⚠️ A TIMEOUT, because this is the one route that can block
		// forever. Following a hyperlink in Office can raise a modal: s8
		// records that the trusted-protocol key may not suppress HLINK.dll's
		// SECOND dialog, and a COM call waiting behind a modal never
		// returns. Without this the whole probe hangs on a dialog instead of
		// reporting that one route could not be answered.
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()
		c = exec.CommandContext(ctx, python,
			filepath.Join(root, "linkprobe", "excel_route.py"),
			r.Argv[0], res.Sent)
	} else if r.Kind == NeedsChrome {
		if python == "" {
			res.Got = Unrunnable
			res.Detail = "no Python interpreter found, and the Chrome routes " +
				"need one for the websocket CDP needs"
			return res
		}
		// Longer than Excel's, because this one launches a browser, waits
		// for a debug port and serves pages. Still bounded: a hung Chrome
		// must not hang the matrix.
		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Second)
		defer cancel()
		c = exec.CommandContext(ctx, python,
			filepath.Join(root, "linkprobe", "chrome_route.py"),
			r.Argv[0], res.Sent)
	} else {
		argv := r.Fill(map[string]string{
			"{{uri}}":    res.Sent,
			"{{shimp}}":  shimpExe,
			"{{python}}": python,
		})
		c = exec.Command(argv[0], argv[1:]...)
	}

	out, err := c.CombinedOutput()
	res.Detail = strings.TrimSpace(string(out))
	if err != nil && errors.Is(err, exec.ErrNotFound) {
		res.Got = Unrunnable
		res.Detail = fmt.Sprintf("%s is not on this machine", c.Path)
		return res
	}
	// Exit 3 is the drivers' "I measured what I could and the rest needs a
	// human". ⚠️ Distinct from a refusal and from silence: the Excel formula
	// route reaches it having successfully written the formula and read back
	// what Excel stored, and only the DELIVERY half is unanswerable.
	if c.ProcessState != nil && c.ProcessState.ExitCode() == exitCannotComplete {
		res.Got = Unrunnable
		return res
	}

	// Poll for arrival. Polled rather than slept, for the live test's reason:
	// a fixed sleep is either flaky or slow and usually both.
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		if got, ok := arrivalSince(logPath, before); ok {
			res.Received = got
			if got == res.Sent {
				res.Got = Verbatim
			} else {
				res.Got = Mangled
			}
			return res
		}
		time.Sleep(200 * time.Millisecond)
	}

	// Nothing arrived. Refused and Absent are different facts: a route that
	// said why it would not try is behaving correctly for its kind, and a
	// route that dropped the link in silence is the failure shape this whole
	// repo is organised against.
	//
	// ⚠️ A REFUSAL NEEDS A MESSAGE, not just a non-zero exit. The first
	// version of this classifier keyed on the exit code alone, and explorer
	// exits 1 whether it dispatched the URI or silently dropped it. So
	// explorer's silent drop came out labelled REFUSED, which reads as the
	// route behaving honestly. It had said nothing at all. The comment on
	// this function already warned against trusting exit codes, and the code
	// below it did exactly that.
	if err != nil && res.Detail != "" {
		res.Got = Refused
		return res
	}
	res.Got = Absent
	return res
}

// arrivalSince returns the raw URI from the first click logged after off.
func arrivalSince(logPath string, off int64) (string, bool) {
	raw, err := os.ReadFile(logPath)
	if err != nil || int64(len(raw)) <= off {
		return "", false
	}
	for _, line := range strings.Split(string(raw[off:]), "\n") {
		i := strings.Index(line, arrivalPrefix)
		if i < 0 {
			continue
		}
		got := strings.TrimRight(line[i+len(arrivalPrefix):], "\r")
		// The handler logs several click: lines. Only one of them is the URI
		// itself, and it is the one that looks like a URI.
		if strings.Contains(got, "://") {
			return got, true
		}
	}
	return "", false
}

func logSize(p string) int64 {
	st, err := os.Stat(p)
	if err != nil {
		return 0
	}
	return st.Size()
}

func probeLogPath() (string, error) {
	h, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot locate the user home, so the probe "+
			"cannot find the handler's log: %w", err)
	}
	return filepath.Join(h, "."+testScheme, logName), nil
}

// linkFor builds the probe link for a route.
//
// ⚠️ THE ROUTE ID GOES IN ARGV, NOT THE FRAGMENT. Two reasons, and the second
// is the one that matters: argv is part of the identity (s44), so each route
// gets a distinct identity and a distinct singleton rather than colliding;
// and a route that STRIPS the fragment would otherwise look like a route that
// delivered nothing. The fragment carries a copy, so fragment loss shows up
// as a mangle rather than as silence.
func linkFor(id string) string { return linkForPadded(id, 0) }

// linkForPadded lengthens the link to at least pad bytes, for routes with a
// length threshold worth crossing. The padding goes in ARGV, so it is inside
// the identity-bearing part and a route that truncates changes the identity.
func linkForPadded(id string, pad int) string {
	const sha = "0123456789abcdef0123456789abcdef01234567"
	head := testScheme + "://v1/go/run/github.com/davidnoz123/" +
		"nielsoln_connectors/probe@" + sha + "/-route/" + id
	tail := "#open?route=" + id
	if n := pad - len(head) - len(tail); n > 0 {
		head += "/-pad/" + strings.Repeat("p", n)
	}
	return head + tail
}

func report(results []Result, logPath string, keep bool) {
	fmt.Fprintf(progress, "\n%s\n", strings.Repeat("-", 72))
	var findings, manual int
	for _, r := range results {
		if r.Got == Manual {
			manual++
			continue
		}
		if !r.AsExpected() {
			findings++
		}
	}

	// ⚠️ THE BASELINE GATE. If ShellExecuteW itself did not deliver, every
	// other row is measuring a broken installation rather than a route, and
	// saying so once is worth more than thirteen false findings.
	for _, r := range results {
		if r.Route.ID == "shellexecute" && r.Got != Verbatim {
			fmt.Fprintf(progress, "⚠️ THE BASELINE FAILED: shellexecute reported %s.\n"+
				"   Every other row below is measuring a broken installation, "+
				"not a route.\n   Fix this first and ignore the rest.\n\n", r.Got)
		}
	}

	for _, r := range results {
		if r.AsExpected() {
			continue
		}
		fmt.Fprintf(progress, "⚠️ %s: expected %s, got %s\n", r.Route.ID, r.Route.Expect, r.Got)
		fmt.Fprintf(progress, "   %s\n", r.Route.What)
		if r.Got == Mangled {
			fmt.Fprintf(progress, "   sent     %d bytes: %s\n", len(r.Sent), r.Sent)
			fmt.Fprintf(progress, "   received %d bytes: %s\n", len(r.Received), r.Received)
			fmt.Fprintf(progress, "   ⚠️ the identity of the received link is NOT the "+
				"identity of the sent one, so this route cannot deliver a "+
				"shimp link at all.\n")
		}
		if r.Detail != "" {
			fmt.Fprintf(progress, "   said: %s\n", firstLine(r.Detail))
		}
	}

	// ⚠️ WHAT EACH NON-VERBATIM ROUTE ACTUALLY SAID, printed whether or not
	// it matched expectation. The first version printed this only for
	// findings, which hid the measurement behind the verdict: the Excel
	// formula route came out "UNRUNNABLE as expected" and silently discarded
	// the one number it had gone to Excel to collect.
	var spoke bool
	for _, r := range results {
		if r.Got == Verbatim || r.Got == Manual || strings.TrimSpace(r.Detail) == "" {
			continue
		}
		if !spoke {
			fmt.Fprintf(progress, "\nwhat the non-delivering routes reported:\n")
			spoke = true
		}
		fmt.Fprintf(progress, "  %s (%s), sent %d bytes\n", r.Route.ID, r.Got, len(r.Sent))
		for _, line := range strings.Split(r.Detail, "\n") {
			if line = strings.TrimSpace(line); line != "" {
				fmt.Fprintf(progress, "    %s\n", line)
			}
		}
	}

	fmt.Fprintf(progress, "\n%d run, %d finding(s), %d awaiting a human.\n",
		len(results)-manual, findings, manual)
	if manual > 0 {
		fmt.Fprintln(progress, "⚠️ The manual routes are UNANSWERED, not passed. "+
			"Named below so the matrix has no invisible holes:")
		for _, r := range results {
			if r.Got == Manual {
				fmt.Fprintf(progress, "   %-26s %s\n", r.Route.ID, r.Route.What)
			}
		}
	}
	if keep {
		fmt.Fprintf(progress, "\n-keep: the installation is STILL IN PLACE. Log: %s\n", logPath)
		fmt.Fprintf(progress, "   purge it with: shimp uninstall -purge -scheme %s\n", testScheme)
	}
}

func firstLine(s string) string {
	if i := strings.IndexAny(s, "\r\n"); i >= 0 {
		return s[:i] + " ..."
	}
	return s
}

// writeClickPage writes the page the human routes need.
//
// A real <a href> on a real page, because s36's open half is what a browser
// does to a URI BEFORE handing it over, and that cannot be measured by
// handing the URI to the browser on a command line.
func writeClickPage(path, uri string) error {
	body := "<!doctype html>\n<meta charset=\"utf-8\">\n" +
		"<title>linkprobe</title>\n" +
		"<p>Click this. The handler logs what it received, verbatim.</p>\n" +
		"<p><a href=\"" + uri + "\">" + uri + "</a></p>\n" +
		"<p>Then compare against the sent bytes above.</p>\n"
	return os.WriteFile(path, []byte(body), 0o644)
}

// repoRoot walks up until it finds the directory holding shimp/.
func repoRoot() (string, error) {
	d, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if st, err := os.Stat(filepath.Join(d, "shimp", "main.go")); err == nil && !st.IsDir() {
			return d, nil
		}
		p := filepath.Dir(d)
		if p == d {
			return "", fmt.Errorf("no shimp/main.go in any parent of the " +
				"working directory, so the probe cannot find the handler to " +
				"build. ⚠️ A module fetched by `go run module@sha` has no " +
				"siblings, which is the expected way for this to fail")
		}
		d = p
	}
}

// findPython locates an interpreter for the Python-driven routes.
//
// ⚠️ It SEARCHES and reports nothing found rather than guessing, which is
// owner_other.go's rule applied to an interpreter. locals.txt is the only
// permitted home for an absolute path, so the venv is read from there and
// never from a literal here.
func findPython(root string) string {
	if p := fromLocals(root, "VENV_PYTHON"); p != "" {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	for _, name := range []string{"python.exe", "python3.exe", "python"} {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}
	return ""
}

// fromLocals reads one key out of locals.txt, which is gitignored and is the
// only place in this workspace allowed to hold an absolute path.
func fromLocals(root, key string) string {
	raw, err := os.ReadFile(filepath.Join(root, "locals.txt"))
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if ok && strings.TrimSpace(k) == key {
			return strings.Trim(strings.TrimSpace(v), "\"")
		}
	}
	return ""
}
