// known_recipes.go -- SEMANTIC duplication, declared, because nothing finds it.
//
// ⚠️ WHY A HAND-WRITTEN LIST IS THE RIGHT ANSWER HERE, which is not a
// sentence this repo writes lightly.
//
// There are two kinds of duplication and only one is mechanically findable:
//
//	copy-paste-and-rename   SharedBodies() finds it, by hashing bodies
//	the same RECIPE twice   nothing finds it, because they are not copies
//
// The second kind is two places performing the same sequence of steps in
// genuinely different code. No body hash matches, no function name matches,
// and no cheap structural check distinguishes it from two unrelated
// functions that happen to do similar work. Pretending otherwise would give
// a check that is green because it is blind, which is the failure shape this
// repo has hit three times already.
//
// So it is written down. And ⚠️ a written-down list rots, which is why each
// entry carries MARKERS: substrings that must still be present in each named
// file. TestEveryRecipeIsStillReal fails when an entry goes stale, so the
// list cannot quietly come to describe code that no longer exists. That is
// the same discipline TestBaselineHasNoStaleEntries applies to
// known_overlaps.go, which caught seven fictional entries on its first run.
package main

// Recipe is one sequence of steps performed in more than one place.
type Recipe struct {
	// Name is what the recipe is, in a few words.
	Name string

	// Sites are repo-relative files that each implement it.
	Sites []string

	// Markers are substrings that must appear in EVERY site. They are what
	// makes the entry falsifiable: if a site stops containing them, either
	// the duplication is gone or it moved, and both need a human.
	Markers []string

	// Why it is tolerated rather than shared. ⚠️ Required. An entry without
	// a reason is an entry nobody dares delete.
	Why string
}

// recipes is the declared inventory.
var recipes = []Recipe{
	{
		Name: "the live install cycle: build the real CLI, install an " +
			"isolated scheme, trigger, poll the handler's log, purge",
		Sites: []string{
			"shimp/live_test.go",
			// ⓘ corpus_run.go, NOT main.go. This said main.go and was wrong
			// before the `install` -> `reinstall` rename, which is what
			// surfaced it: the marker check looked for a verb in a file that
			// had never contained one. A recipe naming the wrong site is a
			// declaration nobody can check, which is the failure this whole
			// baseline exists to prevent.
			"linkprobe/corpus_run.go",
		},
		Markers: []string{
			"go", "build", "reinstall", "-purge", "-scheme",
		},
		// ⚠️ AND SHARING IT WOULD BE WORSE, which is the whole reason this
		// entry exists rather than a shared package.
		//
		// A Go test-only import still lands in go.mod and go.sum. s55's
		// maxim is that go.sum ENUMERATES THE CLOSURE and is the audit's
		// stopping condition. So if shimp imported a shared harness, every
		// auditor pasting a shimp:// link would find test tooling in the
		// audit surface of the thing being shipped -- a cost paid by every
		// user of the protocol, to save one copy of five steps.
		//
		// The process boundary is therefore the correct seam, not a
		// compromise: linkprobe's -json mode is how a test suite consumes
		// it, with no import and no go.mod entry anywhere.
		//
		// The two are also not substitutable. live_test.go tests INSTALL
		// CORRECTNESS: that the registry keys are gone after purge, that a
		// running binary cannot delete itself and says so, that the stamp
		// records the scheme. linkprobe tests DELIVERY across fifteen
		// routes. Same mechanism, different questions, so both earn their
		// place. It is the mechanism that is duplicated, not the purpose.
		Why: "sharing it would put test tooling in shimp's go.sum, which is " +
			"the audit's stopping condition (s55). The process boundary is " +
			"the correct seam, and the two test different questions.",
	},
	{
		Name: "a CDP client in Python: connect, send, match the reply by id",
		Sites: []string{
			"linkprobe/chrome_route.py",
			"probes/cdp.py",
		},
		Markers: []string{
			"suppress_origin", "websocket", "json.loads",
		},
		// ⚠️ NOT COLLAPSED, and the reason is mechanical rather than
		// aesthetic. chrome_route.py binds `websocket` inside
		// safe_local_imports rather than at module level, so importing its
		// CDP class raises NameError on first use. It cannot be reused
		// without changing it, and it is committed, tested and in
		// linkprobe's loop.
		//
		// probes/ exists to stop losing measurement instruments, not to
		// refactor working code, so the duplication is declared instead.
		// ⚠️ Both this entry and acknowledgedBodies["close"] should go the
		// day chrome_route.py's client becomes importable.
		Why: "chrome_route.py's client is not importable: it binds websocket " +
			"inside safe_local_imports, so using its class raises NameError. " +
			"Collapsing it means changing committed tested code, which is " +
			"not what probes/ is for.",
	},
	{
		Name: "a hand-rolled websocket client for CDP",
		Sites: []string{
			"remote_chrome_cdp/ws.go",
			// ⚠️ main.go, NOT ws.go. The first version of this entry said
			// ws.go on the strength of the other site's filename, and
			// TestEveryRecipeIsStillReal refused to read it. remote_ai's
			// client is inline in main.go.
			"remote_ai/main.go",
		},
		Markers: []string{
			"Sec-WebSocket-Key", "opcode",
		},
		// ⚠️ s60 recorded this pair as duplication that the overlap count
		// never saw, because the two use different function names. It is
		// now declared rather than remembered.
		//
		// And it is DELIBERATE, which s60 did not know at the time:
		// remote_chrome_cdp/ws.go's own header says why. remote_ai's client
		// carries TLS, permessage-deflate and a handshake hardened against
		// captive portals, because it crosses the internet through Caddy.
		// The other talks to 127.0.0.1 and nothing else, so every one of
		// those costs buys nothing there. The header even names the right
		// future move: lift remote_ai's dial() when the bridge transport
		// lands, rather than merging the two now.
		//
		// So this is not undiscovered duplication awaiting extraction. It
		// is a recorded decision that the two should differ.
		Why: "deliberate. remote_ai's crosses the internet and needs TLS and " +
			"permessage-deflate; remote_chrome_cdp's talks to 127.0.0.1 " +
			"only. Merging them would overturn a decision, not tidy an " +
			"oversight. See remote_chrome_cdp/ws.go's header.",
	},
}
