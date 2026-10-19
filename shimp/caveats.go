// caveats.go -- what the HANDLER can reach, declared in source.
//
// ⚠️ THIS IS THE ONE THAT MATTERS MOST, AND IT WAS WRITTEN LAST. Every link
// trusts the handler: a caveat on a connector describes what one capability
// does, and a caveat here describes what ANY clicked link can cause. An
// auditor who reads only one file in this repo should read this one.
//
// Checked by `repocheck`, which fails if an outward reach is not named here.
// ⚠️ That check is import-level, so it is a floor: it confirms `os/exec` is
// declared, not that this file describes the exec accurately. The accuracy is
// human work, and the paragraph below is the part to attack first.
package main

// Caveats is what an auditor should be told, keyed by the surface reached.
var Caveats = map[string]string{
	// ------------------------------------------------------------------
	"[exec.Command]": "⚠️ THE WHOLE RISK OF THE PROTOCOL, IN ONE LINE: " +
		"`exec.Command(argv[0], argv[1:]...)` where argv is DERIVED FROM A " +
		"CLICKED URI. Three things bound it, and all three are mechanical " +
		"rather than advisory.\n" +
		"(1) argv[0] is NOT taken from the link. It comes from s43's CLOSED " +
		"runner table in uri.go, whose only entry is {\"go\", \"run\"}, and " +
		"an unrecognised runner FAILS CLOSED. So the only program a link can " +
		"start today is the Go toolchain.\n" +
		"(2) what the toolchain then runs is `module@<40 hex>`, verified " +
		"against go.sum and, for a public module, witnessed by " +
		"sum.golang.org. The handler does not fetch, build or verify " +
		"anything itself.\n" +
		"(3) a human says yes first. See [consent] below, and read its " +
		"history before trusting it.",

	// ------------------------------------------------------------------
	"[advapi32: RegCreateKeyExW, RegSetValueExW, RegDeleteKeyW]": "writes " +
		"TWO keys, both under HKEY_CURRENT_USER, which is s30: per-user, so " +
		"no elevation is ever requested and nothing machine-wide changes.\n" +
		"  Software\\Classes\\<scheme>      the protocol registration\n" +
		"  Trusted Protocols\\...\\<scheme>: the Office key, below\n" +
		"It also READS keys (RegOpenKeyExW, RegQueryValueExW) to report what " +
		"is installed, and DELETES only the two it created.",

	"[Office Trusted Protocols]": "⚠️ THIS ONE DISABLES A SECURITY PROMPT, " +
		"so it is declared separately rather than folded into the registry " +
		"line above. The key `Software\\Policies\\Microsoft\\Office\\16.0\\" +
		"Common\\Security\\Trusted Protocols\\All Applications\\<scheme>:` " +
		"tells every Office application NOT to warn before handing this " +
		"protocol to the OS. That is the intended effect -- s8 records an " +
		"Excel instance held behind exactly that modal for two and a half " +
		"hours, with no symptom but a read-only refusal naming no pid -- but " +
		"an auditor is entitled to know that installing this handler " +
		"suppresses a warning Microsoft chose to show.",

	// ------------------------------------------------------------------
	"[shell32: ShellExecuteW]": "hands a URI to the OS, through the " +
		"registry, by the same path a click takes. ⚠️ Reachable in normal " +
		"use: it is the `dispatch` verb, not test-only code. It exists for " +
		"MEASUREMENT (s36 -- how much of a URI survives into os.Args[1] " +
		"cannot be measured without a registered handler), and it is the one " +
		"verb that deliberately causes the OS to invoke a handler.",

	// ------------------------------------------------------------------
	"[kernel32: AllocConsole, FreeConsole, ReadConsoleInputW]": "allocates " +
		"ITS OWN console to ask for consent, and reads keystrokes from it.\n" +
		"⚠️ READ consent_windows.go's header before trusting this. " +
		"AllocConsole RETURNS EARLY IF A CONSOLE ALREADY EXISTS, so a " +
		"handler launched from a shell inherited that shell's console AND " +
		"ITS INPUT BUFFER -- which still held the Enter key that started the " +
		"command. Consent was therefore granted by a keypress the user had " +
		"made for something else, and Chrome launched in the same second. " +
		"Now FreeConsole precedes AllocConsole and CONIN$ is opened " +
		"explicitly. ⚠️ It was found only because somebody asked whether it " +
		"worked, not by any test.",

	"[user32: SetForegroundWindow, ShowWindow, GetForegroundWindow]": "moves " +
		"windows: brings the thing a click produced to the front, and puts " +
		"focus back where it was when a run is unattended. It reads the " +
		"foreground window handle and shows or minimises its own. It does " +
		"not enumerate, close, or send input to any other process's window.",

	// ------------------------------------------------------------------
	"[os.OpenFile, os.MkdirAll, os.Remove, os.WriteFile]": "writes inside " +
		"ONE directory, `~/.<scheme>`, and nowhere else: the installed " +
		"binary, shimp.log, and a stamp recording what was installed. " +
		"Uninstall removes those and that directory. ⚠️ CONOUT$ and CONIN$ " +
		"are also opened by name -- they are the console device, not files on " +
		"disk, and the only reason os.OpenFile appears in the consent path.",
}

// CaveatsNotReached is the other half of the claim, and the more useful one.
//
// ⚠️ An auditor's question is rarely "what does it do" but "could it do X".
var CaveatsNotReached = []string{
	"⚠️ NO NETWORK OF ANY KIND. The handler opens no socket, resolves no " +
		"name and speaks no protocol. Everything that reaches the network is " +
		"reached by the Go toolchain it starts, under go.sum, which is s55's " +
		"stopping condition for an audit.",
	"no sqlite, no database, and no credential store. It reads no cookie " +
		"jar, no token and no keychain.",
	"⚠️ no normalisation of a link. s35: a URI that is not in the one legal " +
		"spelling is REFUSED rather than tidied up, so there is no code path " +
		"where the handler runs something slightly different from what the " +
		"bytes said. Identity is the sha256 of those bytes verbatim.",
	"no elevation, ever. Both registry keys are per-user, and a failure to " +
		"write them is reported rather than retried as administrator.",
	"no writing outside `~/.<scheme>`, no deletion outside it, and no " +
		"touching of another application's files or settings -- with the one " +
		"declared exception of the Office policy key above.",
	"⚠️ no action delivery and no resident process. An action-bearing link " +
		"is refused BY NAME because the keeper (s20) is not built. So today " +
		"the handler can only start a one-shot, and it says so rather than " +
		"doing something approximate.",
	"no second instance. s46's guarantee is one running instance per " +
		"interactive session; nothing here starts a background service, and " +
		"the uninstall path leaves nothing behind that could.",
}

// CaveatSurfaces are the imports and symbols the checker looks for.
//
// ⚠️ TYPES AND LAZY-LOADED PROCS AS WELL AS FUNCTIONS. A raw syscall through
// NewLazyDLL/NewProc looks nothing like a call an AST walk recognises, so the
// DLL names are listed here to be greppable even though no checker can prove
// the list complete. That limit is stated in chrome_sessions/caveats.go too,
// and it is the reason this file is read by humans rather than trusted.
var CaveatSurfaces = []string{
	"exec.Command", "os.OpenFile", "os.MkdirAll", "os.Remove", "os.WriteFile",
	"syscall", "advapi32.dll", "kernel32.dll", "shell32.dll", "user32.dll",
	// The ones that must stay absent.
	"net/http", "net.Dial", "database/sql", "ShellExecuteExW", "CreateServiceW",
}
