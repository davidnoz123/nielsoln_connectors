"""Can Go ORIGINATE a float the way Python renders it? The Python half.

    & "<venv>\\Scripts\\python.exe" floats.py > floats.out

Reads float64 values as HEX BIT PATTERNS, one per line, so both languages
start from bit-identical doubles and the probe measures RENDERING only. A
decimal literal would leave a parsing question tangled up with the formatting
one, and parsing is not what is in doubt.

Emits: <hex>\t<json.dumps of that float>
"""
import json
import struct
import sys
from pathlib import Path

# ⚠️ CHOSEN TO BREAK THINGS, not to pass. Whole numbers (where Python appends
# .0 and Go does not), the exponent-form thresholds at both ends, values whose
# shortest round-trip is long, and the subnormal floor.
VALUES = [
    0.0, -0.0, 1.0, -1.0, 0.1, 0.3333333333333333,
    1700000000.0, 1700000000.5, 1792373986.6971858, 1730000000.123456,
    1e-5, 1e-4, 1e-7, 2.5e-10, 1e16, 1e17, 1e20, 1e21, 1e22,
    1.5e300, 5e-324, 2.2250738585072014e-308,
    9007199254740992.0, 123456789012345678.0,
    float("1e308"), 0.000001, 0.0000001,
]


def main() -> int:
    if len(sys.argv) > 1 and sys.argv[1] == "--emit-bits":
        out = [f"{struct.unpack('<Q', struct.pack('<d', v))[0]:016x}"
               for v in VALUES]
        Path("floats.bits").write_text("\n".join(out) + "\n", encoding="utf-8")
        print(f"wrote {len(out)} bit patterns to floats.bits")
        return 0

    bits = Path("floats.bits").read_text(encoding="utf-8").split()
    lines = []
    for h in bits:
        v = struct.unpack("<d", struct.pack("<Q", int(h, 16)))[0]
        # json.dumps of the BARE float, which is what patchbucket's
        # json.dumps(patch, ...) does to every number inside the envelope.
        lines.append(f"{h}\t{json.dumps(v)}")
    sys.stdout.buffer.write(("\n".join(lines) + "\n").encode("utf-8"))
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
