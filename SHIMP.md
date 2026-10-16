| item id | item | state | waiting on |
|---|---|---|---|
| s1 | shimp is a connector folder like any other | **DECIDED** 7 Oct 2026 | nothing |
| s2 | Go only: Go does not use Python, Python uses Go | **DECIDED** 7 Oct 2026 | nothing, for a month or more |
| s3 | `runner` must be in the FIRST descriptor ever written | **DECIDED** 7 Oct 2026 | nothing, and it blocks the first descriptor |
| s4 | Declared ports, never allocated | **DECIDED** | nothing |
| s5 | Accept an abbreviated SHA, store the full one | **DECIDED** | nothing |
| s6 | The cold fallback, which SHIMP/1 does not specify | **DECIDED**, and RESTATED 12 Oct 2026 by s51 | nothing |
| s7 | install places a permanent binary, never registers `go run` | **DECIDED** 7 Oct 2026 | nothing |
| s8 | The Office trusted-protocol key, per scheme | **DECIDED** | nothing |
| s9 | The audit is three lines, not two | **DECIDED** 7 Oct 2026, three lines | nothing, and the first descriptor is no longer blocked on this |
| s10 | A long action reports to a visible console, and the click never waits | **DECIDED** 7 Oct 2026 | nothing |
| s11 | Two clicks, not a library, because s15's refusal carries the safety | **DECIDED** 7 Oct 2026, and the library is deferred rather than rejected | nothing |
| s12 | chrome_sessions: liveness built, authority open | **DECIDED** 14 Oct 2026: liveness is the whole job | nothing |
| s13 | Extract the CDP client, do not copy it | **DECIDED** 14 Oct 2026: a leaf module, and the trigger has fired | the probe EXISTING, so extraction happens at second use |
| s14 | patchbucket: transport and ordering, never merging | SPECIFIED | nothing, and it is unblocked |
| s15 | chatgpt_capture needs a SIGNED-IN Chrome, and must refuse without one | SPECIFIED | s12's authority probe |
| s16 | Excel is a session-protocol problem, not a prohibition | PARKED 7 Oct 2026 | a reason to drive Excel through a `shimp://` link |
| s17 | What is NOT in scope for SHIMP/1 | **DECIDED** | nothing |
| s18 | Where this design came from | **REFERENCE** | nothing |
| s19 | The URI grammar, and what identity IS | **DECIDED** in the source (s18), NOT implemented | nothing, except that s21 must close before the first descriptor is written |
| s20 | The keeper, the IPC, and the lifecycle | **DECIDED** in the source (s18), NOT implemented | nothing. These are the spec to build against. |
| s21 | Two more descriptor fields the source froze, and s19's list is short | **DECIDED** 11 Oct 2026: `base` deleted by s29, `policy` dropped | nothing |
| s22 | Windows first, and the spec must not learn the word "pipe" | **DECIDED** in the source (s18), NOT implemented | nothing. It is the build order. |
| s23 | Two audit questions, and only one of them is answerable today | **DECIDED** in the source (s18), NOT implemented | nothing |
| s24 | Four verdicts, because UNVERIFIED is the one that earns the set | **DECIDED** in the source (s18), NOT implemented | nothing |
| s25 | The command surface: nobody hand-builds a capability id | **DECIDED** in the source (s18), NOT implemented | nothing |
| s26 | What a target must carry before it can be a capability | **DECIDED** in the source (s18), NOT implemented, and it has a wrinkle here | nothing |
| s27 | SHIMP is Shim Protocol, and the name collision is known | **DECIDED** in the source (s18) | nothing |
| s28 | Adopting a descriptor: hash the bytes, never trust the id | SUPERSEDED by s35, there is nothing to adopt | nothing |
| s29 | cwd derived from identity rather than declared in it | **DECIDED** 11 Oct 2026: derived, and the directory is SCRATCH | nothing |
| s30 | A singleton per session, and an immutable store per user | **DECIDED** 7 Oct 2026 | nothing |
| s31 | Two link forms, one identity: the short id and the self-describing link | SUPERSEDED by s35, which is SETTLED | a decision on whether a second link form is legal |
| s32 | A link carries a SHORT capability id, because s5's threat model does not transfer | SUPERSEDED by s35, kept for its argument | a decision on how many hex characters |
| s33 | The launch rule: what `runner: "go"` actually constructs | **DECIDED** in the source (s18), NOT implemented, and it was in no row | nothing |
| s34 | The descriptor is BYTES, not an object, so nothing is ever canonicalised | SUPERSEDED by s35, kept for the principle it established | nothing |
| s35 | Drop the descriptor store: the URI IS the canonical form | **BUILDABLE** 12 Oct 2026; grammar frozen and no blockers | nothing |
| s36 | The URI length budget, measured, and what actually eats it | **MEASURED** 12 Oct 2026: at least 8000 bytes survive the OS hop, so length is not a constraint | nothing from the OS hop. The APPLICATION hop still needs a human to click something. |
| s37 | The SHA in the URI: exactly one legal length, and it should be 40 | RAISED 7 Oct 2026, recommending 40 | one measurement: what Go does with an AMBIGUOUS abbreviation |
| s38 | The Excel surface, measured twice: the object refuses, the formula corrupts | **DECIDED** 10 Oct 2026, measured in two repos | nothing. Length is not a constraint on this design. |
| s39 | The minimum SHAPE another tool may refuse a shimp URI on | **DECIDED** 10 Oct 2026, check 6 corrected 11 Oct | nothing, except that check 4's length follows s37 |
| s40 | FIRST CAPABILITY: open an editor at a line of a file | SUPERSEDED IN PART by s58: `vscode:` already does this | the grammar freeze, then the handler |
| s41 | A query is legal AFTER the `#` and banned before it | **DECIDED** 11 Oct 2026 | nothing. It closes one of the four grammar blockers. |
| s42 | Two outbound messages were silently never delivered | RECORDED 11 Oct 2026 | nothing from this repo |
| s43 | The runner is a TABLE, not a parse: `go/run` consumes two segments | **DECIDED** 11 Oct 2026 | nothing |
| s44 | The identity bytes, exactly, with a test vector | **DECIDED** 11 Oct 2026, and it can never be revised | nothing |
| s45 | Argv encoding, and the arguments that are REFUSED | **DECIDED** 11 Oct 2026 | nothing |
| s46 | The TARGET-side IPC client, which s20 requires and s11 declined to provide | **DECIDED** 12 Oct 2026: copied files, synced, with two tests | nothing. The last blocker is closed. |
| s47 | A hashed shimpipc makes the audit cheaper, but only if it is HASHED | **DECIDED** 12 Oct 2026 | nothing |
| s48 | A test installation is a different SCHEME, and that is why it is safe | **DECIDED** and BUILT 12 Oct 2026 | nothing |
| s49 | The one-line install, PROVEN for shimp itself from a bare SHA | **PROVEN** 12 Oct 2026, end to end on a real machine | nothing |
| s50 | NO console field in the descriptor: the console belongs to the keeper | **DECIDED and BUILT** 12 Oct 2026: no field, lazy console, GUI subsystem | nothing |
| s51 | s6 and s10 contradicted each other: the handler never does the work | **DECIDED** 12 Oct 2026, and s6 is restated | nothing |
| s52 | The startup timeout is 120s, and our own measurement rules out any short value | **DECIDED** 12 Oct 2026 | nothing |
| s53 | The named pipe IS the lock, and on Unix it is not | **DECIDED** 12 Oct 2026, with the platform asymmetry reported | nothing |
| s54 | Caveats live in the SOURCE and are checked, because silence must not read as safety | **DECIDED** and BUILT for chrome_sessions, 14 Oct 2026 | rolling it to the other four connectors, and the AGENTS entry |
| s55 | Sharing code across SHAs: the mechanism, the invariant, and what 100% means | **DECIDED and PROVED** 15 Oct 2026: two maxims, and the three untested parts all hold | nothing |
| s56 | NOTHING in the audit payload may be an assertion | **DECIDED** 14 Oct 2026, and four earlier decisions already follow from it | nothing |
| s57 | remote_chrome_cdp is a hand-rolled shimp, and three of its four invented pieces become unnecessary | OBSERVED 14 Oct 2026, no change proposed yet | shimp's dispatch half, before any conversion |
| s58 | `vscode:` is already registered, so s40's first capability already exists | **ANSWERED** 14 Oct 2026 by s61 | nothing |
| s59 | Cached attestations per module@sha: capabilities verified, exposure attributed | **DECIDED** and the first one WRITTEN, 14 Oct 2026 | one per module, as each is written or changed |
| s60 | The leaf candidates, MEASURED, and the divergence nobody was looking for | **MEASURED** and the divergence CLOSED 14 Oct 2026 | nothing |
| s61 | editor_open: the first capability, because NO editor but VS Code has a scheme | DESIGNED 14 Oct 2026, deliberately NOT built | a decision to build it, plus two measurements |
| s62 | linkprobe: fourteen ways in, MEASURED, and two of them cannot carry a shimp link | **BUILT and MEASURED** 15 Oct 2026 | two human routes, which are unanswered rather than passed |
| s63 | The browser route: Chrome keeps the link VERBATIM and refuses to fire it without a gesture | **SUPERSEDED** by s64, 15 Oct 2026. Its conclusion was wrong: an automated browser route exists | nothing |
| s64 | The browser route AUTOMATED, and the security property turned out to be false | **MEASURED** 15 Oct 2026. 13 of 15 routes now run unattended | nothing. Two routes remain human by nature, not by limitation |
| s65 | linkprobe consumable by a test suite, and the duplication check that can see past a rename | **BUILT** 15 Oct 2026 | nothing |
| s66 | The second axis: link SHAPES and the handler's verdict, after a human found what the harness could not | **BUILT and MEASURED** 15 Oct 2026, 8 shapes, 0 findings | nothing |
| s67 | A refused click is VISIBLE: the lazy console, wired at last | **BUILT and CONFIRMED by the user** 15 Oct 2026 | nothing |
| s68 | A hyperlink RAN something. And the consent gate was broken until somebody ran it | **WORKING end to end** 15 Oct 2026, confirmed by the user | nothing |
| s69 | probes/: the instruments promoted, and the first re-run retired a quoted number | **BUILT and RUN** 16 Oct 2026 | nothing |
| s70 | The seven top-level things, and the state of each | **MAPPED** 16 Oct 2026 | nothing |
| s71 | The harness drifted when the handler changed, and the click window is now a choice | **FIXED and MEASURED** 17 Oct 2026 | nothing |
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

#### Proven for the HANDLER too, 12 Oct 2026

This row proved the mechanism with `chrome_sessions`. s49 proves it for `shimp` itself at `e6b4053`, including the `os.Executable()`-under-`go run` path that s7 rests on and that a local `go build` never exercises.

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

#### ⚠️ The mechanism above is contingent, and s34 is why

Added 7 Oct 2026.
"Breaks every link ever written" holds where ids are RE-DERIVED from a parsed object and a schema.
Under s34's proposal, where the descriptor's identity is the sha256 of its bytes as written, an old descriptor's bytes do not change when a later version gains a field, so its id does not change and its links keep working.

The conclusion here is unaffected: put `runner` in the first descriptor, because an explicit field beats an implied one and it costs one line.
But the reason is then tidiness rather than a flag day, and what actually protects an old descriptor is `version` being inside it, which is what `version` is for.

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

#### ⚠️ THE TABLE ABOVE IS WRONG, AND IN THE GENEROUS DIRECTION

Corrected 7 Oct 2026, prompted by the entirely reasonable objection that 9 characters ought to be enough.

The table calls 7 hex "minutes", which implies roughly a million attempts a second.
That is the rate for driving `git commit`.
⚠️ **An attacker does not drive git.** A commit object is `commit <len>\0tree ...\nparent ...\nauthor ...\ncommitter ...\n\n<message>`, and the committer timestamp or trailing whitespace in the message can be varied freely, so the bytes are constructed directly and hashed at raw SHA-1 speed.
Grinding a commit for a vanity prefix is a thing people do for fun.

One current GPU does on the order of 5e10 SHA-1 per second:

| prefix | bits | one GPU |
|---|---|---|
| 7 hex | 28 | about 5 **milliseconds** |
| 9 hex | 36 | about 1.4 **seconds** |
| 12 hex | 48 | about 1.6 hours |
| 40 hex | 160 | not by this route |

So the original table understated 7 hex by around five orders of magnitude, and 9 hex is a second and a half rather than anything to lean on.

#### And the caveat the row never checked: a prefix match still has to be DELIVERED

Grinding the hash is not the attack. The ground object has to be **fetchable at the module path in the descriptor**, because that is what `go run <path>@<abbrev>` resolves the abbreviation against.

* if the attacker can push to that repo, they do not need a collision at all, so the brute-force number is irrelevant
* so is the case the row was implicitly worried about

⚠️ **The route that does NOT need push access is path takeover**: a module path that comes to be served by something other than the intended repo.
A repo renamed, transferred, deleted and its name re-registered, or a proxy answering for that path.
The attacker then owns the path, grinds a commit matching the 9 characters in a descriptor somebody wrote months ago, and the link runs their code.
A full SHA refuses that. Nine characters does not.

#### So the decision is unchanged, on a sounder basis than its own numbers

| where | how much | why |
|---|---|---|
| what a human types | **9 characters, or 7** | s1 PROVED it: `@bdc56ca` resolved and the toolchain expanded it |
| what a descriptor stores | **the full 40** | it costs 31 characters in a file no human reads, and it closes path takeover permanently |

The expansion happens once, at capability creation, performed by the toolchain rather than by a person.
There is no upside to storing nine, which is what decides it: a weak argument beats a zero-cost concession.

The only place the 31 characters could matter is s31's bearer form, where the descriptor travels inside the link. There it is 31 of 255, which does not change whether that form fits.

And `go.sum` plus the checksum database, which s2 records, are what cover the dependency graph below the top-level commit. Neither is affected by abbreviation.

#### ⚠️ AND s35 REMOVES THE STORE THIS RULE DEPENDS ON

7 Oct 2026. "Store the full one" presumes a descriptor to expand into at creation time.
Under s35 the URI is the only artifact, so there is nowhere to put the full SHA and an abbreviation in a link becomes permanent and re-resolved on every click.
s37 carries the question in its new form. This row's input-side generosity survives wherever a tool still accepts typing, but its storage half has no home under s35.

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

#### ⚠️ RESTATED 12 Oct 2026, because it contradicted s10

The sentence above, "the clicking process does the work itself", was inherited from the Python original where that is literally what happened, and the 828ms figure measured exactly that.

It cannot survive s10, which decided the click must never be the thing that waits. s51 resolves it: **the handler never does the work.** It starts a keeper and exits.

What survives of this row is the property rather than the mechanism: **a click never FAILS for want of a live instance, it is only slower.** The singleton is still an optimisation and never a dependency.

⚠️ And what is LOST: this row's claim that the readiness race is "mostly moot". It is not, and s51 records the behaviour that now has to exist, which is the keeper holding an action that arrives before READY.

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

#### Built 12 Oct 2026, and the `os.Executable()` note held up

`install` copies the running binary and registers the copy, exactly as this row describes, and the live cycle confirms Windows routes a click to it.

⚠️ One limit found while building, worth recording beside the measurement above: a binary built by `go run module@sha` can learn its own **pseudo-version** from `debug.ReadBuildInfo`, carrying a 12-hex abbreviation and a timestamp, and **cannot learn its full 40-hex commit**. So `version --audit` reports the pseudo-version. That is what the toolchain actually resolved and what `go.sum` pinned, so it is the honest answer rather than a gap, and s37's 40-hex requirement is about what a LINK carries, where a tool supplies it.

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

#### The rule behind this row's shape, named 14 Oct 2026

s56 states what this row obeys without saying: **nothing in the audit payload may be an assertion.** Both lines here work because each resolves to public bytes, so the sender's honesty is irrelevant. That is also why s54's caveats live in the source rather than in the paste.

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

#### Two decisions that followed from this row, 12 Oct 2026

* **s50**: the console is the KEEPER's and never a descriptor field. The question had to be settled before the first link, because a per-capability console preference would have been a hashed field, and the worry that raised it dissolves once the singleton means one console per INSTANCE rather than one per click.
* **s51**: this row's "the click must never wait" contradicted s6's "the clicking process does the work itself", and the contradiction stood unnamed for five days. The handler now never does the work.

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

#### DECIDED 14 Oct 2026: liveness is the whole job, and this row was right

**`chrome_sessions` answers liveness and nothing else, permanently.** The authority probe is a SEPARATE program.

⚠️ And the reason is this row's own objection, via an invariant neither it nor I had named. I argued on 13 Oct that "two opposite stances cannot live in one tool" covered LAUNCHING and not probing, because a read-only probe decides no policy. That was wrong. `cdp.go`'s header states two things a probe must violate:

| the probe must | `remote_chrome_cdp` refuses |
|---|---|
| attach to an EXISTING tab, to see whether THAT session is signed in | "WE OPEN OUR OWN TAB, ALWAYS", after a measured incident on 2 Oct where attaching drove another program's tab off its page and reported a confident result from the wrong document |
| run a caller-determined query | "THE OP VOCABULARY IS CLOSED, AND `Runtime.evaluate` IS NOT IN IT ... if the wire could carry JavaScript this would be `exec` with extra steps" |

So putting a probe there would not merely enlarge an audit. It would break the invariant standing between that tool and being `exec` with extra steps, in its own words.

The probe is therefore its own program, with its own closed vocabulary of exactly one op, its own JavaScript, and a caveats file (s54) declaring that it attaches to tabs it did not create.

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

#### ⚠️ This now blocks the FIRST capability, not the second consumer

11 Oct 2026. This row says the question is "not urgent until the second consumer exists, which is what s12's authority probe will make it".

That was wrong, and I repeated it twice while listing what blocks building.
s20 requires every target to speak READY and length-prefixed JSON, s11 never provided a target-side library, so the first capability already needs a shared package or a deliberate hand-rolled copy. s46 carries it.

#### A precedent now exists, 12 Oct 2026

s46 faced this same question for the shimp IPC client and chose **copy with a tripwire** over a shared module: one canonical in `shimp/`, byte-identical copies, a per-connector hash test for tampering and a test in `shimp/` that walks siblings for staleness.

That does not decide this row, because a CDP client is 42KB rather than 120 lines and the copy cost scales with size. But it answers the objection this row rests on: copying is only "two implementations" when nothing detects divergence.

#### DECIDED 14 Oct 2026: a leaf module, scoped to transport only

This row said "not urgent until the second consumer exists". ⚠️ **The second consumer now exists and is unavoidable**: s12 closes with the authority probe as a separate program, so there are two programs needing a CDP session.

**`cdpcore`: transport and JSON-RPC only.** Connect, send a method, read a reply, attach to a target. It carries NO op vocabulary.

Each consumer keeps its own closed vocabulary, which preserves the invariant per program rather than centralising it:

* `remote_chrome_cdp`: navigate, show, close. Its own JavaScript. Opens its own tab.
* the probe: one op. Its own JavaScript. Attaches to an existing tab and says so in its caveats.

`cdpcore` being generic is not a weakening. It is a transport, like `net/http`, and nobody faults `net/http` for allowing any URL. The discipline lives in each program and each program's caveats file declares what it actually sends.

s55 carries the mechanism and the three invariants that keep the audit whole.

#### The finding that decided it, from looking at the audit rather than the duplication

Asked to weigh duplication against how easily a model can certify the code, the two axes turned out to AGREE rather than trade off, and the reason is worth keeping:

⚠️ **The cheapest audit is of code that does not exist.** A leaf module and a copy-with-tripwire have IDENTICAL byte counts for an auditor; they differ only in how identity is established, by version string or by hash comparison. What actually reduces audit cost is not arranging 395 lines better but not having them in a second place at all.

That is what made "put the probe in `remote_chrome_cdp`" attractive, and what s12's invariants then ruled out. So: the leaf, because the alternative was foreclosed, not because sharing is free.

#### Trigger corrected, 14 Oct 2026

This row's own wording was "not urgent until the second consumer exists", and on 14 Oct I declared the trigger fired because s12 established the probe must be SEPARATE.

⚠️ **"Must be separate" is not "exists".** The probe is an unwritten program, needed by s15, which is itself unbuilt. So the second consumer is two unwritten programs away, and a shared module with one consumer is strictly worse than no module.

The DECISION above stands: if we share, it is a module, under s55's two maxims. The trigger is restored to **the probe existing**, so that extraction happens at second use rather than in anticipation of it.

And the better reason for waiting: the probe needs connect, attach-to-existing and evaluate, and NOT tab creation, which s12 showed is opposite-stanced. That is perhaps 200 lines rather than 395. Writing it first means deciding what is common with two real implementations in view rather than one and a projection.

#### ⚠️ And measurement says this was not the urgent one. 14 Oct 2026

s60 counted the duplication instead of reasoning about it. `cdpcore` has ONE consumer today. `contained`, the path-containment check, has two, is byte-identical in both, and calls a `foldsCase` that has DIVERGED with a learned lesson stranded in one copy.

So this row's hazard is real and has already happened somewhere else. That strengthens the row's argument and weakens its priority: the extraction worth doing first is not the one that was discussed for a session.

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

#### ⚠️ The hyperlink-length note above is WRONG, corrected 7 Oct 2026

Measured: `Hyperlinks.Add` refuses an `Address` beyond **2084**, so the claim that the Hyperlink object has no cap is false, and the object route is the tighter of the two.
The 255 figure for `HYPERLINK()` is NOT refuted, because what was measured is how much the formula stores (4095) rather than whether a click follows a link longer than 255.
s36 carries the numbers and says which question each one answered.

#### Refined 10 Oct 2026

The correction above said the object "caps at 2084, so the claim that it has no cap is false".
Pinned since: it refuses at 2085 rather than truncating, and 2084 is far above any realistic link.
So this row's original claim is wrong in principle and harmless in practice, and the figure that actually matters is not the size but that **the object refuses loudly where the formula truncates silently**. s38 holds it.

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
⚠️ **And "canonical" is not specified anywhere, which s34 argues is unfixable cheaply and unnecessary**: hash the descriptor's bytes as written and never re-serialise them, so there is no canonical form to agree on.
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
* **and the POSITIVE rule is s33**, because this bullet says what is forbidden and a builder also needs what is built: `runner: "go"` prepends `go run` and the target to the descriptor's argv, verbatim and unquoted.

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

#### The framing was under-specified, and s46 pins it

12 Oct 2026. The table above says "length-prefixed UTF-8 JSON" with no prefix width and no byte order.
That is invisible with one implementation and an interop bug under load with several, which is exactly what s46's copied file would have produced.

Pinned in s46: `uint32` big-endian length, then that many bytes of UTF-8 JSON, 1 MiB maximum, and a declared length above it is REFUSED rather than allocated.
⚠️ Big-endian with a test pinning a known header, for the measured reason in `AGENTS.project.md`: a port read as a plain `uint32` turned 9222 into 30987, which looks like a real port number and so does not announce itself as a bug.

#### Four things this row left unstated, decided 12 Oct 2026

Building the keeper would have hit each of these, and none was on any row:

| | |
|---|---|
| **s52** | the "fixed startup timeout" is **120s**, with progress logged at 5s. s49's 27.4s cold `go run` rules out the 5s default anybody would have reached for, because that breaks the FIRST click of every capability and works thereafter |
| **s53** | the "instance lock" and the named pipe are **the same object** on Windows, which makes this row's lock-lifetime, stale-state and re-check-after-lock properties free rather than implemented. On Unix a socket path persists, so none of it is free there |
| **s51** | the keeper must ACCEPT a connection before READY and HOLD the action, or a cold first click starts the target and silently drops the action it carried |
| **s50** | nothing about a console is ever declared by a capability |

#### ⚠️ NOT the blocker for a one-shot capability, 15 Oct 2026

This row has been described repeatedly as the largest remaining piece and as what stands between a click and anything happening. For a RESIDENT capability that is true. For a one-shot it is not, and the first real use case asked for is two one-shots.

⚠️ `shimp` has NO exec machinery at all. It derives the argv, logs `would run [...]`, and stops. A one-shot click needs only that last step, which is about fifty lines, and none of this row's lock, IPC or handshake.

