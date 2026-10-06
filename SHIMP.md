# SHIMP, and the connectors it launches

Working notes, 7 October 2026.
Decisions that are settled, decisions that are not, and the reasoning for each.
Not a specification yet: the chat that produced the idea judged itself
"75% scoped at the architectural level, perhaps 45 to 50% at protocol-spec
level", and nothing here changes that.

## What SHIMP is

A custom URI scheme whose URI identifies an immutable process specification.
Clicking it ensures exactly one instance of that exact software is running,
then forwards the rest of the URI to it as an action.

```
go run github.com/davidnoz123/<repo>/shimp@FULL_HANDLER_SHA install
shimp://<capability>/<action>
```

The distinguishing property is that the whole trust chain is meant to be
auditable from what a person can paste into a chat window.

## It already exists once, in Python, by accident

`slate_tools/slate_open.py` is a single-purpose SHIMP, built on 5 October
before this idea had a name.

| SHIMP/1 | what `slate://` already does |
|---|---|
| handler registered for a scheme | `--install`, four HKCU keys |
| singleton per capability identity | a daemon keyed by a hash of its own checkout path |
| identity that survives pid reuse | `(pid, started_at)` plus a nonce in a session file |
| local IPC | 127.0.0.1 socket, nonce-authenticated |
| URI remainder as an action | `<repo>/<slate>/<item>/<col>` |
| request/response rather than fire-and-forget | `ok:` and `no:` replies |

⚠️ **And one thing SHIMP/1 does not specify that `slate://` has: the cold
fallback.**
If the singleton cannot be reached, the clicking process does the work itself.
Measured: 828ms cold against 526ms warm.
So the singleton is an optimisation and never a dependency, and there is no
window in which the link is broken, only one in which it is slower.
This belongs in the spec: it makes the readiness race that SHIMP's startup
handshake worries about mostly moot.

## Settled

**Declared ports, never allocated ones.**
A capability's identity must be computable before the process runs, so "pick a
free port" cannot be part of it.
A connector that genuinely needs a dynamic port must report it after starting,
over the keeper's IPC, and never take it as part of identity.
This is the same reasoning that ruled out a caller-dependent `cwd`.

**One capability per Chrome, rather than a port manager.**
A manager of instances and their ports is a singleton with a registry, which is
what SHIMP already is for processes.
Put the profile and the port in the capability's argv and the capability hash
IS that browser's identity.
No second layer doing the same job.

**Accept an abbreviated SHA on input, store the full one.**
Exactly what git does.
The pedantry then appears once, when a capability is created, and never in a
link, because nobody types a `shimp://` URI and a tool emits it.

Matching a 7-hex prefix by brute force is about 2.7e8 attempts, which is
minutes.
At 12 hex it is about 3e14, which is expensive and not out of reach.
Accidental collision is a non-issue for small repos at 7 hex.
The asymmetry that decides it: the thing being authorised is code execution on
click, so the adversarial number is the one that matters.

**The Office trusted-protocol key, per scheme.**
`HKCU\Software\Policies\Microsoft\Office\16.0\Common\Security\Trusted Protocols\All Applications\shimp:`
The trailing colon is part of the key name.
Without it, every click from Excel or Word raises a modal, and that modal
blocks COM invisibly: measured on 5 October, an Excel instance held a workbook
for two and a half hours behind one, and the only symptom was a read-only
refusal naming no pid.
`install` should claim it, `verify` should report it, `uninstall` should
withdraw only its own key.

**`HYPERLINK()` caps `link_location` at 255 characters.**
The Hyperlink object does not.
A URI carrying a sha256 capability id plus an action is comfortably under, and
it is an argument for the capability-id form over inlining target and args.

## Open, and these block code

**Can a capability target a pinned Python?**
This is the central question and the others are consequences of it.
Two of the three connectors wanted are Python today and large: the slate Excel
daemon, which holds a COM apartment through pywin32, and `chatgpt_session.py`,
which is 8,200 lines.
If SHIMP runs only `go run <module>@<sha>` then most existing tooling cannot be
a target and every connector is a rewrite.
If a capability can pin a Python tree by git SHA, fetch it to a
content-addressed cache and run it with a declared interpreter, the direction
works immediately and the audit chain still holds.

**Is the audit two lines or three?**
The chat chose `shimp://<capability-id>/<action>`, keeping startup details out
of the link, AND claimed the audit payload is the install command plus the URI.
Those conflict.
A local sha256 of a local descriptor cannot be resolved by a stranger, so
either the URI carries the target SHA and is self-describing, or the paste is
three lines: install command, descriptor, URI.
Three lines looks right: it keeps links short and stops the audit resting on
something only one machine can resolve.

