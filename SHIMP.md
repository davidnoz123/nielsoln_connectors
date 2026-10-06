| item id | item | state | waiting on |
|---|---|---|---|
| s1 | shimp is a connector folder like any other | **DECIDED** 7 Oct 2026 | nothing |
| s2 | Go only: Go does not use Python, Python uses Go | **DECIDED** 7 Oct 2026 | nothing, for a month or more |
| s3 | `runner` must be in the FIRST descriptor ever written | **DECIDED** 7 Oct 2026 | nothing, and it blocks the first descriptor |
| s4 | Declared ports, never allocated | **DECIDED** | nothing |
| s5 | Accept an abbreviated SHA, store the full one | **DECIDED** | nothing |
| s6 | The cold fallback, which SHIMP/1 does not specify | **DECIDED**, and it is a contribution back to the design | nothing |
| s7 | install places a permanent binary, never registers `go run` | **DECIDED** 7 Oct 2026 | nothing |
| s8 | The Office trusted-protocol key, per scheme | **DECIDED** | nothing |
| s9 | Is the audit two lines or three? | **OPEN**, and it is the only thing blocking a first descriptor | a decision |
| s10 | Where a long action reports | **OPEN** | a decision |
| s11 | shimp as a library, not only a URI | **OPEN** | a decision |
| s12 | chrome_sessions: liveness built, authority open | PART BUILT | the CDP extraction, s13 |
| s13 | Extract the CDP client, do not copy it | RAISED | a decision on where a shared package lives |
| s14 | patchbucket: transport and ordering, never merging | SPECIFIED | nothing, and it is unblocked |
| s15 | chatgpt_capture needs a SIGNED-IN Chrome, and must refuse without one | SPECIFIED | s12's authority probe |
| s16 | Excel is a session-protocol problem, not a prohibition | **OPEN** | a decision on whether the Go side speaks the session protocol |
| s17 | What is NOT in scope for SHIMP/1 | **DECIDED** | nothing |
| s18 | Where this design came from | **REFERENCE** | nothing |
| s19 | The URI grammar, and what identity IS | **DECIDED** in the source (s18), NOT implemented | nothing. These are the spec to build against. |
| s20 | The keeper, the IPC, and the lifecycle | **DECIDED** in the source (s18), NOT implemented | nothing. These are the spec to build against. |
### s1. shimp is a connector folder like any other

No repo reorganisation, no separate trust-root repository. `shimp/` sits beside
`pdf_view_server`, `remote_ai`, `remote_chrome_cdp` and `chrome_sessions`, with
its own `go.mod` declaring
`github.com/davidnoz123/nielsoln_connectors/shimp`.

#### Proven, not assumed

The whole install story rested on an assumption nobody had tested: that
`go run` can fetch a module living in a SUBDIRECTORY at a bare commit SHA.
Measured 7 Oct 2026:

```
go run github.com/davidnoz123/nielsoln_connectors/chrome_sessions@bdc56ca
  go: downloading .../chrome_sessions v0.0.0-20261006210611-bdc56ca708bb
  9222   alive pid 10272        Chrome/154.0.8037.93
```

Four things in one command: a subdirectory module resolves at a bare SHA; the
repo is publicly fetchable with no credentials, so the install line works on
any machine; a 7-hex abbreviation was accepted and expanded to the full SHA by
the toolchain itself, which validates "accept abbreviated, store full"; and the
parent module path is consulted too, which is harmless and worth knowing.

#### An earlier argument of mine, withdrawn

I argued that co-locating shimp with its targets was a risk, because one repo's
compromise would change both what `shimp:` means and what it runs.
⚠️ **The pinning defuses that.** The handler is installed from a specific SHA
and replacement is refused without `--replace`; capabilities pin target SHAs
too. A hostile commit landing in the repo changes NO installed handler and NO
existing link. The real trust boundary is `--replace`, not the folder layout.

### s2. Go only: Go does not use Python, Python uses Go

#### What it buys

The audit story becomes achievable. `go run module@sha` pins the resolved
module graph through `go.sum` and the checksum database. A Python tree at a SHA
pins the source and nothing else: not the interpreter, not pywin32, not one
package. So a Go-only target really can be described by its SHA.

#### What it costs

See s16. A COM reference cannot cross a process boundary, so the Excel daemon
is the one place this bites.

### s3. `runner` must be in the FIRST descriptor ever written

⚠️ Adding the field later changes every descriptor's canonical bytes, which
changes every capability hash, which changes every capability id, **which
breaks every link ever written**.

