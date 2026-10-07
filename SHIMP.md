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
| s9 | The audit is three lines, not two | **DECIDED** 7 Oct 2026, three lines | nothing, and the first descriptor is no longer blocked on this |
| s10 | A long action reports to a visible console, and the click never waits | **DECIDED** 7 Oct 2026 | nothing |
| s11 | Two clicks, not a library, because s15's refusal carries the safety | **DECIDED** 7 Oct 2026, and the library is deferred rather than rejected | nothing |
| s12 | chrome_sessions: liveness built, authority open | PART BUILT | the CDP extraction, s13 |
| s13 | Extract the CDP client, do not copy it | RAISED | a decision on where a shared package lives |
| s14 | patchbucket: transport and ordering, never merging | SPECIFIED | nothing, and it is unblocked |
| s15 | chatgpt_capture needs a SIGNED-IN Chrome, and must refuse without one | SPECIFIED | s12's authority probe |
| s16 | Excel is a session-protocol problem, not a prohibition | PARKED 7 Oct 2026 | a reason to drive Excel through a `shimp://` link |
| s17 | What is NOT in scope for SHIMP/1 | **DECIDED** | nothing |
| s18 | Where this design came from | **REFERENCE** | nothing |
| s19 | The URI grammar, and what identity IS | **DECIDED** in the source (s18), NOT implemented | nothing, except that s21 must close before the first descriptor is written |
| s20 | The keeper, the IPC, and the lifecycle | **DECIDED** in the source (s18), NOT implemented | nothing. These are the spec to build against. |
| s21 | Two more descriptor fields the source froze, and s19's list is short | **OPEN**, and it blocks the first descriptor exactly as s3 does | a decision on whether `policy` is inside the hashed bytes, and s29 on whether `base` survives at all |
| s22 | Windows first, and the spec must not learn the word "pipe" | **DECIDED** in the source (s18), NOT implemented | nothing. It is the build order. |
| s23 | Two audit questions, and only one of them is answerable today | **DECIDED** in the source (s18), NOT implemented | nothing |
| s24 | Four verdicts, because UNVERIFIED is the one that earns the set | **DECIDED** in the source (s18), NOT implemented | nothing |
| s25 | The command surface: nobody hand-builds a capability id | **DECIDED** in the source (s18), NOT implemented | nothing |
| s26 | What a target must carry before it can be a capability | **DECIDED** in the source (s18), NOT implemented, and it has a wrinkle here | nothing |
| s27 | SHIMP is Shim Protocol, and the name collision is known | **DECIDED** in the source (s18) | nothing |
| s28 | Adopting a descriptor: hash the bytes, never trust the id | RAISED, with one rule already clear | a decision on WHEN a descriptor is adopted, and s31 may dissolve this row |
| s29 | cwd derived from identity rather than declared in it | RAISED 7 Oct 2026, and it would close most of s21 | a decision, and s21 waits on this one |
| s30 | A singleton per session, and an immutable store per user | **DECIDED** 7 Oct 2026 | nothing |
| s31 | Two link forms, one identity: the short id and the self-describing link | RAISED 7 Oct 2026, and it would dissolve most of s28 | a decision on whether a second link form is legal |
| s32 | A link carries a SHORT capability id, because s5's threat model does not transfer | RAISED 7 Oct 2026, and it looks decidable immediately | a decision on how many hex characters |
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

#### And the same reasoning does NOT cover the capability id

Added 7 Oct 2026.
The argument above is about a GIT SHA, which names code an adversary may publish, so a prefix match is an attack worth money.
A capability id names bytes already in a local store, where an attacker needs write access before length matters at all.
s32 carries that distinction, because repeating "the adversarial number is the one that matters" for both objects is repeating the words rather than the argument.

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

### s9. The audit is three lines, not two

Settled 7 Oct 2026, from a re-read of all 33 messages of the source rather than the seven s18 names.
What gets pasted is the handler install command, the canonical descriptor and the URI.

```
go run github.com/davidnoz123/shimp@<handler-sha> install
{"version":1,"runner":"go","target":"github.com/davidnoz123/tool@<full-sha>","argv":["-root","."],"base":"capdir","cwd":"workspace"}
shimp://<capability-id>/<action>
```

#### The source left three doors and this row named two

[29] stated the requirement the two-line form fails: "the capability referenced by the URI must itself be independently resolvable and immutable.
If `8b724...` only means 'look in some mutable local database', ChatGPT cannot audit it from the pasted text alone."
It then offered two repairs, and this row considered only the first.

* **put the target in the URI.** Refused.
s17 already records that `HYPERLINK()` caps `link_location` at 255 characters, and Excel is the first surface this exists for.
A module path and a 40-hex SHA and an argv and a cwd overrun that between them, and [11] refused the form on quoting grounds as well.
* **publish descriptors at a deterministic public location.** This is the only thing that would make two lines honestly true, and it is out for SHIMP/1.
It costs a hosting story, a fetch at audit time against something that is not GitHub-at-a-SHA, and it turns creating a capability into a publishing act.
Recorded here because it is what would retire the third line later, and nobody should have to re-derive it.
* **three lines.** The remaining door.

#### The reason is stronger than "the third line supplies what is missing"

⚠️ **The third line makes the second falsifiable.**
A reader with no access to this machine can recompute `sha256(canonical descriptor)` and check it equals the capability id in the URI.
Under two lines that id is an assertion nobody can test.
Under three the payload contains no token only one machine can resolve, which is the property that actually mattered.
Two against three was never the axis.

#### Why the install line cannot be folded in

⚠️ **The same `shimp://` URI means different things depending on which program currently owns the `shimp:` scheme.**
That is [9]'s point and the load-bearing reason line 1 exists: it is the interpreter's identity, not provenance decoration.

