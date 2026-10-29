"""Compare Natural Earth admin-1 Thai names with Wikidata Thai labels, and
list Wikidata Thai labels for the OSM river relations.

Usage (run with python -I):
  python -I thai_labels_check.py <naturalearth_dir> <wikidata labels.json> <osm rivers_relations.json>

A province counts as a match when name_local, with the prefix "จังหวัด"
removed, equals the Wikidata Thai label with "จังหวัด" removed.
"""
import json
import os
import sys

import shapefile


def strip(s):
    s = (s or "").strip()
    for p in ("จังหวัด",):
        if s.startswith(p):
            s = s[len(p):]
    return s


def main():
    ne, wd, osm = sys.argv[1:4]
    lab = {}
    for b in json.load(open(wd))["results"]["bindings"]:
        q = b["item"]["value"].rsplit("/", 1)[1]
        lab[q] = (b.get("th", {}).get("value"), b.get("en", {}).get("value"))
    r = shapefile.Reader(os.path.join(ne, "ne_10m_admin_1_states_provinces", "ne_10m_admin_1_states_provinces.shp"), encoding="utf-8")
    rows = [d for d in (x.as_dict() for x in r.iterRecords()) if d["adm0_a3"] == "THA"]
    exact, prefix_only, wrong = [], [], []
    for d in rows:
        th_wd = lab.get(d["wikidataid"], (None, None))[0]
        nl = d["name_local"]
        if nl == th_wd:
            exact.append(d)
        elif strip(nl) == strip(th_wd):
            prefix_only.append((d["name"], nl, th_wd))
        else:
            wrong.append((d["name"], d["iso_3166_2"], nl, th_wd, d["wikidataid"]))
    print("THA admin-1 rows: %d; Wikidata Thai label found for %d" % (
        len(rows), sum(1 for d in rows if lab.get(d["wikidataid"], (None,))[0])))
    print("name_local == Wikidata label exactly: %d" % len(exact))
    print("same name, differs only by the 'จังหวัด' prefix: %d" % len(prefix_only))
    print("different name: %d" % len(wrong))
    for w in wrong:
        print("   %-26s %-6s NE name_local=%-24s Wikidata th=%-24s %s" % w)

    print("\nOSM river relations: Wikidata Thai and English labels")
    for e in json.load(open(osm))["elements"]:
        t = e["tags"]
        q = t.get("wikidata")
        th, en = lab.get(q, (None, None))
        print("   %-18s osm name:th=%-18s wikidata %-10s th=%-18s en=%s" % (
            t.get("name"), t.get("name:th"), q, th, en))


if __name__ == "__main__":
    main()