s67 records the correction. The keeper remains necessary for the resident class and for action delivery to a running instance, which is what it was always for.

#### ⚠️ And a click now RUNS without it, 15 Oct 2026

s68 ran `chrome_ready` from a pasted link. The keeper is still unbuilt and was not needed, because a one-shot needs no singleton and no action delivery.

What remains for this row is exactly what it was always for: the **resident** class, and delivering an action to something already alive. A link carrying an action other than `ensure` is now refused **by name**, which is a better state than being refused generically: the message says the link is valid, names the action, and says the keeper is what is missing.

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

#### DECIDED 11 Oct 2026, and both halves are closed

* **`base` is deleted**, not decided. s29 derives cwd from identity, so there is no relative path left for a base to anchor. s45 carries the symbolic tokens that argv needs instead.
* **`policy` is DROPPED.**

The reason for dropping it rather than including it: s17 already records that SHIMP never enforces a policy field, so it is a claim and never a lock.
Under s35 every link pays for it in length, and under s44 it would be in the hashed bytes, so two capabilities differing only in a claim about themselves would be different capabilities with separate instances.

⚠️ **A field nobody enforces, that every link pays for, and that forks identity, is a poor trade three times over.**

What the field was FOR survives and lands better: s24's audit report states what a target can actually reach, having read the source, rather than repeating what the target claimed about itself. A verified observation beats an unenforced declaration, which was s17's point all along.

⚠️ And s26's conclusion is now the general rule: a descriptor field cannot be "added later", so there are only two honest positions, in the first version or never. `policy` is never.

#### And s50 reached the same answer by the same three objections

12 Oct 2026. s50 asked whether a capability may declare that it wants a
console, and refused for the three reasons written above for `policy`:
unenforced by SHIMP, paid for in the length of every link, and forking
identity.

⚠️ So this row's reasoning is the general rule, and it is broader than
`policy`: **a preference about how a target behaves is never a descriptor
field.** Two capabilities differing only in a claim about themselves, or in a
preference about their own output, would be different capabilities with
separate instances. That is absurd on its face, and it is exactly what a
hashed field means.

(This paragraph was skipped once, on 12 Oct, because the script that wrote it
tested for the phrase "the general rule" and the paragraph above already
contained it. A marker that matches text the row already has is not an
idempotence check, it is a silent no-op.)

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

#### State corrected 12 Oct 2026: DISSOLVED, not open

This row's state said "RAISED, with one rule already clear", and it pointed at s31 as the row that might dissolve it. s31 was itself superseded, and **s35 dissolved this row outright**: there is no descriptor store, so nothing is ever adopted into one.

What survives is the rule, and it moved rather than died: **hash the bytes, never trust an id you were handed.** Under s35 that is no longer an adoption policy, it is how a URI is verified at all, and s44 is where it now lives, as the identity bytes being hashed verbatim after s35's validator has refused every other spelling.

⚠️ Corrected for the same reason as s34: the state is what a reader scans, and s32 proved that a stale one gets implemented.

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

#### DECIDED 11 Oct 2026: derived, and the directory is declared SCRATCH

**cwd is derived from the identity hash and is not in the URI at all.**

```
cwd = ~/.shimp/run/<session>/<identity-hex>/work
```

Under the session scope of s30, because the working directory of a live instance is live state.

So the grammar has no cwd segment, `base` dies with it, and identity is what code, at what commit, with what arguments. Nothing about where.

#### And the contradiction this row raised is answered

This row set s20's "runtime state is disposable and self-healing" against s17's "uninstall removes only what shimp installed", and observed that one of them becomes a data loss if that directory holds a tool's real data.

⚠️ **The directory is declared SCRATCH.** A target must not keep anything there that matters, uninstall may remove it without asking, and a crash may leave it to be recreated.
Both frozen rules stay true, and the cost is pushed where it can be seen: a target needing durable state must be told where to put it, in argv, which is s45's symbolic tokens.

That is the "relocates the path problem into argv" catch this row predicted, now with somewhere to go.

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

#### s35 supersedes the choice here, and the length objection above was pessimistic

7 Oct 2026. The reason is new and it is operational rather than technical: a local store is state that gets lost, and it is empty exactly when a new user clicks their first link.
Also, the "114 characters before a single argument" measured here is 146 of 255 for a complete realistic link, leaving 109 spare. s36 has the real numbers, including the correction that the Hyperlink OBJECT caps at 2084 while this row relied on s17 saying it does not.

#### ⚠️ SUPERSEDED BY s35, confirmed SETTLED 10 Oct 2026

slate_tools asked whether s31 and s37 are a reversal and which wins.
They are, and the later rows win: **s35 settled the direction, self-describing URI and no descriptor store**, and s37 follows from it.
This row is kept for its canonicalisation argument, which s35 answers rather than discards, by constraining the grammar and refusing instead of normalising.
⚠️ What is SETTLED is the direction. What is still open is grammar DETAIL: s37's SHA length, whether an action may carry a query (s40), and whether `policy` dies (s21). Anyone writing Go may rely on the direction and not yet on the exact segment list.

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
* s5's CONCLUSION stands for the target SHA, full in storage and abbreviated on input, but not on the reasoning it gave.
Its brute-force table was wrong by about five orders of magnitude and its threat model never checked delivery. Corrected in s5 on 7 Oct 2026.
The structural point of this row survives that correction and is in fact strengthened by it: the git SHA has a real non-push attack route, path takeover, and the capability id has none.

#### ⚠️ SUPERSEDED BY s35, and a reader has already been misled

10 Oct 2026. s35 dropped the descriptor store, so **there is no capability id in a shimp URI**, and this row's whole subject no longer exists in the design.
It is kept because its ARGUMENT is still the right one and still unused elsewhere: s5's strictness is about code an adversary may publish, and a hash of local bytes has a different threat model.
But the shape it describes is gone.
Measured consequence rather than a tidiness note: slate_tools read this row on 10 Oct and inferred "a capability id plus an action" as the URI shape to validate, which would have refused every valid link. s39 carries the correction.

### s33. The launch rule: what `runner: "go"` actually constructs

Found 7 Oct 2026, from an entirely fair complaint that a summary of this design had hidden the motivating example behind the abstraction.
Checking why, the abstraction had hidden it in the SLATE too.

[23] froze it as "Go invocation: construct `go run <module>@<full-sha> ...` directly".
s19 records the prohibition, that argv is passed as separate arguments and never as a shell command, and no row records the positive rule.
⚠️ **A builder cannot derive what to run from a statement of what not to run**, and s3 makes `runner` the field that selects between launch rules, so what each legal value MEANS has to be written down.

#### The rule, for the one legal value

```
runner  = "go"
target  = "github.com/davidnoz123/nielsoln_connectors/chatgpt_capture@<full-sha>"
argv    = ["-root", ".", "-mode", "sweep"]
```

becomes, as separate arguments with no shell anywhere:

```
argv[0] = "go"
argv[1] = "run"
argv[2] = "github.com/davidnoz123/nielsoln_connectors/chatgpt_capture@<full-sha>"
argv[3:] = the descriptor's argv, in order, unmodified
```

So `runner: "go"` means exactly: prepend `go run`, then the target, then the descriptor's argv verbatim.
Nothing is quoted, joined, split or expanded on the way, and `SHIMP_ENDPOINT` is added to the ENVIRONMENT rather than to argv, per s20's handshake.

#### Why this is the whole motivating example and not a detail

[4] asked for "a single, tiny custom URI tool that installs with `go run` on a remote git sha, and a url of this scheme is an encoding of another `go run` on a remote git sha".
That sentence is the project.
The chain in full:

```
shimp://8f74b19c5d2a3e01/thread/92817
   |                      |
   |                      +-- the action, delivered over IPC, not part of identity
   +-- 16 hex of sha256(descriptor), which resolves locally to:

{ "version":1, "runner":"go",
  "target":"github.com/.../chatgpt_capture@<full-sha>",
  "argv":["-root",".","-mode","sweep"] }

   which this rule launches as:

go run github.com/.../chatgpt_capture@<full-sha> -root . -mode sweep
```

⚠️ **And the `go run` is why identity could never be the executable's path.**
[5] noted the wrinkle: `go run` downloads, builds and then runs a TEMPORARY executable, so a design identifying processes by pathname would be identifying a path that changes per build.
Identity comes from the requested specification instead, which is what makes the whole scheme work rather than being a nicety.
s7 is the same fact biting the installer, where the binary is copied from `os.Executable()` precisely because that path is the toolchain's scratch output.

#### What this rule does NOT permit

* no `sh -c`, no `cmd /c`, no shell at any point, which is s19 and remains the hard rule
* no `go get`, no `go install`, no building into a cache the descriptor does not name
* no argv the descriptor did not contain, except that the environment carries the endpoint
* an unknown `runner` value **fails closed**, exactly as an unknown `version` does, which is s3's reason for the field existing at all

### s34. The descriptor is BYTES, not an object, so nothing is ever canonicalised

Raised 7 Oct 2026, from the question of whether the JSON with `runner: "go"` needs reconsidering.
It does, in two ways, and the second is larger than the first.

#### ⚠️ "Canonical JSON" is an unimplemented decision hiding a hard problem

s19 says canonical JSON and [25] chose JSON over line-based text for good reasons: arrays, escaping and Unicode are already solved in it.
Neither says what canonical MEANS, and it is not one thing:

| the choice | the trap |
|---|---|
| key order | lexicographic by what? RFC 8785 says by UTF-16 code unit, which differs from code-point order above the BMP |
| numbers | `"version": 1` against `1.0` against `1e0`. Same value, different bytes |
| string escaping | `/` against `\/`, `A` against `A`, which control characters are escaped how |
| Unicode | NFC against NFD in a path or an argument. Visually identical, different bytes |
| duplicate keys | permitted by the grammar, meaningless here, and a silent source of disagreement |

There is a standard, RFC 8785, the JSON Canonicalization Scheme.
⚠️ **Go has no JCS in its standard library, and this repo forbids third-party dependencies**, so adopting it means implementing it, including ES6's shortest-round-trip number formatting, which is the fiddly part.
Get any of it subtly wrong and the same capability hashes to two ids, which is the failure a singleton exists to prevent, arriving through the one door nobody is watching.

#### The fix is to delete the problem rather than solve it

**The descriptor is a byte string. Its id is `sha256` of those bytes, verbatim.**

* `cap create` EMITS the bytes, once, in one fixed form
* the bytes are stored and transmitted exactly as emitted, never re-serialised
* it happens to be valid JSON, so a human and a chat window can read it, which is the whole of [25]'s argument and it is kept
* parsing is for READING. Hashing is over the bytes as received
* verification needs no parser at all: hash what you were given, compare to the id

⚠️ **A tool that round-trips a descriptor through parse-and-reserialise has changed the capability**, and that must be forbidden in as many words, because it is the one mistake this design invites and it looks like a no-op.

It also lands well for s31: the bearer link carries the hashed bytes themselves, so there is no canonical form to agree on there either.

The cost, stated: two authors writing the same capability with different whitespace get two ids and two singletons.
In practice `cap create` is the only producer and emits one form, so they collapse anyway.
And arguably it is correct rather than a cost: different bytes were asked for.

#### ⚠️ And this changes what s3 is protecting against

s3 says adding a field later "changes every descriptor's canonical bytes, which changes every capability hash, which changes every capability id, **which breaks every link ever written**".

That is true under canonicalise-from-structure, where ids are RE-DERIVED from a parsed object and a schema.
It is **false** under bytes-as-identity: an existing descriptor's bytes do not change when a later version gains a field, so its id does not change and its links keep working for ever.

So s3's conclusion survives and its mechanism does not.
What actually protects an old descriptor is `version` inside it, which is precisely what `version` is for: a v1 descriptor is read under v1 rules, where `runner` is implicitly `go`.
s3 remains worth doing, because an explicit field beats an implied one and the cost is one line, but it is a tidiness argument rather than the flag-day emergency it claims, and the slate should not keep a scary reason that is contingent on a choice this row may reverse.

#### And `runner: "go"` specifically

Two things, now that the shape is cheap to change rather than frozen by fear:

* **it pins the module graph and not the toolchain.** `go run module@sha` pins what s2 says it pins, through `go.sum` and the checksum database, and says nothing about which Go built it. So "this exact software" is slightly overclaimed, and the honest home for that gap is s24's UNVERIFIED rather than a promise.
* **string or object.** `"runner":"go"` against `"runner":{"kind":"go","toolchain":"1.27"}`. Under bytes-as-identity this is a free choice later, because a v2 descriptor may carry a different shape without disturbing a single v1 id. s26 put toolchain pinning out of SHIMP/1 and that stands, and this row is what makes deferring it genuinely cheap rather than apparently cheap.

#### s35 inverts this row

Direction set 7 Oct 2026: if the descriptor store is dropped and the URI carries everything, there is no JSON to canonicalise and this row's recommendation is replaced rather than adopted.
The principle survives in a different shape: **constrain the grammar so there is exactly one legal spelling, then REFUSE anything else rather than normalise it.**
The asymmetry is the same one that decides this row. A wrong normaliser maps two spellings to one id silently; a wrong validator rejects a good link loudly.

#### State corrected 12 Oct 2026

This row's STATE said "RAISED, and it is the strongest recommendation on this slate" while its own body said s35 had inverted it.

⚠️ **A row whose state contradicts its body is the s32 failure waiting to happen again**: s32 was read by another repo, which derived a URI shape that no longer existed and would have refused every valid link. The body is not what a reader scans. The state is.

What survives, and it is not small: **constrain the format so there is exactly one legal spelling, then REFUSE anything else rather than normalise it.** s35 applies that to the URI, s45 to argv encoding and s44 to the identity bytes, each for the asymmetry this row first named: a wrong normaliser maps two spellings onto one id silently, while a wrong validator rejects a good input loudly.

### s35. Drop the descriptor store: the URI IS the canonical form

Direction set 7 Oct 2026.
Every argument for a separate descriptor assumed the store would simply be there.

⚠️ **A store is state, and state gets forgotten, ignored, lost and neglected.**
It has to survive a backup, a machine move and a reinstall, and it is empty exactly when a new user clicks their first link, which s9 already discovered from the other end.
That reason was on no row and it outweighs the technical tidiness the descriptor was bought with.

#### The move that answers s34's objection

s31 refused the self-describing URI because hashing a URI makes every spelling a different singleton.
That assumed identity is the hash of the URI text **while the grammar stays loose.** Tighten the grammar and the objection evaporates.

⚠️ **REFUSE, NEVER NORMALISE.**
Define the grammar so that exactly one byte string can denote a given capability, then VALIDATE rather than transform:

* **positional segments, never query parameters.** A query is reorderable and its repeated-key order is parser-dependent. A path is ordered by construction.
* **percent-encoding is canonical**: encode only what must be encoded, uppercase hex always, and reject a URI carrying an unnecessary or lowercase escape.
* lowercase SHA, exact length, no trailing slash, NFC required and non-NFC **refused** rather than normalised.

The asymmetry that makes this the right shape: a normaliser that is subtly wrong maps two spellings onto one id **silently**, which is s34's whole fear.
A validator that is subtly wrong rejects a good link **loudly**.
Only the validator is needed, and it is a far smaller thing to get right than RFC 8785.

#### The fragment carries the action

```
shimp://v1/go/github.com/.../chatgpt_capture@<sha>/-root/./-mode/sweep#thread/92817
     |------------------- identity -------------------------------| |-- action --|
```

Identity is everything before `#`. The action is the fragment.
That makes s19's rule, that the action is not part of identity, **syntactically enforced** instead of a convention, and "just ensure it is running" is the URI with no fragment at all.

#### What this deletes or changes

| row | effect |
|---|---|
| s9 | back to **two lines**, which is what was wanted before the store forced a third |
| s28 | gone. There is nothing to adopt |
| s34 | inverts: not "hash the bytes rather than the structure" but "constrain the grammar, validate, never normalise" |
| s25 | `cap create` becomes a URI builder and `inspect` just parses a link. Still useful, no longer load-bearing |
| s29 | survives, and s36 makes it load-bearing for a second reason |
| s21 | `policy` loses its home. Recommended: drop it. s17 says SHIMP never enforces it, so a field nobody enforces and every link pays for in length is a poor trade |

#### Two costs, stated rather than glossed

* **a rigid grammar.** JSON takes a new field gracefully and positional segments do not. But `v1` is in the path, so a v2 grammar is a different first segment, which is cleaner than a schema migration.
* ⚠️ **links become editable.** Somebody will hand-change `-mode/sweep` to `-mode/delete`, and under this form that is a VALID different capability which runs, where a hash would have failed loudly.
Not a security hole, because the URI is the authority either way and s9 audits the link actually clicked. But it changes a link from an opaque token into an editable instruction, and s10's console echoing the resolved command is the mitigation.

#### One thing this direction invalidates, named rather than left to collide

s5 decided "accept an abbreviated SHA, store the full one", and there is no store here to expand into.
s37 carries what replaces it: the grammar must name exactly ONE legal SHA length, because under this direction two spellings of one commit are two identities, and expanding an abbreviation would make identity depend on network state.

#### SETTLED, confirmed 10 Oct 2026 for anyone about to write Go

Asked by slate_tools, which read s31 and s37 as a reversal and wanted to know which stands.

**The DIRECTION is settled**: the URI is self-describing, there is no descriptor store, and identity is the hash of the identity-bearing bytes.
s31 and s32 are superseded and marked as such.

⚠️ **The GRAMMAR is not frozen**, and these are the open details: s37's SHA length, whether an action may carry a query (s40), and whether `policy` dies with the store (s21).
s39 is the subset that is stable across all of them, and it is what another repo may validate against today.

#### GRAMMAR FROZEN 11 Oct 2026, except one architectural question

The five items this row left open are decided:

| was open | decided as |
|---|---|
| runner segmentation | s43, a closed table, `go/run` consumes two segments |
| cwd derived or declared | s29, derived, and the directory is scratch |
| `policy` in or out | s21, dropped |
| the identity bytes | s44, with a test vector |
| argv encoding and refusals | s45 |

So the grammar is:

```
shimp://v1/go/run/<module>@<40 hex>[/<arg>...][#<action>]
```

⚠️ **One thing still blocks building, and it is not grammar**: s46, the target-side IPC client, which s20 requires and s11 declined to provide, and which lands on s13's unanswered question about where a shared package lives.

s37's SHA length stays RAISED and does NOT block: build with 40, because permitting a shorter form later ADDS a legal spelling and breaks nothing, while starting short is unrecoverable.

#### BUILDABLE, 12 Oct 2026

s46 decided, so the last blocker is closed. Grammar frozen, identity bytes pinned with a test vector in s44, the launch rule in s43, the target-side IPC in s46, and the framing width pinned there too.

What remains is not specification: two measurements that the install half produces or that stand alone, being the handler's `os.Args[1]` limit in s36 and the ambiguous-abbreviation question in s37, and s37 does not block because building with 40 is the recoverable direction.

#### ⚠️ A route caught normalising, 15 Oct 2026

s62 measured `explorer.exe` inserting a byte:

```
sent     shimp-linkprobe://x#y
received shimp-linkprobe://x/#y
```

The `/` lands before the `#`, inside the identity-bearing part, so the hash changes. This row's argument was that a wrong normaliser maps two spellings to one id silently. Here an intermediary did the normalising, unasked, on the way in.

It does not bite today because a real link always has a non-empty path. It is worth recording as the first OBSERVED instance of the hazard rather than a reasoned one.

### s36. The URI length budget, measured, and what actually eats it

Measured 7 Oct 2026 in a scratch workbook in this session's Excel, by binary search on the longest URI that round-trips.

| route | longest URI that round-trips | what it means |
|---|---|---|
| `Hyperlinks.Add`, `.Address` | **2084** | a hard refusal beyond it |
| `=HYPERLINK(...)` as a formula | **4095** | the formula STORES this much. See the caveat |
| a plain string in a cell | 32767 | the cell limit, no link involved |

#### ⚠️ AND THIS CORRECTS s17

s17 says `HYPERLINK()` caps `link_location` at 255 characters "while the Hyperlink object does not".

⚠️ **The second half is false.** `Hyperlinks.Add` refuses an `Address` beyond 2084, measured. The object route is in fact the TIGHTER of the two.

⚠️ **And the first half is not refuted by the number above, because the test measured the wrong thing.**
4095 is how much the formula will STORE. Whether a click FOLLOWS a link longer than 255 is a different question and was not tested, which is the claim s17 actually makes.
So 255 stands unrefuted and must be treated as binding until somebody clicks one.
Saying "measured 4095, so 255 is wrong" would be exactly the false confidence the house rule about fakes warns about.

#### The number that actually decides it is NOT measured

What matters on a click is what the registry handler RECEIVES as `os.Args[1]`.
That needs a registered `shimp:` handler which logs the length, and there is none, because s35's grammar is not built.
So: design to **255** and rely on nothing more, and the first thing the handler should do when it exists is report that number.

#### ⚠️ What actually eats the budget is path-shaped arguments

A bare capability is comfortable. Counted:

```
shimp://v1/go/github.com/davidnoz123/nielsoln_connectors/chatgpt_capture@<40 hex>
= 113 characters, and 82 with the 9-character abbreviation s5 permits on input
+ /-root/./-mode/sweep         =  20
+ #thread/92817                =  13
                               = 146 of 255, or 115 abbreviated
```

An earlier framing of mine called 113 "half the budget gone on provenance alone", which was pessimistic: 109 characters remain and most capabilities will not use them.

The real hazard is a path in argv, and it is worse than its character count:

* every `/` inside an argument must be `%2F`, costing **3 characters per separator**
* ⚠️ every `..` must be `%2E%2E`, costing **6 characters for 2**, and NOT for length reasons: a literal `..` in a URI path is subject to RFC 3986 dot-segment removal by any parser or intermediary that touches it, so an un-encoded relative path can arrive **silently different from the one written**

`--db ../chatgpt_tools/chatgpt_1davidnoz_at_googlemail_dot_com.db` is 62 characters plain and grows past 70 encoded, and it is the dangerous shape as well as the long one.

#### So s36 makes s29 load-bearing twice

s29 proposes that cwd is derived from identity rather than declared, which keeps paths out of the URI.
Recorded there as portability and as not leaking a filesystem layout.
**This row adds a third reason: it is what keeps a link inside 255 and away from dot-segment normalisation.**
A symbolic base token is short and contains no separators, which is the shape a URI wants.

#### Superseded in part, 10 Oct 2026: the boundary is pinned and 255 does not bind

s38 carries the pinned numbers and the reconciliation with slate_tools' independent measurement.
Two corrections to this row:

* the object caps at **2084** and **REFUSES** at 2085 rather than truncating, which this row did not establish
* ⚠️ **the formula route is not an alternative at all.** slate_tools measured that a formula cell reads back as its result, so a slate pull stores the display text and the next publish writes it as a value, destroying the link silently. So the 255 figure is not a budget to design against; it is a route not to use.

Which means **length is not a constraint on this design**, and the arithmetic here stands as arithmetic rather than as a limit. The handler-argv measurement this row asks for is still unmeasured and still the only number that could reintroduce one.

#### MEASURED 12 Oct 2026, and length is not a constraint by a wide margin

The handler now exists, so the number this row said was unobtainable has been taken, through `ShellExecuteW` so that Windows does the routing.

| sent | received as `os.Args[1]` |
|---|---|
| 160 | 160, exact |
| 200, 500, 1000, 2000, 2040, 2060 | exact |
| 4000 | 4000, exact |
| 8000 | **8000, exact** |
| 16000 | not delivered |

⚠️ **AND THE 16000 FAILURE IS NOT WINDOWS'.** That probe dispatched through PowerShell's `Start-Process`, which refused with "The command line is too long". That is PowerShell's limit, not the URI dispatch limit, and reporting it as the latter would be attributing one tool's constraint to another. The honest statement is **at least 8000 bytes survive the OS hop exactly, with no truncation**.

