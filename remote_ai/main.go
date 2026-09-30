// main.go -- the connector that ships to participants.
//
//	go build -ldflags="-X main.defaultHost=<host>" -o nielsoln-bridge.exe .
//
// Same protocol as connector.py, same answers, checked against the same
// fixtures/ that the Python one produced:
//
//	python replay.py --connector nielsoln-bridge.exe
//
// Go because fifteen people have to run this on their own laptops without
// installing anything. Stdlib only -- newline-delimited JSON over TCP needs
// nothing, and the one dependency the house rules allow is spoken for by the
// production WebSocket. Never packed with UPX: a quarantine in front of a room
// is worse than four megabytes.
//
// It dials out and never listens. A listening socket on a participant's laptop
// raises Windows Firewall's "Allow access?" with an admin prompt, and that
// stops people dead.
//
// ONE FILE, and the rule is not a preference. Everything the connector does
// belongs here: path containment, UTF-8 boundaries, magic bytes, the six ops,
// the id cache, WebSocket framing. The only files allowed beside it are
// links_windows.go and links_other.go, and they are split because the
// compiler requires it rather than because it reads better:
// syscall.Win32FileAttributeData does not exist off Windows, and provision.sh
// builds and vets this on a Linux VM. Merging them in was tried -- the
// Windows build succeeds and the Linux build fails with "undefined:
// syscall.Win32FileAttributeData". Build constraints are per file, so there
// is no arrangement that keeps that code here and still compiles there. The
// _windows suffix is itself an implicit constraint, so the split is doubly
// enforced.
//
// Why one file at all: the claim this binary makes is that it is a
// thousand-odd lines of Go and nothing else, checkable in an afternoon by
// somebody deciding whether to run it on their own laptop. A package tree
// makes that audit harder, not easier. Resist splitting by concern; the
// concern is the program.
package main

import (
	"bufio"
	"bytes"
	"compress/flate"
	"crypto/rand"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	defaultLimit = 64 * 1024
	// What Claude Code's Read returns when nobody says otherwise, matched
	// so a caller that omits `lines` gets the same amount of file it is
	// used to. Still bounded by defaultLimit in bytes, because 2000 lines
	// of a minified file is not 2000 lines of prose.
	defaultReadLines = 2000
	maxLine      = 1 << 20
	version      = "draft-1" // the protocol draft implemented, not a repo version
)

var capabilities = []string{"utf8_text"}

// defaultHost is where this connector dials when nobody says otherwise.
//
// Stamped at build time rather than compiled in, because the answer depends on
// who is serving the binary and the participant has no command line:
//
//	go build -ldflags="-X main.defaultHost=bridge.nielsoln.com" \
//	    -o nielsoln-bridge.exe .
//
// Not -s -w: see provision.sh. Windows Defender quarantines the stripped
// build as Trojan:Win32/Sabsik.TE.A!ml once it carries Mark-of-the-Web.
//
// A connector downloaded from the internet that dials 127.0.0.1 reaches the
// participant's own machine and finds nothing -- the failure looks like a
// broken download rather than a wrong address.
var defaultHost = "127.0.0.1"

// defaultPort is stamped the same way, and is a string for the same reason
// defaultHost is one: -X writes into a string symbol and does nothing
// whatsoever to anything else. It was an int in the const block above until
// 21 Sep, which would have produced a binary that builds, links, reports no
// error and dials 8790 for ever.
//
// 443 in the shipped build, because --transport auto reads this: 443 means
// wss through Caddy, anything else means plain TCP. 8790 unstamped keeps
// every local test and all 25 fixtures on the harness transport.
var defaultPort = "8790"

// -- wire ------------------------------------------------------------------

type request struct {
	ID   int             `json:"id"`
	Op   string          `json:"op"`
	Args json.RawMessage `json:"args"`
}

type response struct {
	ID     int      `json:"id"`
	OK     bool     `json:"ok"`
	Result any      `json:"result,omitempty"`
	Error  *opError `json:"error,omitempty"`
}

type opError struct {
	Code    string `json:"code"`    // the closed set in PROTOCOL.md
	Message string `json:"message"` // written for a participant, not a developer
}

type pathArgs struct {
	Path string `json:"path"`
}

type readFileArgs struct {
	Path   string `json:"path"`
	Offset int64  `json:"offset"`
	Limit  int    `json:"limit"`
	// Line addressing, added 27 Sep. Claude Code's Read takes LINE numbers and
	// this took bytes, so "read lines 500 to 600" could not be said at all and
	// the model read the whole file to find them.
	//
	// FirstLine is 1-based, like every editor and like Read. Zero means absent,
	// which is why it is not a pointer: line 0 does not exist, so there is no
	// value to confuse with the default.
	FirstLine int `json:"first_line"`
	Lines     int `json:"lines"`
	// "text" (default) or "base64". The CALLER asks, rather than the
	// connector deciding per file: a read that silently changed shape
	// depending on what it found would make every caller handle both, and
	// the text path is the one that must not move.
	Encoding string `json:"encoding"`
}

type readFileResult struct {
	Path       string `json:"path"`
	Content    string `json:"content"`
	EOF        bool   `json:"eof"`
	BytesSent  int    `json:"bytes_sent"`
	TotalBytes int64  `json:"total_bytes"`
	// Only set when the caller asked in lines, so a byte reader sees exactly
	// what it saw before. FirstLine echoes what was asked; LinesSent is what
	// came back, which is fewer at the end of the file.
	FirstLine  int `json:"first_line,omitempty"`
	LinesSent  int `json:"lines_sent,omitempty"`
	NextLine   int `json:"next_line,omitempty"`
	// Set only when it is not utf8, so a text result is byte for byte what
	// it was before this existed and every fixture recorded until 27 Sep
	// still describes it exactly.
	Encoding string `json:"encoding,omitempty"`
}

// lineWindow finds the byte range covering `count` lines starting at line
// `first`, 1-based. Returns the start offset, the end offset and how many
// lines that actually is, which is fewer than asked at the end of the file.
//
// It SCANS rather than buffering. Finding line 500 costs reading the first 499
// lines, which is unavoidable without an index and is still one pass with a
// fixed buffer rather than the whole file in memory. A caller reading forward
// should pass next_line back, not re-count from the top each time.
func lineWindow(f *os.File, first, count int) (start, end int64, lines int, err error) {
	if _, err = f.Seek(0, io.SeekStart); err != nil {
		return
	}
	r := bufio.NewReaderSize(f, 64*1024)
	var pos int64
	at := 1
	for at < first {
		var chunk []byte
		chunk, err = r.ReadSlice('\n')
		pos += int64(len(chunk))
		if err != nil {
			if err == bufio.ErrBufferFull {
				// A line longer than the buffer. Keep going: it is still one
				// line, it just takes several reads to cross.
				err = nil
				continue
			}
			if err == io.EOF {
				// Asked to start past the end. Not an error: an empty answer
				// with eof set says so, the same as a byte offset past the end.
				err = nil
				return pos, pos, 0, nil
			}
			return
		}
		at++
	}
	start = pos
	for lines < count {
		var chunk []byte
		chunk, err = r.ReadSlice('\n')
		pos += int64(len(chunk))
		if err != nil {
			if err == bufio.ErrBufferFull {
				err = nil
				continue
			}
			if err == io.EOF {
				err = nil
				if len(chunk) > 0 {
					lines++ // a last line with no newline after it
				}
				return start, pos, lines, nil
			}
			return
		}
		lines++
	}
	return start, pos, lines, nil
}

type writeFileArgs struct {
	Path       string `json:"path"`
	Content    string `json:"content"`
	CreateOnly bool   `json:"create_only"`
	// Offset is a POINTER so that absent and zero are different. Absent means
	// "replace the file", which is what every caller before 25 Sep meant and
	// what the op must keep doing. Zero means "write at the start and leave
	// the rest", which is a different thing and now sayable.
	Offset   *int64 `json:"offset"`
	Truncate bool   `json:"truncate"`
}

type writeFileResult struct {
	Path         string `json:"path"`
	BytesWritten int    `json:"bytes_written"`
	Offset       int64  `json:"offset"`
	Size         int64  `json:"size"`
}

type entry struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
	Size int64  `json:"size"`
}

type listDirectoryResult struct {
	Path    string  `json:"path"`
	Entries []entry `json:"entries"`
}

type getFileInfoResult struct {
	Path     string `json:"path"`
	Kind     string `json:"kind"`
	Size     int64  `json:"size"`
	Modified string `json:"modified"`
}

type helloArgs struct {
	Token        string   `json:"token"`
	Platform     string   `json:"platform"`
	Root         string   `json:"root"`
	Version      string   `json:"version"`
	Capabilities []string `json:"capabilities"`
	ResumeOf     string   `json:"resume_of,omitempty"`
}

// refusal is an error from the closed set, carrying a message for a human.
type refusal struct {
	code    string
	message string
}

func (r *refusal) Error() string { return r.message }

func refuse(code, format string, a ...any) error {
	return &refusal{code: code, message: fmt.Sprintf(format, a...)}
}

func logf(format string, a ...any) {
	fmt.Printf("%s %s\n", time.Now().Format("15:04:05"), fmt.Sprintf(format, a...))
}

// encode marshals without Go's default HTML escaping.
//
// encoding/json turns <, > and & into < and friends unless told not to.
// Python's json does not, so the two implementations would produce different
// bytes for the same result and every fixture containing a pound sign in a
// quoted string would diverge for a reason that has nothing to do with this
// protocol.
func encode(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil // Encode already appends the newline
}

// -- paths -----------------------------------------------------------------

