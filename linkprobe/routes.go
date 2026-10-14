// routes.go -- every way a shimp:// link can be triggered, as DATA.
//
// ⚠️ WHY A TABLE AND NOT A TEST EACH. s44 fixes identity as the sha256 of the
// URI's bytes, verbatim, from `shimp://` to the byte before the first `#`.
// s35 then says refuse, never normalise. Put together, those two decisions
// mean that ANY route which alters one byte on the way in produces a
// different identity, and the handler is required to refuse it.
//
// So every entry point is a potential mangler, and until this table existed
// exactly ONE of them had ever been measured: ShellExecuteW, by the live
// install test. Every other way a human would actually trigger a link -- a
// cell in a spreadsheet, a link in an email, the address bar -- was reasoned
// about and never run.
//
// The table is data so that adding a route is adding a row, and so that the
// EXPECTATION is declared next to the route rather than living in a person's
// head. A route whose outcome matches its declared expectation is not news. A
// route that does something else is the whole point of running this.
package main

import "strings"

// Outcome is what a route did to a link. Five values, deliberately, because
// collapsing them loses the distinction the probe exists to draw.
type Outcome string

const (
	// Verbatim: the handler received byte-for-byte what was sent. The only
	// outcome on which a link's identity survives the trip.
	Verbatim Outcome = "VERBATIM"

	// Mangled: the handler received something, and it was NOT what was sent.
	// ⚠️ THE DANGEROUS ONE, and the reason for the whole exercise. A mangled
	// link does not fail cleanly: it hashes to a different identity, so it
	// either refuses a link the user can see is correct, or -- far worse --
	// runs a DIFFERENT pin than the one named in the text they pasted.
	Mangled Outcome = "MANGLED"

	// Absent: the route reported success and nothing ever arrived. The silent
	// failure this repo is organised against, in link form.
	Absent Outcome = "ABSENT"

	// Refused: the route itself declined to try, with a message. An honest
	// outcome, and for an HTTP client it is the CORRECT one.
	Refused Outcome = "REFUSED"

	// Manual: a human has to click something, so an unattended run cannot
	// answer it. Reported by name rather than skipped, because a matrix with
	// invisible holes reads as complete.
	Manual Outcome = "MANUAL"

	// Unrunnable: the route's program is not on this machine. Distinct from
	// Refused, because "wget says no" and "wget is absent" are different
	// facts and only one of them is about shimp.
	Unrunnable Outcome = "UNRUNNABLE"
)

// Kind separates what can run unattended from what cannot.
type Kind string

const (
	// Auto: argv, exec, done.
	Auto Kind = "auto"

	// NeedsExcel: driven through COM by excel_route.py, because Excel is the
	// one route with two genuinely different mechanisms behind one word and
	// Go has no business driving COM here.
	NeedsExcel Kind = "excel"

	// NeedsChrome: driven by chrome_route.py, which serves a page over http
	// and clicks the link through CDP.
	//
	// ⚠️ Python rather than Go for one specific reason: CDP needs a
	// websocket, Go's stdlib has none, and this repo takes no third-party
	// dependencies. remote_chrome_cdp hand-rolled one, and ITS OWN HEADER
	// records why it and remote_ai's are deliberately different -- so a
	// shared leaf would overturn a decision rather than tidy an oversight.
	// The venv already carries websocket-client, and the Excel routes
	// already established the shell-out-to-Python pattern here.
	NeedsChrome Kind = "chrome"

	// NeedsHuman: a consent dialog, a keystroke, or an application nobody
	// should be automating for a measurement this small.
	NeedsHuman Kind = "human"
)

// Route is one way in.
type Route struct {
	ID string

	// What it is, in prose, for the report. A matrix of pass/fail with no
	// prose is unreadable six months later.
	What string

	Kind Kind

	// Argv with {{uri}}, {{shimp}} and {{python}} substituted. NOT
	// shell-quoted by us: where a shell is involved its quoting is part of
	// the template, because the shell's re-parsing is precisely what is
	// under measurement.
	Argv []string

	// Expect is what this route is BELIEVED to do. A run that matches is
	// routine; a run that does not is the finding.
	Expect Outcome

	// Because is why the expectation is what it is. ⚠️ Required for anything
	// expected to be other than Verbatim, since "this one is supposed to
	// fail" is a claim that rots silently.
	Because string

	// Pad lengthens this route's probe link to at least this many bytes.
	//
	// ⚠️ Needed because a threshold nobody crosses is a threshold nobody
	// measures. The default probe link is 183 bytes, and the Excel formula
	// route truncates at 255, so the first run of that route could not have
	// seen truncation even with a correct driver. Zero means the default
	// length.
	Pad int
}

