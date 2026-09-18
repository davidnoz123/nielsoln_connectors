//go:build !windows

// links_other.go -- the same question, everywhere that is not Windows.
//
// The connector only ever ships as a Windows binary, but it is built and
// vetted on Linux, and a file that does not compile there costs the `go vet`
// in provision.sh. Symlinks are the only reparse points off Windows, and the
// mode bit is the whole answer.
package main

import "os"

func isReparse(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return true // cannot inspect it, so cannot vouch for it
	}
	return isReparseInfo(info)
}

func isReparseInfo(info os.FileInfo) bool {
	return info.Mode()&(os.ModeSymlink|os.ModeIrregular) != 0
}
