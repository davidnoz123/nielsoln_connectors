// chatgpt_capture -- get ChatGPT conversations out of an OpenAI export, as
// one raw conversation per line.
//
//	go run github.com/davidnoz123/nielsoln_connectors/chatgpt_capture@<sha> \
//	    -zip <export.zip> -out conversations.jsonl
//
// # WHAT IT DOES, AND DELIBERATELY DOES NOT DO
//
// It opens the zip, finds the `conversations-NNN.json` shards, and writes
// each conversation object to its own line. ⚠️ IT DOES NOT INTERPRET ANY OF
// IT. No title extraction, no message linearising, no model guessing, no
// schema of its own.
//
// That is the whole design, and it is deliberate. `chatgpt_tools`'
// `upsert_conversation()` takes the RAW conversation dict and derives
// everything itself -- id, title, times, model, the linearised turns -- and
// stores `raw_json` as the original object. So anything this program decided
// would be thrown away at best and wrong at worst.
//
// ⚠️ A program that does not interpret cannot misinterpret. That property is
// why the zip path is the first slice: it is offline, deterministic, and has
// no semantics to get wrong, so the module, the JSONL contract, the caveats
// file and the Python intake can all be settled before anything touches a
// live account.
//
// # WHY NOT JUST USE THE PYTHON
//
// Honestly: for the zip alone, barely any reason, and s72 says so. The
// existing `import_from_zip()` already works. What this earns is the shape:
// the same capability will take `-port 9222` and drive an authenticated CDP
// session, and THAT is worth pinning by SHA and auditing. Getting the
// scaffolding right against a 470MB file on disk is cheaper than getting it
// right against a live account.
//
// # THE SEAM
//
//	this program   zip (or later CDP) -> one raw conversation per line
//	chatgpt_tools  that file -> sqlite, including the FTS5 index
//
// ⚠️ The split is forced rather than chosen. `messages_fts` is a virtual
// table using FTS5, whose `_data` holds BLOB segments and whose `_idx` is
// keyed on (segid, term, pgno). Maintaining that from Go means
// reimplementing the internal segment format of a sqlite extension, and this
// repo takes no third-party dependencies. So Go stops at the file.
//
// ⚠️ The Python intake does not exist yet. It is about six lines next to
// `import_conversations_json`, which already holds the upsert logic.
package main

import (
	"archive/zip"
	"bufio"

	"encoding/json"
	"flag"
	"fmt"

	"os"
	"sort"
	"strings"
)

// shardPrefix and shardSuffix select the files that carry conversations.
//
// ⚠️ A 470MB export measured 19 Oct 2026 held 727 entries, of which the
// interesting ones were 100-conversation shards. `chat.html` alone was 79MB
// and is the human-readable render: reading it would triple the work and
// yield nothing.
const (
	shardPrefix = "conversations-"
	shardSuffix = ".json"
)

