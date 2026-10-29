#!/usr/bin/env bash
# Measure compact border and admin-1 outputs built with mapshaper.
# Research script, not the production pipeline.
#
# Usage:
#   geodata/scripts/research/borders_sizes.sh <mapshaper-bin> <naturalearth_dir> <out_dir>
# Example:
#   geodata/scripts/research/borders_sizes.sh .cache/node-scratch/node_modules/.bin/mapshaper \
#       geodata/.cache/naturalearth geodata/.cache/borders_candidates
set -euo pipefail
MS="$1"; NE="$2"; OUT="$3"
BBOX="92,3,111,24"
mkdir -p "$OUT"

size() {  # raw, gzip -9 and brotli q11 sizes of a file, via node's zlib
  node -e '
    const fs=require("fs"),z=require("zlib");const b=fs.readFileSync(process.argv[1]);
    const br=z.brotliCompressSync(b,{params:{[z.constants.BROTLI_PARAM_QUALITY]:11}});
    console.log(String(b.length).padStart(9), String(z.gzipSync(b,{level:9}).length).padStart(8), String(br.length).padStart(8), " ", require("path").basename(process.argv[1]));
  ' "$1"
}

echo "     raw     gzip   brotli   file"
for iv in 250 500 1000; do
  # Thai provinces as polygons (77), shared arcs, ISO code and English name only.
  "$MS" -i "$NE/ne_10m_admin_1_states_provinces/ne_10m_admin_1_states_provinces.shp" \
    -filter 'adm0_a3 == "THA"' -filter-fields iso_3166_2,name_en \
    -simplify interval=$iv keep-shapes \
    -o "$OUT/th_provinces_${iv}m.topojson" format=topojson quantization=100000 force 2>/dev/null
  size "$OUT/th_provinces_${iv}m.topojson"

  # International land boundaries inside the bbox (Thai border plus neighbours).
  "$MS" -i "$NE/ne_10m_admin_0_boundary_lines_land/ne_10m_admin_0_boundary_lines_land.shp" \
    -clip bbox=$BBOX -filter-fields ADM0_A3_L,ADM0_A3_R \
    -simplify interval=$iv \
    -o "$OUT/land_boundaries_${iv}m.topojson" format=topojson quantization=100000 force 2>/dev/null
  size "$OUT/land_boundaries_${iv}m.topojson"

  # Country polygons clipped to the bbox (Thailand mask, neighbour tint).
  "$MS" -i "$NE/ne_10m_admin_0_countries/ne_10m_admin_0_countries.shp" \
    -clip bbox=$BBOX -filter-fields ADM0_A3 \
    -simplify interval=$iv keep-shapes \
    -o "$OUT/countries_${iv}m.topojson" format=topojson quantization=100000 force 2>/dev/null
  size "$OUT/countries_${iv}m.topojson"

  # Same province polygons as GeoJSON (no shared arcs), for comparison.
  "$MS" -i "$NE/ne_10m_admin_1_states_provinces/ne_10m_admin_1_states_provinces.shp" \
    -filter 'adm0_a3 == "THA"' -filter-fields iso_3166_2,name_en \
    -simplify interval=$iv keep-shapes \
    -o "$OUT/th_provinces_${iv}m.geojson" format=geojson precision=0.0001 force 2>/dev/null
  size "$OUT/th_provinces_${iv}m.geojson"
done

# Vertex counts before and after simplification, for the record.
for iv in 0 250 500 1000; do
  if [ "$iv" = 0 ]; then simp=""; else simp="-simplify interval=$iv keep-shapes"; fi
  printf "provinces interval=%sm: " "$iv"
  "$MS" -i "$NE/ne_10m_admin_1_states_provinces/ne_10m_admin_1_states_provinces.shp" \
    -filter 'adm0_a3 == "THA"' $simp -info 2>&1 | grep -iE "^Records|Vertices|vertices" | tr '\n' ' ' || true
  echo
done
