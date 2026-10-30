# Geodata sources and licences

Every shipped asset under `geodata/assets/` is produced by the scripts in
`geodata/scripts/` from the sources below, pinned and re-fetchable. The
research behind each choice (measurements, candidate encodings, licence
texts) is `docs/research/geodata-sources.md`; this file is the shipping
record.

## Assets

| Asset | What it is | Size |
|---|---|---|
| `assets/heightmap.png` | 1815 × 2048 px (width × height), 16-bit grayscale, sqrt-encoded 11-bit heights, sea = code 0, Thailand full precision with a 75 km fade band and neighbours smoothed (r = 2) | 1,269,496 B |
| `assets/provinces.topo.json` | 77 Thai province polygons, Natural Earth 10m admin-1 simplified to 500 m, TopoJSON | 85,557 B |
| `assets/borders.topo.json` | International land boundary lines, Natural Earth 10m, 500 m, TopoJSON | 43,406 B |
| `assets/rivers.geojson` | 13 named rivers (Chao Phraya, Mekong, …), OSM relations simplified to 0.005° | 59,076 B |

Total: about 1.46 MB, within the brief's 1.5 MB budget.

## Sources and licences

### Elevation — AWS Terrain Tiles (Terrarium PNG, zoom 8)

- `https://s3.amazonaws.com/elevation-tiles-prod/terrarium/{z}/{x}/{y}.png`
- Composition over Thailand: **SRTM** and **GMTED2010** on land (U.S.
  Geological Survey), **ETOPO1** at sea (NOAA). U.S. Government public
  domain data; attribution requested.
- Credit line carried on the about page:

  > Elevation data: SRTM, GMTED2010 and ETOPO1 via Mapzen Terrain Tiles
  > (AWS), U.S. Geological Survey and National Oceanic and Atmospheric
  > Administration.

### Borders and provinces — Natural Earth 10m

- `ne_10m_admin_1_states_provinces` v5.1.1 (provinces),
  `ne_10m_admin_0_boundary_lines_land` v5.1.0 (international lines),
  `ne_10m_land` (the sea mask), `ne_10m_admin_0_countries` (the priority
  mask).
- **Public domain.** "Made with Natural Earth" is optional and is carried
  on the about page.
- 9 of 77 Thai `name_local` values are wrong in the source; the Thai names
  are the source's Latin names transliterated where the about page and
  search need them, and the curated table from the research doc
  (Wikidata labels, CC0) is the authority for the map labels.

### Rivers — OpenStreetMap

- 13 river relations fetched via Overpass on 2026-10-29, chosen by Thai
  name and checked against Wikidata IDs (all 13 match).
- **ODbL 1.0.** Credit line carried on the about page:

  > แม่น้ำ: © ผู้ร่วมสร้าง OpenStreetMap, ODbL
  > (openstreetmap.org/copyright)

  The shipped file carries the OSM relation IDs, satisfying the ODbL
  attribution requirement for a produced work.

## Rebuilding

```sh
# from the repository root
.cache/venv/bin/python -I geodata/scripts/build_heightmap.py     # → assets/heightmap.png
.cache/venv/bin/python -I geodata/scripts/build_vectors.py       # → provinces, borders, rivers
```

Both scripts read from and write to the gitignored `geodata/.cache/` and
write the shipped assets into `geodata/assets/`. Fetching the 224
Terrarium tiles (concurrency 4) happens once, on the first run.

`build_heightmap.py` is byte-reproducible: the committed
`assets/heightmap.png` is identical to the research run's candidate for
the shipped configuration (`heightmap_candidates_band75_b2/`).

## Local (gitignored) assets

These are rebuilt by recorded commands, not shipped in git (size or
regenerability puts them outside the commit budget):

- `assets/dem/` — terrarium-encoded DEM tiles z4–8 (462 tiles, ~127 MB),
  built from the committed `assets/heightmap.png` by
  `geodata/scripts/build_terrarium.py` (byte-reproducible from the
  committed heightmap).
- `assets/thailand.pmtiles` — the dark basemap (160 MB, maxzoom 13),
  built with planetiler (`protomaps-basemap` HEAD, `--area=thailand
  --maxzoom=13`) from the Geofabrik `thailand-latest.osm.pbf` extract.
  **Licence: OpenStreetMap data is ODbL**; the basemap is a derivative
  database of OSM and inherits share-alike for the database. The page
  credits "© ผู้ร่วมแก้ไข OpenStreetMap · Protomaps" (protomaps-basemap
  is BSD-3). Thai basemap labels are deferred: the build carries
  `name:th` in the vector tiles, but `fonts.protomaps.com` does not
  resolve, so no glyph set is served; a self-hosted Thai glyph PBF set
  is the follow-up.
