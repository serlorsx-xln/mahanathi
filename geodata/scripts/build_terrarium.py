#!/usr/bin/env python3
"""Re-encode the committed heightmap into terrarium256 raster-DEM tiles.

One-off for ADR 0007 (the MapLibre rebuild): MapLibre's terrain and
hillshade consume a raster-dem source with terrarium encoding
(elevation_m = R*256 + G + B/256 - 32768). Our heightmap is the
committed 1815x2048 16-bit grayscale equirectangular grid over
92-111E, 3-24N (sea = code 0; land = 1 + round(sqrt(h/3300)*2046)),
so this script is a pure re-encode: no new data, no network.

Determinism: same input PNG + same zoom range -> byte-identical tiles
(Pillow PNG encoding is deterministic for fixed inputs and filters).
Run from the project root:

    .cache/venv/bin/python -I geodata/scripts/build_terrarium.py

Outputs to geodata/assets/dem/{z}/{x}/{y}.png for zooms 4..8 (only
tiles whose box intersects the data bbox). Pixels outside the data
bbox extend the nearest edge sample rather than dropping to sea, so
no fake coastline is drawn along the box border; the far edge of the
box fades into the background like the map's horizon.
"""
import math
import os
import sys

from PIL import Image

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
HEIGHTMAP = os.path.join(ROOT, "geodata", "assets", "heightmap.png")
OUT_DIR = os.path.join(ROOT, "geodata", "assets", "dem")

# Must match geodata/scripts/build_heightmap.py and web/src/scene/projection.ts.
LON0, LON1 = 92.0, 111.0
LAT0, LAT1 = 3.0, 24.0          # south, north
HM_CEILING = 3300.0             # metres at code 2047
ZMIN, ZMAX = 4, 8
TILE = 256


def load_heightmap():
    im = Image.open(HEIGHTMAP)
    assert im.mode == "I;16", f"expected 16-bit grayscale, got {im.mode}"
    return im


def decode_height(code):
    """Heightmap code -> metres (sea = 0)."""
    if code <= 0:
        return 0.0
    s = (code - 1) / 2046.0
    return s * s * HM_CEILING


def sample(im, px, u, v):
    """Bilinear sample of the 16-bit grid at normalised u (lon), v (0=north),
    with nearest-edge extension outside [0,1]."""
    w, h = im.size
    x = min(max(u * w - 0.5, 0.0), w - 1.0)
    y = min(max(v * h - 0.5, 0.0), h - 1.0)
    x0, y0 = int(math.floor(x)), int(math.floor(y))
    x1, y1 = min(x0 + 1, w - 1), min(y0 + 1, h - 1)
    fx, fy = x - x0, y - y0
    c00 = px[x0, y0]
    c10 = px[x1, y0]
    c01 = px[x0, y1]
    c11 = px[x1, y1]
    top = c00 * (1 - fx) + c10 * fx
    bot = c01 * (1 - fx) + c11 * fx
    return top * (1 - fy) + bot * fy


def tile_bounds(z, x, y):
    n = 1 << z
    lon0 = x / n * 360.0 - 180.0
    lon1 = (x + 1) / n * 360.0 - 180.0
    # Web Mercator: tile y counts from the north.
    def lat(yy):
        return math.degrees(math.atan(math.sinh(math.pi * (1 - 2 * yy / n))))
    return lon0, lon1, lat(y + 1), lat(y)  # lon0, lon1, south, north


def terrarium_bytes(h):
    v = int(round(h)) + 32768
    if v < 0:
        v = 0
    elif v > 65535:
        v = 65535
    return (v >> 8, v & 0xFF, 0)


def main():
    im = load_heightmap()
    px = im.load()
    os.makedirs(OUT_DIR, exist_ok=True)
    count = 0
    for z in range(ZMIN, ZMAX + 1):
        n = 1 << z
        # Tiles intersecting the data bbox (one-tile margin so the edge
        # extension has neighbours to blend into).
        x0 = max(int((LON0 + 180) / 360 * n) - 1, 0)
        x1 = min(int((LON1 + 180) / 360 * n) + 1, n - 1)
        y0 = max(int((1 - math.asinh(math.tan(math.radians(LAT1))) / math.pi) / 2 * n) - 1, 0)
        y1 = min(int((1 - math.asinh(math.tan(math.radians(LAT0))) / math.pi) / 2 * n) + 1, n - 1)
        for x in range(x0, x1 + 1):
            for y in range(y0, y1 + 1):
                tlon0, tlon1, tsouth, tnorth = tile_bounds(z, x, y)
                row = bytearray()
                for j in range(TILE):
                    # Pixel-centre lat/lon for this row/col.
                    lat = tnorth - (j + 0.5) / TILE * (tnorth - tsouth)
                    v = (LAT1 - lat) / (LAT1 - LAT0)
                    for i in range(TILE):
                        lon = tlon0 + (i + 0.5) / TILE * (tlon1 - tlon0)
                        u = (lon - LON0) / (LON1 - LON0)
                        h = decode_height(sample(im, px, u, v))
                        row += bytes(terrarium_bytes(h))
                tile = Image.frombytes("RGB", (TILE, TILE), bytes(row))
                d = os.path.join(OUT_DIR, str(z), str(x))
                os.makedirs(d, exist_ok=True)
                tile.save(os.path.join(d, f"{y}.png"), optimize=True)
                count += 1
        print(f"z{z}: tiles through x{x0}-{x1}, y{y0}-{y1}", file=sys.stderr)
    print(f"wrote {count} tiles to {OUT_DIR}", file=sys.stderr)


if __name__ == "__main__":
    main()