The source also dropped `handler=` from the descriptor between [11] and [23], deliberately.
Folding it in would change every capability id whenever the handler was upgraded, which is s3's failure mode at a larger scale.
So three lines is the design's own shape, the three layers [27] names: interpreter, capability, action.

#### What it costs, stated rather than glossed

[29] proposed a design test: "if those two lines are not sufficient for an independent safety audit, the protocol is carrying too much hidden state."
Honestly read, SHIMP fails it, because a local content-addressed store is hidden state as far as a stranger is concerned.
Three lines is the smallest repair that keeps everything else intact, and the headline claim restates to something true and stronger: **an audit needs no access to the clicking machine.**

#### Two riders

* **nobody assembles canonical descriptor bytes by hand**, so `shimp audit-text <uri>` emits the three lines ready to paste.
[27] gives the reason: it stops a user omitting the handler SHA, which is the one omission that voids an audit without looking like it has. See s25.
* ⚠️ **an absolute `cwd` would publish this machine's filesystem layout into a chat window** on every paste.
[25]'s symbolic base is what keeps line 2 safe to paste, so it is not only a portability field. See s21.

#### What three lines does not fix

The target's `go.mod` and `go.sum` still need the auditor to fetch the target repo at that SHA.
That is acceptable, it is public GitHub, and it is s26 rather than an assumption.

#### What the webpage changes, 7 Oct 2026

Asked on the sheet: the two lines were for a page offered to customers, who install shimp from line 1 and then click line 2 to have something useful happen.
That is the use case the two-line form was designed for, and it is the strongest argument for three lines on the slate, for a reason that has nothing to do with auditing.

⚠️ **A FRESH MACHINE'S CAPABILITY STORE IS EMPTY.**
[15] puts descriptors in a local content-addressed store, so `shimp://<capability-id>/<action>` is resolved by looking the id up locally.
A customer who has just run line 1 holds no descriptor for that id.
The click finds nothing.
So on a webpage the two-line form does not merely under-audit the capability, **it cannot run it**.

Which re-casts what line 2 is for.
It is not a concession to auditors, it is **the distribution unit**: the page carries the descriptor, the machine adopts it, and only then does line 3 resolve to anything.
The audit property comes along free.

And the cost argument inverts on this surface.
The 255-character ceiling that ruled out a self-describing URI is `HYPERLINK()`'s, which is Excel's, and a webpage has no equivalent.
So the one place length was decisive is not the place this is for.

Two consequences, both recorded rather than assumed:

* **s28** is the question this raises: adopting a descriptor must hash the bytes and refuse on mismatch, never trust the id the page supplies.
* ⚠️ **a browser raises its own "Open shimp?" dialog**, which s8's Office key does nothing about, because that key is Office's.
Reported rather than promised, exactly as s8 reports HLINK.dll: a customer clicking from a page sees one prompt, and SHIMP cannot suppress it.

### s10. A long action reports to a visible console, and the click never waits

The action model is request/response, which fits a navigation at half a second and is wrong for a job that takes minutes.
A click either blocks invisibly or returns having said nothing.
Today the console shows `[full_update] step 1/3 ...`; a hyperlink has no console.

#### Decided 7 Oct 2026: a visible console, and acceptance rather than completion

Asked on the sheet, and the answer is yes on three independent grounds.

* **the house rule already requires it.** `AGENTS.md` forbids hidden windows, for the measured reason that a hidden modal dialog blocks a process with no way for anybody to diagnose the hang.
s8 is that failure in the wild: an Excel instance held a workbook for two and a half hours behind one modal, and the only symptom was a read-only refusal naming no pid.
A console is the honest channel, not a debugging aid.
* **s6 needs it anyway.** When the singleton cannot be reached the clicking process does the work itself, and that process is precisely the one with something to print and nowhere to print it.
* ⚠️ **the hanging has a mechanism, and this is the half that matters.**
A click must never be the thing that waits.
s20 keeps actions request/response so that "delivered" is distinguishable from "started but failed", and the reply returns on **acceptance**, carrying where the work reports, never on completion.
A minutes-long job whose reply arrives at the end is a hyperlink that hangs, which is the irritation this row exists to remove.

So the three candidates collapse into one answer rather than competing: the console is where it reports, the reply is the log or console handle, and `pdf_view_server/raise_windows.go` is how the finished console gets the user's attention without stealing it mid-job.

### s11. shimp as a library, not only a URI

chatgpt_capture needs a signed-in Chrome before it sweeps. Either it ensures
that itself, which means calling shimp from Go, or a human clicks two links in
the right order, which is fragile.

So the same logic needs a Go API and not only a scheme registration. Small, and
it changes what the package exports, so it is worth settling before the
handler is written.

#### And a consequence of keeper-owned IPC, from the source

[25] chose keeper-owned IPC over target-owned, and recorded a consequence that bears directly on this row: with the keeper owning the endpoint, **a target does not necessarily need a SHIMP library at all**, because the keeper can translate `{"action":"/thread/92817"}` into whatever mechanism is agreed for that one target.
That opens adapters for programs never written for SHIMP.

So this row is not only "does shimp export a Go API".
It is two separable questions: whether chatgpt_capture calls shimp's logic directly in-process, and whether a target receives actions through a shimp package or through an adapter the keeper holds.
Found by the s18 re-audit on 7 Oct, which found [25]'s consequence in no row.

#### Decided 7 Oct 2026: two clicks

Answered on the sheet: the human does two clicks, and the CDP instance click is idempotent.
That is right, and it dissolves the fragility this row claimed.