Against s38's measured ceilings that settles it: the Excel Hyperlink object refuses at 2084, and a realistic shimp link is under 200 characters. So the OS hop has roughly forty times the headroom of the tightest surface, and **length is not a constraint on this design** by any margin that matters. The remaining ceiling is academic.

⚠️ **What is still UNANSWERED**, and a passing test must not be read as covering it: what Excel or a browser does to a URI BEFORE handing it to the OS. This measured the registry hop. The application hop still needs a human to click something.

#### ⚠️ The open half is now MEASURED, 15 Oct 2026

s62 built `linkprobe` and ran one link down every entry point. This row asked what Excel or a browser does to a URI BEFORE handing it over, and for Excel the answer is now in:

* **the Hyperlink OBJECT route delivers VERBATIM**, which is the route a slate cell uses
* **the FORMULA route cannot be driven at all**: `Hyperlinks.Count` is 0, so there is nothing to follow and only a human click can complete it
* ⚠️ **and the 255-byte truncation did not reproduce**: 326 bytes stored unaltered in the formula. The limit is real somewhere and is not where this row recorded it

The browser half remains open, and is now open BY NAME rather than by omission: `chrome-address-bar` and `chrome-link-click` are separate rows in the table because the href is resolved against the page's base URL first, which the address bar does not do.

#### ⚠️ The 255 number is RETIRED, 16 Oct 2026

s69 promoted the instrument and re-ran it. `=HYPERLINK()` stores 255, 256, 326, 1000, 2084, 2085 and 4000 bytes **unaltered**, and REFUSES with a com_error by 6000.

⚠️ So "truncates at 255 silently" is wrong in both halves: not at 255, and not silently. The boundary is consistent with Excel's 4096-character formula limit and nothing hyperlink-specific.

The 2084/2085 object boundary reproduced EXACTLY, which is the other half of the same run and the reason to trust it.

The formula route remains unusable for the reason that never depended on 255: `Hyperlinks.Count` is zero, so there is no object to follow.

### s37. The SHA in the URI: exactly one legal length, and it should be 40

Raised 7 Oct 2026, and it is a different question from s5 because the ground moved.

#### ⚠️ s35 DELETED THE PLACE s5 PUT THE FULL SHA

s5's answer was "accept an abbreviated SHA, store the full one", and the whole rule rested on there being a descriptor to expand into at creation time.
s35 drops the descriptor store and makes the URI the only artifact.
⚠️ **So there is nowhere left to store the full one, and an abbreviation in a link is permanent.**
That is a DECIDED row invalidated by a direction set later, which is worth saying plainly rather than leaving the two to coexist.

#### The grammar forces one length, whichever it is

Under s35 identity is the hash of the URI bytes, so `@bdc56ca` and `@bdc56ca708bb...` are **two different capability identities for the same code**, and therefore two singletons.

That leaves two ways out and only one of them works:

* **normalise by expanding the abbreviation.** ⚠️ Fatal. Expanding requires asking the module proxy, so identity would depend on network state and on what the repo contains at the moment of the click. The same link could resolve differently on two machines or on two days, which destroys the one property this design has.
* **refuse every length but one.** Which is s35's refuse-never-normalise rule applied here, and it is the answer. The grammar names a single legal SHA length and rejects anything else.

So the question is not "may a link be abbreviated". It is "what is the one legal length".

#### And the reason for 40 is now stronger than s5's, and different

s5 argued brute force, and s5's corrected numbers still stand: 9 hex is about 1.4 seconds on one GPU, 12 hex about 1.6 hours, and grinding only pays if the attacker can serve that module path at all.

⚠️ **The new reason has nothing to do with attackers.**
Under s35 the abbreviation is re-resolved **by the toolchain, at click time, on every click, for the life of the link.**
Under the old design the full SHA was frozen into a descriptor once and never resolved again.

So a commit pushed to that repo next year, which happens to share nine hex characters with a link written today, changes what that link means.
No adversary is required. The repo simply grew.
A full SHA is the only form whose meaning cannot change, which is the whole of what a link is supposed to buy.

#### The one measurement that could soften this

s1 proved an abbreviation RESOLVES: `@bdc56ca` became `v0.0.0-20261006210611-bdc56ca708bb`.
⚠️ **It did not test what happens when an abbreviation is AMBIGUOUS.**

* if the toolchain REFUSES an ambiguous abbreviation, the failure is loud and an abbreviated link degrades to a broken link, which is survivable and would make a shorter length arguable
* if it picks one, the failure is silent and an abbreviated link is indefensible at any length

Nothing on this slate establishes which, and the difference decides how bad abbreviation is rather than whether 40 is safest.
It is the same shape of experiment as s1 and s7 and it needs a repo holding two commits that share a prefix.

#### ⚠️ If links must be shorter, the SHA is the wrong target

Counted, from s36:

| part | characters |
|---|---|
| `github.com/davidnoz123/nielsoln_connectors/chatgpt_capture` | **58** |
| the full SHA | 40 |
| `shimp://v1/go/` and the `@` | 15 |

The module path is HALF AGAIN the size of the SHA and it is the compressible part: a shorter repo name, a connector at the repo root, or an alias resolved by the handler would all save more than abbreviating the SHA.
Attacking the SHA saves 31 characters of a 255 budget that a realistic link uses 146 of, and it spends the only guarantee the design has to pay for them.

So: 40, and if length ever genuinely bites, shorten the path.

### s38. The Excel surface, measured twice: the object refuses, the formula corrupts

slate_tools measured this on 8 Oct and sent it in `TO_NIELSOLN_CONNECTORS.md`.
This session had measured the same surface on 7 Oct and recorded a contradictory number, so both were pinned exactly on 10 Oct.

| route | cap | at the boundary |
|---|---|---|
| the Hyperlink **object** | **2084** | **REFUSES**, COM error, nothing stored |
| `=HYPERLINK()` as a formula | **255** | **TRUNCATES SILENTLY** |

Pinned, asked against stored:

```
2083 -> 2083  exact
2084 -> 2084  exact
2085 -> REFUSED, nothing stored
```

#### The two measurements were never in conflict

slate_tools found "no cap" and this session found 2084.
Both are right: a search that stops below 2085 finds no cap, because there is none to find until then.
⚠️ **And the number was never the point.** The object refuses loudly and the formula truncates quietly, and that difference decides the design where 2084 against "none" does not.

A realistic shimp link is under 200 characters, so **length is not a constraint on this design.**
s36's budget arithmetic stands as arithmetic and stops being a limit.

#### ⚠️ `=HYPERLINK()` is a trap, not an alternative route

Measured in slate_tools, and it is worse than the silent truncation:

A formula cell reads back as its **result**.
So a slate `pull` stores the DISPLAY TEXT, and the next `publish` writes that text as a plain value, and the link is gone with nothing reporting it.
That is how 33 live hyperlinks were lost on 2 Oct.

So shimp links belong in a column declared `set_format(..., hyperlink=True)`, which uses the object, and **never in a formula**.
⚠️ No shimp documentation may show `=HYPERLINK("shimp://...")`, because that example destroys itself on the next publish.

#### ⚠️ One of this row's two reasons was false, 16 Oct 2026

This row recorded that `=HYPERLINK()` must never appear in shimp documentation. The rule STANDS, and one of its supports has gone.

* ⚠️ **"truncates at 255 silently" is FALSE.** Re-measured by `probes/excel_hyperlink.py`: 4000 bytes stored unaltered, refused by 6000.
* **"a formula cell reads back as its result, so publish destroys the link" is TRUE** and was always the stronger half.
* **And `Hyperlinks.Count` is ZERO afterwards**, measured 15 Oct, so nothing can follow one programmatically either.

Two good reasons remain and the rule is unaffected. ⚠️ Recorded because the false one was sent to `slate_tools` and `dataannotation_tools`, and both notes now carry a correction.

### s39. The minimum SHAPE another tool may refuse a shimp URI on

Asked by slate_tools, which will accept `shimp://` in `link_target` as a one-branch replacement for `slate://`, under a standing rule of **shape only, never liveness**: no network call, no registry lookup, no filesystem check, because "is this a URL" is local and deterministic while "does this handler exist" is slow and fails for unrelated reasons.

```
shimp://<version>/<runner...>/<module>@<sha>[/<arg>...][#<action>]
```

| # | check | why this one |
|---|---|---|
| 1 | scheme is exactly `shimp://`, lowercase | |
| 2 | the first segment matches `^v[0-9]+$` | it is what lets this grammar change without their code changing. An unknown version is OURS to refuse at click time |
| 3 | exactly one `@` before any `#` | |
| 4 | the text from `@` to the next `/`, `#` or end matches `^[0-9a-f]{40}$` | the check worth having: it catches an abbreviated SHA, uppercase hex, a branch name pasted in and a truncated paste, which are the four things a human typing a cell actually gets wrong |
| 5 | at least three non-empty segments before the `@` segment, and a non-empty module part before the `@` | |
| 6 | no `?` anywhere | the grammar is positional; a query is reorderable and its repeated-key order is parser-dependent |
| 7 | at most one `#` | |
| 8 | no empty segment, and no trailing `/` before `#` or end | |

⚠️ **Check 4's length is s37's and s37 is RAISED rather than DECIDED**, so they were asked to make 40 a named constant. It waits on one measurement, what the toolchain does with an AMBIGUOUS abbreviation. Nothing else in the shape would move.

#### One check deliberately NOT delegated

The grammar also requires canonical percent-encoding: uppercase hex, and no unnecessary escapes.
Uppercase-hex is cheap and they may have it.
⚠️ **"No unnecessary escapes" is not cheap**, because it means knowing what the author was obliged to encode, and asking another repo to implement it would make it the second implementation of this grammar.
That belongs in the handler's own validator, which is s35, and the house rule about two implementations of subtle machinery is s13's measured reason.

#### ⚠️ And the correction they needed first

Their note inferred the shape from s32 and s37 as "a capability id plus an action".
True on 7 Oct, false now: s35 dropped the store, so **there is no capability id in a shimp URI at all**.
A tool implementing a capability-id shape would refuse every valid link.
Marked on s32 so the next reader of that row cannot repeat it.

#### ⚠️ Check 6 was WRONG and has been corrected, 11 Oct 2026

As shipped, check 6 said "no `?` anywhere". It would have refused s40's first capability, whose action is `#open?file=...&line=123`.

**Corrected: no `?` BEFORE the `#`, permitted after it.** s41 carries the reasoning, which is that the fragment is not hashed so a query there cannot fork an identity.

⚠️ And the correction travelled through a channel that had been withdrawn, so neither the original nor the fix arrived by that route. See s42. Both are now files in this repo.

### s40. FIRST CAPABILITY: open an editor at a line of a file

Proposed by slate_tools and accepted 10 Oct 2026.

```
shimp://v1/<runner>/<editor-target>@<sha>/<args>#open?file=...&line=123
```

The shape of the action is not settled: `#open?file=...&line=123` carries a query in the FRAGMENT, which s35's no-query rule does not cover because that rule is about the identity part. Whether an action may carry a query at all is a grammar question this row inherits.

#### Why this one first

Tiny, used every day, no authentication, no Chrome and no Excel, so it exercises the handler, the singleton, the IPC and the launch rule without any of the hard dependencies.
And it has a live requester, which beats a clean design nobody is waiting for.

#### The route it replaces, and why that one is dead

The session in `dataannotation_tools/raccoon` is building this as an Excel right-click VBA handler.
Three measured reasons from slate_tools:

* the slate workbook is `sx_<repo>.xlsx`, and an `.xlsx` cannot hold VBA. Saving strips it, so every slate workbook in every repo would have to become `.xlsm`
* ⚠️ `publish` DELETES the worksheet and adds a new one, so every publish hands the sheet a **new CodeName**. A handler bound to the old one silently stops firing and looks identical to a working one
* it needs Trust Center's "Trust access to the VBA project object model", per machine, which `excel_tools` already carries a refusal message for

The second is the dangerous one and it is this repo's recurring shape: it appears to work, then stops, and nothing reports it.

#### ⚠️ What was told to raccoon, stated honestly

Written to `dataannotation_tools` on 10 Oct, needs-action, as `617bc6d2fdc5-034e87e9f030`.
Including the part that matters most: **shimp is designed and NOT built.** There is no `shimp/` folder, no handler and no registered scheme.
So they were told not to block on it, and given slate_tools' own interim answer: a `hyperlink=True` column opens the file today and loses the line number.
Promising a date here would have been the one thing worse than the VBA route.

#### The action's shape is settled, 11 Oct 2026

This row said "the shape of the action is not settled" because `#open?file=...&line=123` carries a query in the fragment.
s41 settles it: a query is legal after the `#` and banned before it, because the fragment is not part of the hashed identity.

So this capability's link needs no exception and the grammar needs no special case.

#### ⚠️ SUPERSEDED IN PART, 14 Oct 2026: this already exists

s58 measured `vscode:` as registered on this machine, so `vscode://file/<path>:<line>` opens an editor at a line today with no shimp at all.

So the capability this row accepted is solved, and raccoon's need is met by a link in a `hyperlink=True` column plus s8's trusted-protocol key for `vscode:`.

What survives: this row's reasoning about what makes a good FIRST capability, being tiny, daily, no auth, no Chrome and no Excel, with a live requester. What does not: that it should be an editor opener. s58 argues the first capability must be something `vscode:` cannot do, which means pinning and audit rather than navigation.

### s41. A query is legal AFTER the `#` and banned before it

Decided 11 Oct 2026, resolving a contradiction between two rows written a day apart.

#### ⚠️ The contradiction, and it had already been sent to another repo

s39 check 6, given to slate_tools as a shape rule: **"no `?` anywhere"**.
s40's first capability, the editor opener: `#open?file=...&line=123`.

So the check this repo asked slate_tools to implement would have refused the first capability's own links.
s40 noted the tension in its own text; the message did not.

#### The rule, and why the ban was never about `?`

s35 bans a query in the identity part for a precise reason: a query is reorderable and its repeated-key order is parser-dependent, so **it cannot appear in the bytes that are hashed**, or two spellings of one capability become two singletons.

⚠️ **The fragment is not hashed.** s19 and s35 put the action outside identity, and s35 made that syntactic by using the fragment for it.
So reordering an action's query cannot fork an identity, and the objection simply does not reach past the `#`.

That makes this a positive argument rather than a concession: the fragment is the one place in the URI where a query is safe, *because* of what the fragment is.

```
no `?` before the `#`      the identity part, hashed, must have one spelling
`?` permitted after it     the action, not hashed, forwarded as an opaque string
```

RFC 3986 is content with it: the fragment is everything after the first `#`, and `?` is a legal fragment character, so `#open?file=x&line=1735` parses without special handling.

#### What follows

* **the target owns action spelling.** Two spellings of one action both reach the running instance, and what it makes of them is its business, per s20's rule that an action is an opaque message.
* s39's check 6 is corrected to "no `?` before the `#`" and re-sent.
* this closes one of the four decisions s35 left open.

### s42. Two outbound messages were silently never delivered

Recorded 11 Oct 2026, because it is the exact failure shape this repo keeps naming and this time this repo was the victim rather than the finder.

On 10 Oct two messages were sent through `inter_repo_minutes`:

| id | to | what |
|---|---|---|
| `617bc6d2fdc5-034e87e9f030` | dataannotation_tools | needs-action **STOP**: you are mid-build on a route that cannot work |
| `617bc6d2fdc5-7778661e9352` | slate_tools | the URI shape they were blocked on |

Both were marked **`withdrawn` at 2026-10-10T23:45** when the IRM deployment was withdrawn, and **neither carries a `seen` mark.**

⚠️ **Nothing reported the non-delivery, at either end.**
The sending command said "sent". The withdrawal said nothing to the sender. It was found a day later only by querying `irm.db` directly, because somebody thought to check.

#### Why this is a row and not a grumble

The STOP was the urgent one. A session was building an Excel right-click VBA handler that cannot work for three measured reasons, and the message telling them so was retracted without telling anybody. The cost of a silent non-delivery is not the message; it is that **the sender stops looking for another way to reach them.**

That is `chrome_sessions`' lesson in a different costume, s12: a browser answering on a port is not a browser that is signed in, and the failure that matters is the one that looks like success.

#### What was done instead

`TO_DATAANNOTATION_TOOLS.md` and `TO_SLATE_TOOLS.md`, in this repo, re-delivering both.
That is the convention `slate_tools` itself used to reach this repo: a committed file in the sender's repo, named for the recipient, with a human carrying the pointer.

⚠️ **Deliberately NOT written into `AGENTS.project.md`.**
That file now says the IRM check is gone and says not to re-add a convention by copying a paragraph, because five repos once held four versions of the same one. Writing a `TO_<REPO>.md` protocol in there would be that mistake with a new name. A file is a file until the generated replacement exists.

#### The requirement this hands to whatever replaces it

Not a better `--thread`, and not a way to clear an obligation. Both of those were already fixed.

⚠️ **A withdrawal, or any other non-delivery, must be visible to the SENDER.**
A channel that can silently fail to deliver a STOP is worse than no channel at all.

### s43. The runner is a TABLE, not a parse: `go/run` consumes two segments

Decided 11 Oct 2026. It was in no row, and it is why a `run` segment went missing from s35's example without anybody noticing.

#### The problem

```
shimp://v1/go/run/github.com/davidnoz123/...@<sha>/...
            ^^^^^^ ^^^^^^^^^^
            runner  module
```

⚠️ **A left-to-right parser cannot tell where the runner ends**, because a module path has slashes too. Nothing in the text marks the boundary.

#### The decision: a closed table, keyed on the first segment after the version

| runner | segments it consumes | argv it contributes |
|---|---|---|
| `go/run` | 2 | `go`, `run` |

Exactly one entry today.
⚠️ **An unrecognised runner FAILS CLOSED**, which is the rule already chosen for `version` in s19 and the reason s3 insisted the field exist before the first link.

So the parse is: take the version, look up the next segment in the table, consume that runner's segment count, and everything from there to the segment containing `@` is the module.

#### Why a table rather than one segment

The alternative was `go+run` as a single segment, which parses with no table at all.
Rejected because the table is needed regardless: s33's launch rule has to map a runner onto literal argv, and that mapping IS the table. A `+` separator would buy a simpler parser and still leave the table to be written, so it saves nothing and reads worse.

#### And this is the positive form of s33

s33 said `runner: "go"` means "prepend `go run`". With the runner in the URI the rule collapses into the table above: **the runner's segments ARE the leading argv**, literally, in order. No implied verb, no prepend rule, nothing to get wrong.

### s44. The identity bytes, exactly, with a test vector

Decided 11 Oct 2026.
Every row said identity is "the hash of the identity-bearing bytes" and none said which bytes.

⚠️ **This is the one decision with no migration path at all.** A version segment lets the grammar change; nothing lets the hashing rule change, because every id that ever existed was computed under it.

#### The rule

```
identity_bytes = the URI from the first byte of "shimp://" up to and
                 including the byte before the first "#", verbatim,
                 as received, with no transformation whatsoever
identity       = sha256(identity_bytes)
```

Spelled out, because each clause was a real alternative:

* **the scheme IS included.** `shimp://` is part of the bytes. A future scheme would be a different protocol, not a different spelling of this one.
* **the version IS included.** `v1` is in the hash, so a v2 link can never collide with a v1 link even if every other segment matches.
* **up to the byte before the first `#`**, so the `#` itself is excluded and the fragment is excluded. s41 is what makes this safe: the action is not hashed, so a query there cannot fork an identity.
* **no `#` at all means the whole URI is the identity bytes.** That is the "ensure it is running, send nothing" form.
* ⚠️ **verbatim, as received.** No case folding, no percent-decoding, no NFC pass, no trailing-slash removal. s35's refuse-never-normalise rule is what makes this correct: anything non-canonical was already REFUSED by the validator, so by the time bytes are hashed there is exactly one legal spelling and nothing left to normalise.

#### The test vector, so an implementation can prove itself

```
identity_bytes (126 bytes):
shimp://v1/go/run/github.com/davidnoz123/nielsoln_connectors/editor_open@1173327f58213898a33698949e49c5aa6d9a83e2/-editor/code

sha256:
7564ba535f3a1d65f96c219b29112e5195b4cd7c3c4f9c007e3b84fc5e34d9bc
```

That is a real sha256 of those exact bytes, computed 10 Oct 2026.
A first implementation that does not reproduce it has the rule wrong, and this is the cheapest possible way to find that out.

#### What the hash names

The lock, the IPC endpoint and the derived working directory of s29. All internal, so the full 64 hex characters are used there and length costs nothing. Only a LINK is ever abbreviated, which is s37.

### s45. Argv encoding, and the arguments that are REFUSED

Decided 11 Oct 2026. Each argv item is one path segment, in order, between the target segment and the `#`.

#### What must be encoded

| in an argument | becomes | why |
|---|---|---|
| `/` | `%2F` | it would otherwise start a new segment, so a new argument |
| `#` | `%23` | it would otherwise start the fragment |
| `?` | `%3F` | s41 bans a query before the `#` |
| a space | `%20` | |
| `%` | `%25` | or a literal `%2F` in an argument could not be distinguished from an encoded `/` |
| anything outside unreserved `A-Za-z0-9-._~` | its `%XX` form | |

Uppercase hex always, and **only what must be encoded**: an unnecessary escape is refused rather than accepted, per s35.

⚠️ **And `..` must be `%2E%2E`.** Not for length: a literal `..` in a URI path is subject to RFC 3986 dot-segment removal by any parser or intermediary that touches it, so an un-encoded relative path can arrive **silently different from the one written**. s36 measured the length cost; this is the correctness cost and it is the larger one.

#### What is REFUSED

* ⚠️ **an empty-string argument.** A path segment cannot be empty, because s39 check 8 bans empty segments so that `//` is never ambiguous.
So a capability needing `-flag ""` cannot be expressed, and that is a refusal rather than a gap: it is stated here so the first person who needs one gets an error naming the reason instead of a link that silently drops an argument.
* an argument that is only dot-segments, for the same RFC 3986 reason.
* a lowercase or unnecessary escape anywhere, per s35.

#### Symbolic expansion, which is s29's relocated problem

s29 moves paths out of cwd and into argv, so argv needs the symbolic tokens [25] proposed.

A **closed set**, with unknown symbols failing closed: `$HOME` and `$CAPDIR` today.

⚠️ **Hashed literally, expanded at LAUNCH.**
The identity bytes contain `$HOME`, never the expansion, which is what keeps one capability one identity across machines. s44's verbatim rule and this are the same rule seen twice.

And the honest cost, which belongs in s24's report rather than in a promise: **an auditor reading the link cannot say which directory will be touched**, only which symbol was named. That is a genuine loss against an absolute path, traded for portability and for not publishing a filesystem layout in every link.

### s46. The TARGET-side IPC client, which s20 requires and s11 declined to provide

Raised 11 Oct 2026.
⚠️ **I dismissed s13 twice as "chrome_sessions, not shimp". That was wrong, and this row is why.**

#### Two rows that do not meet

* **s20 requires every target to speak the protocol.** The keeper "launches the target with the endpoint in its environment, and waits for a `READY` message before forwarding any action". So a target must read `SHIMP_ENDPOINT`, connect, send `READY`, then read length-prefixed UTF-8 JSON and reply `{"ok":true}` or `{"ok":false,"error":"..."}`.
* **s11 decided "no Go API for now"** and that decision was about the CALLER side, one capability ensuring another without a human. It never addressed the target side at all.

So as specified, **every target hand-rolls READY and the framing.**

#### Why that is not survivable past the first one

For one target it is fine. For three it is precisely s13's hazard, which is measured rather than feared: two implementations of subtle machinery in one repo cost three separate incidents in one evening on 5 Oct, all in the one area where two tools both reasoned about the same thing.

And framing is exactly that kind of machinery. A length prefix read one byte short fails intermittently and under load.

#### Which lands it on s13's open question

A shared package needs somewhere to live, and every connector is its own module today, so a target importing `shimp/ipc` takes a cross-module dependency inside one repo. That is s13's question, unanswered, and it now blocks the FIRST capability rather than the second consumer of a CDP client.

The options, none chosen:

