//go:build !windows && !linux && !darwin

// cage_other.go -- there is no cage anywhere but Windows yet.
//
// Windows, Linux and macOS all have one now: AppContainer, Landlock and
// Seatbelt. This is what is left over, which today is nothing anybody runs
// the connector on.
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

// childrenConfinable reports whether a command this connector starts would
// be held by the kernel. Nowhere here, so exec stays shut.
func childrenConfinable() bool { return false }

// cageWrap is the argv prefix that confines a child. Nothing to prepend.
func cageWrap() []string { return nil }

func applyCage(workspace string) (int, bool, error) {
	panic("applyCage on a platform with no cage: guard with cageSupported()")
}

// cageBlocksLoopback is moot where there is no cage.
func cageBlocksLoopback() bool { return false }

// isLoopback is unused off Windows, where there is no cage to be excluded
// from, but main.go refers to it on every platform.
func isLoopback(host string) bool { return false }

// cageAuthority names whatever is enforcing the boundary, for the one line a
// participant is told to look for. There is no cage here.
func cageAuthority() string { return "Nothing" }
