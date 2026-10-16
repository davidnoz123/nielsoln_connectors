//go:build windows

// focus_windows.go -- put a click's console behind, when asked.
//
// ⚠️ RESURRECTED, AND HONEST THIS TIME. An earlier version of this file was
// deleted because it half worked and was the silent default: 5 of 8
// dispatches kept the focus where it was and 3 logged a refusal, while the
// code implied it always worked.
//
// Nothing about the mechanism was wrong. What was wrong was being the
// default and over-claiming. It is now one of three values a caller CHOOSES,
// and it says what it managed.
//
// # WHY IT IS BEST EFFORT AND CANNOT BE BETTER
//
// AllocConsole creates the window AND activates it. There is no flag, no
// ordering and no prior call that prevents that: the window is foreground
// before this process can issue its first instruction about it. So the only
// move available is to put it back afterwards, which is a race this
// sometimes loses.
//
// Three attempts, weakest first, because each is more disruptive:
//
//	SW_SHOWNOACTIVATE  ask for it to be shown without activating
//	SetForegroundWindow give the focus back to whoever had it
//	SW_MINIMIZE        ⚠️ deterministic: a minimised window cannot be
//	                   foreground. Last because it hides output a human may
//	                   want, which is the whole reason `background` exists
//	                   rather than `none`.
package main

var (
	procShowWindow          = user32.NewProc("ShowWindow")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
)

const (
	swShowNoActivate = 4
	swMinimize       = 6
)

// focusKeeper remembers who held the foreground so it can be given back.
type focusKeeper struct {
	prev   uintptr
	active bool
}

// rememberFocus captures the foreground window BEFORE a console is created.
//
// ⚠️ Must be called before allocConsole or ownConsole. Afterwards the console
// is already foreground and the previous holder is unknowable.
func rememberFocus() focusKeeper {
	if windowMode() != WindowBackground {
		return focusKeeper{}
	}
	fg, _, _ := procGetForegroundWindow.Call()
	return focusKeeper{prev: fg, active: true}
}

// restore puts the console behind. It reports what it managed.
func (f focusKeeper) restore() {
	if !f.active {
		return
	}
	hwnd, _, _ := procGetConsoleWindow.Call()
	if hwnd == 0 {
		return
	}
	procShowWindow.Call(hwnd, swShowNoActivate)

	// ⚠️ ONLY GIVE BACK WHAT WAS TAKEN. Measured: with SW_SHOWNOACTIVATE the
	// console often never takes the foreground, and an earlier version
	// logged "Windows refused to return the focus" when nothing had moved.
	cur, _, _ := procGetForegroundWindow.Call()
	if cur != hwnd {
		_log("click: the console did not take the foreground (%s=background)",
			windowEnv)
		return
	}

	if f.prev != 0 && f.prev != hwnd {
		procSetForegroundWindow.Call(f.prev)
		if after, _, _ := procGetForegroundWindow.Call(); after == f.prev {
			_log("click: focus handed back (%s=background)", windowEnv)
			return
		}
	}

	// ⚠️ THE FALLBACK THAT ALWAYS WORKS. SetForegroundWindow is refused when
	// this process is not entitled, and holding the foreground is not always
	// enough. A minimised window cannot be foreground, so this ends the
	// question -- at the cost of hiding the output.
	procShowWindow.Call(hwnd, swMinimize)
	_log("click: focus could not be handed back, so the console was "+
		"MINIMISED instead. Its output is in the log. (%s=background)",
		windowEnv)
}
