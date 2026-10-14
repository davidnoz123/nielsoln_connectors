# attestations/

One file per module per commit, saying what that exact code can reach and
under what conditions that matters.

## ⚠️ Why these cannot live inside the module they describe

An attestation's subject is a **content hash**. Put the file inside the module
and it becomes part of that hash, so the attestation either names a version
that does not include it, or names itself and cannot be computed.

Same circularity as `caveats_test.go`'s: an expected hash cannot live in the
file it hashes. There the fix was a separate test file; here it is a separate
directory, outside every module, so adding one changes no module's content
hash.

## What a file contains

| | |
|---|---|
| **subject** | `module@<full sha>`, and the `go.sum` content hash once a consumer pins it |
| **capabilities** | copied from that module's verified `Caveats` (s54). Facts, mechanically checked |
| **exposure** | what those capabilities mean in practice, **each line citing the code that constrains or enables it** |
| **verdict** | SAFE / CONDITIONAL / UNSAFE / UNVERIFIED, per s24 |
| **attribution** | who, and when |

## ⚠️ Two rules that keep these from becoming P3P

**Every exposure line cites code.** Not "this cannot leave loopback" but "this
cannot leave loopback, because the host is the constant `host` at
`sessions.go:51`, not a flag". That makes the claim spot-checkable rather
than a verdict to believe, which is s56: nothing may be an assertion.

**A leaf's verdict can only be CONDITIONAL.** A mechanism has no stance, so it
has no safety property of its own: `cdpcore` is safe or not depending entirely
on what its caller permits. Only a program, which decides *when*, can be
certified outright. A program's audit then DISCHARGES each condition its
leaves state, which is what makes the audit compose instead of repeating.

## Why cache them at all

A pinned version is immutable, so an attestation about it **cannot go stale**.
Audit `cdpcore@4b91c7e` once and every link whose closure contains it inherits
that work for ever.

⚠️ This is not the thing s47 refused. s47 rejected a *latest-good* list,
because "which version is current" goes stale. A statement about a fixed
subject does not.

## Not a scanner result

Deliberately not VirusTotal or any antivirus verdict. Those detect known-bad
signatures in binaries and answer a different question; a clean result on a
400-line Go connector is an absence of evidence that would be read as a
certification. Go binaries also false-positive routinely, being unsigned,
statically linked and novel on every build.

And uploading anything to such a service is a **publication**. Hash lookup is
safe; upload is not, and nothing carrying a descriptor, a URI with a home
directory in it, or a token may ever be uploaded anywhere.

## Where this goes next

The format is shaped so it can move to a signed transparency log (Sigstore /
Rekor) unchanged, when there is a second party to sign. Until then these are
self-issued, which is weaker than a third-party review and stronger than
nothing: the **capability** half is mechanically verified, so only the
**exposure judgement** is a claim, and it is attributed and append-only
through git history. A wrong one is on the record against a name and a date.
