// shimp -- the SHIMP/1 handler. SHIMP is "Shim Protocol", s27.
//
// A clickable link that names immutable software, guarantees one running copy
// of it per interactive session, and delivers an action to that copy. The
// design lives in the `shimp` slate in this repo, rendered to SHIMP.md, and
// every sN reference in this package points at a row there.
//
// WHAT IS BUILT HERE, AND WHAT IS NOT. This is the INSTALL half:
//
//	install    place a permanent binary and claim the scheme        s7, s8
//	uninstall  withdraw only what install created                   s20
//	verify     report what is TRUE, never what was intended         s8
//	version    the handler's own provenance, for the second audit   s23
//
// The DISPATCH half, which parses a shimp:// URI and runs a keeper, is not
// built. The grammar it will implement is frozen: s43 the runner table, s44
// the identity bytes with a test vector, s45 argv encoding, s41 the action in
// the fragment. Clicking a link today logs the attempt and says so.
//
// Why this half first: it needs no open decision, it is the half that fails
// INVISIBLY if done wrong (s8's modal held an Excel for two and a half hours
// with no symptom but a read-only refusal), and running it produces the one
// measurement s36 cannot get any other way, which is how much of a URI
// survives the registry hop into os.Args[1].
//
//	go run -ldflags "-H=windowsgui" \n//	    github.com/davidnoz123/nielsoln_connectors/shimp@<sha> install
//
// ⚠️ THE LDFLAGS ARE PART OF THE LINE, and s50 is why. A
// console-subsystem handler is given a console window by Windows on EVERY
// click, visible and in the FOREGROUND, which is measured rather than
// supposed. Omitting the flag still installs, still works, and steals
// focus for ever, so `install` warns and `verify` reports FAIL.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

// scheme is the URI scheme, and the single place its name is written.
//
// A VARIABLE rather than a constant, and that is what makes an isolated test
// installation possible at all. There is exactly one HKCU\Software\Classes\shimp
// per user, so without this a test install IS the production install.
//
// ⚠️ And s44 makes a second scheme CORRECT rather than a convenience: the
// identity bytes include the scheme deliberately, so a shimp-test:// link
// hashes differently and addresses a different singleton. A test click can
// never be answered by a production instance. That decision was taken for
// another reason and happens to buy exactly this.
//
// Set by `install -scheme`, recorded in the stamp, and recovered by a running
// handler from the stamp beside its OWN binary (resolveSchemeFromSelf), which
// is why nothing has to go in the registry command.
//
// ⚠️ The TRAILING COLON in Office's trusted-protocol key is a separate
// constant below, because a key named "shimp" looks right in regedit and does
// nothing. See s8.
var scheme = "shimp"

// logName is under the shimp home, and it is not optional.
//
// ⚠️ CORRECTED 12 Oct 2026, having been measured. This comment used to assert
// that a handler invoked from a URI click has NO CONSOLE. That is FALSE:
// Windows gives a clicked handler a console, and the click now records it
// (`has a console: true`). The claim was never checked.
//
// The log's justification survives the correction intact, because the console
// Windows provides is useless for this purpose: the handler exits in about
// 36ms, taking the window with it, so anything written there is gone before
// it can be read. The only symptom a user gets is that clicking the link did
// nothing.
//
// So every run appends here, including the runs that succeed, because "logged
// starting and never finished" is what tells the next person where it died.
// s50 is where the console itself is decided, and that window flashing on
// every click is a defect in its own right.
const logName = "shimp.log"

// stampName records what install knew at install time and nothing else does
// afterwards. s23: the second audit question, "is what is installed the thing
// I think it is", cannot be answered without it.
const stampName = "installed.json"

// out is where the human-facing report goes.
//
// A seam rather than a bare os.Stdout, and AGENTS names the reason: where no
// seam exists, the absence is usually the design problem the test was about
// to find. A `verify` whose output cannot be captured can only be asserted on
// by its EXIT CODE, so a test could say "verify failed" and never which check
// failed. That is how TestVerifyFailsWhenTheBinaryChangedUnderneathIt came to
// be asserting the aggregate rather than its own subject.
var out io.Writer = os.Stdout

// logTo, when set, captures the log instead of writing it. Tests only.
var logTo io.Writer