* a `shimp/ipc` package in the shimp module, imported by targets. One implementation, and every target gains a dependency on the handler's own module, which is the trust root.
* a separate tiny module, `nielsoln_connectors/shimpipc`, depended on by both. More modules, cleanest dependency direction.
* ⚠️ **no library, and the keeper adapts instead.** [25]'s consequence, recorded on s11: with keeper-owned IPC the keeper could translate an action into whatever mechanism a target already speaks, so a target needs no SHIMP code at all. That is the only option that keeps targets dependency-free, and it moves the complexity into the keeper where there is exactly one copy of it.
* hand-roll the first one deliberately, with the duplication recorded as a debt rather than discovered later.

#### Why this is the only architectural question left

Items 1 to 5 of the pre-build list are decided as s43, s29, s21, s44 and s45. This one is not a grammar detail; it decides whether a shimp capability is a normal program or a program that links shimp.

#### A fifth option, recorded because it was PROPOSED AND WITHDRAWN

11 Oct 2026. Asked whether this had been decided as "a small file in each connector folder", which it had not. It was proposed and then withdrawn by its proposer a turn later, and neither half was on this row until now.

**Option E: one source file, copied, with a pinned-hash drift test.**
A single `shimpipc.go` vendored by copy into each target, plus a test asserting its sha256 against a recorded constant.

The argument for it: s13's objection to copying is "two implementations of subtle machinery", and a copy WITH a tripwire is not two implementations, it is one implementation in N places with a loud alarm when they diverge. It is also this repo's own idiom, since `AGENTS.project.md` already requires that any new syscall struct gets a test pinning a known value, for this exact class of problem. It keeps empty `require` blocks, keeps the trust root uncoupled and keeps stdout free.

⚠️ **And then the evidence below made it unnecessary.** Recorded anyway, because a withdrawn option with its reason is worth more than an option list that looks exhaustive.

#### The evidence that changed the recommendation

Asked why a target needs to know anything about shimp at all. The answer is that it is **only** the listener protocol: identity, hashing, the lock, the registry, the URI and the singleton are all invisible to a target, and its entire obligation is to be reachable, so that a SECOND click can arrive at the instance already running.

Then `pdf_view_server` was read rather than assumed, and it already has **two** action channels:

| channel | where |
|---|---|
| `POST /select`, token-guarded | `main.go:132` |
| a watched `.nav` file, polled every 100ms | `main.go:656` |

Neither needs a line of shimp code.

⚠️ **And s4 is what makes the keeper able to use them, which neither row said.**
"Declared ports, never allocated" exists so identity is computable before the process runs. Its second consequence is that the port is in the capability's argv, argv is in the URI, so **the keeper knows where to send an action by reading the link.** No discovery, no registry, nothing to ask the target.

There is a matching answer for `READY`: for a port-serving target, liveness IS the handshake, by polling the port until it answers, which is what `chrome_sessions` already does with `/json/version`. s12's warning that liveness is not authority does not bite, because "is it up" is all the keeper needs.

#### So the question is smaller than this row first stated

| class of target | needs | example |
|---|---|---|
| **one-shot** | nothing. No resident process, so no channel. Every click is s6's cold path at ~650ms | `chrome_sessions` |
| **already reachable** | nothing IN THE TARGET. One adapter per mechanism, one copy, in the keeper | `pdf_view_server`, `remote_chrome_cdp` |
| **resident with no channel of its own** | the listener code | ⚠️ possibly empty for a long time |

#### ⚠️ A correction to this row's own reasoning

When option C was first weighed, it was claimed that s10's visible console kills it, because stdout would be the reply channel.

**That was too broad.** It kills the stdin/stdout VARIANT of adapting, where stdout and the console are the same file descriptor. It does not kill keeper-adapts generally: an adapter that POSTs to a declared port or writes a nav file leaves stdout entirely free for s10's `[full_update] step 1/3`.

#### The recommendation, not yet a decision

**The keeper adapts, with the declared port as the primary adapter, and no target-side library until a target appears with no channel of its own.**

It also lands best against the audit story: a capability stays a normal program that can be `go run` standalone and understood without understanding shimp, which is what s9's second line is for.

The residual cost, stated: the keeper carries an adapter per mechanism and that set is open-ended. But it is one copy in one place rather than N copies in N targets, which is s13's distinction exactly.

#### And one spec gap found while describing this row

⚠️ s20's framing row says only **"length-prefixed UTF-8 JSON"**, with no prefix width and no byte order.
With one implementation that is invisible. With several hand-rolled ones it is an interop bug that appears under load. So the framing needs pinning whichever option wins, and that is a one-line decision nobody has taken.

#### DECIDED 12 Oct 2026: the canonical lives in `shimp/` and is COPIED

Not a shared module, and not the keeper adapting for everyone. Copies, synced, with tests.

#### ⚠️ The enabler, checked rather than assumed

**Every connector in this repo is `package main`.** Verified 12 Oct across `chrome_sessions`, `pdf_view_server`, `remote_ai` and `remote_chrome_cdp`.

So a copy is **byte-identical** with no edit, which is what lets a test be a plain hash equality rather than a fuzzy comparison. Had any connector used a different package name, this option would not work at all.

#### Three files, not one, and the house rule is why

`net.Dial` cannot open a Windows named pipe, and this repo forbids third-party packages. So the platform split that `owner_windows.go` / `owner_other.go` and `raise_windows.go` / `raise_other.go` already use:

| file | what |
|---|---|
| `shimpipc.go` | generic: read `SHIMP_ENDPOINT`, send `READY`, the read and reply loop. No build tag |
| `shimpipc_windows.go` | `CreateFileW` on `\\.\pipe\...`. A pipe client is a file open, and `owner_windows.go` is the precedent for syscall rather than shelling out |
| `shimpipc_other.go` | `net.Dial("unix", ...)`. ⚠️ The catch-all is NOT optional, and under s22's Windows-first it may legitimately say what it cannot do, as `owner_other.go` returns a REASON rather than an empty list |

Around 120 lines in total, and connectors keep **empty `require` blocks**, which is the whole advantage over a module.

#### And the framing is pinned here, closing s20's gap

s20 said only "length-prefixed UTF-8 JSON", with no width and no byte order. Invisible with one implementation, an interop bug under load with several. So:

```
frame = uint32 big-endian length, then that many bytes of UTF-8 JSON
max   = 1 MiB; a larger declared length is REFUSED, not allocated
```

⚠️ **Big-endian, and a test pins a known header.** `AGENTS.project.md` records the measured reason: a port read as a plain `uint32` from `MIB_TCPROW_OWNER_PID` turned 9222 into 30987, which looks like a real port number and so does not announce itself as a bug. The rule there is that any new syscall struct gets a test pinning a known value, and a frame header is the same hazard.

#### ⚠️ Two tests, because they catch different failures

| test | catches | runs |
|---|---|---|
| `shimpipc_test.go`, copied into each connector | **this copy was tampered with** | anywhere. It travels with the module |
| a test in `shimp/` that walks the sibling folders | **this copy is STALE** | only in a checkout, and it skips gracefully when siblings are absent |

The distinction matters and is easy to miss: if the canonical changes and connector A is synced while B is not, **B's own test still passes**, because B's copy matches B's recorded hash. It is internally consistent and out of date. Only a check that knows the canonical can see that.

Putting the staleness check in `shimp/`'s own suite means it runs under the `go test ./...` that `AGENTS.project.md` already mandates, so nobody has to remember a command.

#### ⚠️ The circularity trap, avoided deliberately

An expected hash stored INSIDE `shimpipc.go` can never match, because hashing the file includes the constant.

So the expected hashes live in the copied `shimpipc_test.go`, which hashes the three sources. `sync` rewrites all four files together.

#### `shimpIPCFrom`, which s47 needs

```go
const shimpIPCFrom = "<the handler commit this copy was synced from>"
```

Written by `sync`, covered by the hash test, and not circular because the expected hashes are in the test file. s47 is why it exists: without it, an auditor comparing a target's copy against the canonical cannot tell **stale** from **tampered**.

#### `sync`

```
go run github.com/davidnoz123/nielsoln_connectors/shimp@<sha> sync
```

from the repo root, rewriting all four files in each connector that declares it wants them. Development-time only by design: a connector fetched standalone has no `shimp/` to copy from, and does not need one, because the TEST is what travels.

#### What was NOT chosen, and what survives of it

* a `shimp/ipc` package imported by targets: rejected, it couples every target's audit to the trust root's module graph, which s9 keeps as a separate line on purpose
* a separate `shimpipc` module: rejected for now, it ends the empty-`require` property for a benefit this buys more cheaply
* **the keeper adapting survives and is still right where it applies.** A target that already has a channel needs none of this, and `pdf_view_server` has two, `POST /select` at `main.go:132` and a watched nav file at `main.go:656`. The declared port is in argv and argv is in the link, per s4, so the keeper knows where to send without asking anybody. This decision serves the third class only: resident, with no channel of its own.

#### The cost, stated

N copies to update, and a drift window between editing the canonical and running `sync`. The sibling test closes that window the moment anyone runs shimp's own suite, which is the smallest safeguard that does not depend on memory.

#### One consumer may leave the IPC question entirely, 14 Oct 2026

s57 observes that if `remote_chrome_cdp` stops being a server and becomes a one-shot target, it is this row's FIRST class: no keeper, no IPC, no `shimpipc`. So the copied-file decision would have one fewer consumer than assumed.

### s47. A hashed shimpipc makes the audit cheaper, but only if it is HASHED

Decided 12 Oct 2026, out of s46.

#### The property, and it is cheaper than it looks

**s9's line 1 pins the handler's commit, and `shimp/` is where the canonical `shimpipc.go` lives.**

So an auditor doing the handler half of the audit, which they must do anyway because the handler is the trust root, reads the canonical **as part of that**. Every target's copy then verifies for free: hash the file in the tree already fetched at the target's SHA, and compare.

One read, N recognitions. The handler audit and the target audit share a component, and line 1 is what makes that verifiable rather than assumed.

#### ⚠️ And copies are MORE auditable than a shared module, which inverts the intuition

A module dependency is resolved at build time, so verifying it means reading `go.mod`, reading `go.sum` and resolving the graph.
A copied file is sitting in the tree the auditor already pulled. Nothing to resolve.

It also gives a tamper signal that a module would not give as directly: a target shipping a MODIFIED IPC layer stands out immediately, because its hash does not match.

#### ⚠️ THE TRAP, and it is the one that matters

An auditor who learns to skip "the known IPC file" **by filename** rather than by hash has been handed an attack. A hostile `shimpipc.go` under the right name passes a careless audit and looks exactly like a careful one.

That is the permissive-fake failure from the house rules in a new costume: something that resembles verification and is not. `AGENTS.md` puts it as a bare `Mock` that answers any attribute staying green while the code reaches for surface nobody modelled.

**So the published audit text must say HASH IT, never "this is the standard file".** The efficiency is strictly conditional on the hash, and a shimp document that phrases it as recognition rather than computation has created the hole.

#### ⚠️ Stale against tampered, which is what `shimpIPCFrom` is for

The handler SHA and the target SHA are **independent**. s9's line 1 pins handler@X; the URI pins target@Y. Y's copy was synced from some handler commit, not necessarily X.

So a bare mismatch means **stale or tampered**, and an auditor cannot tell which. That turns a crisp signal into a shrug, and a shrug gets ignored, which is worse than having no signal.

s46 therefore puts the provenance in the copy:

```go
const shimpIPCFrom = "<handler commit this copy came from>"
```

The auditor fetches THAT commit for the canonical and gets a clean verdict: the copy matches its stated origin, or it does not.

⚠️ And deliberately NOT a published list of known-good hashes. That would be state that has to be maintained and that goes stale, which is the objection that killed the descriptor store in s35. The copy carries its own provenance instead.

#### How much this actually saves, honestly

Modest per target, and it compounds across targets.

`shimpipc.go` is the **boring** part. What an auditor genuinely has to assess is what a target does with its arguments: files read and written, network, subprocesses. Skipping 120 lines of framing is a real saving and not the bulk of the work.

And it does not touch either of the audit's standing unverifiables, which stay in s24's UNVERIFIED: `module@sha` does not pin the Go **toolchain**, per s34, and the dependency graph needs separate work, per s26.

#### Refined 14 Oct 2026: this objection does not reach a per-version attestation

This row refused a published list of known-good hashes, as state that has to be maintained and goes stale.

⚠️ The objection is to a **latest-good** list. s59 caches a statement about an **immutable subject**, `module@sha`, and a statement about a fixed thing cannot go stale. What goes stale is "which version is current", which an attestation does not record.

So the distinction worth keeping is not "no lists" but **"no list of what is current"**. s59 also notes what this row's `sum.golang.org` comparison implies: a published hash list DOES work when it is append-only and maintained by somebody else, which is what that checksum database is.

### s48. A test installation is a different SCHEME, and that is why it is safe

Decided and built 12 Oct 2026, because `install` could not be exercised at all without it.

⚠️ **There is exactly one `HKCU\Software\Classes\shimp` per user.** So before this, a test install WAS the production install: one home, one registry class, one Office key. There was no way to run an installer test without putting a real handler on the machine.

#### The mechanism, and s44 already made it correct

The scheme becomes a variable, set by `install -scheme`, recorded in the stamp, and recovered by a running handler from the stamp beside its OWN binary.

| | production | test |
|---|---|---|
| home | `~/.shimp` | `~/.shimp-selftest` |
| class | `...\Classes\shimp` | `...\Classes\shimp-selftest` |
| Office | `...\All Applications\shimp:` | `...\All Applications\shimp-selftest:` |

⚠️ **And s44 makes a second scheme a correctness property rather than a convenience.** The identity bytes include the scheme deliberately, so a `shimp-selftest://` link hashes differently and addresses a different singleton. A test click can never be answered by a production instance. That decision was taken for an unrelated reason and happens to buy exactly this.

The scheme name is VALIDATED rather than trusted, because it becomes a registry key path: lower-case letters, digits and hyphens only. A scheme containing a backslash would claim a key nobody intended.

#### REFUSES rather than defaults

The self-test names the production scheme and **refuses** it, before touching anything.
A default can be overridden by a careless invocation, and on the other side of a careless invocation is somebody's working handler.
It also refuses to run if the test scheme's keys or home ALREADY exist, rather than adopting state it did not create.

#### `uninstall -purge`

Plain `uninstall` keeps the log, which is right for production: a log is the record of what happened, not runtime state, and deleting it at the moment somebody is diagnosing an install is the wrong kind of tidy.
`-purge` removes the home entirely, because a test cycle that leaves anything behind is not a test, it is an accumulation.

#### What the live cycle proves, and it is driven as a user drives it

```
go build  ->  install -scheme  ->  verify  ->  ShellExecuteW(uri)
          ->  read the log     ->  uninstall -purge  ->  assert nothing remains
```

⚠️ **It builds a real binary rather than installing itself.** The first version called `cmdInstall` in process, which copies `os.Executable()`, and under `go test` that is the TEST binary. Windows would then have invoked a test binary as the handler, the click would never have reached `onClick`, and the test would have reported on something it never exercised.

And the click goes through **`ShellExecuteW`**, so WINDOWS does the routing, through the registry, exactly as a click does. Not a simulation.

### s49. The one-line install, PROVEN for shimp itself from a bare SHA

Measured 12 Oct 2026, from `/tmp` with no checkout, after pushing `e6b4053`:

```
go run github.com/davidnoz123/nielsoln_connectors/shimp@e6b4053 install -scheme shimp-selftest

  go: downloading .../shimp v0.0.0-20261012011532-e6b4053c698e
  go: downloading .../nielsoln_connectors v0.0.0-20261012011532-e6b4053c698e
  installed shimp-selftest:
    binary  <home>\shimp-selftest.exe
    version v0.0.0-20261012011532-e6b4053c698e
    sha256  162b97f5a438c0831cc30cf04b47852987bb5b9cfb3dabd7b5f3a4cbe92372ee
```

Then `verify` passed against the live registry, a real click routed through the OS to the installed handler, and `uninstall` withdrew every claim.

#### What this proves that s1 did not

s1 proved `go run module@sha` works for a SUBDIRECTORY module, using `chrome_sessions`. This proves it for the handler, and three things s1 could not:

* **the `os.Executable()` path s7 rests on.** Under `go run` that is the toolchain's temporary build output, and s7 asserts it is self-contained so copying it is enough. A local `go build` does NOT exercise this: its output is an ordinary binary in an ordinary place. Now measured: the copy runs, and the OS invokes it successfully.
* **`version --audit` with a real version.** Every previous run reported `(devel)`. The real block is:

```
SHIMP/1
module=github.com/davidnoz123/nielsoln_connectors/shimp
version=v0.0.0-20261012011532-e6b4053c698e
binary_sha256=162b97f5a438c0831cc30cf04b47852987bb5b9cfb3dabd7b5f3a4cbe92372ee
go=go1.27.0
```

That answers s23's second question, "is what is installed the thing I think it is", with real values for the first time. The version carries the 12-hex `e6b4053c698e` and a timestamp, which is the limit s7 records: a go-run binary can learn its pseudo-version and never its full 40-hex commit.
* **the 7-hex abbreviation on input**, which is s5, expanded by the toolchain itself.

#### ⚠️ And the timings make s7's case STRONGER than s7 does

s7 measured a local module: `go run` 650ms warm against a built binary at 68ms. Measured here for a REMOTE `module@sha`:

| | |
|---|---|
| `go run module@sha`, cold, first ever | **27.4 s** (download and build) |
| `go run module@sha`, warm | **1152 ms, 1628 ms, 4843 ms** |
| the installed binary | **36 ms, 37 ms, 91 ms** |

⚠️ **A remote `module@sha` is far slower warm than s7's local figure**, because resolution still happens on every run. So the gap is roughly 30x rather than 10x, and s7's conclusion that install must place a permanent binary is not merely right, it is understated.

#### ⚠️ One defect this found that the live test structurally could not

`uninstall -purge` invoked as the INSTALLED binary cannot delete itself: Windows refuses to remove a running executable. Measured, "Access is denied".

The live test could never catch it, because that test drives an externally built binary and so purges a file it is not running from. **A harness hides exactly the case it does not model**, which is the house rule about fakes arriving from an unexpected direction.

Not fixed by a reboot flag or a detached helper. It is REPORTED precisely, naming the file and the reason, every registry claim is withdrawn regardless, and the next `install` overwrites the binary. `TestPurgeCannotDeleteItselfAndSaysSo` now models the case.

### s50. NO console field in the descriptor: the console belongs to the keeper

Decided 12 Oct 2026, and it was the only remaining decision that could never be taken later.

#### Why it had to be decided now

s10 decided that a long action reports to a VISIBLE console, and `AGENTS.md` forbids hidden windows for a measured reason. But a console is right for a three-minute sweep and looks wrong for a 75ms editor-open.

⚠️ **If that is a per-capability property then it is a DESCRIPTOR FIELD**, and s34 and s44 leave only two honest positions for one: in the first version, or never. "Optional, added later" is not available, because adding a field changes every identity. So this was s3's problem with a new name, and it was on no row.

#### The decision: no field

**A capability does not declare anything about a console.** The console is the keeper's, and the three classes of s46 settle what each case gets:

| class | console |
|---|---|
| **one-shot** | none. It runs, prints nothing anybody is waiting for, and exits |
| **resident** | the keeper creates one, the target inherits it, and it lives as long as the instance |
| **keeper-adapts** | none. The target already has its own output, as `pdf_view_server` logs to its own file |

#### ⚠️ And the worry that forced the question dissolves

"A console flashing on every click" is wrong, and noticing why is what makes the field unnecessary: **the singleton means the console is created ONCE per live instance.** The first click of a resident capability opens a window. The second click, at 75ms, writes into the window that is already there.

So the cost is one window per instance per session, which is exactly what s10 asked for, and it is not per click at all.

#### And the three objections that killed `policy` apply unchanged

s21 dropped `policy` for being unenforced, for costing link length in every link, and for forking identity. A console bit fails all three in the same way:

* ⚠️ **it would fork identity.** Two capabilities differing only in a preference about their own output would be different capabilities, with different ids and separate instances. That is absurd on its face and it is what a hashed field means.
* every link would pay for it in length, per s36's budget.
* and SHIMP would be carrying a declaration it does not need to interpret, which is s17's whole boundary.

s21's reasoning was written for `policy` and turns out to be the general rule: **a preference about how a target behaves is never a field.**

#### MEASURED 12 Oct 2026: Windows creates it, on every click

Asked who currently creates the command prompt, and whether the keeper could start without one and allocate it later.

The first half was unknown and is now measured, by having a click record it:

| invoked by | `GetConsoleWindow() != 0` |
|---|---|
| **a click, through the OS** | **true** |
| a Git Bash shell, which is a MinTTY pty rather than a Win32 console | false |

⚠️ **So Windows gives the handler a console, and a window flashes on every click today.** `main.go` carried a comment asserting the opposite, that "a handler invoked from a URI click has NO CONSOLE", which was the reason the log exists at all. The log's justification survives, because a flashed console that closes in 36ms is no use to anybody, but the stated fact was wrong.

#### And yes: a console can be allocated LATE

`AllocConsole` gives a console to a process that has none, at any time. So the keeper can start silent and allocate only when it turns out to have something to show.

⚠️ **One trap, and it is this repo's favourite shape:** `AllocConsole` does NOT redirect the existing `os.Stdout` and `os.Stderr`. A window appears, the program writes to its old handles, and the window stays empty. A console that looks like a working one. Opening `CONOUT$` and rebinding is what makes it show anything, and `console_windows.go` does that with the reason recorded.

#### ⚠️ Which makes the NO-FIELD decision stronger, not weaker

The original argument against a `console` field was that it would fork identity, cost link length and be unenforced, which is s21's rule.

Lazy allocation removes the last reason anybody would want the field: **a capability does not need to predict whether it will have something to say, because the keeper can observe whether it did.** A declared preference is a prediction; output is a fact. Preferring the fact is what s24 does with UNVERIFIED and what s21 does with `policy`.

#### The design that follows

* the **keeper** is launched with no console and calls `AllocConsole` the first time the target produces output a human should see, or when s51's anomaly needs reporting
* the **target** inherits the keeper's console once it exists
* a **one-shot** capability never gets one, per s46's three classes

#### ⚠️ And the one real complication, recorded rather than solved

Removing the handler's own flash entirely means the binary must be a **GUI-subsystem** executable, so Windows never creates a console for it, with `AttachConsole(ATTACH_PARENT_PROCESS)` when it is run from a terminal instead of clicked.

⚠️ **That is a BUILD-TIME property, and shimp's install copies whatever binary built it.** So the subsystem is decided by the flags on the documented install line, which would become `go run -ldflags "-H=windowsgui" <module>@<sha> install`. The one-line install story and the no-flash property are therefore coupled, which nothing had noticed.

The cheaper option is to stay console-subsystem and call `FreeConsole` as the first act of a click, which hides the window immediately rather than never creating it. A sub-frame flash in most cases, and not guaranteed.

Not decided here. The flash is 36ms today and the choice is between a short flash with a simple install line and no flash with a longer one.

#### CLOSED 12 Oct 2026: GUI subsystem, measured both ways

The row had three possible outcomes and said which decision each would imply. Measured, three clicks, consistently:

| build | console on a click | CLI output |
|---|---|---|
| console-subsystem, the default | **present, VISIBLE, and in the FOREGROUND** | works |
| GUI-subsystem, `-ldflags "-H=windowsgui"` | **absent entirely** | **still works, including through `cmd.exe`** |

⚠️ **So it was the worst of the three outcomes: the window takes FOREGROUND.** Not a flicker. Every click interrupts whatever was being typed, which is precisely the category s10 exists for and the reason that row was raised at all.

And the GUI build has no cost on the other side. The worry was that a GUI-subsystem binary loses its CLI output and would need `AttachConsole(ATTACH_PARENT_PROCESS)`. Measured: output appears from both Git Bash and `cmd.exe`, because inherited handles work regardless of subsystem. ⚠️ One caveat not tested: `cmd.exe` does not WAIT for a GUI-subsystem process, so an exit code in a batch file is unreliable. Output ordering looked correct here; scripting around it is unverified.

