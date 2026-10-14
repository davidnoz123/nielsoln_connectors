"""chrome_route.py -- trigger a shimp:// link from a real web page, in Chrome.

    & "<venv>\\Scripts\\python.exe" chrome_route.py click      "<uri>"
    & "<venv>\\Scripts\\python.exe" chrome_route.py no-gesture "<uri>"
    & "<venv>\\Scripts\\python.exe" chrome_route.py control    "<uri>"

Driven by linkprobe, which reads the handler's log to find out what arrived.

WHY A SERVER AND NOT A file:// PAGE
-----------------------------------
This started as a file:// page and that measured the wrong thing. Chrome keys
its external-protocol permission on ORIGIN, and a file:// page has an opaque
origin that cannot hold one. So a file:// test can never succeed, and
concluding from it that Chrome blocks synthetic clicks in general was wrong.

An http origin is also the realistic route: a link on a web page is what
shimp links are for.

WHAT IS MEASURED: THE PER-ORIGIN PERMISSION IS THE SOLE GATE
------------------------------------------------------------
    origin  permission   gesture       handler
    file    cannot hold  click         silent
    http    none         click         navigation requested, launch REFUSED
    http    granted      click         FIRES, verbatim
    http    granted      meta refresh  FIRES, verbatim   <-- 3 passes of 3

The last two rows together settle it:

    click,      NO permission  ->  ABSENT  so permission is NECESSARY
    NO gesture, permission     ->  FIRES   so a gesture is NOT required

A CDP-synthesized click does count as a gesture, but that turns out not to
matter, because no gesture is needed at all.

!! THIS DISPROVES A CLAIM MADE TWICE IN THIS REPO. s62 and s63 both recorded
that "a web page cannot silently fire a shimp link" and treated it as a
security property in shimp's favour. It is false. A user grants the
per-origin permission once, through Chrome's "always allow" checkbox, and
from then on that origin can fire shimp links on any later visit with no
interaction whatsoever.

An earlier one-off run of the no-gesture test DID come back silent, which is
why it was believed. That run reused a Chrome instance already driven through
two other tests with its server restarted underneath it. The harness result
reproduces 3 of 3 on a fresh browser each time; the one-off does not
reproduce and is not trusted.

The consequence for shimp is a design one rather than something to fix here:
the handler must treat every arriving link as untrusted and must never act on
one merely because it arrived. The comfort of "somebody must have clicked" is
not available.

NO MACHINE STATE
----------------
The permission is seeded into a throwaway --user-data-dir, not into Chrome's
policy keys under HKCU\\SOFTWARE\\Policies. A probe should not rewrite the
user's browser policy to make its own measurement pass.
"""
import json
import os
import shutil
import subprocess
import sys
import tempfile
import threading
import time
import urllib.request
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

PORT = 8731
DEBUG_PORT = 9769
SCHEME = "shimp-linkprobe"
URI = ""


def _log(msg: str) -> None:
    print(msg, flush=True)


def safe_local_imports(g: dict) -> None:
    """Load the non-stdlib imports, logging a FATAL line before dying."""
    try:
        g["websocket"] = _install_and_import("websocket", "websocket-client")
    except Exception:
        import traceback as _tb
        _log(f"FATAL: safe_local_imports failed:\n{_tb.format_exc()}")
        raise


def _install_and_import(package: str, pip_name: str | None = None):
    """Import *package*, installing via pip if absent. Returns the module."""
    import importlib as _il
    try:
        return _il.import_module(package)
    except ImportError:
        _log(f"installing {pip_name or package} ...")
        subprocess.check_call(
            [sys.executable, "-m", "pip", "install", pip_name or package],
            stdout=subprocess.DEVNULL)
        return _il.import_module(package)


