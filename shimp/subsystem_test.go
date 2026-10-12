// subsystem_test.go -- the PE offset, pinned against binaries built BOTH ways.
//
// ⚠️ THIS IS THE HOUSE RULE, NOT A PRECAUTION. AGENTS.project.md requires that
// any new syscall struct gets a test pinning a known value, for a measured
// reason: a port read as a plain uint32 from MIB_TCPROW_OWNER_PID turned 9222
// into 30987, which looks like a real port number and so does not announce
// itself as a bug.
//
// A PE header offset has exactly that hazard and worse. The values being read
// are 2 and 3, and small integers appear all over a PE header, so a wrong
// offset can return a plausible answer indefinitely. The only way to know the
// offset is right is to build a binary each way and check the reader
// distinguishes them.
package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestSubsystemReaderDistinguishesBothBuilds(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skipf("UNANSWERED on %s: `go build` here does not produce a PE "+
			"file, so the offset cannot be checked against a real one",
			runtime.GOOS)
	}
	dir := t.TempDir()

	for _, c := range []struct {
		name   string
		ldflag string
		want   uint16
	}{
		{"console subsystem, the default", "", subsystemConsole},
		{"GUI subsystem, what install needs", "-H=windowsgui", subsystemGUI},
	} {
		t.Run(c.name, func(t *testing.T) {
			exe := filepath.Join(dir, c.name+".exe")
			args := []string{"build", "-o", exe}
			if c.ldflag != "" {
				args = append(args, "-ldflags", c.ldflag)
			}
			args = append(args, ".")
			if out, err := exec.Command("go", args...).CombinedOutput(); err != nil {
				t.Fatalf("go build: %v\n%s", err, out)
			}
			got, err := subsystemOf(exe)
			if err != nil {
				t.Fatalf("subsystemOf: %v", err)
			}
			if got != c.want {
				t.Fatalf("read subsystem %d (%s), want %d (%s). The PE offset "+
					"is wrong, and the values being read are small integers "+
					"so a wrong offset can look plausible for a long time.",
					got, subsystemName(got), c.want, subsystemName(c.want))
			}
		})
	}
}

func TestSubsystemOfRefusesSomethingThatIsNotAPEFile(t *testing.T) {
	// A refusal naming the reason, not a zero that reads as an answer. The
	// repo's rule about saying what a tool cannot tell you.
	p := filepath.Join(t.TempDir(), "not-a-binary")
	if err := os.WriteFile(p, []byte("this is not a PE file"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := subsystemOf(p); err == nil {
		t.Fatal("a text file was accepted as a PE binary")
	}
}