⚠️ **Idempotence makes the ORDER safe to get wrong, which is not the same as making OMISSION safe.**
Clicking the Chrome capability twice, or clicking it again after a sweep, costs nothing and changes nothing.
Skipping it is the dangerous case, and s15 already has that failure looking exactly like success: no token yields an empty catalog, which reads as "nothing to capture".

So the safety lives where s15 already put it, in a loud refusal when `/api/auth/session` yields no token, rather than in sequencing the clicks.
Two clicks plus that refusal is as safe as a library and smaller, so **no Go API for now**.

The library is deferred rather than rejected: the moment one capability must ensure another without a human in the loop, the API is back, and s28's adoption rule is the other thing that would want it.

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

#### Parked, 7 Oct 2026

Parked on the sheet until there is tooling that actually wants to reach Excel through a `shimp://` link.
Nothing is withdrawn: the constraint above is a property of COM apartments rather than a position anybody took, and s22's Windows-first build order means no capability needs it yet.
What parking costs is nothing, and what it buys is not porting 20,728 lines on speculation.

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

#### One precision added 7 Oct 2026

The guarantee above says "at most one managed live instance".
s30 pins the scope that sentence left out: **per interactive session**, not per machine, because the source's `Global\` mutex contradicted its own per-user descriptor store, and s10's visible console makes the session the right boundary rather than merely the safe one.

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

#### How to read the transcript

⚠️ **zlib, not gzip**, and the reason is recorded in slatex: gzip's header
carries an mtime, so identical bytes would compress to different files.

```python
import json, zlib
from pathlib import Path

raw = zlib.decompress(Path("transcripts/000001.jsonl.z").read_bytes())
lines = raw.decode("utf-8", "replace").splitlines()      # 14,264 of them
for line in lines:
    rec = json.loads(line)                               # one Claude Code event
```

34.5MB decompressed, and **64 lines of the 14,264 mention shimp at all**, which
is the measured reason a hand-cut copy was not made: almost all of it is
something else.

The `transcript` table also records `session`, which names the LIVE session
file in Claude Code's own projects directory, plus `first_uuid` and `last_uuid`
bounding this chunk. The live file is larger and still growing, so it holds
everything including what happened after the capture. Both are on one machine.

#### ⚠️ AND RE-AUDIT THE ROWS AGAINST THE CHATGPT SOURCE

Do not assume this slate captured everything. It did not, the first time: an
audit on 7 Oct found that **ten of the eighteen decisions the source
conversation froze were in no row**, which is how s19 and s20 came to exist.
The core protocol had been left in the chat log.

The method, and it is worth repeating rather than trusting:

```python
import sqlite3
rows = sqlite3.connect("file:db/slate.db?mode=ro", uri=True)
blob = " ".join(r[0] or "" for r in rows.execute(
    "SELECT value FROM cell WHERE col IN ('design','note','item')")).lower()
# then, for each decision in the source conversation, check it appears
```

One caution learned doing it: the probe searched for the WORDING of the old
prose rather than the decision, and reported a rule missing that was present
under different words. Check a miss by eye before believing it.

### s19. The URI grammar, and what identity IS

Frozen in the source conversation (s18) and not re-decided here. Recorded
because an audit found it in no row, which meant the core protocol lived only
in a 61,518 character chat log.

```
shimp://<capability-id>/<action...>
```

#### Identity

    identity = sha256(canonical descriptor)
    descriptor = { version, runner, target, argv, base, cwd }

⚠️ **The action is NOT part of identity.** Two links with different actions
address the SAME running instance, which is the whole point of separating
"ensure this is running" from "tell it to do something".

⚠️ **AND THIS FIELD LIST WAS SHORT.**
`base` is restored above because [25] froze it and a re-audit on 7 Oct found it in no row.
Whether `policy` joins it is the open part of s21, and that question has to close before the first descriptor is written, for s3's reason.


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

### s21. Two more descriptor fields the source froze, and s19's list is short

s19 freezes `descriptor = { version, runner, target, argv, cwd }`.
The source froze two more fields, and the re-audit of 7 Oct found neither in any row.

#### `base`, the symbolic working-directory base

[25] chose it as the first choice over absolute paths: `base=capdir` with `cwd=workspace/project-a`, so one logical capability survives moving between a Windows user directory and a macOS one instead of hashing differently on every machine.

⚠️ **s19 kept the consequence and dropped the field that carries it.**
It says relative paths resolve against a declared capability base, which is right and has nowhere to declare the base.

s9 then gives `base` a second and independent reason: the descriptor is now pasted into a chat window, so an absolute `cwd` publishes this machine's filesystem layout on every audit.

#### `policy`, which is the genuinely open part

[23] lists the capability contents as `version`, `target`, `argv`, `cwd` and **optional metadata/policy**.
s17 records what a policy field MEANS, declarative only and never enforced, and never records that it is a field of the descriptor at all.

So the question is narrow.
Inside the hashed bytes, and two capabilities differing only in a claim about themselves are different capabilities, with different ids and separate singletons.
Beside them, and the claim is not covered by the id a link carries, so an audit cannot check the claim against the thing that was clicked.

#### Why this is urgent rather than tidy

⚠️ s3's argument applies unchanged and with more weight, because this is two fields rather than one.
Adding either later changes every descriptor's canonical bytes, which changes every capability hash, which changes every capability id, **which breaks every link ever written.**
s3 got a whole row for `runner`. These got none, which is how a re-audit came to find them.

#### The same thing, for a ten year old

Think of a **recipe card**.
It says which program to run, exactly which version of it, what to type after it and which folder to be in when you do.

A link is a sticker with the card's **secret number** on it.
You get the number by adding up everything written on the card, and nothing else.