class H(BaseHTTPRequestHandler):
    def _page(self, title: str, body: str) -> bytes:
        return (f'<!doctype html><meta charset="utf-8"><title>{title}</title>'
                f"\n{body}\n").encode("utf-8")

    def do_GET(self) -> None:
        if self.path == "/":
            body = self._page("subject", f'<p><a href="{URI}">click</a></p>')
        elif self.path == "/control":
            body = self._page("control", '<p><a href="/target">click</a></p>')
        elif self.path == "/auto":
            # No gesture and no JavaScript.
            body = self._page(
                "auto",
                f'<meta http-equiv="refresh" content="0;url={URI}">'
                f"<p>auto</p>")
        elif self.path == "/target":
            body = self._page("ARRIVED", "<p>ARRIVED</p>")
        else:
            self.send_error(404)
            return
        self.send_response(200)
        self.send_header("Content-Type", "text/html; charset=utf-8")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def handle_one_request(self) -> None:
        # ⚠️ Chrome aborts requests when it navigates away, and the default
        # handler dumps a socketserver traceback per abort into our output.
        # That noise drowned the control route's actual report, which is the
        # one thing this driver exists to say. Reported as one line, never
        # swallowed silently.
        try:
            super().handle_one_request()
        except (ConnectionAbortedError, ConnectionResetError, BrokenPipeError):
            _log("  (client aborted a request, which Chrome does on navigate)")

    def log_message(self, fmt, *args) -> None:
        # Reported, so "nothing arrived" is never ambiguous between Chrome
        # never asking for the page and Chrome asking and then not launching.
        _log("  served " + (fmt % args))


WATCH = (
    "Page.frameRequestedNavigation",
    "Page.frameScheduledNavigation",
    "Network.requestWillBeSent",
)


class CDP:
    def __init__(self, ws_url: str):
        # suppress_origin, because Chrome rejects a CDP websocket carrying
        # an Origin header with 403 unless launched with
        # --remote-allow-origins. A CDP client has no business claiming a
        # web origin anyway.
        self.ws = websocket.create_connection(
            ws_url, timeout=20, suppress_origin=True)
        self.n = 0
        self.events: list[dict] = []

    def send(self, method: str, params: dict | None = None) -> dict:
        self.n += 1
        self.ws.send(json.dumps({"id": self.n, "method": method,
                                 "params": params or {}}))
        while True:
            msg = json.loads(self.ws.recv())
            if msg.get("id") == self.n:
                if "error" in msg:
                    raise RuntimeError(f"{method} failed: {msg['error']}")
                return msg.get("result", {})
            if "method" in msg:
                self.events.append(msg)

    def drain(self, seconds: float) -> None:
        self.ws.settimeout(0.4)
        end = time.time() + seconds
        while time.time() < end:
            try:
                msg = json.loads(self.ws.recv())
            except Exception:
                continue
            if "method" in msg:
                self.events.append(msg)
        self.ws.settimeout(20)

    def close(self) -> None:
        try:
            self.ws.close()
        except Exception as exc:
            _log(f"WARNING: closing the socket failed: {exc}")


def chrome_exe() -> str:
    """Find Chrome by searching known locations, reporting if absent.

    Searches rather than guesses, which is remote_chrome_cdp's
    chromeCandidates() pattern: report what cannot be found instead of
    failing somewhere less obvious later.
    """
    for base in ("PROGRAMFILES", "PROGRAMFILES(X86)", "LOCALAPPDATA"):
        root = os.environ.get(base)
        if not root:
            continue
        p = os.path.join(root, "Google", "Chrome", "Application", "chrome.exe")
        if os.path.exists(p):
            return p
    return ""


def page_ws() -> str:
    """The websocket of a real page target, not an extension's background."""
    raw = urllib.request.urlopen(
        f"http://127.0.0.1:{DEBUG_PORT}/json/list", timeout=5).read()
    for t in json.loads(raw):
        if (t.get("type") == "page"
                and not t.get("url", "").startswith("chrome-extension://")):
            return t["webSocketDebuggerUrl"]
    return ""