// logSuppressed stops _log recreating a home directory that has just been
// purged. Set by `uninstall -purge` immediately before the removal.
var logSuppressed bool

// _log appends one line, OPENING AND CLOSING THE FILE EACH TIME.
//
// ⚠️ Deliberately not a held handle, and this is a correctness choice rather
// than a style one. Two reasons, the second measured:
//
// A handler can die abruptly, and a buffered line that never reached disk is
// the silent failure this log exists to prevent.
//
// And on Windows an open handle cannot be deleted, so `uninstall -purge`
// failed with "The process cannot access the file because it is being used by
// another process" -- the process being itself. Found by the live test. The
// fix is not to close the handle at the right moment; it is not to hold one.
//
// The cost is a syscall per line, for a program that writes a handful.
func _log(format string, a ...any) {
	line := fmt.Sprintf("%s  %s\n", time.Now().Format(time.RFC3339),
		fmt.Sprintf(format, a...))
	if logTo != nil {
		_, _ = io.WriteString(logTo, line)
		return
	}
	dir, err := home()
	if err == nil && !logSuppressed {
		if err = os.MkdirAll(dir, 0o755); err == nil {
			f, ferr := os.OpenFile(filepath.Join(dir, logName),
				os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
			if ferr == nil {
				_, _ = io.WriteString(f, line)
				_ = f.Close()
				return
			}
		}
	}
	// Nowhere to append. stderr is better than nothing, even though a click
	// cannot see it.
	_, _ = io.WriteString(os.Stderr, line)
}

// home is ~/.shimp, where the binary, the log and the stamp live.
//
// Per-USER rather than per-session, which is s30: immutable state per user,
// live state per session. The binary and the stamp are immutable; the lock
// and the IPC endpoint the dispatch half will add are not, and they go
// somewhere session-scoped.
func home() (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot locate the user home directory, so "+
			"there is nowhere to install to: %w", err)
	}
	return filepath.Join(dir, "."+scheme), nil
}

// stamp is what install records about itself.
//
// Version is the module pseudo-version from the build info, NOT a bare 40-hex
// commit. ⚠️ Measured limit, worth stating rather than hiding: a binary built
// by `go run module@sha` can learn its own pseudo-version, which carries a
// 12-hex abbreviation and a timestamp, and cannot learn the full 40-hex SHA.
// s37 requires 40 hex in a LINK, where a human or a tool supplies it. This is
// the handler describing itself, where the pseudo-version is what the
// toolchain actually resolved and go.sum pinned, so it is the honest answer
// and arguably a better one than a bare SHA.
type stamp struct {
	// Scheme is what this installation claimed, so a handler invoked by the
	// registry can recover it without being told on the command line.
	Scheme    string `json:"scheme"`
	Path      string `json:"path"`
	Module    string `json:"module"`
	Version   string `json:"version"`
	Sha256    string `json:"binary_sha256"`
	GoVersion string `json:"go_version"`
	At        string `json:"at"`
}

// provenance reads what this running process can say about itself.
func provenance() stamp {
	s := stamp{GoVersion: runtime.Version(), At: time.Now().UTC().Format(time.RFC3339)}
	if bi, ok := debug.ReadBuildInfo(); ok {
		s.Module = bi.Main.Path
		s.Version = bi.Main.Version
	}
	// "(devel)" is what a local `go run .` reports, and an empty string is
	// what some build paths report. Neither is a lie, and neither is useful
	// to an auditor, so say which it was rather than leaving a blank.
	if s.Version == "" {
		s.Version = "(unknown: no module version in build info)"
	}
	if exe, err := os.Executable(); err == nil {
		s.Path = exe
		if sum, err := fileSha256(exe); err == nil {
			s.Sha256 = sum
		}
	}
	return s
}

func fileSha256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// resolveSchemeFromSelf recovers the scheme this binary was installed as.
//
// ⚠️ It reads the stamp beside the EXECUTABLE, never the one under home(),
// because home() is derived from the scheme and that would be circular.
// Failure is silent and leaves the default in place, which is correct: an
// uninstalled binary run from a checkout is the production scheme by
// definition, and there is nothing to recover.
func resolveSchemeFromSelf() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	raw, err := os.ReadFile(filepath.Join(filepath.Dir(exe), stampName))
	if err != nil {
		return
	}
	var s stamp
	if json.Unmarshal(raw, &s) == nil && s.Scheme != "" {
		scheme = s.Scheme
	}
}