⚠️ **So adding a new line to the card changes every card's number.**
Every sticker anybody has ever written now points at a number no card has.
The stickers do not break noisily. They just quietly stop finding anything.

That is why every line on the card has to be agreed BEFORE the first card is written.
It is the whole of s3, and it turns out two more lines were forgotten.

#### The first forgotten line: "start counting from here"

A card that says "be in the `workspace` folder" has not actually said where.
`workspace` inside what?
Your computer would pick one place and mine would pick another, so the same card would mean two different things, which is exactly what the secret number exists to make impossible.
So the card needs a line saying where to start counting from, and [25] calls it `base`.

There is now a second reason, which is newer than the row above.
s9 decided the card gets pasted into a chat window so a stranger can check it.
A card holding the full path to a folder inside my own user account tells that stranger how my computer is laid out.
A card that says "start at the capability folder, then `workspace`" tells them nothing about me and still means exactly one place.

#### The second forgotten line: a promise the program makes about itself

A card can carry a note saying "this program only needs to touch its own folder".

⚠️ **That note is a CLAIM and not a lock.**
Nothing stops the program doing more, and s17 is firm about it: SHIMP reads the note out, it does not enforce it.
The note is useful anyway, because somebody checking the card can go and see whether the program really behaves that way.

The part nobody has decided is whether that note is part of the adding-up.

* **inside the sum:** two cards identical except for the note are different recipes, with different numbers, and each gets its own one-running-copy.
* **outside the sum:** the note is not covered by the number on the sticker, so a person checking the sticker cannot tell whether the note they read is the note that was used.

Either answer is workable.
Having no answer is the thing that is not, because the first card cannot be written until it is settled.

#### "So this is about hashing without absolute paths?"

Asked on the sheet, 7 Oct 2026.
For `base`, yes, and that is the whole of it.
For `policy`, no.
The two halves of this row are one problem rather than one topic: both are fields missing from s19's list, which is s3's failure happening twice, and only one of them is about paths.

**The fix is not "use relative paths instead".**
⚠️ **A bare relative `cwd` is WORSE than an absolute one**, because `.` is resolved against whoever clicked the link.
s19 already names the consequence: two identical-looking links mean different things depending on whether Excel or a browser was clicked in.

| what the card stores | deterministic | portable | safe to paste |
|---|---|---|---|
| an absolute path | yes | no, it hashes differently on every machine | no, it publishes your layout |
| a bare relative path | **no** | looks portable | yes |
| `base` plus a relative tail | yes | yes | yes |

Only the third is all three at once, which is why the answer is a named anchor rather than a shorter path.
`base` names one of a small set of meanings, `capdir` and perhaps `home`, each with a single definition.

⚠️ **And the resolution happens at LAUNCH, never at hash time.**
The descriptor stores the symbol and the tail, so what gets hashed is the symbol and the number is identical everywhere.
What the process actually receives is the directory that symbol resolves to on the machine it starts on.
Resolve before hashing and the portability is gone again, which is the trap the words "resolve deterministically before hashing" in s19 can be read into: it means the SYMBOL must have one unambiguous meaning, not that the path must be expanded into the bytes.

**One correction to the question.**
Nothing hashes a command.
s19 forbids a command string existing anywhere, because a URI that becomes `cmd /c` is an arbitrary code transport with a scheme in front of it.
So what is hashed is the descriptor, and the descriptor holds an argv array rather than a line of shell.
There is no command to hash at any point in the design.

**Which half of this row is actually open.**
`base` is frozen. [25] chose it over absolute paths and nothing since has argued with that.
The open decision is `policy` alone, which is what this row's `waiting on` cell has said from the start while the body above read as though both were undecided.

#### "Assuming `shimp://<path>/<params>`, do we need cwd in the path part?"

Asked on the sheet, 7 Oct 2026.
The question splits in two and the answers go opposite ways.

**In identity: yes.**
[5] defined identity as runner, module, commit, args and working directory, and [23] froze it.
The reason is not tidiness.
A tool told `-root .` has its cwd as the ONLY thing distinguishing one job from another, so leaving cwd out of the hash collapses two links on two different folders into one capability id.
⚠️ **The second click is then answered by an instance pointed at the wrong data, and nothing reports it.**
It is the same property s4 is pleased about from the other side: put the profile and the port in a capability's argv and the capability hash IS that browser's identity, so no second registry is needed.
Where argv already names the data, cwd is redundant rather than wrong, and a design cannot depend on every tool being written that way.

**In the URI text: no, never.**
The first segment is the hash of the descriptor, and the descriptor holds cwd.
So cwd is in the "path" part completely by reference and not at all literally, which is the entire point of addressing a capability by content.
It is also what buys the headroom s17 records: `HYPERLINK()` caps `link_location` at 255 characters, and a form carrying module, SHA, argv and a directory overruns that.
If cwd appeared in both the URI and the descriptor there would be two copies of one fact, and two copies can disagree.

⚠️ **And the trap the proposed shape invites: NOTHING after the first slash may affect identity.**
If `<params>` could set cwd, then one capability id would launch different processes depending on what was appended, and the singleton would be meaningless.
s19 freezes the action as outside identity precisely so that two links with different actions reach the same instance.
Params are a message to a running process, never configuration for starting one.

If `<path>` was meant as a module path, as in `shimp://github.com/me/tool@abc123/open/xyz`, that was the early form in [5] and [7] and [15] and [23] replaced it with the capability id.
The reason is this question: once startup needs a cwd, an argv and a base, the readable form stops being readable and starts being quoted.

