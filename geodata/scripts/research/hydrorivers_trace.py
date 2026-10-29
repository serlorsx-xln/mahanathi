"""Trace the 13 Mahanathi rivers through HydroRIVERS and compare with OSM.

Research script, not the production pipeline.

Usage (run with python -I):
  python -I hydrorivers_trace.py <hr_bbox.json> <osm rivers_geom.json>

For each river a rough downstream seed point (lon, lat) is snapped to the
HydroRIVERS reach with the largest upstream area that lies within 8 km of
the seed and mostly within 3 km of the OSM centre line. From there two
upstream traces are made:
  maxup  : at every confluence follow the branch with the largest UPLAND_SKM
  guided : follow the branch whose vertices lie most within 3 km of the OSM
           centre line; stop when no branch has at least 50 % of its
           vertices inside that corridor
Reports length, agreement with OSM, and order / discharge statistics, then
size estimates for a network subset of the bbox.
"""
import collections
import json
import math
import sys

import brotli
import numpy as np
import shapely
from shapely.geometry import LineString, MultiLineString, Point
from shapely.ops import unary_union

# Rough downstream ends (lon, lat). Only used as snapping seeds; they are
# approximate by design and every result is checked against OSM geometry.
SEEDS = {
    "แม่น้ำเจ้าพระยา": ("Chao Phraya", (100.585, 13.545)),
    "แม่น้ำปิง": ("Ping", (100.115, 15.705)),
    "แม่น้ำน่าน": ("Nan", (100.140, 15.705)),
    "แม่น้ำยม": ("Yom", (100.300, 15.880)),
    "แม่น้ำวัง": ("Wang", (99.080, 17.050)),
    "แม่น้ำป่าสัก": ("Pa Sak", (100.580, 14.350)),
    "แม่น้ำท่าจีน": ("Tha Chin", (100.270, 13.520)),
    "แม่น้ำแม่กลอง": ("Mae Klong", (99.995, 13.360)),
    "แม่น้ำมูล": ("Mun", (105.500, 15.320)),
    "แม่น้ำชี": ("Chi", (104.720, 15.200)),
    "Mekong": ("Mekong", (104.930, 11.580)),
    "แม่น้ำตาปี": ("Tapi", (99.330, 9.150)),
    "แม่น้ำบางปะกง": ("Bang Pakong", (100.960, 13.470)),
}
CORRIDOR_DEG = 0.03  # about 3.3 km


def km_of(coords):
    t = 0.0
    for (x0, y0), (x1, y1) in zip(coords[:-1], coords[1:]):
        k = math.cos(math.radians((y0 + y1) / 2))
        t += math.hypot((x1 - x0) * k * 111.32, (y1 - y0) * 110.6)
    return t


def osm_lines(path):
    d = json.load(open(path))
    out = {}
    for rel in d["elements"]:
        name = rel["tags"].get("name")
        lines = []
        for m in rel.get("members", []):
            if m["type"] == "way" and "geometry" in m and (m.get("role") or "") in ("main_stream", ""):
                pts = [(p["lon"], p["lat"]) for p in m["geometry"] if p]
                if len(pts) >= 2:
                    lines.append(LineString(pts))
        out[name] = MultiLineString(lines)
    return out


def frac_inside(corr, coords):
    """Share of a reach's length that lies inside the OSM corridor."""
    if len(coords) < 2:
        return float(shapely.contains_xy(corr, coords[0][0], coords[0][1]))
    ln = LineString(coords)
    return float(ln.intersection(corr).length / ln.length) if ln.length > 0 else 0.0


