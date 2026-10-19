// caveat_scope.go -- which modules OWE a caveats.go, declared rather than guessed.
//
// ⚠️ WHY A DECLARATION AND NOT A RULE. The caveats contract exists because a
// clicked link names a module and an auditor must be able to learn what that
// module reaches. `linkprobe` and `probes/focus` cannot be named by a link:
// they are the test harness and the measurement instruments, AGENTS categories
// 4 and 5, and they run only when a developer runs them. Demanding caveats
// from them would be ceremony, and ceremony is how a check stops being read.
//
// But "cannot be named by a link" is not inferable from a directory name, and
// a checker that infers it starts lying the first time somebody adds a module.
// So it is written down, and every exemption carries its reason -- the same
// discipline as known_overlaps.go, for the same stated reason: an allowlist
// without reasons is an allowlist nobody dares to prune.
//
// ⚠️ AN EXEMPTION HERE IS NOT A STATEMENT THAT THE MODULE IS SAFE. It says
// nobody can reach it through the protocol. `probes/focus` calls
// SetForegroundWindow and spawns processes; it is exempt because a link cannot
// invoke it, not because that is harmless.
package main

import "strings"

// caveatExempt maps a module to why it owes no caveats.go.
//
// ⚠️ EVERY ENTRY IS A LIABILITY. If a module here ever becomes reachable from
// a link, this entry is the thing that hid it, so each reason says what would
// have to change for the exemption to lapse.
var caveatExempt = map[string]string{
	"linkprobe": "the test harness, AGENTS category 4. It DRIVES links rather " +
		"than being named by one: no shimp:// URI can resolve to it, because " +
		"it is not a capability. ⚠️ Lapses if it ever grows a verb a link " +
		"could reach, which would make it a capability wearing a harness's name.",

	"probes/focus": "a measurement instrument, AGENTS category 5. Exists to " +
		"produce a number for a design decision and is run by hand. ⚠️ NOT " +
		"exempt because it is harmless -- it foregrounds windows and starts " +
		"processes -- but because no link can invoke it.",

	"pdf_view_server": "PARKED, 18 Oct 2026, recorded in the repo slate's r3 " +
		"and in a banner in its own header. ⚠️ THE WEAKEST EXEMPTION HERE: " +
		"nothing mechanically prevents a link naming a parked module, so this " +
		"rests on the park holding. It lapses the moment the module is " +
		"unparked, and whoever unparks it owes the caveats first.",

	"remote_chrome_cdp": "PARKED, 18 Oct 2026, same basis and same weakness " +
		"as pdf_view_server. ⚠️ It reaches further than most -- net, " +
		"net/http, os/exec, syscall -- so if it is ever unparked this is the " +
		"first debt to settle.",
}

// CaveatExempt reports whether a module is declared as owing no caveats.go,
// and why.
func CaveatExempt(module string) (string, bool) {
	why, ok := caveatExempt[module]
	return why, ok
}

// firstSentence trims a reason to its first sentence for one-line reporting.
//
// The full text stays in caveat_scope.go, which is where somebody deciding
// whether an exemption still holds should be reading anyway.
func firstSentence(s string) string {
	if i := strings.Index(s, ". "); i > 0 {
		return s[:i+1]
	}
	return s
}
