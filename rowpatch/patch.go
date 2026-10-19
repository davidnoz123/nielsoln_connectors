// patch.go -- a patchbucket row patch, read and written from Go.
//
// The wire format, measured from patchbucket 0.1.2 (PyPI; verified
// byte-identical to its GitHub source, sha256 a539e112...) and from
// chatgpt_tools' apply_row_patch:
//
//	a gzip'd UTF-8 JSON object, canonical per canon.go, carrying
//	  schema_version  int, must be 1
//	  machine_id      string, the writing peer's stable id
//	  seq             string, "000001", zero-padded to six
//	  exported_at     float, the new watermark
//	  patch_hash      sha256 hex of the canonical form WITHOUT this key
//	  description     string, free text
//	  + the domain tables, each an array of row objects
//
// ⚠️ THE HASH IS MANDATORY AND THE REASON IS WORTH READING. patchbucket
// validates it always rather than "when present", because validating only when
// present made DELETING the field a cheaper forgery than recomputing it.
// Nothing here may relax that.
package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
)

// schemaVersion is the only version this code speaks.
//
// ⚠️ AND IT CANNOT DISTINGUISH TWO FORMATS THAT BOTH CLAIM 1. chatgpt_tools'
// hand-rolled row log also declares schema_version = 1 with a DIFFERENT
// envelope -- no seq, no patch_hash, no canonical ordering. A patch from that
// log passes a version check and then applies as a silent no-op on the Python
// side, because apply_row_patch reads every table with .get(name, []).
// Here it fails on the missing hash instead, which is the better accident.
const schemaVersion = 1

// hashKey is excluded from the hash it holds.
const hashKey = "patch_hash"

// domainTables are the row arrays a patch carries, in chatgpt_tools' schema.
//
// ⚠️ MIRRORED, NOT INVENTED, and the conflict rules matter as much as the
// names -- they are what makes replication idempotent:
//
//	conversations      upsert, but ONLY when the incoming update_time is
//	                   strictly newer than what is stored
//	scraped_sessions   same rule
//	messages           INSERT OR IGNORE: immutable once written
//	blobs              INSERT OR IGNORE: immutable once written
//	deleted_sessions   tombstones, applied last
//
// ⚠️ messages.id IS A GLOBAL PRIMARY KEY, not per-conversation. Two different
// messages sharing an id are ONE ROW and the second is silently dropped by the
// insert-or-ignore. Real ids are UUIDs so it does not arise with real data,
// and a test fixture that reused "n1" across conversations made replication
// look broken when it was the fixture at fault.
var domainTables = []string{
	"conversations", "messages", "scraped_sessions", "blobs",
	"deleted_sessions",
}

// Patch is a decoded row patch. The body stays as a generic map so this code
// never has to agree with chatgpt_tools about a column list.
//
// ⚠️ DELIBERATELY NOT A STRUCT PER TABLE. A struct would silently drop any
// column it did not model, and the rows are written back out to be hashed:
// dropping one produces a patch whose hash is right for the wrong content.
type Patch struct {
	Body map[string]any
	Raw  []byte // the canonical bytes as read, for byte-exactness checks
}

// Field returns a top-level scalar as a string, or "" if absent.
func (p Patch) Field(name string) string {
	switch v := p.Body[name].(type) {
	case string:
		return v
	case json.Number:
		return v.String()
	case nil:
		return ""
	default:
		return fmt.Sprint(v)
	}
}

// Rows returns the row count of each domain table present.
func (p Patch) Rows() map[string]int {
	out := map[string]int{}
	for _, t := range domainTables {
		if rows, ok := p.Body[t].([]any); ok {
			out[t] = len(rows)
		}
	}
	return out
}

// TotalRows is how many rows the patch carries in all.
func (p Patch) TotalRows() int {
	n := 0
	for _, c := range p.Rows() {
		n += c
	}
	return n
}

