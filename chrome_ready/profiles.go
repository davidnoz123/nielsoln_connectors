// profiles.go -- R3 and R6: make the profiles visible, and say when one is busy.
//
// # R3, AND WHY IT IS "LIST" AND NOT "REMOVE"
//
// ⚠️ 3.9 GB of Chrome user-data-dirs accumulated on one machine without
// anybody deciding to spend it: 3,015 MB, 265 MB, 232 MB and 356 MB, measured
// 20 Oct 2026. Each is a whole Chrome state tree rather than a profile, so the
// cost scales with the number of CDP PORTS and not with accounts, and the part
// actually wanted is about 80 KB of cookies.
//
// So `-list` reports what exists and what it costs. ⚠️ THERE IS NO `-remove`,
// deliberately: this binary is invocable BY A LINK, and a link that can delete
// a directory tree is a capability nobody asked for. Reporting the path is
// enough, because deleting it is one command the user already knows.
//
// # R6
//
// Chrome writes `lockfile` into a user-data-dir it owns and REMOVES IT on a
// clean close, measured the same day. So its presence answers "is somebody
// already using this profile", which is the difference between a useful
// message and the 25-second timeout the launch path would otherwise give.
//
// ⚠️ A CRASHED CHROME CAN LEAVE A STALE LOCKFILE. That makes "busy" a
// probably rather than a certainly, and the error says so: the failure
// direction is a warning nobody needed rather than a clash nobody saw coming.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// chromeLockName is what Chrome writes into a user-data-dir it holds.
const chromeLockName = "lockfile"

// profileBusy reports whether a Chrome appears to hold this directory.
func profileBusy(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, chromeLockName))
	return err == nil
}

// profileInfo is one user-data-dir and what it costs.
type profileInfo struct {
	Name  string
	Path  string
	Bytes int64
	Busy  bool
	Mark  *marker // nil when this profile was not started by chrome_ready
}

// listProfiles enumerates the profiles under the default base directory.
//
// ⚠️ IT CANNOT SEE PROFILES ELSEWHERE, and says so in the output rather than
// implying the list is complete. `-profile` accepts any path, so a caller who
// declared one is the only record that it exists. On the machine this was
// written for, three of the four live profiles were under C:\Temp and would
// not appear here.
func listProfiles() ([]profileInfo, string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return nil, "", fmt.Errorf("cannot locate the user config directory: %w", err)
	}
	base = filepath.Join(base, "chrome_ready")

	entries, err := os.ReadDir(base)
	if os.IsNotExist(err) {
		return nil, base, nil
	}
	if err != nil {
		return nil, base, err
	}

	var out []profileInfo
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(base, e.Name())
		info := profileInfo{Name: e.Name(), Path: dir, Busy: profileBusy(dir)}
		if m, err := readMarker(dir); err == nil {
			info.Mark = &m
		}
		info.Bytes = dirBytes(dir)
		out = append(out, info)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Bytes > out[j].Bytes })
	return out, base, nil
}

// dirBytes totals the files under dir, ignoring what it cannot read.
//
// ⚠️ A WALK OF A 3 GB TREE IS NOT FREE, which is why this runs only under
// -list and never on the path a click takes. Reporting a size on every launch
// would add seconds to the thing a link is waiting for.
func dirBytes(dir string) int64 {
	var total int64
	_ = filepath.Walk(dir, func(_ string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil // unreadable entries are skipped, not fatal
		}
		if !fi.IsDir() {
			total += fi.Size()
		}
		return nil
	})
	return total
}

// reportProfiles prints the list for a human.
func reportProfiles(out *os.File) int {
	profiles, base, err := listProfiles()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot list profiles: %v\n", err)
		return 1
	}
	fmt.Fprintf(out, "chrome_ready profiles under %s\n\n", base)
	if len(profiles) == 0 {
		fmt.Fprintln(out, "  none")
	}
	var total int64
	for _, p := range profiles {
		total += p.Bytes
		state := "idle"
		if p.Busy {
			state = "BUSY"
		}
		owned := "not started by chrome_ready"
		if p.Mark != nil {
			owned = fmt.Sprintf("port %d, browser %s", p.Mark.Port, p.Mark.BrowserID)
		}
		fmt.Fprintf(out, "  %-20s %7.0f MB  %-5s  %s\n",
			p.Name, float64(p.Bytes)/(1<<20), state, owned)
	}
	if len(profiles) > 0 {
		fmt.Fprintf(out, "\n  %d profile(s), %.0f MB in total\n",
			len(profiles), float64(total)/(1<<20))
		// ⚠️ The honest caveat on the number, not a footnote. Most of a
		// profile is Chrome's own payload rather than anything the capability
		// put there.
		fmt.Fprintln(out, "  ⓘ Most of each is Chrome's own state tree -- ML "+
			"models, Safe Browsing lists, component cache. The part this")
		fmt.Fprintln(out, "    capability needs is the cookie store, tens of "+
			"kilobytes. Delete a directory to reclaim the rest; a new one")
		fmt.Fprintln(out, "    is rebuilt on the next run, and a signed-in "+
			"session is lost with it.")
	}
	// ⚠️ SAID EVERY TIME. A list that looks complete and is not is worse than
	// no list, and -profile accepts any path.
	fmt.Fprintln(out, "\n  ⚠️ This lists only profiles under the default base. "+
		"A -profile given as a PATH lives wherever the caller said,")
	fmt.Fprintln(out, "    and nothing here records that it exists.")
	return 0
}
