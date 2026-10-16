// focus -- who may raise a window to the foreground, and when.
//
//	go run . -mode timeout            SPI_GETFOREGROUNDLOCKTIMEOUT
//	go run . -mode state              who holds the foreground right now
//	go run . -mode target             allocate a window, print its HWND, wait
//	go run . -mode raise -hwnd H -delay N
//	go run . -mode grant-and-run -hwnd H -delay N
//
// WHAT IT MEASURED, 15 October 2026
//
//	SPI_GETFOREGROUNDLOCKTIMEOUT = 2147483647 ms  (MAX_INT, ~24.8 days)
//	a CONSOLE window, raised by an unentitled process:  SUCCEEDS, 3 of 3
//	a real CHROME window, same conditions:              REFUSED
//
// ⚠️ THE FIRST NUMBER KILLED A PLAN. It had been suggested that the grant
// lapses on its own, giving "no irritating late popups" for free. MAX_INT is
// Windows' effectively-never value: foreground stealing by an unentitled
// process is OFF, permanently, not on a timer. So a deadline has to be ours.
//
// ⚠️ AND THE OTHER TWO ARE DIFFERENT REGIMES, which was not expected. A
// console window is owned by conhost and is not protected the way an app
// window is. The window shimp would want raised is Chrome's, so the
// entitlement IS required for the case that matters.
//
// ⚠️ WHICH INVALIDATED A MEASUREMENT OF ITS OWN. An earlier run reported
// that AllowSetForegroundWindow(ASFW_ANY) survives the `go run` hop, because
// a grandchild raised the target. The control then showed the same raise
// succeeding with NO grant at all: the target was a console window, so the
// test measured the permissive case and proved nothing. Still unmeasured:
// whether the grant reaches a grandchild for a NORMAL window.
//
// It exists to answer, with numbers:
//
//  1. does AllowSetForegroundWindow(ASFW_ANY) survive the `go run` hop, given
//     the grant is PER-PROCESS and shimp's child is `go run` while the
//     capability is a grandchild whose PID shimp never sees?
//  2. does the grant LAPSE on a window that arrives late, which would mean
//     Windows already implements "no irritating late popups" for us?
//  3. what is the system's own foreground lock timeout here?
//
// ⚠️ WHAT THIS CANNOT MEASURE, stated up front. The entitlement shimp
// actually relies on comes from the user pressing a key in our console: that
// makes our process the one that received the last input event, which is a
// condition under which Windows permits foregrounding at all. No automated
// run can produce that, so the numbers below describe the UNENTITLED case.
// That is still the useful half: it establishes the baseline refusal, and
// anything that succeeds here would succeed a fortiori after a keypress.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"
	"unsafe"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procSetForeground    = user32.NewProc("SetForegroundWindow")
	procGetForeground    = user32.NewProc("GetForegroundWindow")
	procAllowSetFg       = user32.NewProc("AllowSetForegroundWindow")
	procSystemParameters = user32.NewProc("SystemParametersInfoW")
	procGetWindowText    = user32.NewProc("GetWindowTextW")
	procGetWindowPID     = user32.NewProc("GetWindowThreadProcessId")
	procIsWindowVisible  = user32.NewProc("IsWindowVisible")

	procAllocConsole     = kernel32.NewProc("AllocConsole")
	procGetConsoleWindow = kernel32.NewProc("GetConsoleWindow")
	procFreeConsole      = kernel32.NewProc("FreeConsole")
)

const (
	spiGetForegroundLockTimeout = 0x2000
	asfwAny                     = ^uintptr(0) // (DWORD)-1
)

func titleOf(hwnd uintptr) string {
	buf := make([]uint16, 256)
	procGetWindowText.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), 256)
	return syscall.UTF16ToString(buf)
}

func pidOf(hwnd uintptr) uint32 {
	var p uint32
	procGetWindowThreadProcessIdSafe(hwnd, &p)
	return p
}

func procGetWindowThreadProcessIdSafe(hwnd uintptr, p *uint32) {
	procGetWindowPID.Call(hwnd, uintptr(unsafe.Pointer(p)))
}

