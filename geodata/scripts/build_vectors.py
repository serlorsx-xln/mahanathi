"""Build the shipped vector assets from the cached sources.

Deterministic, one command, reading only from the gitignored
geodata/.cache/ and writing the three vector assets into
geodata/assets/ (see geodata/SOURCES.md for the sources and licences):

- provinces.topo.json: 77 Thai province polygons, Natural Earth 10m
  admin-1, simplified at a 500 m interval (mapshaper keep-shapes),
  TopoJSON with quantization 100000, properties trimmed to the ISO
  code and English name;
- borders.topo.json: international land boundary lines clipped to the
  bbox, same simplification and quantization, ADM0 codes both sides;
- rivers.geojson: the 13 OSM river relations (centre-line members,
  clipped, Douglas-Peucker 0.005 deg, 4-decimal coordinates), produced
  by the research script geodata/scripts/research/osm_rivers.py.

The two TopoJSONs need mapshaper 0.7.101 (installed into the gitignored
.cache/node-scratch/); the rivers need the research venv this script is
run with.

Usage: .cache/venv/bin/python -I geodata/scripts/build_vectors.py
"""
import os
import shutil
import subprocess
import sys

BBOX = "92,3,111,24"
SIMPLIFY_M = 500
MS = ".cache/node-scratch/node_modules/.bin/mapshaper"
NE = "geodata/.cache/naturalearth"
OSM_GEOM = "geodata/.cache/osm/rivers_geom.json"
RESEARCH = os.path.join(os.path.dirname(__file__), "research")


def run(cmd):
    print("$", " ".join(cmd))
    subprocess.run(cmd, check=True, capture_output=True)


def build_provinces(out):
    run([MS, "-i", NE + "/ne_10m_admin_1_states_provinces/ne_10m_admin_1_states_provinces.shp",
         "-filter", 'adm0_a3 == "THA"', "-filter-fields", "iso_3166_2,name_en",
         "-simplify", "interval=%d" % SIMPLIFY_M, "keep-shapes",
         "-o", out, "format=topojson", "quantization=100000", "force"])


def build_borders(out):
    run([MS, "-i", NE + "/ne_10m_admin_0_boundary_lines_land/ne_10m_admin_0_boundary_lines_land.shp",
         "-clip", "bbox=" + BBOX, "-filter-fields", "ADM0_A3_L,ADM0_A3_R",
         "-simplify", "interval=%d" % SIMPLIFY_M,
         "-o", out, "format=topojson", "quantization=100000", "force"])


def build_rivers(out):
    # The research script writes one GeoJSON per tolerance; we ship the
    # 0.005 deg one (about 0.55 km, half a DEM pixel).
    tmp = "geodata/.cache/rivers_build"
    shutil.rmtree(tmp, ignore_errors=True)
    subprocess.run([sys.executable, "-I", os.path.join(RESEARCH, "osm_rivers.py"),
                    OSM_GEOM, "92", "3", "111", "24", "--out", tmp],
                   check=True)
    shutil.copyfile(os.path.join(tmp, "osm13_dp005.geojson"), out)
    shutil.rmtree(tmp, ignore_errors=True)


def main():
    os.makedirs("geodata/assets", exist_ok=True)
    for name, fn in (("provinces.topo.json", build_provinces),
                     ("borders.topo.json", build_borders),
                     ("rivers.geojson", build_rivers)):
        out = os.path.join("geodata/assets", name)
        fn(out)
        print("wrote %s (%d bytes)" % (out, os.path.getsize(out)))


if __name__ == "__main__":
    main()
