// subsystem.go -- which PE subsystem this binary was linked as.
//
// WHY A BINARY HAS TO ASK ITSELF THIS. s50 chose a GUI-subsystem handler,
// because a console-subsystem one is given a console window by Windows on
// every click and that window takes FOREGROUND, measured. So the documented
// install line carries `-ldflags "-H=windowsgui"`.
//
// ⚠️ AND THAT IS THE PROBLEM THIS FILE EXISTS FOR. shimp installs a COPY of
// whatever binary ran `install`, so the subsystem of the handler is decided
// by the flags on the command somebody typed. Omit the flag and the install
// still SUCCEEDS, still registers, still works, and steals focus on every
// click for ever. A silent degradation in the install instructions is exactly
// the failure shape this repo is organised against.
//
// So the binary reads its own PE header and `verify` reports it. The forgotten
// flag becomes a thing somebody is told about rather than a thing they live
// with.
//
// No build tag: a PE file is a PE file, and reading one is the same
// everywhere. What differs is whether the answer MATTERS, which is why
// `verify` only judges it on Windows.
package main

import (
	"encoding/binary"
	"fmt"
	"os"
)

// The two values that occur here. The full set is larger and irrelevant.
const (
	subsystemGUI     = 2 // IMAGE_SUBSYSTEM_WINDOWS_GUI
	subsystemConsole = 3 // IMAGE_SUBSYSTEM_WINDOWS_CUI
)

// subsystemOf reads the Subsystem field from a PE file.
//
// ⚠️ THE OFFSET IS THE DANGEROUS PART, in the way `AGENTS.project.md` records
// for syscall structs: a port read as a plain uint32 turned 9222 into 30987,
// which looks like a real port number and so does not announce itself. A
// wrong offset here returns some other field's value, and 2 or 3 appear all
// over a PE header, so a wrong read can easily look like a right answer.
// subsystem_test.go therefore pins it against binaries built BOTH ways.
//
// The layout, counted rather than copied:
//
//	0x00         "MZ"
//	0x3C         uint32 e_lfanew, the offset of the PE signature
//	e_lfanew     "PE\0\0"            4 bytes
//	+4           COFF file header    20 bytes
//	+24          optional header starts here
//	+24+68       Subsystem, uint16
//
// Subsystem sits at optional-header offset 68 for PE32 AND for PE32+, which
// is worth stating because it looks like it should differ: PE32+ drops
// BaseOfData (4 bytes) and widens ImageBase (4 more), so the two changes
// cancel exactly.
func subsystemOf(path string) (uint16, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	var mz [2]byte
	if _, err := f.ReadAt(mz[:], 0); err != nil {
		return 0, fmt.Errorf("cannot read the DOS header of %s: %w", path, err)
	}
	if mz[0] != 'M' || mz[1] != 'Z' {
		return 0, fmt.Errorf("%s is not a PE file: no MZ signature", path)
	}

	var lfa [4]byte
	if _, err := f.ReadAt(lfa[:], 0x3C); err != nil {
		return 0, fmt.Errorf("cannot read e_lfanew of %s: %w", path, err)
	}
	peAt := int64(binary.LittleEndian.Uint32(lfa[:]))

	var sig [4]byte
	if _, err := f.ReadAt(sig[:], peAt); err != nil {
		return 0, fmt.Errorf("cannot read the PE signature of %s: %w", path, err)
	}
	if sig != [4]byte{'P', 'E', 0, 0} {
		return 0, fmt.Errorf("%s has no PE signature where e_lfanew points, "+
			"so the header offset is wrong rather than the file", path)
	}

	var sub [2]byte
	if _, err := f.ReadAt(sub[:], peAt+24+68); err != nil {
		return 0, fmt.Errorf("cannot read the Subsystem field of %s: %w", path, err)
	}
	return binary.LittleEndian.Uint16(sub[:]), nil
}

// subsystemName is for a report a human reads.
func subsystemName(s uint16) string {
	switch s {
	case subsystemGUI:
		return "GUI"
	case subsystemConsole:
		return "console"
	default:
		return fmt.Sprintf("unknown (%d)", s)
	}
}
