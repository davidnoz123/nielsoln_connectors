// store_handoff.go -- the Python-sqlite backend, as a FILE HANDOFF.
//
// Go writes JSONL and a small sidecar; `chatgpt_tools` reads the JSONL into
// sqlite with FTS5. Neither process calls the other, which is the only shape
// that satisfies caveats.go's "no exec", AGENTS' "Python uses Go; Go does not
// use Python", and s55's closure rule all at once.
//
// # ⚠️ THE WATERMARK IS "CAPTURED", NOT "STORED", AND THE DIFFERENCE BITES
//
// This store advances its watermark when GO has written a conversation. It
// cannot know whether the Python ingest then succeeded: the two never speak.
// So after a failed or skipped ingest, the sidecar claims ground the database
// does not hold, and the next run would skip exactly the conversations that
// never landed.
//
// Said here rather than discovered later, and the sidecar names the hazard in
// its own JSON so somebody reading the file sees it without reading this.
// ⓘ The fix is one line -- delete the sidecar and the next run sweeps fully.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// handoffState is the sidecar beside the JSONL.
type handoffState struct {
	// Watermark is the newest update_time this store has WRITTEN.
	Watermark float64 `json:"watermark"`
	// Count is how many conversations the last run wrote.
	Count int `json:"count"`
	// Warning is carried IN THE FILE on purpose. A reader who finds a stale
	// watermark should not have to find this source file to understand it.
	Warning string `json:"warning"`
}

const handoffWarning = "This watermark is what Go CAPTURED, not what the " +
	"database INGESTED. The two processes never speak. If the Python ingest " +
	"failed or was skipped, delete this file so the next run sweeps fully."

// handoffStore writes one raw conversation per line, plus the sidecar.
type handoffStore struct {
	w        io.Writer
	bw       *bufio.Writer
	outPth   string
	statePth string // empty when writing to stdout: nowhere to put a sidecar
	prior    handoffState
	count    int
	newest   float64
	cbuf     bytes.Buffer
}

// newHandoffStore prepares the store for an -out path, or stdout for "-".
//
// ⚠️ IT DOES NOT OPEN THE OUTPUT FILE. The first version called os.Create
// here, which meant that merely ASKING FOR THE WATERMARK truncated the
// previous capture and leaked a file handle if nothing was then written. Found
// by this module's own tests, which could not clean up their temp directories.
//
// So the file is created on the FIRST Put. A run that reads the watermark and
// then captures nothing leaves the previous output intact, which is the only
// behaviour that makes a watermark safe to consult.
func newHandoffStore(out string) (*handoffStore, error) {
	s := &handoffStore{outPth: out}
	if out == "-" {
		// ⚠️ NO SIDECAR FOR STDOUT, and so no watermark. A pipe has nowhere
		// to keep state, and inventing a location would put a file somewhere
		// the caller did not name.
		s.w = os.Stdout
		s.bw = bufio.NewWriterSize(s.w, 1<<20)
		return s, nil
	}
	s.statePth = out + ".state.json"
	if raw, err := os.ReadFile(s.statePth); err == nil {
		// A malformed sidecar is IGNORED rather than fatal: the cost of
		// ignoring it is a full sweep, and the cost of trusting it is
		// skipping conversations.
		_ = json.Unmarshal(raw, &s.prior)
	}
	return s, nil
}

// open creates the output file, once, on first write.
func (s *handoffStore) open() error {
	if s.bw != nil {
		return nil
	}
	f, err := os.Create(s.outPth)
	if err != nil {
		return fmt.Errorf("cannot write %s: %w", s.outPth, err)
	}
	s.w = f
	s.bw = bufio.NewWriterSize(s.w, 1<<20)
	return nil
}

// Watermark reports the newest update_time already captured.
func (s *handoffStore) Watermark() (float64, string) {
	switch {
	case s.statePth == "":
		return 0, "stdout has no sidecar, so this store cannot say: a FULL sweep"
	case s.prior.Watermark == 0:
		return 0, "no sidecar yet, so nothing has been captured: a FULL sweep"
	default:
		return s.prior.Watermark, fmt.Sprintf(
			"from %s, which the last run wrote after %d conversation(s)",
			s.statePth, s.prior.Count)
	}
}

// Put writes one conversation, compacted and otherwise untouched.
func (s *handoffStore) Put(raw json.RawMessage) error {
	if err := s.open(); err != nil {
		return err
	}
	// ⚠️ Compacted because JSONL cannot carry embedded newlines, and the API
	// returns pretty-printed JSON. json.Compact elides insignificant
	// whitespace only: key order and number text survive, which matters
	// because raw_json on the other side is meant to be the original object.
	s.cbuf.Reset()
	if err := json.Compact(&s.cbuf, raw); err != nil {
		return fmt.Errorf("conversation %d: %w", s.count, err)
	}
	if _, err := s.bw.Write(s.cbuf.Bytes()); err != nil {
		return err
	}
	if _, err := s.bw.Write([]byte{'\n'}); err != nil {
		return err
	}
	s.count++
	return nil
}

// Observe records an update_time so the watermark can advance.
//
// ⚠️ SEPARATE FROM Put AND NOT PART OF THE Store INTERFACE. Put must not read
// fields -- that is the module's whole property -- so the ONE field the
// watermark needs comes from the list, which the capture already parsed for
// its two fields. Keeping it out of the interface stops a future store being
// obliged to understand conversation shape.
func (s *handoffStore) Observe(updateTime float64) {
	if updateTime > s.newest {
		s.newest = updateTime
	}
}

// Commit flushes and writes the sidecar.
func (s *handoffStore) Commit() (int, error) {
	if s.bw == nil {
		// Nothing was ever written, so there is no file and no sidecar.
		return 0, nil
	}
	if err := s.bw.Flush(); err != nil {
		return s.count, err
	}
	if f, ok := s.w.(*os.File); ok && s.statePth != "" {
		if err := f.Close(); err != nil {
			return s.count, err
		}
	}
	if s.statePth == "" || s.count == 0 || s.newest == 0 {
		// ⚠️ NO SIDECAR ON AN EMPTY RUN, and none when nothing was Observed.
		// Writing a watermark for zero conversations would claim ground
		// nothing covered.
		//
		// ⚠️ THE newest == 0 CASE IS THE ZIP HALF, and it was writing a
		// sidecar saying `"watermark": 0` until this was noticed. An export
		// has no watermark to report -- it is a point-in-time dump, not an
		// incremental source -- so there is nothing to record. It only looked
		// harmless because Watermark() reads 0 as "no sidecar yet", which is
		// agreement by accident rather than by design.
		return s.count, nil
	}
	// ⚠️ The watermark only ever MOVES FORWARD. A run capturing older
	// conversations than a previous one must not rewind it, or the next run
	// re-fetches everything in between.
	wm := s.newest
	if s.prior.Watermark > wm {
		wm = s.prior.Watermark
	}
	raw, err := json.MarshalIndent(handoffState{
		Watermark: wm, Count: s.count, Warning: handoffWarning,
	}, "", "  ")
	if err != nil {
		return s.count, err
	}
	if err := os.WriteFile(s.statePth, append(raw, '\n'), 0o644); err != nil {
		// ⚠️ Reported, NOT fatal. The conversations are written; the only loss
		// is that the next run sweeps fully, which is the safe direction.
		return s.count, fmt.Errorf(
			"the conversations were written but the sidecar %s was not, so "+
				"the next run will sweep fully: %w", s.statePth, err)
	}
	return s.count, nil
}
