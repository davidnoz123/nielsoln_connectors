// verdict.go -- read the handler's own account of what it did.
//
// ⚠️ THE HANDLER'S LOG IS THE ONLY CHANNEL. It runs as a separate process the
// OS started, so there is no return value to inspect. It logs the raw URI
// before parsing, then an identity and a derived argv if the link was good,
// then a refusal if it was not.
//
// ⚠️ AND THE ONE TRAP, which would invert every result. A VALID link also
// ends in a refusal today:
//
//	click REFUSED: the KEEPER is not built. The link is valid and its
//	identity is above
//
// So acceptance is decided by the PRESENCE OF AN IDENTITY LINE, never by the
// absence of the word REFUSED. Keying on "REFUSED" would classify every
// correct link as rejected and every rejection as correct.
package main

import (
	"strconv"
	"strings"
)

// HandlerSaw is what the log reveals about one click.
type HandlerSaw struct {
	Verdict Verdict

	// Raw is the URI as delivered, verbatim.
	Raw string

	// Identity and IdentityBytes come from the identity line.
	Identity      string
	IdentityBytes int

	// ArgvLog is the `would run [...]` text, compared as a string against
	// what the handler would print for the expected argv.
	ArgvLog string

	// Action is logged separately, with the handler itself stating that it
	// is not part of identity.
	Action string

	// Refusal is the reason text, for either kind of refusal.
	Refusal string

	// KeeperRefusal is true when the refusal was "the keeper is not built",
	// which means the link PARSED. Distinguished so that building the keeper
	// does not silently change every corpus result.
	KeeperRefusal bool

	// Aborted is true when the handler reached the CONSENT GATE and nobody
	// answered. ⚠️ Added 17 Oct 2026 because s67 made a one-shot RUN, so
	// the four accepted shapes stopped hitting the keeper refusal and
	// started hitting the gate instead. readHandler could not see the line,
	// so the corpus kept passing while measuring something else.
	Aborted bool

	// Ran is true when the handler EXECUTED something.
	//
	// ⚠️ THE INVARIANT THIS EXISTS FOR: an unattended corpus run must never
	// execute anything. Nobody is at the keyboard, so consent cannot have
	// been given, so a `click: running` line means the gate let something
	// through without a human. That is the most serious thing this harness
	// could discover and before today it had no way to notice it.
	Ran bool
}

const keeperMarker = "the KEEPER is not built"

// abortedPrefix and runningPrefix are the two lines s67 added. Both must be
// read, because between them they say whether the consent gate held.
const abortedPrefix = "click ABORTED: "
const runningPrefix = "click: running "

// readHandler parses every line written after off into one verdict.
func readHandler(body string) HandlerSaw {
	saw := HandlerSaw{Verdict: NotSeen}
	for _, line := range strings.Split(body, "\n") {
		i := strings.Index(line, arrivalPrefix)
		refused := strings.Index(line, "click REFUSED: ")
		switch {
		case refused >= 0:
			saw.Refusal = strings.TrimSpace(line[refused+len("click REFUSED: "):])
			if strings.Contains(saw.Refusal, keeperMarker) {
				saw.KeeperRefusal = true
			}
		case strings.Contains(line, abortedPrefix):
			saw.Aborted = true
		case strings.Contains(line, runningPrefix):
			saw.Ran = true
		case i < 0:
			continue
		default:
			rest := strings.TrimRight(line[i+len(arrivalPrefix):], "\r")
			switch {
			case strings.HasPrefix(rest, "identity "):
				f := strings.Fields(rest)
				// identity <hex> over <n> bytes
				if len(f) >= 4 {
					saw.Identity = f[1]
					saw.IdentityBytes, _ = strconv.Atoi(f[3])
				}
			case strings.HasPrefix(rest, "would run "):
				saw.ArgvLog = strings.TrimSpace(rest[len("would run "):])
			case strings.HasPrefix(rest, "action "):
				saw.Action = rest
			case strings.Contains(rest, "://"):
				saw.Raw = rest
			}
		}
	}

	// ⚠️ The decision, and the order is the whole point. An identity line
	// means the link parsed, whatever refusal followed it.
	// ⚠️ An identity line means the link PARSED, whatever followed it: the
	// keeper refusal, a consent abort, or a run. Keying on any of those
	// instead would invert the result the day the handler's ending changed,
	// which is exactly what happened on 17 Oct.
	switch {
	case saw.Identity != "":
		saw.Verdict = Accepted
	case saw.Refusal != "":
		saw.Verdict = HandlerRefused
	}
	return saw
}

// CorpusResult is one shape's outcome.
type CorpusResult struct {
	Shape Shape
	Saw   HandlerSaw

	// Problems is empty when the shape behaved exactly as declared. Each
	// entry is a sentence naming what disagreed, because "FAIL" on its own
	// sends a reader back to the source to guess.
	Problems []string
}

func (c CorpusResult) OK() bool { return len(c.Problems) == 0 }

// judge compares what the handler did against what the shape declared.
func judge(s Shape, saw HandlerSaw, scheme string) CorpusResult {
	res := CorpusResult{Shape: s, Saw: saw}
	add := func(f string) { res.Problems = append(res.Problems, f) }

	// ⚠️ THE INVARIANT, CHECKED FOR EVERY SHAPE BEFORE ANYTHING ELSE.
	// Nobody is at the keyboard during a corpus run, so consent cannot have
	// been given. Anything that executed got through the gate without a
	// human, and that matters more than whether the shape matched.
	if saw.Ran {
		add("⚠️ SOMETHING EXECUTED during an unattended run. Consent cannot " +
			"have been given, so the gate in run_windows.go let a link " +
			"through without a human. Treat this before any other finding.")
		return res
	}

	if saw.Verdict != s.Expect {
		add("verdict " + string(saw.Verdict) + ", expected " + string(s.Expect))
		// ⚠️ Returning early: once the verdict is wrong, the detailed
		// comparisons below are noise that buries the one finding.
		return res
	}

	switch s.Expect {
	case Accepted:
		if want := s.ExpectedIdentityBytes(scheme); saw.IdentityBytes != want {
			add("identity computed over " + strconv.Itoa(saw.IdentityBytes) +
				" bytes, expected " + strconv.Itoa(want) +
				" (s44: everything before the first #)")
		}
		if want := s.ExpectedArgvLog(); want != "" && saw.ArgvLog != want {
			add("argv was " + saw.ArgvLog + ", expected " + want +
				" (s43: the runner consumes two segments, the rest are the " +
				"leading arguments)")
		}
		if sent := s.Link(scheme); saw.Raw != sent {
			add("the handler received " + strconv.Itoa(len(saw.Raw)) +
				" bytes, not the " + strconv.Itoa(len(sent)) + " sent")
		}
		// ⚠️ Reported, not asserted. Today every accepted link ends in the
		// keeper refusal; the day the keeper lands, that stops being true
		// and this must not then fail the whole corpus.
	case HandlerRefused:
		if s.Reason == "" {
			add("the shape declares no expected reason, so this would pass " +
				"for a link refused on the wrong grounds")
		} else if !strings.Contains(saw.Refusal, s.Reason) {
			add("refused for the wrong reason. Wanted text containing " +
				strconv.Quote(s.Reason) + ", got " + strconv.Quote(saw.Refusal))
		}
		if saw.KeeperRefusal {
			add("⚠️ this was the KEEPER refusal, which means the link PARSED. " +
				"A shape expecting rejection was accepted.")
		}
	}
	return res
}