// UnknownTables are top-level array keys this code does not model.
//
// ⚠️ REPORTED RATHER THAN IGNORED. A new table added on the Python side would
// otherwise pass through unmentioned, and a peer would replicate rows it
// cannot describe. That is not an error -- the hash still covers them and they
// are written back verbatim -- but it must be visible.
func (p Patch) UnknownTables() []string {
	known := map[string]bool{}
	for _, t := range domainTables {
		known[t] = true
	}
	var out []string
	for k, v := range p.Body {
		if _, isArray := v.([]any); isArray && !known[k] {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

// ReadPatch loads a .json.gz patch and VERIFIES its integrity hash.
//
// Errors are returned with patchbucket's own reason strings where they
// correspond, so a Go reader and a Python reader describe the same refusal the
// same way.
func ReadPatch(path string) (*Patch, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	zr, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("corrupt_gzip: %w", err)
	}
	defer zr.Close()
	raw, err := io.ReadAll(zr)
	if err != nil {
		return nil, fmt.Errorf("corrupt_gzip: %w", err)
	}

	// ⚠️ UseNumber so every number keeps its original text. Without it the
	// re-encode needed to verify the hash would reformat floats and the hash
	// would never match a patch Python wrote.
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var body map[string]any
	if err := dec.Decode(&body); err != nil {
		return nil, fmt.Errorf("corrupt_patch_json: %w", err)
	}

	p := &Patch{Body: body, Raw: raw}

	if v := p.Field("schema_version"); v != fmt.Sprint(schemaVersion) {
		return nil, fmt.Errorf("unrecognised schema_version %q", v)
	}
	stored, ok := body[hashKey].(string)
	if !ok || stored == "" {
		return nil, fmt.Errorf("missing_patch_hash")
	}
	computed, err := HashOf(body)
	if err != nil {
		return nil, err
	}
	if computed != stored {
		return nil, fmt.Errorf("invalid_patch_hash: stored %s, computed %s",
			stored, computed)
	}
	if p.Field("machine_id") == "" {
		return nil, fmt.Errorf("missing_patch_machine_id")
	}
	if p.Field("seq") == "" {
		return nil, fmt.Errorf("missing_patch_seq")
	}
	return p, nil
}

// HashOf computes the integrity hash of a patch body, excluding patch_hash.
func HashOf(body map[string]any) (string, error) {
	without := make(map[string]any, len(body))
	for k, v := range body {
		if k != hashKey {
			without[k] = v
		}
	}
	s, err := canonical(without)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", sum), nil
}

// WritePatch canonicalises a body, stamps its hash, and writes the .json.gz.
//
// ⚠️ THE HASH IS COMPUTED BEFORE IT IS INSERTED, over the body without it,
// which is the only ordering that can ever verify. Getting this backwards
// produces a file that fails its own check.
func WritePatch(path string, body map[string]any) (string, error) {
	h, err := HashOf(body)
	if err != nil {
		return "", err
	}
	stamped := make(map[string]any, len(body)+1)
	for k, v := range body {
		stamped[k] = v
	}
	stamped[hashKey] = h

	s, err := canonical(stamped)
	if err != nil {
		return "", err
	}

	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	zw := gzip.NewWriter(f)
	if _, err := zw.Write([]byte(s)); err != nil {
		return "", err
	}
	if err := zw.Close(); err != nil {
		return "", err
	}
	return h, nil
}

// Reproduce re-canonicalises a patch as read and reports whether the bytes are
// identical to what was on disk.
//
// ⚠️ THIS IS THE INTEROP TEST, and it is stronger than "the hash verified".
// A hash match proves the canonical form of the DECODED body matches; this
// proves Go's canonical form equals PYTHON'S OWN OUTPUT BYTES. They can differ
// only if the hash was computed over something other than the file contents,
// which is exactly the kind of mistake worth catching once.
func (p Patch) Reproduce() (bool, string, error) {
	s, err := canonical(p.Body)
	if err != nil {
		return false, "", err
	}
	return s == string(p.Raw), s, nil
}
