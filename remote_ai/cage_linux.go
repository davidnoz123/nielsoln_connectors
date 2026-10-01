//go:build linux

// cage_linux.go -- the session confines itself with Landlock, so the kernel
// refuses the participant's files rather than this program remembering to.
//
// A different shape from the Windows cage, and a simpler one. Landlock is
// SELF-APPLIED: this process restricts itself, every child inherits the
// restriction, and it cannot be lifted afterwards. So there is no relaunch, no
// security identity, no permission written onto the participant's folder and
// nothing to clean up.
//
// That absence matters. Two of the six findings against the Windows cage were
// about a grant outliving its session and an identity a later session could
// land on again. Neither can happen here, because neither a grant nor an
// identity is created. The same job, with less to get wrong.
//
// Raw syscalls rather than github.com/landlock-lsm/go-landlock: go.mod has no
// requires, and the trust argument is that the security-critical code is short
// enough for a participant to read.
//
// Measured 1 Oct 2026 in WSL2 (kernel 6.6, ABI v3), as control/real pairs:
//
//	workspace read and write   allowed -> allowed
//	read /etc/ssh              allowed -> DENIED
//	read /var/log              allowed -> DENIED
//	a CHILD process reads      allowed -> DENIED
//	a GRANDCHILD reads         allowed -> DENIED
//	reach the network          allowed -> allowed
//
// And one platform limit found the same way: on a Windows folder mounted into
// WSL (/mnt/c, v9fs) the rule does not take, and the session is denied its own
// workspace. Fail-closed rather than an escape, but it would look like a broken
// connector, so confineHere checks for it and says so.
package main

import (
	"fmt"
	"net"
	"os"
	"syscall"
	"unsafe"
)

const (
	sysLandlockCreateRuleset = 444
	sysLandlockAddRule       = 445
	sysLandlockRestrictSelf  = 446

	prSetNoNewPrivs = 38

	// Names a file without reading it, which is what a rule wants. Not in the
	// Go standard library.
	oPath = 0x200000

	ruleTypePathBeneath = 1

	accExecute    = 1 << 0
	accWriteFile  = 1 << 1
	accReadFile   = 1 << 2
	accReadDir    = 1 << 3
	accRemoveDir  = 1 << 4
	accRemoveFile = 1 << 5
	accMakeChar   = 1 << 6
	accMakeDir    = 1 << 7
	accMakeReg    = 1 << 8
	accMakeSock   = 1 << 9
	accMakeFifo   = 1 << 10
	accMakeBlock  = 1 << 11
	accMakeSym    = 1 << 12
	accRefer      = 1 << 13 // ABI v2
	accTruncate   = 1 << 14 // ABI v3

	// Everything the kernel's own ABI can police. Anything left OUT of this
	// stays permitted for ever, so under-declaring is how a cage quietly has a
	// hole in it rather than an error.
	handledV1 = accExecute | accWriteFile | accReadFile | accReadDir |
		accRemoveDir | accRemoveFile | accMakeChar | accMakeDir | accMakeReg |
		accMakeSock | accMakeFifo | accMakeBlock | accMakeSym
	handledV2 = handledV1 | accRefer
	handledV3 = handledV2 | accTruncate

	// v9fs, the filesystem WSL uses for a Windows drive mounted at /mnt/c.
	// Landlock rules do not take on it: the session ends up denied the very
	// folder it was granted. Measured, not read about.
	magicV9FS = 0x01021997

	// Rights that mean something on a plain file. A rule asking for READ_DIR
	// on one is rejected, and a single such rule makes the WHOLE ruleset fail
	// to build -- measured as "add_rule /etc/resolv.conf: invalid argument".
	fileRights = accReadFile | accWriteFile | accExecute | accTruncate
)

// systemPaths are read-only grants without which nothing runs.
//
// NAMED FILES under /etc rather than the whole of it. Granting /etc wholesale
// was the first attempt and handed over the machine's ssh host keys: measured
// by probing /etc/ssh inside the cage and finding it readable. A resolver needs
// a few files and a certificate bundle, not the machine's secrets.
//
// /dev and /proc are not optional. Without them TLS and the resolver have no
// randomness and no configuration, and the first caged run could not reach the
// network at all.
var systemPaths = []string{
	"/usr", "/bin", "/sbin", "/lib", "/lib64", "/lib32",
	"/dev", "/proc",
	"/etc/resolv.conf", "/etc/hosts", "/etc/nsswitch.conf", "/etc/localtime",
	"/etc/ssl", "/etc/pki", "/etc/ca-certificates", "/etc/ssl/certs",
}

// landlockABI is what this kernel supports, or 0 for none.
func landlockABI() int {
	r, _, _ := syscall.Syscall(sysLandlockCreateRuleset, 0, 0, 1)
	if int(r) < 0 {
		return 0
	}
	return int(r)
}

func handledFor(v int) uint64 {
	switch {
	case v >= 3:
		return handledV3
	case v == 2:
		return handledV2
	default:
		return handledV1
	}
}

func cageSupported() bool { return landlockABI() >= 1 }

// cageBlocksLoopback reports whether confining this session would cut it off
// from a bridge on this same machine.
//
// False here. Landlock gained network rules at ABI v4 and this asks for none
// of them, so a confined session can still reach a local bridge. The Windows
// cage cannot, which is why that platform stands down for a local bridge and
// this one does not.
func cageBlocksLoopback() bool { return false }

