"""Render shaped Thai/Latin sample lines for several fonts into one SVG.

Usage (from the project root):

    .cache/venv/bin/python -I scripts/research/render_specimen.py \
        --out .cache/fonts/_render/sans.svg [--size 40] [--lh 1.6] \
        [--features ss01,tnum] [--text "line one|line two"] \
        FONT[@wght=400,wdth=100] [FONT ...]

Text is shaped with HarfBuzz (script Thai, language th) and drawn from the
font outlines, so what you see is the font's own GSUB/GPOS result, not a
browser fallback. Red lines mark each line's hhea ascender and descender,
grey dashed lines mark the line box at the --lh line-height. This is a
research aid for looking at letterforms and mark stacking. The real
specimen is built in the browser in Phase 3.
"""

import argparse
import sys

import uharfbuzz as hb
from fontTools.pens.svgPathPen import SVGPathPen
from fontTools.pens.transformPen import TransformPen
from fontTools.ttLib import TTFont

LINES = [
    "น้ำล้นตลิ่ง ผู้ใหญ่บ้าน ปิ่นเกล้า",
    "จังหวัดพระนครศรีอยุธยา ปั๊มน้ำ ฟื้นฟู ฤดูฝน",
    "บปนพฟมวหยลสรคถภดต กฎหมาย ปฏิบัติ ญาติ",
    "Mahanathi 3.42 m 14:00 น. 29 ต.ค. 69 1234567890",
]


class _Pen:
    """Adapter from HarfBuzz draw callbacks to a fontTools pen."""

    def __init__(self, pen):
        self.pen = pen

    def moveTo(self, p):
        self.pen.moveTo(p)

    def lineTo(self, p):
        self.pen.lineTo(p)

    def qCurveTo(self, *pts):
        self.pen.qCurveTo(*pts)

    def curveTo(self, *pts):
        self.pen.curveTo(*pts)

    def closePath(self):
        self.pen.closePath()


def parse(spec):
    if "@" in spec:
        path, var = spec.rsplit("@", 1)
        v = {}
        for kv in var.split(","):
            k, val = kv.split("=")
            v[k] = float(val)
        return path, v
    return spec, None


def main(argv):
    ap = argparse.ArgumentParser()
    ap.add_argument("fonts", nargs="+")
    ap.add_argument("--out", required=True)
    ap.add_argument("--size", type=float, default=40.0)
    ap.add_argument("--lh", type=float, default=1.6)
    ap.add_argument("--width", type=int, default=1800)
    ap.add_argument("--features", default="", help="comma-separated, e.g. ss01")
    ap.add_argument("--text", default="", help="lines separated by |")
    a = ap.parse_args(argv)
    feats = {f: True for f in a.features.split(",") if f}
    lines = a.text.split("|") if a.text else LINES

    size, lh = a.size, a.lh
    margin = 20
    parts = []
    y = margin
    for spec in a.fonts:
        path, var = parse(spec)
        tt = TTFont(path, lazy=True)
        upm = tt["head"].unitsPerEm
        asc, desc = tt["hhea"].ascent, tt["hhea"].descent
        fam = str(tt["name"].getDebugName(16) or tt["name"].getDebugName(1))
        sub = str(tt["name"].getDebugName(17) or tt["name"].getDebugName(2))
        face = hb.Face(hb.Blob.from_file_path(path))
        font = hb.Font(face)
        if var:
            font.set_variations(var)
        scale = size / upm
        label = f"{fam} {sub}" + (f" {var}" if var else "") + (f" features {a.features}" if feats else "")
        parts.append(
            f'<text x="{margin}" y="{y + 14}" font-family="Helvetica" '
            f'font-size="13" fill="#555">{label}  (lh {lh})</text>')
        y += 22
        line_px = size * lh
        for text in lines:
            # line box top at y; baseline from half-leading model
            content = (asc - desc) * scale
            half = (line_px - content) / 2
            base = y + half + asc * scale
            parts.append(f'<rect x="0" y="{y:.1f}" width="{a.width}" height="{line_px:.1f}" '
                         f'fill="none" stroke="#bbb" stroke-dasharray="4 4"/>')
            for yy in (base - asc * scale, base - desc * scale):
                parts.append(f'<line x1="0" x2="{a.width}" y1="{yy:.1f}" y2="{yy:.1f}" '
                             f'stroke="#e33" stroke-width="0.6"/>')
            buf = hb.Buffer()
            buf.add_str(text)
            buf.direction = "ltr"
            buf.script = "Thai"
            buf.language = "th"
            hb.shape(font, buf, feats)
            x = margin
            for info, pos in zip(buf.glyph_infos, buf.glyph_positions):
                svg = SVGPathPen(None)
                tx = x + pos.x_offset * scale
                ty = base - pos.y_offset * scale
                tp = TransformPen(svg, (scale, 0, 0, -scale, tx, ty))
                font.draw_glyph_with_pen(info.codepoint, _Pen(tp))
                d = svg.getCommands()
                if d:
                    parts.append(f'<path d="{d}" fill="#111"/>')
                x += pos.x_advance * scale
            y += line_px
        y += 16
    height = int(y + margin)
    svg = (f'<svg xmlns="http://www.w3.org/2000/svg" width="{a.width}" height="{height}" '
           f'viewBox="0 0 {a.width} {height}"><rect width="100%" height="100%" fill="#fff"/>'
           + "".join(parts) + "</svg>")
    with open(a.out, "w", encoding="utf-8") as fh:
        fh.write(svg)
    print(a.out, a.width, height)


if __name__ == "__main__":
    main(sys.argv[1:])