// Fill substitutes into a route's argv.
func (r Route) Fill(repl map[string]string) []string {
	out := make([]string, len(r.Argv))
	for i, a := range r.Argv {
		for k, v := range repl {
			a = strings.ReplaceAll(a, k, v)
		}
		out[i] = a
	}
	return out
}

// routes is the matrix. Order is roughly "closest to the OS" first, because a
// failure near the top invalidates everything below it.
var routes = []Route{
	{
		ID:   "shellexecute",
		What: "shell32!ShellExecuteW via `shimp dispatch`. The OS hop with nothing in front of it",
		Kind: Auto,
		// ⚠️ THE BASELINE. Every other row is a comparison against this one,
		// so if this row is not VERBATIM then nothing else in the table means
		// anything, and the run says exactly that rather than reporting
		// thirteen independent failures.
		Argv:    []string{"{{shimp}}", "dispatch", "{{uri}}"},
		Expect:  Verbatim,
		Because: "the registry hop is the only thing between send and receive, and the live install test already measured it",
	},
	{
		ID:   "cmd-start",
		What: "`cmd /c start \"\" \"<uri>\"`, the oldest documented way to open a URI",
		Kind: Auto,
		// ⚠️ The empty "" is the window TITLE, not a stray quote. Without it
		// `start` treats a quoted URI as the title and opens nothing, which
		// is a silent no-op and exactly what this table exists to catch.
		Argv:    []string{"cmd", "/c", "start", "", "{{uri}}"},
		Expect:  Verbatim,
		Because: "believed to pass the URI through, NOT measured. cmd re-parses, and `&` is a command separator there, so a link with two action arguments is the interesting case",
	},
	{
		ID:      "powershell-start-process",
		What:    "PowerShell `Start-Process -FilePath '<uri>'`",
		Kind:    Auto,
		Argv:    []string{"powershell", "-NoProfile", "-Command", "Start-Process", "-FilePath", "'{{uri}}'"},
		Expect:  Verbatim,
		Because: "⚠️ KNOWN to impose its OWN ceiling: the 16000-byte failure in s36's length sweep was PowerShell's \"command line is too long\", not Windows'. Fine for short links, wrong for a length measurement",
	},
	{
		ID:   "explorer",
		What: "`explorer.exe \"<uri>\"`, the shell itself",
		Kind: Auto,
		Argv: []string{"explorer.exe", "{{uri}}"},
		// ⚠️ MEASURED 15 Oct 2026, AND IT CANNOT CARRY A SHIMP LINK.
		//
		// Expected ABSENT rather than Verbatim, so that a future run flags a
		// CHANGE instead of re-finding this. Two separate defects, both
		// measured, both deterministic over repeats:
		//
		//  1. A URI containing `?` is DROPPED IN SILENCE. Exit code 1 and no
		//     message, in the path or the fragment alike:
		//       ...#open          VERBATIM
		//       ...#open?route=x  nothing
		//       ...//x?q=1        nothing
		//     shimp's action syntax is `#open?file=...&line=...`, so explorer
		//     cannot deliver ANY link that carries action arguments. s61's
		//     editor_open link is exactly this shape.
		//
		//  2. It NORMALISES an empty path, inserting a slash:
		//       sent     shimp-linkprobe://x#y
		//       received shimp-linkprobe://x/#y
		//     The inserted byte is before the `#`, so it is inside the
		//     identity-bearing part and the hash changes. s35's "refuse,
		//     never normalise" hazard, observed in the wild rather than
		//     argued about. Real links always have a non-empty path, so this
		//     one does not bite today, but it proves explorer rewrites URIs.
		//
		// And it does dispatch schemes in general: ms-settings: and vscode:
		// both worked, vscode being HKCU-registered exactly as shimp is. So
		// this is not about our registration.
		Expect:  Absent,
		Because: "measured: explorer silently drops any URI containing `?`, and shimp's action syntax requires one. Exit code 1 with no message, which is why a refusal here needs a message to count as one",
	},
	{
		ID:      "python-startfile",
		What:    "Python `os.startfile(uri)`, which is ShellExecuteW under another name",
		Kind:    Auto,
		Argv:    []string{"{{python}}", "-c", "import os,sys; os.startfile(sys.argv[1])", "{{uri}}"},
		Expect:  Verbatim,
		Because: "documented as wrapping ShellExecuteW, so it should match the baseline exactly. Worth running because 'should' is doing a lot of work in that sentence",
	},
	{
		ID:      "python-webbrowser",
		What:    "Python `webbrowser.open(uri)`, the route a script reaches for by habit",
		Kind:    Auto,
		Argv:    []string{"{{python}}", "-c", "import webbrowser,sys; webbrowser.open(sys.argv[1])", "{{uri}}"},
		Expect:  Verbatim,
		Because: "⚠️ may route through a BROWSER rather than the shell, in which case it inherits the browser's consent dialog and is a human route wearing an automatable costume",
	},
	{
		ID:   "curl",
		What: "`curl <uri>`, an HTTP client handed a scheme it does not know",
		Kind: Auto,
		// ⚠️ EXPECTED TO REFUSE, and that refusal is a PASS.
		Argv:    []string{"curl", "-sS", "{{uri}}"},
		Expect:  Refused,
		Because: "curl speaks the protocols it implements and never consults the registry. It CANNOT trigger a shimp link, and pinning that stops someone assuming a URL is a URL",
	},
	{
		ID:      "wget",
		What:    "`wget <uri>`, the same question asked of a different client",
		Kind:    Auto,
		Argv:    []string{"wget", "-q", "-O", "-", "{{uri}}"},
		Expect:  Unrunnable,
		Because: "same reasoning as curl, but ⚠️ wget is ABSENT from this machine, so the honest expectation is that it cannot be run rather than that it refused. Two different facts",
	},
	{
		ID:      "excel-hyperlink-object",
		What:    "Excel `Hyperlinks.Add` then `.Follow` -- the route a slate cell actually uses",
		Kind:    NeedsExcel,
		Argv:    []string{"object"},
		Expect:  Verbatim,
		Because: "⚠️ THE MOST IMPORTANT ROW IN THE TABLE, because this is how a link reaches a human in this workspace, and it has never been run. Separately measured: the object caps at 2084 bytes and REFUSES at 2085, which is honest behaviour",
	},
	{
		ID:   "excel-hyperlink-formula",
		What: "Excel `=HYPERLINK(\"<uri>\")`, the obvious way, which is a trap",
		Kind: NeedsExcel,
		// ⚠️ EXPECTED TO MANGLE. Declared, so that a VERBATIM result here
		// would itself be the surprise.
		Argv: []string{"formula"},
		// ⚠️ 320 bytes, deliberately past the 255 threshold. The first run of
		// this route used the default 183-byte link, so truncation could not
		// have appeared even if the driver had been correct. It was not: it
		// followed the URI variable rather than the cell's stored target and
		// reported VERBATIM for a link this route had never touched.
		Pad: 320,
		// ⚠️ UNRUNNABLE, which is the honest verdict and was a surprise.
		// A formula HYPERLINK creates no Hyperlink OBJECT, so there is
		// nothing to .Follow() and Hyperlinks.Count is 0. Excel resolves the
		// target only when a human clicks the cell, so the DELIVERY half of
		// this route cannot be answered unattended at all. What the driver
		// can still measure, and does, is what Excel STORED.
		// ⚠️ AND THE MEASUREMENT DID NOT REPRODUCE THE 255 TRUNCATION.
		// A 326-byte URI written into =HYPERLINK() came back from
		// cell.Formula at 326 bytes, unaltered, with Hyperlinks.Count 0.
		// So whatever truncates at 255 is NOT what the formula stores. It
		// may be the Hyperlink.Address property, or what a click resolves,
		// or the earlier number may have measured a different route
		// altogether. Recorded as open rather than quietly dropped: the
		// claim is not disproved, it is unlocated.
		Expect:  Unrunnable,
		Because: "exposes NO Hyperlink object (Hyperlinks.Count is 0), so delivery is only answerable by a human clicking the cell. Also a formula cell reads back as its result, so a pull/publish cycle destroys the link. ⚠️ The 255-byte truncation is NOT where it was believed to be: 326 bytes stored unaltered",
	},
	{
		ID:   "chrome-link-click",
		What: "clicking an <a href> on an http-served page, through CDP",
		Kind: NeedsChrome,
		Argv: []string{"click"},
		// ⚠️ THIS ROUTE WORKS, AND AN EARLIER VERSION OF THIS TABLE SAID IT
		// COULD NOT. The correction is worth more than the result.
		//
		// The first attempt used a file:// page, measured nothing, and
		// concluded that Chrome blocks an external-protocol launch from
		// CDP-synthesized input in general. Two things were wrong with it:
		//
		//  1. ⚠️ file:// IS AN OPAQUE ORIGIN. Chrome keys its
		//     external-protocol permission on origin, so a file:// page
		//     cannot hold one and that test could never have succeeded. The
		//     conclusion was generalised from the one origin incapable of
		//     passing.
		//  2. ⚠️ THE CONTROL WAS UNRELIABLE. Hit-testing an inactive page
		//     silently finds nothing, which is indistinguishable from a
		//     refused click. Page.bringToFront alone was not enough;
		//     Page.setWebLifecycleState active and
		//     Emulation.setFocusEmulationEnabled were both needed. Until
		//     they were, an ordinary link did not navigate either, and a
		//     null result from a harness whose control fails says nothing.
		//
		// Measured properly, both conditions are required and a CDP click
		// DOES count as a user gesture:
		//
		//	origin  permission   gesture       handler
		//	file    cannot hold  click         silent
		//	http    none         click         navigation requested, REFUSED
		//	http    granted      click         FIRES, verbatim
		//	http    granted      meta refresh  silent
		Expect:  Verbatim,
		Because: "",
	},
	{
		ID:   "chrome-no-gesture",
		What: "an http page trying to fire the link with NO gesture, by meta refresh",
		Kind: NeedsChrome,
		Argv: []string{"no-gesture"},
		// ⚠️⚠️ IT FIRES. A PAGE CAN LAUNCH A SHIMP LINK WITH NO GESTURE.
		//
		// Measured 15 Oct 2026, reproducible 3 passes out of 3: a meta
		// refresh, no JavaScript and no click of any kind, fires the handler
		// and delivers the URI verbatim, once the protocol is permitted for
		// the page's origin.
		//
		// ⚠️ THIS KILLS A CLAIM MADE TWICE IN THIS REPO. s62 and s63 both
		// recorded that "a web page cannot silently fire a shimp link" and
		// treated it as a security property in shimp's favour. It is not
		// true. Taken with chrome-control-no-permission below, the two rows
		// settle the mechanism between them:
		//
		//	click,     NO permission  ->  ABSENT   (so permission is NECESSARY)
		//	NO gesture, permission    ->  FIRES    (so gesture is NOT required)
		//
		// So the PER-ORIGIN PERMISSION IS THE SOLE GATE. A user grants it
		// once, through Chrome's "always allow" checkbox, and from then on
		// that origin can fire shimp links on any later visit with no
		// interaction at all.
		//
		// What follows for shimp, and it is a design consequence rather than
		// a bug to fix here: the handler must treat every arriving link as
		// untrusted and must never act on one merely because it arrived. The
		// comfort of "somebody must have clicked" is not available.
		//
		// Expect is Verbatim because that is what happens, so the row flags
		// a CHANGE. ⚠️ An ABSENT here would be good news, not a failure.
		Expect:  Verbatim,
		Because: "",
	},
	{
		ID:   "chrome-control-no-permission",
		What: "the same click with the protocol NOT permitted for the origin",
		Kind: NeedsChrome,
		Argv: []string{"control"},
		// ⚠️ THE CONTROL, kept in the table rather than done once by hand.
		//
		// Two things it establishes, and the harness is worthless without
		// the first: that the click LANDS, by navigating an ordinary link on
		// the same origin; and that without the per-origin permission the
		// shimp launch is refused even though the click landed.
		//
		// chrome_route.py reports the navigation events Chrome emitted, so
		// "nothing arrived" is never ambiguous between a click that missed
		// and a launch that was refused. Conflating those two is exactly
		// what produced the wrong conclusion the first row above records.
		Expect:  Absent,
		Because: "no permission for the origin, so the launch is refused even though the click lands and Chrome requests the navigation. Proves the harness can tell a landed click from a refused launch",
	},
	{
		ID:      "run-dialog",
		What:    "Win+R, paste, Enter",
		Kind:    NeedsHuman,
		Expect:  Manual,
		Because: "not automatable without synthesising keystrokes into a system dialog, which is more risk than this measurement is worth",
	},
	{
		ID:      "outlook-email",
		What:    "a link in a received email, which is how a link reaches someone off this machine",
		Kind:    NeedsHuman,
		Expect:  Manual,
		Because: "⚠️ the route with the most intermediaries and so the likeliest to mangle: the sending client, the transport, the receiver's link rewriter and Outlook's own protocol check each touch it",
	},
}
