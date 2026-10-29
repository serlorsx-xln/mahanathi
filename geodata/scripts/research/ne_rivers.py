"""Natural Earth 10m rivers and lakes inside a bbox, matched against the
13 rivers Mahanathi must show.

Usage (run with python -I):
  python -I ne_rivers.py <naturalearth_dir> W S E N

Prints the field list, every named river feature in the bbox grouped by
name (segments, length in km, scalerank, strokeweig, featurecla), the
match table for the 13 target rivers, Thai-script presence, and lakes.
"""
import collections
import math
import os
import re
import sys

import shapefile
from shapely.geometry import box, shape

THAI = re.compile("[฀-๿]")
TARGETS = [
    ("เจ้าพระยา", "Chao Phraya", r"chao\s*phraya"),
    ("ปิง", "Ping", r"\bping\b"),
    ("วัง", "Wang", r"\bwang\b"),
    ("ยม", "Yom", r"\byom\b"),
    ("น่าน", "Nan", r"\bnan\b"),
    ("ป่าสัก", "Pa Sak", r"pa\s*sak"),
    ("ท่าจีน", "Tha Chin", r"tha\s*chin|suphan|nakhon\s*chai\s*si"),
    ("แม่กลอง", "Mae Klong", r"mae\s*klong|meklong|khwae|kwai"),
    ("มูล", "Mun", r"\bmun\b|\bmoon\b"),
    ("ชี", "Chi", r"\bchi\b"),
    ("โขง", "Mekong", r"mekong|lancang"),
    ("ตาปี", "Tapi", r"\btapi\b|ta\s*pi\b"),
    ("บางปะกง", "Bang Pakong", r"bang\s*pakong|prachin"),
]


def geod_len_km(geom):
    """Length of a (multi)line in km using a local equirectangular approximation."""
    total = 0.0
    lines = [geom] if geom.geom_type == "LineString" else list(getattr(geom, "geoms", []))
    for ln in lines:
        c = list(ln.coords)
        for (x0, y0), (x1, y1) in zip(c[:-1], c[1:]):
            k = math.cos(math.radians((y0 + y1) / 2))
            total += math.hypot((x1 - x0) * k * 111.32, (y1 - y0) * 110.6)
    return total


def layer(root, name):
    return shapefile.Reader(os.path.join(root, name, name + ".shp"), encoding="utf-8")


def rivers(root, name, bb):
    r = layer(root, name)
    fields = [f[0] for f in r.fields[1:]]
    print("\n== %s  (fields: %d)" % (name, len(fields)))
    print("fields:", fields)
    print("name_* fields:", [f for f in fields if f.lower().startswith("name")])
    groups = collections.defaultdict(list)
    thai_fields = set()
    nfeat = 0
    for sr in r.iterShapeRecords():
        b = sr.shape.bbox
        if b[2] < bb.bounds[0] or b[0] > bb.bounds[2] or b[3] < bb.bounds[1] or b[1] > bb.bounds[3]:
            continue
        g = shape(sr.shape.__geo_interface__).intersection(bb)
        if g.is_empty:
            continue
        nfeat += 1
        d = sr.record.as_dict()
        for k, v in d.items():
            if isinstance(v, str) and THAI.search(v):
                thai_fields.add(k)
        key = d.get("name") or "(unnamed)"
        groups[key].append((d, g))
    print("features intersecting bbox:", nfeat, "| distinct names:", len(groups))
    print("fields with Thai script in bbox:", sorted(thai_fields) or "none")
    rows = []
    for nm, items in groups.items():
        km = sum(geod_len_km(g) for _, g in items)
        sr_ = sorted(set(d.get("scalerank") for d, _ in items))
        sw = sorted(set(d.get("strokeweig") for d, _ in items if d.get("strokeweig") is not None))
        fc = sorted(set(d.get("featurecla") for d, _ in items))
        mz = sorted(set(d.get("min_zoom") for d, _ in items))
        en = sorted(set(str(d.get("name_en")) for d, _ in items))
        alt = sorted(set(str(d.get("name_alt")) for d, _ in items if d.get("name_alt")))
        rows.append((nm, len(items), km, sr_, sw, fc, mz, en, alt))
    rows.sort(key=lambda r: -r[2])
    print("%-28s %4s %8s  %-10s %-14s %-22s %-10s %s" % ("name", "segs", "km", "scalerank", "strokeweig", "featurecla", "min_zoom", "name_en / name_alt"))
    for nm, n, km, sr_, sw, fc, mz, en, alt in rows:
        print("%-28s %4d %8.0f  %-10s %-14s %-22s %-10s %s %s" % (nm[:28], n, km, sr_, [round(x, 2) for x in sw], ",".join(fc)[:22], mz, en, alt))
    print("\ntarget match:")
    for th, en, pat in TARGETS:
        hits = [r for r in rows if re.search(pat, (r[0] + " " + " ".join(r[7]) + " " + " ".join(r[8])).lower())]
        if hits:
            for h in hits:
                print("  %-10s %-12s -> %-24s segs=%d km=%.0f scalerank=%s strokeweig=%s" % (th, en, h[0], h[1], h[2], h[3], [round(x, 2) for x in h[4]]))
        else:
            print("  %-10s %-12s -> NOT PRESENT" % (th, en))


def lakes(root, bb):
    r = layer(root, "ne_10m_lakes")
    fields = [f[0] for f in r.fields[1:]]
    print("\n== ne_10m_lakes fields:", fields)
    for sr in r.iterShapeRecords():
        b = sr.shape.bbox
        if b[2] < bb.bounds[0] or b[0] > bb.bounds[2] or b[3] < bb.bounds[1] or b[1] > bb.bounds[3]:
            continue
        d = sr.record.as_dict()
        g = shape(sr.shape.__geo_interface__)
        c = g.representative_point()
        k = math.cos(math.radians(c.y))
        area = g.area * 111.32 * k * 110.6
        print("  %-30s featurecla=%-12s scalerank=%s area~%.0f km2 at (%.2f, %.2f) name_alt=%s" % (
            d.get("name"), d.get("featurecla"), d.get("scalerank"), area, c.x, c.y, d.get("name_alt")))


def main():
    root = sys.argv[1]
    w, s, e, n = map(float, sys.argv[2:6])
    bb = box(w, s, e, n)
    rivers(root, "ne_10m_rivers_lake_centerlines", bb)
    rivers(root, "ne_10m_rivers_lake_centerlines_scale_rank", bb)
    lakes(root, bb)


if __name__ == "__main__":
    main()
