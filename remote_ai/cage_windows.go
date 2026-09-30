//go:build windows

// cage_windows.go -- the session runs inside a Windows AppContainer, so the
// kernel refuses the participant's files rather than this program remembering
// to.
//
// Why this exists. Everything else in this connector is careful path checking,
// and four independent reviews of the published source found four ways past
// it: an unchecked --record path, a temporary name derived after containment,
// a containment test that folded case on a filesystem that does not, and a
// probe that named a path outside the share. Each was fixed. The pattern is
// the point: a path check is only as wide as the paths it is shown, and the
// next one will be a path nobody thought to show it.
//
// An AppContainer is not shown paths. It is a property of the process token,
// checked by the kernel on the object, and it holds for child processes too.
//
// A running process cannot join one: the container identity is part of the
// token and the token is made when the process is. So this relaunches its own
// executable with the container attributes and waits. Same binary both times,
// which matters because the whole trust argument is that a participant can
// read the source they are about to run.
//
// NO os/exec here, deliberately. The obvious way to set the workspace ACL is
// to shell out to icacls, and every review of this source has checked for
// os/exec and reported it absent. Buying a shorter file with "the connector
// can run arbitrary commands" is a bad trade.
//
// Measured 1 Oct 2026 against a standalone spike of this, each as a
// control/real pair because a refusal that would also have happened without
// the cage proves nothing:
//
//	read ~\Documents          allowed -> DENIED
//	write ~\OneDrive\Desktop  allowed -> DENIED
//	read through a junction   allowed -> DENIED
//	a CHILD process reads     allowed -> DENIED
//	a GRANDCHILD reads        allowed -> DENIED
//	an exe in the workspace   runs    -> runs, and is DENIED
//	reach the network         allowed -> allowed
package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

var (
	userenvDLL  = syscall.NewLazyDLL("userenv.dll")
	advapi32DLL = syscall.NewLazyDLL("advapi32.dll")
	kernel32DLL = syscall.NewLazyDLL("kernel32.dll")

	procCreateAppContainerProfile = userenvDLL.NewProc("CreateAppContainerProfile")
	procDeriveAppContainerSid     = userenvDLL.NewProc("DeriveAppContainerSidFromAppContainerName")
	procDeleteAppContainerProfile = userenvDLL.NewProc("DeleteAppContainerProfile")

	procConvertStringSidToSid = advapi32DLL.NewProc("ConvertStringSidToSidW")
	procGetNamedSecurityInfo  = advapi32DLL.NewProc("GetNamedSecurityInfoW")
	procSetNamedSecurityInfo  = advapi32DLL.NewProc("SetNamedSecurityInfoW")
	procSetEntriesInAcl       = advapi32DLL.NewProc("SetEntriesInAclW")
	procOpenProcessToken      = advapi32DLL.NewProc("OpenProcessToken")
	procGetTokenInformation   = advapi32DLL.NewProc("GetTokenInformation")

	procInitializeProcThreadAttributeList = kernel32DLL.NewProc("InitializeProcThreadAttributeList")
	procUpdateProcThreadAttribute         = kernel32DLL.NewProc("UpdateProcThreadAttribute")
	procDeleteProcThreadAttributeList     = kernel32DLL.NewProc("DeleteProcThreadAttributeList")
	procLocalFree                         = kernel32DLL.NewProc("LocalFree")
)

const (
	// Every name this program makes begins with this, so its own leavings can
	// be told from somebody else's.
	//
	// The name USED to be a single per-machine "nielsoln.bridge.session", on
	// the reasoning that one profile per session would litter the machine.
	// That reasoning was the bug, and an outside review of b6c835a found it.
	// A container name hashes to a SID, the workspace grant is an inheritable
	// ACE for that SID, and nothing removed it. So every folder ever shared
	// kept a permanent full-control grant for the identity that EVERY later
	// session would run as -- and since exec is confined by that ACL rather
	// than by any path check, a later session could read and write every
	// earlier one's folder. Measured on this machine: a folder shared hours
	// before still carried (OI)(CI)(F) for the shared SID.
	//
	// So the identity is per session now. A grant left behind by a crash then
	// names a SID no future session will ever run as, which makes the leak
	// untidy instead of dangerous. Tidiness was the thing being optimised for
	// in the first place, and it was the wrong thing.
	cagePrefix = "nielsoln.bridge."

	// Creating a profile that exists is an error rather than a no-op, so this
	// is the expected result on every run after the first.
	hrAlreadyExists = 0x800700B7

	// An AppContainer has NO network at all without this, measured as curl
	// exit 6, could not resolve host. The connector must dial the bridge, so
	// it is not optional. It is also why the honest word is "prison" and not
	// "sandbox": outbound network stays open by design, and anything the
	// session can read in the workspace can still leave.
	capInternetClient = "S-1-15-3-1"
	seGroupEnabled    = 0x00000004

	tokenQuery          = 0x0008
	tokenIsAppContainer = 29

	procThreadAttrSecurityCapabilities = 0x00020009
	extendedStartupInfoPresent         = 0x00080000

	seFileObject                   = 1
	daclSecurityInfo               = 0x00000004
	grantAccess                    = 1
	revokeAccess                   = 4
	trusteeIsSID                   = 0
	trusteeIsGroup                 = 2
	subContainersAndObjectsInherit = 0x00000003

	genericAll     = 0x10000000
	genericRead    = 0x80000000
	genericExecute = 0x20000000
)

