// caveats.go -- what this program can reach, declared in source.
//
// s54: a caveat declared here is CHECKED by `repocheck`, which walks every
// module's imports and fails if an outward reach is not named in this file.
//
// ⚠️ THIS HEADER USED TO SAY "verified by an AST walk in the test", and there
// was no such test in this module -- only `chrome_sessions` had one. So the
// audit file asserted its own enforcement, which is exactly what s56 forbids:
// nothing in the audit payload may be an assertion. Corrected 19 Oct 2026 to
// name the checker that actually runs. The check is import-level, so it is a
// floor rather than a ceiling; `repocheck/caveats.go` states its limits.
//
// ⚠️ WRITTEN WITH THE MODULE, not afterwards. The `repo` slate's r6 records
// the state this is correcting: one caveats file and one attestation in the
// whole repo, BOTH on `chrome_sessions`, which nothing runs, and ABSENT from
// `chrome_ready`, the only module a link executes. The apparatus was on the
// wrong module. This one starts with it.
//
// ⚠️ AND THE SUBJECT MATTER IS WHY IT MATTERS MOST HERE. This program's input
// is a human's entire ChatGPT history: 1,904 conversations and 34,488
// messages in the measured export. An auditor reading a link that names it
// is entitled to know exactly what it touches, and the honest answer is
// short, which is the point.
package main

// Caveats is what an auditor should be told, keyed by the surface reached.
//
// The bracketed tag is the symbol a reader can grep for, and the test
// requires one declaration per REACH rather than per file: s54's checker was
// per-file once and a planted `exec.Command` passed.
var Caveats = map[string]string{
	"[archive/zip]": "opens a zip the CALLER names and reads only entries " +
		"matching conversations-*.json. It does not extract, does not write " +
		"into the archive, and never looks outside it. chat.html, which is " +
		"79MB of the measured export, is not read.",

	"[os.Create]": "writes ONE file, the path given by -out, or stdout when " +
		"that path is `-`. It appends nothing elsewhere, creates no " +
		"directories and removes nothing.",

	"[encoding/json]": "decodes each conversation as a RawMessage and " +
		"compacts it. ⚠️ It does not interpret the contents: no field is " +
		"read, no value is transformed, and the bytes written are the bytes " +
		"read with insignificant whitespace removed.",
}

// CaveatsNotReached is the other half of the claim, and the more useful one.
//
// ⚠️ An auditor's question is rarely "what does it do" but "could it do X".
// Saying what is ABSENT answers that in a way a list of present reaches
// cannot, and the test pins each of these as genuinely unreachable.
var CaveatsNotReached = []string{
	"no network of any kind. No http, no net, no websocket, no CDP. ⚠️ The " +
		"CDP capture half of this capability is NOT BUILT; when it is, this " +
		"line must change and the attestation must be re-issued.",
	"no exec. It starts no process and shells out to nothing.",
	"no registry, no COM, no syscall. It is portable Go with no build tags.",
	"no sqlite. It stops at a file, which is forced rather than chosen: " +
		"maintaining the FTS5 index from Go would mean reimplementing a " +
		"sqlite extension's segment format.",
	"no deletion, and no writing outside the single -out path.",
	"no credentials. It reads a file the caller already has and authenticates " +
		"to nothing.",
}

// CaveatSurfaces are the imports and symbols the checker looks for.
//
// ⚠️ TYPES AS WELL AS FUNCTIONS. s54's walk missed `client.Get`, a method on
// a value, which was its package's primary reach. Listing the import paths
// catches what a function-name list does not.
var CaveatSurfaces = []string{
	"archive/zip", "os.Create", "encoding/json",
	// The ones that must stay absent. Named here so the checker can assert
	// their absence rather than nobody noticing they arrived.
	"net/http", "net", "os/exec", "syscall", "database/sql",
}
