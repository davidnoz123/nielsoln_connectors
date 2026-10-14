"""excel_route.py -- trigger a shimp:// link from Excel, two ways.

    & "<venv>\\Scripts\\python.exe" excel_route.py object  "<uri>"
    & "<venv>\\Scripts\\python.exe" excel_route.py formula "<uri>"

Driven by linkprobe, which supplies the URI and reads the handler's log to
find out what actually arrived. This script's only job is to make Excel
follow a link; it deliberately draws no conclusion, because the conclusion
lives where the received bytes are.

WHY TWO MODES, AND WHY THAT IS THE POINT
----------------------------------------
Excel has two things called a hyperlink and they behave differently:

  object   Hyperlinks.Add(anchor, Address=uri) then .Follow()
           Caps at 2084 bytes and REFUSES at 2085, which is honest.

  formula  =HYPERLINK("uri") then .Follow() on the resulting link
           Truncates at 255 bytes SILENTLY.

The second is the one everybody reaches for and it is a trap. Worse than the
truncation: a formula cell reads back as its RESULT, so a pull/publish cycle
over a sheet destroys the link without touching the row.

Both were measured before this script existed. What was never measured is
whether either route delivers the URI's bytes UNCHANGED, which is what s44
requires and what linkprobe is here to find out.
"""
import sys


def safe_local_imports(g: dict) -> None:
    """Load the non-stdlib imports, logging a FATAL line before dying.

    Centralised so that an ImportError in COM support produces a message
    rather than a silent exit, which matters more than usual here: this
    script is launched as a subprocess and its stderr goes into a pipe
    somebody may not be reading.
    """
    try:
        import win32com.client as w32
        g["w32"] = w32
    except Exception:
        import traceback as _tb
        print(f"FATAL: safe_local_imports failed:\n{_tb.format_exc()}")
        raise


def _log(msg: str) -> None:
    print(msg, flush=True)


def follow(mode: str, uri: str) -> int:
    """Put a link in a scratch workbook and follow it. Returns an exit code."""
    # DispatchEx, NOT Dispatch, and this is not a style preference.
    #
    # Dispatch ATTACHES to a running Excel if one exists. Measured on this
    # machine: eighteen instances were already open. Attaching would have
    # added a scratch workbook to somebody's live session and, worse, set
    # DisplayAlerts = False on the application they were working in, which
    # outlives this script. DispatchEx starts its own process.
    app = w32.DispatchEx("Excel.Application")

    # Visible, always. A hidden Excel holding a modal dialog is a process
    # that hangs with no way for anyone to see why, which is the single
    # failure mode this workspace is most organised against.
    app.Visible = True

    # Scoped to OUR instance only, which is the whole reason DispatchEx
    # matters above.
    app.DisplayAlerts = False

    wb = app.Workbooks.Add()
    try:
        ws = wb.Worksheets(1)
        cell = ws.Range("A1")

        if mode == "object":
            # The Hyperlink OBJECT route. Refuses over 2084 bytes, so a
            # failure here is Excel being honest and is reported as such.
            ws.Hyperlinks.Add(Anchor=cell, Address=uri, TextToDisplay="probe")
            _log(f"added a Hyperlink object, Address is {len(uri)} bytes")
            stored = ws.Hyperlinks(1).Address
            if stored != uri:
                _log(f"NOTE: Excel STORED {len(stored)} bytes, not {len(uri)}. "
                     f"The link was altered before it was ever followed.")
            ws.Hyperlinks(1).Follow()

        elif mode == "formula":
            # The FORMULA route, and the measurement here is what Excel
            # STORED, not whether a link can be followed.
            #
            # ⚠️ THE FIRST VERSION OF THIS BRANCH MEASURED NOTHING. It wrote
            # the formula and then called FollowHyperlink(Address=uri) with
            # the original string, so it followed the VARIABLE rather than
            # the cell's stored target. It reported VERBATIM, which was true
            # of a link this route had never touched.
            cell.Formula = f'=HYPERLINK("{uri}","probe")'
            _log(f"wrote =HYPERLINK() over a {len(uri)}-byte URI")

            # What Excel kept. This is the real question for this route.
            stored = cell.Formula
            inner = ""
            if '"' in stored:
                inner = stored.split('"')[1]
            _log(f"Excel stored a {len(inner)}-byte address inside the formula")
            if inner != uri:
                _log(f"TRUNCATED-OR-ALTERED: sent {len(uri)}, stored "
                     f"{len(inner)} bytes")
                _log(f"stored: {inner}")
            else:
                _log("the formula still holds the address unaltered")

            # ⚠️ AND THERE IS NOTHING TO FOLLOW. A formula HYPERLINK creates
            # no Hyperlink OBJECT, so Hyperlinks.Count is 0 and there is no
            # .Follow() to call. Excel resolves the target only when a human
            # clicks the cell.
            _log(f"Hyperlinks.Count on the sheet is {ws.Hyperlinks.Count}")
            _log("CANNOT-COMPLETE: a formula hyperlink exposes no object to "
                 "follow, so the delivery half of this route is only "
                 "answerable by a human clicking the cell. Reporting that "
                 "rather than following the URI string directly, which is "
                 "what the previous version did and it measured nothing.")
            return 3

        else:
            _log(f"unknown mode {mode!r}: expected 'object' or 'formula'")
            return 2

        return 0

    except Exception as exc:
        import traceback
        _log(f"FATAL in mode {mode!r}:\n{traceback.format_exc()}")
        _log(f"summary: {exc}")
        return 1
    finally:
        # Closed without saving, and OUR instance quit. Because this is a
        # DispatchEx process of our own, quitting it cannot take anyone
        # else's workbooks with it.
        try:
            wb.Close(SaveChanges=False)
        except Exception as exc:
            _log(f"WARNING: could not close the scratch workbook: {exc}")
        try:
            app.Quit()
        except Exception as exc:
            _log(f"WARNING: could not quit our Excel instance: {exc}")


def main() -> int:
    if len(sys.argv) != 3:
        _log(__doc__)
        return 2
    return follow(sys.argv[1], sys.argv[2])


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