// schemePattern is what a scheme name may contain.
//
// ⚠️ Validated rather than trusted, because this string becomes a REGISTRY
// KEY PATH. A scheme containing a backslash would claim a key somewhere
// nobody intended, and one containing a dot or a space would produce a class
// Windows will not dispatch. RFC 3986 allows letters, digits, "+", "-" and
// "." after a leading letter; the dot is excluded here because a dotted
// class name is indistinguishable from a file association.
func validScheme(s string) bool {
	if s == "" || !(s[0] >= 'a' && s[0] <= 'z') {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		ok := (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-'
		if !ok {
			return false
		}
	}
	return true
}

// useScheme switches every derived path to a different scheme.
func useScheme(s string) error {
	if !validScheme(s) {
		return fmt.Errorf("%q is not a usable scheme name. Lower-case "+
			"letters, digits and hyphens, starting with a letter: the name "+
			"becomes a registry key path, so anything else claims a key "+
			"nobody intended", s)
	}
	scheme = s
	return nil
}

func usage() {
	fmt.Fprintf(os.Stderr, `shimp -- SHIMP/1 handler (install half only)

  shimp install [-replace] [-scheme NAME]
  shimp uninstall [-purge] [-scheme NAME]
  shimp verify [-scheme NAME]
  shimp version [-audit]

-scheme installs under a DIFFERENT scheme, which is how a test installation
is kept away from a working one: separate registry class, separate home,
separate Office key, and a different identity hash per s44. Default %[1]s.
-purge also removes the log and the home directory, for a test cycle.

Clicking a %[1]s:// link works only once the DISPATCH half exists. Today a
click is logged and refused, and the log says so.
`, scheme)
}

func main() {
	// The scheme this binary was installed as, which is what a click needs.
	resolveSchemeFromSelf()

	args := os.Args[1:]
	if len(args) == 0 {
		usage()
		os.Exit(2)
	}

	// ⚠️ THE LOG IS OPENED AFTER THE SCHEME IS KNOWN, never before. An
	// earlier version opened it first, which created the PRODUCTION home
	// directory on every invocation including `-scheme shimp-selftest`, so
	// a test command touched production's directory and no -purge could
	// ever clean up after it. Caught by looking at the machine after a test
	// run rather than by any assertion, which is why the live test now
	// checks for it.
	// ⚠️ A URI argument comes FIRST, before any flag parsing, because the
	// registry invokes us as `shimp.exe "%1"` and that argument starts with
	// the scheme rather than a verb. Letting flag.Parse see it would produce
	// "flag provided but not defined" into a console nobody is looking at.
	//
	// Detected by "://" rather than by OUR scheme, so a link for a DIFFERENT
	// installation is recognised as a URI and refused with a reason.
	if strings.Contains(args[0], "://") {
		os.Exit(onClick(args[0]))
	}

	verb, rest := args[0], args[1:]

	// schemeFlag is added to every verb that touches an installation, so a
	// test cycle can name its own and never has to be trusted to remember.
	withScheme := func(fs *flag.FlagSet) *string {
		return fs.String("scheme", scheme,
			"the URI scheme to act on. A different one gives a wholly "+
				"separate installation: its own registry class, home and "+
				"Office key, and a different identity hash (s44)")
	}

	switch verb {
	case "install":
		fs := flag.NewFlagSet("install", flag.ExitOnError)
		replace := fs.Bool("replace", false,
			"replace an installation from a different version. Refused "+
				"without this, because the handler is the trust root (s20)")
		s := withScheme(fs)
		_ = fs.Parse(rest)
		if err := useScheme(*s); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		os.Exit(cmdInstall(*replace))
	case "uninstall":
		fs := flag.NewFlagSet("uninstall", flag.ExitOnError)
		purge := fs.Bool("purge", false,
			"also remove the log and the home directory. For a test cycle, "+
				"which has to leave the machine as it found it")
		s := withScheme(fs)
		_ = fs.Parse(rest)
		if err := useScheme(*s); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		os.Exit(cmdUninstall(*purge))
	case "verify":
		fs := flag.NewFlagSet("verify", flag.ExitOnError)
		s := withScheme(fs)
		_ = fs.Parse(rest)
		if err := useScheme(*s); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		os.Exit(cmdVerify())
	case "version":
		fs := flag.NewFlagSet("version", flag.ExitOnError)
		audit := fs.Bool("audit", false, "the full provenance block (s23)")
		_ = fs.Parse(rest)
		os.Exit(cmdVersion(*audit))
	// dispatch exists for MEASUREMENT, and it is the one verb that sends
	// rather than receives.
	//
	// ⚠️ It is here so that the OS-direct route is an argv template like
	// every other route in linkprobe's table, rather than the one route
	// that needs its own code path. ShellExecuteW is the baseline every
	// other way in is compared against: `cmd /c start`, Start-Process,
	// explorer and a browser all re-parse the URI on the way, and the
	// whole point of the probe matrix is to find out what each of them
	// changes before Windows ever sees it.
	//
	// It does NOT need an installation of its own, so it takes no -scheme:
	// the scheme is whatever the URI says, and if nothing is registered
	// for it ShellExecuteW says so.
	case "dispatch":
		if len(rest) != 1 {
			fmt.Fprintln(os.Stderr,
				"dispatch takes exactly one argument, the URI to open")
			os.Exit(2)
		}
		if err := dispatchURI(rest[0]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	default:
		usage()
		os.Exit(2)
	}
}

// onClick is what the registry invokes. The dispatch half is not built, so
// this does the one useful thing it can: it RECORDS the measurement s36 says
// nothing else can produce.
func onClick(uri string) int {
	// ⚠️ THIS IS s36's MISSING NUMBER. How much of a URI survives the hop
	// from the clicking application, through the registry's "%1", into
	// os.Args[1] is unmeasurable without a registered handler. Every click
	// logs it, so the answer accumulates from real use rather than from a
	// guess about Windows.
	// ⚠️ MEASURING "who creates the command prompt". A Go binary is a
	// console-subsystem executable, so Windows may give it a console when
	// the registry launches it from a GUI context, and that would be a
	// window flashing on every click before shimp has decided anything.
	// Nobody knew, so every click records it.
	present, visible, fg := consoleState()
	_log("click: console present=%v visible=%v foreground=%v",
		present, visible, fg)
	_log("click: os.Args[1] is %d bytes", len(uri))
	_log("click: %s", uri)

	// The parser is built, so a click now gets a real answer about its own
	// link. ⚠️ A refusal here is the GOOD case: under s35 every
	// non-canonical spelling is rejected rather than repaired, so a link
	// that parses is a link with exactly one identity.
	u, err := ParseURI(uri, scheme)
	if err != nil {
		_log("click REFUSED: %v", err)
		return 1
	}

	// Everything the dispatch half will need, recorded now so that the
	// pieces can be checked against a real click before the keeper exists.
	_log("click: identity %s over %d bytes", u.Identity(), len(u.Raw))
	_log("click: would run %q", u.Argv())
	if u.HasAction {
		_log("click: action %q, which is NOT part of identity", u.Action)
	} else {
		_log("click: no action, so this is ensure-running only")
	}

	_log("click REFUSED: the KEEPER is not built. The link is valid and its " +
		"identity is above; s20's lock, IPC and handshake and s46's " +
		"target-side client are what remain.")
	return 1
}

func cmdVersion(audit bool) int {
	p := provenance()
	if !audit {
		fmt.Printf("shimp %s\n", p.Version)
		return 0
	}
	// The block s23 specified, as text rather than JSON, because its reader
	// is a person pasting it into a chat window.
	fmt.Printf("SHIMP/1\nmodule=%s\nversion=%s\nbinary_sha256=%s\ngo=%s\n",
		p.Module, p.Version, p.Sha256, p.GoVersion)
	return 0
}

func readStamp() (stamp, error) {
	var s stamp
	dir, err := home()
	if err != nil {
		return s, err
	}
	raw, err := os.ReadFile(filepath.Join(dir, stampName))
	if err != nil {
		return s, err
	}
	return s, json.Unmarshal(raw, &s)
}

func writeStamp(s stamp) error {
	dir, err := home()
	if err != nil {
		return err
	}
	raw, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, stampName), append(raw, '\n'), 0o644)
}
