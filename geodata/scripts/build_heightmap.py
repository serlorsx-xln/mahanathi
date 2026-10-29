"""Build the shipped heightmap from the cached Terrarium tiles.

Production pipeline: one command, deterministic, reading only from the
gitignored geodata/.cache/ and writing geodata/assets/heightmap.png.

Encoding (decisions measured in docs/research/geodata-sources.md):
- a 1815 x 2048 px (width x height) equirectangular grid over
  92-111 E, 3-24 N, pixels square at 13 N;
- 16-bit grayscale; the sea is reserved code 0;
- land is 1 + round(sqrt(h / HM) * 2046) with h clamped to >= 0,
  where HM is the land height maximum rounded up to a whole 100 m
  (printed by this script; the client decoder needs the same constant);
- Thailand keeps full precision; neighbours are box-blurred (r = 2,
  three passes) and coarsely quantised, blended across a 75 km band.

Usage: .cache/venv/bin/python -I geodata/scripts/build_heightmap.py
"""
import math
import os
import sys

import numpy as np
from PIL import Image

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "research"))
import heightmap_measure as hm  # noqa: E402

BBOX = (92.0, 3.0, 111.0, 24.0)
LAT0 = 13.0
ROWS = 2048
SQRT_BITS = 11
NB_BLUR = 2
BAND_KM = 75.0


def build(tiles: str, z: int, land_shp: str, tha_shp: str):
    w, s, e, n = BBOX
    km_per_deg_lat = 110.574 + (111.694 - 110.574) * (LAT0 / 90.0)
    km_per_deg_lon = 111.320 * math.cos(math.radians(LAT0))
    width_km = (e - w) * km_per_deg_lon
    height_km = (n - s) * km_per_deg_lat

    mos, x0, y0 = hm.stitch(tiles, z, BBOX)
    ref_rows = int(round((hm.lat2y(s, z) - hm.lat2y(n, z)) * 256))
    ref_cols = int(round(ref_rows * width_km / height_km))
    eq = hm.reproject(mos, x0, y0, z, BBOX, ref_rows, ref_cols)

    cols = int(round(ROWS * width_km / height_km))
    hf = np.asarray(Image.fromarray(eq).resize((cols, ROWS), Image.BOX), dtype=np.float32)

    ne_land = hm.raster_land(land_shp, BBOX, ROWS, cols) >= 0.5
    is_land = ne_land | (hf > 1.0)
    ref = np.where(is_land, np.maximum(hf, 0.0), 0.0).astype(np.float32)
    # The encoding ceiling, exactly as the research computed it: the land
    # maximum rounded up to a whole 100 m. It is part of the format.
    hm_ceiling = math.ceil(float(ref.max()) / 100.0) * 100.0

    km_px = height_km / ROWS
    tmask = hm.raster_land(tha_shp, BBOX, ROWS, cols, ss=2, only_a3="THA")
    band_px = max(1, int(round(BAND_KM / km_px)))
    wgt = np.clip(hm.box_blur(tmask, band_px) * 2.0, 0.0, 1.0)
    wgt = np.maximum(wgt, (tmask >= 0.5).astype(np.float32))

    nb = hm.box_blur(hm.box_blur(hm.box_blur(ref, NB_BLUR), NB_BLUR), NB_BLUR)
    nb = np.where(is_land, np.maximum(nb, 0), 0)

    top = (1 << SQRT_BITS) - 2
    coarse = 4  # the divisor measured for 11-bit in the research
    fine = np.round(np.sqrt(np.maximum(ref, 0) / hm_ceiling) * top)
    crs = np.round(np.round(np.sqrt(nb / hm_ceiling) * top) / coarse) * coarse
    v = np.where(wgt >= 0.999, fine, np.round(wgt * fine + (1 - wgt) * crs))
    u = np.where(is_land, 1 + v, 0).astype(np.uint16)
    return u, hm_ceiling


def main() -> None:
    out = "geodata/assets/heightmap.png"
    u, hm_ceiling = build(
        tiles="geodata/.cache/terrarium",
        z=8,
        land_shp="geodata/.cache/naturalearth/ne_10m_land/ne_10m_land.shp",
        tha_shp="geodata/.cache/naturalearth/ne_10m_admin_0_countries/ne_10m_admin_0_countries.shp",
    )
    os.makedirs(os.path.dirname(out), exist_ok=True)
    _, best = hm.png_bytes(Image.fromarray(u))
    with open(out, "wb") as f:
        f.write(best)
    print("wrote %s (%d bytes, %d x %d px, HM=%d m)" % (out, len(best), u.shape[1], u.shape[0], hm_ceiling))


if __name__ == "__main__":
    main()
