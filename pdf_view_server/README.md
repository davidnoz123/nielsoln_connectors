# pdf_view_server

Serve one folder over loopback, and tell the page which item to show.

```
go run github.com/davidnoz123/nielsoln_connectors/pdf_view_server@<sha> \
    -root . -host 127.0.0.1 -port 8790 -token K7M2XQ4P
```

A reviewer works down rows in a spreadsheet and wants the matching passage
shown, highlighted, in a browser. Making the spreadsheet open a URL gives **a
new tab per row**, which is unusable by about the fifth one, and nothing in the
hyperlink path can ask a browser to reuse a tab.

So the browser is never driven. A page served from here holds an SSE stream
open and the spreadsheet pushes an id; the page moves itself.

| | |
|---|---|
| **Any browser** | Safari, Chrome, Firefox. Nothing is automated, so there is no AppleScript, no CDP, and no *"Excel wants to control Chrome"* permission dialog |
| **One tab** | for the whole session |
| **`fetch()` works** | it is `http://`, not `file://`, so a page can run PDF.js on a real PDF instead of pre-rendered images — text stays selectable and zoom stays sharp |

## It knows nothing about PDFs

It serves bytes out of `-root` and relays ids. The viewer, the PDF and whatever
maps an id to a page and a rectangle are **data**, supplied by whoever runs it.

That is deliberate: this repo is public and the documents people review with it
are not.

## Flags

Same vocabulary as `remote_ai`, so there is one shape to learn.

| | |
|---|---|
| `-root` | the only folder this server may read. Default `.` |
| `-host` `-port` | default `127.0.0.1:8790`. Loopback on purpose: it does not trip the macOS incoming-connection firewall prompt |
| `-token` | required on every request. Empty disables the check and says so loudly |
| `-nav` | file under `-root` carrying the current id. Default `.nav` |
| `-poll` | how often to stat it. Default `100ms` |
| `-open` | open a browser once listening |

## Two ways in, because Excel has two

`POST /select?id=...` works from Windows VBA through `MSXML2.XMLHTTP`.

**Excel for Mac has no HTTP client in VBA.** MSXML does not exist there, so the
alternatives are `AppleScriptTask` shelling out to `curl` — which drags in a
script file in a sandboxed folder, and the Automation permission that goes with
it — or writing a file, which VBA does identically on both platforms with no
permissions at all.

So the nav file is watched as well:

```vba
Open navPath For Output As #1: Print #1, id: Close #1
```

Two lines, same on both platforms, so the event handler around it stays
identical.

## Endpoints

| | |
|---|---|
| `GET /events` | SSE. Sends the current id on connect, then each change. Keepalive comment every 25s |
| `POST /select` | `?id=` or the body |
| `GET /state` | `{"id": ..., "clients": n}` |
| `GET /*` | files under `-root`. A directory serves `index.html` or 404s — never a listing |

## What it refuses

- any path that resolves outside `-root`, **including through a symlink inside
  it** — the resolved path is re-tested, not the requested one
- a directory listing
- any request without the token

Containment uses a prefix test and folds case **only when the filesystem
measurably folds**, asked of the filesystem rather than inferred from
`runtime.GOOS`. That is carried over from `remote_ai`'s `case_test.go` rather
than rediscovered: `filepath.Rel` folds case on Windows whatever it is handed,
and folding unconditionally made a path *outside* a share read as inside it.
NTFS carries a per-directory case-sensitivity flag that `fsutil` sets without
elevation, so the OS name is not the answer.

## Build and test

```
go test ./...
GOOS=darwin GOARCH=arm64 go build .
```

Zero dependencies, no cgo, no build constraints — one `main.go` that compiles
for every target. The only place an operating system is named is opening a
browser, and those are different *arguments* to the same call rather than
different APIs.

Polling rather than `fsnotify`: native watching means inotify, FSEvents and
`ReadDirectoryChangesW` — a dependency and three platform files, to watch one
file.
