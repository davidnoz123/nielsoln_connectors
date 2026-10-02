//go:build !windows

// raise_other.go -- the catch-all, so no platform is left with raisePID
// undefined. See raise_windows.go for why that file exists at all.
//
// macOS raising happens in `activate` through `open -a`, which is
// LaunchServices and needs no Automation grant. There is no permission-free
// way to raise a SPECIFIC window there, so there is nothing for a pid-targeted
// version to do.
package main

// raisePID reports that it did nothing, so `activate` falls back.
func raiseWindow(pid int, title string) bool { return false }

// windowExists cannot be answered here without AppleScript and the Automation
// grant that comes with it, so it says no and the caller falls back to waiting
// for an existing page to reconnect.
func windowExists(title string) bool { return false }