**One module or many.**
Each connector is its own module today.
A shared CDP client argues for one module with `cmd/` subdirectories, because
then one SHA pins a connector AND its shared code, which is the property the
audit story wants.
Separate modules plus a shared module means auditing one connector requires
resolving a second SHA, and connectors drift onto different client versions.

**Where a long action reports.**
SHIMP's request/response model fits a navigation at half a second.
`full_update` is minutes: two rclone transfers and a catalog sweep at 0.5s per
session.
A hyperlink has no console, so a click either blocks invisibly or returns
having said nothing.
Needs a `status` action to poll, or a log path returned on start, or a
completion notification.
`pdf_view_server/raise_windows.go` already has the primitive for the last.

**shimp as a library, not only a URI.**
If a connector must ensure another is running, it needs to call shimp in
process.
Everything in the chat is about the clickable URI.

## The connectors

### chrome sessions, and why not `remote_chrome_cdp`

The whole access story pivots on a browser that is LOGGED IN to the accounts
being reached.
`remote_chrome_cdp` takes the opposite stance, deliberately, and says so:

> A DEDICATED PROFILE, ALWAYS. Launching with the default profile while an
> ordinary Chrome is running does NOT give a debuggable browser ... The cost is
> that this Chrome is signed out ... the honest answer there is to point
> `-chrome-port` at a browser you have signed in yourself.

That dedicated profile is a feature for sweeping public pages and a blocker for
anything behind a login.
Two opposite profile policies cannot live in one tool without one being a
special case, and the special case would be the credential-bearing one, which
is the one that must never silently degrade.

So: a separate connector, whose stance is attach first and never launch a fresh
anonymous profile.

⚠️ **EXTRACT THE CDP CLIENT, DO NOT COPY IT.**
`cdp.go` and `ws.go` are the reusable 42KB.
`main.go`, `manifest.go` and `sweep.go` are the task runner and belong where
they are.
Copying the client gives two implementations of subtle machinery, and
5 October cost three separate incidents in the one area where two tools both
reasoned about Excel sessions.

**Liveness and authority are different questions.**

* is Chrome up on this port: generic, and `/json/version` answers it over
  plain HTTP with no websocket
* is it still signed in, as which account: site-specific, and the only honest
  probe is site-specific too

The second matters because it is the failure that looks like success.
A signed-out browser yields an empty ChatGPT catalog, which reads as "nothing
to capture".
So a connector must be able to refuse loudly when authority is absent rather
than proceeding with an empty answer.

### patchbucket

Split out of `chatgpt_session.py`'s sync, and the seam is not where the names
suggest.
`apply_row_patch` hardcodes `conversations`, `messages`, `scraped_sessions` and
`blobs`, their per-table conflict rules and the FTS5 update.
That is the chatgpt database's merge semantics, not transport.

**patchbucket owns:** the rclone wrapper and its timeout, which exists for a
measured hazard (a wedged rclone that finishes transferring and never exits),
per-machine directories under a remote prefix, `latest.txt` as the head
pointer, sequence numbering, per-peer progress, and download-and-apply in
order while skipping your own machine.

**The owning tool keeps:** `export_rows_since`, `apply_row_patch` and the
`schema_version` gate.

⚠️ **TWO WATERMARKS THAT LOOK ALIKE, AND MUST NOT BE CONFLATED.**
patchbucket owns "how far have I consumed peer X".
The owning tool owns "how far have I exported my own rows", and that one is
read inside a DEFERRED transaction together with the export so that no rows
slip through the gap.
Move it across a process boundary and the atomicity is gone, silently.

Worth more than its first caller: `slate_tools` and `inter_repo_minutes` both
write patches per turn.
Though for anything already committed, git IS the transport and patchbucket
would be redundant there.

### chatgpt capture

`op_full_sync` is already incremental: `clean_stop_pages=3` stops after three
consecutive clean catalog pages.
Its hard dependency is a signed-in Chrome, because it authenticates by running
`fetch()` inside the browser context through `Runtime.evaluate`, so the request
carries the session cookies and a Bearer token from `/api/auth/session`.

So it consumes the chrome sessions connector rather than launching anything,
and it must refuse when `/api/auth/session` yields no token.

`full_update` then reduces to one orchestrating action: pull, sweep, push.
One URI is one capability and one action, so the sequencing lives in code
rather than in the link.

## What is NOT in scope

Taken from the chat and kept: sandboxing, filesystem and network confinement,
auto-update, and any general service manager.
SHIMP guarantees one thing only, that this exact process specification has at
most one managed live instance and that this action was delivered to it.

## Code conventions

One file of generic code per concern, with platform specifics in their own
files behind build tags, and a catch-all so no platform is left with a symbol
undefined.
`pdf_view_server/raise_windows.go` and `raise_other.go` are the pattern.
