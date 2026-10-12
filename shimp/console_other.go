//go:build !windows

// console_other.go -- the catch-all, so no platform is left with hasConsole
// or allocConsole undefined. See console_windows.go for why they exist.
//
// ⚠️ It reports that it CANNOT SAY rather than guessing. There is no portable
// equivalent of AllocConsole: a Unix process either inherited a terminal or
// did not, and nothing lets it create one for itself. The nearest equivalent
// is spawning a terminal emulator, which is per-desktop-environment and so
// the least portable thing in the design.
//
// So hasConsole asks whether stdout is a terminal, which is the honest local
// question, and allocConsole refuses with a reason rather than silently
// succeeding and leaving output nowhere.
package main

import (
	"fmt"
	"os"
	"runtime"
)

// hasConsole reports whether stdout looks like a terminal.
func hasConsole() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// allocConsole refuses, and says what would have to be built.
func allocConsole() error {
	if hasConsole() {
		return nil
	}
	return fmt.Errorf(
		"cannot create a console on %s: there is no portable equivalent of "+
			"AllocConsole, and a process that did not inherit a terminal "+
			"cannot give itself one. Spawning a terminal emulator is "+
			"per-desktop-environment and is the work this would need",
		runtime.GOOS)
}
