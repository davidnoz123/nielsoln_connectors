//go:build !windows

// scheme_other.go -- the catch-all, so no platform is left with openClaims
// undefined. See scheme_windows.go for why that file exists at all.
//
// ⚠️ THIS FILE IS NOT OPTIONAL, and it does not return a value that reads as
// an answer. The repo's rule, with chrome_sessions/owner_other.go as the
// worked example: a catch-all should say what it CANNOT do, because an empty
// result reads as "nothing is there" and that is the failure shape everything
// here is organised against.
//
// So openClaims REFUSES on a non-Windows platform rather than succeeding
// against a store that does not exist. An install that silently claimed
// nothing would leave a user with a binary in place, a stamp recording
// success, and a scheme nobody handles.
//
// s22 is the row: Windows gets the first complete implementation, with the
// protocol defined cross-platform from the start, and the spec says "local
// IPC endpoint" rather than "named pipe" so that nothing here leaks upward.
// The work when somebody needs it:
//
//	macOS  an app bundle with CFBundleURLTypes in Info.plist, registered
//	       with Launch Services. A bare executable cannot claim a scheme.
//	Linux  a .desktop file with MimeType=x-scheme-handler/shimp, plus
//	       update-desktop-database, which is per-desktop-environment and
//	       therefore the least portable of the three.
//
// Neither should be written for symmetry. Write it when a capability
// actually has to be clicked there.
package main

import (
	"fmt"
	"runtime"
)

func openClaims() (claims, error) {
	return nil, fmt.Errorf(
		"claiming the %s: scheme is Windows-only so far, and this is %s. "+
			"Refusing rather than installing a binary that nothing would "+
			"ever invoke. s22 records the build order and scheme_other.go "+
			"records what the macOS and Linux work would be",
		scheme, runtime.GOOS)
}
