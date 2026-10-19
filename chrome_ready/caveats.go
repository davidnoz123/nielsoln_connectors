// caveats.go -- what this program can reach, declared in source.
//
// ⚠️ WRITTEN LAST, WHICH IS THE WRONG ORDER AND THE REASON THIS FILE EXISTS.
// `chrome_ready` is the ONLY module a shimp link has ever actually executed
// (s68, proven by a real click) and it was the one module with no caveats at
// all, while the single caveats file in the repo sat on `chrome_sessions`,
// which nothing runs. The apparatus was on the wrong module for five days.
// Recorded here rather than quietly fixed, because the lesson is the ordering:
// the audit surface belongs on whatever a link names, first.
package main

// Caveats is what an auditor should be told, keyed by the surface reached.
var Caveats = map[string]string{
	"[exec.Command]": "LAUNCHES CHROME, visibly, and never detached. The " +
		"executable is discovered from PROGRAMFILES, PROGRAMFILES(X86) and " +
		"LOCALAPPDATA; it is not taken from PATH and not configurable, so a " +
		"link cannot name an arbitrary binary to run. ⚠️ It passes " +
		"--remote-debugging-port, --user-data-dir, --remote-allow-origins=*, " +
		"--no-first-run and --no-default-browser-check, and nothing else. " +
		"The process is left RUNNING on purpose: that is the capability.",

	"[--remote-allow-origins=*]": "⚠️ THE ONE AN AUDITOR SHOULD STOP ON. It " +
		"tells Chrome to accept CDP websocket connections from ANY origin, " +
		"so any page able to reach the port can drive the browser. The port " +
		"is bound to loopback, so this is a same-machine exposure rather than " +
		"a network one, but on a shared or multi-user machine it is real. It " +
		"is here because CDP clients are otherwise refused by origin, and it " +
		"is declared rather than defended.",

	"[http.Client]": "probes ONE url: GET http://127.0.0.1:<port>/json/version, " +
		"on loopback only, with a timeout. ⚠️ It is a liveness probe and not " +
		"evidence of authority: a browser answering on a port is not a " +
		"browser signed in to anything.",

	"[os.MkdirAll]": "creates ONE directory, {os.TempDir()}/{-profile}, mode " +
		"0755, to hold the browser profile. It writes nothing else, deletes " +
		"nothing and reads no file. ⚠️ A DEDICATED profile, not the user's " +
		"main Chrome profile, so the first run is signed out of everything. " +
		"The directory persists under its name between runs, which is what " +
		"makes a sign-in stick -- and it lives in TEMP, so a disk cleanup " +
		"silently signs the user out again.",
}

// CaveatsNotReached is the other half of the claim, and the more useful one.
//
// ⚠️ An auditor's question is rarely "what does it do" but "could it do X".
// Saying what is ABSENT answers that in a way a list of present reaches
// cannot.
var CaveatsNotReached = []string{
	"no network beyond loopback. The only address it forms is 127.0.0.1, and " +
		"the only path is /json/version. It sends no request anywhere else " +
		"and transmits nothing about the user.",
	"no CDP traffic of its own. It asks whether the port answers and stops " +
		"there: it opens no websocket, evaluates no script and reads no tab. " +
		"⚠️ Driving the browser is a DIFFERENT module's reach, and whichever " +
		"one does it must declare it.",
	"no credentials. It reads no cookie jar, no token and no password store, " +
		"and the profile it creates starts empty.",
	"no deletion, no writing outside the one profile directory, and no " +
		"registry or COM.",
	"no sqlite, and no reading of the user's main Chrome profile.",
	"no choice of binary. ⚠️ This is the one that stops a link being an " +
		"arbitrary code path: the executable comes from a fixed list of " +
		"install locations, so no flag and no argument can redirect it.",
}

// CaveatSurfaces are the imports and symbols the checker looks for.
//
// The ones that must stay ABSENT are named so their absence is asserted
// rather than nobody noticing they arrived.
var CaveatSurfaces = []string{
	"exec.Command", "http.Client", "os.MkdirAll",
	// must stay absent
	"database/sql", "syscall", "os.Remove", "os.WriteFile", "websocket",
}
