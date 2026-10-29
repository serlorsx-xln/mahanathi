"""Compare the Latin outlines of two fonts glyph by glyph.

Usage (from the project root):

    .cache/venv/bin/python -I scripts/research/compare_outlines.py \
        FONT_A[@wght=400] FONT_B[@wght=400] [--chars "..."]

For each character, both glyphs are drawn through HarfBuzz at the given
variation location, scaled to 1000 units per em, and compared on:
  - advance width
  - outline structure (sequence of pen operations)
  - largest point deviation when the structure matches
  - bounding box when it does not
It prints per-glyph rows and a tally. A tally of mostly identical or
near-identical outlines means one Latin is derived from the other.
"""

import argparse
import sys

import uharfbuzz as hb
from fontTools.ttLib import TTFont

DEFAULT_CHARS = (
    "abcdefghijklmnopqrstuvwxyz"
    "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
    "0123456789"
)


class Rec:
    """Records HarfBuzz draw callbacks as (op, points)."""

    def __init__(self):
        self.ops = []

    def moveTo(self, p):
        self.ops.append(("M", [p]))

    def lineTo(self, p):
        self.ops.append(("L", [p]))

    def qCurveTo(self, *pts):
        self.ops.append(("Q", list(pts)))

    def curveTo(self, *pts):
        self.ops.append(("C", list(pts)))

    def closePath(self):
        self.ops.append(("Z", []))


def load(spec):
    path, var = spec, None
    if "@" in spec:
        path, v = spec.rsplit("@", 1)
        var = {k: float(x) for k, x in (kv.split("=") for kv in v.split(","))}
    tt = TTFont(path, lazy=True)
    upm = tt["head"].unitsPerEm
    font = hb.Font(hb.Face(hb.Blob.from_file_path(path)))
    if var:
        font.set_variations(var)
    name = str(tt["name"].getDebugName(16) or tt["name"].getDebugName(1))
    return font, upm, name, tt.getBestCmap()


def glyph(font, upm, cmap, ch):
    if ord(ch) not in cmap:
        return None
    buf = hb.Buffer()
    buf.add_str(ch)
    buf.guess_segment_properties()
    hb.shape(font, buf, {})
    gid = buf.glyph_infos[0].codepoint
    adv = buf.glyph_positions[0].x_advance * 1000 / upm
    rec = Rec()
    font.draw_glyph_with_pen(gid, rec)
    s = 1000 / upm
    ops = [(op, [(x * s, y * s) for x, y in pts]) for op, pts in rec.ops]
    return adv, ops


def bbox(ops):
    pts = [p for _, ps in ops for p in ps]
    if not pts:
        return (0, 0, 0, 0)
    xs, ys = [p[0] for p in pts], [p[1] for p in pts]
    return (min(xs), min(ys), max(xs), max(ys))


def main(argv):
    ap = argparse.ArgumentParser()
    ap.add_argument("a")
    ap.add_argument("b")
    ap.add_argument("--chars", default=DEFAULT_CHARS)
    ap.add_argument("--quiet", action="store_true")
    a = ap.parse_args(argv)
    fa, ua, na, ca = load(a.a)
    fb, ub, nb, cb = load(a.b)
    tally = {"identical": 0, "near": 0, "same_structure": 0, "different": 0, "missing": 0}
    print(f"A = {na}  B = {nb}")
    for ch in a.chars:
        ga, gb = glyph(fa, ua, ca, ch), glyph(fb, ub, cb, ch)
        if ga is None or gb is None:
            tally["missing"] += 1
            continue
        (adva, opsa), (advb, opsb) = ga, gb
        sig_a = [(op, len(p)) for op, p in opsa]
        sig_b = [(op, len(p)) for op, p in opsb]
        if sig_a == sig_b:
            dev = max(
                (abs(pa[0] - pb[0]) + abs(pa[1] - pb[1])
                 for (_, psa), (_, psb) in zip(opsa, opsb)
                 for pa, pb in zip(psa, psb)),
                default=0.0)
            if dev <= 1 and abs(adva - advb) <= 1:
                cls = "identical"
            elif dev <= 15:
                cls = "near"
            else:
                cls = "same_structure"
            detail = f"max point dev {dev:.1f}"
        else:
            cls = "different"
            ba, bb = bbox(opsa), bbox(opsb)
            detail = "bbox A " + str(tuple(round(v) for v in ba)) + " B " + str(tuple(round(v) for v in bb))
        tally[cls] += 1
        if not a.quiet:
            print(f"  {ch}  adv {adva:.0f}/{advb:.0f}  {cls:15s} {detail}")
    print("tally:", tally)


if __name__ == "__main__":
    main(sys.argv[1:])
