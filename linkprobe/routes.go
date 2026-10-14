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
		ID:   "chrome-address-bar",
		What: "navigating Chrome to the link, by command line or by CDP",
		Kind: NeedsHuman,
		// ⚠️ ATTEMPTED AUTOMATICALLY 15 Oct 2026, AND IT CANNOT BE. Two
		// independent routes were tried against a throwaway profile:
		//
		//   chrome.exe <uri>                     handler silent
		//   CDP PUT /json/new?<uri>              handler silent
		//
		// The second is the informative one. Chrome CREATED the tab and
		// `/json/list` showed it holding the URI complete, fragment and
		// all, byte-for-byte as sent. So Chrome does not mangle a shimp
		// link. It simply declines to hand it to the OS.
		//
		// ⚠️ THE READING, and it is good news rather than an obstacle:
		// Chrome gates an external-protocol launch behind a real user
		// gesture, and a programmatic navigation is not one. So a web page
		// CANNOT silently fire a shimp link. That is a security property
		// worth having, and it is why this route stays NeedsHuman: the
		// gesture is the mechanism, not an inconvenience around it.
		//
		// ⚠️ AND A SYNTHETIC CLICK DOES NOT HELP. MEASURED, which killed the
		// plan that was written here first.
		//
		// The previous version of this comment said the route needed
		// Input.dispatchMouseEvent over a websocket, and named extracting
		// remote_chrome_cdp's wsConn as the work. That was a guess. It was
		// tested before being built, and it is wrong:
		//
		//   CONTROL: synthetic click on an ordinary <a href>  NAVIGATES
		//   SHIMP:   same click on a shimp:// <a href>        handler silent
		//
		// The control is what makes this conclusive. The click mechanism
		// works, the coordinates are right, and Chrome processes it as a
		// real link activation, because the ordinary link navigated. The
		// same click on a custom scheme fires nothing, raises NO dialog
		// (window enumeration on the test Chrome found only the page), and
		// leaves the tab where it was. Two passes, same result.
		//
		// So Chrome distinguishes CDP-injected input from real user input
		// for an external-protocol launch specifically, and blocks it
		// SILENTLY. There is no automated browser route to be had.
		Expect:  Manual,
		Because: "measured twice: Chrome holds the URI verbatim and will not launch it from a programmatic navigation OR from a CDP-synthesized click, while an ordinary link clicked the same way navigates fine. Human-only by Chrome's design, not by our tooling",
	},
	{
		ID:   "chrome-link-click",
		What: "clicking an <a href> on a page, which is NOT the address bar",
		Kind: NeedsHuman,
		// Still distinct from the row above: the href is resolved against
		// the page's base URL before anything else sees it, and the page
		// origin is what the protocol permission is keyed on. The probe
		// writes click_me.html so a human only has to click.
		//
		// ⓘ A profile-local preference DOES exist for pre-allowing the
		// protocol per origin, and seeding a throwaway profile with
		//   protocol_handler.allowed_origin_protocol_pairs
		// touches no machine state. It did not make a programmatic
		// navigation fire, consistent with the gesture being the gate
		// rather than the permission. Chrome's AutoLaunchProtocolsFromOrigins
		// POLICY is the documented suppression mechanism and was
		// deliberately not used: it is machine state under HKCU\Policies,
		// and a probe should not rewrite the user's browser policy.
		// ⚠️ This is the row the CDP click was actually tested against, and
		// it is the one that proves there is no automated route: the click
		// lands on a real <a href> on a real file:// page and still fires
		// nothing. See chrome-address-bar above for the control.
		Expect:  Manual,
		Because: "⚠️ measured: a CDP-synthesized click on this exact page fires nothing, while the same click on an ordinary link navigates. Chrome blocks an external-protocol launch from synthetic input, silently and with no dialog",
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
