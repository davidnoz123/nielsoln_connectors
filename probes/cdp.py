"""cdp.py -- the smallest CDP client the probes need.

Not a general client. It connects, sends a method, reads the matching reply
and collects events, which is all any instrument here does.

!! TWO THINGS IN IT WERE LEARNED THE HARD WAY, and both are one-line fixes
whose absence cost a measurement each.

1. suppress_origin. Chrome 155 rejects a CDP websocket carrying an Origin
   header with 403 unless launched with --remote-allow-origins. The first run
   of the click probe reported "PREMISE FAILS" on that 403, which was the
   harness mislabelling its own transport error as a result about shimp. Only
   the traceback revealed it. A CDP client has no business claiming a web
   origin anyway.

2. Replies are matched BY ID. Events interleave with replies on the same
   socket, so reading the next frame and assuming it is the answer works
   until the page emits anything.
"""
import json
import subprocess
import sys

import time
import urllib.request

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


def install_and_import(package: str, pip_name: str | None = None):
    """Import *package*, installing via pip if absent. Returns the module.

    Never call pip from a terminal: an install done by hand is not auditable
    in source and is invisible when this runs unattended.
    """
    import importlib
    try:
        return importlib.import_module(package)
    except ImportError:
        _log(f"installing {pip_name or package} ...")
        subprocess.check_call(
            [sys.executable, "-m", "pip", "install", pip_name or package],
            stdout=subprocess.DEVNULL)
        return importlib.import_module(package)


def page_target(debug_port: int) -> str:
    """The websocket of a real page, not an extension's background page.

    ⚠️ Picking the FIRST entry in /json/list is wrong and looks right: on a
    fresh Chrome the first target is often a chrome-extension background
    page, and a probe attached to it reports "no <a> on the page" for a page
    it never loaded.
    """
    raw = urllib.request.urlopen(
        f"http://127.0.0.1:{debug_port}/json/list", timeout=5).read()
    for t in json.loads(raw):
        if (t.get("type") == "page"
                and not t.get("url", "").startswith("chrome-extension://")):
            return t["webSocketDebuggerUrl"]
    return ""


class CDP:
    """One websocket to one target."""

    def __init__(self, ws_url: str, timeout: int = 20):
        websocket = install_and_import("websocket", "websocket-client")
        self.ws = websocket.create_connection(
            ws_url, timeout=timeout, suppress_origin=True)
        self.n = 0
        self.events: list[dict] = []

    def send(self, method: str, params: dict | None = None) -> dict:
        self.n += 1
        self.ws.send(json.dumps(
            {"id": self.n, "method": method, "params": params or {}}))
        while True:
            msg = json.loads(self.ws.recv())
            if msg.get("id") == self.n:
                if "error" in msg:
                    raise RuntimeError(f"{method} failed: {msg['error']}")
                return msg.get("result", {})
            if "method" in msg:
                self.events.append(msg)

    def drain(self, seconds: float) -> None:
        """Collect events for a while without sending anything."""
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

    def activate(self) -> None:
        """Make the page hit-testable.

        ⚠️ ALL THREE CALLS, and the omission of the last two cost two rounds
        of measurement. Hit-testing an inactive page silently finds nothing,
        which is indistinguishable from a refused click: an ordinary link did
        not navigate either, so a null result said nothing at all.
        bringToFront alone was not enough.
        """
        for method, params in (("Page.bringToFront", None),
                               ("Page.setWebLifecycleState", {"state": "active"}),
                               ("Emulation.setFocusEmulationEnabled",
                                {"enabled": True})):
            try:
                self.send(method, params)
            except Exception as exc:
                _log(f"  WARNING: {method} failed: {exc}")

    def click_first_link(self) -> tuple[float, float]:
        """Click the page's first <a>. Returns the coordinates used.

        DOM.getBoxModel plus Input.dispatchMouseEvent, with NO
        Runtime.evaluate anywhere: remote_chrome_cdp/cdp.go's vocabulary is
        closed and evaluate is not in it, and a real mouse event at real
        coordinates is a more faithful model of a click in any case.
        """
        root = self.send("DOM.getDocument", {"depth": 1})["root"]["nodeId"]
        node = self.send("DOM.querySelector",
                         {"nodeId": root, "selector": "a"})["nodeId"]
        if not node:
            raise RuntimeError("no <a> on the page")
        q = self.send("DOM.getBoxModel", {"nodeId": node})["model"]["content"]
        x, y = (q[0] + q[2]) / 2, (q[1] + q[5]) / 2
        for kind in ("mouseMoved", "mousePressed", "mouseReleased"):
            p = {"type": kind, "x": x, "y": y, "button": "left"}
            if kind != "mouseMoved":
                p["clickCount"] = 1
            self.send("Input.dispatchMouseEvent", p)
        return x, y

    def close(self) -> None:
        try:
            self.ws.close()
        except Exception as exc:
            _log(f"WARNING: closing the socket failed: {exc}")
