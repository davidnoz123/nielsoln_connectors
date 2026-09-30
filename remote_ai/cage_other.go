//go:build !windows

// cage_other.go -- there is no cage anywhere but Windows yet.
//
// Linux is the next one and the cheapest: Landlock restricts a process and
// every child it goes on to create, cannot be lifted once applied, and does
// not touch networking, so the connector could confine ITSELF in place with no
// relaunch at all. macOS is the awkward one, because Apple's supported model
// is entitlement-based and sandbox_init is deprecated.
//
// Saying so out loud matters more than it looks. The page tells participants
// which platforms the cage covers, and a stub that quietly reported success
// here would make that page lie on exactly the platforms with no protection.
package main

// inCage reports whether this process is confined. Nowhere but Windows is,
// yet, and answering anything else would gate `exec` open on a platform with
// no boundary behind it.
func inCage() bool { return false }

// cageSupported reports whether a cage can be built on this platform at all.
// False here means the connector runs exactly as it did before: the file
// operations, their path checking, and no exec.
func cageSupported() bool { return false }

func enterCage(workspace string) (int, error) {
	panic("enterCage on a platform with no cage: guard with cageSupported()")
}

// isLoopback is unused off Windows, where there is no cage to be excluded
// from, but main.go refers to it on every platform.
func isLoopback(host string) bool { return false }