**One consequence for s9's webpage, and it is a third independent reason for `base`.**
One folder is one capability, one id and one singleton, which is correct and means the folder has to be decided when the descriptor is authored.
⚠️ You cannot know a customer's folder when you write the page.
An absolute path is unknowable at authoring time, and a caller-relative path is non-deterministic by the table above.
So a capability that can be distributed on a page is only writable as `base=capdir` plus a relative tail.
The webpage case does not merely prefer `base`; it is impossible without it.

#### s29 may delete half of this row

Raised 7 Oct 2026: if cwd is derived from the capability id rather than declared, then `cwd` leaves the descriptor and `base` has nothing left to anchor.
Half of this row would then be answered by deletion rather than by decision, and the `policy` half would be all that is left.
See s29, which also records why a derived cwd cannot itself be a descriptor field.

### s22. Windows first, and the spec must not learn the word "pipe"

[21] lists platform scope among the ten decisions that had to be frozen before the design could be called ready, [23] froze it, and the re-audit of 7 Oct found it in no row.

**Windows gets the first complete implementation**, with the protocol defined cross-platform from the start.
[25]'s reason is that URI registration, singleton lifetime, keeper behaviour, process launching, the startup handshake, IPC and crash cleanup are already enough difficult semantics.
Adding Launch Services and Unix sockets at the same time makes it hard to tell whether a problem is in the protocol or in one platform's implementation of it.

⚠️ **And the rule that makes Windows-first safe rather than a trap: do not let Windows leak into the spec.**
The spec says *local IPC endpoint*.
The Windows implementation happens to use a named pipe.
s20's IPC table names both transports, which is right for an implementation note and wrong in a protocol sentence.

This repo already has the shape for it, and `AGENTS.project.md` makes the catch-all file non-optional: `owner_windows.go` against `owner_other.go` defining the same symbols, where the non-Windows file says what it cannot do rather than returning a value that reads as an answer.
The second choice recorded in [25] was Windows and macOS together, which pressure-tests the abstraction earlier and is only worth it if cross-platform deployment is immediately needed.

### s23. Two audit questions, and only one of them is answerable today

[27] froze two distinct audit modes, and the re-audit of 7 Oct found the second in no row.

| question | the evidence it needs |
|---|---|
| can I safely install this and click this link | the install command's SHA, which is s9's line 1 |
| is what is already installed on this machine the thing I think it is | the installed binary's own account of itself |

The second needs a command, and [27] specified it:

```
shimp version --audit
    SHIMP/1
    repo=github.com/davidnoz123/shimp
    commit=<full sha>
    binary_sha256=<hash of the running binary>
```

⚠️ **Which means install has to record its own provenance while it still knows it.**
s7 has the binary copied from `os.Executable()`, which under `go run` is the toolchain's temporary build output.
The SHA that produced it is known to the installing process and to nothing afterwards unless that process writes it down.
A handler that cannot say which commit built it turns the first question's answer into a claim about a machine that nobody can check, which is s9's defect one level down.

`verify` is the natural home for the report, since s8 already gives it the job of saying what is true rather than what was intended.

### s24. Four verdicts, because UNVERIFIED is the one that earns the set

[27] froze the vocabulary an audit answers in, and the re-audit of 7 Oct found it in no row.

| verdict | means |
|---|---|
| SAFE | no material capability beyond the declared contract was found |
| CONDITIONAL | safe only under stated assumptions or permissions |
| UNSAFE | a concrete dangerous capability or violation was found |
| UNVERIFIED | important source, dependency, binary or runtime behaviour could not be established |

⚠️ **UNVERIFIED is what makes the other three honest.**
Three verdicts force a reader who could not fetch the target, or could not resolve a dependency, to choose between a reassurance and an accusation.
That is the failure this repo names everywhere else.
`AGENTS.project.md` puts it as "say what a tool cannot tell you", and s12 is the worked example: liveness was reported, authority was not, and the dangerous version would have been a single word covering both.

[27] also specified what the report lists beneath the verdict: what exact code will execute, what resources it can reach, what persists afterwards, what network access exists and any unverified assumptions.

### s25. The command surface: nobody hand-builds a capability id

[15] added three scope items it called easy to miss, and the re-audit of 7 Oct found two of them in no row.

* **`shimp cap create ...`** produces the canonical descriptor and its hash.
⚠️ A human assembling canonical bytes by hand gets a different hash for the same capability, and the symptom is a link that resolves to nothing rather than an error naming the cause.
s5 already assumes this command exists: the full-SHA pedantry "appears once, when a capability is created", and that moment is this command.
* **`shimp inspect <id>`** prints the exact target, SHA, argv, cwd and current running state.
This is what makes a local content-addressed store auditable by its owner, and s9 leaves it load-bearing: what a stranger cannot resolve, the machine's owner must be able to.
* **`shimp audit-text <uri>`** emits s9's three lines ready to paste.
[27]'s reason is that it stops a user omitting the handler SHA, which is the one omission that voids an audit without looking like it has.

With s8's `verify` and s23's `version --audit`, that is the whole surface besides `install`, `uninstall` and the handler's own URI dispatch.

### s26. What a target must carry before it can be a capability

[15] froze a requirement on any target: a full SHA **plus committed `go.mod` and `go.sum`**, with the Go toolchain version pinned for high assurance, and an optional later `deps-digest` field hashing the resolved module graph.
s2 records `go.sum` as a property the Go runner gives us.
The re-audit of 7 Oct found nothing recording it as something a target must satisfy, which is the different and checkable claim.