func main() {
	zipPath := flag.String("zip", "", "an OpenAI export .zip")
	port := flag.Int("port", 0,
		"capture from a signed-in ChatGPT tab on this CDP port instead of a "+
			"zip. Run chrome_ready first")
	since := flag.Float64("since", -1,
		"with -port: only conversations whose update_time is newer than this "+
			"Unix time. ⓘ The default of -1 ASKS THE STORE, which keeps a "+
			"sidecar beside -out, so repeat runs are incremental with no "+
			"bookkeeping by the caller. 0 forces a full sweep")
	out := flag.String("out", "", "where to write the JSONL, or - for stdout")
	max := flag.Int("max", 0,
		"stop after this many conversations, for a quick check. 0 means all")
	flag.Parse()

	// ⚠️ EXACTLY ONE SOURCE. Both would be ambiguous about which produced a
	// line, and neither leaves nothing to do.
	switch {
	case *zipPath == "" && *port == 0:
		flag.Usage()
		fmt.Fprintln(os.Stderr,
			"\nOne of -zip or -port is required, and -out always.\n"+
				"  -zip <export.zip>   the OpenAI export, offline\n"+
				"  -port 9655          a signed-in browser, incremental")
		os.Exit(2)
	case *zipPath != "" && *port != 0:
		fmt.Fprintln(os.Stderr,
			"-zip and -port are both set. Pick one: a line's provenance "+
				"should not depend on reading the flags.")
		os.Exit(2)
	case *out == "":
		fmt.Fprintln(os.Stderr, "-out is required. Use - for stdout.")
		os.Exit(2)
	}

	// ⚠️ ONE STORE FOR BOTH HALVES. The zip and the live capture write the
	// same thing, so they share the destination rather than each opening a
	// file their own way. The Store seam is in store.go, with its own warning
	// that it is unproven until a second implementation exists.
	st, err := newHandoffStore(*out)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	if *port != 0 {
		n, seen, err := runLive(*port, *since, st, *max)
		if err != nil {
			fmt.Fprintf(os.Stderr,
				"FAILED after %d conversation(s): %v\n", n, err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr,
			"%d conversation(s) captured, from %d that had changed\n", n, seen)
		if *since == 0 {
			// ⚠️ A full sweep reads as an incremental one unless it says so.
			fmt.Fprintln(os.Stderr,
				"ⓘ -since was 0, so this was a FULL sweep, forced rather "+
					"than asked of the store.")
		}
		if *max > 0 && n >= *max {
			fmt.Fprintf(os.Stderr,
				"⚠️ STOPPED AT THE -max LIMIT OF %d. This is NOT everything "+
					"that changed.\n", *max)
		}
		return
	}

	n, shards, err := runZip(*zipPath, st, *max)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAILED after %d conversation(s): %v\n", n, err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "%d conversation(s) from %d shard(s)\n", n, shards)
	if *max > 0 && n >= *max {
		// ⚠️ SAID OUT LOUD. A silent cap reads as "that was everything",
		// which is the failure shape this repo is organised against.
		fmt.Fprintf(os.Stderr,
			"⚠️ STOPPED AT THE -max LIMIT OF %d. This is NOT the whole "+
				"export.\n", *max)
	}
}

// runZip streams the shards into the store and returns how many it wrote.
//
// ⓘ Named runZip since -port arrived: `run` no longer said which half.
func runZip(zipPath string, st Store, max int) (int, int, error) {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return 0, 0, fmt.Errorf("opening the zip: %w", err)
	}
	defer zr.Close()

	var shards []*zip.File
	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, shardPrefix) &&
			strings.HasSuffix(f.Name, shardSuffix) {
			shards = append(shards, f)
		}
	}
	if len(shards) == 0 {
		// ⚠️ A REASON, not an empty result. An export with no shards is
		// either the wrong file or a changed format, and "0 conversations"
		// reads as an empty account.
		return 0, 0, fmt.Errorf(
			"no %s*%s entries in %s, out of %d. Either this is not an "+
				"OpenAI conversations export, or the format has changed",
			shardPrefix, shardSuffix, zipPath, len(zr.File))
	}
	// Sorted so a run is reproducible and a resume is conceivable.
	sort.Slice(shards, func(i, j int) bool { return shards[i].Name < shards[j].Name })

	total := 0
	for _, f := range shards {
		n, err := streamShard(f, st, max, total)
		total += n
		if err != nil {
			// Partial results committed, as the live half does.
			kept, _ := st.Commit()
			return kept, len(shards), fmt.Errorf("in %s: %w", f.Name, err)
		}
		if max > 0 && total >= max {
			break
		}
	}
	kept, err := st.Commit()
	return kept, len(shards), err
}

// streamShard writes one line per conversation in a single shard.
//
// ⚠️ STREAMED WITH json.Decoder RATHER THAN json.Unmarshal OF THE WHOLE FILE.
// The shards are up to 18MB each and the export is 470MB; the Python reads a
// whole shard at a time, which works but is not necessary. Decoding element
// by element keeps memory bounded whatever OpenAI ships next.
func streamShard(f *zip.File, st Store, max, soFar int) (int, error) {
	rc, err := f.Open()
	if err != nil {
		return 0, err
	}
	defer rc.Close()

	dec := json.NewDecoder(bufio.NewReaderSize(rc, 1<<20))

	// The shard is a top-level ARRAY. Consume the opening bracket so that
	// dec.More() then walks its elements.
	tok, err := dec.Token()
	if err != nil {
		return 0, fmt.Errorf("reading the opening token: %w", err)
	}
	if d, ok := tok.(json.Delim); !ok || d != '[' {
		// ⚠️ Refused rather than guessed at. A shard that is an object
		// rather than an array means the export format changed, and
		// silently reading nothing would look like an empty export.
		return 0, fmt.Errorf(
			"expected a JSON array at the top level, got %v. The export "+
				"format has changed and this program should be re-read "+
				"before being trusted", tok)
	}

	n := 0
	for dec.More() {
		if max > 0 && soFar+n >= max {
			return n, nil
		}
		// ⚠️ RawMessage, so the bytes pass through WITHOUT being re-encoded.
		// Re-marshalling a decoded map would reorder keys and reformat
		// numbers, and `raw_json` on the other side is meant to be the
		// original object.
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return n, fmt.Errorf("decoding conversation %d: %w", n, err)
		}
		// ⓘ The compacting that used to be here now lives in the store, so
		// both halves get it from one place rather than each doing its own.
		if err := st.Put(raw); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}