#### So the documented install line carries the flag

```
go run -ldflags "-H=windowsgui" \
    github.com/davidnoz123/nielsoln_connectors/shimp@<sha> install
```

#### ⚠️ And the condition this row set for accepting that is now met

The objection to the flag was never its length. It was that **forgetting it degrades silently**: the install succeeds, the scheme works, and focus is stolen on every click for ever, with nothing to read.

So the binary reads its own PE header, and the omission is made loud twice:

* **`install` WARNS** at the moment it matters, naming the flag, because somebody who omitted it has no reason to run `verify`. A warning rather than a refusal, so the plain line still works for somebody who does not care.
* **`verify` reports FAIL**, not a note, because the install did succeed and this is the only thing that would ever tell them.

`subsystem.go` reads Subsystem at PE + 24 + 68. ⚠️ Worth stating because it looks like it should differ: that offset is the same for PE32 and PE32+, since PE32+ drops `BaseOfData` and widens `ImageBase`, and the two changes cancel exactly.

⚠️ **And the offset is pinned by a test that builds a binary BOTH ways**, per the house rule about syscall structs. The values being read are 2 and 3, and small integers are everywhere in a PE header, so a wrong offset could return a plausible answer indefinitely. Nothing but building both ways distinguishes right from lucky.

#### One defect the new check exposed in an existing test

`TestVerifyFailsWhenTheBinaryChangedUnderneathIt` asserted `verify`'s EXIT CODE, so it broke the moment an unrelated check was added, which it duly did: a `go test` binary is console-subsystem, so the subsystem check correctly fails and the aggregate is non-zero.

⚠️ The real defect was that `verify`'s output could not be captured, so a test could only ever assert the aggregate and never which check failed. `AGENTS.md` names that exactly: where there is no seam, the absence is usually the design problem the test was about to find. `verify` now writes through an `out` seam and the test asserts on the `binary sha256` line, which is its actual subject.

#### ⚠️ The lazy console is WIRED, 15 Oct 2026

This row decided the console belongs to the keeper rather than to a capability, and `console_windows.go`'s header answered the sharper question put back to it: a handler can start with no console and allocate one only when it has something to show.

⚠️ `allocConsole` was then written and **never called by anything**, until a human clicked eight links and saw an empty flash.

s67 wires it, and the asymmetry strengthens this row rather than weakening it: a refusal gets a window because it has something to say, a success gets none because it does not, and nothing has to be declared in advance.

### s51. s6 and s10 contradicted each other: the handler never does the work

Decided 12 Oct 2026.

#### The contradiction, which was on no row

* **s6**: "If the singleton cannot be reached, **the clicking process does the work itself**."
* **s10**: "**the click must never be the thing that waits.** The action returns on ACCEPTANCE, carrying where it reports, never on completion."

⚠️ **Those cannot both hold literally**, and both rows existed for five days without either naming it.

s6's mechanism was inherited from the Python original, `slate_tools/slate_open.py`, where the clicking process genuinely did the work and the 828ms cold figure measured exactly that. Under shimp the handler is a 36ms binary and the keeper is a separate process, so the premise changed and the sentence did not.

#### The decision: the handler NEVER does the work

A handler resolves the link, computes the identity, finds no live keeper, **starts one, and exits.** It never runs a target, never waits for one and never holds a console.

So s6 is restated. What it claimed:

> the clicking process does the work itself

What survives, and it is the part that mattered:

> **a click never FAILS for want of a live instance. It is only slower.**

The singleton remains an optimisation rather than a dependency, which is s6's real content. The cold path is slow because the TARGET has to start, which s49 measured at 650ms to 27.4s for a `go run`, and not because anything waits in the click.

#### ⚠️ And one thing s6 claimed is genuinely LOST

s6 said the startup-handshake race is "mostly moot" because not being ready is survivable. Under this decision it is not moot: an action can arrive while a keeper is starting and READY has not come.

So the keeper owes one behaviour s20 never stated: **it accepts a connection before READY and HOLDS the action until READY arrives or the timeout expires.** Without that, the first click of a cold capability starts the target and silently drops the action it was carrying, which is the failure shape of this whole repo.

That is the real cost of resolving the contradiction, and it is an addition to s20 rather than a detail.

#### ⚠️ THREE outcomes, not two. Added 12 Oct 2026 from the sheet

Put back: a click should not be silent when there is an ANOMALY in the configuration, and a singleton that cannot be accessed is an anomaly.

That is right, and it names a distinction both this row and s6 had merged. "The singleton cannot be reached" was doing duty for two completely different situations:

| | what it means | how it should behave |
|---|---|---|
| **no keeper exists** | the normal first click of a capability | slow, logged, progress shown. NOT an anomaly |
| **a keeper exists and answers** | the normal warm click | fast, silent |
| ⚠️ **a keeper exists and cannot be reached** | something is wrong | **loud and visible** |

#### s53 is what makes them separable

Under pipe-as-lock the three cases are distinguished by the kernel rather than inferred:

* `CreateNamedPipeW` **succeeds** -> nobody held it -> I am the keeper -> cold start
* it **fails** and connecting **succeeds** -> I lost the race or a keeper was already live -> forward the action
* ⚠️ it **fails** and connecting **fails or times out** -> a keeper holds the endpoint and will not answer -> **anomaly**

The third case cannot be recovered from by force: the pipe's owner cannot be evicted. So it is REPORTED, visibly, which is the first real use for s50's lazy console, because an anomaly is precisely the output a human needs to see. And it names what it can: the identity, the endpoint and the pid if the platform will say.

This is the same rule the rest of the repo keeps restating. s12: a browser answering on a port is not a browser that is signed in. s8: `verify` says the key is present, not that clicks are silent. A click that quietly carries on after finding a wedged singleton would be the version of that failure that this design cannot afford, because the user's conclusion would be "the link does not work" with nothing to read.

### s52. The startup timeout is 120s, and our own measurement rules out any short value

Decided 12 Oct 2026.

s20 says "a fixed startup timeout, with a clear failure when readiness never arrives" and names no number. The number was on no row, and the two facts that decide it sat in two rows that never met.

#### ⚠️ The measurement that rules out the obvious choice

s49 measured the target side of a cold launch:

| | |
|---|---|
| `go run module@sha`, cold, first ever | **27.4 s** (download, then build) |
| `go run module@sha`, warm | 1.1 s to 4.8 s |

⚠️ **So any timeout under about 30 seconds fails the FIRST click of every new capability and succeeds on every click after.**

That is the worst failure shape available: a link that is broken once and then works, which teaches a user that clicking twice is the fix and teaches nobody that anything is wrong. And 5 seconds is what anybody writing this would reach for.

#### The decision

**120 seconds**, with progress LOGGED at 5 seconds and every 15 after.

The headroom is deliberate. 27.4s was one small module on a warm network; a larger target, a cold module cache or a slow proxy can be multiples of it, and the cost of a generous timeout is a slow failure while the cost of a tight one is a broken first click.

⚠️ **And the progress logging is not decoration.** A click that takes 30 seconds with nothing reported is indistinguishable from a click that did nothing, which is the exact failure `AGENTS.md` is organised against. The keeper logs what it is waiting for, so a slow first click is visible rather than mysterious.

#### What this is NOT

Not a timeout on the action, and not a timeout on the target's lifetime. Only on READY. A target that takes three minutes to do its job is s10's case and reports to a console; this is about a target that never becomes reachable at all.

#### ⚠️ What the number is FOR changed, 12 Oct 2026

Put back: this depends on whether anything VISIBLE is happening, a console showing a heartbeat for instance.

That is right and it reframes the row. 120s was chosen because a silent 30-second wait is indistinguishable from a click that did nothing, so the timeout was carrying the whole burden of not being mysterious. With visible progress it stops being that and becomes a backstop.

#### And the heartbeat does not have to be invented

⚠️ **`go run` already says what it is doing.** s49 captured it:

```
go: downloading github.com/.../shimp v0.0.0-20261012011532-e6b4053c698e
go: downloading github.com/.../nielsoln_connectors v0.0.0-...
```

The keeper launches the target and can capture its stderr, so the console can show the toolchain's own output. That is real progress, naming the module being fetched, rather than a spinner that proves only that a process is alive.

#### So the rule, restated

**Never wait silently.** Concretely:

| at | what happens |
|---|---|
| ~2 s | the keeper allocates a console (s50) and starts echoing the target's output |
| throughout | `go: downloading ...` and anything else the target says |
| 120 s | the backstop. A clear failure naming what never arrived |

The 120s stays, because the cost of a generous backstop is a slow failure while the cost of a tight one is s52's original finding: a link broken on its first click and working thereafter. But it is no longer the thing that keeps a slow click honest. The console is.

### s53. The named pipe IS the lock, and on Unix it is not

Decided 12 Oct 2026.

s20 says "the keeper owns the instance lock" and, separately, that the transport is a named pipe on Windows. It never asks whether those are the same object. [5] had suggested a named mutex, and s30 quoted that only to settle `Local\` against `Global\`.

#### The decision: one primitive, not two

**On Windows the named pipe's existence IS the instance lock.** `CreateNamedPipeW` with `FILE_FLAG_FIRST_PIPE_INSTANCE` fails for a second creator of the same name, so creating the endpoint and taking the lock are one indivisible act.

Three of s20's own properties stop needing an implementation:

* **"the lock's lifetime IS the instance's lifetime"** becomes true by construction. The handle dies with the process, so there is no pid-and-start-time dance and no pid reuse to avoid.
* **"stale state is detected and recreated"** becomes free. There is no stale pipe to detect: a dead keeper's pipe does not exist.
* **"concurrent clicks serialise around the identity, and re-check after taking the lock"** becomes one atomic operation rather than a check, a lock and a re-check. The loser of the race gets a failure from `CreateNamedPipeW` and then simply connects as a client.

⚠️ **And two primitives can disagree, which is the real argument.** A mutex held while the pipe is gone, or a pipe alive while the mutex was released, are both states that a mutex-plus-pipe design can reach and neither has an obvious correct response. One object cannot contradict itself.

#### ⚠️ THE ASYMMETRY, reported rather than hidden

**A Unix domain socket's path PERSISTS after the process dies.** So none of the three properties above is free there: a stale socket file is left behind, and the keeper must try to connect, treat a refusal as proof the owner is gone, unlink and recreate.

That is a real difference in the amount of code per platform, and it is the shape s30 already recorded for session scoping: Windows gives something free that Unix must implement, and the honest move is to say so rather than to pretend the design is symmetric.

s22's rule still holds and is what keeps this out of the spec: the protocol says **the keeper owns the endpoint and only one keeper per identity can exist**. That the Windows implementation gets the second half for free is an implementation note, and `scheme_other.go`'s pattern is where the Unix version says what it has to do instead.

#### The requirement, which is a better way to put it. 12 Oct 2026

Put back: the question was unclear, and surely the driving consideration is robust singleton creation with no racing.

That is exactly it, and it is a better framing than the one this row opened with. The row posed a choice between primitives when the requirement decides it:

⚠️ **`CreateNamedPipeW` with `FILE_FLAG_FIRST_PIPE_INSTANCE` is ATOMIC.** Either this process created the endpoint or somebody else holds it, decided by the kernel, with **no window between checking and acting**.

A mutex plus a pipe has that window: take the mutex, then create the pipe, and a crash between them leaves a held lock with no endpoint. It also has two objects that can disagree, and neither disagreement has an obvious correct response.

So there was nothing to weigh. One atomic primitive satisfies "no racing" by construction; two primitives cannot, however carefully they are ordered. The rest of this row follows from that rather than arguing towards it.

And s51 now rests on the same atomicity: it is the pipe result, succeed or fail, plus the connect result that separates a normal cold start from a lost race from a wedged keeper.

### s54. Caveats live in the SOURCE and are checked, because silence must not read as safety

Decided 14 Oct 2026, from the question of where an auditor's caveats should live.

#### Why the source and not the paste

s56 is the general rule, and this is its first application: **a caveat pasted alongside a link is an assertion** and so unverifiable. A caveat in the source at a pinned commit is self-verifying, because a sender cannot swap it and it cannot drift from the code it describes, being in the same commit.

#### ⚠️ And a declaration nobody checks is worse than none

An auditor finding no caveats cannot tell "this reaches for nothing" from "nobody wrote any", so silence reads as safety. That is the failure shape the whole repo is organised against.

So `caveats.go` declares, and `caveats_test.go` walks the package with `go/ast` and asserts every outward reach of a known kind is declared. A new `exec.Command` nobody wrote down breaks the build.

#### ⚠️ THE PROOF FOUND THREE DEFECTS, ALL IN THE CHECKER

Worth recording in full, because each is a different way for a safety tool to be comfortable and useless.

1. **It missed the program's PRIMARY reach.** The HTTP request is `client.Get(url)`, a method on a value, and an AST-only walk cannot resolve what `client` is. So `http.Get` sat in the surface list, the code never calls it, and the checker passed while blind to the thing the program exists to do. Fixed by matching TYPE references as well as calls: `&http.Client{...}` is a selector the walk can see.
2. **The declaration asserted the checker was blind to the syscall.** It is not: `iphlpapi` is a package-level var, so `iphlpapi.NewProc` is an ordinary selector. Wrong claim, corrected in place rather than dropped.
3. ⚠️ **FATAL: the check was per-FILE.** It asked whether the file was mentioned anywhere in the declaration, and `sessions.go` was already named for two other reaches, so a deliberately planted `exec.Command` in that file PASSED. Every file gets mentioned eventually, so the checker was on its way to always passing while staying green. Each caveat line now opens with its surface in brackets and the check matches on that.

Proven by planting the reach and watching it fail, then reverting:

```
UNDECLARED REACH: exec.Command at [sessions.go:323]
```

#### What it turned up about chrome_sessions

```
[http.Client]      HTTP GET to 127.0.0.1:<port>/json/version and /json/list
[net.DialTimeout]  a bare TCP dial, to tell a dead port from a wedged one
[iphlpapi.NewProc] reads the WHOLE machine's TCP listener table, including the
                   owning pid of every listening socket, not only the ports asked about
```

⚠️ The third is the point. A tool advertised as "which Chromes can be driven" reads system-wide information about every process holding a socket. No auditor would guess it from the description, and now none has to.

#### The limit, stated in the file rather than implied

`CaveatSurfaces` names exactly what the checker can see. A raw syscall, reflection, or any helper outside that list is **UNVERIFIED rather than absent**, and the iphlpapi line is declared because a human noticed.

#### The signpost

It belongs in the HANDLER's source, at the SHA in s9's line 1, which an auditor reads anyway: "read `Caveats` in the target, verify it against the source, and **if there is no `Caveats`, report UNVERIFIED rather than SAFE**." Living there rather than in the paste means it cannot be softened for a particular link.

#### The capability half of s59, 14 Oct 2026

What this row verifies is exactly the **facts** half of an attestation. s59 adds the exposure judgement, which is the part that needs a human and is therefore the part worth caching per `module@sha`.

⚠️ And it is the division that makes a self-issued attestation worth anything: the capabilities are mechanically checked, so only the judgement is a claim.

#### A leaf candidate in its own right, 14 Oct 2026

s60 ranks this second of the measured candidates. The AST checker is pure stdlib and pure mechanism, has one consumer and five planned, and rolling it out to the other four connectors is what will CREATE the duplication.

⚠️ So the ordering matters: extract it before copying it four times, or `repocheck` will report four identical copies of a checker whose whole purpose is catching undeclared duplication.

### s55. Sharing code across SHAs: the mechanism, the invariant, and what 100% means

Decided 14 Oct 2026.

#### It works natively, and the install line is untouched

A connector's `go.mod` can require a sibling module at a specific commit, and `go run A@shaX` resolves and fetches `B@shaY`:

```
require github.com/davidnoz123/nielsoln_connectors/cdpcore v0.0.0-20261013084211-4b91c7e8f2a0
```

Nobody types the pseudo-version; `go get <mod>@<sha>` derives it. **The user still types one line and the toolchain resolves the graph**, which is the property that had to survive.

#### ⚠️ THE INVARIANT THAT PRESERVES THE AUDIT, and it is three rules

MVS is only ambiguous when there are two paths to one module. So:

1. **every shared module is a LEAF**: empty `require`, depends on nothing
2. **the graph is at most ONE level deep**: a connector requires leaves, never another connector
3. **no `replace` and no `exclude`, ever.** A `replace` in the target's own `go.mod` DOES apply under `go run module@sha`, and one line pointing at a local path makes the build unauditable while looking ordinary

Under those, the resolved version IS the required version, derivable by reading one `go.mod` with no tooling and no graph walk.

⚠️ **This is the empty-require property RELOCATED, not abandoned**: it moves from every connector to every shared module. And the enforcement has to exist BEFORE the first dependency, which is s3's argument again.

#### What "100%" actually means

Setting the Go toolchain aside, the claim that holds is:

> **Every line of Go that this link compiles and runs is identified and fetchable from public sources named in the two pasted lines.**

What it does NOT cover, and this gap is larger than the toolchain one: **a target can exec other programs.** s40's `editor_open` runs `code --goto`. `remote_chrome_cdp` launches Chrome. An auditor sees THAT it execs something, because the `exec.Command` is in the fetched source, and cannot audit what it execs. s17 already concedes this and s24's UNVERIFIED is where it belongs.

#### ⚠️ And the thing that degrades rather than breaks

Today an auditor reads one folder and is PROVABLY done. With a dependency they must follow `go.mod`, and **stopping early is silent**. So the invariant must be enforced mechanically and the audit text must say "follow every require", or the property is true and nobody can tell.

#### Still unproven, and s1 is why that matters

Three parts are untested: a subdirectory module exposing an importable sub-package, a sibling required at a pseudo-version, and the whole thing resolving from a bare `go run` with no checkout. s1's lesson is that the install story rested on an untested assumption for weeks. One throwaway push settles it.

#### TWO MAXIMS, 14 Oct 2026, replacing the invariant list above

The list above went through four forms in one afternoon: four invariants, then three, then two, then these. Each reduction came from a question rather than from tidying, and the final form constrains something different from the first.

> **1. Every module in the closure lives in this repo.**
> **2. Nothing may stand between the declaration and the thing** — no `replace`, no `exclude`.

That is all. Depth is not constrained. A leaf may aggregate leaves, to any depth.

#### ⚠️ WHY DEPTH TURNED OUT NOT TO MATTER

Three of the earlier invariants existed to make the resolved version derivable by reading, so that an auditor could reason about MVS. They are unnecessary, for a reason that was available all along:

**`go.sum` enumerates the resolved closure.** Every module version in the build, transitive ones included. So the audit's stopping condition is one file:

> Read `go.sum`. That is the complete set of modules. You are done when you have read all of them.

No graph walk, no MVS reasoning, no depth limit. ⚠️ That also answers the worry recorded earlier in this row, that an auditor who stops after one folder has silently under-audited: they stop when `go.sum` is exhausted, and `go.sum` says when that is.

This came from looking for prior art. Nix's `flake.lock` is the canonical "one file naming the whole closure", and Go already ships the same thing.

#### What the earlier invariants were really protecting, and what replaces them

| earlier invariant | fate |
|---|---|
| every shared module is a leaf | **dropped.** Aggregation is allowed. Mechanism-versus-stance is a design judgement, not a rule |
| one level deep | **dropped.** `go.sum` makes depth irrelevant to the audit |
| the repo moves as one | **dropped**, and it was the expensive one: it coupled the release of unrelated connectors |
| no `replace`, no `exclude` | **kept** as maxim 2 |
| | **maxim 1 is new**: provenance, which is what all of it was groping at |

⚠️ **Provenance is the constraint, not shape.** With depth 1, every module in a build is one we named; with arbitrary depth, an aggregating leaf could pull in something nobody looked at. That is the third-party-dependency risk in a different hat, and it is what `AGENTS.project.md` means by calling zero dependencies "a property to keep rather than an accident". "Empty `require`" was never the point; "nothing we did not write" was.

Checkable in one line: every path in `go.sum` must begin with `github.com/davidnoz123/nielsoln_connectors/`.

#### And why the AWS-shaped pain cannot reach us

`aws-sdk-go-v2` and OpenTelemetry Go are the production examples of this exact pattern, one repo with a module per folder and a shared core, and both are known for version-alignment pain.

⚠️ **It does not transfer, and not because we are more disciplined.** Their consumers combine many service modules into ONE application build, so MVS has dozens of requirements on the core to reconcile. Our connectors are separate programs that are never linked together, and under maxim 1 every pin in any closure is ours to fix. Their pain is consumers inheriting a core they do not control.

#### The development mechanism, which maxim 2 would otherwise cost

Maxim 2 forbids the commonest multi-module move:

```
replace github.com/davidnoz123/nielsoln_connectors/cdpcore => ../cdpcore
```

Without it, editing a leaf and testing a consumer against it needs a commit and a push. `go.work` is the answer, and the property that makes it safe is the one that made it irrelevant to the audit:

| | travels with the repo | seen by `go run module@sha` |
|---|---|---|
| `replace` in `go.mod` | **yes**, which is the problem | yes |
| `go.work` | no, gitignored | **no, ignored by design** |

So: develop with `go.work`, never commit it, never use `replace`. ⚠️ It was NOT in `.gitignore`; added 14 Oct.

#### A note rather than a rule

Skew becomes possible once depth exceeds 1, and a `go.work` build is where it will surface, because a workspace unifies the graph. So: if you aggregate, expect to align pins, and the workspace is where you find out. A cost to know, not a prohibition.

#### Prior art, for comparison rather than adoption

| | what it does | what it offers us |
|---|---|---|
| Nix flakes | pins the closure INCLUDING the compiler | solves the toolchain gap s34 concedes, if it ever matters enough |
| `go.sum` + sum.golang.org | content hashes, plus an append-only public log | ⚠️ interesting against s47, which refused a published hash list. `sum.golang.org` IS one, and works because it is append-only and maintained by somebody else rather than hand-curated |
| SLSA / in-toto | provenance attestations; a layout of allowed materials | the vocabulary. in-toto's "layout" is s54's caveats file under another name |
| Deno | `--allow-net=example.com` on the invocation | the same shape as argv in a shimp link. Deno enforces at runtime; s17 refuses to |
| object-capability systems | an unforgeable reference IS the authority | ⚠️ we borrow the word "capability", not the semantics. Ours is a name; authority comes from the OS |
| Android / extension manifests | declared permissions, reviewed at a gate | the cautionary tale below |

⚠️ **The cautionary one.** An unenforced declaration becomes noise: P3P, Do Not Track and `robots.txt` were all declarations nobody could verify, and all ended up ignored or lied in. Android permissions work because the OS ENFORCES them, and s17 refuses enforcement.

s54 uses a third mechanism neither of those: not "the OS prevents it" and not "trust the declaration", but **a test proves the declaration complete against the code**. That is the only reason to expect our manifests to survive where P3P did not, and it is worth being clear about.

#### ⚠️ PROVED, 15 Oct 2026, by the first real click

This row waited on one throwaway push to settle three untested assumptions. s68's capability forced it, and all three hold:

* a module in a **subdirectory** of a multi-module repo resolves by `@<sha>`
* the **proxy** serves it, as `v0.0.0-20261015021557-4252cb9149e8`
* **no checkout anywhere**: run from an empty temp directory, 24s cold

```
go: downloading github.com/davidnoz123/nielsoln_connectors/chrome_ready
    v0.0.0-20261015021557-4252cb9149e8
