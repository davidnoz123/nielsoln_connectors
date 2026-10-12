//go:build windows

// dispatch_windows.go -- hand a URI to the OS exactly as a click does.
//
// WHY THIS EXISTS, AND IT IS NOT A CONVENIENCE. s36 records that the number
// which actually bounds a shimp URI is what the handler receives as
// os.Args[1], and that it cannot be measured without a registered handler.
// ShellExecuteW routes a URI THROUGH THE REGISTRY, which is the same path a
// click takes, so the live test can take that measurement instead of
// simulating it.
//
// ⚠️ AND THE LIMIT OF WHAT IT MEASURES, stated rather than glossed. This
// answers the OS hop: scheme lookup, the "%1" substitution, CreateProcess,
// os.Args[1]. It says NOTHING about what Excel or a browser does to the URI
// BEFORE handing it over, which is the other half of s36's question and the
// half that still needs a human to click something.
//
// Raw shell32 rather than `cmd /c start`, for the reason
// pdf_view_server/raise_windows.go records about PowerShell's AppActivate,
// and because a shell would re-parse the URI and so measure the shell rather
// than the registry.
package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	shell32             = syscall.NewLazyDLL("shell32.dll")
	procShellExecuteW   = shell32.NewProc("ShellExecuteW")
	swShowNormal        = 1
	shellExecuteSuccess = uintptr(32) // anything above 32 is success
)

// dispatchURI asks the OS to open uri, which routes through whatever is
// registered for its scheme.
func dispatchURI(uri string) error {
	op, err := syscall.UTF16PtrFromString("open")
	if err != nil {
		return fmt.Errorf("cannot encode the verb: %w", err)
	}
	target, err := syscall.UTF16PtrFromString(uri)
	if err != nil {
		return fmt.Errorf("the URI is not representable as UTF-16: %w", err)
	}
	// ⚠️ ShellExecuteW's return is NOT an HRESULT. It is an HINSTANCE-shaped
	// value where anything <= 32 is an error code, which is a convention old
	// enough to be easy to get backwards: treating it as "0 means failure"
	// would report every success as a failure.
	r, _, _ := procShellExecuteW.Call(0,
		uintptr(unsafe.Pointer(op)), uintptr(unsafe.Pointer(target)),
		0, 0, uintptr(swShowNormal))
	if r <= shellExecuteSuccess {
		return fmt.Errorf("ShellExecuteW refused the URI with code %d, which "+
			"usually means no handler is registered for its scheme", r)
	}
	return nil
}
