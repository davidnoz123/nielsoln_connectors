"""Measure Excel's two hyperlink routes and their real length limits.

    & "<venv>\\Scripts\\python.exe" probes\\excel_hyperlink.py
    & "<venv>\\Scripts\\python.exe" probes\\excel_hyperlink.py --boundary-only

WHAT IT MEASURED, re-run 16 October 2026
----------------------------------------
    Hyperlinks.Add .Address     round-trips to 2084, REFUSES at 2085
    =HYPERLINK() formula        stores to 4000 UNALTERED, refuses by 6000
    plain cell text             far longer, not a constraint

!! THE FIRST NUMBER REPRODUCED EXACTLY. 2084 round-trips, 2085 refuses, on
the same Excel build, from a freshly written instrument. That is what this
directory is for.

!! AND THE SECOND NUMBER WAS WRONG IN BOTH HALVES. The claim recorded on
7 Oct, quoted since in linkprobe/routes.go and the shimp slate, was
"=HYPERLINK() truncates at 255 bytes, SILENTLY". Measured here:

    255, 256, 326, 1000, 2084, 2085, 4000   all stored UNALTERED
    6000, 8000, 32000                       REFUSED, with a com_error

So it does not truncate at 255, and where it fails it REFUSES rather than
truncating silently. The refusal sits between 4000 and 6000, which is
consistent with Excel's 4096-character formula limit being the real
constraint and nothing hyperlink-specific being involved at all.

!! WHICH DOES NOT RESCUE THE FORMULA ROUTE, and the reason it was always
unusable is unchanged and stronger: Hyperlinks.Count is ZERO after writing
=HYPERLINK(), so there is no Hyperlink object to follow and the delivery half
needs a human clicking the cell. That objection never depended on 255.

!! THE BOUNDARY IS THE POINT, NOT THE NUMBER. slate_tools reported "no cap
found" on 8 Oct while this measured 2084 on 7 Oct, and both were true: a
search that never reaches 2084 finds no cap. So this reports what happens AT
the boundary -- exact round-trip, silent change, or refusal -- rather than
only where the boundary is. A route that truncates silently and a route that
refuses are different kinds of thing, and only one of them is safe.

!! AND A LATER MEASUREMENT COMPLICATED THE FORMULA NUMBER. On 15 Oct, a
326-byte URI written into =HYPERLINK() came back from cell.Formula at 326
bytes, unaltered. So whatever truncates at 255 is NOT what the formula
stores; it may be the Address property, or what a click resolves. The claim
is not disproved, its subject is unlocated. --boundary-only reproduces both
halves so the two can be compared directly.

WHY DispatchEx RATHER THAN ATTACHING
------------------------------------
The original ran against the session's existing Excel through slate_tools.
This starts its own instance instead, for two reasons measured on this
machine: eighteen Excel instances were open at one point, and attaching sets
DisplayAlerts = False on somebody's live application, which outlives the
script. A length limit is a property of the Excel BUILD rather than of the
instance, so nothing about the measurement changes.

It also makes this directory self-contained: no sibling repo is needed to run
a probe.
"""
import sys

# ⚠️ UTF-8 ON STDOUT, OR THE INSTRUMENT CRASHES PRINTING ITS OWN RESULT.
# Measured 16 Oct 2026: the Windows console is cp1252 here, and a warning
# glyph in a conclusion raised UnicodeEncodeError AFTER the measurement had
# been taken. A probe that dies while reporting is a probe nobody can use,
# and the traceback buries the number it was run for.
try:
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")
    sys.stderr.reconfigure(encoding="utf-8", errors="replace")
except Exception:
    pass


def _log(msg: str) -> None:
    print(msg, flush=True)


def safe_local_imports(g: dict) -> None:
    """Load the non-stdlib imports, logging a FATAL line before dying."""
    try:
        import win32com.client as w32
        g["w32"] = w32
    except Exception:
        import traceback as _tb
        _log(f"FATAL: safe_local_imports failed:\n{_tb.format_exc()}")
        raise


def url_of(n: int) -> str:
    """A syntactically plausible shimp URI of exactly n characters."""
    head = "shimp://v1/go/run/example.com/a@"
    tail = "0" * 40
    pad = n - len(head) - len(tail) - 1
    if pad < 0:
        return (head + tail)[:n]
    return head + tail + "/" + "a" * pad


def largest_ok(fn, lo: int, hi: int) -> int:
    """Largest n in [lo, hi] for which fn(n) succeeds. 0 if none.

    A binary search, which is why the boundary report below exists: this
    finds WHERE the limit is and says nothing about what happens when it is
    crossed, and the second question is the one that matters.
    """
    best = 0
    while lo <= hi:
        mid = (lo + hi) // 2
        try:
            ok = fn(mid)
        except Exception:
            ok = False
        if ok:
            best = mid
            lo = mid + 1
        else:
            hi = mid - 1
    return best


