"""Constrained search for the five-step situation colour scale.

The scale must pass WCAG AA as small text on every surface (ground,
panel, elevated), stay quiet below level 4, and survive Machado
deuteranopia and protanopia simulation with every pair of levels
clearly distinct. Reds and greens both collapse toward olive under
those simulations, so the search maximises the minimum pairwise
delta-E (CIE76, on the simulated colours) and lets the separation be
carried by lightness, colorfulness and the blue axis (level 1).

Level 4 (the amber) is pinned: it is the scale's anchor and was chosen
by eye first. The output is pasted into web/src/design/tokens.ts; the
evidence screenshots live in docs/design/screens/phase-3/.

Usage (from the project root):
  .cache/venv/bin/python -I scripts/research/cvd_scale_search.py
"""
import colorsys
import math
import random

DEUTAN = [
    [0.367322, 0.860646, -0.227968],
    [0.280085, 0.672501, 0.047413],
    [-0.011820, 0.042940, 0.968881],
]
PROTAN = [
    [0.152286, 1.052583, -0.204868],
    [0.114503, 0.786281, 0.099216],
    [-0.003882, -0.048116, 1.051998],
]

SURFACES = ["#0e161d", "#15202b", "#1a2733"]
LEVEL4 = "#e3a93f"  # the amber, pinned by eye


def lin(c):
    return c / 12.92 if c <= 0.04045 else ((c + 0.055) / 1.055) ** 2.4


def unlin(c):
    return c * 12.92 if c <= 0.0031308 else 1.055 * c ** (1 / 2.4) - 0.055


def luminance(hex_):
    r, g, b = (lin(int(hex_[i:i + 2], 16) / 255) for i in (1, 3, 5))
    return 0.2126 * r + 0.7152 * g + 0.0722 * b


def contrast(a, b):
    la, lb = luminance(a), luminance(b)
    return (max(la, lb) + 0.05) / (min(la, lb) + 0.05)


def passes_aa(hex_):
    return all(contrast(hex_, s) >= 4.55 for s in SURFACES)


def hsv_saturation(hex_):
    rgb = [int(hex_[i:i + 2], 16) for i in (1, 3, 5)]
    mx, mn = max(rgb), min(rgb)
    return 0 if mx == 0 else (mx - mn) / mx


def to_lab(rgb):
    r, g, b = (lin(c / 255) for c in rgb)
    X = 0.4124 * r + 0.3576 * g + 0.1805 * b
    Y = 0.2126 * r + 0.7152 * g + 0.0722 * b
    Z = 0.0193 * r + 0.1192 * g + 0.9505 * b

    def f(t):
        return t ** (1 / 3) if t > 0.008856 else 7.787 * t + 16 / 116

    fx, fy, fz = f(X / 0.95047), f(Y / 1.0), f(Z / 1.08883)
    return (116 * fy - 16, 500 * (fx - fy), 200 * (fy - fz))


def simulate(hex_, matrix):
    """Machado simulation in linear light, back to sRGB hex."""
    rgb = [int(hex_[i:i + 2], 16) / 255 for i in (1, 3, 5)]
    out = [unlin(min(1, max(0, sum(lin(rgb[j]) * matrix[i][j] for j in range(3)))))
           for i in range(3)]
    return "#%02x%02x%02x" % tuple(round(c * 255) for c in out)


def candidates(hue_lo, hue_hi, hls_sat_lo, hls_sat_hi, sat_lo=None, sat_hi=None):
    """The first lightness per (hue, HLS saturation) that passes AA."""
    out = []
    for hue in range(hue_lo, hue_hi + 1, 2):
        for i in range(25):
            s = hls_sat_lo + (hls_sat_hi - hls_sat_lo) * i / 25
            for v in range(40, 92):
                r, g, b = colorsys.hls_to_rgb(hue / 360, v / 100, s)
                hex_ = "#%02x%02x%02x" % (round(r * 255), round(g * 255), round(b * 255))
                if passes_aa(hex_):
                    if sat_lo is not None and not (sat_lo <= hsv_saturation(hex_) <= sat_hi):
                        break
                    out.append(hex_)
                    break
    return out


def main():
    # level 1: muted slate blue; 2: near-grey; 3: quiet green;
    # 5: strong red-family. Level 4 is pinned above.
    pools = [
        [h for h in candidates(198, 218, 0.10, 0.40) if hsv_saturation(h) <= 0.35],
        [h for h in candidates(60, 140, 0.02, 0.20) if hsv_saturation(h) <= 0.16],
        [h for h in candidates(110, 175, 0.18, 0.48) if 0.28 <= hsv_saturation(h) <= 0.45],
        [LEVEL4],
        [h for h in candidates(0, 18, 0.40, 0.90) if hsv_saturation(h) >= 0.55],
    ]
    print("candidate pools:", [len(p) for p in pools])

    random.seed(11)
    best = None
    for _ in range(8000):
        hexes = [random.choice(p) for p in pools]
        deutan = [to_lab(tuple(int(simulate(h, DEUTAN)[i:i + 2], 16) for i in (1, 3, 5)))
                  for h in hexes]
        protan = [to_lab(tuple(int(simulate(h, PROTAN)[i:i + 2], 16) for i in (1, 3, 5)))
                  for h in hexes]
        normal = [to_lab(tuple(int(h[i:i + 2], 16) for i in (1, 3, 5))) for h in hexes]

        def min_pair(labs):
            return min(math.dist(labs[i], labs[j])
                       for i in range(5) for j in range(i + 1, 5))

        score = min(min_pair(deutan), min_pair(protan), min_pair(normal) * 0.5)
        if best is None or score > best[0]:
            best = (score, hexes, min_pair(deutan), min_pair(protan), min_pair(normal))

    score, hexes, d, p, n = best
    print("min pairwise dE76: deutan %.1f  protan %.1f  (normal %.1f)" % (d, p, n))
    for level, hex_ in enumerate(hexes, 1):
        print("  level %d  %s  sat %.2f  contrast %s" % (
            level, hex_, hsv_saturation(hex_),
            "/".join("%.2f" % contrast(hex_, s) for s in SURFACES)))


if __name__ == "__main__":
    main()