⚠️ **And here it has a wrinkle worth stating rather than discovering.**
`AGENTS.project.md` makes zero third-party dependencies a property of every connector in this repo, so a connector's `go.sum` is empty or absent.
A check written as "the target has a committed `go.sum`" therefore fails every target this repo will ever offer, for the best possible reason.
So the check is on the pair: a `go.mod` with an empty `require` block and no `go.sum` is **stronger** evidence than a `go.sum` full of hashes, and an audit that cannot tell those two apart is reporting the wrong thing.

#### `deps-digest` cannot be added later, so it is never a field

⚠️ By s3 and s21's argument there are only two honest positions for a descriptor field: it is in the FIRST descriptor, or it is never in a descriptor at all.
"Optional, added later" is not available, because adding it changes every capability id.
[15] put `deps-digest` out of v1, so the second position follows: dependency provenance belongs in s24's audit report rather than in the hashed bytes.

### s27. SHIMP is Shim Protocol, and the name collision is known

[19] settled the name, and the re-audit of 7 Oct found it in no row, which left the expansion and the terminology as folklore.

**SHIMP = Shim Protocol.**
A shim is a small interposition layer between two systems, and this is one: between an OS URI scheme and a content-pinned process invocation.
The expansion describes the abstraction rather than the Go implementation, so it survives s2 relaxing.

The terminology that comes with it, worth using consistently:

| term | means |
|---|---|
| shimp handler | the OS URI shim, and the trust root |
| shimp capability | an immutable process definition |
| shimp action | a message delivered to that process |
| SHIMP/1 | the protocol version |

#### The collision was searched for, not assumed away

⚠️ **SHIMP is an established term in vestibular medicine**, the Suppression Head Impulse Paradigm, in use since about 2016.
So searching "shimp protocol" today returns the medical meaning, and that ambiguity is a known cost rather than a surprise waiting inside a README.
The domains are far enough apart that [19] recommended keeping the name, and ranked `Latch` as the alternative if the name ever has to carry its meaning to strangers.

### s28. Adopting a descriptor: hash the bytes, never trust the id

Raised 7 Oct 2026, out of s9.
Once a webpage is the way a capability reaches a customer, something on that machine has to take the descriptor from the page and put it in the local store, and nothing in the source conversation covers that step.

#### The rule that is already clear

⚠️ **Adopt by hashing, never by being told the hash.**
The id a page prints beside a descriptor is the page's claim.
The store's key must be `sha256(canonical descriptor)` computed locally from the bytes, and a mismatch against the id in the URI is a refusal rather than a warning.
Otherwise a page can map an id a reader audited onto bytes that reader never saw, which breaks the single property this design has.

#### The open part

**When does adoption happen, and does it need consent?**

Silent adoption is defensible, and the reason is worth stating because it looks alarming: a descriptor in the store is inert.
It names a process; it does not run one.
Nothing executes until a link is clicked, and the clicked link is the consent.
s20 refuses to replace the HANDLER silently because the handler is the trust root; a capability is not, as long as its id is verified against its bytes.

Against that, an explicit step gives the customer something to read before anything is written, and the three-action flow is not obviously worse than two.
The decision needs one of: install takes descriptors as arguments, a `shimp cap add` the page links to, or the handler adopting on first click of an unknown id with the descriptor alongside.
The third is the smallest and is the one that needs the hash rule hardest.

Related: s25's `cap create` is the author's side of the same object, and this is the consumer's.

#### s31 may dissolve this row

Raised 7 Oct 2026: if a self-describing bearer link is legal, the descriptor arrives WITH the click rather than ahead of it, and this row's open question stops being "when is a descriptor adopted" because the answer is "on the click that carries it".
The rule above survives either way, and under s31 it stops being an adoption policy and becomes the bearer form's defining behaviour.

### s29. cwd derived from identity rather than declared in it

Raised 7 Oct 2026, from the observation that a capability is a system-wide singleton, so the folder it runs in cannot sensibly be some arbitrary place the author happened to be standing.

The proposal: the working directory is `~/.shimp/caps/<capability-id>/work`, owned by shimp and derived from the id.

#### Taken fully, it removes TWO fields rather than adding one

⚠️ **A derived cwd must not be IN the descriptor, or there is no fixed point.**
The id is the hash of the descriptor. If the descriptor holds cwd, and cwd is a function of the id, then computing the hash requires the hash.
That is not a hard problem to work around, it is a problem that dissolves: a value derived from the id carries no information, so it does not belong in the hashed bytes at all.

So the descriptor loses `cwd`.
And with no cwd to anchor, `base` has nothing left to do, so the descriptor loses that too, which is half of s21 answered by deletion rather than by decision.

What remains is clean: **identity is what code, at what commit, with what arguments.** Nothing about where.

And it is exactly what s9's webpage needs.
A descriptor containing no path is machine-independent by construction rather than by a careful choice of symbol, so the same bytes hash to the same id on every machine, and there is nothing in the pasted line 2 that describes anybody's disk.

#### ⚠️ The catch: it relocates the path problem, it does not solve it

Most of the tools this is for operate on data that lives somewhere else: a workbook in a repo, a database in `chatgpt_tools`, a folder of captures.
A connector started in an empty per-capability directory finds none of it.
So the data path moves out of cwd and into argv.

That is a real improvement rather than a shuffle, and the reason is worth keeping: an implicit cwd that silently differs per machine is the dangerous kind of path, and an explicit `--db <path>` sitting in the pasted descriptor is one an auditor can read.
But portability and leaking come with it, so they are not gone.

[25] anticipated this exactly: support symbols such as `$HOME` or `$CAPDIR`, resolved deterministically before identity is hashed.
⚠️ So `base` does not disappear so much as **change job**: from "what is cwd relative to" into "the expansion rule for argv".
Whether that is still a descriptor field, or a rule the whole descriptor is read under, is the part this row has to settle.