def click(c: CDP, page: str) -> list[str]:
    """Navigate to *page* and click its first link. Returns navigation events.

    Returning the events rather than a boolean is deliberate: an empty list
    means the click never landed, which is a harness fault and must not be
    reported as a finding about shimp. The first version of this could not
    tell the two apart and drew a wrong conclusion from a null result.
    """
    c.send("Page.enable")
    c.send("Network.enable")
    c.send("DOM.enable")
    c.send("Page.navigate", {"url": page})
    c.drain(2.5)

    # ⚠️ ALL THREE, because bringToFront alone was not enough: hit-testing
    # an inactive page silently finds nothing, which looks exactly like a
    # refused click. This was the harness fault that produced a wrong
    # conclusion, and it cost two rounds of measurement to find.
    for method, params in (("Page.bringToFront", None),
                           ("Page.setWebLifecycleState", {"state": "active"}),
                           ("Emulation.setFocusEmulationEnabled",
                            {"enabled": True})):
        try:
            c.send(method, params)
        except Exception as exc:
            _log(f"WARNING: {method} failed: {exc}")

    root = c.send("DOM.getDocument", {"depth": 1})["root"]["nodeId"]
    node = c.send("DOM.querySelector",
                  {"nodeId": root, "selector": "a"})["nodeId"]
    if not node:
        _log("FATAL: no <a> on the page")
        return []
    q = c.send("DOM.getBoxModel", {"nodeId": node})["model"]["content"]
    x, y = (q[0] + q[2]) / 2, (q[1] + q[5]) / 2
    _log(f"  link centre {x:.0f},{y:.0f}")

    # DOM.getBoxModel plus Input.dispatchMouseEvent, with no
    # Runtime.evaluate anywhere: cdp.go's vocabulary is closed and evaluate
    # is not in it, and a real mouse event at real coordinates is a more
    # faithful model of a click in any case.
    c.events.clear()
    for kind in ("mouseMoved", "mousePressed", "mouseReleased"):
        p = {"type": kind, "x": x, "y": y, "button": "left"}
        if kind != "mouseMoved":
            p["clickCount"] = 1
        c.send("Input.dispatchMouseEvent", p)
    c.drain(4.0)
    return [e["method"] for e in c.events if e.get("method") in WATCH]


def run(mode: str, uri: str) -> int:
    global URI
    URI = uri

    exe = chrome_exe()
    if not exe:
        _log("CANNOT-COMPLETE: no chrome.exe in any known location")
        return 3

    srv = ThreadingHTTPServer(("127.0.0.1", PORT), H)
    # Logged after the socket is bound, never before: a server that
    # announces a port it has not got reports a measurement it cannot take.
    _log(f"  serving http://127.0.0.1:{PORT}/")
    threading.Thread(target=srv.serve_forever, daemon=True).start()

    prof = tempfile.mkdtemp(prefix="linkprobe-chrome-")
    os.makedirs(os.path.join(prof, "Default"), exist_ok=True)
    if mode != "control":
        # Profile-local permission for THIS origin only. No machine state.
        pref = {"protocol_handler": {"allowed_origin_protocol_pairs": {
            f"http://127.0.0.1:{PORT}": {SCHEME: True}}}}
        with open(os.path.join(prof, "Default", "Preferences"), "w",
                  encoding="utf-8") as fh:
            json.dump(pref, fh)
        _log(f"  permitted {SCHEME} for http://127.0.0.1:{PORT} in a "
             f"throwaway profile")

    proc = subprocess.Popen(
        [exe, f"--user-data-dir={prof}", "--no-first-run",
         "--no-default-browser-check",
         f"--remote-debugging-port={DEBUG_PORT}", "about:blank"])
    c = None
    try:
        ws = ""
        for _ in range(40):
            try:
                ws = page_ws()
                if ws:
                    break
            except Exception:
                pass
            time.sleep(0.5)
        if not ws:
            _log("CANNOT-COMPLETE: Chrome never offered a page target")
            return 3

        c = CDP(ws)
        if mode == "no-gesture":
            # No click at all. A meta refresh to the shimp URI.
            c.send("Page.enable")
            c.send("Page.navigate", {"url": f"http://127.0.0.1:{PORT}/auto"})
            c.drain(6.0)
            _log("  navigated to /auto with NO gesture of any kind")
            return 0

        page = (f"http://127.0.0.1:{PORT}/control" if mode == "control"
                else f"http://127.0.0.1:{PORT}/")
        events = click(c, page)
        if not events:
            _log("CANNOT-COMPLETE: the click did not land, so this run "
                 "measures the harness rather than Chrome")
            return 3
        _log(f"  click landed, Chrome reported: {', '.join(sorted(set(events)))}")
        return 0
    except Exception:
        import traceback
        _log(f"FATAL in mode {mode!r}:\n{traceback.format_exc()}")
        return 1
    finally:
        if c is not None:
            c.close()
        try:
            proc.terminate()
            proc.wait(timeout=15)
        except Exception as exc:
            _log(f"WARNING: could not stop our Chrome: {exc}")
        srv.shutdown()
        shutil.rmtree(prof, ignore_errors=True)


def main() -> int:
    if len(sys.argv) != 3:
        _log(__doc__)
        return 2
    mode = sys.argv[1]
    if mode not in ("click", "no-gesture", "control"):
        _log(f"unknown mode {mode!r}")
        return 2
    return run(mode, sys.argv[2])


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
