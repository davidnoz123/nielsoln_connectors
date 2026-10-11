//go:build windows

// scheme_windows.go -- claiming a URL scheme through advapi32.
//
// WHY RAW SYSCALL RATHER THAN SHELLING OUT. `reg.exe add` would be four lines
// shorter and the repo already records what that costs: pdf_view_server's
// first raise shelled out to PowerShell's AppActivate at 330ms per call,
// spawning a whole runtime for something the platform API does in
// microseconds. It also keeps the zero third-party dependency property, since
// the obvious alternative, golang.org/x/sys/windows/registry, is a module.
//
// ⚠️ AND THE WIDE-STRING LENGTHS ARE THE DANGEROUS PART, in the same way
// chrome_sessions/owner_windows.go records a port in network byte order
// reading as 30987 instead of 9222. cbData is a count of BYTES including the
// terminating NUL, and getting it wrong does not fail: it writes a value that
// looks right in regedit and is subtly truncated. scheme_windows_test.go pins
// the conversion against known bytes for exactly that reason.
package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	advapi32            = syscall.NewLazyDLL("advapi32.dll")
	procRegCreateKeyExW = advapi32.NewProc("RegCreateKeyExW")
	procRegOpenKeyExW   = advapi32.NewProc("RegOpenKeyExW")
	procRegSetValueExW  = advapi32.NewProc("RegSetValueExW")
	procRegQueryValueEx = advapi32.NewProc("RegQueryValueExW")
	procRegDeleteKeyW   = advapi32.NewProc("RegDeleteKeyW")
	procRegCloseKey     = advapi32.NewProc("RegCloseKey")
)

const (
	// HKEY_CURRENT_USER. Per-user, so no elevation, which is s30: a
	// capability is a thing a person installed.
	hkeyCurrentUser = 0x80000001

	keyRead  = 0x20019
	keyWrite = 0x20006

	regSZ = 1

	errSuccess       = 0
	errFileNotFound  = 2
	errMoreData      = 234
	regOptionNonVol  = 0
	maxValueBytes    = 1 << 16 // a registry string this long is already wrong
	dispositionUnset = 0
)

type winClaims struct{}

func openClaims() (claims, error) { return winClaims{}, nil }

func (winClaims) where() string { return `HKEY_CURRENT_USER (live registry)` }

// status turns an LSTATUS into an error naming the call, because "the install
// failed" with no API name is the kind of report that costs an hour.
func status(call string, r uintptr) error {
	if r == errSuccess {
		return nil
	}
	return fmt.Errorf("%s returned %d: %w", call, r, syscall.Errno(r))
}

func (winClaims) set(subkey, name, value string) error {
	sub, err := syscall.UTF16PtrFromString(subkey)
	if err != nil {
		return fmt.Errorf("subkey %q is not representable: %w", subkey, err)
	}
	var h uintptr
	r, _, _ := procRegCreateKeyExW.Call(
		uintptr(hkeyCurrentUser), uintptr(unsafe.Pointer(sub)), 0, 0,
		regOptionNonVol, keyWrite, 0,
		uintptr(unsafe.Pointer(&h)), dispositionUnset)
	if err := status("RegCreateKeyExW", r); err != nil {
		return err
	}
	defer procRegCloseKey.Call(h)

	// An empty name is the key's DEFAULT value, which is what Windows reads
	// for a class and for a shell command.
	var np *uint16
	if name != "" {
		if np, err = syscall.UTF16PtrFromString(name); err != nil {
			return fmt.Errorf("value name %q is not representable: %w", name, err)
		}
	}

	// ⚠️ UTF16FromString appends the terminating NUL, and cbData counts
	// BYTES, so it is len(u)*2 and never len(value). An empty string is one
	// UTF-16 unit, the NUL, so two bytes.
	u, err := syscall.UTF16FromString(value)
	if err != nil {
		return fmt.Errorf("value %q is not representable: %w", value, err)
	}
	r, _, _ = procRegSetValueExW.Call(h, uintptr(unsafe.Pointer(np)), 0, regSZ,
		uintptr(unsafe.Pointer(&u[0])), uintptr(len(u)*2))
	return status("RegSetValueExW", r)
}

func (winClaims) hasKey(subkey string) (bool, error) {
	sub, err := syscall.UTF16PtrFromString(subkey)
	if err != nil {
		return false, fmt.Errorf("subkey %q is not representable: %w", subkey, err)
	}
	var h uintptr
	r, _, _ := procRegOpenKeyExW.Call(uintptr(hkeyCurrentUser),
		uintptr(unsafe.Pointer(sub)), 0, keyRead, uintptr(unsafe.Pointer(&h)))
	if r == errFileNotFound {
		return false, nil
	}
	if err := status("RegOpenKeyExW", r); err != nil {
		return false, err
	}
	procRegCloseKey.Call(h)
	return true, nil
}

func (w winClaims) get(subkey, name string) (string, bool, error) {
	sub, err := syscall.UTF16PtrFromString(subkey)
	if err != nil {
		return "", false, fmt.Errorf("subkey %q is not representable: %w", subkey, err)
	}
	var h uintptr
	r, _, _ := procRegOpenKeyExW.Call(uintptr(hkeyCurrentUser),
		uintptr(unsafe.Pointer(sub)), 0, keyRead, uintptr(unsafe.Pointer(&h)))
	if r == errFileNotFound {
		return "", false, nil
	}
	if err := status("RegOpenKeyExW", r); err != nil {
		return "", false, err
	}
	defer procRegCloseKey.Call(h)

	var np *uint16
	if name != "" {
		if np, err = syscall.UTF16PtrFromString(name); err != nil {
			return "", false, fmt.Errorf("value name %q is not representable: %w", name, err)
		}
	}

	// Two passes: ask the size, then read it. A fixed buffer would either
	// truncate a long command silently or over-allocate on every call, and
	// the first of those is the failure this repo keeps naming.
	var size uint32
	r, _, _ = procRegQueryValueEx.Call(h, uintptr(unsafe.Pointer(np)), 0, 0, 0,
		uintptr(unsafe.Pointer(&size)))
	if r == errFileNotFound {
		return "", false, nil
	}
	if err := status("RegQueryValueExW(size)", r); err != nil && r != errMoreData {
		return "", false, err
	}
	if size == 0 {
		return "", true, nil // present and empty, which is a real state here
	}
	if size > maxValueBytes {
		return "", true, fmt.Errorf("value at %s is %d bytes, which is beyond "+
			"anything this program writes, so it is not read", subkey, size)
	}
	buf := make([]uint16, (size+1)/2)
	r, _, _ = procRegQueryValueEx.Call(h, uintptr(unsafe.Pointer(np)), 0, 0,
		uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)))
	if err := status("RegQueryValueExW(read)", r); err != nil {
		return "", false, err
	}
	return syscall.UTF16ToString(buf), true, nil
}

func (winClaims) removeKey(subkey string) error {
	sub, err := syscall.UTF16PtrFromString(subkey)
	if err != nil {
		return fmt.Errorf("subkey %q is not representable: %w", subkey, err)
	}
	// ⚠️ RegDeleteKeyW deletes a key only when it has no subkeys, which is
	// why install.go removes them deepest-first rather than trusting a
	// recursive delete that does not exist here.
	r, _, _ := procRegDeleteKeyW.Call(uintptr(hkeyCurrentUser),
		uintptr(unsafe.Pointer(sub)))
	if r == errFileNotFound {
		return fmt.Errorf("%s was not present", subkey)
	}
	return status("RegDeleteKeyW", r)
}