So: `"runner": "go"`, exactly one legal value today, unknown values fail
closed, which is the rule already chosen for `version`.

This is the cheapest decision on the list and the most expensive to defer,
because s2 is stated as likely to relax.

### s4. Declared ports, never allocated

A capability's identity must be computable before its process runs, so "pick a
free port" cannot be part of it.

A connector that genuinely needs a dynamic port must REPORT it after starting,
over the keeper's IPC, and never take it as part of identity.

Consequence, and a happy one: a port manager is unnecessary. Put the profile
and the port in a capability's argv and the capability hash IS that browser's
identity, so SHIMP's singleton logic does the whole job with no second
registry.

### s5. Accept an abbreviated SHA, store the full one

The pedantry then appears once, when a capability is created, and never in a
link, because nobody types a `shimp://` URI and a tool emits it.

#### The numbers that decide it

Matching a prefix by brute force, which is the adversarial case:

| prefix | brute force | accidental, within a repo |
|---|---|---|
| 7 hex | ~2.7e8, minutes | ~16k objects |
| 12 hex | ~3e14, expensive | ~17M objects |

Accidental collision is a non-issue for small repos at 7 hex, which is what
makes abbreviation tempting. The asymmetry that decides it: the thing being
authorised is **code execution on click**, so the adversarial number is the one
that matters, and git grows its own abbreviations with repo size for a weaker
reason than that.

### s6. The cold fallback, which SHIMP/1 does not specify

If the singleton cannot be reached, **the clicking process does the work
itself**.

That is what makes the singleton an optimisation and never a dependency, so
there is no window in which a link is broken, only one in which it is slower.
It also makes SHIMP's startup-handshake race mostly moot: the readiness
question stops being urgent when not being ready is survivable.

Measured in the Python original: a click is 828ms cold and 526ms warm, so the
singleton saves 301ms or 36%. Of the warm figure about 100ms is imports and the
rest is interpreter start, which no daemon can remove.

### s7. install places a permanent binary, never registers `go run`

Measured 7 Oct 2026, same program, module already cached:

```
go run : 2609 ms   (first warm run, still linking)
go run :  649 ms
go run :  653 ms
binary :   68 ms
binary :   68 ms
```

So `go run github.com/.../shimp@SHA install` is the PROVENANCE, and what it
leaves behind is a permanent executable the registry points at. One-line
install, immutable source, and 68ms clicks.

⚠️ The binary is copied from `os.Executable()`, which under `go run` is the
toolchain's temporary build output. It is self-contained, so copying it is
enough, and that is why none of this needs a repo layout change.

### s8. The Office trusted-protocol key, per scheme

```
HKCU\Software\Policies\Microsoft\Office\16.0\Common\Security
     \Trusted Protocols\All Applications\shimp:
```

⚠️ **The trailing colon is part of the key name.** Microsoft's own example is
"to disable the display of a security warning for the `Notes:` protocol, enter
`Notes:`". A key named `shimp` looks right in regedit and does nothing.

Without it, every click from Excel or Word raises a modal, and **that modal
blocks COM invisibly**: on 5 Oct an Excel instance held a workbook for two and
a half hours behind one, and the only symptom was a read-only refusal naming no
pid.

`install` claims it, `verify` reports it, `uninstall` withdraws only its own
subkey because `All Applications` is shared.

Reported rather than promised: a second dialog comes from HLINK.dll and this
key does not suppress it, by Microsoft's own account, so `verify` says the key
is present and not that clicks are silent.

### s9. Is the audit two lines or three?

Two choices were made that cannot both hold.

* the URI is `shimp://<capability-id>/<action>`, keeping startup details out of
the clickable link
* the audit payload is two lines, the install command and the URI, because a
reader "can derive the exact target and action from the `shimp:` URI"

⚠️ **A local sha256 of a local descriptor cannot be resolved by a stranger.**
Not by a colleague, not by a chat window. So either the URI carries the target
SHA and is self-describing, or the paste is three lines.

Three looks right: it keeps links short, keeps the descriptor as the
authoritative object, and stops the audit resting on something only one machine
can resolve. But it is a change to the headline claim, so it is a decision
rather than a detail.

### s10. Where a long action reports

The action model is request/response, which fits a navigation at half a second
and is wrong for a job that takes minutes.

A click either blocks invisibly or returns having said nothing. Today the
console shows `[full_update] step 1/3 ...`; a hyperlink has no console.

