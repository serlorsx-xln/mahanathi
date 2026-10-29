"""Positional agreement between OSM river centre lines and other sources.

Usage (run with python -I):
  python -I river_agreement.py <osm rivers_geom.json> <hr_bbox.json> \
      <ne_10m_rivers_lake_centerlines_scale_rank.shp>

Samples each OSM centre line (clipped to Thailand plus 0.5 degree) every
1 km and reports the share of samples that have a HydroRIVERS reach with
UPLAND_SKM >= 500, or any Natural Earth 10m river line, within 1, 2 and 5 km,
plus the median distance. Distances are computed in a local equirectangular
km grid (x scaled by cos 13 deg), which is accurate to a few per cent here.
"""
import json
import math
import sys

import numpy as np
import shapefile
import shapely
from shapely.geometry import LineString, MultiLineString, box

KX = 111.32 * math.cos(math.radians(13.0))
KY = 110.6
CLIP = box(96.8, 5.1, 106.2, 21.0)


def to_km(g):
    return shapely.transform(g, lambda c: np.column_stack([c[:, 0] * KX, c[:, 1] * KY]))


def osm_centre(path):
    d = json.load(open(path))
    out = {}
    for rel in d["elements"]:
        lines = []
        for m in rel.get("members", []):
            if m["type"] == "way" and "geometry" in m and (m.get("role") or "") in ("main_stream", ""):
                pts = [(p["lon"], p["lat"]) for p in m["geometry"] if p]
                if len(pts) >= 2:
                    lines.append(LineString(pts))
        out[rel["tags"].get("name:en") or rel["tags"].get("name")] = MultiLineString(lines).intersection(CLIP)
    return out


def samples(g, step=1.0):
    pts = []
    for ln in (g.geoms if hasattr(g, "geoms") else [g]):
        if ln.geom_type != "LineString" or ln.length == 0:
            continue
        n = max(int(ln.length / step), 1)
        pts += [ln.interpolate(i * ln.length / n) for i in range(n + 1)]
    return pts


def main():
    osm = osm_centre(sys.argv[1])
    hr = json.load(open(sys.argv[2]))
    hr_lines = [to_km(LineString(f["coords"])) for f in hr["features"]
                if f["UPLAND_SKM"] >= 500 and len(f["coords"]) > 1]
    ne_lines, ne_names = [], []
    rd = shapefile.Reader(sys.argv[3], encoding="utf-8")
    for sr in rd.iterShapeRecords():
        b = sr.shape.bbox
        if b[2] < 92 or b[0] > 111 or b[3] < 3 or b[1] > 24:
            continue
        ne_lines.append(to_km(shapely.geometry.shape(sr.shape.__geo_interface__)))
        ne_names.append(sr.record.as_dict().get("name") or "(unnamed)")
    t_hr, t_ne = shapely.STRtree(hr_lines), shapely.STRtree(ne_lines)
    print("HydroRIVERS reaches with UPLAND>=500 km2 in bbox: %d; NE lines: %d" % (len(hr_lines), len(ne_lines)))
    print("%-18s %7s | %-24s %6s | %-24s %6s" % ("OSM river", "km", "HR within 1/2/5 km", "med", "NE within 1/2/5 km", "med"))
    for name, g in osm.items():
        gk = to_km(g)
        pts = samples(gk)
        if not pts:
            continue
        res = []
        for tree in (t_hr, t_ne):
            idx, dist = tree.query_nearest(pts, return_distance=True, all_matches=False)
            dist = np.asarray(dist)
            res.append(("%3.0f%% %3.0f%% %3.0f%%" % tuple(100 * np.mean(dist <= r) for r in (1, 2, 5)),
                        float(np.median(dist))))
        # Which NE names carry the samples that lie within 2 km of an NE line.
        near = [ne_names[j] for j, dd in zip(np.asarray(idx[1]), dist) if dd <= 2.0]
        tally = {}
        for nm in near:
            tally[nm] = tally.get(nm, 0) + 1
        top = ", ".join("%s %d%%" % (k, round(100 * v / len(pts))) for k, v in
                        sorted(tally.items(), key=lambda kv: -kv[1])[:3])
        print("%-18s %7.0f | %-24s %6.2f | %-24s %6.2f | NE names within 2 km: %s" % (
            name[:18], gk.length, res[0][0], res[0][1], res[1][0], res[1][1], top or "-"))


if __name__ == "__main__":
    main()
