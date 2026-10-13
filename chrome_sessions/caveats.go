// caveats.go -- everything this program reaches for beyond its own process.
//
// WHY THIS FILE EXISTS, AND WHY IT IS NOT A COMMENT. The audit story is that
// somebody pastes a shimp:// link and gets a complete picture of what
// clicking it can do. The code is fetchable at a pinned commit, so it is
// self-verifying: unlike a caveat pasted alongside a link, a caveat in the
// source cannot be swapped by a sender and cannot drift from the code it
// describes, because it is in the same commit.
//
// ⚠️ BUT A DECLARATION NOBODY CHECKS IS WORSE THAN NONE. If this were a
// convention and no more, an auditor finding no caveats could not tell
// "this program reaches for nothing" from "nobody wrote any", and silence
// would read as safety. That is the failure shape this whole repo is
// organised against.
//
// So caveats_test.go VERIFIES it: every outward reach of a known kind in this
// package must appear below, or the test fails naming the one that does not.
// A new exec.Command that nobody declared breaks the build.
//
// ⚠️ AND THE LIMIT IS STATED RATHER THAN GLOSSED. The check is complete with
// respect to the surfaces named in CaveatSurfaces, and no further. A program
// can reach outward through a raw syscall, through reflection, or through a
// helper the checker does not recognise -- and THIS PACKAGE ALREADY DOES, via
// iphlpapi. That reach is declared below because a human noticed it, not
// because the checker found it. An auditor should treat anything outside the
// named surfaces as UNVERIFIED rather than absent.
package main

// Caveats is what this program reaches for beyond its own process.
//
// Read by an auditor, and checked by caveats_test.go.
//
// ⚠️ EACH LINE OPENS WITH ITS SURFACE IN BRACKETS, and that is not
// decoration. An earlier version of the checker asked only whether the FILE
// was mentioned somewhere, and sessions.go was already named for two other
// reaches -- so a deliberately added exec.Command in that file passed. The
// check was per-file when it had to be per-reach, and every file gets
// mentioned eventually, so the checker was on its way to being useless while
// staying green.
var Caveats = []string{
	"[http.Client] connects: HTTP GET to 127.0.0.1:<port>/json/version " +
		"and /json/list, " +
		"for every port named by -ports or found by -discover. Loopback " +
		"only; the host is a constant, not a flag. (sessions.go, getJSON)",

	"[net.DialTimeout] connects: a bare TCP dial to 127.0.0.1:<port>, " +
		"opened and immediately " +
		"closed, to tell a dead port from a wedged one. (sessions.go, " +
		"listening)",

	"[iphlpapi.NewProc] reads system state: the whole machine's TCP " +
		"listener table, through " +
		"iphlpapi!GetExtendedTcpTable, including the owning pid of every " +
		"listening socket on the machine and not only the ports asked " +
		"about. ⚠️ This is the reach least likely to be guessed from the " +
		"program's description. ⚠️ And a correction: an earlier version of " +
		"this line claimed the checker cannot see it. It CAN -- iphlpapi is " +
		"a package-level var, so iphlpapi.NewProc is an ordinary selector. " +
		"What the checker misses is a method call on a value, which was " +
		"this package's HTTP request. (owner_windows.go)",
}

// CaveatsNotReached is stated because absence is only meaningful when it is
// claimed. An auditor can verify each of these against the source.
var CaveatsNotReached = []string{
	"execs nothing. No exec.Command anywhere in the package.",
	"listens on nothing. It is a client only, so no inbound socket exists.",
	"writes no files. Output goes to stdout, and -json changes the format " +
		"rather than the destination.",
	"persists nothing. No registry key, no scheduled task, no startup entry.",
	"reaches no non-loopback address. The host is the constant `host`, not " +
		"a flag, so there is no input that redirects it off the machine.",
}

// CaveatSurfaces is exactly what caveats_test.go knows how to look for, and
// therefore the boundary of what the declaration above can be trusted to be
// complete about.
//
// ⚠️ Anything a program does outside this list is UNDECLARED AND UNDETECTED
// unless a human wrote it down. Raw syscalls are the worked example, in this
// very package.
var CaveatSurfaces = []string{
	// Types as well as functions, because a method call on a value
	// (client.Get) is invisible to an AST-only checker while the TYPE it was
	// made from is not. Found the hard way: see caveats_test.go.
	"http.Client", "exec.Cmd", "net.Dialer",
	"exec.Command", "exec.CommandContext",
	"net.Dial", "net.DialTimeout", "net.Listen",
	"http.Get", "http.Post", "http.NewRequest",
	"os.Create", "os.WriteFile", "os.OpenFile", "os.Remove", "os.RemoveAll",
	"os.MkdirAll", "os.Rename",
}
