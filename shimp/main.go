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
//	go run github.com/davidnoz123/nielsoln_connectors/shimp@<sha> install
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
// ⚠️ The TRAILING COLON in Office's trusted-protocol key is a separate
// constant below, because a key named "shimp" looks right in regedit and does
// nothing. See s8.
const scheme = "shimp"

// logName is under the shimp home, and it is not optional.
//
// ⚠️ A handler invoked from a URI click has NO CONSOLE. Anything written to
// stderr goes nowhere, the process exits, and the only symptom is that
// clicking the link did nothing. That is the exact failure mode the house
// rules are organised against, so every run appends here, including the runs
// that succeed, because "it logged starting and never finished" is what tells
// the next person where it died.
const logName = "shimp.log"

// stampName records what install knew at install time and nothing else does
// afterwards. s23: the second audit question, "is what is installed the thing
// I think it is", cannot be answered without it.
const stampName = "installed.json"

var logTo io.Writer // set by main; tests replace it

func _log(format string, a ...any) {
	line := fmt.Sprintf("%s  %s\n", time.Now().Format(time.RFC3339), fmt.Sprintf(format, a...))
	if logTo != nil {
		_, _ = io.WriteString(logTo, line)
		return
	}
	// Before the home exists there is nowhere to append, so stderr is better
	// than nothing even though a click cannot see it.
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

// openLog appends to the log, creating the home if needed.
func openLog() (*os.File, error) {
	dir, err := home()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("cannot create %s: %w", dir, err)
	}
	return os.OpenFile(filepath.Join(dir, logName),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
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

func usage() {
	fmt.Fprintf(os.Stderr, `shimp -- SHIMP/1 handler (install half only)

  shimp install [-replace]   place the binary and claim %[1]s:
  shimp uninstall            withdraw only what install created
  shimp verify               report what is true of this machine
  shimp version [-audit]     this handler's own provenance

Clicking a %[1]s:// link works only once the DISPATCH half exists. Today a
click is logged and refused, and the log says so.
`, scheme)
}

func main() {
	if f, err := openLog(); err == nil {
		defer f.Close()
		logTo = f
	}

	args := os.Args[1:]
	if len(args) == 0 {
		usage()
		os.Exit(2)
	}

	// ⚠️ A URI argument comes FIRST, before any flag parsing, because the
	// registry invokes us as `shimp.exe "%1"` and that argument starts with
	// the scheme rather than a verb. Letting flag.Parse see it would produce
	// "flag provided but not defined" into a console nobody is looking at.
	if strings.HasPrefix(strings.ToLower(args[0]), scheme+":") {
		os.Exit(onClick(args[0]))
	}

	verb, rest := args[0], args[1:]
	switch verb {
	case "install":
		fs := flag.NewFlagSet("install", flag.ExitOnError)
		replace := fs.Bool("replace", false,
			"replace an installation from a different version. Refused "+
				"without this, because the handler is the trust root (s20)")
		_ = fs.Parse(rest)
		os.Exit(cmdInstall(*replace))
	case "uninstall":
		os.Exit(cmdUninstall())
	case "verify":
		os.Exit(cmdVerify())
	case "version":
		fs := flag.NewFlagSet("version", flag.ExitOnError)
		audit := fs.Bool("audit", false, "the full provenance block (s23)")
		_ = fs.Parse(rest)
		os.Exit(cmdVersion(*audit))
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
	_log("click: os.Args[1] is %d bytes", len(uri))
	_log("click: %s", uri)
	_log("click REFUSED: the dispatch half is not built. The grammar is " +
		"frozen (s43, s44, s45, s41) and no parser exists yet.")
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