func describe(label string, hwnd uintptr) {
	if hwnd == 0 {
		fmt.Printf("  %-22s none\n", label)
		return
	}
	vis, _, _ := procIsWindowVisible.Call(hwnd)
	fmt.Printf("  %-22s hwnd=0x%X pid=%d visible=%v  %q\n",
		label, hwnd, pidOf(hwnd), vis != 0, titleOf(hwnd))
}

func main() {
	mode := flag.String("mode", "state", "timeout|state|target|raise|grant-and-run")
	hwnd := flag.String("hwnd", "0", "target window, hex or decimal")
	delay := flag.Int("delay", 0, "seconds to wait before raising")
	grant := flag.Bool("grant", false, "call AllowSetForegroundWindow first")
	flag.Parse()

	h, _ := strconv.ParseUint(*hwnd, 0, 64)

	switch *mode {
	case "timeout":
		measureTimeout()
	case "state":
		measureState()
	case "target":
		beTarget()
	case "raise":
		doRaise(uintptr(h), *delay, *grant)
	case "grant-and-run":
		grantAndRun(uintptr(h), *delay)
	default:
		fmt.Fprintf(os.Stderr, "unknown mode %q\n", *mode)
		os.Exit(2)
	}
}

// measureTimeout reads the system's own notion of how long after user input
// a process may take the foreground.
func measureTimeout() {
	var ms uint32
	r, _, err := procSystemParameters.Call(spiGetForegroundLockTimeout, 0,
		uintptr(unsafe.Pointer(&ms)), 0)
	if r == 0 {
		fmt.Printf("SPI_GETFOREGROUNDLOCKTIMEOUT: could not read it: %v\n", err)
		return
	}
	fmt.Printf("SPI_GETFOREGROUNDLOCKTIMEOUT = %d ms\n", ms)
	switch {
	case ms == 0:
		fmt.Println("  ⚠️ ZERO, so the system imposes NO lock period: a process")
		fmt.Println("     that is otherwise entitled may foreground at any time,")
		fmt.Println("     and a deadline would have to be ours rather than the OS's.")
	case ms >= 0x7FFFFFFF:
		// ⚠️ THE READING THAT KILLED A PLAN, and the first version of this
		// branch got it badly wrong by printing "the OS refuses foreground
		// changes for 2147483647 ms after input, which is the same idea as
		// the proposed deadline". At MAX_INT that sentence is nonsense: it
		// is not a lock PERIOD, it is a permanent refusal.
		fmt.Println("  ⚠️ MAX_INT, which is Windows' effectively-never value.")
		fmt.Println("     Foreground stealing by an unentitled process is OFF,")
		fmt.Println("     permanently. This is NOT a timer, and a deadline")
		fmt.Println("     cannot be derived from it: ours to build, or none.")
	default:
		fmt.Printf("  a real lock period of %d ms after input, which IS the\n", ms)
		fmt.Println("     same idea as a deadline and could be leaned on.")
	}
}

func measureState() {
	fg, _, _ := procGetForeground.Call()
	describe("foreground now", fg)
	cw, _, _ := procGetConsoleWindow.Call()
	describe("our console", cw)
	fmt.Printf("  %-22s %v\n", "we hold foreground", fg != 0 && fg == cw)
	fmt.Printf("  %-22s %d\n", "our pid", os.Getpid())
}

// beTarget allocates a console window, prints its handle, and waits to be
// raised. A window we own, so nothing of the user's is disturbed.
func beTarget() {
	procFreeConsole.Call()
	if r, _, err := procAllocConsole.Call(); r == 0 {
		fmt.Fprintf(os.Stderr, "AllocConsole failed: %v\n", err)
		os.Exit(1)
	}
	co, err := os.OpenFile("CONOUT$", os.O_WRONLY, 0)
	if err != nil {
		os.Exit(1)
	}
	os.Stdout = co
	cw, _, _ := procGetConsoleWindow.Call()
	// ⚠️ Printed to the ORIGINAL stderr as well, because the new console is
	// the thing being measured and the harness needs the handle on a channel
	// it can read.
	fmt.Fprintf(os.Stderr, "TARGET_HWND=0x%X\n", cw)
	fmt.Fprintf(co, "focusprobe target. hwnd=0x%X pid=%d\n", cw, os.Getpid())
	time.Sleep(120 * time.Second)
}

