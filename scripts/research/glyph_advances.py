"""Print shaped advance widths of figures and figure-adjacent characters.

Usage (from the project root):

    .cache/venv/bin/python -I scripts/research/glyph_advances.py \
        FONT[@wght=400,wdth=100] [FONT ...]

For each font, prints in units per 1000 em:
  - ASCII digits 0-9 (min-max, and whether all are equal)
  - Thai digits ๐-๙ (min-max)
  - the characters a measurement line uses next to figures: minus sign
    U+2212, hyphen-minus, plus, period, comma, colon, space, no-break
    space, figure space U+2007, degree, percent
Advances come from HarfBuzz shaping in a Latin run, so they include any
default substitutions. A character the font lacks is reported as "none".
"""

import sys

import uharfbuzz as hb
from fontTools.ttLib import TTFont

FIGURE_NEIGHBOURS = [
    ("minus U+2212", "−"),
    ("hyphen", "-"),
    ("plus", "+"),
    ("period", "."),
    ("comma", ","),
    ("colon", ":"),
    ("space", " "),
    ("nbsp", " "),
    ("figure space U+2007", " "),
    ("degree", "°"),
    ("percent", "%"),
]


def load(spec):
    path, var = spec, None
    if "@" in spec:
        path, v = spec.rsplit("@", 1)
        var = {k: float(x) for k, x in (kv.split("=") for kv in v.split(","))}
    tt = TTFont(path, lazy=True)
    font = hb.Font(hb.Face(hb.Blob.from_file_path(path)))
    if var:
        font.set_variations(var)
    upm = tt["head"].unitsPerEm
    name = str(tt["name"].getDebugName(16) or tt["name"].getDebugName(1))
    sub = str(tt["name"].getDebugName(17) or tt["name"].getDebugName(2))
    return font, upm, f"{name} {sub}" + (f" {var}" if var else ""), tt.getBestCmap()


def advances(font, upm, text, script="Latn"):
    buf = hb.Buffer()
    buf.add_str(text)
    buf.direction = "ltr"
    buf.script = script
    buf.language = "en" if script == "Latn" else "th"
    hb.shape(font, buf, {})
    return [round(p.x_advance * 1000 / upm) for p in buf.glyph_positions]


def main(argv):
    for spec in argv:
        font, upm, label, cmap = load(spec)
        print(f"== {label}")
        a = advances(font, upm, "0123456789")
        print(f"   ASCII digits {min(a)}-{max(a)} {'all equal' if len(set(a)) == 1 else 'proportional'}")
        thai_digits = "๐๑๒๓๔๕๖๗๘๙"
        if all(ord(c) in cmap for c in thai_digits):
            t = advances(font, upm, thai_digits, "Thai")
            print(f"   Thai digits  {min(t)}-{max(t)} {'all equal' if len(set(t)) == 1 else 'proportional'}")
        else:
            print("   Thai digits  none (not in cmap)")
        cells = []
        for label_, ch in FIGURE_NEIGHBOURS:
            if ord(ch) not in cmap:
                cells.append(f"{label_} none")
            else:
                cells.append(f"{label_} {advances(font, upm, ch)[0]}")
        print("   " + " | ".join(cells))


if __name__ == "__main__":
    main(sys.argv[1:])
