//go:build windows

// links_windows.go -- is this path a reparse point?
//
// Split out by build tag because the only reliable answer on Windows comes
// from the raw file attributes, and syscall.Win32FileAttributeData does not
// exist anywhere else. `go vet` on a Linux box would not compile otherwise,
// and that check is worth keeping.
//
// Why the raw attribute rather than the mode bits. Go decides ModeSymlink
// from the reparse TAG, which means a second call, and a FileInfo obtained
// from ReadDir does not always carry it -- a junction then arrives looking
// like an ordinary directory. FILE_ATTRIBUTE_REPARSE_POINT is in the
// directory scan itself and is set for every reparse point whatever its tag:
// symlinks, junctions, mount points, and the OneDrive placeholders that sit
// in a great many Documents folders.
package main

import (
	"os"
	"syscall"
)

const fileAttributeReparsePoint = 0x400

// isReparse reports whether path is a link of any kind, without following it.
func isReparse(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		// Cannot tell, so assume the worst. A path we cannot inspect is not
		// one we can prove is inside the shared folder.
		return true
	}
	return isReparseInfo(info)
}

func isReparseInfo(info os.FileInfo) bool {
	if info.Mode()&(os.ModeSymlink|os.ModeIrregular) != 0 {
		return true
	}
	if d, ok := info.Sys().(*syscall.Win32FileAttributeData); ok {
		return d.FileAttributes&fileAttributeReparsePoint != 0
	}
	return false
}
