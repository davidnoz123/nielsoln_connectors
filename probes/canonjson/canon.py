"""The Python half of the canonical-JSON agreement probe.

    & "<venv>\\Scripts\\python.exe" canon.py records.jsonl

Emits one line per record: <sha256>\t<canonical bytes>

This is the REFERENCE, not a candidate. patchbucket computes patch_hash as
sha256 over json.dumps(patch, ensure_ascii=False, sort_keys=True), so whatever
this prints IS the thing a Go peer has to reproduce byte for byte. If they
disagree, every Go-written patch is refused as invalid_patch_hash.
"""
import hashlib
import json
import sys
from pathlib import Path


def canonical(obj) -> bytes:
    # EXACTLY patchbucket's call. Not "equivalent to": the same arguments in
    # the same order, so this probe cannot drift from the thing it models.
    return json.dumps(obj, ensure_ascii=False, sort_keys=True).encode("utf-8")


def main() -> int:
    src = Path(sys.argv[1] if len(sys.argv) > 1 else "records.jsonl")
    out = []
    for raw in src.read_text(encoding="utf-8").splitlines():
        if not raw.strip():
            continue
        # ⚠️ parse_float / parse_int left at the default on purpose. The
        # question is what Python's OWN round trip produces, since that is
        # what patchbucket hashes.
        obj = json.loads(raw)
        b = canonical(obj)
        out.append(f"{hashlib.sha256(b).hexdigest()}\t{b.decode('utf-8')}")
    sys.stdout.buffer.write(("\n".join(out) + "\n").encode("utf-8"))
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except SystemExit:
        raise
    except Exception:
        import traceback
        print(f"FATAL unhandled exception:\n{traceback.format_exc()}")
        raise