Candidates: a `status` action to poll, a log path returned on start, or a
completion notification. `pdf_view_server/raise_windows.go` already has the
window-raising primitive for the last.

### s11. shimp as a library, not only a URI

chatgpt_capture needs a signed-in Chrome before it sweeps. Either it ensures
that itself, which means calling shimp from Go, or a human clicks two links in
the right order, which is fragile.

So the same logic needs a Go API and not only a scheme registration. Small, and
it changes what the package exports, so it is worth settling before the
handler is written.

### s12. chrome_sessions: liveness built, authority open

⚠️ **LIVENESS AND AUTHORITY ARE DIFFERENT QUESTIONS**, and only the first is
answered.

| | |
|---|---|
| is Chrome up on this port | generic, and `/json/version` answers it over plain HTTP |
| is it signed in, as which account | site-specific, needs a websocket and `Runtime.evaluate` |

The second matters because it is the failure that looks like success: a
signed-out browser yields an empty ChatGPT catalog, which reads as "nothing to
capture". `Tabs` is a hint about what a browser has open and **never** a claim
about who it is logged in as.

#### Why not extend remote_chrome_cdp

The two take opposite stances. `launch.go` says: "A DEDICATED PROFILE, ALWAYS
... The cost is that this Chrome is signed out ... the honest answer there is
to point `-chrome-port` at a browser you have signed in yourself." That is
right for sweeping public pages and a blocker for anything behind a login. Two
opposite profile policies cannot live in one tool without one being a special
case, and the special case would be the credential-bearing one, which is the
one that must never silently degrade.

#### Two defects the first real run found, both mine

* `-discover` probed EVERY listening port, including 135, 139, 443 and 445,
which is RPC and SMB; one answered by forcibly closing the connection. And the
flag's help promised a filter that was never written, "whose owner looks like a
browser". Documentation claiming behaviour the code lacks, in the first hour.
* titles came back HTML-escaped, `Edit &quot;Home&quot; with Elementor`.
Decoded in the human view only; the JSON keeps Chrome's own string.

### s13. Extract the CDP client, do not copy it

The authority probe in s12 needs a websocket and `Runtime.evaluate`, which
`remote_chrome_cdp/cdp.go` and `ws.go` already implement in 42KB.

⚠️ Copying them would put two implementations of subtle machinery in one repo.
Measured cost of that pattern, 5 Oct 2026: three separate incidents in one
evening, all in the one area where two tools both reasoned about Excel
sessions.

Open: where a shared package lives, given every connector is its own module
today. Not urgent until the second consumer exists, which is what s12's
authority probe will make it.

### s14. patchbucket: transport and ordering, never merging

**patchbucket owns:** the rclone wrapper and its timeout, which exists for a
measured hazard (a wedged rclone that finishes transferring and never exits),
per-machine directories under a remote prefix, `latest.txt` as the head
pointer, sequence numbering, per-peer progress, and download-and-apply in order
while skipping your own machine.

**The owning tool keeps:** `export_rows_since`, `apply_row_patch` and the
`schema_version` gate. `apply_row_patch` hardcodes `conversations`, `messages`,
`scraped_sessions` and `blobs`, their per-table conflict rules and the FTS5
update. That is merge semantics, not transport.

⚠️ **TWO WATERMARKS THAT LOOK ALIKE AND MUST NOT BE CONFLATED.** patchbucket
owns "how far have I consumed peer X". The owning tool owns "how far have I
exported my own rows", and that one is read inside a DEFERRED transaction
together with the export so that no rows slip through the gap. Move it across a
process boundary and the atomicity is gone, silently.

Worth more than its first caller: slate_tools and inter_repo_minutes both write
patches per turn. Though for anything already committed, git IS the transport
and patchbucket would be redundant there.

The name should disclaim merging. The moment it looks like it merges, somebody
will give it a schema.

### s15. chatgpt_capture needs a SIGNED-IN Chrome, and must refuse without one

`op_full_sync` is already incremental: `clean_stop_pages=3` stops after three
consecutive clean catalog pages. Its hard dependency is authority, not
liveness.

⚠️ **AND THE FAILURE LOOKS EXACTLY LIKE SUCCESS.** No token means an empty
catalog, which reads as "nothing to capture". So a capability that "ensures
Chrome" would ensure the WRONG Chrome, because the one `remote_chrome_cdp`
launches is signed out by design.

