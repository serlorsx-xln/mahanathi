"""Inspect Natural Earth 10m admin layers for the Mahanathi geodata research.

Usage (run with python -I):
  python -I ne_inspect.py <naturalearth_dir>

<naturalearth_dir> holds one unzipped directory per layer, for example
<dir>/ne_10m_admin_1_states_provinces/ne_10m_admin_1_states_provinces.shp

Prints: field lists, language name fields, Thai admin-1 count and names,
whether any Thai-script names exist, Thailand POV class fields, and every
land boundary line touching Thailand with its featurecla.
"""
import os
import re
import sys

import shapefile
from shapely.geometry import shape

THAI = re.compile("[฀-๿]")


def open_layer(root, name):
    return shapefile.Reader(os.path.join(root, name, name + ".shp"), encoding="utf-8")


def main():
    root = sys.argv[1]

    a0 = open_layer(root, "ne_10m_admin_0_countries")
    f0 = [f[0] for f in a0.fields[1:]]
    print("admin0 fields (%d):" % len(f0), f0)
    print("admin0 name_* fields:", [f for f in f0 if f.lower().startswith("name_")])
    print("admin0 fclass_* fields:", [f for f in f0 if f.lower().startswith("fclass_")])
    for rec in a0.iterRecords():
        d = rec.as_dict()
        if d["ADM0_A3"] in ("THA", "KHM", "LAO", "MMR", "MYS", "VNM", "CHN"):
            print(" ", d["ADM0_A3"], d["NAME"], "TYPE", d["TYPE"], "FCLASS_ISO", d.get("FCLASS_ISO"),
                  "MIN_ZOOM", d["MIN_ZOOM"], "NAME_TH?", d.get("NAME_TH"))

    a1 = open_layer(root, "ne_10m_admin_1_states_provinces")
    f1 = [f[0] for f in a1.fields[1:]]
    print("\nadmin1 fields (%d):" % len(f1), f1)
    print("admin1 name_* fields:", [f for f in f1 if f.lower().startswith("name_")])
    tha = [r.as_dict() for r in a1.iterRecords() if r.as_dict()["adm0_a3"] == "THA"]
    print("admin1 THA rows:", len(tha))
    thai_script_fields = set()
    for d in tha:
        for k, v in d.items():
            if isinstance(v, str) and THAI.search(v):
                thai_script_fields.add(k)
    print("admin1 fields containing Thai script in THA rows:", sorted(thai_script_fields))
    keys = ["name", "name_en", "name_local", "name_alt", "iso_3166_2", "type_en", "region", "gn_name", "woe_name"]
    for d in sorted(tha, key=lambda d: d["name"] or ""):
        print("  ", " | ".join(str(d.get(k)) for k in keys))
    for probe in ("Bueng Kan", "Bangkok", "Nong Khai"):
        hits = [d["name"] for d in tha if probe.lower() in (d["name"] or "").lower()]
        print("probe", probe, hits)

    bl = open_layer(root, "ne_10m_admin_0_boundary_lines_land")
    fb = [f[0] for f in bl.fields[1:]]
    print("\nboundary_lines_land fields:", fb)
    for sr in bl.iterShapeRecords():
        d = sr.record.as_dict()
        pair = (d.get("ADM0_LEFT"), d.get("ADM0_RIGHT"))
        if "Thailand" in pair:
            g = shape(sr.shape.__geo_interface__)
            print("  %-10s | %-10s | %-28s | len=%.3f deg | pts=%d | min_zoom=%s scalerank=%s" % (
                pair[0], pair[1], d.get("FEATURECLA"), g.length, len(sr.shape.points),
                d.get("MIN_ZOOM"), d.get("SCALERANK")))


if __name__ == "__main__":
    main()
