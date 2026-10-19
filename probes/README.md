# probes -- the instruments behind the numbers

Every measured claim in this repo should have a program here that can take the
measurement again.

## Why this directory exists

The numbers below are quoted in `linkprobe/routes.go`, in
`repocheck/known_overlaps.go` and across the `shimp` slate, and they are
load-bearing: the route table's expectations and several frozen decisions rest
on them.

They were produced by scripts in an agent's temp directory. ⚠️ The numbers
survived, in the slate. The instruments did not. So every figure had become a
claim nobody could re-measure without rewriting the tool, which is exactly the
position s59's attestations and s66's corpus exist to prevent.

## The rule

**A number quoted as measured must name the instrument that produced it.**

If there is no instrument, the number is a recollection. Say so, or write one.

## What is here

| instrument | what it measures |
|---|---|
| `excel_hyperlink.py` | Excel's two hyperlink routes and their real length limits |
| `cdp_click.py` | whether a CDP-synthesized click lands, and what Chrome did with it |
| `cdp.py` | the shared CDP client the Python instruments use |
| `focus/` | foreground-window entitlement: who may raise what, and when |
| `canonjson/` | whether Go can reproduce Python's canonical JSON byte for byte |
| `canonjson/floats/` | whether Go can *originate* a float the way Python renders it |

### What `canonjson` measured, 19 October 2026

patchbucket computes a patch's integrity hash as `sha256` over
`json.dumps(patch, ensure_ascii=False, sort_keys=True)` and refuses any patch
whose hash does not match.
So this decides whether a Go peer can write a patch the Python side will accept.

| strategy | score |
|---|---|
| A: `json.Marshal` of a decoded map, `SetEscapeHTML(false)` | **0 / 32** |
| B: `json.Number` + hand-sorted keys + Python's separators | **31 / 32** |
| C: B plus a hand-written string escaper | ✅ **32 / 32** |

⚠️ **A fails on every record**, and the dominant reason is not exotic: Python's
default separators are `", "` and `": "` with spaces, Go's are `","` and `":"`,
and Go's `Encoder` has no option to change them.
Its number handling is separately wrong, because decoding into `any` turns
every JSON number into a `float64`: `1700000000.0` comes back as
`1700000000`, `1e-07` as `1e-7`, `1e+20` as `100000000000000000000`, and an
integer above 2^53 loses precision outright.

**B's single remaining failure is `U+2028`.**
`encoding/json` always escapes U+2028 and U+2029 and offers no way to stop it:
a raw line separator inside a `<script>` block terminates a JavaScript string
literal, so Go escapes it deliberately.
Python with `ensure_ascii=False` emits it literally.
Matching Python therefore requires writing the escaper, which is what C does.

**So four things are required, and none is optional:** `json.Number` for
verbatim number text, hand-sorted keys, Python's `", "` / `": "` separators,
and a hand-written string escaper.

⚠️ The verdict is **all-or-nothing**, which is why a per-case score is only
diagnostic: `patch_hash` covers the whole envelope, so one divergent value
anywhere makes the entire patch unacceptable.

### And what `canonjson/floats` measured, same day

C only proves Go can *echo* number text Python wrote.
A Go peer originating a patch from its own store holds a `float64` and has no
text to echo, so it was measured separately, from **hex bit patterns** so that
rendering is tested without a parsing question tangled into it.

| | score |
|---|---|
| Go's shortest round-trip `'g'` | 18 / 27 |
| `pythonFloat()`, three rules | ✅ **27 / 27** |

The three rules:

1. ⚠️ **A whole float keeps its `.0`.** Python prints `1700000000.0`; Go's `'g'`
   prints `1.7e+09` and Go's `json` prints `1700000000`. This is the one that
   would bite in practice, because `create_time` and `update_time` are `REAL`
   columns and a timestamp with no fractional part is ordinary.
2. The exponent carries a sign and **at least two digits**: `1e-07`, not `1e-7`.
3. ⚠️ **The thresholds are Python's**, not Go's: `repr` switches to exponent
   form at `1e16` and below `1e-4`.

ⓘ One residual, not a blocker but worth knowing: an integer above 2^53 cannot
survive a `float64` round trip, so a Go peer originating one needs `int64`
rather than a float. C preserves such values only because `json.Number` keeps
the original text.

## Running them

```powershell
& "<venv>\Scripts\python.exe" probes\excel_hyperlink.py
& "<venv>\Scripts\python.exe" probes\cdp_click.py --help
cd probes\focus ; go run . -mode timeout
```

⚠️ **Run these in a visible terminal.** `excel_hyperlink.py` starts an Excel
and `cdp_click.py` starts a Chrome. AGENTS forbids starting either invisibly,
because a hidden window holding a modal is a process nobody can diagnose, and
an agent broke that rule with `linkprobe -serve` on 15 Oct by backgrounding it
where the user could neither see nor stop it.

## What an instrument owes the reader

* **its own result, in the docstring**, with the date. A probe whose output
  nobody recorded is a probe that will be run again to learn what is already
  known.
* **the distinction between measured and inferred.** `cdp_click.py` reports
  Chrome's navigation events rather than only success or failure, because a
  null result from a harness whose control fails says nothing at all, and that
  conflation produced one wrong conclusion already.
* **a control.** Twice in one session a control caught a false positive that
  would otherwise have been written down as a finding.
* **no machine state left behind.** Scratch workbooks closed without saving,
  throwaway Chrome profiles removed, ports released.

## Known duplication, declared

`cdp.py` and `linkprobe/chrome_route.py` both implement a CDP client.
`chrome_route.py` is committed, tested and in `linkprobe`'s loop, and its
client is not importable because it binds `websocket` inside
`safe_local_imports` rather than at module level.

Collapsing them is worth doing and is not done here, because the point of this
directory is to stop losing instruments, not to refactor working code.
Declared in `repocheck/known_recipes.go` so it is a decision rather than an
oversight.