type sidAndAttributes struct {
	sid        uintptr
	attributes uint32
}

type securityCapabilities struct {
	appContainerSid uintptr
	capabilities    uintptr
	capabilityCount uint32
	reserved        uint32
}

type startupInfoEx struct {
	startupInfo   syscall.StartupInfo
	attributeList uintptr
}

type trusteeW struct {
	multipleTrustee          uintptr
	multipleTrusteeOperation uint32
	trusteeForm              uint32
	trusteeType              uint32
	name                     uintptr
}

type explicitAccessW struct {
	accessPermissions uint32
	accessMode        uint32
	inheritance       uint32
	trustee           trusteeW
}

// inCage reports whether THIS process is already inside an AppContainer.
//
// Asked of the token rather than tracked with an environment variable, because
// an environment variable is something this program sets and a token is
// something the kernel knows. It is also what gates `exec`: a capability
// offered on the strength of a variable we set ourselves would be a capability
// offered on our own say-so.
func inCage() bool {
	var token syscall.Handle
	r, _, _ := procOpenProcessToken.Call(
		uintptr(syscall.Handle(^uintptr(0))), // GetCurrentProcess()
		tokenQuery, uintptr(unsafe.Pointer(&token)))
	if r == 0 {
		return false
	}
	defer syscall.CloseHandle(token)

	var isContainer uint32
	var returned uint32
	r, _, _ = procGetTokenInformation.Call(
		uintptr(token), tokenIsAppContainer,
		uintptr(unsafe.Pointer(&isContainer)),
		unsafe.Sizeof(isContainer),
		uintptr(unsafe.Pointer(&returned)))
	return r != 0 && isContainer != 0
}

// cageNameFor is this session's container name.
//
// Derived from the pairing token, so a session that drops and resumes gets the
// SAME cage and the same grant rather than accumulating one per reconnect,
// while two different sessions never share an identity. Where there is no
// token, the workspace path stands in: it is the thing the grant is about.
//
// Only letters and digits survive, because the name becomes a filesystem
// directory under AppData\Local\Packages.
func cageNameFor(token, workspace string) string {
	seed := token
	if seed == "" {
		seed = workspace
	}
	var b strings.Builder
	b.WriteString(cagePrefix)
	for _, r := range seed {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r + 32)
		}
	}
	return b.String()
}

// cageSID returns the container's SID, creating the profile the first time.
func cageSID(cageName string) (uintptr, error) {
	name, err := syscall.UTF16PtrFromString(cageName)
	if err != nil {
		return 0, err
	}
	display, _ := syscall.UTF16PtrFromString("Nielsoln session workspace")
	desc, _ := syscall.UTF16PtrFromString(
		"The one folder a Nielsoln session may read")

	var sid uintptr
	hr, _, _ := procCreateAppContainerProfile.Call(
		uintptr(unsafe.Pointer(name)),
		uintptr(unsafe.Pointer(display)),
		uintptr(unsafe.Pointer(desc)),
		0, 0, uintptr(unsafe.Pointer(&sid)))
	if hr == 0 {
		return sid, nil
	}
	if uint32(hr) != hrAlreadyExists {
		return 0, fmt.Errorf("could not create the container profile (0x%X)",
			uint32(hr))
	}
	hr, _, _ = procDeriveAppContainerSid.Call(
		uintptr(unsafe.Pointer(name)), uintptr(unsafe.Pointer(&sid)))
	if hr != 0 {
		return 0, fmt.Errorf("could not find the container profile (0x%X)",
			uint32(hr))
	}
	return sid, nil
}

