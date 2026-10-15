//go:build !windows

// refusal_other.go -- the catch-all, so no platform is left with showRefusal
// undefined. See refusal_windows.go for why that file exists at all.
//
// ⚠️ It writes to stderr and does NOT pretend to have shown anybody
// anything. The repo's rule, with owner_other.go as the worked example: a
// catch-all should say what it CANNOT do. A click cannot reach this platform
// yet, because scheme_other.go refuses to install, so there is no console to
// allocate and nothing to hold open.
package main

import (
	"fmt"
	"os"
)

func showRefusal(reason string) {
	fmt.Fprintf(os.Stderr, "%s:// REFUSED: %s\n", scheme, reason)
}
