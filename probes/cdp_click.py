"""Does a CDP-synthesized click land, and what did Chrome do with it?

    & "<venv>\\Scripts\\python.exe" probes\\cdp_click.py <page-url>
    & "<venv>\\Scripts\\python.exe" probes\\cdp_click.py --self-test

WHAT IT MEASURED, 15 October 2026
---------------------------------
    CONTROL: a synthetic click on an ordinary <a href>   NAVIGATES
    a click on a shimp:// href, origin NOT permitted     launch REFUSED
    a click on a shimp:// href, origin permitted         FIRES, verbatim
    no click at all, meta refresh, origin permitted      FIRES, verbatim

!! THIS INSTRUMENT OVERTURNED A WRITTEN CONCLUSION. s63 had recorded that
Chrome blocks an external-protocol launch from CDP-synthesized input, and
that there was therefore no automated browser route. Both halves were wrong,
and this is what showed it:

  * the earlier test used a file:// page, and file:// is an OPAQUE ORIGIN
    that cannot hold Chrome's per-origin protocol permission. The conclusion
    had been generalised from the one origin incapable of passing.
  * the control was broken. Until Page.setWebLifecycleState and
    Emulation.setFocusEmulationEnabled were added, an ordinary link did not
    navigate either, so the null result said nothing at all.

!! AND IT DISPROVED A SECURITY PROPERTY CLAIMED TWICE. s62 and s63 both said
a web page cannot silently fire a shimp link. The last row above shows it
can: a meta refresh, no JavaScript and no click, once the origin is
permitted. The per-origin permission is the SOLE gate; the gesture is not
required. A user grants it once through Chrome's "always allow" checkbox.

WHY IT REPORTS EVENTS RATHER THAN SUCCESS OR FAILURE
----------------------------------------------------
Three outcomes look identical from outside and mean completely different
things:

    no navigation events at all    the click never landed -- HARNESS FAULT
    a navigation request           the click landed and the link activated
    a request and nothing after    it activated and something refused it

Conflating the first two is what produced the wrong conclusion. So this
prints what Chrome reported and leaves the reading to a human.
"""
import sys

from pathlib import Path

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

sys.path.insert(0, str(Path(__file__).resolve().parent))


def _log(msg: str) -> None:
    print(msg, flush=True)


def safe_local_imports(g: dict) -> None:
    """Load the non-stdlib local imports, logging FATAL before dying."""
    try:
        from cdp import CDP, page_target
        g["CDP"], g["page_target"] = CDP, page_target
    except Exception:
        import traceback as _tb
        _log(f"FATAL: safe_local_imports failed:\n{_tb.format_exc()}")
        raise


DEBUG_PORT = 9772

WATCH = (
    "Page.frameRequestedNavigation",
    "Page.frameScheduledNavigation",
    "Page.frameStartedLoading",
    "Page.windowOpen",
    "Network.requestWillBeSent",
)


def diagnose(page_url: str, ws_url: str) -> int:
    c = CDP(ws_url)
    try:
        c.send("Page.enable")
        c.send("Network.enable")
        c.send("DOM.enable")
        c.send("Page.navigate", {"url": page_url})
        c.drain(2.5)
        c.activate()

        # The viewport, reported rather than assumed: if layout metrics
        # disagree with the box model the click lands somewhere else, and
        # that is indistinguishable from a refusal.
        m = c.send("Page.getLayoutMetrics")
        vp = m.get("cssLayoutViewport") or m.get("layoutViewport") or {}
        _log(f"  viewport {vp.get('clientWidth')}x{vp.get('clientHeight')}")

        c.events.clear()
        x, y = c.click_first_link()
        _log(f"  clicked at {x:.0f},{y:.0f}")
        c.drain(4.0)

        seen = [e for e in c.events if e.get("method") in WATCH]
        if not seen:
            _log("  NO NAVIGATION EVENTS AT ALL.")
            _log("  ⚠️ The click did not land. This is a HARNESS FAULT and "
                 "not a finding about anything else.")
            return 1
        for e in seen:
            p = e.get("params", {})
            what = p.get("url") or p.get("request", {}).get("url", "")
            _log(f"    {e['method']}: {what[:110]}")
        _log("  the click LANDED and the link activated. Whether the target "
             "then ran is a separate question, answered by the handler's log.")
        return 0
    except Exception:
        import traceback
        _log(f"FATAL:\n{traceback.format_exc()}")
        return 1
    finally:
        c.close()


def self_test() -> int:
    """⚠️ THE CONTROL, as a mode rather than a thing done once by hand.

    Serves two pages and clicks the first. If the ordinary link does not
    navigate, every other result this instrument gives is worthless, and
    that is the state it was in when it produced a wrong conclusion.
    """
    import http.server
    import threading

    class H(http.server.BaseHTTPRequestHandler):
        def do_GET(self):
            if self.path == "/target":
                body = b"<!doctype html><title>ARRIVED</title><p>ARRIVED</p>"
            else:
                body = (b'<!doctype html><title>control</title>'
                        b'<p><a href="/target">click</a></p>')
            self.send_response(200)
            self.send_header("Content-Type", "text/html; charset=utf-8")
            self.send_header("Content-Length", str(len(body)))
            self.end_headers()
            self.wfile.write(body)

        def handle_one_request(self):
            try:
                super().handle_one_request()
            except (ConnectionAbortedError, ConnectionResetError,
                    BrokenPipeError):
                pass

        def log_message(self, fmt, *args):
            _log("    served " + (fmt % args))

    srv = http.server.ThreadingHTTPServer(("127.0.0.1", 0), H)
    port = srv.server_address[1]
    threading.Thread(target=srv.serve_forever, daemon=True).start()
    _log(f"  control server on http://127.0.0.1:{port}/")
    _log("  ⚠️ an http origin, NOT file://: a file page has an opaque origin "
         "and the measurement this instrument exists for cannot be taken on "
         "one")
    try:
        ws = page_target(DEBUG_PORT)
        if not ws:
            _log(f"  CANNOT RUN: no Chrome on --remote-debugging-port="
                 f"{DEBUG_PORT}. Start one with a throwaway profile first.")
            return 3
        return diagnose(f"http://127.0.0.1:{port}/control", ws)
    finally:
        srv.shutdown()


def main() -> int:
    if "--self-test" in sys.argv:
        return self_test()
    if len(sys.argv) != 2:
        _log(__doc__)
        return 2
    ws = page_target(DEBUG_PORT)
    if not ws:
        _log(f"CANNOT RUN: no Chrome on --remote-debugging-port={DEBUG_PORT}")
        return 3
    return diagnose(sys.argv[1], ws)


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