```

So the module graph a link pins is fetchable exactly as the design assumed, and `go.sum` enumerating the closure is now a property of something that has actually run.

### s56. NOTHING in the audit payload may be an assertion

Decided 14 Oct 2026, prompted by the suggestion that per-link caveats be pasted alongside a link.

> **Every element of the audit payload must be resolvable to public bytes, or it does not belong in the paste.**

#### Why

s9's two lines work because **nothing in them is claimed.** Line 1 is a SHA you fetch. Line 2 is a URI you resolve. Every assertion is checkable against public bytes, so the sender's honesty is irrelevant.

⚠️ Add prose and you add something the auditor must either trust or ignore. A hostile sender writes reassuring caveats; a careless one writes stale ones; and **the auditor cannot distinguish either from the truth**, because the text is not derived from anything. Worse, a declared link "type" would itself be a claim, so a link could claim a type with mild caveats.

#### Four decisions already follow from it, and none states it

| row | what it refused | because |
|---|---|---|
| s21 | a `policy` field | an unenforced claim about a target's own behaviour |
| s35 | the descriptor store | state beside a link, able to drift |
| s47 | a published list of known-good hashes | state to maintain, and it goes stale |
| s54 | caveats pasted with a link | an assertion the auditor cannot check |

It also explains why s54's caveats work: they are not in the payload at all. They are in the **source at a pinned commit**, which resolves to public bytes like everything else.

#### What replaces a caveat, when one is needed

A **question list**, published once with SHIMP and emitted by `shimp audit-text` (s25), applying to whatever was fetched. [27] had this shape and s24 records it: "report what exact code will execute, what resources it can reach, what persists, what network access exists, any unverified assumptions."

Those are prompts to LOOK, and the answers are derived. "This execs VS Code" is not a caveat somebody wrote; it is `exec.Command` in the fetched source.

#### ⚠️ And the honest corollary

An audit that always returns SAFE is worthless. s24 has four verdicts so that a link which SHOULD make somebody pause does. The design's job is to make the unease **accurate**: present when warranted, absent when not, and never manufactured by a sender who wanted a click.

### s57. remote_chrome_cdp is a hand-rolled shimp, and three of its four invented pieces become unnecessary

Observed 14 Oct 2026. The tool was written for the `amazeme` sibling to automate navigation from a URL, which is shimp's job description reached independently and earlier.

Its links look like `http://127.0.0.1:8790/show/logo-url?t=K7M2XQ4P`: a clickable thing in a document that causes one named action from a closed set to happen locally. That is `shimp://<capability>/<action>` with a hand-rolled transport.

#### What it had to invent, and what shimp makes unnecessary

| invented | why | under shimp |
|---|---|---|
| `-listen` port | a link needs something to hit | gone. s20: local only, no port to manage |
| `-token` | any page in the browser can call localhost | gone, and this is the real win |
| `-manifest` | a closed set of allowed ids | gone. Each allowed page is its own capability, so its own link |
| a long-running server | so links keep working | optional. s46's one-shot class |

#### ⚠️ The token is substantive, not tidiness

Its own header concedes it: "it travels in a document, so it is not a secret. The thing that makes this safe to run is that the vocabulary is closed."

A localhost HTTP server is reachable by **any page the browser has open**. A `shimp://` link is not: a sandboxed page cannot silently ask the OS to open a custom scheme. So the conversion removes an exposure the current design can only mitigate.

#### ⚠️ And the manifest point is subtler

A manifest is an EDITABLE allowlist consulted at request time. Under shimp the allowlist becomes **which links exist**, and each link is a pinned specification with no editable middle layer. That is s35's move exactly: state between the link and the behaviour is state that can drift or be tampered with.

#### What stays, and it is what the tool is for

Open a tab it created, navigate to one address, highlight one string in the page source, bring the window forward. "Replaces trust with looking."

Of 3,517 lines, the HTTP server, manifest, token guard, sweep lanes and health endpoint are transport scaffolding. Strip them and what remains is the page interaction, which is exactly what an auditor should see, undiluted rather than buried.

#### Three cautions

* **nothing forces this.** It works, and the measured lessons in it, the 2 Oct tab hijack and the dedicated-profile discovery, are embodied in that code and must survive a rewrite rather than be lost with it.
* **sequence.** shimp's dispatch half does not exist, so: shimp first, then convert. Converting in anticipation leaves a tool that works neither way.
* ⚠️ **one real loss.** The server form works from any document that can make an HTTP request. A `shimp://` link only works where shimp is installed. Irrelevant for a local playbook, and worth knowing before calling the server form obsolete.

#### And `amazeme` has no live consumer

Checked 14 Oct: it appears in `amazeme`'s slate, a `drive-server.log` and `scripts.txt`, and in no `.py` or `.md` code path. So a conversion would break nothing today.

### s58. `vscode:` is already registered, so s40's first capability already exists

Measured 14 Oct 2026, while looking for prior art:

```
HKCU\Software\Classes\vscode\shell\open\command
  "...\Microsoft VS Code\Code.exe" --open-url -- "%1"
```

So `vscode://file/<path>:<line>:<col>` works on this machine **today**, with no shimp, no install and no code. s40's first capability, "open an editor at a line of a file", is shipped prior art from Microsoft.

#### Three consequences

* ⚠️ **raccoon's actual need is met now.** A `vscode://file/...:1735` link in a `hyperlink=True` column keeps the line number, which is strictly better than slate_tools' interim suggestion of losing it. s40 and `TO_DATAANNOTATION_TOOLS.md` both offered a worse answer than the one already installed.
* ⚠️ **and they will hit s8's modal**, because Office does not know `vscode:` either. Same fix, same trailing colon: `...\Trusted Protocols\All Applications\vscode:`
* **shimp's first capability should NOT be `editor_open`.** That problem is solved. The first capability should be something `vscode:` cannot do, which means pinning and audit rather than navigation.

#### What it says about shimp, which is the more useful part

The pattern shimp generalises is already in production from Microsoft, and from JetBrains, GitHub, Slack and Zoom. A custom scheme that takes an action and acts locally is ordinary.

⚠️ So shimp's novelty is NOT the scheme, and never was. [32] reached the same conclusion from a precedents search on 7 Oct: every ingredient is established and only the composition is not, with the two-line audit "the most original and valuable part".

This is that conclusion arriving from the other direction, by measurement rather than by search, and it sharpens what the first capability has to demonstrate: **immutable pinning and a two-line audit of code a stranger can fetch.** `vscode:` has neither. It runs whatever VS Code happens to be installed, with no way to say which.

#### Answered 14 Oct 2026

s61 is the answer: **open an editor you chose**, at a line. Measured, nothing but `vscode:` registers an editor scheme, so for every other editor there is no link-based route at all. A capability doing something that cannot currently be done is a better demonstration than one doing something better.

And the configurability is the demonstration: the editor is in argv, so it is in the link, pinned and auditable, with no config file between the link and the behaviour.

### s59. Cached attestations per module@sha: capabilities verified, exposure attributed

Decided 14 Oct 2026, from the question of whether certifications could be cached by hash the way VirusTotal caches scan results.

#### ⚠️ Not a scanner verdict, and that ruling-out matters

Antivirus detects known-bad signatures in binaries and answers a different question from "what can this reach". A clean result on a 400-line Go connector is an **absence of evidence** that would be read as a certification, which is s56's failure exactly. Go binaries also false-positive routinely, being unsigned, statically linked and novel on every build, so the noise runs the harmful way too.

And uploading is a **publication**: hash lookup is safe, upload is not, and nothing carrying a descriptor, a URI with a home directory in it, or a token may go anywhere.

#### What is actually cached

| part | source |
|---|---|
| **subject** | `module@<full sha>`, plus the `go.sum` content hash once a consumer pins it |
| **capabilities** | that module's `Caveats`, which s54 verifies against the source mechanically. Facts |
| **exposure** | what those mean in practice, each line CITING the code that constrains or enables it. Judgement |
| **verdict** | s24's four |
| **attribution** | who, and when |

The expensive half is the exposure judgement, it is stable for a fixed version, and it is what a reader actually wants. That is what makes caching worth anything.

#### ⚠️ THE STRUCTURAL RESULT: a leaf can only ever be CONDITIONAL

Writing one for `cdpcore` makes it obvious. It opens a websocket to a caller-supplied host and sends caller-supplied methods, so **its safety is determined entirely by its caller**. You cannot certify a mechanism safe or unsafe in isolation; all you can certify is the conditions.

That is the mechanism-versus-stance split arriving from the audit side, and it sharpens it: **a leaf has no stance, therefore no safety property. A program has a stance, therefore can be certified outright.**

#### Which makes the audit COMPOSE rather than repeat

```
audit of a program = read its own code
                   + for each leaf in its closure, check the program
                     DISCHARGES each condition that leaf's attestation states
```

Auditing `remote_chrome_cdp` becomes: does it constrain `cdpcore`'s host? Yes, a constant. Its method? Yes, "THE OP VOCABULARY IS CLOSED". Both discharged, and that is evidence `cdp.go`'s header already volunteers. Close to in-toto's layout, where a step declares what it requires of its materials.

#### ⚠️ An attestation cannot live inside what it attests

The subject is a content hash, so a file inside the module becomes part of that hash: the attestation either names a version not including it, or names itself and cannot be computed.

Same circularity as s54's, where an expected hash could not live in the file it hashed. There the fix was a separate test file; here it is `attestations/`, outside every module, so adding one changes no module's content hash.

#### Why this is NOT what s47 refused

s47 rejected a published list of known-good hashes, as state to maintain that goes stale. ⚠️ **A statement about an immutable subject cannot go stale.** What goes stale is "which version is current", and that is not what these record. s47's objection stands for a latest-good list and does not reach a per-version attestation.

#### And it is a new argument for s13

Because a pinned version is immutable, its audit is **cacheable**: audit `cdpcore@4b91c7e` once and every link whose closure contains it inherits that work for ever. **Copies get none of this**, since each is bytes in a different module at a different version and needs its own look every time.

So over many links the leaf route gets cheaper to certify while the copy route stays flat. That argument only appears when you think about caching rather than about one audit.

#### ⚠️ The guard against becoming P3P

The failure mode is a reader trusting the attestation INSTEAD of looking. So every exposure line must carry **the input to its own judgement**: not "this cannot leave loopback" but "because `host` is a `const` at `sessions.go:51`, not a flag". Spot-checkable rather than believed.

#### What the first one found, which is why it was worth writing

`attestations/chrome_sessions@5c46b13.md`. Verdict **CONDITIONAL**, and three findings the exercise produced rather than confirmed:

* ⚠️ **it reads the whole machine's socket table on EVERY run, with no flag.** `ownerOfPort(port)` is unconditional at `sessions.go:114`, and `owner_windows.go` fetches the entire table to answer for one port.
* ⚠️ **but by default only one pid per probed port is DISCLOSED.** The rest is read into the process and never emitted. The distinction between reading machine-wide state and disclosing it is the one an auditor needs, and I had the framing wrong in both directions before checking the code.
* ⚠️ **`-discover` turns the capability into a disclosure**, because `discoverPorts()` enumerates listening ports from that same table, so the output then IS a machine inventory.

A first attestation returning SAFE would have demonstrated nothing about the format. This one disagrees with the tool's own description of itself, which is the point.

#### Where it goes next

Shaped to move to a signed transparency log, Sigstore or Rekor, unchanged, when there is a second party to sign. Until then self-issued, which is weaker than third-party review and stronger than nothing: the capability half is verified, so only the judgement is a claim, and it is attributed and append-only through git.

### s60. The leaf candidates, MEASURED, and the divergence nobody was looking for

Measured 14 Oct 2026, after a whole session of reasoning about what should be shared.

#### ⚠️ REASONING FOUND cdpcore. MEASUREMENT FOUND PATH SAFETY.

Counting function names present in two or more connectors, and hashing each body so divergence shows:

| name | consumers | state |
|---|---|---|
| `contained` | pdf_view_server, remote_ai | **BYTE-IDENTICAL** |
| `foldsCase` | pdf_view_server, remote_ai | ⚠️ **DIVERGED**, 23 lines against 16 |
| `resolve` | both | same name, different jobs, 12 lines against 92 |
| `raiseWindow` | pdf_view_server, remote_chrome_cdp | different jobs, 96 against 14 |

`cdpcore`, which s13 spent the session on, has ONE consumer today. The path cluster has two, now.

⚠️ **And the websocket duplication did not appear at all**, because the two implementations use different function names. So name-matching UNDERSTATES duplication, and the only reason that pair is known is that it came up in conversation.

#### The divergence, and what is actually at stake

`contained(root, real)` decides whether a resolved path sits inside the permitted root. It is the traversal defence, and it is byte-identical in both connectors.

It calls `foldsCase`, which is not:

* **pdf_view_server** stats the directory, flips its case, stats again, and compares with `os.SameFile`. On a failed flipped-stat it **caches `false`**.
* **remote_ai** delegates to `probeFold` and deliberately does **not** cache an undetermined answer, with the reason in its own comment: *"An empty share cannot be asked whether it folds, and the moment it has a file in it, it can. Caching the guess would keep answering from the one moment the question was unanswerable."*

⚠️ **So a learned lesson is stranded in one copy**, which is precisely s13's hazard, realised, in the one area nobody had looked at.

#### Severity, stated accurately rather than alarmingly

The failure direction is over-strict **REJECTION**, not admission: a wrong `false` makes `contained` compare case-sensitively, so a legitimate path is refused. Both copies are conservative in the dangerous direction, where a wrong `true` would treat two different directories as one.

So this is a **divergence in safety-adjacent code**, not a hole. ⚠️ The live defect is pdf_view_server caching a `false` derived from a transient stat failure, for the life of the process, which is exactly what remote_ai's comment warns against.

Not fixed here: propagating remote_ai's behaviour changes a working tool, and that is not a change to make silently.

#### The candidate list, ranked by evidence

1. ⚠️ **path safety** -- `contained`, `foldsCase`, `flipCase`, `probeFold`. Two consumers TODAY, partly identical, partly diverged, safety-adjacent. Never mentioned in any discussion.
2. **the caveats checker** (s54). One consumer, five planned. Pure stdlib, pure mechanism, and rolling s54 out is about to create the duplication.
3. **`cdpcore`** (s13). One consumer today, two once the probe exists.
4. the websocket pair. Two implementations, legitimately different environments.
5. `raiseWindow`. Probably not: one name, two jobs.

#### What this says about the method, which is the transferable part

We reasoned our way to a leaf from a PROJECTED second consumer, while the live problem had two ACTUAL ones and a diverged helper.

⚠️ So the answer is not better reasoning, it is a cheap recurring measurement. `repocheck` is that: a program outside every module, because neither of s55's maxims is checkable from inside one, with a declared baseline so that a NEW overlap fails rather than being noticed by somebody in a year.

Overlap is not a defect, and the baseline is the point: `main` is in every connector, and two functions can share a name while doing different jobs. What is forbidden is UNNOTICED overlap.

#### ⚠️ And the baseline immediately caught itself

The first `known_overlaps.go` listed seven names that were not overlaps at all -- `flipCase`, `probeFold`, `windowExists`, `isReparse`, `isReparseInfo`, `isLoopback`, `contains` -- taken from a grep that had not distinguished a cross-connector overlap from a build-tag pair inside ONE module.

`TestBaselineHasNoStaleEntries` found all seven. That is the same discipline in the other direction: an allowlist must not claim duplication that does not exist, or the next reader cannot tell which entries are live.

Every check in `repocheck` also has a companion proving it CAN fail, because three times in this repo a safety check has passed by being blind: s54's walk missed its package's primary reach, then matched per-file rather than per-reach, then could not see inside an import.

#### CLOSED 14 Oct 2026: the lesson is propagated

pdf_view_server was cleared for modification, so remote_ai's lesson went in rather than staying stranded.

⚠️ **Only ONE of three early returns was guessing.** A failed stat of the directory itself, and a directory with no letters to flip, were both already uncached. The flipped-stat failure cached `false` unconditionally, so a permission error, a share hiccup or an antivirus lock would be remembered as "case-sensitive" for the life of the process. Now only `os.IsNotExist` caches, because that is the one failure that IS evidence.

**The lesson is propagated, not the code.** remote_ai delegates to `probeFold` while pdf_view_server inlines the stat pair, so the bodies still differ and `repocheck` still reports DIVERGED. That is accurate rather than untidy: the behaviours agree and the implementations do not.

⚠️ And `repocheck` SAW the change, the hash moving from `aa622d57ec75` to `a85800ddcab6`. That is the only evidence that the check is live rather than decorative.

#### And writing the test found something about the code

An assertion claimed the "nothing to flip" branch fires for a name like `123`. It does not: `flipCase` operates on the WHOLE PATH, and every absolute Windows path carries a drive letter, so `C:\123` flips to `c:\123` and **that branch is unreachable on Windows.** It is live only where a path can contain no letters at all, which means Unix.

Said plainly in the test: the non-ENOENT branch is the new logic and is NOT directly testable, because a permission error is not something a test can portably arrange. The two reachable not-cached paths are covered and the third is argued rather than proved.

#### ⚠️ A SECOND CONSUMER appeared for the websocket, 15 Oct 2026

s63 needs `Input.dispatchMouseEvent` to automate the browser routes, which needs a websocket, which Go's stdlib does not have and this repo will not take as a dependency.

This row noted that the websocket duplication never showed up in the overlap count, because the two implementations use different function names, and that name-matching therefore UNDERSTATES duplication. It now has what `cdpcore` was ranked third for lacking: **a real second consumer, today, not a projected one**.

`remote_chrome_cdp`'s `wsConn` is the extraction candidate, and `linkprobe` is the caller waiting on it.

#### ⚠️ THE SECOND CONSUMER IS WITHDRAWN, 15 Oct 2026

The note above claimed `linkprobe` was a real second consumer for the websocket, waiting on `wsConn` being extracted. **s63 measured the premise and it is false**: a CDP-synthesized click does not satisfy Chrome's external-protocol gate, so a websocket would not have moved the browser route at all.

So the websocket candidate is back to where this row originally put it: **two implementations, and the duplication never showed up in the overlap count because they use different function names.**

⚠️ And `remote_chrome_cdp/ws.go`'s own header already explains why the two are deliberately different: remote_ai's client carries TLS, permessage-deflate and a handshake hardened against captive portals because it crosses the internet; this one talks to `127.0.0.1` and nothing else, so each of those costs buys nothing here. It even names the right future move, which is to lift remote_ai's `dial()` when the bridge transport lands rather than to merge the two now.

So this is not undiscovered duplication awaiting extraction. It is **a recorded decision that the two should differ**, and a shared leaf would have to overturn it rather than tidy it.

#### ⚠️ THE BLIND SPOT IS NOW A CHECK, 15 Oct 2026

This row said name-matching UNDERSTATES duplication, pointing at the two websocket clients that never appeared in the overlap count. s65 acted on it, after it bit a second time: `shimp/live_test.go` and `linkprobe/main.go` had both grown the same five-step install cycle, invisible for two reasons -- the name key, and `Overlaps` skipping `_test.go`.

`SharedBodies` now inverts the key and reads tests. ⚠️ It caught an `itoa` I had just hand-rolled against one already sitting in `chrome_sessions/caveats_test.go`, which is exactly the class of thing this row predicted and exactly the file class `Overlaps` cannot see.

And what no check can find is now DECLARED, in `known_recipes.go`, with markers and a staleness test. Both websocket clients are in it, with the reason they deliberately differ.

### s61. editor_open: the first capability, because NO editor but VS Code has a scheme

Designed 14 Oct 2026 and deliberately not built. This is what s58 asked for: a first capability demonstrating something `vscode:` cannot do.

#### ⚠️ The measurement that makes it the right one

| scheme | registered on this machine |
|---|---|
| `vscode` | **yes**, hardwired to `Code.exe --open-url -- "%1"` |
| `notepad++`, `nppp`, `sublime`, `subl`, `idea`, `textmate` | **none of them** |

Two facts follow, and the second is the stronger:

* ⚠️ **`vscode:` cannot open anything but VS Code.** The scheme name IS the program, so there is no version of that link which opens another editor.
* ⚠️ **There is no link-based route to Notepad++ at all.** Not a worse one. None. Nothing registers a handler, so no clickable thing can open it at a line.

So this capability does something that currently **cannot be done**, rather than something done better. That is what a first capability has to show.

#### The configurability, which comes for free

The editor goes in the capability's **argv**, so it is in the link:

```
shimp://v1/go/run/.../editor_open@<sha>/-editor/notepad++#open?file=...&line=1735
shimp://v1/go/run/.../editor_open@<sha>/-editor/code#open?file=...&line=1735
```

Different argv, different identity (s44), different singleton (s53). ⚠️ **So configuration becomes "which link you click"**, which is better than a config file for the same reason s35 deleted the descriptor store: there is no hidden state between the link and the behaviour that can drift from what the link says. It is pinned, and it is visible in s9's two-line paste.

A column can hold whichever editor is wanted, per row.

#### Finding the executable is already a solved pattern here

`HKLM:\SOFTWARE\Notepad++` holds `C:\Program Files\Notepad++`, measured, so **no path need go in the link** for the editor itself. `remote_chrome_cdp`'s `chromeCandidates()` is the house pattern: search known locations and REPORT what cannot be found rather than guessing, which is `owner_other.go`'s rule applied to an exe.

#### ⚠️ One honest cost, which is unavoidable here

The **file path** must be in the link, because it is the payload. So s9's concern applies: pasting `...#open?file=C:/Users/.../x.py&line=1735` into a chat window publishes that path.

Nothing fixes it for this capability, since naming the file is the whole point. It is an argument for relative paths against a declared base wherever the file sits inside a known root, and for stating the cost rather than discovering it.

#### Still to measure before building

* **the Notepad++ invocation.** `notepad++.exe -n<line> <file>` is believed to open at a line and has NOT been run. s1's lesson is that an install story resting on an untested assumption is a liability.
* **whether an editor needs to be resident.** s46's three classes: a one-shot pays `go run`'s 650ms to 4.8s on every click (s49), which for opening a file is tolerable but noticeable. A resident pays it once. s40 argued resident; that argument stands and is worth re-measuring against a real editor launch, because the editor itself may dominate.

#### What it inherits

* s50: no console field. A one-shot editor launch prints nothing, so it gets no console
* s54: a caveats file, which will read `[exec.Command] launches an editor binary found by searching known locations`
* s59: an attestation, whose verdict will be CONDITIONAL at best, because it execs a program the audit cannot see
* s45: the file path percent-encoded, and `%2E%2E` rather than a literal `..`

### s62. linkprobe: fourteen ways in, MEASURED, and two of them cannot carry a shimp link

Built and run 15 Oct 2026. `linkprobe` is a seventh module: one link, every way a human or a script can trigger it, and a report of what each route did to it.

#### ⚠️ Why this is not about coverage

s44 makes a link's identity the sha256 of its bytes, verbatim. s35 then says refuse, never normalise. Together those mean a route that alters ONE byte does not fail cleanly: it either refuses a link the user can see is correct, or runs a different pin from the one named in the text they pasted.

So every entry point is a potential mangler. Before this, exactly ONE had ever been run: `ShellExecuteW`, by the live install test. The spreadsheet cell, the address bar, the email, the command line were all reasoned about and never measured.

#### The matrix

| route | result | |
|---|---|---|
| `shellexecute` | **VERBATIM** | the baseline, shell32 direct |
| `cmd-start` | **VERBATIM** | |
| `powershell-start-process` | **VERBATIM** | ⚠️ but imposes its own length ceiling |
| `explorer` | ⚠️ **ABSENT** | see below |
| `python-startfile` | **VERBATIM** | |
| `python-webbrowser` | **VERBATIM** | and it goes via the shell, not a browser |
| `curl` | REFUSED | correct: `Protocol "shimp-linkprobe" not supported` |
| `wget` | UNRUNNABLE | absent from this machine, which is a different fact |
| `excel-hyperlink-object` | **VERBATIM** | ⚠️ **the row that matters most** |
| `excel-hyperlink-formula` | UNRUNNABLE | see below |
| four human routes | UNANSWERED | named, not skipped |

#### ⚠️ FINDING 1: explorer.exe silently drops any URI containing `?`

Deterministic over repeated passes:

| sent | arrived |
|---|---|
| `...@<sha>/-route/x#open` | VERBATIM |
| `...@<sha>/-route/x#open?route=x` | **nothing** |
| `shimp-linkprobe://x?q=1` | **nothing** |

Exit code 1 and no message, in the path or the fragment alike.

**shimp's action syntax is `#open?file=...&line=...`**, so explorer cannot deliver any link that carries action arguments. s61's `editor_open` link is exactly that shape.

And it is not our registration: `ms-settings:` and `vscode:` both dispatched through explorer, `vscode` being HKCU-registered exactly as shimp is.

