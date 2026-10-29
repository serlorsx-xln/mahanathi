"""Download AWS Terrain Tiles (Terrarium PNG) for a lon/lat bbox at one zoom.

Research script, not the production pipeline. Polite by construction:
at most 4 concurrent requests, a project User-Agent, and tiles already on
disk are never fetched again.

Usage (run with python -I):
  python -I fetch_terrarium.py --z 7 --bbox 92.5 3.5 110 24 \
      --out geodata/.cache/terrarium

Writes <out>/<z>/<x>/<y>.png and <out>/<z>/headers.json (status, bytes and
the x-amz-meta-x-imagery-sources header for each tile). With --heads it only
issues HEAD requests and refreshes headers.json.
"""
import argparse
import json
import math
import os
import sys
import time
import urllib.request
from concurrent.futures import ThreadPoolExecutor

URL = "https://s3.amazonaws.com/elevation-tiles-prod/terrarium/{z}/{x}/{y}.png"
HEADS_ONLY = False
UA = "mahanathi-geodata-research/0.1 (river gauge relief map; research fetch)"


def lon2x(lon, z):
    return (lon + 180.0) / 360.0 * (1 << z)


def lat2y(lat, z):
    r = math.radians(lat)
    return (1.0 - math.asinh(math.tan(r)) / math.pi) / 2.0 * (1 << z)


def tile_range(bbox, z):
    w, s, e, n = bbox
    x0, x1 = int(math.floor(lon2x(w, z))), int(math.floor(lon2x(e, z) - 1e-9))
    y0, y1 = int(math.floor(lat2y(n, z))), int(math.floor(lat2y(s, z) - 1e-9))
    return x0, x1, y0, y1


def head(job):
    z, x, y, out = job
    req = urllib.request.Request(URL.format(z=z, x=x, y=y), method="HEAD",
                                 headers={"User-Agent": UA})
    with urllib.request.urlopen(req, timeout=60) as r:
        return (x, y, r.status, int(r.headers.get("Content-Length", 0)),
                r.headers.get("x-amz-meta-x-imagery-sources"))


def fetch(job):
    z, x, y, out = job
    path = os.path.join(out, str(z), str(x), "%d.png" % y)
    if HEADS_ONLY:
        return head(job)
    if os.path.exists(path) and os.path.getsize(path) > 0:
        return (x, y, "cached", os.path.getsize(path), None)
    os.makedirs(os.path.dirname(path), exist_ok=True)
    req = urllib.request.Request(URL.format(z=z, x=x, y=y), headers={"User-Agent": UA})
    for attempt in range(4):
        try:
            with urllib.request.urlopen(req, timeout=60) as r:
                data = r.read()
                # S3 serves the joerd source list as user metadata.
                src = r.headers.get("x-amz-meta-x-imagery-sources")
            with open(path + ".part", "wb") as f:
                f.write(data)
            os.replace(path + ".part", path)
            return (x, y, 200, len(data), src)
        except Exception as ex:  # noqa: BLE001 - research script, report and retry
            err = repr(ex)
            time.sleep(2 ** attempt)
    return (x, y, "error", 0, err)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--z", type=int, required=True)
    ap.add_argument("--bbox", type=float, nargs=4, required=True, metavar=("W", "S", "E", "N"))
    ap.add_argument("--out", required=True)
    ap.add_argument("--dry-run", action="store_true")
    ap.add_argument("--heads", action="store_true",
                    help="HEAD every tile and record sources, no download")
    a = ap.parse_args()
    global HEADS_ONLY
    HEADS_ONLY = a.heads
    x0, x1, y0, y1 = tile_range(a.bbox, a.z)
    n = (x1 - x0 + 1) * (y1 - y0 + 1)
    print("z=%d x=%d..%d y=%d..%d tiles=%d" % (a.z, x0, x1, y0, y1, n))
    if a.dry_run:
        return
    jobs = [(a.z, x, y, a.out) for x in range(x0, x1 + 1) for y in range(y0, y1 + 1)]
    results = []
    with ThreadPoolExecutor(max_workers=4) as ex:  # never more than 4
        for res in ex.map(fetch, jobs):
            results.append(res)
    hdr_path = os.path.join(a.out, str(a.z), "headers.json")
    old = {}
    if os.path.exists(hdr_path):
        with open(hdr_path) as f:
            old = json.load(f)
    for x, y, status, size, src in results:
        key = "%d/%d" % (x, y)
        if status == "cached" and key in old:
            continue
        old[key] = {"status": status, "bytes": size, "sources": src}
    with open(hdr_path, "w") as f:
        json.dump(old, f, indent=1, sort_keys=True)
    total = sum(r[3] for r in results)
    errs = [r for r in results if r[2] == "error"]
    print("bytes=%d errors=%d" % (total, len(errs)))
    for e in errs:
        print("ERR", e, file=sys.stderr)


if __name__ == "__main__":
    main()