// grantToCage adds one ACE letting the container reach *path*.
//
// Through the Win32 API rather than icacls, so this file adds no ability to
// run a program. See the note at the top.
func grantToCage(path string, sid uintptr, rights uint32) error {
	target, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}

	var oldDACL, secDesc uintptr
	r, _, _ := procGetNamedSecurityInfo.Call(
		uintptr(unsafe.Pointer(target)), seFileObject, daclSecurityInfo,
		0, 0, uintptr(unsafe.Pointer(&oldDACL)), 0,
		uintptr(unsafe.Pointer(&secDesc)))
	if r != 0 {
		return fmt.Errorf("could not read the permissions on %s (%d)", path, r)
	}
	defer procLocalFree.Call(secDesc)

	access := explicitAccessW{
		accessPermissions: rights,
		accessMode:        grantAccess,
		inheritance:       subContainersAndObjectsInherit,
		trustee: trusteeW{
			trusteeForm: trusteeIsSID,
			trusteeType: trusteeIsGroup,
			name:        sid,
		},
	}

	var newDACL uintptr
	r, _, _ = procSetEntriesInAcl.Call(1,
		uintptr(unsafe.Pointer(&access)), oldDACL,
		uintptr(unsafe.Pointer(&newDACL)))
	if r != 0 {
		return fmt.Errorf("could not build the permissions for %s (%d)", path, r)
	}
	defer procLocalFree.Call(newDACL)

	r, _, _ = procSetNamedSecurityInfo.Call(
		uintptr(unsafe.Pointer(target)), seFileObject, daclSecurityInfo,
		0, 0, newDACL, 0)
	if r != 0 {
		return fmt.Errorf("could not set the permissions on %s (%d)", path, r)
	}
	return nil
}

// enterCage relaunches this executable inside the container and waits for it,
// returning the child's exit code.
func enterCage(workspace, token string) (int, error) {
	cageName := cageNameFor(token, workspace)
	self, err := os.Executable()
	if err != nil {
		return 0, err
	}
	sid, err := cageSID(cageName)
	if err != nil {
		return 0, err
	}

	// The workspace, read and write. This is the whole grant.
	if err := grantToCage(workspace, sid, genericAll); err != nil {
		return 0, err
	}
	// And this binary, read and execute, or the container cannot start it.
	// Under `go run` the binary lives in %TEMP%\go-build..., which carries no
	// container ACE at all, and that is the route the page recommends.
	if err := grantToCage(self, sid, genericRead|genericExecute); err != nil {
		return 0, fmt.Errorf("the container could not be given access to this "+
			"program, so it could never start it: %w", err)
	}

	var attrSize uintptr
	procInitializeProcThreadAttributeList.Call(0, 1, 0,
		uintptr(unsafe.Pointer(&attrSize)))
	attrList := make([]byte, attrSize)
	r, _, err := procInitializeProcThreadAttributeList.Call(
		uintptr(unsafe.Pointer(&attrList[0])), 1, 0,
		uintptr(unsafe.Pointer(&attrSize)))
	if r == 0 {
		return 0, fmt.Errorf("could not prepare the container: %v", err)
	}
	defer procDeleteProcThreadAttributeList.Call(
		uintptr(unsafe.Pointer(&attrList[0])))

	netSID, err := parseSID(capInternetClient)
	if err != nil {
		return 0, err
	}
	caps := []sidAndAttributes{{sid: netSID, attributes: seGroupEnabled}}
	sc := securityCapabilities{
		appContainerSid: sid,
		capabilities:    uintptr(unsafe.Pointer(&caps[0])),
		capabilityCount: 1,
	}
	r, _, err = procUpdateProcThreadAttribute.Call(
		uintptr(unsafe.Pointer(&attrList[0])), 0,
		procThreadAttrSecurityCapabilities,
		uintptr(unsafe.Pointer(&sc)), unsafe.Sizeof(sc), 0, 0)
	if r == 0 {
		return 0, fmt.Errorf("could not attach the container: %v", err)
	}

	si := startupInfoEx{}
	si.startupInfo.Cb = uint32(unsafe.Sizeof(si))
	si.attributeList = uintptr(unsafe.Pointer(&attrList[0]))
	// The participant must SEE what the confined copy says. A connector whose
	// output disappears is the silent failure this workspace treats as the
	// worst outcome.
	si.startupInfo.Flags = syscall.STARTF_USESTDHANDLES
	si.startupInfo.StdInput, _ = syscall.GetStdHandle(syscall.STD_INPUT_HANDLE)
	si.startupInfo.StdOutput, _ = syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)
	si.startupInfo.StdErr, _ = syscall.GetStdHandle(syscall.STD_ERROR_HANDLE)

	cmdline := syscall.EscapeArg(self)
	for _, a := range os.Args[1:] {
		cmdline += " " + syscall.EscapeArg(a)
	}
	argv, err := syscall.UTF16PtrFromString(cmdline)
	if err != nil {
		return 0, err
	}
	dir, err := syscall.UTF16PtrFromString(workspace)
	if err != nil {
		return 0, err
	}

	var pi syscall.ProcessInformation
	err = syscall.CreateProcess(nil, argv, nil, nil, true,
		extendedStartupInfoPresent|syscall.CREATE_UNICODE_ENVIRONMENT,
		nil, dir, (*syscall.StartupInfo)(unsafe.Pointer(&si)), &pi)
	if err != nil {
		return 0, fmt.Errorf("could not start the confined copy: %v", err)
	}
	defer syscall.CloseHandle(pi.Thread)
	defer syscall.CloseHandle(pi.Process)

	syscall.WaitForSingleObject(pi.Process, syscall.INFINITE)
	var code uint32
	syscall.GetExitCodeProcess(pi.Process, &code)

	// Hand the folder back. This process is the UNCAGED parent, which is the
	// only one that can: the confined copy cannot change permissions on the
	// thing confining it.
	//
	// Best-effort by necessity -- a machine that loses power skips it -- which
	// is why the identity is per session as well. Belt and braces, because a
	// grant that outlives its session is what an outside review of b6c835a
	// found, and one of the two has to hold when the other does not.
	releaseCage(workspace, self, sid, cageName)
	return int(code), nil
}