def run(boundary_only: bool) -> int:
    # DispatchEx, never Dispatch: see the module docstring.
    app = w32.DispatchEx("Excel.Application")
    # Visible, always. A hidden Excel holding a modal is a process that hangs
    # with no way for anyone to see why.
    app.Visible = True
    app.DisplayAlerts = False
    _log(f"our own Excel: {app.Version}")

    wb = app.Workbooks.Add()
    try:
        ws = wb.Worksheets(1)

        def try_formula(n: int) -> bool:
            u = url_of(n)
            ws.Range("A1").Formula = f'=HYPERLINK("{u}","click")'
            got = ws.Range("A1").Formula
            return isinstance(got, str) and u in got

        def try_object(n: int) -> bool:
            u = url_of(n)
            cell = ws.Range("B1")
            cell.ClearContents()
            try:
                ws.Hyperlinks.Delete()
            except Exception:
                pass
            ws.Hyperlinks.Add(Anchor=cell, Address=u, TextToDisplay="click")
            return cell.Hyperlinks(1).Address == u

        def try_cell_text(n: int) -> bool:
            u = url_of(n)
            ws.Range("C1").Value = u
            return ws.Range("C1").Value == u

        if not boundary_only:
            _log("\nLONGEST URI THAT ROUND-TRIPS")
            for name, fn, hi in (("=HYPERLINK() formula", try_formula, 8192),
                                 ("Hyperlinks.Add .Address", try_object, 8192),
                                 ("plain cell text", try_cell_text, 40000)):
                _log(f"  {name:<26} {largest_ok(fn, 1, hi)}")

        # ⚠️ THE BOUNDARY REPORT, which is the part worth keeping.
        _log("\nWHAT HAPPENS AT THE BOUNDARY (Hyperlinks.Add)")
        _log(f"  {'asked':>7}  {'stored':>7}  outcome")
        for n in (255, 256, 1000, 2083, 2084, 2085, 2100, 4000, 8000):
            u = url_of(n)
            cell = ws.Range("A1")
            try:
                if cell.Hyperlinks.Count:
                    cell.Hyperlinks.Delete()
                cell.ClearContents()
                ws.Hyperlinks.Add(Anchor=cell, Address=u, TextToDisplay="click")
                got = cell.Hyperlinks(1).Address
                if got == u:
                    _log(f"  {n:>7}  {len(got):>7}  exact round-trip")
                else:
                    _log(f"  {n:>7}  {len(got):>7}  ⚠️ SILENTLY CHANGED "
                         f"(by {n - len(got)})")
            except Exception as exc:
                _log(f"  {n:>7}  {'-':>7}  REFUSED: "
                     f"{str(exc).split(chr(10))[0][:80]}")

        # ⚠️ The 15 Oct complication, reproduced here so the two halves can be
        # compared in one run rather than remembered separately.
        _log("\nWHAT =HYPERLINK() STORES, past the believed 255 limit")
        # ⚠️ EACH CASE CAUGHT SEPARATELY. The first version let the
        # exception escape, so the run died at 8000 and threw away the six
        # results it had already taken. An instrument must survive the
        # interesting case.
        for n in (255, 256, 326, 1000, 2084, 2085, 4000, 6000, 8000, 32000):
            u = url_of(n)
            try:
                ws.Range("D1").Formula = f'=HYPERLINK("{u}","click")'
                stored = ws.Range("D1").Formula
                inner = stored.split('"')[1] if '"' in stored else ""
                verdict = ("unaltered" if inner == u
                           else f"⚠️ CHANGED to {len(inner)}")
                _log(f"  asked {n:>6}  stored {len(inner):>6}  {verdict}")
            except Exception as exc:
                _log(f"  asked {n:>6}  {'-':>6}  REFUSED: "
                     f"{str(exc).split(chr(10))[0][:70]}")
        _log(f"  Hyperlinks.Count after a formula: {ws.Hyperlinks.Count}")
        _log("  ⚠️ zero, so a formula hyperlink exposes NO object to follow, "
             "and its delivery half needs a human click")
        return 0
    except Exception:
        import traceback
        _log(f"FATAL:\n{traceback.format_exc()}")
        return 1
    finally:
        # Closed without saving, and OUR instance quit. Because this is a
        # DispatchEx process of our own, quitting cannot take anyone else's
        # workbooks with it.
        try:
            wb.Close(SaveChanges=False)
        except Exception as exc:
            _log(f"WARNING: could not close the scratch workbook: {exc}")
        try:
            app.Quit()
        except Exception as exc:
            _log(f"WARNING: could not quit our Excel: {exc}")
        _log("\nscratch workbook closed without saving, our Excel quit")


def main() -> int:
    return run("--boundary-only" in sys.argv)


if __name__ == "__main__":
    safe_local_imports(globals())
    try:
        raise SystemExit(main())
    except SystemExit:
        raise
    except Exception:
        import traceback
        _log(f"FATAL unhandled exception:\n{traceback.format_exc()}")
        raise
