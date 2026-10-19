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
		"link cannot name an arbitrary binary to run. It passes exactly seven " +
		"flags and nothing else: --remote-debugging-port, --user-data-dir, " +
		"--remote-allow-origins=*, --no-first-run, " +
		"--no-default-browser-check, --disable-sync and " +
		"--allow-browser-signin=false. The process is left RUNNING on " +
		"purpose: that is the capability.",

	"[--disable-sync, --allow-browser-signin=false]": "⚠️ s73's R4, and the " +
		"rule that prevents the worst surprise. MEASURED 19 Oct 2026: a " +
		"profile this program created, ONE HOUR OLD, held 10 extensions and " +
		"106MB of them (Grammarly at 47.8MB), active Sync Data, 272KB of " +
		"history and 184KB of autofill. Nothing here read a credential: " +
		"signing in to ChatGPT with \"Continue with Google\" signed the " +
		"BROWSER in as well, and sync pulled the account into a directory a " +
		"LINK had created.\n" +
		"⚠️ SO THE HONEST WARNING TO A USER IS: an SSO sign-in can bring " +
		"account data with it. These two flags are the defence -- signin.allowed " +
		"was measured as False with them and True without -- and a later " +
		"throwaway profile came to 88MB against the earlier 356MB, which is " +
		"the same finding from the other end.\n" +
		"⚠️ VERIFIED AS A SETTING, NOT AS AN OUTCOME. No Google sign-in was " +
		"attempted against the flags, so the claim is that Chrome records " +
		"sign-in as disallowed, not that sync has been proven unreachable.",

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
		"browser signed in to anything.\n" +
		"⚠️ AND IT IS NOT EVIDENCE OF IDENTITY EITHER, which is s73's R2. " +
		"Until 20 Oct 2026 an answering port was accepted whatever was behind " +
		"it, so `-port 9333 -profile chatgpt` reported ALREADY READY against " +
		"a different browser entirely -- and on the machine where that was " +
		"found, port 9222 held live work sessions. The reply's browser GUID " +
		"is now matched against a marker this program wrote into the profile " +
		"it launched, and a mismatch or a missing marker REFUSES, naming both " +
		"sides. -accept-unverified overrides it and says so on every run.",

	"[os.WriteFile: the profile marker]": "writes one small JSON file, " +
		"`.chrome_ready.json`, INSIDE the profile it launched, holding the " +
		"port, the browser's own GUID and the pid. That is the whole of R2's " +
		"mechanism. It contains no credential and nothing about the user, and " +
		"it lives inside the profile so deleting the profile deletes the " +
		"claim. ⓘ The alternative was port-to-pid-to-command-line, rejected " +
		"because CDP refuses it without --enable-automation (which makes the " +
		"browser announce itself as automated) and because the TCP-table " +
		"route would mean copying syscall struct layouts AGENTS names as the " +
		"dangerous part.",

	"[os.MkdirAll]": "creates ONE directory to hold the browser profile, " +
		"mode 0755: the path the caller declared via -profile, or " +
		"{user config dir}/chrome_ready/{name} for a bare name. It writes " +
		"nothing else, deletes nothing and reads no file.\n" +
		"⚠️ A DEDICATED profile, never the user's main Chrome profile, for " +
		"the two reasons in profile.go: Chrome allows one process per " +
		"user-data-dir, and recent Chrome restricts remote debugging on the " +
		"default profile precisely because a process that can enable " +
		"debugging on a signed-in profile can read its cookies.\n" +
		"⚠️ SO THE FIRST RUN IS SIGNED OUT OF EVERYTHING, and the program " +
		"says so on the run that creates the profile. A sign-in then " +
		"persists, because the directory is DURABLE -- it was under " +
		"os.TempDir() until 19 Oct 2026, where a disk cleanup would remove " +
		"it and the only symptom was a capture finding an empty account. A " +
		"temp path is now REFUSED rather than warned about.\n" +
		"⚠️ AND IT DOES NOT COPY AN EXISTING PROFILE. Copying a signed-in " +
		"one would copy its cookie jar and token store: credential copying, " +
		"by a program a link can invoke. Refused on purpose.",
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
	"⚠️ no credentials ARE READ, and that is a narrower claim than it looks. " +
		"This program reads no cookie jar, no token and no password store, " +
		"and the profile it creates starts empty. ⚠️ BUT A USER SIGNING IN " +
		"CAN FILL IT: with \"Continue with Google\" the browser signs in " +
		"too and sync brings extensions, history and autofill, measured on " +
		"19 Oct 2026. R4's flags are the defence. The old wording here said " +
		"only that the profile starts empty, which was true of the code and " +
		"misleading about the outcome -- s56's failure shape in a new place.",
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
	// ⚠️ os.WriteFile MOVED FROM THE ABSENT LIST TO THIS ONE on 20 Oct 2026,
	// when R2's marker started using it. Caught because the file then
	// declared the same surface both present (in Caveats) and forbidden
	// (here), and an audit document that contradicts itself is worse than one
	// that says nothing. Any surface added to the code must be moved here in
	// the same change.
	"exec.Command", "http.Client", "os.MkdirAll", "os.WriteFile",
	// must stay absent
	"database/sql", "syscall", "os.Remove", "websocket",
}
