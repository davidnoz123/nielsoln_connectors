//go:build !windows

// dispatch_other.go -- the catch-all, so no platform is left with dispatchURI
// undefined. See dispatch_windows.go for why that file exists at all.
//
// ⚠️ It REFUSES rather than shelling out to `open` or `xdg-open`. Those would
// work, and the result would be worthless for the purpose: s36 wants to know
// what the OS passes to a registered handler, and nothing is registered here
// because scheme_other.go refuses to install. A dispatcher that succeeded
// while nothing could receive the URI would report a measurement that was
// never taken, which is the failure shape this repo is organised against.
package main

import (
	"fmt"
	"runtime"
)

func dispatchURI(uri string) error {
	return fmt.Errorf(
		"dispatching a %s:// URI through the OS is Windows-only so far, and "+
			"this is %s. Nothing is registered here because scheme_other.go "+
			"refuses to install, so there would be nothing to receive it",
		scheme, runtime.GOOS)
}
