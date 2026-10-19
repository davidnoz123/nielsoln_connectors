// capture_test.go -- a synthetic export, then the real 470MB one.
//
// AGENTS' rule, applied: one contract run twice, once against a fast fake and
// once against the real dependency. The fake here is a zip this test builds,
// which is fast and lets the awkward cases be arranged deliberately. ⚠️ The
// real run is not optional and is not a formality: the format of an OpenAI
// export is exactly the kind of belief a fake would encode from the same
// assumption the code did.
package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// buildZip writes a zip with the given entries and returns its path.

// memStore is a Store that keeps everything in memory, for the tests.
//
// ⚠️ IT IS THE INTERFACE'S SECOND IMPLEMENTATION, and that is worth saying
// plainly because store.go warns the seam is unproven until one exists. This
// is a WEAK second: it is a buffer, not a backend, so it exercises the shape
// of the interface without testing any storage decision. The warning in
// store.go therefore still stands -- a real second store may still find the
// seam wrong.
//
// What it does prove is narrower and still useful: that nothing in the
// capture needs more from a Store than these three methods.
type memStore struct {
	buf       bytes.Buffer
	cbuf      bytes.Buffer
	n         int
	watermark float64
	how       string
}

func (m *memStore) Watermark() (float64, string) {
	if m.how == "" {
		return m.watermark, "memStore, set by the test"
	}
	return m.watermark, m.how
}

func (m *memStore) Put(raw json.RawMessage) error {
	m.cbuf.Reset()
	if err := json.Compact(&m.cbuf, raw); err != nil {
		return err
	}
	m.buf.Write(m.cbuf.Bytes())
	m.buf.WriteByte('\n')
	m.n++
	return nil
}

func (m *memStore) Commit() (int, error) { return m.n, nil }

func (m *memStore) String() string { return m.buf.String() }

func buildZip(t *testing.T, entries map[string]string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "export.zip")
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	for name, body := range entries {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return p
}

const twoConvs = `[
  {
    "conversation_id": "aaa",
    "title": "first",
    "create_time": 1.5,
    "mapping": {"n1": {"id": "n1"}}
  },
  {
    "conversation_id": "bbb",
    "title": "second",
    "update_time": 2.5,
    "mapping": {}
  }
]`

func TestOneLinePerConversation(t *testing.T) {
	z := buildZip(t, map[string]string{
		"conversations-000.json": twoConvs,
		// ⚠️ Noise that must be ignored. The real export carries a 79MB
		// chat.html and 700-odd other entries.
		"chat.html":                          "<html>not json at all</html>",
		"conversation_asset_file_names.json": `{"not":"a shard"}`,
	})
	var out memStore
	n, shards, err := runZip(z, &out, 0)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 || shards != 1 {
		t.Fatalf("got %d conversation(s) from %d shard(s), want 2 from 1", n, shards)
	}
	lines := strings.Split(strings.TrimRight(out.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d line(s), want 2:\n%s", len(lines), out.String())
	}
	// ⚠️ EACH LINE MUST BE ONE COMPLETE JSON OBJECT. A pretty-printed source
	// carries newlines, and a single stray one turns one conversation into
	// two unparseable halves on the other side.
	for i, line := range lines {
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Errorf("line %d is not one JSON object: %v\n%s", i, err, line)
		}
		if strings.ContainsAny(line, "\n\r") {
			t.Errorf("line %d has an embedded newline", i)
		}
	}
}

// TestTheBytesPassThroughUNINTERPRETED is the property the design rests on.
//
// ⚠️ chatgpt_tools' upsert_conversation derives id, title, times, model and
// the linearised turns from the RAW dict, and stores raw_json as the
// original. So this program must not reorder, rename, coerce or drop
// anything. A re-marshalled map would reorder keys and reformat numbers.
func TestTheBytesPassThroughUNINTERPRETED(t *testing.T) {
	// Key order here is deliberately NOT alphabetical, and the number has a
	// form a float round-trip would change.
	const odd = `[{"zzz":1,"conversation_id":"x","create_time":1730000000.123456,"aaa":{"b":[1,2,3]}}]`
	z := buildZip(t, map[string]string{"conversations-000.json": odd})
	var out memStore
	if _, _, err := runZip(z, &out, 0); err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(out.String())
	want := `{"zzz":1,"conversation_id":"x","create_time":1730000000.123456,"aaa":{"b":[1,2,3]}}`
	if got != want {
		t.Errorf("the bytes were altered.\n got %s\nwant %s\n"+
			"⚠️ Key order or number formatting changed, which means raw_json "+
			"on the other side would not be the original object.", got, want)
	}
}

