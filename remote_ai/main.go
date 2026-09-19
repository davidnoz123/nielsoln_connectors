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
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
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
	defaultPort  = 8790
	defaultLimit = 64 * 1024
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
}

type readFileResult struct {
	Path       string `json:"path"`
	Content    string `json:"content"`
	EOF        bool   `json:"eof"`
	BytesSent  int    `json:"bytes_sent"`
	TotalBytes int64  `json:"total_bytes"`
}

type writeFileArgs struct {
	Path       string `json:"path"`
	Content    string `json:"content"`
	CreateOnly bool   `json:"create_only"`
}

type writeFileResult struct {
	Path         string `json:"path"`
	BytesWritten int    `json:"bytes_written"`
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

	f, err := os.Open(path)
	if err != nil {
		return nil, refuse("not_allowed", "%s could not be opened.", filepath.Base(path))
	}
	defer f.Close()

	head := make([]byte, 16)
	nHead, _ := io.ReadFull(f, head)
	head = head[:nHead]

	raw := make([]byte, limit)
	n, err := f.ReadAt(raw, offset)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, refuse("internal", "%s could not be read.", filepath.Base(path))
	}
	raw = raw[:n]
	atEOF := offset+int64(n) >= total

	// Sniff before decoding, not after. A failed decode is not a reliable test
	// for binary: the first chunk of a real PDF is mostly ASCII object headers,
	// and NUL is perfectly valid UTF-8 -- so a decode-only check reads a PDF
	// happily and hands Claude line noise it will try to interpret.
	if name := magicName(head); name != "" {
		return nil, refuse("not_text", "%s looks like %s, so it cannot be read as text.",
			filepath.Base(path), name)
	}
	if bytes.IndexByte(raw, 0) >= 0 {
		// The same sentinel git uses. No real text file carries a NUL.
		return nil, refuse("not_text", "%s is not a text file, so it cannot be read as text.",
			filepath.Base(path))
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

	return readFileResult{
		Path:       path,
		Content:    text,
		EOF:        offset+int64(used) >= total,
		BytesSent:  used,
		TotalBytes: total,
	}, nil
}

func opWriteFile(root string, args json.RawMessage) (any, error) {
	var a writeFileArgs
	json.Unmarshal(args, &a)
	path, err := resolve(root, a.Path)
	if err != nil {
		return nil, err
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
	// Write beside the target and rename. An interrupted write then leaves the
	// original intact rather than a half-written file.
	tmp := path + ".part"
	f, err := os.Create(tmp)
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
	return writeFileResult{Path: path, BytesWritten: len(data)}, nil
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
)

type searchFilesArgs struct {
	Pattern string `json:"pattern"`
	Path    string `json:"path"`
	Limit   int    `json:"limit"`
}

type searchMatch struct {
	Path string `json:"path"`
	Kind string `json:"kind"`
	Size int64  `json:"size"`
}

type searchFilesResult struct {
	Matches  []searchMatch `json:"matches"`
	Complete bool          `json:"complete"`
	Scanned  int           `json:"scanned"`
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
	// Matched lowercased rather than with a case-insensitive matcher, because
	// filepath.Match has none and the filesystems this runs on fold case
	// anyway -- see caseInsensitiveFS.
	lowered := strings.ToLower(pattern)

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
			ok, _ := filepath.Match(lowered, strings.ToLower(it.Name()))
			if ok {
				if len(result.Matches) >= limit {
					result.Complete = false
					stack = nil
					break
				}
				m := searchMatch{Path: full, Kind: "file"}
				if it.IsDir() {
					m.Kind = "dir"
				} else if st, err := it.Info(); err == nil {
					m.Size = st.Size()
				} else {
					m.Size = -1
				}
				result.Matches = append(result.Matches, m)
			}
			if it.IsDir() {
				stack = append(stack, full)
			}
		}
	}
	sort.Slice(result.Matches, func(i, j int) bool {
		return strings.ToLower(result.Matches[i].Path) <
			strings.ToLower(result.Matches[j].Path)
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
}

// -- the session -----------------------------------------------------------

type connector struct {
	host, port, root, token string
	record                  string
	session                 string
	// id -> response, for the life of the session including across a resume.
	// PROTOCOL.md leaves the bound open; a session is one afternoon, so this
	// keeps everything, deliberately and on the record.
	done map[int][]byte
}

func (c *connector) send(w io.Writer, v any) error {
	line, err := encode(v)
	if err != nil {
		return err
	}
	_, err = w.Write(line)
	return err
}

func (c *connector) runOnce() error {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(c.host, c.port), 30*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
	logf("connected to %s:%s", c.host, c.port)

	r := bufio.NewReaderSize(conn, 64*1024)
	if err := c.hello(conn, r); err != nil {
		return err
	}
	return c.serve(conn, r)
}

func (c *connector) readLine(r *bufio.Reader) ([]byte, error) {
	line, err := r.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	if len(line) > maxLine {
		logf("dropping a %d-byte line; ceiling is %d", len(line), maxLine)
		return nil, nil
	}
	return bytes.TrimSpace(line), nil
}

func (c *connector) hello(conn net.Conn, r *bufio.Reader) error {
	err := c.send(conn, map[string]any{
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
		line, err := c.readLine(r)
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

func (c *connector) serve(conn net.Conn, r *bufio.Reader) error {
	for {
		line, err := c.readLine(r)
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
			if _, err := conn.Write(cached); err != nil {
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
		if _, err := conn.Write(out); err != nil {
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
	fmt.Println("  This lets Claude read files on this computer, and nothing else.")
	fmt.Println("  Nothing is uploaded: Claude asks, and this program answers.")
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
	return line
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
	port := flag.Int("port", defaultPort, "bridge port")
	root := flag.String("root", "", "the only folder this session may read or write")
	token := flag.String("token", "", "pairing token, sent in hello")
	record := flag.String("record", "", "append every exchange here as JSON lines")
	once := flag.Bool("once", false, "do not reconnect; exit when the socket closes")
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
	info, err := os.Stat(abs)
	if err != nil || !info.IsDir() {
		fatal("There is no folder called %s, so nothing was shared.", abs)
	}

	c := &connector{
		host: *host, port: fmt.Sprint(*port), root: abs,
		token: *token, record: *record, done: map[int][]byte{},
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