#### ⚠️ FINDING 2: explorer also normalises, which is the s35 hazard in the wild

```
sent     shimp-linkprobe://x#y
received shimp-linkprobe://x/#y
```

An inserted `/`, before the `#`, so **inside the identity-bearing part**. The hash changes. Real links always have a non-empty path so this one does not bite today, but it is the first observed case of a route rewriting a URI rather than an argument that one might.

#### ⚠️ FINDING 3: the 255-byte truncation is not where it was recorded

A 326-byte URI written into `=HYPERLINK()` came back from `cell.Formula` at **326 bytes, unaltered**.

So whatever truncates at 255 is not what the formula STORES. It may be the `Hyperlink.Address` property, or what a click resolves, or the original number may have measured a different route. **Recorded as unlocated rather than quietly dropped**: the claim is not disproved, its subject is.

#### ⚠️ FINDING 4: a formula hyperlink exposes no object to follow

`Hyperlinks.Count` is **0** after writing `=HYPERLINK()`. There is no `.Follow()` to call, so the DELIVERY half of that route is answerable only by a human clicking the cell.

That is now a third independent reason the formula route is unusable, alongside the pull/publish destruction and whatever the 255 limit turns out to be.

#### And the Excel object route passes, which is the one that had to

`Hyperlinks.Add` then `.Follow()` delivered **VERBATIM**. That is how a link reaches a human in this workspace, and it had never been run. No Office modal blocked it, though that is reported rather than concluded: a `.Follow()` from COM may not be the same path as a click.

#### Three defects in the probe itself, which is the pattern by now

1. **The classifier called explorer's silent drop a REFUSAL.** It keyed on the exit code, and explorer exits 1 whether it dispatched or dropped. The function's own comment warned against trusting exit codes and the code below it did exactly that. A refusal now requires a message.
2. **The formula driver measured nothing.** It wrote the formula and then followed the URI *variable*, so it reported VERBATIM for a link the route had never touched.
3. **The probe link was too short to cross the threshold it was testing.** 183 bytes against a 255-byte limit. Routes now declare a `Pad`.

Each was caught by looking at what the handler's log actually held rather than at the verdict the probe printed. ⚠️ And the report now prints every non-delivering route's detail whether or not it matched expectation, because the first version hid the Excel measurement behind "UNRUNNABLE as expected".

#### What makes it recur rather than be re-reasoned

* the route table is DATA, so adding a way in is adding a row
* each route declares its EXPECTED outcome next to itself, so a known-bad route passing is not news and a change is
* a non-Verbatim expectation must carry WHY, enforced by a test, because "this one is supposed to fail" is the easiest claim to leave behind after its reason stops being true
* the four human routes are reported BY NAME as unanswered, so the matrix has no invisible holes
* it rests on s48: an isolated scheme, installed and purged per run, refusing the production scheme by name

`repocheck` saw the new module on its first run and flagged `probe` as an undeclared overlap with `chrome_sessions`. Different signatures, different jobs, now declared.

#### ⚠️ The browser rows, attempted and understood. 15 Oct 2026

s63 tried to automate `chrome-address-bar` and `chrome-link-click` and established why they cannot be:

* **Chrome keeps the URI VERBATIM**, fragment included, which is the opposite of what `explorer` does
* **and refuses to hand it to the OS without a real user gesture**, which is a security property rather than an obstacle: a web page cannot silently fire a link that names code to run

So the two rows stay `NeedsHuman` on purpose. Automating them needs a TRUSTED input event, so a websocket, so `remote_chrome_cdp`'s `wsConn` extracted as a leaf.

⚠️ And one apparent finding was withdrawn: the fragment looked stripped, and that was an unencoded `#` in the probe's own `/json/new` request rather than anything Chrome did.

#### ⚠️ CORRECTION 15 Oct 2026: the security claim above is FALSE

This row recorded that a web page cannot silently fire a shimp link, and treated the gesture requirement as a security property. **s64 measured it and it is wrong.**

A page served over http, with the protocol permitted for its origin, fires the handler from a **meta refresh**: no JavaScript, no click, no interaction of any kind. Reproducible 3 passes of 3 on a fresh browser.

The per-origin permission is the sole gate, granted once by a user ticking "always allow". Everything after that is silent.

### s63. The browser route: Chrome keeps the link VERBATIM and refuses to fire it without a gesture

Measured 15 Oct 2026, attempting to automate s62's two Chrome rows.

#### What was tried, and what each attempt said

| attempt | handler |
|---|---|
| `chrome.exe <uri>`, throwaway profile | silent |
| CDP `PUT /json/new?<uri>` | silent |

The second is the informative one. Chrome **created the tab**, and `/json/list` showed it holding the URI complete, fragment and all, byte-for-byte as sent.

#### ⚠️ SO CHROME DOES NOT MANGLE A SHIMP LINK

That is the first half of the answer and it is the reassuring half. s62 found `explorer` dropping any URI with a `?` and inserting a byte into an empty path. Chrome did neither: the full 139-byte URI including `#open?route=...` came back from the tab unaltered.

#### ⚠️ AND IT DECLINES TO HAND IT TO THE OS, WHICH IS A FEATURE

Chrome gates an external-protocol launch behind a **real user gesture**, and a programmatic navigation is not one. The tab sits on the URI with an empty title and nothing reaches the handler.

Read the right way round, that is a **security property in shimp's favour**: a web page cannot silently fire a `shimp://` link. Given that a shimp link names code to run, a browser that launched one without a gesture would be the hole. So the gesture is the mechanism, not an inconvenience around it.

This is why both Chrome rows stay `NeedsHuman` rather than becoming automatable: the thing that cannot be automated is the thing that makes the route safe.

#### ⚠️ A NON-FINDING, recorded so nobody inherits it

The first CDP attempt appeared to show Chrome **stripping the fragment**. It was not Chrome. The `#` was unencoded in the `/json/new` request URL, so Chrome parsed it as a fragment of *its own* endpoint and never saw it. With `%23` the fragment survived intact.

An artifact of the probe's transport, indistinguishable from a finding until it was encoded properly. Written down because a stripped fragment is exactly the kind of result that would have been believed.

#### What would close it, and what it costs

A **trusted** input event: `Input.dispatchMouseEvent`, which Chrome treats as genuine user activation because it is injected below the page.

That needs a websocket. Go's stdlib has none and this repo takes no third-party dependencies, so it needs `remote_chrome_cdp`'s `wsConn` extracted as a leaf. s60 already ranked that extraction among the measured candidates; this is a **second consumer** for it, which is the thing s60 said `cdpcore` lacked.

⚠️ So the browser rows are blocked on the leaf extraction, not on anything about Chrome.

#### Two mechanisms deliberately NOT used

* **`AutoLaunchProtocolsFromOrigins`**, Chrome's documented policy for pre-allowing a protocol. It is real and it would work. It is also machine state under `HKCU\SOFTWARE\Policies`, and a probe should not rewrite the user's browser policy to make its own measurement pass. Attempted once and correctly refused by the environment, which was the right answer.
* **the profile-local preference** `protocol_handler.allowed_origin_protocol_pairs`, seeded into a throwaway profile. This touches no machine state and was tried. It did not make a programmatic navigation fire, which is consistent with the **gesture** being the gate rather than the permission.

#### And the whole attempt touched nothing

A throwaway `--user-data-dir` per launch, the test scheme purged, the temp profiles removed, and the user's own 25 Chrome processes untouched throughout. Verified afterwards rather than assumed: no scheme class, no policy key, no probe home.

`remote_chrome_cdp`'s rule is *we open our own tab, always*, because attaching to an existing one once drove somebody else's tab away from its own page and reported a confident result measured on the wrong document. The same rule one level up is **our own profile, always**.

#### ⚠️ CLOSED 15 Oct 2026, and the answer is NO

This row named extracting `remote_chrome_cdp`'s `wsConn` as the work that would close the browser route. **That was a guess, it was tested before it was built, and it is wrong.**

The premise was proved out with a throwaway Python CDP client rather than by paying for the Go extraction first, since the venv already carries `websocket-client` and `linkprobe` already has a Python-driven route in the Excel pair.

| | |
|---|---|
| **CONTROL**: synthetic click on an ordinary `<a href>` | **NAVIGATES** |
| **SHIMP**: the same click on a `shimp://` `<a href>` | handler **silent** |

⚠️ **The control is what makes this conclusive.** Without it, a null result is indistinguishable from a probe that missed the link. The ordinary link navigated from `control.html` to `target.html`, so the click mechanism works, the coordinates are right, and Chrome processes the event as a genuine link activation.

The same click on a custom scheme fires nothing, raises **no dialog** (window enumeration on the test Chrome found only the page itself) and leaves the tab where it was. Two passes, same result.

#### What that means

Chrome distinguishes **CDP-injected input from real user input for an external-protocol launch specifically**, and blocks it silently.

So s62's reading holds and strengthens: the gesture requirement is not a permission that can be granted or a dialog that can be dismissed. It is a genuine trust boundary, and **CDP sits on the wrong side of it**.

⚠️ There is no automated browser route to be had. Both Chrome rows are human-only by Chrome's design rather than by any limitation in our tooling, and that is now measured rather than inferred.

#### ⚠️ And the extraction is NOT justified by this row

Which is the useful outcome: a 205-line leaf extraction, plus a second consumer for a module whose page-interaction content is the most audit-sensitive thing in the repo, was avoided by a twenty-minute test.

#### Two more things measured on the way

* ⓘ **Chrome 155 rejects a CDP websocket carrying an `Origin` header** with 403 unless launched with `--remote-allow-origins`. `suppress_origin` is the right fix: a CDP client has no business claiming a web origin. ⚠️ The first run of the premise test reported "PREMISE FAILS" on this 403, which was the harness mislabelling its own transport error as a result. The traceback is the only reason it was caught.
* `DOM.getBoxModel` plus `Input.dispatchMouseEvent` is enough to click a link with **no `Runtime.evaluate` anywhere**, which respects `cdp.go`'s closed vocabulary and is a more faithful model of a click in any case.

#### ⚠️ CORRECTION 15 Oct 2026: this row's conclusion was WRONG

**There IS an automated browser route.** s64 has it working, from the user's suggestion of serving the links over http rather than from a file.

Two faults in the measurement this row rests on:

* ⚠️ **`file://` is an opaque origin** and cannot hold Chrome's external-protocol permission, so that test could never have succeeded. The conclusion was generalised from the one origin incapable of passing.
* ⚠️ **the control was unreliable.** `Page.bringToFront` alone does not make a page hit-testable; `Page.setWebLifecycleState` active and `Emulation.setFocusEmulationEnabled` were also needed. Until then an ordinary link did not navigate either, so the null result said nothing.

A CDP click **does** count as a user gesture. And that turns out not to matter, because **no gesture is required at all** once the origin is permitted, which also disproves this row's security claim.

### s64. The browser route AUTOMATED, and the security property turned out to be false

Measured 15 Oct 2026, from the user's suggestion: serve the test links from a tiny Python server and click them with CDP.

#### ⚠️ The suggestion corrected an error of mine, and that is the main event

s63 concluded there was **no automated browser route** and that Chrome blocks an external-protocol launch from CDP-synthesized input as a trust boundary. Both halves were wrong, for two separate reasons:

1. ⚠️ **`file://` IS AN OPAQUE ORIGIN.** Chrome keys its external-protocol permission on origin, so a `file://` page **cannot hold one**. The test could never have succeeded, and I generalised from the single origin incapable of passing.
2. ⚠️ **MY CONTROL WAS UNRELIABLE.** Hit-testing an inactive page silently finds nothing, which is indistinguishable from a refused click. `Page.bringToFront` alone was not enough: `Page.setWebLifecycleState` active **and** `Emulation.setFocusEmulationEnabled` were both needed. Until they were, an ordinary link did not navigate either, so the null result said nothing at all.

An http origin is also the **realistic** route: a link on a web page is what shimp links are for. `file://` was never the thing to measure.

#### The measurement, with a control that passes

| origin | permission | gesture | handler |
|---|---|---|---|
| `file` | cannot hold one | click | silent |
| `http` | none | click | navigation requested, launch **REFUSED** |
| `http` | granted | click | **FIRES**, verbatim |
| `http` | granted | meta refresh | ⚠️ **FIRES**, verbatim, 3 passes of 3 |

#### ⚠️ AND THE SECURITY PROPERTY IS FALSE

s62 and s63 both recorded that *a web page cannot silently fire a shimp link*, and both treated it as a property in shimp's favour. **It is not true.**

The last two rows settle the mechanism between them:

```
click,      NO permission  ->  ABSENT   so the permission is NECESSARY
NO gesture, permission     ->  FIRES    so a gesture is NOT required
```

So the **per-origin permission is the sole gate**. A user grants it once, through Chrome's "always allow" checkbox, and from then on that origin can fire shimp links on any later visit with **no interaction whatsoever**. A meta refresh does it, with no JavaScript and no click.

⚠️ An earlier one-off run DID come back silent, which is why the claim was believed twice. That run reused a Chrome already driven through two other tests with its server restarted underneath it. The harness reproduces 3 of 3 on a fresh browser; the one-off does not reproduce and is not trusted.

#### What follows for shimp, which is a design consequence rather than a bug

The handler must treat **every arriving link as untrusted** and must never act on one merely because it arrived. The comfort of "somebody must have clicked" is not available.

Much of this is already the design: identity over verbatim bytes (s44), refuse rather than normalise (s35), the keeper owning one instance per identity (s20). What changes is the **threat model's entry assumption**, and the `chrome-no-gesture` row now holds that fact where it cannot be forgotten.

#### The centralised harness, as it now stands

**15 routes, 13 of them unattended.**

| | |
|---|---|
| VERBATIM | `shellexecute`, `cmd-start`, `powershell-start-process`, `python-startfile`, `python-webbrowser`, `excel-hyperlink-object`, `chrome-link-click`, `chrome-no-gesture` |
| ABSENT, correctly | `explorer`, `chrome-control-no-permission` |
| REFUSED / UNRUNNABLE | `curl`, `wget`, `excel-hyperlink-formula` |
| still human | `run-dialog`, `outlook-email` |

The two remaining are human **by nature**: one needs keystrokes into a system dialog, the other needs a real email to arrive. Both are named in the report rather than skipped.

#### Three things that make the Chrome driver honest

* **it reports Chrome's navigation events**, so "nothing arrived" is never ambiguous between a click that missed and a launch that was refused. Conflating those is exactly what produced the wrong conclusion.
* **the control is a ROW IN THE TABLE**, not a thing done once by hand. `chrome-control-no-permission` proves on every run that the click lands, by navigating an ordinary link on the same origin.
* **no machine state.** The permission is seeded into a throwaway `--user-data-dir`. Chrome's `AutoLaunchProtocolsFromOrigins` policy would also have worked and was deliberately not used: a probe should not rewrite the user's browser policy to make its own measurement pass.

#### And the wsConn question is settled by not arising

s63 said the browser route needed `remote_chrome_cdp`'s `wsConn` extracted as a leaf. It does need a websocket, and the answer is **Python**, which already has one and was already the pattern here for Excel. No Go extraction, and `ws.go`'s own header still stands: the two Go websocket clients differ deliberately.

### s65. linkprobe consumable by a test suite, and the duplication check that can see past a rename

Built 15 Oct 2026, from the question "it may not be a test suite, but can it be used by one, and do we reuse its tooling?"

The answer to the second half was **no, and there was undeclared duplication**.

#### ⚠️ What the question found

`shimp/live_test.go` and `linkprobe/main.go` both drive the same five steps:

| step | `live_test.go` | `linkprobe/main.go` |
|---|---|---|
| build the real CLI | l.136 | l.175 |
| install an isolated scheme | l.140 | l.183 |
| trigger | `dispatchURI` | driver per `Kind` |
| poll the handler's log | l.176-186 | `arrivalSince` |
| `uninstall -purge` | l.206 | deferred |

⚠️ **And `repocheck` could not see it, for TWO reasons**: it keys on the function NAME, and it skips `_test.go` entirely. s60 had already written down the first one, calling name-matching a thing that *understates* duplication. Nothing was done, and it bit again.

#### ⚠️ Sharing it would be WORSE, which is the real finding

A Go test-only import still lands in `go.mod` and `go.sum`. s55's maxim is that **`go.sum` enumerates the closure and IS the audit's stopping condition**.

So a shared harness package would put **test tooling in the audit surface of the thing being shipped**. Every auditor pasting a `shimp://` link would find it. That is a cost paid by every user of the protocol to save one copy of five steps.

⚠️ **So the process boundary is the correct seam rather than a compromise.** `-json` is how a test suite consumes the probe, with no import and no `go.mod` entry anywhere.

The two are also not substitutable. `live_test.go` tests **install correctness**: registry keys gone after purge, a running binary refusing to delete itself, the stamp recording its scheme. `linkprobe` tests **delivery across fifteen routes**. Same mechanism, different questions, so both earn their place. It is the mechanism that is duplicated, not the purpose.

#### 1. `linkprobe -json`

One object on **stdout**, the human report moved to **stderr**, both always produced. A consumer gets the object; a human watching the same run still sees why.

⚠️ **Exit 0 even with findings**, and that is the decision a consumer most needs to know. A finding is a measurement, not a crash: `explorer` is *expected* to drop the link and `chrome-no-gesture` is *expected* to deliver it. Non-zero is reserved for the probe failing to take a measurement at all. So a suite asserts on `findings` or a named route's `got`, never on the exit code.

The JSON type is **separate from `Result`**, not tags on it: `Result` embeds the whole `Route` including its prose, and a consumer must not break when somebody rewords a comment. Two tests pin the shape, one of them listing every field name, because renaming a field silently breaks every consumer and nothing else here would notice.

#### 2. `repocheck` can now see past a rename

`SharedBodies` inverts the key: group by **body**, report when one body appears in two or more modules, and name the functions it is hiding behind. It also **reads `_test.go`**, which `Overlaps` does not, because a copied helper is likelier to live in a test than in shipped code.

A floor of 5 significant lines, comments excluded, because without it the report is pure noise: counting comments would let a function clear the floor on the strength of its documentation.

#### ⚠️ And it caught two things immediately, one of them mine

1. **`itoa` in `chrome_sessions` and `repocheck`** -- I had hand-rolled one in the new file while `strconv.Itoa` exists. ⚠️ `Overlaps` had never reported it because the other copy lives in `caveats_test.go`, which it skips. So reading test files was vindicated on the first run. Mine is deleted.
2. ⚠️ **The check did not work at all, and its own companion test found out.** `bodyOf()` starts at `func <name>(`, so the NAME sat inside the hashed text and two differently-named copies hashed differently. **A check built to find renamed copies could not find one.** It would have passed every run, green, for exactly the reason three earlier checks in this repo passed: it found nothing because it could not see. `afterSignature` strips the declaration line.

That is the fourth time a check in this repo has been blind, and the second time the companion test is the only reason anybody knows.

#### 3. A declared inventory for what nothing can find

Two kinds of duplication, and only one is mechanically findable:

| | |
|---|---|
| copy-paste-and-rename | `SharedBodies` finds it |
| the same RECIPE twice | **nothing** finds it |

The `live_test`/`linkprobe` cycle is the second kind: the same sequence of steps in genuinely different code. No body hash matches, no name matches, and no cheap structural check tells it from two unrelated functions doing similar work. Claiming otherwise would give a check that is green because it is blind.

So `known_recipes.go` writes it down. ⚠️ And because a written list rots, every entry carries **markers** that must still be present in each named file, with `TestEveryRecipeIsStillReal` failing when an entry goes stale. Same discipline as `TestBaselineHasNoStaleEntries`, which caught seven fictional entries on its first run.

⚠️ **It worked immediately**: the websocket entry claimed `remote_ai/ws.go` and the test refused to read it. remote_ai's client is inline in `main.go`. The entry was written on the strength of the *other* site's filename.

Two recipes declared: the install cycle, and the websocket pair that s60 noticed and never recorded.

#### ⚠️ The Python gap bit within one turn, 15 Oct 2026

This row widened the duplication check to see past a rename and noted that `repocheck` reads Go only. s66 then added 543 lines of Python to `linkprobe`, including a third copy of serve-a-page-and-click, and none of it was visible.

⚠️ **The identical bug recurred too.** `afterSignature` strips to the first `{\n`, and Python has no braces, so the widening did not work for Python either, for the same reason it had not worked for Go. The same companion test caught it both times.

The lesson this row drew, that writing down a blind spot and not acting is how it bites again, held on a one-turn timescale rather than a one-month one.

### s66. The second axis: link SHAPES and the handler's verdict

Built 15 Oct 2026, prompted by a question with four question marks: *do you not have what I have just done in the testing suite?*

The answer was no, and measuring the gap was more useful than the feature.

#### ⚠️ What a human did that the harness could not

| | by hand | `linkprobe` before |
|---|---|---|
| link shapes sent | **6** | **1**, a single template |
| parser refusals checked | **2**, with reasons | ⚠️ **zero** |
| identity hash and derived argv read | yes | ⚠️ never, only sent-vs-received bytes |
| handler built `windowsgui` | yes | ⚠️ never, console-subsystem only |
| profile | the real one | throwaway only |

`linkprobe` was **15 routes x 1 link**: it measured TRANSPORT. So s37's forty-character rule and s45's one-spelling rule had only ever been unit-tested, and the first real evidence they work through a browser came from a human clicking.

⚠️ And the word `Refused` in the route table means *the route declined to try* (curl: "protocol not supported"), **not** *the handler refused the link*. A reader would assume the second. The harness could not express it at all.

#### The corpus: one route, many shapes

⚠️ **One route, not fifteen**, and the reason matters: the handler's verdict is a function of the bytes it receives, and `routes.go` already establishes which routes deliver bytes faithfully. Crossing the tables would run ninety combinations to learn nothing the two axes do not say separately. Eight dispatches down the baseline is the whole cost.

All eight behaved as declared on the first run:

| shape | verdict | |
|---|---|---|
| `plain` | ACCEPTED | identity over 124 bytes |
| `argv-and-action` | ACCEPTED | the ChatGPT shape; argv derived exactly, action NOT in identity |
| `action-with-arguments` | ACCEPTED | the `?` case, 135-byte identity under a 26-byte action |
| `long` | ACCEPTED | 331-byte identity, past Excel's formula threshold |
| `bad-sha-39` | HANDLER-REFUSED | *"must be exactly 40"* (s37) |
| `double-encoded-colon` | HANDLER-REFUSED | *"second spelling of the same value"* (s45) |
| `control-byte` | HANDLER-REFUSED | *"a NUL truncates a Windows command line"* |
| `unknown-runner` | HANDLER-REFUSED | *"unknown runners fail closed"* (s43) |

The last two were **predictions**, not observations, and the table says so per shape. Nobody has clicked them in a browser; they hold through `ShellExecuteW` and nowhere else.

#### ⚠️ The one mistake that would have inverted everything

A VALID link still ends in a refusal today, because the keeper is not built:

```
click REFUSED: the KEEPER is not built. The link is valid and its identity is above
```

So acceptance is decided by **the presence of an identity line**, never by the absence of the word REFUSED. Keying on "REFUSED" would classify every correct link as rejected and every rejection as correct: not merely blind but exactly inverted. `TestJudgeCatchesARejectedShapeThatActuallyParsed` is the companion that pins it.

And a refusal must match a **declared reason**. Asserting only that a link was refused would pass for a link refused on unrelated grounds.

#### ⚠️ The finding the harness could never have made

The human clicked six links and saw **nothing at all**. Built `-H=windowsgui`, every click silent, every click correctly refused.

He could not have distinguished that from: the scheme unregistered, the handler crashing before `_log` existed, Chrome blocking it, or the link being malformed. ⚠️ **That is this repo's cardinal sin, silent failure, appearing in the PRODUCT rather than the tooling.**

It is a real tension, not an oversight. Measured: console-subsystem gives a visible foreground console on every click, which is intrusive and which s50 deliberately has no field for; `windowsgui` gives silence. **Both are wrong for a refusal.** A refused click needs a third thing: a tray balloon, a toast, a window that appears only on error.