func TestNoShardsIsAnErrorWithAReason(t *testing.T) {
	z := buildZip(t, map[string]string{"chat.html": "<html></html>"})
	var out memStore
	n, _, err := runZip(z, &out, 0)
	if err == nil {
		t.Fatal("a zip with no conversation shards succeeded, reporting 0 " +
			"conversations. That reads as an empty account.")
	}
	if n != 0 {
		t.Errorf("wrote %d conversation(s) from a zip with no shards", n)
	}
	for _, want := range []string{"conversations-", "format has changed"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the error does not mention %q, so a caller cannot tell "+
				"a wrong file from a changed format: %v", want, err)
		}
	}
}

// TestAnObjectAtTheTopLevelIsRefused pins the format check.
func TestAnObjectAtTheTopLevelIsRefused(t *testing.T) {
	z := buildZip(t, map[string]string{
		"conversations-000.json": `{"conversations": []}`,
	})
	var out memStore
	if _, _, err := runZip(z, &out, 0); err == nil {
		t.Fatal("an OBJECT at the top level was accepted. The real shards " +
			"are arrays, so this would silently read nothing and look like " +
			"an empty export.")
	} else if !strings.Contains(err.Error(), "array") {
		t.Errorf("the error does not say what was expected: %v", err)
	}
}

func TestShardsAreReadInOrder(t *testing.T) {
	z := buildZip(t, map[string]string{
		"conversations-002.json": `[{"conversation_id":"c"}]`,
		"conversations-000.json": `[{"conversation_id":"a"}]`,
		"conversations-001.json": `[{"conversation_id":"b"}]`,
	})
	var out memStore
	n, shards, err := runZip(z, &out, 0)
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 || shards != 3 {
		t.Fatalf("got %d from %d shards, want 3 from 3", n, shards)
	}
	// Order matters for reproducibility, and a zip's entry order is whatever
	// the writer chose.
	if i, j, k := strings.Index(out.String(), `"a"`),
		strings.Index(out.String(), `"b"`),
		strings.Index(out.String(), `"c"`); !(i < j && j < k) {
		t.Errorf("shards were not read in name order:\n%s", out.String())
	}
}

// TestTheMaxLimitSaysItTruncated is the no-silent-caps rule.
func TestTheMaxLimitSaysItTruncated(t *testing.T) {
	z := buildZip(t, map[string]string{
		"conversations-000.json": `[{"conversation_id":"a"},{"conversation_id":"b"},{"conversation_id":"c"}]`,
	})
	var out memStore
	n, _, err := runZip(z, &out, 2)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Errorf("got %d with -max 2, want 2", n)
	}
	// ⚠️ main() prints the warning; this pins that the COUNT reaches the
	// limit exactly, which is what main() keys the warning on. Without that
	// the cap would be silent, and a silent cap reads as "that was all".
	if lines := strings.Count(strings.TrimSpace(out.String()), "\n") + 1; lines != 2 {
		t.Errorf("wrote %d line(s) with -max 2", lines)
	}
}

// TestAgainstTheRealExport is the second half of the contract.
//
// ⚠️ SKIPS BY NAME rather than passing when the file is absent. AGENTS: "say
// which question each run answered", and a suite that silently skips the real
// run reports more confidence than it earned.
func TestAgainstTheRealExport(t *testing.T) {
	const env = "CHATGPT_EXPORT_ZIP"
	p := os.Getenv(env)
	if p == "" {
		t.Skipf("UNANSWERED: the real-export half did not run. Set %s to an "+
			"OpenAI export .zip. The synthetic cases above describe a format "+
			"this test has not confirmed.", env)
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("%s=%s cannot be read: %v", env, p, err)
	}
	var out memStore
	n, shards, err := runZip(p, &out, 250)
	if err != nil {
		t.Fatalf("the real export failed: %v", err)
	}
	if n == 0 {
		t.Fatal("the real export yielded no conversations")
	}
	t.Logf("MEASURED: %d conversation(s) read from %d shard(s) (capped at 250)",
		n, shards)
	// Every line must parse, and carry the one field the Python requires.
	for i, line := range strings.Split(strings.TrimRight(out.String(), "\n"), "\n") {
		var m map[string]json.RawMessage
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("real line %d does not parse: %v", i, err)
		}
		if _, ok := m["conversation_id"]; !ok {
			if _, ok2 := m["id"]; !ok2 {
				t.Fatalf("real line %d has neither conversation_id nor id, "+
					"so upsert_conversation would skip it", i)
			}
		}
	}
}
