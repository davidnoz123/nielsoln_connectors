// known_overlaps.go -- the baseline, so that NEW duplication fails the build.
//
// ⚠️ THE BASELINE IS THE WHOLE POINT. Overlap is not a defect: `main` is in
// every connector, and two functions can share a name while doing different
// jobs. What this file forbids is UNNOTICED overlap. Anything not written
// down here fails, so a new one stops the build until somebody has looked and
// either shared the code or recorded why not.
//
// Measured 14 Oct 2026. Every entry carries why it is acceptable, because an
// allowlist without reasons is an allowlist nobody dares to prune.
//
// ⚠️ AND IT IS SHORT BY CONSTRUCTION. The first version listed seven more
// names, taken from a grep that had not distinguished a cross-connector
// overlap from a build-tag pair inside ONE module: flipCase, probeFold,
// windowExists, isReparse, isReparseInfo, isLoopback, contains. All seven
// were fiction. TestBaselineHasNoStaleEntries found them, which is the same
// discipline in the other direction: an allowlist must not claim duplication
// that does not exist, or the next reader cannot tell which entries are live.
package main

import "strings"

// acknowledged maps a function name to why its duplication is tolerated.
var acknowledged = map[string]string{
	"main": "every connector is a program. Nothing to share.",

	// ⚠️ THE ONE THAT MATTERS, and the reason this check exists at all.
	//
	// `contained` is the path-containment check: it decides whether a
	// resolved path sits inside the permitted root, which is the traversal
	// defence. It is BYTE-IDENTICAL in pdf_view_server and remote_ai.
	//
	// And it calls `foldsCase`, which is NOT identical: 23 lines against 16.
	// remote_ai's version deliberately does not cache an undetermined
	// answer, with the reason in its own comment -- "an empty share cannot be
	// asked whether it folds, and the moment it has a file in it, it can" --
	// and pdf_view_server's caches `false` on a failed stat. So a learned
	// lesson is stranded in one copy.
	//
	// Being accurate about severity: the failure direction is over-strict
	// REJECTION, not admission. Both are conservative where it matters. So
	// this is a divergence in safety-adjacent code rather than a hole.
	//
	// ⚠️ THE LESSON IS NOW PROPAGATED, 14 Oct 2026, but the code still
	// differs. pdf_view_server caches only `os.IsNotExist`, so a permission
	// error or a share hiccup is no longer remembered as a determined
	// answer. remote_ai still delegates to `probeFold` while this inlines
	// the stat pair, so the bodies remain different and this check still
	// says DIVERGED. That is accurate rather than a failure to tidy: the
	// behaviours agree and the implementations do not.
	//
	// The check also SAW the change, the hash moving from aa622d57ec75 to
	// a85800ddcab6, which is the one piece of evidence that it is live
	// rather than decorative.
	"contained": "identical copies of the traversal defence. See the note above and s60.",
	"foldsCase": "different bodies, same behaviour since 14 Oct. s60.",
	// Same name, genuinely different jobs. Measured: 12 lines against 92.
	"resolve": "unrelated functions that happen to share a name.",

	// pdf_view_server raises a specific window through user32 in 96 lines;
	// remote_chrome_cdp's 14-line version does a different and smaller job.
	"raiseWindow": "different jobs behind one name, measured at 96 lines against 14.",

	// Added 15 Oct 2026, and the check found it on the first run after
	// linkprobe existed, which is the evidence that it is live:
	//
	//	chrome_sessions  probe(port int, timeout time.Duration) Session
	//	linkprobe        probe(r Route, ...) Result
	//
	// One asks a CDP port over HTTP what it is; the other triggers a link
	// down one route and classifies what came back. Different signatures,
	// different return types, no shared logic. The word is just the English
	// for what both do.
	// ⚠️ THREE modules now, after chrome_ready landed 15 Oct 2026:
	//
	//	chrome_sessions  probe(port, timeout) Session   -- CDP port query
	//	linkprobe        probe(Route, ...) Result       -- link-route runner
	//	chrome_ready     probe(port, timeout) (version, error)
	//
	// ⓘ The first and third are genuinely CLOSE: both ask a CDP port what it
	// is over HTTP. They are not shared because chrome_sessions returns a
	// Session carrying an owner PID and an error STRING for rendering, while
	// chrome_ready needs the websocket URL and a real error to retry on. A
	// shared one would have to return the union and each caller would ignore
	// half of it.
	//
	// Written down rather than extracted, and if a FOURTH appears that is
	// the signal to reconsider. Two near-identical probes is a coincidence;
	// three with one more coming is a leaf.
	"probe": "three functions asking similar questions with different return needs. See the note above.",

	// chrome_ready prints a CDP version block in 5 lines; linkprobe renders
	// a route matrix with findings and detail in about 40.
	"report": "different jobs behind one name: a version block and a result matrix.",
}

