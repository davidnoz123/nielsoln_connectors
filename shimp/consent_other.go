//go:build !windows

// consent_other.go -- the catch-all, so no platform is left with askConsent
// undefined. See consent_windows.go for why that file exists at all.
//
// ⚠️ It returns FALSE, always, and that is the correct behaviour rather than
// a stub. There is no consent prompt written for this platform, and shimp
// will not execute a link without one. run_other.go already refuses before
// reaching here; this exists so the refusal cannot be bypassed by any future
// caller that forgets.
package main

import "time"

func askConsent(show func(), d time.Duration) bool {
	_log("click: consent cannot be asked for on this platform, so nothing " +
		"will run")
	return false
}
