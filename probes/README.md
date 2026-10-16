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