// doRaise waits, then tries to take the foreground, and reports what the OS
// actually did rather than what was asked.
func doRaise(target uintptr, delaySec int, grant bool) {
	fmt.Printf("raise: pid=%d target=0x%X delay=%ds grant=%v\n",
		os.Getpid(), target, delaySec, grant)
	if grant {
		r, _, err := procAllowSetFg.Call(asfwAny)
		fmt.Printf("  AllowSetForegroundWindow(ASFW_ANY) -> %d  (%v)\n", r, err)
	}
	if delaySec > 0 {
		fmt.Printf("  waiting %ds, so a LATE arrival is what gets measured\n", delaySec)
		time.Sleep(time.Duration(delaySec) * time.Second)
	}
	before, _, _ := procGetForeground.Call()
	r, _, _ := procSetForeground.Call(target)
	time.Sleep(300 * time.Millisecond)
	after, _, _ := procGetForeground.Call()

	// ⚠️ NO ERRNO HERE, DELIBERATELY, and the first version printed one.
	//
	// SetForegroundWindow returns 0 for "refused" and does NOT call
	// SetLastError, so whatever GetLastError holds is left over from some
	// earlier call. It printed "Invalid window handle" for a handle Windows
	// had just confirmed was valid, and two readings were written down as
	// "REFUSED, and it said so" when it had said nothing at all.
	//
	// The return value and the before/after foreground are the whole
	// evidence. A stale errno dressed as a reason is worse than no reason,
	// because it gets quoted.
	fmt.Printf("  SetForegroundWindow -> %d\n", r)
	fmt.Printf("  foreground before 0x%X, after 0x%X\n", before, after)
	switch {
	case after == target:
		fmt.Println("  RESULT: RAISED. The target now holds the foreground.")
	case r != 0:
		fmt.Println("  ⚠️ RESULT: the call SUCCEEDED and the window did NOT come")
		fmt.Println("     forward. Windows flashed the taskbar instead, which is")
		fmt.Println("     what it does when the caller is not entitled.")
	default:
		fmt.Println("  RESULT: REFUSED. The return was 0 and the foreground")
		fmt.Println("     did not move. Windows gives no reason for this and")
		fmt.Println("     none is invented here.")
	}
}

// grantAndRun is the hop under test: grant here, then go through `go run` to
// a grandchild that does the raising.
//
// ⚠️ This is the exact topology shimp has: shimp -> go run -> capability.
// The grant is PER-PROCESS, so the question is whether a grant made here
// reaches a process two hops away.
func grantAndRun(target uintptr, delaySec int) {
	fmt.Printf("granter: pid=%d\n", os.Getpid())
	fg, _, _ := procGetForeground.Call()
	describe("foreground at grant", fg)
	cw, _, _ := procGetConsoleWindow.Call()
	fmt.Printf("  we hold foreground: %v\n", fg != 0 && fg == cw)

	r, _, err := procAllowSetFg.Call(asfwAny)
	fmt.Printf("  AllowSetForegroundWindow(ASFW_ANY) -> %d  (%v)\n", r, err)
	if r == 0 {
		fmt.Println("  ⚠️ THE GRANT ITSELF FAILED, which is expected when the")
		fmt.Println("     granter does not hold the foreground. Everything below")
		fmt.Println("     therefore measures the UNENTITLED case.")
	}

	// Through `go run`, so the intermediary shimp cannot avoid is included.
	fmt.Println("  --- through `go run` to a grandchild ---")
	cmd := exec.Command("go", "run", ".", "-mode", "raise",
		"-hwnd", fmt.Sprintf("0x%X", target),
		"-delay", strconv.Itoa(delaySec))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("  grandchild failed: %v\n", err)
	}
}