// Acknowledged reports whether an overlap has been written down.
func Acknowledged(name string) bool {
	_, ok := acknowledged[name]
	return ok
}

// acknowledgedBodies is the baseline for SharedBodies: one entry per set of
// NAMES that share a body across modules.
//
// ⚠️ Keyed on the joined names rather than the body hash, deliberately. A
// hash changes the moment anybody edits either copy, so a hash-keyed
// baseline would go stale on every touch and teach people to re-bless it
// without looking. The names are what a reader recognises.
var acknowledgedBodies = map[string]string{}

// AcknowledgedBody reports whether a renamed copy has been written down.
func AcknowledgedBody(names []string) bool {
	_, ok := acknowledgedBodies[strings.Join(names, "+")]
	return ok
}

// ⚠️ Declared 16 Oct 2026, and it is the ONE real duplicate that survived
// eleven Python name collisions when the body-keyed check was widened.
//
//	linkprobe/chrome_route.py   CDP.close()
//	probes/cdp.py               CDP.close()
//
// Byte-identical, 5 lines: close the socket, log a warning if it fails.
// Both belong to a CDP client, and the clients themselves are the real
// duplication -- declared as a recipe in known_recipes.go, with the reason
// they are not yet collapsed.
//
// Kept as an entry rather than fixed here because probes/ exists to stop
// losing instruments, not to refactor linkprobe. ⚠️ If chrome_route.py ever
// becomes importable, both this entry and the recipe should go.
func init() {
	acknowledgedBodies["close"] = "the CDP client's socket close, identical in two clients. See known_recipes.go."
}

// ⚠️ Declared 19 Oct 2026, and SharedBodies found it under TWO DIFFERENT
// NAMES, which is exactly the blind spot the body-keyed check exists for:
//
//	probes/canonjson   pythonString
//	rowpatch           canonString
//
// Thirty lines, same body: CPython's string escaping under
// ensure_ascii=False, which Go's own escaper cannot be configured to produce
// because encoding/json always escapes U+2028 and U+2029.
//
// ⚠️ THE TWO COPIES HAVE DIFFERENT JOBS, and that is why this is declared
// rather than collapsed. probes/canonjson MEASURES whether Go can match
// Python and must keep its own copy to stay an independent instrument -- an
// instrument that imported the thing it measures would only ever confirm
// itself. rowpatch USES the rule in anger.
//
// ⚠️ AND THIS IS s13'S TRIGGER FIRING FOR THE SECOND TIME. s13 decided
// extraction happens at SECOND USE, and this is a second use. The fix is a
// leaf module holding the canonical encoder, which would be this repo's first
// `require` block and first go.sum -- a decision belonging on the slate, not
// in a commit that happened to notice. Until then: two copies, one declared
// reason, and a check that will say so again if a third appears.
func init() {
	acknowledgedBodies["canonString+pythonString"] = "CPython's " +
		"ensure_ascii=False string escaper. The probe MEASURES the rule and " +
		"must stay independent of the code that uses it; rowpatch applies " +
		"it. s13's extraction trigger has now fired twice -- see the comment " +
		"above."
}