#### ⚠️ And the new directory has a lifecycle that contradicts two frozen rules

s20 says runtime state is disposable and self-healing, detected and recreated after a crash.
s17 says uninstall removes only what shimp installed: its registration, its binary, its runtime state.

If `~/.shimp/caps/<id>/work` holds a tool's real working data, it is **not** disposable, and an uninstall removing "what shimp installed" would delete the user's files.
Both rules stay true only if the directory is declared SCRATCH and targets are forbidden from keeping anything there that matters.
The alternative is that uninstall refuses it and says why.
Either is fine. Having neither written down is how a cleanup becomes a data loss, so it belongs in this row before anything is built.

### s30. Local or Global: the singleton's scope, and the two halves disagree

Found 7 Oct 2026, while answering what "system-wide singleton" actually means.
This is not an open choice the source left. It is two frozen decisions that contradict each other.

| frozen in | says | scope |
|---|---|---|
| [5] | the lock is a named mutex, `Global\shimp-<hash>` | every logged-in session on the machine |
| [15] | descriptors live under `~/.shimp/caps/<sha256>` | the calling user alone |

⚠️ **So two users on one machine would share a singleton while holding separate capability stores.**
User B clicks a link, the lock says an instance is already live, and the action is forwarded to a process that user A started, running under A's account, against A's files, writing to A's directories.
B's own descriptor for that id might not even be the one that was launched, because B's store was never consulted.

The failure is quiet in the way this repo keeps naming: B's click SUCCEEDS.
Something ran. The reply says delivered. The work happened somewhere B cannot see and to data B does not own.

#### Only one answer looks defensible

Per-user, which on Windows means `Local\shimp-<hash>` rather than `Global\`, because `Local\` is scoped to the caller's session.
The descriptor store is already per-user and that is the right scope for it: a capability is a thing a person installed, not a thing a machine has.
s17's guarantee then reads "at most one managed live instance **per user**", which is what it has to say rather than what it currently says.

The case for `Global\` is a shared service two users should not duplicate, and SHIMP/1 explicitly is not a service manager, s17.
Where a target genuinely must be machine-wide, that is the target's problem and it can say so, rather than every capability inheriting the hazard above by default.

Worth stating plainly: this is a documentation-stage find, not a bug report. Nothing is built, so the cost of fixing it is one word.

#### Decided 7 Oct 2026, and the intent was never in doubt

To be clear about whose decision this was: **`Global\` came from [5] in the source conversation**, not from anybody here.
Asked, the intent was per-user and had been all along, which is what makes this a contradiction inside the frozen design rather than a disagreement about what was wanted.

So the lock is NOT machine-wide. But per-user is not quite the answer either.

#### Per-session, and s10 is why

⚠️ **`Local\` on Windows is scoped to the SESSION, not to the user.**
The same person logged in twice, over RDP and at the console, or across fast user switching, gets two `Local\` namespaces and therefore two singletons, while sharing one `~/.shimp`.

That looked like a mismatch to accept until s10 was decided, and now it is the better answer rather than the tolerable one.
s10 puts a long action's output in a **visible console**.
A console belongs to one session's desktop and cannot be shown on another, so an instance running in session 1 physically cannot report to a user sitting in session 2.
A per-user singleton would hand session 2's click to a process whose output session 2 can never see, which is s10's irritation restored by a different door.

So: **one live instance per interactive session.**

#### The split that falls out, and it is a clean one

| state | scope | why |
|---|---|---|
| descriptors, the content-addressed store | per user | immutable, content-addressed, and nothing is gained by duplicating them per session |
| the instance lock, the IPC endpoint, the per-capability working directory | per session | they name a LIVE thing, and liveness is what a session bounds |

Immutable state per user, live state per session.
s17's guarantee has to be read with that in it: at most one managed live instance **per session**, not per machine.

#### ⚠️ And one asymmetry to report rather than hide

Windows gives per-session scoping for free in its object namespace.
The Unix equivalent, `$XDG_RUNTIME_DIR`, is per-user and typically shared across that user's sessions, so it is COARSER.
The protocol says the live-state scope is the user's interactive session and each platform implements the closest thing it has.
Where a platform cannot be that precise, it says so, per the house rule about saying what a tool cannot tell you.
s22 keeps the spec from naming either mechanism.

### s31. Two link forms, one identity: the short id and the self-describing link

Raised 7 Oct 2026, from the question of why the URI is not the self-describing form the design started with:

```
shimp://go/run/github.com/...?go_arg1=xyz&arg2=bbb/actionx?arg1=123&arg2=1234
```

That was the form in [5] and [7], and [15] and [23] replaced it with `shimp://<capability-id>/<action>`.
Worth being straight about why, because the fat form has a real advantage and the short one has to buy it back.

#### What the fat form genuinely buys

It is **self-describing**, which is the property s9 spends a whole third line acquiring.
Taken fully it deletes a surprising amount: s9's line 2, s28's adoption problem, and the empty-store failure that stops a fresh customer's first click resolving at all.
The customer flow becomes install, then click, with nothing in between.
That is what was wanted, so the short form is a cost rather than an obvious win.

#### ⚠️ The objection that settles it: canonicalisation, which was in no row

If identity is the hash of the URI, then **every spelling difference is a different singleton.**

Query parameter order. Percent-encoding choices, `.` against `%2E`. `+` against `%20`. A trailing slash. Case in the host part. The order of repeated `arg=` keys, which URI query semantics do not guarantee any parser preserves.

Each variation hashes differently, so two links a human would call identical start two copies, which is the one thing a singleton exists to prevent, and it fails silently.

