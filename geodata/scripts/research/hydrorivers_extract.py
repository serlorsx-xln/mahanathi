"""Extract HydroRIVERS reaches that intersect a bbox into a compact JSON cache.

Usage (run with python -I):
  python -I hydrorivers_extract.py <HydroRIVERS_v10_as.shp> W S E N <out.json>

Keeps every attribute and the coordinates rounded to 1e-5 degrees.
"""
import json
import sys
import time

import shapefile


def main():
    shp = sys.argv[1]
    w, s, e, n = map(float, sys.argv[2:6])
    out = sys.argv[6]
    t0 = time.time()
    r = shapefile.Reader(shp)
    fields = [f[0] for f in r.fields[1:]]
    total = len(r)
    feats = []
    for sr in r.iterShapeRecords(bbox=[w, s, e, n]):
        d = dict(zip(fields, sr.record))
        d["coords"] = [[round(x, 5), round(y, 5)] for x, y in sr.shape.points]
        feats.append(d)
    with open(out, "w") as f:
        json.dump({"fields": fields, "total_reaches_in_file": total, "bbox": [w, s, e, n],
                   "features": feats}, f)
    print("reaches in file: %d, in bbox: %d, fields: %s, %.0fs" % (total, len(feats), fields, time.time() - t0))


if __name__ == "__main__":
    main()
