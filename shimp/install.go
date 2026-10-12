// install.go -- the generic half of install, uninstall and verify.
//
// No build tag. Everything platform-specific is behind the `claims` seam
// below, which is the repo's "one file of generic code, platform specifics in
// their own files" rule, and it is also the test seam: AGENTS asks that a
// fake be injected through a seam that already exists rather than through a
// parameter added for the test, and this one has to exist anyway because
// Windows and everything else claim a scheme by different means.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// The keys install claims. FUNCTIONS rather than constants, because the
// scheme is now a variable: that is what gives a test installation its own
// registry class, its own home and its own Office key, so it cannot disturb a
// working one.
func classKey() string { return `Software\Classes\` + scheme }

// commandKey() is the subkey Windows reads to find the program.
func commandKey() string { return classKey() + `\shell\open\command` }

// trustedProtocolKey() suppresses Office's own security prompt.
//
// ⚠️ THE TRAILING COLON IS PART OF THE KEY NAME. Microsoft's own example is
// "to disable the display of a security warning for the Notes: protocol,
// enter Notes:". A key named "shimp" looks correct in regedit and does
// nothing at all. This is s8, and the cost of getting it wrong is not a
// visible error: on 5 Oct an Excel instance held a workbook for two and a
// half hours behind one such modal, and the only symptom was a read-only
// refusal naming no pid.
func trustedProtocolKey() string {
	return `Software\Policies\Microsoft\Office\16.0\Common\Security\` +
		`Trusted Protocols\All Applications\` + scheme + `:`
}

// claims is every durable thing install creates outside its own directory.
//
// Deliberately tiny and string-only. There is no transaction and no rollback:
// if a later claim fails, install says which ones it made, and `uninstall`
// withdraws exactly those. A half-registered scheme is visible in `verify`
// rather than hidden by a cleanup that might itself fail.
type claims interface {
	set(subkey, name, value string) error
	get(subkey, name string) (value string, present bool, err error)
	// hasKey asks about the KEY alone, which is a different question from
	// reading a value under it.
	//
	// ⚠️ It exists because of s8's trusted-protocol key, where the claim IS
	// the key's existence and Office reads no value. Checking for a default
	// value instead would report FAIL for a key a user had created by hand
	// in regedit, which Office would be perfectly happy with. A verify that
	// reports a working machine as broken gets ignored, and then so does the
	// one that reports a broken one.
	hasKey(subkey string) (bool, error)
	removeKey(subkey string) error
	// where names the store for a human, so a report cannot be mistaken for
	// one taken against the real registry when it was not.
	where() string
}

// claimsFor is the one seam the tests use, and it is the indirection the
// platform split already required: openClaims lives in scheme_windows.go or
// scheme_other.go and the generic code cannot name either directly.
//
// AGENTS asks that a fake be injected through a seam that already exists
// rather than through a parameter added for the test, and that where no seam
// exists, the absence is usually the design problem the test was about to
// find. Here the seam is load-bearing in production too.
var claimsFor = openClaims

func binaryName() string {
	if runtime.GOOS == "windows" {
		return scheme + ".exe"
	}
	return scheme
}

// cmdInstall places the binary, claims the scheme and records what it did.
func cmdInstall(replace bool) int {
	c, err := claimsFor()
	if err != nil {
		_log("FATAL install: %v", err)
		fmt.Fprintf(os.Stderr, "install failed: %v\n", err)
		return 1
	}
	_log("install starting, claims store is %s", c.where())

	dir, err := home()
	if err != nil {
		_log("FATAL install: %v", err)
		return 1
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		_log("FATAL install: cannot create %s: %v", dir, err)
		return 1
	}

	me := provenance()

	// ⚠️ The handler is never silently replaced, s20, because it is the trust
	// root. A DIFFERENT version is refused; the SAME version is allowed
	// through, so that re-running the documented install line is idempotent
	// rather than an error a user has to think about.
	if old, err := readStamp(); err == nil {
		if old.Version != me.Version && !replace {
			msg := fmt.Sprintf("refusing to replace the installed handler. "+
				"Installed: %s. This one: %s. The handler is the trust root, "+
				"so pass -replace if that is what you mean.",
				old.Version, me.Version)
			_log("install REFUSED: %s", msg)
			fmt.Fprintln(os.Stderr, msg)
			return 1
		}
		_log("install: replacing %s with %s", old.Version, me.Version)
	}

	// s7: what gets registered is a PERMANENT binary, never `go run`.
	// Measured, same program, module already cached: go run 649ms and 653ms
	// warm, a built binary 68ms. The binary is not an optimisation, it is the
	// difference between a usable click and an annoying one.
	//
	// ⚠️ And os.Executable() under `go run` is the TOOLCHAIN'S TEMPORARY
	// BUILD OUTPUT, which is exactly why this works: it is self-contained, so
	// copying it is enough, and that is why none of this needs a repo layout
	// change.
	target := filepath.Join(dir, binaryName())
	if err := copySelf(me.Path, target); err != nil {
		_log("FATAL install: cannot place the binary: %v", err)
		fmt.Fprintf(os.Stderr, "install failed: %v\n", err)
		return 1
	}
	_log("install: placed %s", target)

	sum, err := fileSha256(target)
	if err != nil {
		_log("FATAL install: cannot hash the placed binary: %v", err)
		return 1
	}

	// ⚠️ NOT fmt.Sprintf("%q %q", ...). Go's %q escapes a backslash, so a
	// Windows path came out as "C:\\Users\\..." and the registry would have
	// held a command naming a path that does not exist. It looks almost
	// right, which is why the contract test asserts the command names the
	// placed binary rather than merely being non-empty: that assertion is
	// what caught it.
	command := `"` + target + `" "%1"`
	for _, claim := range []struct{ subkey, name, value string }{
		{classKey(), "", "URL:SHIMP Protocol"},
		// An empty "URL Protocol" value is what marks a class as a URL
		// scheme. The VALUE being empty is the convention; the NAME existing
		// is the signal.
		{classKey(), "URL Protocol", ""},
		{commandKey(), "", command},
		// The key itself is the claim here; Office reads its presence, not a
		// value. s8.
		{trustedProtocolKey(), "", ""},
	} {
		if err := c.set(claim.subkey, claim.name, claim.value); err != nil {
			_log("FATAL install: cannot claim %s\\%s: %v",
				claim.subkey, claim.name, err)
			fmt.Fprintf(os.Stderr, "install failed at %s: %v\n", claim.subkey, err)
			return 1
		}
		_log("install: claimed %s name=%q", claim.subkey, claim.name)
	}

	// Written LAST, so a stamp's existence means every claim above succeeded.
	//
	// ⚠️ Scheme is set HERE and not in provenance(), because provenance()
	// describes the running binary while this describes the INSTALLATION.
	// Omitting it was a real defect the live test caught: the stamp recorded
	// an empty scheme, so resolveSchemeFromSelf kept the default and a
	// handler installed as shimp-selftest believed it was production. It
	// would then have refused its own links as belonging to another scheme,
	// and on a machine with a real install the two would have fought over
	// one home directory.
	me.Scheme = scheme
	me.Path, me.Sha256 = target, sum
	if err := writeStamp(me); err != nil {
		_log("FATAL install: cannot record the stamp: %v", err)
		return 1
	}

	_log("install complete: %s %s", me.Version, sum)
	fmt.Printf("installed %s\n  binary  %s\n  version %s\n  sha256  %s\n",
		scheme+":", target, me.Version, sum)
	fmt.Printf("\nRun `%s verify` to see what is true. It will NOT promise "+
		"that clicks are silent.\n", binaryName())
	return 0
}

func copySelf(from, to string) error {
	if from == "" {
		return fmt.Errorf("cannot determine this executable's own path, so " +
			"there is nothing to copy")
	}
	src, err := os.Open(from)
	if err != nil {
		return fmt.Errorf("cannot read this executable at %s: %w", from, err)
	}
	defer src.Close()

	// Written to a temporary name and renamed, so a failed copy cannot leave
	// a truncated binary registered as the handler.
	tmp := to + ".new"
	dst, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return fmt.Errorf("cannot create %s: %w", tmp, err)
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		return fmt.Errorf("cannot write %s: %w", tmp, err)
	}
	if err := dst.Close(); err != nil {
		return fmt.Errorf("cannot close %s: %w", tmp, err)
	}
	// ⚠️ Windows refuses to rename over a RUNNING executable, which is the
	// case when install replaces a handler that is currently handling a
	// click. Removing first narrows that window rather than closing it, and
	// the failure is reported rather than swallowed.
	_ = os.Remove(to)
	if err := os.Rename(tmp, to); err != nil {
		return fmt.Errorf("cannot move %s into place at %s: %w", tmp, to, err)
	}
	return nil
}

// cmdVerify reports what is TRUE, which is not the same as what install
// intended. s8 is the row, and s12 is the pattern: a report that cannot
// distinguish "absent" from "unknown" is worse than no report.
func cmdVerify() int {
	c, err := claimsFor()
	if err != nil {
		_log("FATAL verify: %v", err)
		fmt.Fprintf(os.Stderr, "verify failed: %v\n", err)
		return 1
	}
	fmt.Printf("claims store: %s\n", c.where())

	ok := true
	note := func(good bool, label, detail string) {
		mark := "FAIL"
		if good {
			mark = " ok "
		} else {
			ok = false
		}
		fmt.Printf("[%s] %-28s %s\n", mark, label, detail)
	}

	st, stErr := readStamp()
	if stErr != nil {
		note(false, "install record", fmt.Sprintf("absent or unreadable: %v", stErr))
	} else {
		note(true, "install record", st.Version)
	}

	// The binary, and whether it is still the one that was installed. ⚠️ This
	// is the check that answers s23's second audit question, "is what is
	// installed the thing I think it is". A path existing proves nothing.
	if stErr == nil {
		if sum, err := fileSha256(st.Path); err != nil {
			note(false, "binary", fmt.Sprintf("recorded at %s but unreadable: %v", st.Path, err))
		} else if sum != st.Sha256 {
			note(false, "binary sha256",
				fmt.Sprintf("CHANGED since install. recorded %s, found %s", st.Sha256, sum))
		} else {
			note(true, "binary sha256", sum)
		}
	}

	cmd, present, err := c.get(commandKey(), "")
	switch {
	case err != nil:
		note(false, "scheme registration", fmt.Sprintf("cannot read: %v", err))
	case !present:
		note(false, "scheme registration", "absent, so clicks go nowhere")
	case stErr == nil && !strings.Contains(cmd, st.Path):
		// Points somewhere, but not at us. Louder than absent, because a
		// link still opens something.
		note(false, "scheme registration",
			fmt.Sprintf("points elsewhere: %s", cmd))
	default:
		note(true, "scheme registration", cmd)
	}

	// hasKey rather than get, because Office reads the key's EXISTENCE and
	// no value under it. See the interface comment.
	if present, err := c.hasKey(trustedProtocolKey()); err != nil {
		note(false, "Office trusted protocol", fmt.Sprintf("cannot read: %v", err))
	} else if present {
		note(true, "Office trusted protocol",
			`key "`+scheme+`:" present (the trailing colon is part of the name)`)
	} else {
		note(false, "Office trusted protocol",
			`absent, so every click from Excel or Word raises a modal, and `+
				`⚠️ that modal blocks COM invisibly`)
	}

	// ⚠️ Reported rather than promised, which is the whole point of this
	// command. Microsoft's own account is that a second dialog comes from
	// HLINK.dll and the key above does not suppress it. So verify says the
	// key is present. It does NOT say clicks are silent.
	fmt.Printf("\nverify reports what is present. It does NOT promise silent " +
		"clicks:\n  a second dialog comes from HLINK.dll and the trusted " +
		"protocol key\n  does not suppress it, by Microsoft's own account.\n")
	fmt.Printf("  and the DISPATCH half is not built, so a click is logged " +
		"and refused.\n")

	if !ok {
		return 1
	}
	return 0
}

// cmdUninstall withdraws exactly what install created and nothing else.
func cmdUninstall(purge bool) int {
	c, err := claimsFor()
	if err != nil {
		_log("FATAL uninstall: %v", err)
		fmt.Fprintf(os.Stderr, "uninstall failed: %v\n", err)
		return 1
	}
	_log("uninstall starting, claims store is %s", c.where())

	// ⚠️ Deepest first, because a key with subkeys cannot be deleted. And
	// ⚠️ only shimp's OWN subkey of "All Applications", never that parent:
	// s8 records that the parent is shared, and deleting it would withdraw
	// every other application's trusted protocol.
	failed := 0
	for _, k := range []string{trustedProtocolKey(), commandKey(),
		classKey() + `\shell\open`, classKey() + `\shell`, classKey()} {
		if err := c.removeKey(k); err != nil {
			// Not fatal and not silent: a key that was never there is the
			// normal case when uninstall runs twice.
			_log("uninstall: %s not removed: %v", k, err)
			failed++
			continue
		}
		_log("uninstall: removed %s", k)
	}

	dir, err := home()
	if err == nil {
		for _, name := range []string{binaryName(), stampName} {
			p := filepath.Join(dir, name)
			if err := os.Remove(p); err != nil {
				_log("uninstall: %s not removed: %v", p, err)
				continue
			}
			_log("uninstall: removed %s", p)
		}
		// ⚠️ The LOG is deliberately kept BY DEFAULT. s20 has uninstall
		// remove its runtime state, and a log is not runtime state, it is the
		// record of what happened. Deleting the evidence of an install at the
		// moment somebody is most likely to be diagnosing it would be the
		// wrong kind of tidy.
		//
		// -purge removes it, and the home directory with it. That exists for
		// a TEST cycle, which has to leave the machine exactly as it found
		// it or it is not a test, it is an accumulation.
		if !purge {
			fmt.Printf("kept %s, which is the record of what happened\n",
				filepath.Join(dir, logName))
		}
	}

	_log("uninstall complete, %d claim(s) were already absent", failed)
	fmt.Printf("uninstalled %s:\n", scheme)

	// ⚠️ THE PURGE IS LAST, AND IT SILENCES THE LOG FIRST. Both halves are
	// needed and the live test found out why twice:
	//
	// Purging before the final _log meant that line RECREATED the directory,
	// because _log makes the home if it is missing. The removal succeeded and
	// the directory was back a millisecond later.
	//
	// So logToFile is turned off before the removal, and anything logged
	// afterwards goes to stderr instead of resurrecting what was just
	// deleted. Ordering alone would work today and would break the moment
	// somebody adds a log line below this one.
	if purge {
		if dir, err := home(); err == nil {
			logSuppressed = true
			if err := os.RemoveAll(dir); err != nil {
				fmt.Fprintf(os.Stderr, "could not purge %s: %v\n", dir, err)
				return 1
			}
			fmt.Printf("purged %s\n", dir)
		}
	}
	return 0
}
