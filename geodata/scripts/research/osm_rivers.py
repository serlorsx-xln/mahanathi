"""Measure OSM river relations fetched with Overpass `out geom`.

Usage (run with python -I):
  python -I osm_rivers.py <rivers_geom.json> W S E N [--out DIR]

For each relation: member ways by role, raw vertex count, length (km) of
all ways and of centre-line ways only, how many disconnected pieces the
centre line falls into after merging, and vertex counts after
Douglas-Peucker simplification at 0.002 and 0.005 degrees. Optionally
(--out DIR) writes the simplified centre lines of all relations as one
GeoJSON per tolerance (coordinates rounded to 4 decimals, properties: OSM
relation id and wikidata) and prints raw, gzip -9 and brotli q11 sizes.
"""
import collections
import gzip
import json
import math
import os
import sys

import brotli

from shapely.geometry import LineString, MultiLineString, box
from shapely.ops import linemerge


def km(line):
    c = list(line.coords)
    t = 0.0
    for (x0, y0), (x1, y1) in zip(c[:-1], c[1:]):
        k = math.cos(math.radians((y0 + y1) / 2))
        t += math.hypot((x1 - x0) * k * 111.32, (y1 - y0) * 110.6)
    return t


def parts(g):
    if g.is_empty:
        return []
    if g.geom_type == "LineString":
        return [g]
    return [p for p in getattr(g, "geoms", []) if p.geom_type == "LineString"]


def main():
    d = json.load(open(sys.argv[1]))
    w, s, e, n = map(float, sys.argv[2:6])
    out = sys.argv[sys.argv.index("--out") + 1] if "--out" in sys.argv else None
    feats = {0.002: [], 0.005: []}
    clip = box(w, s, e, n)
    print("osm base timestamp:", d.get("osm3s", {}).get("timestamp_osm_base"))
    print("%-18s %-16s %5s %7s %7s %7s %6s %7s %7s  %s" % (
        "name", "name:en", "ways", "nodes", "km_all", "km_ctr", "pieces", "v@.002", "v@.005", "roles"))
    for rel in d["elements"]:
        t = rel["tags"]
        roles = collections.Counter()
        lines_all, lines_ctr = [], []
        nodes = 0
        for m in rel.get("members", []):
            if m["type"] != "way" or "geometry" not in m:
                roles["(non-way:%s)" % m["type"]] += 1
                continue
            role = m.get("role") or "(empty)"
            roles[role] += 1
            pts = [(p["lon"], p["lat"]) for p in m["geometry"] if p]
            nodes += len(pts)
            if len(pts) < 2:
                continue
            ln = LineString(pts)
            lines_all.append(ln)
            if role in ("main_stream", "(empty)"):
                lines_ctr.append(ln)
        ctr = MultiLineString(lines_ctr).intersection(clip) if lines_ctr else MultiLineString()
        merged = linemerge(parts(ctr)) if not ctr.is_empty else ctr
        pieces = parts(merged)
        km_all = sum(km(p) for l in lines_all for p in parts(l.intersection(clip)))
        km_ctr = sum(km(p) for p in pieces)
        v2 = sum(len(p.simplify(0.002).coords) for p in pieces)
        v5 = sum(len(p.simplify(0.005).coords) for p in pieces)
        for tol in feats:
            geoms = [[[round(x, 4), round(y, 4)] for x, y in p.simplify(tol).coords] for p in pieces]
            feats[tol].append({"type": "Feature", "properties": {"osm": rel["id"], "wikidata": t.get("wikidata")},
                               "geometry": {"type": "MultiLineString", "coordinates": geoms}})
        # Largest gap between a piece and its nearest other piece (km, local scale).
        gap = 0.0
        if len(pieces) > 1:
            for i, p in enumerate(pieces):
                others = MultiLineString([q for j, q in enumerate(pieces) if j != i])
                gap = max(gap, p.distance(others) * 110.0)
        print("%-18s %-16s %5d %7d %7.0f %7.0f %6d %7d %7d  gap<=%.1fkm  %s" % (
            t.get("name", "")[:18], (t.get("name:en") or "-")[:16], sum(roles.values()), nodes,
            km_all, km_ctr, len(pieces), v2, v5, gap, dict(roles)))
    if out:
        os.makedirs(out, exist_ok=True)
        for tol, fs in feats.items():
            b = json.dumps({"type": "FeatureCollection", "features": fs}, separators=(",", ":")).encode()
            name = os.path.join(out, "osm13_dp%s.geojson" % str(tol).replace("0.", ""))
            open(name, "wb").write(b)
            print("simplified %.3f deg: vertices=%d raw=%d gzip9=%d brotli11=%d  %s" % (
                tol, sum(len(l) for f in fs for l in f["geometry"]["coordinates"]), len(b),
                len(gzip.compress(b, 9)), len(brotli.compress(b, quality=11)), name))


if __name__ == "__main__":
    main()