// resolve turns raw into an absolute path, following symlinks first and then
// checking containment. The order matters: checking containment on the
// unresolved path lets a symlink inside the root hand out anything on the disk.
func resolve(root, raw string) (string, error) {
	p := raw
	if p == "" {
		p = root
	}
	// A path that starts with a separator was meant as an absolute one, even
	// when Windows disagrees -- "/mnt/c/analytics" is absolute on the machine
	// Claude is running on and merely drive-less here. Joining it to the root
	// produced "C:\Users\you\Documents\mnt\c\analytics", and the refusal then
	// named a folder from the middle of a path nobody asked for. Refusing it
	// as outside is both true and something Claude can act on.
	if !filepath.IsAbs(p) && (strings.HasPrefix(p, "/") || strings.HasPrefix(p, `\`)) {
		return "", refuse("not_allowed",
			"%s is not inside the folder shared for this session, which is %s.",
			raw, root)
	}
	if !filepath.IsAbs(p) {
		p = filepath.Join(root, p)
	}
	p = filepath.Clean(p)

	// No links, anywhere under the root. Checked BEFORE any resolution,
	// because resolution is what cannot be trusted here.
	//
	// filepath.EvalSymlinks does not follow a Windows junction. Measured
	// 18 Sep 2026 against a junction inside the shared folder: EvalSymlinks
	// returned the path unchanged, containment therefore passed, and ReadDir
	// then followed the junction and listed a folder outside the share. The
	// containment check was being asked a question about a path that had
	// never been resolved.
	//
	// So links are refused rather than resolved. A shared folder of somebody's
	// documents has no need of them, and they are the only way a path inside
	// the root can name a file outside it. Refusing is a guarantee;
	// resolving correctly on every filesystem is a hope.
	if bad := firstLink(root, p); bad != "" {
		return "", refuse("not_allowed",
			"%s is a shortcut to somewhere else, so it is not readable in this "+
				"session. The folder shared for this session is %s.",
			filepath.Base(bad), root)
	}

	// EvalSymlinks fails on a path that does not exist yet, which write_file
	// legitimately produces. Fall back to resolving the parent, so a symlinked
	// directory still cannot be used to escape.
	real, err := filepath.EvalSymlinks(p)
	if err != nil {
		// The fallback below exists for a path that does not exist YET, which
		// write_file legitimately produces. It must not be used for a path
		// that exists and merely will not resolve: joining the parent's real
		// path to the base name yields something that passes containment
		// without the final component ever having been followed, and if the
		// open then succeeds we would read whatever it points at.
		//
		// Found in the wild, 18 Sep 2026, on a participant's own Documents
		// folder. Windows keeps three legacy junctions there -- My Music, My
		// Pictures, My Videos -- pointing at the sibling user folders, with
		// access denied. Every listing shows them, EvalSymlinks fails on
		// them, and this fallback waved them through containment. Only the
		// subsequent Stat failing kept it honest, and it reported
		// "There is no file called My Pictures there", which is false.
		if _, lerr := os.Lstat(p); lerr == nil {
			return "", refuse("not_allowed",
				"%s cannot be followed, so it cannot be shown to be inside the "+
					"folder shared for this session, which is %s.",
				filepath.Base(p), root)
		}
		parent, err2 := filepath.EvalSymlinks(filepath.Dir(p))
		if err2 != nil {
			return "", refuse("not_found",
				"There is nothing at %s. The folder shared for this session is %s.",
				raw, root)
		}
		real = filepath.Join(parent, filepath.Base(p))
	}

	if !contained(root, real) {
		// not_allowed, never not_found. "No such file" sends Claude hunting for
		// the right name; "not allowed" tells it to stop.
		// Naming the shared folder is what lets Claude correct itself. Without
		// it, a refusal is a dead end: it cannot tell whether it guessed the
		// wrong path or has no access at all, and it concludes the latter.
		return "", refuse("not_allowed",
			"That is outside the folder shared for this session, which is %s.", root)
	}
	return real, nil
}

// firstLink returns the first path component at or below root that is a link,
// or "" if there is none. p must already be cleaned and absolute.
//
// Walked one component at a time from the root downwards, because a link
// anywhere along the way redirects everything after it. Checking only the
// final component would let "My Pictures/notes.txt" through.
//
// Components at or above the root are not checked. The participant chose that
// folder, and if they typed a path that is itself reached through a link, that
// is their share and their decision -- the question here is only whether a
// path can leave it.
func firstLink(root, p string) string {
	rel, err := filepath.Rel(root, p)
	if err != nil || rel == "." {
		return ""
	}
	if strings.HasPrefix(rel, "..") {
		return "" // outside already; containment reports that more clearly
	}
	at := root
	for _, part := range strings.Split(rel, string(filepath.Separator)) {
		if part == "" {
			continue
		}
		at = filepath.Join(at, part)
		if _, err := os.Lstat(at); err != nil {
			return "" // does not exist yet: write_file's legitimate case
		}
		if isReparse(at) {
			return at
		}
	}
	return ""
}

// skipEntry reports whether a directory entry should be left out of a listing
// because reading it could only be refused.
//
// Only links are considered. An ordinary file that cannot be stat'd stays in
// the listing -- it exists, naming it is useful, and the size comes back as
// -1. A link is refused by resolve, so naming it only invites that refusal.
//
// os.Lstat on the full path rather than it.Info(): a FileInfo from ReadDir
// does not always carry the reparse tag, so a junction can arrive looking
// like an ordinary directory. That is exactly how one got into a listing.
func skipEntry(full string) bool {
	return isReparse(full)
}

func contained(root, real string) bool {
	rel, err := filepath.Rel(root, real)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	// "Rel" is case-sensitive but these filesystems are not, so a lowercase
	// root against a capitalised real path would look like an escape and a
	// folder the participant can see would be reported as outside their own
	// share.
	//
	// macOS belongs here as much as Windows: APFS is case-insensitive by
	// default. Left out, the Mac connector would refuse legitimate paths for
	// a reason no participant could act on.
	if caseInsensitiveFS() {
		rel, _ = filepath.Rel(strings.ToLower(root), strings.ToLower(real))
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// -- what kind of file is this ---------------------------------------------

// Enough to name the thing in a refusal. Claude Code's own Read says "appears
// to be a binary .pyc file" rather than erroring blankly, and naming the type
// is what lets Claude stop guessing instead of retrying.
var magic = []struct {
	prefix []byte
	name   string
}{
	{[]byte("%PDF-"), "a PDF document"},
	{[]byte("\x89PNG\r\n\x1a\n"), "a PNG image"},
	{[]byte("\xff\xd8\xff"), "a JPEG photo"},
	{[]byte("GIF8"), "a GIF image"},
	{[]byte("PK\x03\x04"), "a zip file (Word, Excel and PowerPoint files are zips too)"},
	{[]byte("\x1f\x8b"), "a gzip archive"},
	{[]byte("MZ"), "a Windows program"},
	{[]byte("\x7fELF"), "a Linux program"},
	{[]byte("\xd0\xcf\x11\xe0"), "an old-style Office document"},
}

func magicName(head []byte) string {
	for _, m := range magic {
		if bytes.HasPrefix(head, m.prefix) {
			return m.name
		}
	}
	return ""
}

// trimToRune returns the longest prefix of raw that ends on a character
// boundary, and how many bytes that was. ok is false when raw is not text at
// all rather than merely cut mid-character.
func trimToRune(raw []byte, atEOF bool) (string, int, bool) {
	if utf8.Valid(raw) {
		return string(raw), len(raw), true
	}
	if atEOF {
		return "", 0, false // nothing more is coming; it is simply not text
	}
	// A character is at most 4 bytes, so a boundary is within 3 of the end.
	for n := len(raw) - 1; n >= 0 && n > len(raw)-utf8.UTFMax; n-- {
		if utf8.Valid(raw[:n]) {
			return string(raw[:n]), n, true
		}
	}
	return "", 0, false
}

// -- operations ------------------------------------------------------------

func opPing(root string, args json.RawMessage) (any, error) {
	return map[string]any{"pong": true}, nil
}

func opListDirectory(root string, args json.RawMessage) (any, error) {
	var a pathArgs
	json.Unmarshal(args, &a)
	path, err := resolve(root, a.Path)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, refuse("not_found", "There is no file called %s there.", filepath.Base(path))
	}
	if !info.IsDir() {
		return nil, refuse("not_dir", "%s is a file, not a folder.", filepath.Base(path))
	}
	items, err := os.ReadDir(path)
	if err != nil {
		return nil, refuse("not_allowed", "That folder could not be opened.")
	}
	entries := make([]entry, 0, len(items))
	for _, it := range items {
		// Do not offer what read_file and list_directory will then refuse.
		// A listing that names something the very next call says does not
		// exist is worse than a shorter listing: Claude reports the denial to
		// the participant as fact, and the participant is told a folder they
		// can see in Explorer is not there.
		//
		// This is every Windows Documents folder: My Music, My Pictures and
		// My Videos are hidden junctions to the sibling user folders, outside
		// any share rooted at Documents. Explorer hides them too, so omitting
		// them matches what the person is looking at.
		if skipEntry(filepath.Join(path, it.Name())) {
			continue
		}
		e := entry{Name: it.Name(), Kind: "file"}
		if it.IsDir() {
			e.Kind = "dir"
		} else if st, err := it.Info(); err == nil {
			e.Size = st.Size()
		} else {
			// A file we cannot stat still exists and is worth naming; dropping
			// it silently would make the listing quietly wrong.
			e.Size = -1
		}
		entries = append(entries, e)
	}
	sort.Slice(entries, func(i, j int) bool {
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})
	return listDirectoryResult{Path: path, Entries: entries}, nil
}

func opReadFile(root string, args json.RawMessage) (any, error) {
	var a readFileArgs
	json.Unmarshal(args, &a)
	path, err := resolve(root, a.Path)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, refuse("not_found", "There is no file called %s there.", filepath.Base(path))
	}
	if info.IsDir() {
		return nil, refuse("is_dir", "%s is a folder, not a file.", filepath.Base(path))
	}

	offset := a.Offset
	if offset < 0 {
		offset = 0
	}
	limit := a.Limit
	if limit <= 0 || limit > defaultLimit {
		limit = defaultLimit
	}
	total := info.Size()

	binary := false
	switch strings.ToLower(strings.TrimSpace(a.Encoding)) {
	case "", "text", "utf8":
	case "base64":
		binary = true
	default:
		return nil, refuse("bad_request",
			"encoding is \"text\" or \"base64\", not %q.", a.Encoding)
	}
	if binary && a.FirstLine > 0 {
		// Lines are a property of text. Asking for line 3 of a JPEG is a
		// question with no answer, so it is refused rather than guessed at.
		return nil, refuse("bad_request",
			"first_line counts lines, which binary content does not have. "+
				"Use offset and limit with encoding base64.")
	}
	if a.FirstLine < 0 {
		return nil, refuse("bad_request",
			"A first line cannot be negative, and %d is.", a.FirstLine)
	}
	if a.FirstLine > 0 && a.Offset > 0 {
		// One asks in lines and the other in bytes. Honouring both would mean
		// choosing which the caller meant, and the wrong choice returns
		// plausible text from the wrong part of the file.
		return nil, refuse("bad_request",
			"first_line counts lines and offset counts bytes, so they cannot "+
				"both apply. Use one or the other.")
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, refuse("not_allowed", "%s could not be opened.", filepath.Base(path))
	}
	defer f.Close()

	head := make([]byte, 16)
	nHead, _ := io.ReadFull(f, head)
	head = head[:nHead]

	// Line addressing turns into a byte window here and nowhere else, so the
	// whole of the rest of this op, the UTF-8 boundary walk included, is the
	// code that was already proved.
	lineCount := 0
	if a.FirstLine > 0 {
		want := a.Lines
		if want <= 0 {
			want = defaultReadLines
		}
		start, end, got, lerr := lineWindow(f, a.FirstLine, want)
		if lerr != nil {
			return nil, refuse("internal", "%s could not be read.",
				filepath.Base(path))
		}
		offset, lineCount = start, got
		if span := int(end - start); span < limit {
			limit = span
		}
		if limit == 0 {
			// Past the last line. An empty answer with eof set, which is what
			// a byte offset past the end already does.
			limit = 1
		}
	}

	raw := make([]byte, limit)
	n, err := f.ReadAt(raw, offset)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, refuse("internal", "%s could not be read.", filepath.Base(path))
	}
	raw = raw[:n]
	atEOF := offset+int64(n) >= total

	if binary {
		// Returns BEFORE every check below, because all of them are about
		// text: the magic-byte refusal, the NUL sentinel and the rune
		// boundary walk. A byte range has no character boundaries to respect
		// and no encoding to get wrong, which is the whole reason this is
		// the primitive rather than a format reader.
		return readFileResult{
			Path:       path,
			Content:    base64.StdEncoding.EncodeToString(raw),
			EOF:        offset+int64(n) >= total,
			BytesSent:  n,
			TotalBytes: total,
			Encoding:   "base64",
		}, nil
	}

	// Sniff before decoding, not after. A failed decode is not a reliable test
	// for binary: the first chunk of a real PDF is mostly ASCII object headers,
	// and NUL is perfectly valid UTF-8 -- so a decode-only check reads a PDF
	// happily and hands Claude line noise it will try to interpret.
	if name := magicName(head); name != "" {
		return nil, refuse("not_text",
			"%s looks like %s, so it cannot be read as text. Ask again with "+
				"encoding base64 to get the bytes.",
			filepath.Base(path), name)
	}
	if bytes.IndexByte(raw, 0) >= 0 {
		// The same sentinel git uses. No real text file carries a NUL.
		return nil, refuse("not_text",
			"%s is not a text file, so it cannot be read as text. Ask again "+
				"with encoding base64 to get the bytes.",
			filepath.Base(path))
	}

	if a.FirstLine > 0 && lineCount == 0 {
		// Started past the last line. Say so rather than returning whatever
		// byte happened to be at that offset.
		raw = raw[:0]
		atEOF = true
	}

	text, used, ok := trimToRune(raw, atEOF)
	if !ok {
		return nil, refuse("not_text", "%s is not a text file, so it cannot be read as text.",
			filepath.Base(path))
	}
	if used == 0 && len(raw) > 0 && !atEOF {
		// limit is smaller than the first character, so a naive reader would
		// return nothing, report eof false, and be asked again forever.
		return nil, refuse("too_large",
			"That chunk was too small to hold a single character. Ask again with a larger limit.")
	}

	out := readFileResult{
		Path:       path,
		Content:    text,
		EOF:        offset+int64(used) >= total,
		BytesSent:  used,
		TotalBytes: total,
	}
	if a.FirstLine > 0 {
		// Counted from the text ACTUALLY returned, not from what the window
		// found: the UTF-8 boundary walk above can drop the tail of the last
		// line, and reporting a line as sent when half of it was trimmed is
		// how a caller reading forward loses a line without noticing.
		out.FirstLine = a.FirstLine
		out.LinesSent = strings.Count(text, "\n")
		if out.EOF && text != "" && !strings.HasSuffix(text, "\n") {
			// A final line with no newline after it still counts. A PARTIAL
			// last line mid-file does not: the boundary walk may have cut it,
			// and calling it sent is how a caller reading forward skips the
			// rest of it without noticing.
			out.LinesSent++
		}
		out.NextLine = a.FirstLine + out.LinesSent
	}
	return out, nil
}

// opWriteFile writes a file, whole or in pieces.
//
// read_file has always chunked and write_file never did, so a file larger than
// the message ceiling could be read and not written back. Worse, the failure
// was a dropped connection: the bridge framed a message it knew the far end
// must refuse, the connector refused it, and the socket died without the
// request ever being answered. Nothing in that sequence says "too big", which
// is the one thing the caller could act on.
//
// With `offset` the caller writes in pieces, and the message ceiling stops
// being a ceiling on the FILE.
func opWriteFile(root string, args json.RawMessage) (any, error) {
	var a writeFileArgs
	json.Unmarshal(args, &a)
	path, err := resolve(root, a.Path)
	if err != nil {
		return nil, err
	}
	if a.Offset != nil && *a.Offset < 0 {
		return nil, refuse("bad_request",
			"An offset cannot be negative, and %d is.", *a.Offset)
	}
	if a.Offset != nil && *a.Offset > 0 && a.CreateOnly {
		// create_only asks "does this file already exist"; a non-zero offset
		// says "carry on with the one that does". Together they are a
		// contradiction rather than a combination.
		return nil, refuse("bad_request",
			"create_only writes a new file and an offset continues an "+
				"existing one, so they cannot both apply.")
	}
	if a.CreateOnly {
		if _, err := os.Stat(path); err == nil {
			return nil, refuse("not_allowed", "%s already exists and was left untouched.",
				filepath.Base(path))
		}
	}
	if _, err := os.Stat(filepath.Dir(path)); err != nil {
		return nil, refuse("not_found", "There is no folder called %s to write into.",
			filepath.Base(filepath.Dir(path)))
	}

	data := []byte(a.Content)

	// A piece of a file, written in place. NOT the rename dance: a rename
	// replaces the whole file, which is exactly what the caller is trying to
	// avoid by sending pieces. The cost is that an interrupted chunked write
	// leaves a partly written file, which is why the op reports the size it
	// left behind and why a caller resuming one should ask for it.
	if a.Offset != nil {
		f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o644)
		if err != nil {
			return nil, refuse("not_allowed", "%s could not be written.",
				filepath.Base(path))
		}
		defer f.Close()
		if _, err := f.WriteAt(data, *a.Offset); err != nil {
			return nil, refuse("not_allowed", "%s could not be written at %d.",
				filepath.Base(path), *a.Offset)
		}
		if a.Truncate {
			if err := f.Truncate(*a.Offset + int64(len(data))); err != nil {
				return nil, refuse("not_allowed",
					"%s could not be shortened.", filepath.Base(path))
			}
		}
		if err := f.Sync(); err != nil {
			return nil, refuse("not_allowed", "%s could not be written.",
				filepath.Base(path))
		}
		size := int64(0)
		if st, err := f.Stat(); err == nil {
			size = st.Size()
		}
		return writeFileResult{Path: path, BytesWritten: len(data),
			Offset: *a.Offset, Size: size}, nil
	}

	// Write beside the target and rename. An interrupted write then leaves the
	// original intact rather than a half-written file.
	// The .part file is a path like any other, and it has to earn containment
	// the same way. It did not: resolve() vetted `path`, then this appended a
	// suffix and handed the result straight to os.Create, which follows a
	// symlink. So a pre-existing report.txt.part inside the shared folder,
	// pointing anywhere, turned a legitimate write of report.txt into a write
	// outside the folder -- without the requested path ever being suspicious.
	//
	// Found by an outside review of 8399b67. Worth noting how it hid: the
	// containment check was correct, and the escape was on a path derived
	// AFTER it. A check is only as wide as the paths it is shown.
	tmp := path + ".part"
	// Remove the NAME first, then refuse to create it if anything is still
	// there. os.Remove deletes the directory entry and never what it points
	// at, so this defuses both shapes of the attack without touching the
	// victim's file, and O_EXCL closes the gap between the two calls.
	//
	// resolve() alone was not enough. It refuses symlinks and reparse points,
	// and a HARD link is neither -- so the containment check would have passed
	// it, and `mklink /H` needs no administrator on Windows. The easier half
	// of the attack was the half a path check could not see.
	os.Remove(tmp)
	f, err := os.OpenFile(tmp, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return nil, refuse("not_allowed", "%s could not be written.", filepath.Base(path))
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		os.Remove(tmp)
		return nil, refuse("not_allowed", "%s could not be written.", filepath.Base(path))
	}
	f.Sync()
	f.Close()
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return nil, refuse("not_allowed", "%s could not be written.", filepath.Base(path))
	}
	return writeFileResult{Path: path, BytesWritten: len(data),
		Offset: 0, Size: int64(len(data))}, nil
}

// -- edit_file -------------------------------------------------------------
//
// Answers Claude Code's Edit. Without it the model reads a file and writes it
// back whole, which for a short file is correct and for a long one is two full
// transfers and a window in which somebody else's change is lost. It also
// cannot be done at all to a file only partially read, which chunked reads
// make the normal case.
//
// The refusals are the substance. An edit that matched more places than the
// caller meant is the one mistake here that destroys work rather than failing,
// so ambiguity returns zero and changes nothing.

type editFileArgs struct {
	Path string `json:"path"`
	Old  string `json:"old"`
	New  string `json:"new"`
	All  bool   `json:"all"`
}

type editFileResult struct {
	Path         string `json:"path"`
	Replacements int    `json:"replacements"`
	BytesWritten int    `json:"bytes_written"`
	LineEndings  string `json:"line_endings"`
}

func opEditFile(root string, args json.RawMessage) (any, error) {
	var a editFileArgs
	json.Unmarshal(args, &a)
	path, err := resolve(root, a.Path)
	if err != nil {
		return nil, err
	}
	if a.Old == "" {
		return nil, refuse("bad_request",
			"An edit needs the text to replace. To create or overwrite a file, "+
				"use write_file.")
	}
	if a.Old == a.New {
		return nil, refuse("bad_request",
			"The old and new text are identical, so there is nothing to do.")
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, refuse("not_found", "There is no file called %s there.",
			filepath.Base(path))
	}
	if info.IsDir() {
		return nil, refuse("not_dir", "%s is a folder, not a file.",
			filepath.Base(path))
	}
	if info.Size() > editMaxBytes {
		return nil, refuse("too_large",
			"%s is %d bytes, and editing reads the whole file. The ceiling is %d.",
			filepath.Base(path), info.Size(), editMaxBytes)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, refuse("not_allowed", "%s could not be read.",
			filepath.Base(path))
	}
	head := body
	if len(head) > 16 {
		head = head[:16]
	}
	if name := magicName(head); name != "" {
		return nil, refuse("not_text",
			"%s looks like %s, so it cannot be edited as text.",
			filepath.Base(path), name)
	}

	// CRLF, and it has to be handled rather than hoped about. The model sends
	// "a\nb" because that is what it read: read_file hands back the bytes on
	// disk, but a model composing new text writes \n. On a CRLF file the
	// literal match then finds nothing and the edit refuses a change that is
	// plainly there. So: try exactly, and if that finds nothing on a CRLF
	// file, try again in the file's own endings. The result says which.
	oldText, newText := a.Old, a.New
	endings := "as written"
	count := strings.Count(string(body), oldText)
	if count == 0 && bytes.Contains(body, []byte("\r\n")) {
		crlfOld := strings.ReplaceAll(oldText, "\n", "\r\n")
		if n := strings.Count(string(body), crlfOld); n > 0 {
			oldText = crlfOld
			newText = strings.ReplaceAll(newText, "\n", "\r\n")
			count, endings = n, "crlf"
		}
	}

	if count == 0 {
		return nil, refuse("not_found",
			"That text is not in %s, so nothing was changed. Read the file "+
				"again: it may have moved on since you last saw it.",
			filepath.Base(path))
	}
	if count > 1 && !a.All {
		// The one mistake here that destroys work rather than failing. Naming
		// the count is what lets the caller decide between more context and
		// meaning all of them.
		return nil, refuse("ambiguous",
			"That text appears %d times in %s, so nothing was changed. Include "+
				"enough surrounding lines to name one of them, or set all to "+
				"replace every occurrence.", count, filepath.Base(path))
	}

	replacements := 1
	if a.All {
		replacements = count
	}
	updated := strings.Replace(string(body), oldText, newText, replacements)

	// Beside the target and renamed, the same as write_file: an interrupted
	// edit leaves the original rather than half of it.
	// Same derived path, same treatment. See opWriteFile.
	tmp := path + ".part"
	os.Remove(tmp)
	f, err := os.OpenFile(tmp, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return nil, refuse("not_allowed", "%s could not be written.",
			filepath.Base(path))
	}
	if _, err := f.WriteString(updated); err != nil {
		f.Close()
		os.Remove(tmp)
		return nil, refuse("not_allowed", "%s could not be written.",
			filepath.Base(path))
	}
	f.Sync()
	f.Close()
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return nil, refuse("not_allowed", "%s could not be written.",
			filepath.Base(path))
	}
	return editFileResult{Path: path, Replacements: replacements,
		BytesWritten: len(updated), LineEndings: endings}, nil
}

func opGetFileInfo(root string, args json.RawMessage) (any, error) {
	var a pathArgs
	json.Unmarshal(args, &a)
	path, err := resolve(root, a.Path)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, refuse("not_found", "There is no file called %s there.", filepath.Base(path))
	}
	kind, size := "file", info.Size()
	if info.IsDir() {
		kind, size = "dir", 0
	}
	return getFileInfoResult{
		Path: path,
		Kind: kind,
		Size: size,
		// Python's isoformat writes the offset as +00:00, not Z. Matching it
		// keeps the two implementations byte-identical on the wire.
		Modified: info.ModTime().UTC().Format("2006-01-02T15:04:05.000000-07:00"),
	}, nil
}

const (
	searchLimit   = 200              // matches returned unless asked for fewer
	searchSeconds = 10               // the walk's own deadline, inside the bridge's
	searchMaxScan = 200000           // entries looked at before stopping regardless

	// search_content reads files rather than listing them, so it needs its own
	// ceilings. A repository of source is a few megabytes; a folder someone
	// shared might hold a disk image.
	grepLimit    = 100        // matching lines returned unless asked for fewer
	grepMaxBytes = 2 << 20    // a file larger than this is skipped, and said so
	grepMaxLine  = 64 << 10   // a longer line is not prose and not worth carrying
	// Raised from 2 on 27 Sep. Two was a guess and Grep has no such cap;
	// ten is still small enough that a hundred matches cannot blow the
	// message ceiling, and large enough to see a function around a line.
	grepContext = 10

	// Multiline matching scans a file as one string instead of line by
	// line, so a pattern's cost stops being bounded by the length of a
	// line. Python's engine backtracks and this one does not, so the same
	// pattern can be cheap here and ruinous there, and a ceiling is the
	// only thing that keeps the two answering together. A file above this
	// is skipped and counted in TooBig, exactly as the larger ceiling
	// already does.
	grepMultilineMaxBytes = 256 << 10
	// Matches CONSIDERED in one file before the search gives up on the
	// rest of it. A pattern that can match at nearly every position finds
	// one per byte, and hitting this costs Complete like every other
	// ceiling here.
	grepMaxPerFile = 20000

	// Sorting newest-first cannot stop early: you have to see every candidate
	// before you know which are newest. This bounds what is held while doing
	// so, and hitting it costs `complete`, as every other ceiling here does.
	searchMaxCollect = 5000

	// edit_file reads the whole file to replace inside it. Generous, because
	// editing a large document is a reasonable thing to ask, and bounded
	// because the file arrives in this process's memory.
	editMaxBytes = 8 << 20
)

type searchFilesArgs struct {
	Pattern string `json:"pattern"`
	Path    string `json:"path"`
	Limit   int    `json:"limit"`
	// "path" (default) or "modified". Claude Code's Glob always sorts newest
	// first, which is often the whole reason for reaching for it, and this op
	// could only sort by path.
	Sort string `json:"sort"`
}

type searchMatch struct {
	Path string `json:"path"`
	Kind string `json:"kind"`
	// Only carried when sorting by it. A caller who asked for newest-first
	// and cannot see the times has to take the order on trust.
	Modified string `json:"modified,omitempty"`
	Size int64  `json:"size"`
}

type searchFilesResult struct {
	Matches  []searchMatch `json:"matches"`
	Complete bool          `json:"complete"`
	Scanned  int           `json:"scanned"`
}

// grepTypes is `type`: sugar over `glob`, a language name standing for the
// extensions that language is written in.
//
// It is matched as an EXTENSION rather than expanded into a glob, and that is
// a decision rather than a shortcut. Every language here with more than one
// extension would need *.{py,pyi}, and the glob engine has no brace syntax.
// Adding one would change search_files too, in two languages, for a feature
// only this op asked for, and brace expansion has its own arguments to have
// (nesting, a comma inside a character class). An extension set has none.
//
// Deliberately small. A name that is not here is a refusal that says so and
// names what there is, because a `type` silently matching nothing is the
// absence-reported-as-fact failure this whole op is built around.
var grepTypes = map[string][]string{
	"c":    {".c", ".h"},
	"cpp":  {".cpp", ".cc", ".cxx", ".hpp", ".hh"},
	"cs":   {".cs"},
	"css":  {".css"},
	"csv":  {".csv"},
	"go":   {".go"},
	"html": {".html", ".htm"},
	"ini":  {".ini", ".cfg", ".conf"},
	"java": {".java"},
	"js":   {".js", ".mjs", ".cjs", ".jsx"},
	"json": {".json"},
	"log":  {".log"},
	"md":   {".md", ".markdown"},
	"php":  {".php"},
	"ps1":  {".ps1", ".psm1"},
	"py":   {".py", ".pyi"},
	"rb":   {".rb"},
	"rs":   {".rs"},
	"sh":   {".sh", ".bash", ".zsh"},
	"sql":  {".sql"},
	"tex":  {".tex"},
	"toml": {".toml"},
	"ts":   {".ts", ".mts", ".cts", ".tsx"},
	"txt":  {".txt", ".text"},
	"vb":   {".vb", ".bas", ".cls", ".frm"},
	"xml":  {".xml"},
	"yaml": {".yaml", ".yml"},
}

// The spellings a model actually writes. Claude Code's own Grep description
// offers "py" and "rust" in one breath, so both have to work or the parameter
// refuses the example it was given.
var grepTypeAliases = map[string]string{
	"bash": "sh", "c++": "cpp", "csharp": "cs", "golang": "go",
	"htm": "html", "javascript": "js", "markdown": "md", "powershell": "ps1",
	"python": "py", "ruby": "rb", "rust": "rs", "shell": "sh", "text": "txt",
	"typescript": "ts", "vba": "vb", "yml": "yaml",
}

// grepTypeNames is the sorted list of what grepTypes knows, for the refusal.
// Built once: a map's order is randomised in Go, and a refusal that named the
// same types in a different order every time would be a diff for no reason.
func grepTypeNames() string {
	names := make([]string, 0, len(grepTypes))
	for name := range grepTypes {
		names = append(names, name)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

// extensionOf is the lowercased extension of a file name, or "" when it has
// none.
//
// Spelled out rather than taken from the standard library, because the two
// libraries disagree. filepath.Ext(".gitignore") calls the extension
// ".gitignore" and Python's os.path.splitext calls it empty, so a dotfile
// would have matched a `type` on one connector and not the other. A leading
// dot is a name here, not an extension.
func extensionOf(name string) string {
	cut := strings.LastIndexByte(name, '.')
	if cut > 0 {
		return strings.ToLower(name[cut:])
	}
	return ""
}

// typeSuffixes is the extensions a `type` names, or a refusal saying what
// there is.
//
// Several may be named, comma-separated, because asking for Python and Go at
// once is an ordinary thing to want and the alternative is two searches.
func typeSuffixes(spec string) (map[string]bool, error) {
	out := map[string]bool{}
	for _, word := range strings.Split(spec, ",") {
		name := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(word)), ".")
		if name == "" {
			continue
		}
		if alias, ok := grepTypeAliases[name]; ok {
			name = alias
		}
		exts, ok := grepTypes[name]
		if !ok {
			return nil, refuse("bad_request",
				"I do not know a file type called %q. The ones I know are: %s. "+
					"For anything else, use glob.",
				strings.TrimSpace(word), grepTypeNames())
		}
		for _, ext := range exts {
			out[ext] = true
		}
	}
	if len(out) == 0 {
		return nil, refuse("bad_request",
			"type was given as %q, which names no file type.", spec)
	}
	return out, nil
}

// globToRegexp turns a shell-style pattern into a regular expression.
//
// filepath.Match cannot do this job. It has no ** at all, so every pattern
// naming a directory matched nothing and search_files reported complete:true,
// which says "I looked everywhere and there are none". That is the exact false
// claim the complete field was added to prevent, reachable through the field's
// own op, and with the pattern a model writes first.
//
//	**      any number of path segments, separators included
//	*       anything within one segment
//	?       one character, not a separator
//	[abc]   passed through as a character class
//
// Anchored at both ends, because a pattern is a description of the whole name
// and not a substring search. That is what search_content is for.
func globToRegexp(pattern string) (*regexp.Regexp, error) {
	var b strings.Builder
	b.WriteString("^")
	for i := 0; i < len(pattern); i++ {
		switch c := pattern[i]; c {
		case '*':
			if i+1 < len(pattern) && pattern[i+1] == '*' {
				i++
				// **/ swallows the separator too, so **/x matches a bare x at
				// the root as well as a/b/x. Without that, the commonest
				// pattern of all would miss the files directly in front of it.
				if i+1 < len(pattern) && pattern[i+1] == '/' {
					i++
					b.WriteString("(?:.*/)?")
				} else {
					b.WriteString(".*")
				}
			} else {
				b.WriteString("[^/]*")
			}
		case '?':
			b.WriteString("[^/]")
		case '[':
			end := strings.IndexByte(pattern[i:], ']')
			if end < 0 {
				return nil, fmt.Errorf("unclosed [ in %q", pattern)
			}
			b.WriteString(pattern[i : i+end+1])
			i += end
		default:
			b.WriteString(regexp.QuoteMeta(string(c)))
		}
	}
	b.WriteString("$")
	return regexp.Compile(b.String())
}

// globMatcher decides what a pattern is matched AGAINST, which is the whole
// substance of the fix.
//
// No separator in the pattern: the basename, because *.pdf has never meant
// anything else and every caller since 19 Sep has relied on it.
//
// A separator: the path relative to the search root, because docs/*.pdf can
// mean nothing else and matching it against a basename is how it silently
// found nothing.
type globMatcher struct {
	re      *regexp.Regexp
	onPath  bool
	pattern string
}

func newGlobMatcher(pattern string) (*globMatcher, error) {
	lowered := strings.ToLower(strings.ReplaceAll(pattern, "\\", "/"))
	re, err := globToRegexp(lowered)
	if err != nil {
		return nil, err
	}
	return &globMatcher{re: re, onPath: strings.Contains(lowered, "/"),
		pattern: pattern}, nil
}

// match takes both because which one is used is the matcher's decision, and
// the caller has them already from the walk.
func (g *globMatcher) match(rel, base string) bool {
	if g.onPath {
		return g.re.MatchString(strings.ToLower(filepath.ToSlash(rel)))
	}
	return g.re.MatchString(strings.ToLower(base))
}

// opSearchFiles finds files by name without a round trip per directory.
//
// It exists because walking cost sixty calls at a drive root, found nothing,
// and ended by asking the participant where to look.
//
// Every limit here sets Complete to false, and that field is the point. A
// search that stops quietly reports "not found" -- a stronger claim than "I
// stopped looking", and read as fact. That has already been said about a
// drive that did contain the file.
func opSearchFiles(root string, args json.RawMessage) (any, error) {
	var a searchFilesArgs
	json.Unmarshal(args, &a)
	start, err := resolve(root, a.Path)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(start)
	if err != nil {
		return nil, refuse("not_found", "There is no folder called %s there.",
			filepath.Base(start))
	}
	if !info.IsDir() {
		return nil, refuse("not_dir", "%s is a file, not a folder.", filepath.Base(start))
	}
	pattern := strings.TrimSpace(a.Pattern)
	if pattern == "" {
		return nil, refuse("bad_request", "Searching needs a pattern, like *.pdf.")
	}
	limit := a.Limit
	if limit <= 0 || limit > searchLimit {
		limit = searchLimit
	}
	// Case folding is inside the matcher now, for the same reason it was
	// here: the filesystems this runs on fold case anyway, see
	// caseInsensitiveFS.
	matcher, err := newGlobMatcher(pattern)
	if err != nil {
		return nil, refuse("bad_request",
			"%s is not a pattern I can read. Try something like *.pdf or "+
				"docs/**/*.txt.", pattern)
	}
	byTime := false
	switch strings.ToLower(strings.TrimSpace(a.Sort)) {
	case "", "path":
	case "modified":
		byTime = true
	default:
		return nil, refuse("bad_request",
			"sort is \"path\" or \"modified\", not %q.", a.Sort)
	}
	// Newest-first cannot stop at the limit, because the newest file may be
	// the last one the walk reaches. So the limit stops applying to the WALK
	// and starts applying to the answer, and the collection ceiling takes
	// over as the thing that bounds the work.
	collect := limit
	if byTime {
		collect = searchMaxCollect
	}

	deadline := time.Now().Add(searchSeconds * time.Second)
	result := searchFilesResult{Matches: []searchMatch{}, Complete: true}
	stack := []string{start}
	for len(stack) > 0 {
		if time.Now().After(deadline) || result.Scanned >= searchMaxScan {
			result.Complete = false
			break
		}
		here := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		entries, err := os.ReadDir(here)
		if err != nil {
			// A drive root is full of these. Skipping is right; pretending we
			// looked is not, so it costs Complete.
			result.Complete = false
			continue
		}
		for _, it := range entries {
			result.Scanned++
			full := filepath.Join(here, it.Name())
			// The same rule as reading: a link is neither followed nor
			// reported, so a search cannot name something read_file would
			// refuse, and cannot walk out of the shared folder.
			if skipEntry(full) {
				continue
			}
			// rel is what a pattern carrying a separator is matched
			// against, and it is relative to the folder the search STARTED
			// in rather than to the shared root: a caller who asked about
			// docs/ means docs/*.pdf to be about what is under docs.
			rel, relErr := filepath.Rel(start, full)
			if relErr != nil {
				rel = it.Name()
			}
			if matcher.match(rel, it.Name()) {
				if len(result.Matches) >= collect {
					result.Complete = false
					stack = nil
					break
				}
				m := searchMatch{Path: full, Kind: "file"}
				st, stErr := it.Info()
				if it.IsDir() {
					m.Kind = "dir"
				} else if stErr == nil {
					m.Size = st.Size()
				} else {
					m.Size = -1
				}
				if byTime && stErr == nil {
					m.Modified = st.ModTime().UTC().Format(time.RFC3339)
				}
				result.Matches = append(result.Matches, m)
			}
			if it.IsDir() {
				stack = append(stack, full)
			}
		}
	}
	if byTime {
		// Newest first, and the path settles a tie so the answer is the same
		// twice running. Two files written in the same second are common when
		// something generated them.
		sort.Slice(result.Matches, func(i, j int) bool {
			if result.Matches[i].Modified != result.Matches[j].Modified {
				return result.Matches[i].Modified > result.Matches[j].Modified
			}
			return strings.ToLower(result.Matches[i].Path) <
				strings.ToLower(result.Matches[j].Path)
		})
		if len(result.Matches) > limit {
			// Truncated AFTER sorting, so these really are the newest, and
			// complete is false because there were more.
			result.Matches = result.Matches[:limit]
			result.Complete = false
		}
	} else {
		sort.Slice(result.Matches, func(i, j int) bool {
			return strings.ToLower(result.Matches[i].Path) <
				strings.ToLower(result.Matches[j].Path)
		})
	}
	return result, nil
}

// -- search_content --------------------------------------------------------
//
// The op that answers Claude Code's Grep, and the last denied tool that had no
// answer at all. search_files could never take this job: it matches NAMES,
// which is what Glob asks for, and TOOLING.md records the correction.
//
// Why it belongs on this end rather than the server's: searching a hundred
// pages by pulling every chunk across the wire and scanning it in the model is
// not practical, and matching on the machine the files are already on is
// nearly free. That is the same argument read_file's chunking makes.
//
// Type narrows by language and Multiline lets a pattern cross a newline.
// Both are Grep's, and both had to land on this end: the filter because
// the walk happens here, and the matching because the files are here.

type searchContentArgs struct {
	Query string `json:"query"`
	Path  string `json:"path"`
	Glob  string `json:"glob"`
	// A language name, or several comma-separated, standing for the
	// extensions it is written in. Sugar over Glob, and ANDed with it.
	Type string `json:"type"`
	// Let a pattern cross a newline: the file is matched as one string
	// rather than line by line, and `.` crosses a line ending.
	Multiline     bool `json:"multiline"`
	Regex         bool `json:"regex"`
	CaseSensitive bool `json:"case_sensitive"`
	Limit         int  `json:"limit"`
	Context       int  `json:"context"`
	// "content" (default), "files" or "count". Grep calls the middle one
	// files_with_matches; the shorter name is used here because the answer
	// carries it back in a field of that name.
	OutputMode string `json:"output_mode"`
	// Skip this many results before the limit applies, which is how a caller
	// pages past what it has already seen. Grep spells it offset; that word
	// means a byte position everywhere else in this protocol.
	Skip int `json:"skip"`
}

type fileCount struct {
	Path  string `json:"path"`
	Count int    `json:"count"`
}

type contentMatch struct {
	Path string `json:"path"`
	Line int    `json:"line"`
	// Where a multiline match stopped. Absent when it did not span,
	// so a match inside one line answers as it always did.
	EndLine int      `json:"end_line,omitempty"`
	Text    string   `json:"text"`
	Before  []string `json:"before,omitempty"`
	After   []string `json:"after,omitempty"`
}

type searchContentResult struct {
	Matches []contentMatch `json:"matches"`
	// Only one of these is ever populated, decided by output_mode. Empty
	// rather than absent so a caller can tell "asked and found none" from
	// "did not ask".
	Files  []string    `json:"files,omitempty"`
	Counts []fileCount `json:"counts,omitempty"`
	Complete bool           `json:"complete"`
	Scanned  int            `json:"scanned"`
	Read     int            `json:"read"`
	Binary   int            `json:"binary"`
	TooBig   int            `json:"too_big"`
}

func opSearchContent(root string, args json.RawMessage) (any, error) {
	var a searchContentArgs
	json.Unmarshal(args, &a)
	start, err := resolve(root, a.Path)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(start)
	if err != nil {
		return nil, refuse("not_found", "There is no folder called %s there.",
			filepath.Base(start))
	}
	if !info.IsDir() {
		return nil, refuse("not_dir", "%s is a file, not a folder.",
			filepath.Base(start))
	}
	query := strings.TrimSpace(a.Query)
	if query == "" {
		return nil, refuse("bad_request",
			"Searching needs something to look for.")
	}

	// Case-INSENSITIVE unless asked otherwise, which is not what grep does and
	// is deliberate. These are somebody's documents rather than source, and a
	// missed match is reported as an absence. That is the failure this whole
	// family of ops keeps having to design around.
	flags := ""
	if !a.CaseSensitive {
		flags = "i"
	}
	if a.Multiline {
		// `s` so `.` crosses a newline, which is the whole point of the
		// parameter. `m` so `^` and `$` are line anchors, and that one is
		// about parity rather than taste: Python's `$` matches before a
		// file's trailing newline and this engine's does not, so without
		// `m` the two connectors disagree on the commonest pattern there
		// is.
		flags += "sm"
	}
	expr := query
	if !a.Regex {
		expr = regexp.QuoteMeta(query)
	}
	if flags != "" {
		expr = "(?" + flags + ")" + expr
	}
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, refuse("bad_request",
			"%s is not a regular expression I can read: %v", query, err)
	}

	var matcher *globMatcher
	if strings.TrimSpace(a.Glob) != "" {
		if matcher, err = newGlobMatcher(a.Glob); err != nil {
			return nil, refuse("bad_request",
				"%s is not a pattern I can read. Try something like *.md or "+
					"docs/**/*.txt.", a.Glob)
		}
	}

	// AND, not OR. type="py" with glob="tests/**" means the Python files
	// under tests, which is the only reading either one can have
	// alongside the other.
	var suffixes map[string]bool
	if strings.TrimSpace(a.Type) != "" {
		if suffixes, err = typeSuffixes(strings.TrimSpace(a.Type)); err != nil {
			return nil, err
		}
	}

	mode := strings.ToLower(strings.TrimSpace(a.OutputMode))
	switch mode {
	case "":
		mode = "content"
	case "content", "files", "count":
	default:
		return nil, refuse("bad_request",
			"output_mode is \"content\", \"files\" or \"count\", not %q.",
			a.OutputMode)
	}
	if a.Skip < 0 {
		return nil, refuse("bad_request",
			"skip cannot be negative, and %d is.", a.Skip)
	}

	limit := a.Limit
	if limit <= 0 || limit > grepLimit {
		limit = grepLimit
	}
	around := a.Context
	if around < 0 {
		around = 0
	} else if around > grepContext {
		around = grepContext
	}

	deadline := time.Now().Add(searchSeconds * time.Second)
	result := searchContentResult{Matches: []contentMatch{}, Complete: true}
	// Counted across the whole walk, not per file: skip pages through the
	// answer, and an answer that restarted its numbering in every folder
	// would page over the same results forever.
	kept := 0
	stack := []string{start}
	for len(stack) > 0 {
		if time.Now().After(deadline) || result.Scanned >= searchMaxScan {
			result.Complete = false
			break
		}
		here := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		entries, err := os.ReadDir(here)
		if err != nil {
			// A drive root is full of these. Skipping is right; pretending we
			// looked is not, so it costs Complete.
			result.Complete = false
			continue
		}
		for _, it := range entries {
			result.Scanned++
			full := filepath.Join(here, it.Name())
			// The same rule as reading: a link is neither followed nor
			// reported, so a search cannot name something read_file would
			// refuse, and cannot walk out of the shared folder.
			if skipEntry(full) {
				continue
			}
			if it.IsDir() {
				stack = append(stack, full)
				continue
			}
			if suffixes != nil && !suffixes[extensionOf(it.Name())] {
				continue
			}
			if matcher != nil {
				rel, relErr := filepath.Rel(start, full)
				if relErr != nil {
					rel = it.Name()
				}
				if !matcher.match(rel, it.Name()) {
					continue
				}
			}
			st, err := it.Info()
			if err != nil {
				result.Complete = false
				continue
			}
			// COUNTED, not silently passed over. A search that skipped the one
			// file the answer was in and said nothing would be the same lie as
			// stopping early and claiming completeness.
			ceiling := int64(grepMaxBytes)
			if a.Multiline {
				ceiling = grepMultilineMaxBytes
			}
			if st.Size() > ceiling {
				result.TooBig++
				continue
			}
			// The walk's own deadline was only checked once per DIRECTORY, so a
			// flat one of five thousand files did all its reading inside one
			// iteration and never looked at the clock. Multiline turns each of
			// those reads into a whole-file match, which is where that stops
			// being theoretical.
			if time.Now().After(deadline) {
				result.Complete = false
				stack = nil
				break
			}
			body, err := os.ReadFile(full)
			if err != nil {
				result.Complete = false
				continue
			}
			// Sniffed rather than decoded, for read_file's reason: the first
			// chunk of a PDF is mostly ASCII object headers, so a decode test
			// reads one happily and hands back line noise.
			head := body
			if len(head) > 16 {
				head = head[:16]
			}
			if magicName(head) != "" {
				result.Binary++
				continue
			}
			result.Read++

			text := strings.ReplaceAll(string(body), "\r\n", "\n")
			lines := strings.Split(text, "\n")
			inFile := 0
			// Two loops rather than one, because a multiline match is a
			// different object from a matching line: it has two line numbers,
			// it is found by position rather than by iterating lines, and it
			// has its own ceilings. Folding them together meant a generator
			// in one language and a slice in the other, which is exactly the
			// shape the two connectors drift apart in.
			if a.Multiline {
				// Every line's start offset, so a match's position becomes a
				// line number. Built once per file, not once per match.
				starts := make([]int, 0, len(lines))
				at := 0
				for _, one := range lines {
					starts = append(starts, at)
					at += len(one) + 1
				}
				// The engine's own enumeration, asked for one more than the
				// ceiling so that hitting it is distinguishable from ending on
				// it. Sliced by hand it would re-anchor `^` at the cut, and
				// Python's search-from-offset does not, which is a divergence
				// on the first pattern anyone writes.
				spans := re.FindAllStringIndex(text, grepMaxPerFile+1)
				for k, span := range spans {
					if k >= grepMaxPerFile {
						result.Complete = false
						break
					}
					atStart, atEnd := span[0], span[1]
					if atEnd == atStart {
						// A zero-width match is not something anyone searched
						// for, and the two engines need not agree on where one
						// lands. Dropping them is what lets this op promise the
						// same answer in both languages.
						continue
					}
					first := sort.SearchInts(starts, atStart+1) - 1
					last := sort.SearchInts(starts, atEnd) - 1
					// A match returns the whole LINES it touches rather than the
					// character range, so what comes back reads as the file
					// reads and Line still means what it means everywhere else.
					// EndLine says where it stopped.
					found := strings.Join(lines[first:last+1], "\n")
					if len(found) > grepMaxLine {
						// The same judgement as the per-line ceiling: past this
						// it is not prose. Dropping it costs Complete rather
						// than truncating, because half a match is not a match.
						result.Complete = false
						continue
					}
					inFile++
					if mode == "files" {
						break
					}
					if mode != "content" {
						continue
					}
					kept++
					if kept <= a.Skip {
						continue // paged past
					}
					if len(result.Matches) >= limit {
						result.Complete = false
						stack = nil
						break
					}
					m := contentMatch{Path: full, Line: first + 1,
						Text: strings.ToValidUTF8(found, "\uFFFD")}
					if last != first {
						// Only when it really spans more than one line, so a
						// multiline search that happened to match inside one
						// line answers exactly as it would without the flag.
						m.EndLine = last + 1
					}
					for i := first - around; i < first; i++ {
						if i >= 0 {
							m.Before = append(m.Before,
								strings.ToValidUTF8(lines[i], "\uFFFD"))
						}
					}
					for i := last + 1; i <= last+around && i < len(lines); i++ {
						m.After = append(m.After,
							strings.ToValidUTF8(lines[i], "\uFFFD"))
					}
					result.Matches = append(result.Matches, m)
				}
			} else {
				for n, line := range lines {
					if len(line) > grepMaxLine {
						// Not prose. Carrying it would blow the message ceiling
						// for one minified file.
						continue
					}
					if !re.MatchString(line) {
						continue
					}
					inFile++
					if mode == "files" {
						// One is enough to know the file matched, and reading on
						// costs the rest of the file for an answer already known.
						break
					}
					if mode != "content" {
						continue
					}
					kept++
					if kept <= a.Skip {
						continue // paged past
					}
					if len(result.Matches) >= limit {
						result.Complete = false
						stack = nil
						break
					}
					m := contentMatch{Path: full, Line: n + 1,
						Text: strings.ToValidUTF8(line, "\uFFFD")}
					for i := n - around; i < n; i++ {
						if i >= 0 {
							m.Before = append(m.Before,
								strings.ToValidUTF8(lines[i], "\uFFFD"))
						}
					}
					for i := n + 1; i <= n+around && i < len(lines); i++ {
						m.After = append(m.After,
							strings.ToValidUTF8(lines[i], "\uFFFD"))
					}
					result.Matches = append(result.Matches, m)
				}
			}
			if inFile > 0 && mode != "content" {
				kept++
				if kept > a.Skip {
					if len(result.Files)+len(result.Counts) >= limit {
						result.Complete = false
						stack = nil
					} else if mode == "files" {
						result.Files = append(result.Files, full)
					} else {
						result.Counts = append(result.Counts,
							fileCount{Path: full, Count: inFile})
					}
				}
			}
			if stack == nil {
				break
			}
		}
	}
	sort.Slice(result.Matches, func(i, j int) bool {
		if result.Matches[i].Path != result.Matches[j].Path {
			return result.Matches[i].Path < result.Matches[j].Path
		}
		return result.Matches[i].Line < result.Matches[j].Line
	})
	sort.Strings(result.Files)
	sort.Slice(result.Counts, func(i, j int) bool {
		return result.Counts[i].Path < result.Counts[j].Path
	})
	return result, nil
}

var ops = map[string]func(string, json.RawMessage) (any, error){
	"ping":           opPing,
	"list_directory": opListDirectory,
	"read_file":      opReadFile,
	"write_file":     opWriteFile,
	"get_file_info":  opGetFileInfo,
	"search_files":   opSearchFiles,
	"search_content": opSearchContent,
	"edit_file":      opEditFile,
}

// -- the session -----------------------------------------------------------

type connector struct {
	host, port, root, token string
	// "auto", "tcp" or "wss". See dial().
	wire string
	record                  string
	session                 string
	// id -> response, for the life of the session including across a resume.
	// PROTOCOL.md leaves the bound open; a session is one afternoon, so this
	// keeps everything, deliberately and on the record.
	done map[int][]byte
}

func (c *connector) send(t transport, v any) error {
	line, err := encode(v)
	if err != nil {
		return err
	}
	return t.sendRaw(line)
}

func (c *connector) runOnce() error {
	t, err := c.dial()
	if err != nil {
		return err
	}
	defer t.Close()
	if err := c.hello(t); err != nil {
		return err
	}
	return c.serve(t)
}

// wsPath is the single route Caddy forwards to the bridge. The bridge
// refuses a handshake for anything else, which test_ws_transport.py checks.
const wsPath = "/bridge"

// dial opens whichever transport this session uses.
//
// "auto" means wss on 443 and plain TCP anywhere else. That rule is what
// keeps the 25 recorded fixtures and every local test working unchanged
// while the shipped build talks WSS: the harness runs on 8790 and stays on
// newline-delimited JSON, exactly as PROTOCOL.md says it should.
//
// Only ever dials out, never listens. A listening socket on a participant's
// laptop raises Windows Firewall's "Allow access?" with an admin prompt, and
// that stops people dead.
func (c *connector) dial() (transport, error) {
	addr := net.JoinHostPort(c.host, c.port)

	// "ws" is the same framing without TLS. It exists because the bridge
	// itself serves plain WebSocket on loopback and Caddy is what terminates
	// TLS in front of it -- so this is the mode that lets framing be tested
	// end to end against bridge_mcp.py with no certificate anywhere.
	useWS := c.wire == "wss" || c.wire == "ws"
	useTLS := c.wire != "ws"
	if c.wire == "" || c.wire == "auto" {
		useWS = c.port == "443"
	}

	if !useWS {
		conn, err := net.DialTimeout("tcp", addr, 30*time.Second)
		if err != nil {
			return nil, err
		}
		logf("connected to %s, plain TCP", addr)
		return &plainTransport{conn: conn, r: bufio.NewReaderSize(conn, 64*1024)}, nil
	}

	var conn net.Conn
	var err error
	if useTLS {
		// ServerName set explicitly: the certificate is checked against the
		// hostname stamped into this binary, not against whatever a network
		// in the middle would prefer we accepted.
		conn, err = tls.DialWithDialer(
			&net.Dialer{Timeout: 30 * time.Second}, "tcp", addr,
			&tls.Config{ServerName: c.host})
	} else {
		conn, err = net.DialTimeout("tcp", addr, 30*time.Second)
	}
	if err != nil {
		return nil, err
	}
	r := bufio.NewReaderSize(conn, 64*1024)
	// Offered every time. A bridge that has not been updated yet simply does
	// not confirm it, and this end drops back to plain frames: the binary a
	// participant downloaded last week has to keep working, because there is
	// no way to make them download it again.
	deflate, err := wsClientHandshake(conn, r, c.host, wsPath, true)
	if err != nil {
		conn.Close()
		return nil, err
	}
	scheme := "wss"
	if !useTLS {
		scheme = "ws"
	}
	logf("connected to %s://%s%s", scheme, addr, wsPath)
	if deflate != nil {
		logf("  compression agreed: %s", wsExtDeflate)
	}
	return &wsTransport{conn: conn, r: r, d: deflate}, nil
}


func (c *connector) hello(t transport) error {
	err := c.send(t, map[string]any{
		"id": 0, "op": "hello",
		"args": helloArgs{
			Token:        c.token,
			Platform:     runtime.GOOS,
			Root:         c.root,
			Version:      version,
			Capabilities: capabilities,
			ResumeOf:     c.session,
		},
	})
	if err != nil {
		return err
	}
	for {
		line, err := t.readLine()
		if err != nil {
			return err
		}
		if len(line) == 0 {
			continue
		}
		var msg struct {
			ID     int   `json:"id"`
			OK     *bool `json:"ok"`
			Result struct {
				Session string `json:"session"`
				Resumed bool   `json:"resumed"`
			} `json:"result"`
			Error *opError `json:"error"`
		}
		if err := json.Unmarshal(line, &msg); err != nil || msg.ID != 0 {
			logf("expected a hello answer, got %.120s", line)
			continue
		}
		if msg.OK != nil && !*msg.OK {
			// A refused hello is permanent: the code is wrong, and retrying it
			// every two seconds forever would look like a network problem and
			// hide the one thing the participant can actually fix.
			why := "The session refused this connection."
			if msg.Error != nil && msg.Error.Message != "" {
				why = msg.Error.Message
			}
			fatal("%s", why)
		}
		if !msg.Result.Resumed {
			// A fresh session renumbers ids from 1, so cached answers from the
			// old one would be returned for unrelated requests.
			if len(c.done) > 0 {
				logf("resume refused; discarding %d cached answers", len(c.done))
			}
			c.done = map[int][]byte{}
		}
		c.session = msg.Result.Session
		logf("session=%s resumed=%v root=%s", c.session, msg.Result.Resumed, c.root)
		return nil
	}
}

func (c *connector) serve(t transport) error {
	for {
		line, err := t.readLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		if len(line) == 0 {
			continue
		}
		var req request
		if err := json.Unmarshal(line, &req); err != nil {
			logf("bridge sent a line that is not JSON: %v", err)
			continue
		}
		if req.Op == "" {
			logf("ignoring a message with no op: %.120s", line)
			continue
		}

		if cached, seen := c.done[req.ID]; seen {
			// The exactly-once rule. Do not re-execute; the write already
			// happened, and the bridge is only asking because the answer was
			// lost on the way back.
			logf("<- %d %s (already done; replaying the cached answer)", req.ID, req.Op)
			if err := t.sendRaw(cached); err != nil {
				return err
			}
			continue
		}

		logf("<- %d %s %.100s", req.ID, req.Op, req.Args)
		out, err := encode(c.execute(req))
		if err != nil {
			return err
		}
		c.done[req.ID] = out
		if err := t.sendRaw(out); err != nil {
			return err
		}
		c.appendRecord(req, out)
	}
}

func (c *connector) execute(req request) response {
	handler, ok := ops[req.Op]
	if !ok {
		return response{ID: req.ID, OK: false, Error: &opError{
			Code: "internal", Message: fmt.Sprintf("This connector has no %q.", req.Op)}}
	}
	result, err := handler(c.root, req.Args)
	if err != nil {
		var ref *refusal
		if errors.As(err, &ref) {
			logf("   refused: %s -- %s", ref.code, ref.message)
			return response{ID: req.ID, OK: false,
				Error: &opError{Code: ref.code, Message: ref.message}}
		}
		// Never let one bad request end the session: the participant would see
		// their files stop working with no explanation.
		logf("   %s failed: %v", req.Op, err)
		return response{ID: req.ID, OK: false, Error: &opError{
			Code:    "internal",
			Message: "Something went wrong reading that, which is a bug rather than anything you did."}}
	}
	return response{ID: req.ID, OK: true, Result: result}
}

func (c *connector) appendRecord(req request, out []byte) {
	if c.record == "" {
		return
	}
	f, err := os.OpenFile(c.record, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		// Said out loud. This was a bare `return`, so a record that could not
		// be opened was indistinguishable from a session in which nothing
		// happened -- and the file exists to be the account of what happened.
		logf("could not append to the record at %s (%v)", c.record, err)
		return
	}
	defer f.Close()
	var resp any
	json.Unmarshal(out, &resp)
	line, _ := encode(map[string]any{"request": req, "response": resp})
	f.Write(line)
}

// fatal says why, then waits, then leaves.
//
// A double-clicked console application closes its window the instant the
// process exits, so an error printed and returned from is an error nobody can
// read -- the participant sees a black flash and has nothing to tell anyone.
// Waiting for a keypress costs nothing when a human is there and is skipped
// when one is not, because a script's stdin is closed and Scanln returns at
// once.
func fatal(format string, a ...any) {
	fmt.Println()
	logf("%s", fmt.Sprintf(format, a...))
	fmt.Println()
	fmt.Print("Press Enter to close this window. ")
	var discard string
	fmt.Scanln(&discard)
	os.Exit(1)
}

// askForFolder asks which folder to share, offering Documents.
//
// Deliberately not defaulting silently: a participant should see the folder
// name before anything is shared from it, and "just press Enter" is a low
// enough bar that offering the default costs nobody anything.
// newline is the delimiter ReadString wants; named so the confirmation
// below reads without an escape sequence buried in it.
const newline = 10

func askForFolder() string {
	suggested := ""
	if home, err := os.UserHomeDir(); err == nil {
		docs := filepath.Join(home, "Documents")
		if info, err := os.Stat(docs); err == nil && info.IsDir() {
			suggested = docs
		} else {
			suggested = home
		}
	}

	fmt.Println()
	fmt.Println("  This lets Claude read files in one folder on this computer.")
	// "Nothing is uploaded" was here until 21 Sep 2026 and was not true. The
	// contents of whatever Claude reads DO leave this machine: that is what
	// read_file returns, and Claude cannot read a file it never receives.
	// The demo page carried the same sentence and was corrected on the 19th;
	// this copy was missed because that sweep looked at documents and not at
	// Go source. It is the more load-bearing of the two, because it is read
	// at the moment somebody chooses what to share.
	fmt.Println("  Whatever it reads is sent over the internet to answer you,")
	fmt.Println("  encrypted, the same as pasting text into a chat.")
	fmt.Println()
	fmt.Println("  Which folder should it be allowed to read?")
	if suggested != "" {
		fmt.Printf("     press Enter for   %s\n", suggested)
	}
	fmt.Println("     or type a folder and press Enter")
	fmt.Println()
	fmt.Print("  Folder: ")

	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	// Windows Explorer's "Copy as path" wraps the path in quotes, and a
	// participant pasting it should not have to know that.
	line = strings.Trim(strings.TrimSpace(line), "\"'")
	if line == "" {
		if suggested == "" {
			fatal("No folder given, so nothing was shared.")
		}
		return suggested
	}

	// Anything not absolute is resolved against the working directory, which
	// for a double-clicked program is wherever it was downloaded to. Typing
	// "C:" -- which Windows treats as drive-relative, not absolute -- would
	// silently share the Downloads folder and report success. Refusing is the
	// only safe answer: this program exists to keep files from leaving, and a
	// folder nobody chose is the one way it could.
	if !filepath.IsAbs(line) {
		fatal("%q is not a full folder path, so nothing was shared.\n%s", line,
			howToGiveAFolder())
	}
	if why := tooBroad(line); why != "" {
		fmt.Println()
		fmt.Printf("  %s is %s.\n", line, why)
		fmt.Println("  Everything inside it becomes readable, including saved")
		fmt.Println("  passwords, keys and anything downloaded.")
		fmt.Println()
		fmt.Print("  Share it anyway? Type yes to confirm: ")
		answer, _ := reader.ReadString(newline)
		if strings.ToLower(strings.TrimSpace(answer)) != "yes" {
			fatal("Nothing was shared. Run this again and give a narrower folder.")
		}
		logf("sharing %s after an explicit confirmation", line)
	}
	return line
}

// tooBroad reports why a folder is a bad thing to share, or "" if it is fine.
//
// A confirmation rather than a refusal: somebody may genuinely mean it, and
// this program cannot know. But the default prompt offers Documents, and the
// gap between pressing Enter and typing a drive letter should not be the gap
// between sharing a folder and sharing a machine.
//
// Written after a test session shared C:\ without comment, on a drive that
// held a code-signing private key.
func tooBroad(p string) string {
	clean := filepath.Clean(p)

	// A drive root or /: Dir of a root is itself.
	if filepath.Dir(clean) == clean {
		return "the whole drive"
	}
	// One level down from a root, which catches C:\Users, C:\Windows,
	// /home and /Users.
	parent := filepath.Dir(clean)
	if filepath.Dir(parent) == parent {
		base := strings.ToLower(filepath.Base(clean))
		for _, risky := range []string{"users", "windows", "home", "system32",
			"program files", "program files (x86)", "etc", "var"} {
			if base == risky {
				return "a system folder holding every account on this computer"
			}
		}
	}
	// The home directory itself: everything the person owns, which is the
	// case people actually reach for without meaning all of it.
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		if strings.EqualFold(filepath.Clean(home), clean) {
			return "your whole home folder"
		}
	}
	return ""
}

// howToGiveAFolder is the advice that follows a refused path, in the terms of
// the machine the participant is actually sitting at. Telling a Mac user about
// File Explorer and drive letters is worse than saying nothing: it reads as a
// program written for somebody else, which is exactly the impression this is
// meant to avoid.
func howToGiveAFolder() string {
	switch runtime.GOOS {
	case "darwin":
		// Dragging is better than any copy-path instruction: there is nothing
		// to find in a menu, and the path arrives correct and complete.
		return "  A full path starts with a slash, like /Users/you/Documents.\n" +
			"  Easiest way: drag the folder from Finder into this window, " +
			"then press Enter."
	case "windows":
		return "  A full path starts with a drive, like C:\\Users\\you\\Documents.\n" +
			"  In File Explorer: right-click the folder, Copy as path, " +
			"then paste it here."
	default:
		return "  A full path starts with a slash, like /home/you/Documents."
	}
}

// caseInsensitiveFS reports whether this platform's filesystem folds case by
// default. Both defaults are conventions rather than guarantees -- NTFS can be
// made case-sensitive per directory and APFS can be formatted that way -- but
// folding when the filesystem does not is harmless here, while not folding
// when it does refuses paths that are genuinely inside the share.
func caseInsensitiveFS() bool {
	return runtime.GOOS == "windows" || runtime.GOOS == "darwin"
}

// keyPattern finds the pairing key inside the downloaded file's name.
//
// Anchored on the key's own shape rather than on position, because a browser
// appends its own decoration when the same file is downloaded twice and every
// browser does it differently: Chrome and Edge give "name (1).exe", Firefox
// gives "name(1).exe" with no space, and Safari gives "name-1.exe". Splitting
// on the last hyphen reads Safari's key as "1". Searching for the shape leaves
// all of that outside the match.
//
// The alphabet omits 0, 1, I and O: the key is read off a screen and typed by
// someone who should not have to tell them apart.
var keyPattern = regexp.MustCompile(`(?i)nielsoln-([2-9A-HJ-NP-Z]{8})`)

// keyFromName returns the pairing key carried in this program's own filename,
// or "" if there is none. "nielsoln-bridge.exe" -- the development build --
// deliberately does not match.
func keyFromName(argv0 string) string {
	base := filepath.Base(argv0)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	if m := keyPattern.FindStringSubmatch(base); m != nil {
		return strings.ToUpper(m[1])
	}
	return ""
}

// askForKey is the fallback when the filename carries no key -- renamed on
// download, saved by something exotic, or copied by hand. The code is on the
// screen they downloaded it from, so asking costs one line.
func askForKey() string {
	fmt.Println()
	fmt.Println("  This copy does not carry a code in its filename.")
	fmt.Println("  There is one on the page you downloaded it from.")
	fmt.Println()
	fmt.Print("  Code: ")
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	return strings.ToUpper(strings.TrimSpace(line))
}

// keepWindowOpen is the last-resort handler AGENTS.base.md requires of every
// entry point, in the form Go needs.
//
// Every Python file here has one; this did not, which is the same defect the
// rule was written about. A panic prints its stack to stderr and exits, and a
// double-clicked console window closes the instant the process does -- so the
// participant sees a black flash, has nothing to report, and the one place the
// cause was written down is already gone.
//
// Caveat worth knowing: a panic in a *goroutine* cannot be recovered here.
// Everything in this program runs on the main goroutine for exactly that
// reason; keep it that way.
func keepWindowOpen() {
	r := recover()
	if r == nil {
		return
	}
	fmt.Println()
	logf("Something went wrong inside this program, which is a bug rather than")
	logf("anything you did. Nothing was shared.")
	fmt.Println()
	fmt.Printf("%v\n\n%s\n", r, debug.Stack())
	fmt.Print("Please show this to whoever is helping, then press Enter. ")
	var discard string
	fmt.Scanln(&discard)
	os.Exit(2)
}

func main() {
	defer keepWindowOpen()
	host := flag.String("host", defaultHost, "bridge address")
	port := flag.String("port", defaultPort, "bridge port")
	root := flag.String("root", "", "the only folder this session may read or write")
	token := flag.String("token", "", "pairing token, sent in hello")
	record := flag.String("record", "", "append every exchange here as JSON lines")
	once := flag.Bool("once", false, "do not reconnect; exit when the socket closes")
	wire := flag.String("transport", "auto",
		"auto, tcp, ws or wss. auto means wss on 443 and tcp elsewhere")
	flag.Parse()

	// The key rides in this program's own filename, because a participant has
	// no command line. --token still wins, for scripts and for the tests.
	if *token == "" {
		*token = keyFromName(os.Args[0])
		if *token != "" {
			logf("Code %s, from this file's name.", *token)
		}
	}

	// No --root means somebody double-clicked this, which is how every
	// participant will start it. Asking is better than failing: they have no
	// command line, and sharing the wrong folder is a privacy problem rather
	// than an inconvenience, so it should be a deliberate answer.
	if *root == "" {
		*root = askForFolder()
	}
	abs, err := filepath.Abs(*root)
	if err == nil {
		if real, err2 := filepath.EvalSymlinks(abs); err2 == nil {
			abs = real
		}
	}
	// --root skips the prompt, and with it the confirmation. Scripts and the
	// test suite pass it deliberately, so refusing here would break them for
	// no gain -- but saying nothing would mean the one path with no human in
	// it is also the one that shares a drive in silence.
	if why := tooBroad(abs); why != "" {
		logf("WARNING: %s is %s. Everything inside it is readable.", abs, why)
	}
	info, err := os.Stat(abs)
	if err != nil || !info.IsDir() {
		fatal("There is no folder called %s, so nothing was shared.", abs)
	}

	// --record wrote through os.OpenFile and never went near resolve(), so
	// `-record ../outside.jsonl` appended outside the shared folder: exactly
	// what the line printed below promises cannot happen. Found by an
	// independent review of commit 8b76568, not by us, which is the whole
	// argument for publishing the source rather than asking to be trusted.
	//
	// Checked HERE, at startup, rather than at the first write. A containment
	// rule enforced only when it is exercised is one nobody hears about until
	// something has already been written outside.
	if *record != "" {
		if _, err := resolve(abs, *record); err != nil {
			fatal("The record file %s is outside %s. Nothing was shared, "+
				"because a record written outside the shared folder would "+
				"break the only promise this program makes.", *record, abs)
		}
	}

	c := &connector{
		host: *host, port: *port, root: abs,
		token: *token, record: *record, wire: *wire, done: map[int][]byte{},
	}
	// The resolved folder, not what was typed. A participant should be able to
	// read back exactly what they have shared before anything is read from it.
	fmt.Println()
	logf("Sharing %s", abs)
	logf("Claude can read files here, and nowhere else. Close this window to stop.")

	backoff := time.Second
	for {
		if err := c.runOnce(); err != nil {
			logf("not connected (%v)", err)
		} else {
			backoff = time.Second
			logf("the bridge closed the connection")
		}
		if *once {
			return
		}
		// The laptop sleeps, the WiFi drops, the bridge restarts. Reconnecting
		// by ourselves is what keeps the session alive across all three.
		logf("reconnecting in %.0fs ...", backoff.Seconds())
		time.Sleep(backoff)
		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
}

// ---------------------------------------------------------------- wsframe --
//
// RFC 6455, client side only, translated from nielsoln_bridge/wsframe.py.
//
// That file is the reference and it is tested, including against the
// handshake example published in RFC 6455 section 1.3. This is a translation
// with an oracle rather than a fresh design, which is the entire basis for
// trusting hand-rolled framing here. If the two ever disagree, the Python is
// right until proven otherwise.
//
// Stdlib only, deliberately. The connector is the file a stranger in a
// library is asked to download and run, and "a thousand-odd lines of Go and
// nothing else" is a claim that has to stay true. A WebSocket library would
// have been three lines of go.mod and the end of that sentence.

const (
	wsGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

	// Prefixed wsOp, not op: this file already uses "op" for the six
	// protocol operations, and opPing is one of them. The compiler caught
	// the collision, which is the second near-identical pair of names to
	// surface in two days.
	// RFC 7692, offered bare: no window-bits or no-context-takeover
	// parameters, because both ends of this protocol are ours and a
	// negotiation with nothing to negotiate is only more surface.
	wsExtDeflate = "permessage-deflate"

	// Level 6 rather than 1 or 9. Measured in the bridge repo's
	// spike/flate_cost.go on 22 Sep 2026: level 6 beats level 1 on any link
	// below 120 Mbps, level 9 costs 27 ms more for 7,764 bytes, and level 6
	// stops paying for itself only above 519 Mbps of uplink.
	wsDeflateLevel = 6

	// The deflate window. The read side keeps exactly this much history,
	// which is what makes a dictionary equivalent to context takeover.
	wsDeflateWindow = 32 << 10

	wsOpCont   = 0x0
	wsOpText   = 0x1
	wsOpBinary = 0x2
	wsOpClose  = 0x8
	wsOpPing   = 0x9
	wsOpPong   = 0xA

	wsMaxControl = 125
)

// wsAcceptKey is the value the server must echo: base64(sha1(key + GUID)).
// The GUID is not a hash of anything, it is a constant published in the RFC,
// and getting one character wrong produces a handshake that fails only
// against conforming servers.
// wsSyncTail is what Flush appends and RFC 7692 section 7.2.1 says to strip.
var wsSyncTail = []byte{0x00, 0x00, 0xff, 0xff}

// wsInflateTail is that same tail followed by a final empty stored block.
// Without the final block the inflater sits waiting for input that will never
// arrive, because a message ends and a deflate stream does not.
var wsInflateTail = []byte{0x00, 0x00, 0xff, 0xff, 0x01, 0x00, 0x00, 0xff, 0xff}

// wsDeflate is one connection's compression state. See the note at the top of
// the file on why the two directions are built differently.
type wsDeflate struct {
	out bytes.Buffer
	w   *flate.Writer
	win []byte // the last wsDeflateWindow bytes we have inflated
}

func newWSDeflate() (*wsDeflate, error) {
	d := &wsDeflate{}
	w, err := flate.NewWriter(&d.out, wsDeflateLevel)
	if err != nil {
		return nil, fmt.Errorf("could not start the compressor: %w", err)
	}
	d.w = w
	return d, nil
}

// compress returns the payload of a compressed message: deflate, sync-flushed,
// with the four tail bytes removed.
//
// There is deliberately no "send it raw if it did not get smaller" path. The
// window persists, so feeding the compressor a message and then not sending it
// would leave the two ends holding different history, and everything after it
// would decode to plausible rubbish rather than an error. Skipping has to be
// decided before this is called. Nothing skips today: it is all JSON.
func (d *wsDeflate) compress(payload []byte) ([]byte, error) {
	d.out.Reset()
	if _, err := d.w.Write(payload); err != nil {
		return nil, err
	}
	if err := d.w.Flush(); err != nil {
		return nil, err
	}
	b := d.out.Bytes()
	if len(b) < len(wsSyncTail) || !bytes.HasSuffix(b, wsSyncTail) {
		return nil, errors.New("sync flush did not end as RFC 7692 requires")
	}
	return append([]byte(nil), b[:len(b)-len(wsSyncTail)]...), nil
}

// inflate expands one message, refusing to produce more than max bytes.
//
// The ceiling is on the OUTPUT, which is the whole point: a few hundred bytes
// on the wire can become gigabytes in memory, and a limit applied to what
// arrived would be no limit at all. This runs on a participant's laptop.
func (d *wsDeflate) inflate(payload []byte, max int) ([]byte, error) {
	src := io.MultiReader(bytes.NewReader(payload), bytes.NewReader(wsInflateTail))
	fr := flate.NewReaderDict(src, d.win)
	defer fr.Close()

	out, err := io.ReadAll(io.LimitReader(fr, int64(max)+1))
	if err != nil {
		return nil, fmt.Errorf("could not inflate a compressed message: %w", err)
	}
	if len(out) > max {
		return nil, fmt.Errorf("compressed message inflates past the %d ceiling", max)
	}
	d.win = wsKeepWindow(d.win, out)
	return out, nil
}

// wsKeepWindow keeps the tail of what we have inflated, so the next message
// can reference it exactly as the peer's persistent window expects.
func wsKeepWindow(win, add []byte) []byte {
	win = append(win, add...)
	if len(win) > wsDeflateWindow {
		win = append([]byte(nil), win[len(win)-wsDeflateWindow:]...)
	}
	return win
}

func wsAcceptKey(key string) string {
	sum := sha1.Sum([]byte(key + wsGUID))
	return base64.StdEncoding.EncodeToString(sum[:])
}

// wsClientHandshake completes the upgrade and VERIFIES the accept value.
//
// Verified rather than assumed, because a proxy or captive portal answering
// with a cheerful 200 is the exact failure this transport was chosen to
// survive. An unchecked handshake turns that into a connection that looks
// open and never delivers a byte, which is indistinguishable from a hang.
// wsClientHandshake upgrades the connection and returns the compression
// state, which is nil unless permessage-deflate was both offered and
// confirmed. An intermediary that strips the header therefore costs us
// compression rather than the connection, and there is an intermediary in the
// path: every production connection goes through Caddy.
func wsClientHandshake(conn net.Conn, r *bufio.Reader, host, path string, offerDeflate bool) (*wsDeflate, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return nil, fmt.Errorf("no randomness for the websocket key: %w", err)
	}
	key := base64.StdEncoding.EncodeToString(raw)

	lines := []string{
		"GET " + path + " HTTP/1.1",
		"Host: " + host,
		"Upgrade: websocket",
		"Connection: Upgrade",
		"Sec-WebSocket-Key: " + key,
		"Sec-WebSocket-Version: 13",
	}
	if offerDeflate {
		lines = append(lines, "Sec-WebSocket-Extensions: "+wsExtDeflate)
	}
	req := strings.Join(lines, "\r\n") + "\r\n\r\n"
	if _, err := conn.Write([]byte(req)); err != nil {
		return nil, err
	}

	resp, err := http.ReadResponse(r, nil)
	if err != nil {
		return nil, fmt.Errorf("no HTTP response to the upgrade: %w", err)
	}
	// The body is never read: on 101 there is none, and on anything else the
	// status is the whole story. Closing it is still correct.
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusSwitchingProtocols {
		return nil, fmt.Errorf("server refused the upgrade: %s", resp.Status)
	}
	if got := resp.Header.Get("Sec-WebSocket-Accept"); got != wsAcceptKey(key) {
		return nil, errors.New("Sec-WebSocket-Accept did not match the key we sent")
	}

	agreed := false
	for _, offer := range strings.Split(resp.Header.Get("Sec-WebSocket-Extensions"), ",") {
		if strings.TrimSpace(strings.Split(offer, ";")[0]) == wsExtDeflate {
			agreed = true
		}
	}
	if !agreed {
		return nil, nil
	}
	if !offerDeflate {
		// We would have no decompressor for what follows, and RSV1 would be
		// refused frame by frame a moment later. Saying so here names the
		// fault instead of leaving it looking like corruption.
		return nil, errors.New("server confirmed an extension we did not offer")
	}
	return newWSDeflate()
}

// wsSendFrame writes one frame, always masked. This end is a client and
// RFC 6455 section 5.1 makes masking a MUST; a conforming server closes the
// connection on an unmasked frame rather than warning about it. The bridge
// enforces exactly that, and test_ws_transport.py checks that it does.
// The d argument compresses the payload and sets RSV1. Control frames are
// never compressed, which RFC 7692 section 6.1 requires: a ping has to stay
// readable by anything that speaks RFC 6455, extension or not.
func wsSendFrame(w io.Writer, opcode byte, payload []byte, d *wsDeflate) error {
	if opcode >= wsOpClose && len(payload) > wsMaxControl {
		return fmt.Errorf("control frame of %d bytes", len(payload))
	}
	var rsv1 byte
	if d != nil && opcode < wsOpClose {
		var err error
		if payload, err = d.compress(payload); err != nil {
			return err
		}
		rsv1 = 0x40
	}
	var head []byte
	head = append(head, 0x80|rsv1|opcode) // FIN always set: we never fragment on send

	n := len(payload)
	switch {
	case n < 126:
		head = append(head, 0x80|byte(n))
	case n < 1<<16:
		head = append(head, 0x80|126)
		head = binary.BigEndian.AppendUint16(head, uint16(n))
	default:
		head = append(head, 0x80|127)
		head = binary.BigEndian.AppendUint64(head, uint64(n))
	}

	var mask [4]byte
	if _, err := rand.Read(mask[:]); err != nil {
		return fmt.Errorf("no randomness for the frame mask: %w", err)
	}
	head = append(head, mask[:]...)

	// Masked into a copy. Masking in place would corrupt a caller's buffer,
	// and the caller here hands us the encoded JSON line it may still log.
	body := make([]byte, n)
	for i := 0; i < n; i++ {
		body[i] = payload[i] ^ mask[i%4]
	}
	if _, err := w.Write(append(head, body...)); err != nil {
		return err
	}
	return nil
}

// wsReadFrame reads exactly one frame.
//
// io.ReadFull everywhere, never Read. A frame arriving across several TCP
// segments is the single most common hand-rolled WebSocket bug, and it hides
// until the network is slow -- which, for this tool, means it hides until the
// venue.
// allowRSV1 is set only when permessage-deflate was agreed. The payload comes
// back still compressed: RSV1 marks the first frame of a message rather than
// every frame of it, so inflating has to wait for reassembly.
func wsReadFrame(r io.Reader, allowRSV1 bool) (fin, rsv1 bool, opcode byte, payload []byte, err error) {
	var h [2]byte
	if _, err = io.ReadFull(r, h[:]); err != nil {
		return
	}
	fin = h[0]&0x80 != 0
	rsv1 = h[0]&0x40 != 0
	reserved := byte(0x70)
	if allowRSV1 {
		reserved = 0x30
	}
	if h[0]&reserved != 0 {
		// RSV1-3 are legal only once an extension has been negotiated. A bit
		// set for something we did not agree means the peer believes it is
		// talking to a different implementation, and every byte after this
		// would be misread rather than merely unexpected.
		err = errors.New("reserved bits set, but no extension was agreed")
		return
	}
	opcode = h[0] & 0x0F
	masked := h[1]&0x80 != 0
	if masked {
		// A server MUST NOT mask. One that does is not the bridge, or the
		// stream has been rewritten on the way through.
		err = errors.New("masked frame from a server")
		return
	}

	length := uint64(h[1] & 0x7F)
	switch length {
	case 126:
		var b [2]byte
		if _, err = io.ReadFull(r, b[:]); err != nil {
			return
		}
		length = uint64(binary.BigEndian.Uint16(b[:]))
	case 127:
		var b [8]byte
		if _, err = io.ReadFull(r, b[:]); err != nil {
			return
		}
		length = binary.BigEndian.Uint64(b[:])
	}

	if opcode >= wsOpClose {
		if !fin {
			err = errors.New("fragmented control frame")
			return
		}
		if length > wsMaxControl {
			err = fmt.Errorf("control frame of %d bytes", length)
			return
		}
	}
	if length > uint64(maxLine) {
		err = fmt.Errorf("frame of %d bytes exceeds the %d ceiling", length, maxLine)
		return
	}
	if length > 0 {
		payload = make([]byte, length)
		_, err = io.ReadFull(r, payload)
	}
	return
}

// ------------------------------------------------------------- transports --
//
// The bridge only ever does three things with its connection: get a line,
// send a line, close. Both transports offer exactly those, and nothing above
// this point knows which it has. bridge_mcp.py draws the same seam in the
// same place, and test_ws_transport.py drives one conversation through both
// and compares the JSON objects to prove the contract does not move.

type transport interface {
	// sendRaw takes an already-encoded line. The exactly-once cache holds
	// encoded bytes and has to replay them unchanged, so the seam is bytes
	// rather than values.
	sendRaw(raw []byte) error
	readLine() ([]byte, error)
	Close() error
}

// plainTransport is newline-delimited JSON on a bare socket: the harness
// transport, and what every fixture in fixtures/ was recorded over.
type plainTransport struct {
	conn net.Conn
	r    *bufio.Reader
}

func (t *plainTransport) sendRaw(raw []byte) error {
	_, err := t.conn.Write(raw)
	return err
}

func (t *plainTransport) readLine() ([]byte, error) {
	line, err := t.r.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	if len(line) > maxLine {
		logf("dropping a %d-byte line; ceiling is %d", len(line), maxLine)
		return nil, nil
	}
	return bytes.TrimSpace(line), nil
}

func (t *plainTransport) Close() error { return t.conn.Close() }

// wsTransport is the same messages under RFC 6455 framing: the production
// transport, because a library's wifi will pass 443 and will not pass 8790.
type wsTransport struct {
	conn net.Conn
	r    *bufio.Reader
	d    *wsDeflate // nil unless permessage-deflate was agreed
}

func (t *wsTransport) sendRaw(raw []byte) error {
	// encode() ends the line with a newline for the byte-stream
	// transport. A frame carries its own length, so that newline is
	// noise which would arrive inside the message the far end parses.
	return wsSendFrame(t.conn, wsOpText, bytes.TrimRight(raw, "\r\n"), t.d)
}

// readLine returns one complete message, reassembling fragments and
// answering control frames on the way past.
//
// Fragmentation is handled even though the bridge does not fragment: whether
// a message arrives in one frame is the SENDER's choice, and there is a
// reverse proxy in the path whose behaviour is not ours to assume.
func (t *wsTransport) readLine() ([]byte, error) {
	var (
		parts      [][]byte
		size       int
		started    bool
		isText     bool
		compressed bool
	)
	for {
		fin, rsv1, opcode, payload, err := wsReadFrame(t.r, t.d != nil)
		if err != nil {
			return nil, err
		}

		if opcode >= wsOpClose {
			// Control frames may arrive BETWEEN the fragments of a message,
			// which is why they are handled here rather than by the caller.
			switch opcode {
			case wsOpPing:
				if err := wsSendFrame(t.conn, wsOpPong, payload, nil); err != nil {
					return nil, err
				}
			case wsOpClose:
				// Echo the close and let the read fail next time round, which
				// the reconnect loop already treats as a dropped connection.
				_ = wsSendFrame(t.conn, wsOpClose, payload, nil)
				return nil, io.EOF
			}
			continue
		}

		if opcode == wsOpCont {
			if !started {
				return nil, errors.New("continuation frame with nothing to continue")
			}
			if rsv1 {
				// RFC 7692 section 6.1: RSV1 belongs on the first frame of a
				// message. On a continuation it means the peer is fragmenting
				// in a way we would reassemble wrongly.
				return nil, errors.New("RSV1 set on a continuation frame")
			}
		} else {
			if started {
				return nil, errors.New("a new message began before the last finished")
			}
			started = true
			isText = opcode == wsOpText
			compressed = rsv1
		}

		parts = append(parts, payload)
		size += len(payload)
		if size > maxLine {
			return nil, fmt.Errorf("message of at least %d bytes exceeds the %d ceiling",
				size, maxLine)
		}
		if !fin {
			continue
		}

		if !isText {
			// Nothing in this protocol is binary: file contents travel as
			// JSON strings, which PROTOCOL.md decided long before this.
			return nil, errors.New("binary message, but this protocol is text")
		}
		raw := bytes.Join(parts, nil)
		if compressed {
			inflated, ierr := t.d.inflate(raw, maxLine)
			if ierr != nil {
				return nil, ierr
			}
			raw = inflated
		}
		if !utf8.Valid(raw) {
			// Replaced rather than fatal, and loud -- the same rule the TCP
			// transport follows. Corruption must not end the session and must
			// not pass unremarked.
			logf("WARNING: invalid UTF-8 in a %d-byte message; replacing", len(raw))
			raw = []byte(strings.ToValidUTF8(string(raw), "\uFFFD"))
		}
		return bytes.TrimSpace(raw), nil
	}
}

func (t *wsTransport) Close() error { return t.conn.Close() }