Both halves are needed: the Chrome capability for this purpose attaches and
never launches, and the capturer refuses loudly when `/api/auth/session` yields
no token rather than sweeping an empty catalog.

`full_update` then reduces to one orchestrating action: pull, sweep, push. One
URI is one capability and one action, so the sequencing lives in code rather
than in the link.

### s16. Excel is a session-protocol problem, not a prohibition

#### An earlier framing of mine, withdrawn

I proposed an attach-only Go reader that would REFUSE rather than open a
workbook, on the grounds that opening creates the rival-instance condition.
That was over-constrained. The read-only incidents of 5 and 6 October were not
caused by opening a workbook; they were caused by opening it in an instance
that was not the one already holding it. `ensure_workbook_open` exists to do
that correctly, and slatex's own comment calls it "AND THEN THROUGH THE FRONT
DOOR, which is the whole point of this line".

#### The real constraint

⚠️ **A COM reference cannot cross a process boundary**, so whichever process
owns the apartment does ALL the Excel work in that process. And the session
discipline is a PROTOCOL, held in a small JSON handle file, whose identity test
is two signals: the pid and the caption nonce, neither sufficient alone.

A Go implementation that opens Excel without speaking that protocol will be
seen by the Python side as a stranger holding the file, and refused with "held
by an Excel that is not this session's". The two tools would then deadlock each
other politely.

So the options are: port Excel to Go wholesale, which is publish, pull,
formats, character runs and the silos across 15,541 and 5,187 lines; or have
the Go side read and write the SAME session handle and honour the same rule.

If the second, the house rule names the way to stop two implementations
drifting: **one contract suite, two drivers**, the same assertions run against
the Python implementation and the Go one.

### s17. What is NOT in scope for SHIMP/1

Sandboxing, filesystem and network confinement, auto-update, and any general
service manager.

SHIMP guarantees one thing: that this exact process specification has at most
one managed live instance, and that this action was delivered to it.

The temptation is `policy=filesystem:cwd` meaning "this is enforced". In
SHIMP/1 a policy field is **declarative metadata only**: it says what a target
claims to need, and the audit can then check whether the target or the OS
actually enforces it. Otherwise the tiny launcher becomes a security container,
which is a much larger project.

Also noted rather than scoped: `HYPERLINK()` caps `link_location` at 255
characters while the Hyperlink object does not, which is an argument for the
capability-id form over inlining target and args.

### s18. Where this design came from

⚠️ **The whole of SHIMP was designed in a ChatGPT session, and nothing in this
repo recorded which one** until this row. That is the provenance problem the
other slates exist to prevent, repeated here on day one.

#### The source

| | |
|---|---|
| title | Custom URL Scheme |
| id | `6ac4ee15-845c-83ec-8352-47f52b596f32` |
| when | 7 Oct 2026, 33 messages, 61,518 characters |
| model | gpt-5-6-thinking |
| where | `chatgpt_tools/chatgpt_1davidnoz_at_googlemail_dot_com.db` |

Read it whole, which is the only way worth reading it:

```python
import sqlite3
c = sqlite3.connect("file:<db>?mode=ro", uri=True)
rows = c.execute("SELECT role, content FROM messages "
                 "WHERE conversation_id=? ORDER BY turn_index, id",
                 ("6ac4ee15-845c-83ec-8352-47f52b596f32",))
```

The messages worth going back to: [5] the launcher model and why the URI must
not encode a shell command, [11] the capability descriptor, [13] its own
assessment of scope at "75% architectural, 45 to 50% protocol", [15] ideas for
each open point, [23] the decisions-to-freeze table, [25] where it leans on the
four real choices, and [32] the precedents search, which found close analogues
for every individual piece and none for the combination.

#### The accidental first implementation

`slate_tools/slate_open.py`, built 5 Oct before this idea had a name, is a
single-purpose SHIMP: scheme registration, a singleton keyed by identity, a
nonce-authenticated local socket, the URI remainder as an action, and
request/response replies. Worth reading before writing the handler, because
every measurement in s6 and s7 came out of it and so did five defects.

The rows that retired it: slate_tools `r178` records the supersession and what
is carried over, `r147` the daemon itself, and `r148` the cross-platform work
that became moot.

#### What is NOT written down anywhere else

The reasoning in this slate is the record. There is no second document, and
`SHIMP.md` is a render of these rows rather than a source. A decision changed
in the file is lost at the next turn.

#### The transcript, which already exists

The session that produced all of this was captured automatically at the first
turn boundary:

    transcripts/000001.jsonl.z    9,267,237 bytes, 14,264 lines, turn 1