def main():
    hr = json.load(open(sys.argv[1]))
    osm = osm_lines(sys.argv[2])
    feats = {f["HYRIV_ID"]: f for f in hr["features"]}
    up = collections.defaultdict(list)
    for f in hr["features"]:
        up[f["NEXT_DOWN"]].append(f["HYRIV_ID"])
    # Spatial index on reach geometries for seed snapping.
    ids = list(feats)
    geoms = [LineString(feats[i]["coords"]) if len(feats[i]["coords"]) > 1 else Point(feats[i]["coords"][0]) for i in ids]
    tree = shapely.STRtree(geoms)

    print("reaches in bbox:", len(feats))
    hdr = ("river", "osm_km", "maxup_km", "maxup_in%", "maxup_ok_km", "guided_km", "osm_cov%",
           "STRA", "CLAS", "FLOW", "UPLAND_max", "DIS_max", "reaches")
    print("%-12s %7s %8s %9s %11s %9s %8s %7s %6s %6s %10s %8s %7s" % hdr)
    traces = {}
    for osm_name, (en, seed) in SEEDS.items():
        line = osm[osm_name]
        corr = line.buffer(CORRIDOR_DEG)
        shapely.prepare(corr)
        osm_km = sum(km_of(list(l.coords)) for l in line.geoms)
        cand = tree.query(Point(seed).buffer(0.075))

        def guided_trace(start, max_misses=6):
            """Follow the upstream branch with the largest corridor share,
            tolerating up to max_misses consecutive reaches below 50 %."""
            cur, out, misses, pending = start, [], 0, []
            while cur is not None:
                pending.append(cur)
                if frac_inside(corr, cur["coords"]) >= 0.5:
                    out.extend(pending)
                    pending, misses = [], 0
                else:
                    misses += 1
                    if misses > max_misses:
                        break
                ups = [feats[u] for u in up.get(cur["HYRIV_ID"], []) if u in feats]
                scored = [(frac_inside(corr, f["coords"]), f["UPLAND_SKM"], f) for f in ups]
                scored = [x for x in scored if x[0] >= 0.2]
                # Among branches at least half inside the corridor the largest
                # upstream area wins (keeps the Mekong over small streams that
                # sit inside a wide river's corridor); otherwise best overlap.
                cur = max(scored, key=lambda x: (x[0] >= 0.5, x[1] if x[0] >= 0.5 else x[0]))[2] if scored else None
            return out

        # Seed = the candidate (>= 50 % in corridor) whose guided trace is
        # longest. This avoids picking the receiving river at a confluence.
        best, guided = None, []
        for ci in cand:
            f = feats[ids[ci]]
            if frac_inside(corr, f["coords"]) < 0.5:
                continue
            g = guided_trace(f)
            if sum(x["LENGTH_KM"] for x in g) > sum(x["LENGTH_KM"] for x in guided):
                best, guided = f, g
        if best is None:
            print("%-12s no seed reach found" % en)
            continue

        # max-upland trace from the same seed; how far it stays in the corridor
        cur, maxup, ok_km, left = best, [], 0.0, False
        while cur is not None:
            maxup.append(cur)
            if not left and frac_inside(corr, cur["coords"]) >= 0.5:
                ok_km += cur["LENGTH_KM"]
            else:
                left = True
            ups = [feats[u] for u in up.get(cur["HYRIV_ID"], []) if u in feats]
            cur = max(ups, key=lambda f: f["UPLAND_SKM"]) if ups else None
        maxup_km = sum(f["LENGTH_KM"] for f in maxup)
        inside = sum(f["LENGTH_KM"] for f in maxup if frac_inside(corr, f["coords"]) >= 0.5)

        g_km = sum(f["LENGTH_KM"] for f in guided)
        gline = MultiLineString([LineString(f["coords"]) for f in guided if len(f["coords"]) > 1])
        gcorr = gline.buffer(CORRIDOR_DEG)
        cov_km = sum(km_of(list(p.coords)) for l in line.geoms
                     for p in ([l.intersection(gcorr)] if l.intersection(gcorr).geom_type == "LineString"
                               else getattr(l.intersection(gcorr), "geoms", [])) if p.geom_type == "LineString")
        stra = [f["ORD_STRA"] for f in guided]
        clas = [f["ORD_CLAS"] for f in guided]
        flow = [f["ORD_FLOW"] for f in guided]
        print("%-12s %7.0f %8.0f %8.0f%% %11.0f %9.0f %7.0f%% %7s %6s %6s %10.0f %8.0f %7d" % (
            en, osm_km, maxup_km, 100 * inside / maxup_km, ok_km, g_km, 100 * cov_km / osm_km,
            "%d-%d" % (min(stra), max(stra)), "%d-%d" % (min(clas), max(clas)),
            "%d-%d" % (min(flow), max(flow)), max(f["UPLAND_SKM"] for f in guided),
            max(f["DIS_AV_CMS"] for f in guided), len(guided)))
        traces[en] = guided
        top = guided[-1]["coords"]  # HydroRIVERS lines run downstream: [0] is the upstream end
        dep = next((f for f in maxup if frac_inside(corr, f["coords"]) < 0.5), None)
        print("    seed reach UPLAND=%.0f; guided stops at (%.3f, %.3f) DIST_UP_KM=%.0f; maxup first leaves at %s" % (
            best["UPLAND_SKM"], top[0][0], top[0][1], guided[-1]["DIST_UP_KM"],
            "(%.3f, %.3f)" % tuple(dep["coords"][0]) if dep else "never"))

    # Network subsets for the whole bbox and their compact sizes.
    print("\nnetwork subsets in bbox (simplified at 0.005 deg, quantised to 1e-3 deg):")
    for label, pred in [
        ("ORD_FLOW<=4 (>=100 m3/s)", lambda f: f["ORD_FLOW"] <= 4),
        ("ORD_FLOW<=5 (>=10 m3/s)", lambda f: f["ORD_FLOW"] <= 5),
        ("ORD_STRA>=5", lambda f: f["ORD_STRA"] >= 5),
        ("ORD_STRA>=6", lambda f: f["ORD_STRA"] >= 6),
        ("UPLAND>=1000 km2", lambda f: f["UPLAND_SKM"] >= 1000),
        ("UPLAND>=5000 km2", lambda f: f["UPLAND_SKM"] >= 5000),
        ("13 named traces only", None),
    ]:
        sel = [f for t in traces.values() for f in t] if pred is None else [f for f in hr["features"] if pred(f)]
        if not sel:
            continue
        verts, km_total, buf = 0, 0.0, bytearray()
        for f in sel:
            ls = LineString(f["coords"]) if len(f["coords"]) > 1 else None
            if ls is None:
                continue
            c = np.asarray(ls.simplify(0.005).coords)
            q = np.round(c * 1000).astype(np.int32)
            d = np.diff(q, axis=0, prepend=np.zeros((1, 2), dtype=np.int32))
            d[0] = q[0] - [92000, 3000]
            buf += np.int16(len(q)).tobytes() + np.int8(min(f["ORD_STRA"], 127)).tobytes() + d.astype(np.int16).tobytes()
            verts += len(q)
            km_total += f["LENGTH_KM"]
        gj = json.dumps({"type": "FeatureCollection", "features": [
            {"type": "Feature", "properties": {"o": f["ORD_STRA"]},
             "geometry": {"type": "LineString", "coordinates": [[round(x, 3), round(y, 3)] for x, y in
                                                                 LineString(f["coords"]).simplify(0.005).coords]}}
            for f in sel if len(f["coords"]) > 1]}, separators=(",", ":")).encode()
        print("  %-26s reaches=%6d km=%7.0f verts=%7d bin=%7d bin.br=%7d geojson=%8d geojson.br=%7d" % (
            label, len(sel), km_total, verts, len(buf), len(brotli.compress(bytes(buf), quality=11)),
            len(gj), len(brotli.compress(gj, quality=11))))


if __name__ == "__main__":
    main()