// releaseCage takes back what enterCage granted, and removes the profile.
func releaseCage(workspace, self string, sid uintptr, cageName string) {
	for _, p := range []string{workspace, self} {
		if err := revokeFromCage(p, sid); err != nil {
			// Said, not swallowed. A grant that quietly outlived its session
			// is the bug this function exists for, so a failure to remove one
			// is the last thing that should be silent.
			logf("WARNING: could not take back this session's access to %s "+
				"(%v). It is granted to an identity no later session uses, so "+
				"nothing inherits it, but it is still there.", p, err)
		}
	}
	deleteCageProfile(cageName)
}

// deleteCageProfile removes the container profile and the directory Windows
// made for it under AppData\Local\Packages.
func deleteCageProfile(cageName string) {
	name, err := syscall.UTF16PtrFromString(cageName)
	if err == nil {
		procDeleteAppContainerProfile.Call(uintptr(unsafe.Pointer(name)))
	}
}

// revokeFromCage removes the ACEs this session added for *sid*.
//
// REVOKE_ACCESS rather than DENY: a deny ACE would be a second permanent mark
// on the participant's folder, which is the shape of the problem rather than
// the fix.
func revokeFromCage(path string, sid uintptr) error {
	target, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	var oldDACL, secDesc uintptr
	r, _, _ := procGetNamedSecurityInfo.Call(
		uintptr(unsafe.Pointer(target)), seFileObject, daclSecurityInfo,
		0, 0, uintptr(unsafe.Pointer(&oldDACL)), 0,
		uintptr(unsafe.Pointer(&secDesc)))
	if r != 0 {
		return fmt.Errorf("could not read the permissions (%d)", r)
	}
	defer procLocalFree.Call(secDesc)

	access := explicitAccessW{
		accessMode: revokeAccess,
		trustee: trusteeW{
			trusteeForm: trusteeIsSID,
			trusteeType: trusteeIsGroup,
			name:        sid,
		},
	}
	var newDACL uintptr
	r, _, _ = procSetEntriesInAcl.Call(1,
		uintptr(unsafe.Pointer(&access)), oldDACL,
		uintptr(unsafe.Pointer(&newDACL)))
	if r != 0 {
		return fmt.Errorf("could not build the permissions (%d)", r)
	}
	defer procLocalFree.Call(newDACL)

	r, _, _ = procSetNamedSecurityInfo.Call(
		uintptr(unsafe.Pointer(target)), seFileObject, daclSecurityInfo,
		0, 0, newDACL, 0)
	if r != 0 {
		return fmt.Errorf("could not set the permissions (%d)", r)
	}
	return nil
}

func parseSID(s string) (uintptr, error) {
	p, err := syscall.UTF16PtrFromString(s)
	if err != nil {
		return 0, err
	}
	var sid uintptr
	r, _, err := procConvertStringSidToSid.Call(
		uintptr(unsafe.Pointer(p)), uintptr(unsafe.Pointer(&sid)))
	if r == 0 {
		return 0, fmt.Errorf("could not read the capability %s: %v", s, err)
	}
	return sid, nil
}

func cageSupported() bool { return true }

// isLoopback reports whether *host* names this machine.
//
// A confined session cannot reach it, so the cage and a bridge here are
// mutually exclusive. Answered by parsing rather than by string comparison:
// 127.0.0.1 is not the only loopback address, and ::1 and 127.0.0.2 are both
// reachable spellings of the same machine.
func isLoopback(host string) bool {
	if host == "" || host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