with `first_uuid`, `last_uuid`, the line count and the capture time in the
`transcript` table. So there is no need for a hand-cut copy, and one was
deliberately not made.

⚠️ **Read it as history, never as the current position.** It contains two
arguments of mine that were later withdrawn, in their original confident form:
an attach-only Excel reader, corrected in s16, and a security objection to
co-locating shimp with its targets, corrected in s1. Flat text marks neither as
superseded. The rows do, which is why the rows are the record.

And `transcripts/` is gitignored, because the first capture was 33MB raw. So
this is a THIS-MACHINE fallback and not a portable one: a clone elsewhere will
not have it.

### s19. The URI grammar, and what identity IS

Frozen in the source conversation (s18) and not re-decided here. Recorded
because an audit found it in no row, which meant the core protocol lived only
in a 61,518 character chat log.

```
shimp://<capability-id>/<action...>
```

#### Identity

    identity = sha256(canonical descriptor)
    descriptor = { version, runner, target, argv, cwd }

⚠️ **The action is NOT part of identity.** Two links with different actions
address the SAME running instance, which is the whole point of separating
"ensure this is running" from "tell it to do something".

#### The rules that go with it

* **the descriptor is the authoritative object**, and the URI is the
convenient clickable representation of it. Descriptors are stored
content-addressed locally, and the hash of the canonical bytes is the id.
* **canonical JSON**, with `version` inside it, and an unknown version
**fails closed** rather than being interpreted hopefully.
* **a full immutable commit SHA, never a branch or a tag.** A branch makes the
same link mean different code on different days, which destroys the only
property this design has. Abbreviations are accepted on input and expanded
before storage, which is s5.
* ⚠️ **`cwd` must resolve deterministically BEFORE the descriptor is hashed.**
A caller-dependent `.` would make two identical-looking links mean different
things depending on which application was clicked in. Relative paths resolve
against a declared capability base, never against the browser's or Excel's
current directory.
* ⚠️ **NEVER ENCODE A SHELL COMMAND, and never invoke one.** argv is
constructed directly and passed as separate arguments. A URI that becomes
`sh -c` or `cmd /c` is an arbitrary code transport with a scheme in front of
it, and the quoting problems alone would be reason enough without the security
ones.

### s20. The keeper, the IPC, and the lifecycle

Frozen in the source conversation (s18), recorded here because an audit found
none of it in a row.

#### One keeper per capability identity

The keeper owns the instance lock, starts the target, exposes the IPC
endpoint, waits for the target, and exits. So the lock's lifetime IS the
instance's lifetime, which avoids the pid-reuse problem rather than papering
over it: a short-lived launcher cannot hold a lock for a process that outlives
it.

#### IPC

| | |
|---|---|
| transport | named pipe on Windows, Unix domain socket elsewhere |
| framing | length-prefixed UTF-8 JSON |
| request | `{"v":1,"action":"/thread/92817"}` |
| reply | `{"ok":true}` or `{"ok":false,"error":"..."}` |

Deliberately boring: no RPC framework, no reflection. Local only, so no port
to manage, which is also why s4's declared-ports rule is about the TARGET's
ports rather than shimp's own.

#### Lifecycle

* **the handshake.** The keeper creates the endpoint first, launches the target
with the endpoint in its environment, and waits for a `READY` message before
forwarding any action. A fixed startup timeout, with a clear failure when
readiness never arrives, rather than waiting for ever.
* **actions are request/response**, even where the caller ignores the reply,
because that is what distinguishes "started, but the action failed" from
"delivered".
* ⚠️ **concurrent clicks serialise around the identity, and re-check after
taking the lock.** Without the second check two simultaneous clicks both
observe "not running" and start two copies, which is the one thing a singleton
exists to prevent.
* **stale state is disposable and self-healing.** A stale pipe, socket or lock
is detected and recreated rather than requiring a human to clean up after a
crash.
* **the handler is never silently replaced.** Installing a different shimp SHA
refuses unless asked explicitly, because the handler is the trust root.
* **uninstall removes only what shimp installed**: its registration, its
binary, its runtime state. It does not kill unrelated target processes unless
told to.

#### And one of ours that changes the handshake's weight

s6's cold fallback makes the readiness race survivable: if the singleton is not
there yet, the clicking process does the work itself. So a slow `READY` costs
latency rather than correctness, which is a weaker requirement than the source
design assumed.