// inCage reports whether this process is confined, by ASKING rather than by
// remembering.
//
// Landlock has no introspection: there is no call that answers "am I
// restricted?". The alternative to measuring would be a package variable set
// when restrict_self succeeded, which is the connector's own say-so -- and
// `exec` is gated on this answer, so its own say-so is exactly what it must
// not be.
//
// So it opens "/", which is world-readable on every Linux and is deliberately
// NOT granted. Being refused it means something refused it, and the only thing
// that would is the cage.
func inCage() bool {
	fd, err := syscall.Open("/", syscall.O_RDONLY|syscall.O_DIRECTORY, 0)
	if err != nil {
		return true
	}
	syscall.Close(fd)
	return false
}

// applyCage confines this process in place and returns done=false, meaning
// carry on in this same process.
//
// Windows has to relaunch, because a container identity is fixed when a token
// is made. Landlock restricts the running process, so there is no second copy
// and no parent left outside holding permissions it has to remember to hand
// back.
func applyCage(workspace string) (int, bool, error) {
	v := landlockABI()
	if v < 1 {
		return 0, false, fmt.Errorf("this kernel has no landlock")
	}
	if cageWouldCripple(workspace) {
		return 0, false, fmt.Errorf("a folder on a Windows drive mounted into "+
			"WSL cannot be confined by Landlock, and confining it would leave "+
			"this session unable to read %s at all. Share a folder on the "+
			"Linux side to have the boundary enforced", workspace)
	}
	if err := confineHere(v, workspace); err != nil {
		return 0, false, err
	}
	return 0, false, nil
}

// cageWouldCripple reports whether confining a session on this workspace
// would leave it unable to read the workspace itself.
//
// Checked BEFORE the cage goes on, because Landlock cannot be lifted. Get this
// wrong and the session is confined, useless, and still advertising `exec`
// because inCage() quite correctly says yes. Better to run uncaged and say so:
// the participant gets a working session and an honest warning rather than a
// dead one.
//
// A named filesystem rather than a general rule, because one is measured and
// the other would be a guess. Anything else that misbehaves is caught after the
// fact by the check at the end of confineHere, which is fatal.
func cageWouldCripple(workspace string) bool {
	var st syscall.Statfs_t
	if err := syscall.Statfs(workspace, &st); err != nil {
		return false
	}
	return st.Type == magicV9FS
}

func confineHere(v int, workspace string) error {
	// Without this the kernel refuses to restrict an unprivileged process at
	// all, because a setuid binary could otherwise be used to shed the cage.
	if _, _, e := syscall.Syscall6(syscall.SYS_PRCTL, prSetNoNewPrivs,
		1, 0, 0, 0, 0); e != 0 {
		return fmt.Errorf("could not drop privilege escalation: %v", e)
	}

	handled := handledFor(v)
	attr := struct{ handledAccessFS uint64 }{handled}
	fd, _, e := syscall.Syscall(sysLandlockCreateRuleset,
		uintptr(unsafe.Pointer(&attr)), unsafe.Sizeof(attr), 0)
	if e != 0 {
		return fmt.Errorf("could not build a ruleset: %v", e)
	}
	defer syscall.Close(int(fd))

	if err := addRule(fd, workspace, handled); err != nil {
		return err
	}
	for _, p := range systemPaths {
		if err := addRule(fd, p, accExecute|accReadFile|accReadDir); err != nil {
			return err
		}
	}

	if _, _, e := syscall.Syscall(sysLandlockRestrictSelf, fd, 0, 0); e != 0 {
		return fmt.Errorf("could not apply the ruleset: %v", e)
	}

	// The workspace must still be reachable from inside its own cage. On a
	// Windows folder mounted into WSL (/mnt/c, v9fs) the rule does not take and
	// the session is denied the very folder it was given. Nothing escapes, but
	// a connector that cannot read its own workspace looks broken rather than
	// careful, so it says which it is.
	if _, err := os.ReadDir(workspace); err != nil {
		// FATAL, not a warning. The cage is on and cannot be taken off, so
		// carrying on means a session that can read nothing while inCage()
		// truthfully reports it is confined -- which would also offer `exec`.
		// There is no useful state left to continue into.
		fatal("This session was confined, and then could not read %s inside "+
			"its own boundary (%v). Nothing has been shared. A confinement "+
			"cannot be undone once applied, so this stops rather than "+
			"continuing as a session that can read nothing.", workspace, err)
	}
	return nil
}

func addRule(ruleset uintptr, path string, rights uint64) error {
	fd, err := syscall.Open(path, oPath|syscall.O_CLOEXEC, 0)
	if err != nil {
		return nil // not on this distribution; nothing to grant
	}
	defer syscall.Close(fd)

	// Rights must suit what they name. See fileRights.
	var st syscall.Stat_t
	if err := syscall.Fstat(fd, &st); err == nil &&
		st.Mode&syscall.S_IFMT != syscall.S_IFDIR {
		rights &= fileRights
	}
	if rights == 0 {
		return nil
	}

	// struct landlock_path_beneath_attr is PACKED: __u64 then __s32, twelve
	// bytes. The equivalent Go struct is sixteen, and the kernel rejects it as
	// the wrong size.
	var buf [12]byte
	*(*uint64)(unsafe.Pointer(&buf[0])) = rights
	*(*int32)(unsafe.Pointer(&buf[8])) = int32(fd)
	if _, _, e := syscall.Syscall6(sysLandlockAddRule, ruleset,
		ruleTypePathBeneath, uintptr(unsafe.Pointer(&buf[0])), 0, 0, 0); e != 0 {
		return fmt.Errorf("could not grant %s: %v", path, e)
	}
	return nil
}

// isLoopback reports whether *host* names this machine.
func isLoopback(host string) bool {
	if host == "" || host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// cageAuthority names whatever is enforcing the boundary, for the one line a
// participant is told to look for. Landlock.
func cageAuthority() string { return "Linux" }
