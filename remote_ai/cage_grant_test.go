//go:build windows

// The workspace grant must not outlive the session that needed it.
//
// This is the fourth review finding and the most serious: the container name
// was one fixed string, so its SID was the same for every session, and the
// inheritable full-control ACE that enterCage puts on the workspace was never
// removed. Every folder ever shared therefore kept a permanent grant to the
// identity the NEXT session would run as. Since exec is confined by that ACL
// and by nothing else, a later session could read and write an earlier one's
// folder while its own -root said otherwise.
//
// Two things fix it and this file pins both, because either alone leaves a
// hole. The grant is taken back when the session ends, and the identity is per
// session so that a grant a crash left behind names a SID nobody will ever run
// as again.
//
// icacls is used to ask what is really on the folder. A test may shell out;
// the connector may not, which is why the production code sets the ACL through
// the Win32 API instead.
package main

import (
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"unsafe"
)

// sidString renders a SID as S-1-..., which is how icacls names it.
//
// Here rather than in the connector: the shipped code never needs to print a
// SID, and a function that exists only for a test belongs with the test.
var procConvertSidToStringSid = advapi32DLL.NewProc("ConvertSidToStringSidW")

func sidString(sid uintptr) (string, error) {
	var out *uint16
	r, _, err := procConvertSidToStringSid.Call(sid, uintptr(unsafe.Pointer(&out)))
	if r == 0 {
		return "", err
	}
	return syscall.UTF16ToString((*[512]uint16)(unsafe.Pointer(out))[:]), nil
}

func aclOf(t *testing.T, path string) string {
	t.Helper()
	out, err := exec.Command("icacls", path).CombinedOutput()
	if err != nil {
		t.Fatalf("icacls %s: %v: %s", path, err, out)
	}
	return string(out)
}

func TestTheWorkspaceGrantIsTakenBackAtTheEnd(t *testing.T) {
	dir := t.TempDir()
	name := cageNameFor("granttest1", dir)
	sid, err := cageSID(name)
	if err != nil {
		t.Skipf("no container profile available here: %v", err)
	}
	// A test that leaves profiles behind is the litter the old design was
	// wrongly trying to avoid. Cleaning up is cheap when somebody does it.
	t.Cleanup(func() { deleteCageProfile(name) })
	str, err := sidString(sid)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(aclOf(t, dir), str) {
		t.Fatal("the folder already grants this identity before anything ran")
	}

	if err := grantToCage(dir, sid, genericAll); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(aclOf(t, dir), str) {
		t.Fatal("the grant this session depends on was not made")
	}

	// The whole point. Before the fix there was no code that could do this.
	if err := revokeFromCage(dir, sid); err != nil {
		t.Fatal(err)
	}
	if got := aclOf(t, dir); strings.Contains(got, str) {
		t.Fatalf("the grant outlived the session, which is the escape:\n%s", got)
	}
}

func TestTwoSessionsNeverShareAnIdentity(t *testing.T) {
	// The second half of the fix, and the half that still holds when a machine
	// loses power before anything can be taken back. A grant left on a folder
	// must name an identity no later session runs as.
	a := cageNameFor("sessionAAA", "")
	b := cageNameFor("sessionBBB", "")
	if a == b {
		t.Fatalf("two sessions produced one identity: %s", a)
	}

	sidA, err := cageSID(a)
	if err != nil {
		t.Skipf("no container profile available here: %v", err)
	}
	t.Cleanup(func() { deleteCageProfile(a) })
	sidB, err := cageSID(b)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { deleteCageProfile(b) })
	strA, _ := sidString(sidA)
	strB, _ := sidString(sidB)
	if strA == strB {
		t.Fatalf("two names hashed to one SID, so a leak would be inherited: %s",
			strA)
	}
}

func TestAResumedSessionKeepsItsOwnCage(t *testing.T) {
	// The reason the name comes from the pairing token rather than from
	// something random per run: a session that drops and reconnects must land
	// in the SAME cage, or every reconnect leaves another grant behind and the
	// leak this file exists to close comes back by a different route.
	if cageNameFor("KEY123", "") != cageNameFor("key123", "") {
		t.Error("the same session produced two identities across a reconnect")
	}
}