`-gui` is now a flag on the harness rather than a constant, so both cases can be measured. The design question stays open and is recorded here rather than rediscovered.

#### One table, two consumers

`-serve` renders the SAME shapes as a clickable page. A throwaway page of hand-written links was built one turn earlier and had already drifted from what the harness checked. Now a human clicking and a machine dispatching read one table by construction.

#### ⚠️ And the hard-kill hole, measured

`-serve` purges on Ctrl-C through `signal.Notify`. Nothing survives being killed outright, and then the temp build directory is gone too, so the user has a registered scheme and no obvious binary to withdraw it with.

⚠️ The binary install PLACED in the home cannot do it either: Windows will not delete a running executable, so `uninstall -purge` from there withdraws the registry claim and then refuses to remove its own directory, which is exactly what `TestPurgeCannotDeleteItselfAndSaysSo` pins. `-purge` builds a fresh copy in a temp directory and sidesteps it. Measured both ways.

#### repocheck now reads Python

⚠️ Flagged one turn earlier and not acted on, then it bit: 543 lines of Python in `linkprobe` doing real work, a third copy of serve-a-page-and-click, none of it visible to a check whose purpose is noticing duplication.

**And `afterSignature` broke again, in exactly the same way, one commit after being fixed.** The rule was "strip to the first `{\n`", and Python has no braces, so every Python signature stayed inside the hash and two renamed copies hashed differently. The same companion test caught it the second time.

It reports nothing today, because both Python files live in one module and only cross-module sharing is reported. Covered for when that changes, which is the point.

#### ⚠️ It drifted the day the handler changed, 17 Oct 2026

This row built the corpus against a handler that ALWAYS refused. s67 then made a one-shot RUN, and nobody updated the corpus.

It kept passing, 54 seconds slower, for a changed reason: four shapes were exercising the CONSENT GATE rather than the keeper refusal, `waitForVerdict` waited only for `click REFUSED:` and polled out, and `readHandler` could not see the new `ABORTED` line at all. The verdict came out right only because it keys on the identity line.

⚠️ s71 fixed it and added the invariant this row should have had from the start: **nothing may EXECUTE during an unattended run.** Nobody is at the keyboard, so anything that ran got through the gate without a human.

### s67. A refused click is VISIBLE: the lazy console, wired at last

Built 15 Oct 2026, from two reports by the user, both correct:

> *"nothing seems to be happening"* — then — *"all I see when I click is a command prompt flash"*

| build | what a human saw |
|---|---|
| `-H=windowsgui` | no window at all, **total silence** |
| console-subsystem | a window appears and **shows NOTHING** |

⚠️ **The second is worse than the first.** A window appearing implies something happened and then tells you nothing, because `_log` writes to the file and falls back to stderr only when it cannot.

So a user could not distinguish a refused link from: the scheme unregistered, the handler dying before `_log` existed, Chrome blocking the launch, or a malformed link. ⚠️ **This repo's cardinal sin, a silent failure, in the thing being shipped rather than in the tooling.**

#### ⚠️ And the answer was already decided, written down, built, and never called

`console_windows.go`'s header, from s50:

> *"can a keeper start with NO console and allocate one only when it turns out to have something to show? The answer is yes, and it is AllocConsole ... no window for a click that has nothing to say, and a window the moment something does."*

`allocConsole` was written for exactly this and **nothing invoked it**. Reasoned, recorded, implemented, unwired. ⚠️ That is the third instance of the same pattern inside two days, after s60's blind spot and s65's Python gap.

#### The asymmetry is the design

A refusal is an error the user must see. A success has nothing to say. So s50's refusal of a descriptor field for the console gets **stronger**: nothing is declared in advance, and the window appears only where there is something to read.

Confirmed by the user, verbatim:

```
  shimp-linkprobe:// REFUSED

  the KEEPER is not built, so there is nothing to hand this to yet. The
  link is VALID: identity 08f7fefa65fd4929... over 135 bytes.

  the full record is in C:\Users\david\.shimp-linkprobe\shimp.log

  press any key, or this closes in 25s
```

⚠️ The message says **VALID** deliberately. A keeper refusal otherwise reads exactly like a malformed link, and the user goes away and edits a correct URI.

#### Four things that had to be got right, each measured

* **a timeout AS WELL AS a keypress.** Keypress alone hangs an unattended run; a timer alone makes a human stare at something already read.
* ⚠️ **the input buffer is FLUSHED first.** A freshly allocated console already holds focus and mouse events, and treating one as the keypress would close the window before it was read, which is this row's own bug one layer down.
* ⚠️ **`SHIMP_NO_HOLD` suppresses the WAIT, never the MESSAGE.** A variable that hid the reason would be the opposite of the purpose. Measured: it propagates through `ShellExecuteW`, so `-corpus` takes **28s for 8 shapes** rather than 200. It does NOT propagate through a browser click, where the handler is a child of Chrome, so `chrome-link-click` went from ~40s to **65s** and still passes with no orphan consoles.
* **polled, not blocking.** `ReadConsoleInput` has no timeout, so a blocking read would ignore the deadline entirely.

#### ⚠️ The test passed by ACCIDENT, which was worse than failing

`TestClickLogsTheLengthS36CannotOtherwiseMeasure` already drives `onClick` down a refusal path. The real `showRefusal` allocates a console and waits 25 seconds, so it should have blocked the repo's own loop.

It did not, and the reason is luck: **`AllocConsole` happens to FAIL under `go test`** because the process is already attached to one, so the early return made the suite fast by accident of environment. Run in a plain console it would have hung.

⚠️ A test that passes by luck is precisely the failure shape this repo is organised against. So `showRefusalFn` is now a seam, which buys two things: no console in tests, and **both refusal paths become assertable**. AGENTS puts it the sharper way round: where there is no seam, the absence is usually the design problem the test was about to find. Here it was two.

#### And one more caught by vetting another platform

`wrap` and `splitWords` were written inside `refusal_windows.go` and inherited its build tag by proximity. `GOOS=linux go vet` failed on a platform-independent test referencing them. Nothing about breaking text at word boundaries is Windows-specific; they are in `text.go` now. Worth keeping cross-platform vet in the loop for exactly that.

#### ⚠️ A separate correction: the keeper is NOT what blocks the first real use case

Asked when a shimp link could ensure a CDP Chrome and then run a ChatGPT `full_update`, the answer required withdrawing something said repeatedly in this slate.

**Both of those are ONE-SHOTS.** Ensure-CDP is idempotent and exits; `full_update` syncs and exits. s20's keeper exists to guarantee one RESIDENT instance and deliver actions to it, which is s46's resident class. A one-shot click needs none of it.

⚠️ What it needs is the thing shimp has never had:

```
=== does shimp have ANY exec machinery? ===
NONE -- the handler has never run anything
```

It logs `would run [...]` and then does not. **That** is the gap for a one-shot capability, and it is roughly fifty lines rather than s20's lock, IPC and handshake.

⚠️ **And the day it lands, s64 bites.** A permitted origin can fire a link with no gesture, measured. Today that is harmless because nothing runs. The moment exec exists, a visited page can make the machine run `go run <module>@<sha>`. The consent mechanism has to be designed in before exec, not after, and this row's lazy console is the obvious candidate: a click about to RUN something shows what it will run and waits.

### s68. A hyperlink RAN something. And the consent gate was broken until somebody ran it

15 Oct 2026. The protocol did its job for the first time.

```
15:58:09  click: shimp-try://.../chrome_ready@4252cb91.../-port/9333/-profile/demo_profile
15:58:09  click: identity 9eed99a9074c66ef... over 151 bytes
15:58:09  click: would run ["go" "run" ".../chrome_ready@4252cb91..." "-port" "9333" "-profile" "demo_profile"]
15:58:13  click: consent given by a keypress on a console we own
15:58:13  click: running [...]
15:58:17  click: exited 0
```

A Chrome appeared on 9333 with a dedicated profile. ⚠️ **And it survives its launcher**, measured separately on 9555: `chrome_ready` exits, the browser stays. Without that, "ensure a CDP Chrome" would have been theatre.

#### What that single click exercised

| | |
|---|---|
| s7, s8 | the registry claim and the Office trusted-protocol key |
| s44 | identity over 151 bytes, the fragment excluded |
| s43 | the runner table: two segments consumed, the rest ARE the argv |
| s55 | **all three untested parts**: a module in a subdirectory of a multi-module repo, resolved by BARE SHA, fetched by the proxy with no checkout anywhere |
| s49 | `go run module@sha` cold, measured again at 24s |
| s46 | the one-shot class, which needs no keeper |

s55's proof push had been queued for days. The use case forced it, and it passed.

#### ⚠️ THE CONSENT GATE WAS BROKEN, AND ONLY RUNNING IT FOUND OUT

Asked to test before clicking, the first attempt dispatched a link, **pressed nothing**, and the handler logged *"consent given by keypress"* and launched a Chrome. **In the same second as the dispatch.**

Two faults, the first structural:

1. ⚠️ **`allocConsole` RETURNS EARLY IF A CONSOLE ALREADY EXISTS.** A handler launched from a shell INHERITS that shell's console, input buffer and all. Consent was read from the buffer holding the Enter key that had started the command. The user's answer was never involved.
2. `STD_INPUT_HANDLE` follows whatever the process inherited, so even on a fresh console it is not reliably that console's input.

⚠️ This is the worst defect this program has had, and it failed in the one direction that cannot be tolerated: it **granted** permission that was never given.

#### The fix is structural

```
FreeConsole    detach from anything inherited
AllocConsole   create one that is OURS, input buffer empty
CONIN$         open the console's own input, never a std handle
settle+flush   discard the arrival burst a new console emits
minSettle      reject any key arriving too fast to have been typed
```

And `isKeyDown` now checks `EventType == KEY_EVENT` properly. A console emits focus, mouse, menu and buffer-size events, and a fresh one emits several at once; any of those counting as a keypress is how the first version authorised a run.

**Measured both directions, which is the whole point:**

| | |
|---|---|
| no keypress | `ABORTED` after 31s, nothing on 9444 |
| a real keypress | `consent given by a keypress on a console we own`, 4s after the prompt |

⚠️ The second could not be tested by me at all: `SHIMP_ASSUME_CONSENT` takes a different branch, so it proves nothing about reading a key. It needed a human, which is exactly what the question *"can you run it first"* produced.

#### ⚠️ The lesson, and it is not "write more tests"

The gate had a seam, a test file and a careful comment explaining that the timeout must abort. **None of that caught it**, because every test exercised the decision logic and the bug was in where the input came from.

What caught it was dispatching a real link on a real machine and pressing nothing. The repo already says this, in AGENTS, about fakes: *"it can never show that the model matches reality, because the fake was written from the same belief the code was."* A seam is a fake of a console. The belief it encoded was that the console was ours.

#### And a second thing found while fixing the first

The hold check lived only in `showRefusal`, so the ABORTED path in `runLink` called `waitForKeyOrTimeout` directly and sat for 25 seconds. The suite went from 3s to 61s, and the variable meant to prevent exactly that had no effect there. `waitForKeyOrTimeout` honours it itself now: **13s**.

⚠️ A guard every caller must remember is a guard one caller will forget.

#### What this does NOT yet do

* **no action delivery.** A link carrying an action other than `ensure` is refused by name, because that needs the keeper.
* **no second capability.** `chatgpt_tools`' `full_update` is 8200 lines of Python in another repo, so s55 maxim 1 means a `go/run` link cannot name it. The decision recorded in s67 stands open: may shimp run code its audit does not cover, if it declares the boundary loudly?
* ⚠️ **s64 is now live, not theoretical.** A permitted origin can fire a link with no gesture, and something now runs at the end of it. The keypress is the only thing between a visited page and `go run <module>@<sha>`, which is why it is in from the first commit that could execute anything.

### s69. probes/: the instruments promoted, and the first re-run retired a quoted number

16 Oct 2026, from the question *"have we captured a lot of Python from your scratchpad and are we using it?"*

The answer was **no**, and git settled it: three `.py` files had ever been committed, two of them drivers written directly into `linkprobe` and one the pre-existing `update_agents.py`. Nothing had been promoted.

#### ⚠️ The instruments behind the numbers were in a temp directory

| instrument | the number it produced |
|---|---|
| `probe_2084.py` | "the object caps at 2084 and REFUSES at 2085" |
| `measure_url_limits.py` | "=HYPERLINK() truncates at 255 silently" |
| `click_diag.py` | the navigation-event diagnosis that overturned s63 |
| `focusprobe` | yesterday's foreground measurements |

Those figures are quoted in `linkprobe/routes.go`, in the slate, and in two notes delivered to OTHER repos. ⚠️ **The numbers survived. The instruments did not.** Every one had become a claim nobody could re-measure without rewriting the tool, which is the position s59 and s66 exist to prevent.

`probes/` now holds them, with the rule stated in its README: **a number quoted as measured must name the instrument that produced it.** If there is no instrument, the number is a recollection.

#### ⚠️ AND THE FIRST RE-RUN RETIRED A NUMBER

`probes/excel_hyperlink.py`, on the same Excel build:

```
Hyperlinks.Add       2084 exact round-trip, 2085 REFUSED   <- reproduced exactly
=HYPERLINK() stores  255, 256, 326, 1000, 2084, 2085, 4000  all UNALTERED
                     6000, 8000, 32000                      REFUSED, com_error
```

So the 2084 boundary is confirmed by a freshly written instrument, which is exactly what this directory is for.

⚠️ **And "=HYPERLINK() truncates at 255, SILENTLY" is wrong in BOTH halves.** It does not truncate at 255, and where it fails it REFUSES rather than truncating. The boundary sits between 4000 and 6000, consistent with Excel's 4096-character formula limit and nothing hyperlink-specific being involved at all.

⚠️ That claim was recorded on 7 Oct and quoted for nine days, including outward to `slate_tools` and `dataannotation_tools`. Both notes now carry a correction naming the instrument.

**It does not rescue the formula route**, and the reason it was always unusable is unchanged and stronger: `Hyperlinks.Count` is ZERO after writing `=HYPERLINK()`, so there is no object to follow and only a human click completes it. That objection never depended on 255.

#### What each instrument owes, and why

* **its own result in the docstring, with the date.** A probe whose output nobody recorded gets re-run to learn what is already known.
* **a control.** Twice in one session a control caught a false positive that would have been written down as a finding.
* ⚠️ **the distinction between measured and inferred.** `cdp_click.py` reports Chrome's navigation events rather than success or failure, because three outcomes look identical from outside and conflating two of them produced a wrong conclusion.
* **no machine state left behind**, and `DispatchEx` rather than attaching: the original ran against the session's Excel, and attaching sets `DisplayAlerts = False` on somebody's live application, which outlives the script.

#### ⚠️ Three defects in the instruments, found by running them

1. **A probe died printing its own conclusion.** The console is cp1252 and a warning glyph raised `UnicodeEncodeError` AFTER the measurement had been taken. A probe that crashes while reporting buries the number it was run for. Both streams are reconfigured to UTF-8 now.
2. **The formula loop let an exception escape**, so the run died at 8000 and discarded the six results it already had. Each case is caught separately: an instrument must survive the interesting case.
3. ⚠️ **The focus probe printed a STALE ERRNO as a reason.** `SetForegroundWindow` returns 0 for "refused" and does NOT call `SetLastError`, so it reported "Invalid window handle" for a handle Windows had just confirmed valid, and two readings were written down as *"REFUSED, and it said so"* when it had said nothing. The return value and the before/after foreground are the whole evidence. **A stale errno dressed as a reason is worse than no reason, because it gets quoted.**

#### ⚠️ And repocheck could not see any of it

`modulesIn` meant "directory with a go.mod", which missed a Python-only directory entirely and missed `probes/focus` for being one level down. ⚠️ **The third instance of the same blind spot in two days**: SharedBodies could not see test files, then could not see Python, and then discovery could not see a directory holding only Python. Each was found after adding the code it could not see.

A directory now counts if it holds a `go.mod` OR any `.py`, and immediate subdirectories are checked for a `go.mod` too.

#### ⚠️ THE INTERESTING PART: name-matching is the WRONG check for Python

Widening `Overlaps` to read Python reported **eleven** new overlaps, of which **one** was real. The rest were names Python does not let you choose:

| | |
|---|---|
| `do_GET`, `log_message`, `handle_one_request` | required overrides of `BaseHTTPRequestHandler` |
| `_log`, `safe_local_imports` | mandated by AGENTS |
| `__init__`, `send`, `close`, `drain` | ordinary method names |

⚠️ A name-keyed check assumes the name was CHOSEN. In Python it is frequently dictated, by a framework, a convention or the language. So the check that **understates** duplication in Go massively **overstates** it in Python, and an allowlist of eleven forced names is exactly the rubber-stamp `known_overlaps.go`'s own header warns about.

So `Overlaps` stays Go-only, deliberately, and Python is covered by `SharedBodies`, which keys on the BODY and does not care what anything is called. **That is the right key for a language where the name is not a choice, and it found the one true duplicate among the eleven**: `CDP.close()`, byte-identical in two clients.

Two further corrections fell out of it: `SharedBodies` now flags single-name shared bodies, because the old condition assumed `Overlaps` covered them and it does not; and it does NOT flag a body whose names are already acknowledged, because demanding two declarations for one fact is how an allowlist becomes a formality.

#### Declared rather than collapsed

`probes/cdp.py` and `linkprobe/chrome_route.py` both implement a CDP client. ⚠️ `chrome_route.py`'s is **not importable**: it binds `websocket` inside `safe_local_imports`, so using its class raises `NameError`. Collapsing it means changing committed, tested code in `linkprobe`'s loop, and `probes/` exists to stop losing instruments rather than to refactor. Declared as a recipe, with the note that both the recipe and `acknowledgedBodies["close"]` should go the day that client becomes importable.

### s70. The seven top-level things, and the state of each

Mapped 16 Oct 2026, from a proposed four-part taxonomy: the handler, the leaf modules, the shimpable tools, the Python testing code. Measured against the repo, four of seven, and one of the four mislabelled.

#### The seven, and they live in their OWN SLATE now

⚠️ 17 Oct 2026. The table that was here has moved to the **`repo` slate**, one row per category, because seven things that evolve independently are seven rows rather than one design cell.

This row keeps the REASONING, which is what it was actually for:

* a four-part taxonomy was proposed: handler, leaf modules, shimpable tools, Python testing code
* **four of seven**, and one of the four mislabelled
* ⚠️ "the Python testing code" undersold a 2,345-line Go harness and swept `repocheck` into the wrong box. It splits into the LINK HARNESS, which drives routes and shapes, and the INSTRUMENTS, which measure
* ⚠️ the two missing were the two that matter most: the **audit surface**, which is the point of the project and is barely built, and the **design record**, which is the largest artefact here

⚠️ And the state belongs in a table with a GAP column, not in prose. Three of the seven are nearly empty, and a flat list reads as seven healthy pillars. That is the whole reason the `repo` slate has that column.

### s71. The harness drifted when the handler changed, and the click window is now a choice

17 Oct 2026, from two observations the harness could not make about itself.

#### ⚠️ "Everything takes 30s per hit"

The corpus had gone from 28s to 82s. I blamed the harness polling for a line that never came. **Wrong: the HANDLER was waiting.**

s67 gave a one-shot a consent gate. Four of the eight shapes are one-shots, nobody is at the keyboard during a corpus run, and each sat for the full thirty seconds. ⚠️ `SHIMP_NO_HOLD`, which the harness already set, only suppresses the hold AFTER something is shown. It does nothing to the consent window.

`SHIMP_CONSENT_SECONDS` shortens the wait and **cannot grant consent**: it still aborts, so the real branch runs and nothing executes. Deliberately not `SHIMP_ASSUME_CONSENT`, which skips the gate and then runs. **Shortening a refusal is safe; shortening a decision is not.**

#### ⚠️ AND THE CORPUS HAD DRIFTED, PASSING FOR A CHANGED REASON

This is the part worth keeping. s66 wrote the corpus against a handler that ALWAYS refused. s67 made a one-shot run. Nobody updated the corpus, and it kept passing:

* `waitForVerdict` waited only for `click REFUSED: `, so the four one-shots polled the full 20s before falling through to a late read
* `readHandler` could not SEE the new `click ABORTED: ` line, so it could not tell a keeper refusal from a consent abort
* the four accepted shapes were silently exercising the **consent gate** instead of the keeper refusal they were written against
* and the verdict still came out right, because it keys on the identity line, which survived the change

⚠️ **A harness that passes 54 seconds slower is a harness telling you something and being ignored.** It now waits for any terminal line: REFUSED, ABORTED or exited.

#### ⚠️ THE NEW INVARIANT, which matters more than any shape

`judge()` now fails ANY shape where `click: running` appears in the log.

Nobody is at the keyboard during an unattended run, so consent cannot have been given. Anything that executed got through the gate without a human. ⚠️ That is the most serious thing this harness could discover, and **before today it had no way to notice it**. Measured: zero.

#### The window is a choice: focus, background or none

A boolean could only say window or no window, so it had no way to express the case a human watching a run actually wants.

| | |
|---|---|
| `focus` | **default.** A console appears and TAKES the foreground. What a real click should do: the user clicked, so the thing should come to them |
| `background` | appears, does not take focus. ⚠️ **best effort**, and says what it managed |
| `none` | no console. Consent cannot be asked, so the click ABORTS. A refusal is logged, not shown |

⚠️ None of the three can cause anything to RUN. `none` fails closed by construction. An unrecognised value falls back to `focus` **and says so**: a typo must not turn a visible click into an invisible one.

`linkprobe -window` defaults to `none`.

#### ⚠️ `background` was deleted and then resurrected, and the reason is the lesson

It was built, measured at 5 of 8 dispatches keeping the focus, and deleted as half-working. ⚠️ **Nothing was wrong with the mechanism.** What was wrong was that it was the SILENT DEFAULT and the code implied it always worked.

`AllocConsole` creates and activates the window; there is no flag, ordering or prior call that prevents it, so the only move is to put it back afterwards and that is a race. It now has the fallback it lacked: `SW_MINIMIZE`, deterministic because a minimised window cannot be foreground, tried LAST because it hides output.

A mechanism that works most of the time is fine when it is CHOSEN and says what it achieved. The same mechanism as a silent default is a lie.

#### ⚠️ AND THE ACTUAL CAUSE OF THE SPLATTER WAS DUPLICATION

The user reported command prompts still covering the desktop after I had "fixed" it and measured the fix. Both were true: **`-run` was never fixed.**

`doRun` carried its **own copy** of `buildAndInstall`'s logic, without the `-ldflags`, and `probe()` set no environment at all. So fifteen routes each dispatched to a console-subsystem handler with nothing suppressed. The corpus was fixed and measured; `-run` was not; the duplicated build hid the difference.

⚠️ **This is what `repocheck` exists for and it could not see it.** Two call sites where one gets a fix is the exact failure mode, but the duplication was a RECIPE -- the same sequence written twice -- rather than a shared body, which is the category `known_recipes.go` was created for and which this pair was never declared in. s65 wrote that nothing finds a recipe mechanically. Here is the cost of that being true.

Collapsed onto one `buildAndInstall`, and the environment onto one `quietEnv()` every child path calls.

#### Measured

| | |
|---|---|
| corpus, 8 shapes | 82s → **12s**, then 5s warm |
| unit tests, 9 modules | 27s |
| repocheck | under 1s |
| **the whole quiet suite** | **32s** |
| focus | ⚠️ **held**: same window identity before and after |

#### ⚠️ Two limits, stated rather than discovered

* **Excel and Chrome will always show windows.** The Python drivers launch them and AGENTS forbids starting either invisibly, because a hidden window holding a modal is a process nobody can diagnose. 4 of 15 routes. `-window` covers the handler, which is what was multiplying fifteenfold.
* ⚠️ **The conhost count is not a usable signal.** Sampled four times with nothing running: 15, 15, 15, 18. Every measurement call spawns its own console, so the measurement perturbs the thing measured. The reliable signal is the foreground window's IDENTITY, not a process count -- which is the third time today my own tooling produced a false reading, after the hwnd arithmetic and the stale errno.
