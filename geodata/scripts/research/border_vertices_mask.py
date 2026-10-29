"""Vertex counts of the Thai admin-1 source and simplified TopoJSON outputs,
plus the encoded size of a raster Thailand mask on the heightmap grid.

Usage (run with python -I):
  python -I border_vertices_mask.py <naturalearth_dir> <borders_candidates_dir> <out_dir>

The mask uses the same grid as heightmap_measure.py (bbox 92,3,111,24,
lat0 13 deg, square pixels after cos(lat0) scaling).
"""
import io
import json
import math
import os
import sys

import numpy as np
import shapefile
from PIL import Image, ImageDraw
from shapely.geometry import shape

BBOX = (92.0, 3.0, 111.0, 24.0)
LAT0 = 13.0


def grid(rows):
    w, s, e, n = BBOX
    width_km = (e - w) * 111.32 * math.cos(math.radians(LAT0))
    height_km = (n - s) * (110.574 + (111.694 - 110.574) * LAT0 / 90.0)
    return rows, int(round(rows * width_km / height_km))


def main():
    ne, cand, out = sys.argv[1], sys.argv[2], sys.argv[3]
    os.makedirs(out, exist_ok=True)
    r = shapefile.Reader(os.path.join(ne, "ne_10m_admin_1_states_provinces", "ne_10m_admin_1_states_provinces.shp"))
    fields = [f[0] for f in r.fields[1:]]
    ia = fields.index("adm0_a3")
    pts = sum(len(sr.shape.points) for sr in r.iterShapeRecords() if sr.record[ia] == "THA")
    print("source THA admin-1 polygon vertices (rings stored per province, shared edges duplicated):", pts)
    for iv in (250, 500, 1000):
        for nm in ("th_provinces", "land_boundaries", "countries"):
            p = os.path.join(cand, "%s_%dm.topojson" % (nm, iv))
            t = json.load(open(p))
            print("  %-16s interval=%4dm arcs=%5d arc vertices=%6d" % (
                nm, iv, len(t["arcs"]), sum(len(a) for a in t["arcs"])))

    c = shapefile.Reader(os.path.join(ne, "ne_10m_admin_0_countries", "ne_10m_admin_0_countries.shp"))
    cf = [f[0] for f in c.fields[1:]]
    tha = next(shape(sr.shape.__geo_interface__) for sr in c.iterShapeRecords() if sr.record[cf.index("ADM0_A3")] == "THA")
    w, s, e, n = BBOX
    for rows in (1024, 2048):
        H, W = grid(rows)
        ss = 4
        im = Image.new("L", (W * ss, H * ss), 0)
        dr = ImageDraw.Draw(im)
        for pg in getattr(tha, "geoms", [tha]):
            dr.polygon([((x - w) / (e - w) * W * ss, (n - y) / (n - s) * H * ss) for x, y in pg.exterior.coords], fill=255)
        m = im.resize((W, H), Image.BOX)
        b1, b2 = io.BytesIO(), io.BytesIO()
        m.save(b1, format="PNG", optimize=True)
        m.convert("RGB").save(b2, format="WEBP", lossless=True, quality=100, method=6)
        frac = float((np.asarray(m) >= 128).mean())
        print("Thailand mask %dx%d (antialiased 8-bit): png=%d webp=%d bytes, Thai share of pixels=%.3f" % (
            W, H, len(b1.getvalue()), len(b2.getvalue()), frac))
        m.save(os.path.join(out, "tha_mask_%d.png" % rows), optimize=True)


if __name__ == "__main__":
    main()
