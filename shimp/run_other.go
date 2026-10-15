//go:build !windows

// run_other.go -- the catch-all, so no platform is left with runLink
// undefined. See run_windows.go for why that file exists at all.
//
// ⚠️ It REFUSES rather than running the argv, and that is not laziness. The
// Windows version will not execute without a console in which to ask a human
// for consent, because s64 measured that a web page can fire a link with no
// gesture. This platform has no equivalent prompt written yet, so the honest
// behaviour is to decline and say so, not to run code with no gate.
//
// scheme_other.go already refuses to install here, so no click can reach
// this today. When somebody builds the install half for another platform,
// the consent prompt is part of the work, not an optional extra.
package main

import "runtime"

func runLink(u *URI) int {
	reason := "running a link is Windows-only so far, and this is " +
		runtime.GOOS + ". The consent prompt that gates execution is not " +
		"written for this platform, and shimp will not execute a link " +
		"without one"
	_log("click REFUSED: %s", reason)
	showRefusalFn(reason)
	return 1
}
