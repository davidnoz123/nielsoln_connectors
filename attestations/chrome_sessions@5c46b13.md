# chrome_sessions @ 5c46b13

    subject       github.com/davidnoz123/nielsoln_connectors/chrome_sessions
    commit        5c46b13847551b9ae06697c79dc3cbf227c8444d
    content hash  not yet available: no consumer pins this module, so it has no
                  go.sum entry anywhere. The commit is the binding until one does.
    kind          PROGRAM, so an absolute verdict is possible (see attestations/README.md)
    checked by    Claude Opus 5, on David Nielsen's instruction
    at            14 Oct 2026

## Verdict: CONDITIONAL

Safe for its stated purpose. Three conditions below change what it discloses,
and one of them needs no flag to apply.

⚠️ Not SAFE, and the difference matters: a first attestation that came back
clean would have demonstrated nothing about the format.

## Capabilities

Copied from the module's `Caveats`, which `caveats_test.go` verifies against
the source on every `go test`. Facts, not judgement.

| surface | what |
|---|---|
| `http.Client` | HTTP GET to `127.0.0.1:<port>/json/version` and `/json/list` |
| `net.DialTimeout` | a bare TCP dial, opened and immediately closed |
| `iphlpapi.NewProc` | `GetExtendedTcpTable`, which returns the **whole machine's** TCP listener table |

## Exposure

Each line cites the code that constrains or enables it.

**It cannot reach off the machine.** `host` is a `const` at `sessions.go:51`,
not a flag. There is no input that redirects it, so every request and dial is
to loopback.

⚠️ **It reads the whole machine's socket table on every run, with no flag.**
`ownerOfPort(port)` is called unconditionally in `probe()` at
`sessions.go:114`, and `owner_windows.go` fetches the entire
`MIB_TCPROW_OWNER_PID` table to answer for one port. So the machine-wide read
is a capability of every invocation.

**But by default only one pid per probed port is disclosed.** `render()` at
`sessions.go:304` prints the `Session` values, and `OwnerPID` carries the owner
of a port you named. The rest of the table is read into this process and never
emitted. ⚠️ The distinction between *reading* machine-wide state and
*disclosing* it is the one an auditor needs, and it is easy to get wrong in
either direction.

⚠️ **`-discover` turns that capability into a disclosure.** `discoverPorts()`
at `sessions.go:272` enumerates listening ports from the same table, so the
output then reflects machine-wide state: which ports are listening and which
process owns each. Piped anywhere, that is a machine inventory.

⚠️ **It makes unsolicited HTTP requests to whatever loopback port it is given,
and prints what comes back.** `getJSON` at `sessions.go:142` fetches
`/json/version` and `/json/list` and the result reaches stdout. Point it at a
non-Chrome service and that service's response at those paths is printed.
s12 records this happening: `-discover` probed 135, 139, 443 and 445, which is
RPC and SMB, and one answered by forcibly closing the connection.

**Tab titles are decoded for display only.** `sessions.go:233` unescapes HTML
entities in the human view; the JSON keeps Chrome's own string. So the printed
form is lossy relative to the data, deliberately, and a caller parsing output
should use `-json`.

## Conditions for a SAFE reading

1. `-discover` is not used, or its output is not retained anywhere
2. the ports named are ones you expect a browser on, since anything listening
   there will receive two unsolicited GETs
3. the output is not treated as evidence of authority. `Tabs` is a hint about
   what a browser has open and never a claim about who it is signed in as (s12)

## Not reached, and verifiable

From `CaveatsNotReached`, each checkable against the source:

- **execs nothing.** No `exec.Command` anywhere, asserted by `caveats_test.go`
- **listens on nothing.** Client only; no inbound socket exists
- **writes no files.** Output goes to stdout; `-json` changes the format, not the destination
- **persists nothing.** No registry key, no scheduled task, no startup entry
- **no non-loopback address**, per the `const host` above

## What this attestation does NOT establish

- ⚠️ **the Go toolchain is not pinned** by a commit, so a different compiler
  could produce a different binary from this source. s34, and the reason s24
  has UNVERIFIED as a verdict
- raw syscalls, reflection, and anything outside `CaveatSurfaces` are
  UNVERIFIED rather than absent. The `iphlpapi` line above is declared because
  a human noticed it, and the checker can see it only by accident of
  `iphlpapi` being a package-level var
- this is **self-issued**. The capability half is mechanically verified; the
  exposure half is a judgement, attributed and on the record