The fix is to canonicalise the URI before hashing. But a canonical form for a structure with arrays, escaping and Unicode in it **is** the descriptor, just expressed in a syntax with worse escaping rules than JSON.
[25] made exactly this point choosing JSON over line-based text: once you need arrays, escaping and nested metadata, you are inventing a format, and JSON already solved those problems.

⚠️ **So the descriptor is not an extra layer. It is the canonical form the fat URI would have needed anyway.**
The only real question is whether it is also the transport.

And the short form has a property worth stating positively rather than defensively: **it has no canonicalisation surface at all.** Sixty-four hex characters and a path. There is nothing to spell two ways.

#### Two smaller objections, stated for completeness

* ⚠️ **the example's grammar does not parse.** Only the first `?` starts a query; everything after it is query. So `/actionx?arg1=123` is part of the first query string rather than a separate path and query, and recovering the three parts needs a custom mini-grammar inside the query string. Which is, again, inventing a format.
* **length, measured rather than asserted.** `shimp://go/run/github.com/davidnoz123/nielsoln_connectors/chatgpt_capture@` plus a 40-hex SHA is **114 characters before a single argument**. s17 records `HYPERLINK()` capping `link_location` at 255, so roughly half the budget is gone on provenance alone. The Hyperlink OBJECT has no such cap, so this bites a formula and not a VBA-inserted link, which is worth knowing rather than overstating.

#### The proposal: both forms, one identity

The tension is real and it is not resolvable by picking a side, because the two surfaces want opposite things.
Excel wants short. A webpage and a chat window want self-describing.

So make two forms legal, distinguished unambiguously by their first segment:

| form | looks like | for |
|---|---|---|
| **reference** | `shimp://<64-hex>/<action>` | Excel, and anything on a machine that already holds the descriptor |
| **bearer** | `shimp://v1+<encoded canonical descriptor>/<action>` | a webpage, a chat window, a first click on a fresh machine |

⚠️ **And the point that makes it work: the id is the SAME in both.**
Both are the hash of the same canonical descriptor bytes, so a bearer link on a page and a reference link in a workbook address **one singleton**.
The bearer form is how a capability ARRIVES; the reference form is how it is referred to afterwards.

Encoding the canonical bytes rather than a query string keeps the canonicalisation objection answered: the encoding of an already-canonical byte string is itself canonical, so there is still nothing to spell two ways.

What it would change elsewhere:

* **s28 mostly dissolves.** Its rule, hash the bytes and never trust the id, becomes the bearer form's defining behaviour rather than a separate adoption step.
* **s9 keeps three lines for the reference form and needs only two for the bearer form**, because a bearer link IS its own line 2. That is a genuinely better answer than either row currently gives.
* the bearer form is long, so it does not rescue the Excel case, and nothing here claims it does.

### s32. A link carries a SHORT capability id, because s5's threat model does not transfer

Raised 7 Oct 2026, from an objection that the 40-hex SHA is too pedantic.
Two separate things, and the first is already settled in the objector's favour.

#### Nothing types a SHA, and no link contains one

s5 decided it: accept an abbreviated SHA, store the full one, and the pedantry "appears once, when a capability is created, and never in a link, because nobody types a `shimp://` URI and a tool emits it."

So under the reference form there is no SHA in a link at all.
The 40 hex characters measured in s31 are the cost of the **fat self-describing form**, which puts the target in the link, and they are therefore an argument against that form rather than a requirement imposed by this one.

#### ⚠️ But the objection lands somewhere worse, and this part is conceded

The capability id is **sha256, so 64 hex characters**, which is *longer* than the 40-hex SHA being objected to.
If the complaint is that there is too much hex in a link, the current form is the bigger offender, and s31 quietly used the fat form's SHA as the example while the short form was carrying more.

#### ⚠️ And the reason for the strictness does not transfer

s5's argument is specific and it is about GIT SHAs:

> the thing being authorised is **code execution on click**, so the adversarial number is the one that matters

That holds because a git SHA names code **an adversary may publish**. A prefix match is then an attack: grind out a commit whose hash shares the prefix, get it fetched, and the link runs it.

A capability id is a different object with a different threat model.
It names **bytes already on this machine**, in a local content-addressed store.
For a shortened id to resolve to the wrong descriptor, an attacker must get a colliding descriptor INTO that store, and anybody who can write the store has already won at full length.
s28's rule closes the remaining door: the store is keyed by what the bytes actually hash to, never by an id somebody supplied.

⚠️ So the 64 hex is **inherited pedantry rather than reasoned pedantry**, and applying s5's asymmetry to it is repeating the words rather than the argument.

Under s31's bearer form it is weaker still: the descriptor travels with the link, so the id is a CHECK against bytes in hand rather than a lookup key, and a shorter check is still a check.

#### What is left is only accidental collision, which is arithmetic

| prefix | bits | collision chance across 1,000,000 capabilities |
|---|---|---|
| 8 hex | 32 | far too high to consider |
| 12 hex | 48 | about 1 in 550 |
| 16 hex | 64 | about 3 in 100,000,000 |

16 hex looks comfortable and 12 is defensible.
`shimp://8f74b19c5d2a3e01/thread/92817` is the shape, which is short enough to read and to sit in a cell.

And the rule that makes any prefix safe is the one git already uses: **refuse an ambiguous prefix rather than guess at it.** The full id stays in the store, exactly as s5 has the full SHA stay in the descriptor.

#### What this does not change

* the descriptor still hashes to a full sha256. Only the LINK is abbreviated.
* the IPC endpoint and the lock keep using the full identity internally, where length costs nothing.
* s5 stands unaltered for the target SHA, because that argument is about published code and it is still right.
