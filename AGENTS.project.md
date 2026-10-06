# AGENTS.project.md — nielsoln_connectors

Hand-written project rules. `AGENTS.md` is generated from this plus the central
standards and must never be edited directly.

---

## What this repo is

Small, separate connectors, each its own Go module, each fetchable and runnable
without a checkout:

```
go run github.com/davidnoz123/nielsoln_connectors/<connector>@<sha>
```

⚠️ **PROVEN, 7 Oct 2026, not assumed.** A module in a SUBDIRECTORY resolves at
a bare commit SHA: `go run .../chrome_sessions@bdc56ca` downloaded
`v0.0.0-20261006210611-bdc56ca708bb` and ran. The repo is publicly fetchable
with no credentials, so that line works on any machine, and a 7-hex
abbreviation is expanded to the full SHA by the toolchain itself.

Each `go.mod` carries a comment saying why its declared path must match where
the code lives: go refuses with "module declares its path as" **before it
downloads anything**, and that failure lands on the user's machine rather than
ours.

The design the connectors exist to serve is `shimp`, and its decisions live in
the `shimp` slate in this repo rather than in prose. `SHIMP.md` is a RENDER of
that slate, like `PLAN.md` is of slate_tools' queue. Do not hand-edit it.

---

## ⚠️ THIS IS A GO REPO AND THE CENTRAL STANDARDS ARE WRITTEN FOR PYTHON

There is no Go profile in `nielsoln_agent_standards`: the eleven profiles are
Python or domain-specific. So this repo declares `AGENTS.base` alone, and
several of its rules are inapplicable as written. **This section says which,
and what the Go equivalent is.** Nothing here overrides a rule that does apply.

| base rule | status here |
|---|---|
| the workspace Python interpreter | n/a. Go 1.27, `go` from PATH |
| `py_compile` after every edit | **replaced by `go vet ./...` and `go test ./...`**, and `gofmt -l .` must print nothing |
| `safe_local_imports` | n/a. Go has no import-time execution to guard |
| `_install_and_import` | n/a, and the stronger rule below replaces it |
| stdlib-only module-level imports | **strengthened: no third-party dependencies at all** |
| never run `pip install` | reads as: never `go get` a dependency into a connector without deciding it in the `shimp` slate first |

Everything else in the base applies unchanged and is not optional: the branch
model and virtual tags, no absolute paths in source or docs, no silent
failures, the prose rules, the worktree merging rule, and fakes verified
against the real thing.

---

## Rules

### Zero third-party dependencies

Every connector so far has an empty `require` block, and that is a property to
keep rather than an accident. A connector is meant to be auditable by reading
it, and a dependency tree is the thing that stops being possible.

Where the standard library is genuinely insufficient, raise it as a row on the
`shimp` slate before adding anything, so the decision is recorded next to the
reason.

### One file of generic code, platform specifics in their own files

```
sessions.go        the generic logic, no build tag
owner_windows.go   //go:build windows
owner_other.go     //go:build !windows, defining the SAME symbols
```

⚠️ **The catch-all file is not optional.** It exists so no platform is left
with a symbol undefined, and it should say what it cannot do rather than
returning a value that reads as an answer. `owner_other.go` returns 0 meaning
"cannot say", and `discoverPorts` there returns a REASON instead of an empty
list, because an empty list reads as "nothing is listening".

`pdf_view_server/raise_windows.go` and `raise_other.go` are the original
pattern and the one to copy.

### Platform APIs by syscall, not by shelling out

`raise_windows.go` records why: the first version shelled out to PowerShell's
`AppActivate` and cost **330ms per call**, spawning a whole runtime, for
something `user32` does in microseconds. `owner_windows.go` follows it, reading
the TCP table through `iphlpapi` rather than parsing `netstat`.

This also keeps the zero-dependency property, since the alternative is usually
a module.

⚠️ **And the struct layouts are the dangerous part.** A port in
`MIB_TCPROW_OWNER_PID` is network byte order in the low 16 bits: read
`dwLocalPort` as a plain `uint32` and 9222 comes back as 30987, which looks
like a real port number and so does not announce itself as a bug. Any new
syscall struct gets a test that pins a known value.

### Tests: a strict fake, and then the real thing

The base rule applies with full force here, and the connectors make it easy to
honour because most speak HTTP or a protocol rather than COM.

* **the fake is strict.** `sessions_test.go`'s fake Chrome answers the two
paths it models and 404s anything else, naming the path. A permissive fake
stays green while the code reaches for surface nobody modelled.
* **the real run is not optional**, and it has already earned its place: the
first real run of `chrome_sessions` found two defects the fake could not,
including a help string promising a filter that was never written.
* **say which question each run answered.** "Tests pass" after the fake alone
is a false statement about what was verified.

### Say what a tool cannot tell you

A connector's job is often to report on something it does not control, and the
failure that matters is the one that looks like success.

* `chrome_sessions` separates **liveness** from **authority**: a browser
answering on a port is not a browser that is signed in, and a signed-out
browser yields an empty catalog that reads as "nothing to capture". So `Tabs`
is documented as a hint and never as evidence.
* a refusal carries the reason verbatim, because "connection refused" and
"timed out" mean different things: nothing listening, against something
listening and wedged.

### Human output and machine output are different

`-json` exists so Python can call a connector today, before `shimp` exists, and
that is the allowed direction: **Python uses Go; Go does not use Python.**

Where a human view differs from the data, the difference belongs in the human
view only. `chrome_sessions` unescapes HTML entities in a printed title and
leaves Chrome's own string in the JSON, because a caller parsing it should get
the value rather than this program's idea of the value.

### Declared, never allocated

A port, a profile or a working directory that a connector needs is **declared**
by the caller. Nothing picks a free port.

The reason is `shimp`: a capability's identity has to be computable before its
process runs, so anything chosen at runtime cannot be part of it. A connector
that genuinely needs a dynamic port must report it after starting and never
take it as part of its identity.

---

## Build and check

```powershell
go vet ./...
go test ./...
gofmt -l .          # must print nothing
```

Per connector, from its own directory, because each is its own module.

---

## Slate

This repo has a slate, onboarded 7 Oct 2026. The `shimp` slate holds the design
decisions with a state each, so any one of them can be queried on the sheet.

`db/` is DERIVED and gitignored; `patches/` is the tracked record and must be
committed. The workbook and `transcripts/` are gitignored for the same reason
as everywhere else.

⚠️ **Write cells INSIDE an open turn.** A write made with no turn open is
absorbed into the next `begin_turn`'s snapshot, reads as standing still, and is
reverted by the sheet. Measured in slate_tools on 5 Oct: six cells lost,
including a 2,584 character design append.
