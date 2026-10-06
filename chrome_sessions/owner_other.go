//go:build !windows

// owner_other.go -- the catch-all, so no platform is left with ownerOfPort
// undefined. See owner_windows.go for why that file exists at all.
//
// There is no portable way to ask who holds a TCP port without either raw
// privileges or shelling out to lsof, and a report that shells out to a tool
// that may not be installed is worse than a report with a blank in it. So the
// honest answer here is "cannot say", which is what 0 means to the caller.
//
// A macOS or Linux implementation would most likely parse `lsof -nP -iTCP
// -sTCP:LISTEN`, and it should only be written when something actually needs
// the owner there, rather than for symmetry.
package main

// ownerOfPort reports that it cannot say.
func ownerOfPort(port int) int { return 0 }

// discoverPorts reports that it cannot enumerate, and says why rather than
// returning an empty list that reads as "nothing is listening".
func discoverPorts() ([]int, string) {
	return nil, "port discovery is Windows-only so far, so -discover found " +
		"nothing here. Name the ports with -ports."
}
