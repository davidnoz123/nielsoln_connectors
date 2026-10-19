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

	"[net/http]": "⚠️ ADDED 20 Oct 2026 WITH -port, and the absence this " +
		"replaces said it would be. ONE loopback request: GET " +
		"http://127.0.0.1:<port>/json/list, to find a tab already on " +
		"chatgpt.com. It forms no other address and reaches no remote host.",

	"[net, net.Dial: the websocket to 127.0.0.1]": "opens a CDP websocket to that tab and sends " +
		"Runtime.evaluate. ⚠️ ws:// ON LOOPBACK ONLY, checked rather than " +
		"assumed: a wss:// would silently skip TLS and the only symptom " +
		"would be a connection that worked. " +
		"⚠️ EVERY NETWORK CALL TO CHATGPT IS MADE BY THE BROWSER, NOT BY " +
		"THIS PROGRAM. It sends JavaScript; the page does the fetching with " +
		"the session it already has. So this program never speaks to " +
		"openai.com and never holds a cookie.",

	"[the access token]": "⚠️ READ AND SPENT INSIDE THE PAGE, NEVER RETURNED. " +
		"/backend-api/* needs Authorization: Bearer <accessToken>, and the " +
		"response carrying that token says in its own payload that sharing " +
		"it is like sharing your password. One expression fetches the " +
		"session, builds the header, makes the call and returns only the " +
		"conversation JSON. The token is never sent across the websocket, " +
		"never held in a Go variable and never written anywhere. " +
		"⚠️ THIS IS A FIX FOR A MISTAKE MADE THE SAME DAY. A probe answering " +
		"\"is this signed in\" mis-parsed the CDP reply, fell through to a " +
		"diagnostic that printed the raw response, and put a token prefix in " +
		"a transcript. The lesson was to reduce inside the page so there is " +
		"nothing to leak, not to be careful with output.",

	"[the conversation list]": "the ONE place this program interprets " +
		"anything, and it reads TWO FIELDS: `id` and `update_time`, to " +
		"decide which conversations to fetch. Bodies still pass through " +
		"uninterpreted. ⓘ It stops at the first page whose entries are all " +
		"older than -since, which is what makes it incremental.",

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
	"⚠️ no network BEYOND LOOPBACK, which is a narrower claim than the one " +
		"this replaces. Until -port landed this read \"no network of any " +
		"kind ... when it is built, this line must change and the " +
		"attestation must be re-issued\". It was built on 20 Oct 2026, so " +
		"the line changed and ⚠️ THE ATTESTATION IS NOW OWED. The only " +
		"addresses formed are 127.0.0.1: one HTTP GET and one websocket.",
	"no request to openai.com or chatgpt.com FROM THIS PROCESS. The browser " +
		"makes those, with the session it already had. ⚠️ Which is the " +
		"distinction that matters for an audit: this program cannot " +
		"authenticate to anything, because it holds no credential to do it " +
		"with.",
	"no cookie jar, no token store, no password store, and no reading of " +
		"any profile directory. It does not know where the profile is.",
	"⚠️ no navigation and no tab creation. If no chatgpt.com tab is open it " +
		"REFUSES with a reason rather than opening one: moving somebody's " +
		"tabs is not this program's business.",
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
	"archive/zip", "os.Create", "encoding/json", "net/http", "wsDial",
	"Runtime.evaluate",
	// The ones that must stay absent. Named here so the checker can assert
	// their absence rather than nobody noticing they arrived.
	"os/exec", "syscall", "database/sql", "os.Remove", "Page.navigate",
}
