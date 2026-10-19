// profile.go -- where the browser profile lives, and why not the two obvious
// places.
//
// # THE DEFECT THIS FIXES
//
// ⚠️ THE PROFILE WAS UNDER os.TempDir(). The stated goal of the whole
// capability is "a CDP Chrome with MY profile so I automatically access my
// ChatGPT account", and a profile in the temp directory delivers the opposite:
// a sign-in that works, persists for a while, and then vanishes on a disk
// cleanup with no symptom except a capture that suddenly finds an empty
// account. Found on 19 Oct 2026 while writing caveats.go, which is an argument
// for writing caveats: describing a reach honestly meant reading what it
// actually touched.
//
// # WHY NOT CHROME'S OWN DEFAULT PROFILE, WHICH WOULD NEED NO SIGN-IN
//
// Two reasons, and the first is enough on its own.
//
//  1. ⚠️ CHROME ALLOWS ONE PROCESS PER user-data-dir. A browser the user
//     already has open holds the lock, so a launch either fails to bind the
//     debugging port or silently opens a tab in the existing window and
//     returns success. Measured today: 40 chrome processes were running, which
//     is the normal state of that machine.
//  2. ⓘ RECENT CHROME RESTRICTS REMOTE DEBUGGING ON THE DEFAULT PROFILE.
//     Google added this because a local process able to enable debugging on a
//     signed-in profile can read its cookies. ⚠️ STATED AS UNVERIFIED HERE:
//     Chrome 155.0.8059.41 is installed and the restriction is believed to
//     apply, but it could not be TESTED without launching Chrome against the
//     user's live profile, which would disturb the session it was meant to
//     measure. Do not quote this as measured until a machine with Chrome
//     closed is available.
//
// # AND WHY NOT COPY THE DEFAULT PROFILE
//
// It would work, and it is refused on purpose. Copying a signed-in profile
// copies its cookie jar and token store: a credential-copying operation
// performed by a program a LINK can invoke. The caveat for that would have to
// say so, and no capability here needs it when signing in once achieves the
// same end.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// profileDir resolves the -profile flag to an absolute directory.
//
// ⚠️ A PATH IS USED AS GIVEN, which is AGENTS' "declared, never allocated"
// applied to a directory. The previous version took a bare NAME and joined it
// to the temp directory, so the connector chose the location and the caller
// could not override it. A caller who declares a path gets that path.
func profileDir(nameOrPath string) (string, error) {
	if nameOrPath == "" {
		return "", fmt.Errorf("-profile is empty. Name a directory, or a " +
			"bare name to place one under the user config directory")
	}

	// A path, by any of the spellings Windows and POSIX use.
	if filepath.IsAbs(nameOrPath) ||
		strings.ContainsAny(nameOrPath, `/\`) ||
		strings.HasPrefix(nameOrPath, ".") {
		abs, err := filepath.Abs(nameOrPath)
		if err != nil {
			return "", fmt.Errorf("cannot resolve -profile %q: %w",
				nameOrPath, err)
		}
		if err := refuseTemp(abs); err != nil {
			return "", err
		}
		return abs, nil
	}

	// A bare name. os.UserConfigDir is %AppData% on Windows -- durable, and
	// NOT a subdirectory of the temp directory the way %LocalAppData%\Temp is.
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("cannot locate the user config directory, so "+
			"there is nowhere durable to put a profile named %q. Pass a full "+
			"path instead: %w", nameOrPath, err)
	}
	return filepath.Join(base, "chrome_ready", nameOrPath), nil
}

// refuseTemp rejects a profile inside the temp directory.
//
// ⚠️ REFUSED RATHER THAN WARNED ABOUT. The failure it prevents is silent and
// delayed: everything works, the sign-in sticks for days, and then a cleanup
// removes it and the capability reports an empty account. A warning on the run
// that sets it up would be read once and forgotten long before the symptom.
func refuseTemp(dir string) error {
	tmp, err := filepath.Abs(os.TempDir())
	if err != nil {
		return nil // cannot tell, so do not pretend to
	}
	rel, err := filepath.Rel(tmp, dir)
	if err != nil || strings.HasPrefix(rel, "..") {
		return nil // outside the temp tree, which is what we want
	}
	return fmt.Errorf("REFUSED: %s is inside the temp directory (%s).\n"+
		"  A browser profile there loses its sign-in to any disk cleanup, and "+
		"the symptom is a capture that\n"+
		"  finds an empty account rather than an error. Pass a durable path, "+
		"or a bare name to use the default\n"+
		"  location under the user config directory", dir, tmp)
}

func dirExists(dir string) bool {
	fi, err := os.Stat(dir)
	return err == nil && fi.IsDir()
}
