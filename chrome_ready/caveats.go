// caveats.go -- what this program can reach, declared in source.
//
// ⚠️ WRITTEN LAST, WHICH IS THE WRONG ORDER AND THE REASON THIS FILE EXISTS.
// `chrome_ready` is the ONLY module a shimp link has ever actually executed
// (s68, proven by a real click) and it was the one module with no caveats at
// all, while the single caveats file in the repo sat on `chrome_sessions`,
// which nothing runs. The apparatus was on the wrong module for five days.
// Recorded here rather than quietly fixed, because the lesson is the ordering:
// the audit surface belongs on whatever a link names, first.
//
// # WHAT AN AUDITOR SHOULD UNDERSTAND THIS MODULE TO BE
//
// It ensures an authenticated browser on a declared port, FOR ANY SITE. It is
// not a ChatGPT accessory, and this file read like one until 20 Oct 2026: the
// profile measured on the machine it was written for carried ClickUp,
// DataAnnotation, Slack and iCloud sessions simultaneously.
//
// ⚠️ THAT WIDENS WHAT IS AT STAKE RATHER THAN NARROWING IT. The directory this
// program creates can come to hold every session its user signs in to, so the
// absences below -- no cookie reading, no credential store, no profile copying
// -- are load-bearing for all of them at once, not for one account.
package main

// Caveats is what an auditor should be told, keyed by the surface reached.
var Caveats = map[string]string{
	"[exec.Command]": "LAUNCHES CHROME, visibly, and never detached. The " +
		"executable is discovered from PROGRAMFILES, PROGRAMFILES(X86) and " +
		"LOCALAPPDATA; it is not taken from PATH and not configurable, so a " +
		"link cannot name an arbitrary binary to run. It passes exactly SIX " +
		"flags and nothing else: --remote-debugging-port, --user-data-dir, " +
		"--no-first-run, --no-default-browser-check, --disable-sync and " +
		"--allow-browser-signin=false. The process is left RUNNING on " +
		"purpose: that is the capability.",

	"[--disable-sync, --allow-browser-signin=false]": "⚠️ s73's R4, and the " +
		"rule that prevents the worst surprise. MEASURED 19 Oct 2026: a " +
		"profile this program created, ONE HOUR OLD, held 10 extensions and " +
		"106MB of them (Grammarly at 47.8MB), active Sync Data, 272KB of " +
		"history and 184KB of autofill. Nothing here read a credential: " +
		"signing in to A SITE with \"Continue with Google\" signed the " +
		"BROWSER in as well, and sync pulled the account into a directory a " +
		"LINK had created. ⓘ The site measured happened to be ChatGPT; the " +
		"mechanism is Google SSO and applies to any of them.\n" +
		"⚠️ SO THE HONEST WARNING TO A USER IS: an SSO sign-in can bring " +
		"account data with it. These two flags are the defence -- signin.allowed " +
		"was measured as False with them and True without -- and a later " +
		"throwaway profile came to 88MB against the earlier 356MB, which is " +
		"the same finding from the other end.\n" +
		"⚠️ VERIFIED AS A SETTING, NOT AS AN OUTCOME. No Google sign-in was " +
		"attempted against the flags, so the claim is that Chrome records " +
		"sign-in as disallowed, not that sync has been proven unreachable.",

	"[lockfile, read]": "reads one filename inside the profile directory, " +
		"`lockfile`, to answer whether a Chrome already holds it. s73's R6. " +
		"⚠️ Chrome removes it on a clean close, measured 20 Oct 2026, so its " +
		"presence is the signal. It is STATTED, never opened or written. " +
		"ⓘ A crashed Chrome can leave it behind, which makes a busy report a " +
		"probably rather than a certainly; the refusal says so and names the " +
		"file, so a stale lock is a warning nobody needed rather than a " +
		"clash nobody saw coming.",

	"[-list: filepath.Walk]": "under `-list` ONLY, totals the file sizes " +
		"under each profile in the default base. s73's R3, because 3.9 GB of " +
		"user-data-dirs accumulated on one machine without anybody deciding " +
		"to spend it. It reads sizes and never contents, skips what it " +
		"cannot read, and ⚠️ never runs on the path a click takes: walking a " +
		"3 GB tree would add seconds to the thing a link is waiting for.\n" +
		"⚠️ AND THERE IS DELIBERATELY NO -remove. This binary is invocable " +
		"BY A LINK, and a link able to delete a directory tree is a " +
		"capability nobody asked for. Reporting the path is enough.",

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
	"⚠️ no --remote-allow-origins. s73's R5, REMOVED 20 Oct 2026 after " +
		"measuring that it was never needed: launched without it, a CDP " +
		"client that SUPPRESSES its Origin header connects and answers " +
		"normally, while a client sending a page-like Origin is refused with " +
		"403 Forbidden. probes/cdp.py already passes suppress_origin=True. " +
		"⚠️ So the exposure an auditor was previously told to stop on is " +
		"GONE: a web page can no longer drive this browser, because Chrome " +
		"refuses any connection carrying an origin.",
	"no deletion of anything, no -remove verb, and no writing outside the " +
		"one profile directory. No registry and no COM.",
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
	"os.Stat", "filepath.Walk", "os.ReadDir",
	// must stay absent
	"database/sql", "syscall", "os.Remove", "os.RemoveAll", "websocket",
	"remote-allow-origins",
}
