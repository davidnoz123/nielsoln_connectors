// rowpatch_test.go -- one contract, run twice.
//
// AGENTS' rule applied: the fake half is a patch this test writes, which is
// fast and lets the awkward refusals be arranged deliberately. ⚠️ The real
// half reads a patch written by PYTHON's patchbucket, and it is not optional
// and not a formality -- the canonical form is exactly the kind of belief a
// fake would encode from the same assumption the code did. A Go-written patch
// that only ever satisfies Go proves nothing at all about interop.
package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func samplePatch() map[string]any {
	return map[string]any{
		"schema_version": 1,
		"machine_id":     "machine_A",
		"seq":            "000001",
		"exported_at":    1700000000.0,
		"description":    "test",
		"conversations": []any{map[string]any{
			"id": "c1", "title": "x & y < z", "create_time": 1700000000.0,
			"update_time": 1700000000.5, "model": nil,
			"raw_json": `{"a":1}`,
		}},
		"messages": []any{map[string]any{
			"id": "m1", "conversation_id": "c1", "role": "user",
			"content": "hi", "create_time": 1700000000.0,
			"turn_index": 0, "source": "zip",
		}},
		"scraped_sessions": []any{},
		"blobs":            []any{},
		"deleted_sessions": []any{},
	}
}

func TestWriteThenReadVerifies(t *testing.T) {
	p := filepath.Join(t.TempDir(), "patch-000001.json.gz")
	h, err := WritePatch(p, samplePatch())
	if err != nil {
		t.Fatal(err)
	}
	got, err := ReadPatch(p)
	if err != nil {
		t.Fatalf("a patch we just wrote failed its own verification: %v", err)
	}
	if got.Field("patch_hash") != h {
		t.Errorf("hash %q in the file, %q returned", got.Field("patch_hash"), h)
	}
	if got.TotalRows() != 2 {
		t.Errorf("got %d rows, want 2", got.TotalRows())
	}
	if n := got.Rows()["conversations"]; n != 1 {
		t.Errorf("got %d conversations, want 1", n)
	}
}

// TestTheHashCoversTheContent is the property the refusal depends on.
func TestTheHashCoversTheContent(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.json.gz")
	if _, err := WritePatch(a, samplePatch()); err != nil {
		t.Fatal(err)
	}
	// Change one character of one row and the hash must no longer match.
	tampered := samplePatch()
	tampered["conversations"].([]any)[0].(map[string]any)["title"] = "x & y < Z"
	b := filepath.Join(dir, "b.json.gz")
	if _, err := WritePatch(b, tampered); err != nil {
		t.Fatal(err)
	}
	pa, _ := ReadPatch(a)
	pb, _ := ReadPatch(b)
	if pa.Field("patch_hash") == pb.Field("patch_hash") {
		t.Error("⚠️ ONE CHANGED CHARACTER PRODUCED THE SAME HASH. The hash " +
			"does not cover the row content, so it guards nothing.")
	}
}

