"""Natural Earth 10m border facts for Thailand: disputed features near
Thailand, where named sites fall, and vertex counts after simplification.

Usage (run with python -I):
  python -I ne_borders.py <naturalearth_dir> W S E N

Sites are (name, lon, lat) taken from the English Wikipedia coordinates API
(accessed 2026-10-29); see docs/research/geodata-sources.md.
"""
import math
import os
import sys

import shapefile
from shapely.geometry import Point, box, shape
from shapely.ops import unary_union

SITES = [
    ("Prasat Preah Vihear (Phra Wihan)", 104.68027778, 14.39055556),
    ("Prasat Ta Muen Thom", 103.2666, 14.3493),
    ("Ko Kut district (Trat)", 102.54222222, 11.65833333),
]
KX = 111.32 * math.cos(math.radians(14.0))
KY = 110.6


def km_dist(geom, pt):
    """Distance in km using a local scale (fine for distances of a few km)."""
    p1, p2 = None, None
    from shapely.ops import nearest_points
    p1, p2 = nearest_points(geom, pt)
    return math.hypot((p1.x - p2.x) * KX, (p1.y - p2.y) * KY)


def layer(root, name):
    return shapefile.Reader(os.path.join(root, name, name + ".shp"), encoding="utf-8")


def main():
    root = sys.argv[1]
    w, s, e, n = map(float, sys.argv[2:6])
    bb = box(w, s, e, n)

    countries = {}
    for sr in layer(root, "ne_10m_admin_0_countries").iterShapeRecords():
        d = sr.record.as_dict()
        g = shape(sr.shape.__geo_interface__)
        if g.intersects(bb):
            countries[d["ADM0_A3"]] = g
    print("admin-0 units intersecting bbox:", sorted(countries))

    print("\ndisputed areas (polygons) intersecting bbox:")
    for sr in layer(root, "ne_10m_admin_0_disputed_areas").iterShapeRecords():
        g = shape(sr.shape.__geo_interface__)
        if g.intersects(bb):
            d = sr.record.as_dict()
            print("  ", d.get("NAME") or d.get("BRK_NAME"), "|", d.get("NOTE_BRK"), "|", d.get("SOVEREIGNT"), "| type", d.get("TYPE"))
    print("disputed boundary lines intersecting bbox:")
    for sr in layer(root, "ne_10m_admin_0_boundary_lines_disputed_areas").iterShapeRecords():
        g = shape(sr.shape.__geo_interface__)
        if g.intersects(bb):
            d = sr.record.as_dict()
            print("  ", d.get("NAME"), "|", d.get("FEATURECLA"), "|", d.get("BRK_NAME"), "|", d.get("NOTE"))
    print("maritime indicator lines intersecting bbox:")
    for sr in layer(root, "ne_10m_admin_0_boundary_lines_maritime_indicator").iterShapeRecords():
        g = shape(sr.shape.__geo_interface__)
        if g.intersects(bb):
            d = sr.record.as_dict()
            print("  ", d.get("NAME"), "|", d.get("FEATURECLA"), "| len %.2f deg" % g.length)

    th_kh = None
    for sr in layer(root, "ne_10m_admin_0_boundary_lines_land").iterShapeRecords():
        d = sr.record.as_dict()
        if {d.get("ADM0_LEFT"), d.get("ADM0_RIGHT")} == {"Thailand", "Cambodia"}:
            th_kh = shape(sr.shape.__geo_interface__)
    print("\nsites vs NE admin-0 polygons and the Thailand-Cambodia line:")
    for name, lon, lat in SITES:
        p = Point(lon, lat)
        inside = [k for k, g in countries.items() if g.contains(p)]
        print("  %-34s in %-6s  distance to TH-KH line: %.2f km" % (name, inside or "none", km_dist(th_kh, p)))

    tha = countries["THA"]
    print("\nThailand polygon: parts=%d, vertices=%d, area~%.0f km2" % (
        len(getattr(tha, "geoms", [tha])), sum(len(pg.exterior.coords) for pg in getattr(tha, "geoms", [tha])),
        tha.area * KX * KY))
    # Vertex spacing of the TH-KH land line: a proxy for positional resolution.
    parts = [th_kh] if th_kh.geom_type == "LineString" else list(th_kh.geoms)
    seg, nv = [], 0
    for part in parts:
        c = list(part.coords)
        nv += len(c)
        seg += [math.hypot((x1 - x0) * KX, (y1 - y0) * KY) for (x0, y0), (x1, y1) in zip(c[:-1], c[1:])]
    seg.sort()
    print("TH-KH line: %d parts, %d vertices, segment length median %.2f km, p90 %.2f km, max %.2f km" % (
        len(parts), nv, seg[len(seg) // 2], seg[int(len(seg) * 0.9)], seg[-1]))


if __name__ == "__main__":
    main()
