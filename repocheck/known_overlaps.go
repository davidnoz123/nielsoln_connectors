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
}

// Acknowledged reports whether an overlap has been written down.
func Acknowledged(name string) bool {
	_, ok := acknowledged[name]
	return ok
}