func TestRefusals(t *testing.T) {
	dir := t.TempDir()

	write := func(name string, body map[string]any) string {
		p := filepath.Join(dir, name)
		s, err := canonical(body)
		if err != nil {
			t.Fatal(err)
		}
		f, err := os.Create(p)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		zw := gzip.NewWriter(f)
		if _, err := zw.Write([]byte(s)); err != nil {
			t.Fatal(err)
		}
		if err := zw.Close(); err != nil {
			t.Fatal(err)
		}
		return p
	}

	// ⚠️ THE ONE THAT MATTERS MOST. A patch with no hash must be refused, not
	// accepted unverified: patchbucket's own comment records that validating
	// "when present" made DELETING the field a cheaper forgery than
	// recomputing it. A patch from chatgpt_tools' old hand-rolled log looks
	// exactly like this.
	noHash := samplePatch()
	if _, err := ReadPatch(write("nohash.json.gz", noHash)); err == nil {
		t.Error("a patch with NO patch_hash was accepted")
	} else if !strings.Contains(err.Error(), "missing_patch_hash") {
		t.Errorf("wrong reason for a hash-less patch: %v", err)
	}

	// A hash that does not match the content.
	bad := samplePatch()
	bad["patch_hash"] = strings.Repeat("0", 64)
	if _, err := ReadPatch(write("badhash.json.gz", bad)); err == nil {
		t.Error("a patch with a WRONG patch_hash was accepted")
	} else if !strings.Contains(err.Error(), "invalid_patch_hash") {
		t.Errorf("wrong reason for a bad hash: %v", err)
	}

	// An unknown schema_version, hashed correctly so only the version is wrong.
	v2 := samplePatch()
	v2["schema_version"] = 2
	h, err := HashOf(v2)
	if err != nil {
		t.Fatal(err)
	}
	v2["patch_hash"] = h
	if _, err := ReadPatch(write("v2.json.gz", v2)); err == nil {
		t.Error("schema_version 2 was accepted")
	} else if !strings.Contains(err.Error(), "schema_version") {
		t.Errorf("wrong reason for a future version: %v", err)
	}

	// Not gzip at all.
	plain := filepath.Join(dir, "plain.json.gz")
	if err := os.WriteFile(plain, []byte(`{"schema_version":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadPatch(plain); err == nil {
		t.Error("a non-gzip file was accepted")
	} else if !strings.Contains(err.Error(), "corrupt_gzip") {
		t.Errorf("wrong reason for non-gzip: %v", err)
	}
}

// TestUnknownTablesAreReportedNotSwallowed pins the forward-compatibility rule.
func TestUnknownTablesAreReportedNotSwallowed(t *testing.T) {
	body := samplePatch()
	body["future_table"] = []any{map[string]any{"id": "f1"}}
	p := filepath.Join(t.TempDir(), "future.json.gz")
	if _, err := WritePatch(p, body); err != nil {
		t.Fatal(err)
	}
	got, err := ReadPatch(p)
	if err != nil {
		t.Fatalf("a patch with an extra table was refused: %v", err)
	}
	u := got.UnknownTables()
	if len(u) != 1 || u[0] != "future_table" {
		t.Errorf("got unknown tables %v, want [future_table]", u)
	}
	// ⚠️ And it must survive a rewrite, because the hash covers it. A build
	// that dropped it would produce a patch whose hash is right for content
	// the sender never sent.
	same, _, err := got.Reproduce()
	if err != nil || !same {
		t.Errorf("a patch with an unmodelled table did not reproduce "+
			"byte-identically (err=%v)", err)
	}
}

// TestAgainstAPatchPythonWrote is the second half of the contract.
//
// ⚠️ SKIPS BY NAME rather than passing when absent. AGENTS: say which question
// each run answered. A suite that silently skips the real run reports more
// confidence than it earned, and here the fake half cannot answer the only
// question that matters -- whether Go's canonical form equals Python's.
func TestAgainstAPatchPythonWrote(t *testing.T) {
	const env = "ROWPATCH_PYTHON_PATCH"
	path := os.Getenv(env)
	if path == "" {
		t.Skipf("UNANSWERED: the interop half did not run. Set %s to a "+
			".json.gz written by patchbucket. Everything above describes a "+
			"format Go agreed with ITSELF about.", env)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("%s=%s cannot be read: %v", env, path, err)
	}

	p, err := ReadPatch(path)
	if err != nil {
		t.Fatalf("⚠️ A PATCH PYTHON WROTE WAS REFUSED: %v\n"+
			"If this says invalid_patch_hash then canon.go does not match "+
			"json.dumps(ensure_ascii=False, sort_keys=True) and no "+
			"Go-written patch will ever be accepted by a peer.", err)
	}
	t.Logf("MEASURED: read a Python-written patch, machine_id=%s seq=%s "+
		"rows=%d", p.Field("machine_id"), p.Field("seq"), p.TotalRows())

	// The strong check: not merely that the hash verified, but that Go's
	// canonical form IS Python's output bytes.
	same, got, err := p.Reproduce()
	if err != nil {
		t.Fatal(err)
	}
	if !same {
		t.Fatalf("⚠️ Go's canonical form differs from Python's bytes.\n"+
			"  python: %d bytes\n  go    : %d bytes\n  first divergence at %d",
			len(p.Raw), len(got), firstDiffAt(string(p.Raw), got))
	}
	t.Log("MEASURED: Go's canonical form is BYTE-IDENTICAL to Python's output")

	// And a Go rewrite must verify under Python's rules too.
	out := filepath.Join(t.TempDir(), "rewritten.json.gz")
	body := map[string]any{}
	for k, v := range p.Body {
		if k != hashKey {
			body[k] = v
		}
	}
	h, err := WritePatch(out, body)
	if err != nil {
		t.Fatal(err)
	}
	if h != p.Field("patch_hash") {
		t.Errorf("⚠️ a Go REWRITE produced hash %s where Python's patch "+
			"carries %s. Round-tripping loses something.", h, p.Field("patch_hash"))
	}
	// The rewritten file must be byte-identical to the original gzip payload.
	orig, err := gunzip(path)
	if err != nil {
		t.Fatal(err)
	}
	ours, err := gunzip(out)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(orig, ours) {
		t.Errorf("a Go rewrite differs from the original payload by %d bytes",
			len(ours)-len(orig))
	}
}

func gunzip(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(zr); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func firstDiffAt(a, b string) int {
	i := 0
	for i < len(a) && i < len(b) && a[i] == b[i] {
		i++
	}
	return i
}

// TestCanonicalRefusesAnUnmodelledType pins the refuse-rather-than-guess rule.
func TestCanonicalRefusesAnUnmodelledType(t *testing.T) {
	if _, err := canonical(map[string]any{"t": struct{ A int }{1}}); err == nil {
		t.Error("canonical() accepted a struct. An unmodelled type rendered " +
			"some other way produces a patch whose hash is wrong, and the " +
			"failure surfaces on a peer's machine with nothing to point at.")
	}
}

// TestNumberTextSurvives is the rule that scored 0/32 when got wrong.
func TestNumberTextSurvives(t *testing.T) {
	const src = `{"a": 1700000000.0, "b": 1, "c": 1e-07, "d": 9007199254740993}`
	dec := json.NewDecoder(strings.NewReader(src))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		t.Fatal(err)
	}
	got, err := canonical(v)
	if err != nil {
		t.Fatal(err)
	}
	if got != src {
		t.Errorf("number text was altered.\n got %s\nwant %s\n⚠️ Decoding "+
			"without UseNumber turns every number into a float64, and "+
			"1700000000.0 then renders as 1700000000.", got, src)
	}
}
