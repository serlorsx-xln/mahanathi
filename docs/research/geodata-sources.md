# Geodata sources: elevation, borders, provinces and rivers

Researched 2026-10-29 by the implementing agent (an AI model), as decision support for the geodata pipeline (`geodata/`, built in Phase 3). This document chooses sources and an encoding and shows the measurements behind each choice. **It does not build the production pipeline.** The proposed decisions are listed at the end for the lead.

Every number comes from a script in `geodata/scripts/research/` run on files downloaded on 2026-10-29, or from a calculation that is written out. Anything not checked is marked **unverified**. Downloads and candidate outputs are under `geodata/.cache/` (gitignored). The commands are listed in section 6.

## Summary

- **Elevation:** AWS Terrain Tiles, Terrarium PNG, zoom 8. That is 224 tiles and 22.4 MB, fetched with 4 workers. The tiles come from SRTM and GMTED2010 (USGS) on land and ETOPO1 (NOAA) at sea. All three are U.S. Government data with attribution requested.
- **Grid:** a regular lon/lat (equirectangular) grid.
  - Extent 92 to 111°E and 3 to 24°N, with pixels made square at 13°N.
  - **2048 × 1815 px, about 1.13 km per pixel.**
  - Grid sizes in this document are written **rows × columns**. As an image, the 2048 × 1815 grid is 1815 px wide and 2048 px tall, and 1536 × 1361 is 1361 wide and 1536 tall. (Implementer's review note: checked with Pillow on the cached candidates.)
  - Stations map into it linearly from lat/lon. No Web Mercator reaches the client.
- **Heightmap encoding:** heights on a square-root curve with 11 bits. The sea is a reserved code 0 and is drawn flat and dark, with no bathymetry.
  - Thailand plus a fade band keeps full precision. The neighbours are smoothed, so they cost fewer bytes.
  - Stored as a **16-bit grayscale PNG of 1.17 to 1.45 MB**, depending on how much the neighbours are smoothed (measured).
  - In the central plain the error is **0.045 m RMSE**, against the DEM's own pixel-to-pixel noise of 1.31 m.
  - The browser decodes it in a worker with `DecompressionStream('deflate')`, with no canvas involved.
- **Borders and provinces:** Natural Earth 10m (public domain), files v5.1.1 and v5.1.0.
  - 77 Thai provinces, including Bueng Kan (TH-38) and Bangkok (TH-10).
  - There is no Thai point-of-view edition and no `name_th` field.
  - 9 of the 77 Thai `name_local` values are wrong or inconsistent, so the Thai names come from a curated table checked against Wikidata.
  - TopoJSON at a 500 m simplification: provinces 26.6 KB and international lines 13.8 KB, both brotli-compressed.
- **Rivers:** OpenStreetMap (ODbL) for all 13 named rivers. The relations were found by their Thai names, and all 13 carry Wikidata IDs whose Thai labels match the OSM Thai names exactly. The simplified file is 11.9 KB brotli.
  - Natural Earth has only 8 of the 13, and labels two of those wrongly.
  - HydroRIVERS is positionally good but has no names. Its WWF licence restricts redistribution, which is a risk for a public web app.
- **Total initial geodata transfer** is about 1.32 MB (section 6), almost all of it the heightmap.

## 1. Elevation (DEM)

### 1.1 Source, URL scheme and encoding

- **URL:** `https://s3.amazonaws.com/elevation-tiles-prod/terrarium/{z}/{x}/{y}.png`. These are Web Mercator (EPSG:3857) tiles of 256 × 256 px with XYZ numbering, documented in `tilezen/joerd/docs/formats.md`.
- **Encoding**, from the same document: "All values are positive with a 32,768 offset, split into the red, green, and blue channels, with 16 bits of integer and 8 bits of fraction." To decode:

  ```
  height_m = (R * 256 + G + B / 256) - 32768
  ```

  `heightmap_measure.py` decodes every tile with this formula. The stitched results are plausible:
  - land up to 3,297 m and sea down to -4,709 m at 1.13 km/px;
  - the lower Chao Phraya plain at 3.05 to 15.78 m (1st to 99th percentile);
  - a hillshade preview with the expected coastlines (`geodata/.cache/heightmap_candidates/preview_full_full.png`).
- **Sources per zoom** (`joerd/docs/data-sources.md`): ETOPO1 covers the ocean at every zoom. On land the source is ETOPO1 at z0 to 3, GMTED at z4 to 6, and SRTM from z7.
- **Per-tile headers:** each tile names its sources in the S3 header `x-amz-meta-x-imagery-sources`. The docs call it `X-Imagery-Sources`, but that header is not returned.
  - All 64 z7 tiles were checked: 14 are SRTM plus GMTED, 43 are SRTM plus GMTED plus ETOPO1, 6 are GMTED plus ETOPO1, and 1 is ETOPO1 only.
  - A z8 sample, tile 8/200/118, lists `srtm/N13E101.tif … gmted/10N090E_… etopo1/ETOPO1_Bed_g.tif`.
  - So the attribution covers exactly three datasets: SRTM, GMTED2010 and ETOPO1.
- **Age:** the tiles carry `Last-Modified: 2017-11-19`.
  - An undocumented `v2/` prefix serves a near-identical tile: 69,100 bytes against 69,115, dated 2017-09-01. It is not used.
  - A registry YAML is fetched with the tiles. That copy (cached at `geodata/.cache/docs/joerd_terrain-tiles.yaml`) is identical to the canonical file in awslabs/open-data-registry.

### 1.2 Licence and attribution

The AWS registry entry gives the licence as a link to `tilezen/joerd/docs/attribution.md` and says the dataset is "Managed By Mapzen, a Linux Foundation project". That page lists required attribution per source. The lines that apply to tiles over Thailand are:

> * Global ETOPO1 terrain data U.S. National Oceanic and Atmospheric Administration
> * United States 3DEP (formerly NED) and global GMTED2010 and SRTM terrain data courtesy of the U.S. Geological Survey.

The fine print in the same page adds two points:
- SRTM, GMTED2010 and ETOPO1 are public domain U.S. Government works, and USGS asks for credit.
- ETOPO1 cites 17 U.S.C. 403: "third parties producing copyrighted works consisting predominantly of the material produced by U.S. government agencies must provide notice with such work(s) identifying the U.S. Government material incorporated and stating that such material is not subject to copyright protection within the United States."

Our heightmap consists predominantly of that material, so it needs the notice.

**Proposed credit line**, for the About page and `geodata/SOURCES.md`:

> Elevation: SRTM and GMTED2010 courtesy of the U.S. Geological Survey; ETOPO1 from the U.S. National Oceanic and Atmospheric Administration. These are U.S. Government data, not subject to copyright protection in the United States. Tiles: Terrain Tiles (Mapzen, a Linux Foundation project) on the Registry of Open Data on AWS.

The "Mapzen" line is only *required* for Mapzen's former hosted service. Crediting it here is a courtesy.

### 1.3 Bounding box and margins

Thailand's extent in Natural Earth admin-0 is 97.351 to 105.651°E and 5.630 to 20.445°N. The proposed bbox is **92 to 111°E, 3 to 24°N**, centred on 101.5°E and 13.5°N.

The km figures below use 108.48 km per degree of longitude and 110.63 km per degree of latitude (WGS84 at 13°N, section 1.4).

| Side | Margin | Where the edge falls | Why |
|---|---|---|---|
| West, 92°E | 5.35°, about 580 km | Bay of Bengal, west of Myanmar's Rakhine coast | the sea frames the west side, so there is no cut through land |
| East, 111°E | 5.35°, about 580 km | South China Sea, east of Vietnam's coast | includes the whole lower Mekong (Laos, Cambodia, the delta), which the Mekong, Mun and Chi belong to |
| South, 3°N | 2.63°, about 290 km | Malay peninsula and the Strait of Malacca | the peninsula runs on past the frame instead of ending at the Thai border |
| North, 24°N | 3.55°, about 390 km | land: Myanmar, Yunnan, northern Laos and Vietnam | the only land edge; about 390 km of terrain beyond Chiang Rai lets fog and aerial perspective fade it out (assumes a default camera looking roughly north, **unverified** until look-dev) |

- **Thailand is 10.9% of the pixels** (`tha_px_share` = 0.1092). The wide frame costs little because the neighbours are encoded cheaply (section 1.8).
- **Tighter boxes were not measured.**

### 1.4 Projection and how stations map

**Proposal: plate carrée (equirectangular) with square pixels at 13°N, done once in the pipeline.** The client never sees Web Mercator.

- **The grid** is regular in lon/lat.
  - With 2048 rows, 21° / 2048 = 0.0102539° per row and 19° / 1815 = 0.0104683° per column.
  - At 13°N that is 1134.4 m by 1135.7 m per pixel, square to within 0.1%.
- **WGS84 metres per degree at 13°N** are calculated as follows:
  - Meridian: 111132.954 - 559.822 cos 2φ + 1.175 cos 4φ = 110,630.5 m.
  - Parallel: 111412.84 cos φ - 93.5 cos 3φ + 0.118 cos 5φ = 108,484.7 m.
  - The research scripts used a linear approximation of 110.736 km per degree of latitude, 0.10% high. That moves the column count by 2 (1815 against 1817) and does not change any size or error figure in this document.
- **Reprojection** from the Mercator mosaic is separable: columns are linear in longitude in both projections, so only the row positions change. `reproject()` in `heightmap_measure.py` samples bilinearly at native z8 resolution and then box-filters down to the target size.
- **Distortion over Thailand**, relative to true ground distance:

  | Latitude | Equirectangular, x scaled by cos 13°: east-west | north-south | Web Mercator rescaled to be true at 13°: both axes |
  |---|---|---|---|
  | 5.63°N, the southern tip | -2.1% | 0 | -2.1% |
  | 13°N | 0 | 0 | 0 |
  | 20.45°N, the northern tip | +4.0% | 0 | +4.0% |

  - The numbers are cos(13°) / cos(lat) - 1. The equirectangular grid only stretches east-west, by at most 4%.
  - Mercator scales both axes, so its area error is roughly twice as large (+8.2% and -4.2%). It is conformal, but a 4% shape error is not visible at country scale.
  - Equirectangular wins on simplicity, because every consumer uses one linear formula.
- **Station mapping**, assuming ThaiWater lat/lon are WGS84 (**unverified**):

  ```
  u = (lon - 92) / 19            v = (24 - lat) / 21          // texture coords, v down
  X_km = (lon - 101.5) * 108.4847
  Z_km = (lat - 13.5)  * -110.6305                             // sign follows the scene's north
  Y    = height(u, v) * exaggeration                           // bilinear on the CPU copy
  ```

  The terrain mesh is a plain rectangle of 2061 × 2323 km (19° × 108.48 by 21° × 110.63). Rivers and borders go through the same function. One pure function shared by the client and the tests is enough.

### 1.5 Zoom, tile count and download size

| Zoom | Tile range (x by y) | Tiles | Bytes | Native resolution at 13°N | Use |
|---|---|---|---|---|---|
| 7 | 96..103 by 55..62 | 64 | 4,283,863 (measured) | 1191.7 m | headers and a first check only; coarser than the 2048-row output |
| **8** | **193..206 by 110..125** | **224** | **22,449,608 (measured)** | **595.8 m** | **proposed source** |
| 9 | 386..413 by 220..251 | 896 | not downloaded | 297.9 m | |
| 10 | 773..827 by 441..503 | 3,465 | not downloaded | 149.0 m | |

- **Native resolution** is 156543.03 × cos φ / 2^z metres per pixel.
- **Tile counts** for z9 and z10 come from `fetch_terrarium.py --dry-run`.
- **Why z8:** at 2048 rows each output pixel averages about 1.9 × 1.9 source pixels, which also damps SRTM speckle. Only an output of about 4000 rows or more would need z9, and section 1.7 shows that cannot fit in 1.5 MB.
- **Politeness:** the download used `ThreadPoolExecutor(max_workers=4)` with a descriptive User-Agent, skipped tiles already cached and retried with backoff.

### 1.6 Sea, coast and bathymetry

The design asks for "dark and calm" sea, so **bathymetry is dropped**. The sea becomes a reserved code 0 and the client draws it as one flat dark plane.

Measured costs and facts, all at 2048 rows:

- **Bathymetry is most of the raw file.** Raw Terrarium RGB with bathymetry and the 1/256 m fraction is 7,809,452 bytes. With the sea flattened and heights rounded to whole metres, the same container is 2,285,121 bytes, 3.4× smaller.
- **"Height ≤ 0" is not a sea mask.**
  - 50.61% of pixels have DEM ≤ 0, but only 48.94% are sea by Natural Earth land coverage.
  - 18,624 pixels outside NE land have DEM above 1 m: islands that NE 10m omits, plus coastline offsets.
  - 4,855 NE-land pixels have DEM ≤ 0, as low as -644 m, where NE's coast overlaps ETOPO1 sea.
  - In a Bangkok and Samut Prakan box (100.30 to 100.90°E, 13.55 to 14.00°N), only 2 of 2,649 land pixels are at or below 0 m.
- **The proposed mask is land = (NE 10m land covers 50% of the pixel or more) OR (DEM above 1 m).** Land heights are clamped to 0 m or more. After this, no sea pixel is above 5 m (measured: 0).
- **Coastline smoothness:** codes 0 and 1 make a hard edge. If look-dev wants an antialiased coast, ship a separate 8-bit coverage mask. A Thailand-only mask at 2048 rows measured 23,334 bytes as PNG and 12,188 as WebP. A full land mask has a longer outline, so its size is **unverified**.

### 1.7 Encodings measured

All encodings below use the same 2048 × 1815 grid and the same land mask. "Plain" is a test box in the lower Chao Phraya plain: 100.15 to 100.60°E, 14.0 to 15.0°N (Ayutthaya, Ang Thong, Pathum Thani and southern Sing Buri), with 4,356 land pixels. It was placed west of the Lop Buri hills because a wider first box reached a 95th percentile of 75 m.

Errors are measured against the float reference, with the sea flattened. PNG sizes are the smaller of Pillow `optimize` and `oxipng` level 6. WebP is lossless, method 6, `exact`.

| Encoding | Bytes | Plain RMSE (m) | Plain max error (m) | Distinct levels in plain | All-land RMSE (m) |
|---|---|---|---|---|---|
| Terrarium RGB PNG, raw (bathymetry, 1/256 m) | 7,809,452 | 0.002 | 0.004 | 2,160 | 0.002 |
| Terrarium RGB PNG, sea 0, whole metres | 2,285,121 | 0.289 | 0.500 | 20 | 0.288 |
| Terrarium RGB WebP lossless, sea 0, whole metres | 1,677,956 | 0.289 | 0.500 | 20 | 0.288 |
| 16-bit gray PNG, linear 1 m | 2,131,988 | 0.289 | 0.500 | 20 | 0.288 |
| 16-bit gray PNG, linear 0.5 m | 2,363,258 | 0.145 | 0.250 | 37 | 0.144 |
| 16-bit gray PNG, linear 0.1 m | 2,845,307 | 0.029 | 0.050 | 157 | 0.029 |
| RG-packed WebP lossless, linear 1 m | 1,678,124 | 0.289 | 0.500 | 20 | 0.288 |
| RG-packed WebP lossless, linear 0.5 m | 1,905,252 | 0.145 | 0.250 | 37 | 0.144 |
| uint16 row-delta + brotli 11, linear 1 m | 1,836,963 | 0.289 | 0.500 | 20 | 0.288 |
| float16 + zstd 19 (proxy for KTX2 R16F + zstd) | 3,097,899 | 0.002 | 0.007 | 1,923 | 0.134 |
| 8-bit PNG, linear over 0 to 3,300 m | 823,857 | 4.393 | 6.494 | 4 | 3.768 |
| 8-bit PNG, sqrt curve | 996,662 | 0.365 | 1.077 | 17 | 2.749 |
| 8-bit PNG, log curve | 1,004,194 | 0.085 | 0.394 | 60 | 5.963 |
| 16-bit gray PNG, sqrt curve, 10 bits | 1,926,595 | 0.091 | 0.238 | 55 | 0.682 |
| RG WebP, sqrt 10 bits | 1,500,000 | 0.091 | 0.238 | 55 | 0.682 |
| 16-bit gray PNG, sqrt 11 bits | 2,250,938 | 0.045 | 0.113 | 101 | 0.341 |
| RG WebP, sqrt 11 bits | 1,774,744 | 0.045 | 0.113 | 101 | 0.341 |
| MED residual + brotli 11, sqrt 11 bits | 1,874,969 | 0.045 | 0.113 | 101 | 0.341 |
| 16-bit gray PNG, sqrt 12 bits | 2,477,550 | 0.022 | 0.076 | 192 | 0.170 |

Notes on the table:
- **Sqrt curve:** code = 1 + round(sqrt(h / 3300) × (2^N - 2)), and 0 is the sea. The range of 3,300 m is the land maximum, 3,296.9 m, rounded up.
- **MED:** the JPEG-LS median edge predictor (left, up and up-left neighbours). Its residuals decode losslessly in one sequential pass, which was checked on a 200 × 200 crop.
- **The 1,500,000-byte WebP** is a real measurement, not a rounded one, and it was reproduced in a second run.
- **8-bit WebP sizes:** sqrt 923,822, log 923,734 and linear 763,436 bytes.
- **Other compressors** on uint16 row-delta at 1 m: zstd 19 gives 2,046,917, gzip 9 gives 2,316,614 and xz gives 1,889,480.

Step size (metres per code) at a given height, from the same report:

| Height | 8-bit linear | 8-bit sqrt | 8-bit log | sqrt 10-bit | **sqrt 11-bit** | sqrt 12-bit |
|---|---|---|---|---|---|---|
| 5 m | 12.99 | 1.06 | 0.19 | 0.26 | **0.13** | 0.06 |
| 15 m | 12.99 | 1.80 | 0.52 | 0.44 | **0.22** | 0.11 |
| 30 m | 12.99 | 2.53 | 1.01 | 0.62 | **0.31** | 0.15 |
| 1,000 m | 12.99 | 14.36 | 32.44 | 3.56 | **1.78** | 0.89 |
| 2,500 m | 12.99 | 22.67 | 81.06 | 5.62 | **2.81** | 1.40 |

Size by output resolution, in bytes. "n/m" means not measured.

| Rows × cols | m/px | Terrarium raw | 16-bit 1 m PNG | 16-bit 0.5 m PNG | RG WebP 0.5 m | sqrt-11 PNG | sqrt-11 WebP | 8-bit sqrt PNG |
|---|---|---|---|---|---|---|---|---|
| 1024 × 907 | 2,271 | 2,032,720 | 568,934 | 620,587 | 509,200 | n/m | n/m | 279,808 |
| 1536 × 1361 | 1,514 | 4,473,011 | 1,232,633 | 1,360,220 | 1,103,400 | 1,300,913 | 1,026,286 | 589,372 |
| 1792 × 1588 | 1,298 | n/m | n/m | n/m | n/m | 1,747,893 | 1,377,152 | n/m |
| 2048 × 1815 | 1,136 | 7,809,452 | 2,131,988 | 2,363,258 | 1,905,252 | 2,250,938 | 1,774,744 | 996,662 |
| 2560 × 2269 | 908 | 12,063,026 | 3,277,113 | 3,626,428 | 2,907,282 | n/m | n/m | 1,489,711 |

What the tables show:

1. **Precision in the plain is not the binding constraint.**
   - The plain spans 3.05 to 15.78 m (1st to 99th percentile).
   - The DEM itself varies between horizontal neighbours with a standard deviation of 1.31 m (median absolute difference 0.57 m).
   - Every 16-bit option quantises more than 10× below that noise.
   - At a 10× exaggeration that noise becomes 13 m bumps per 1.1 km pixel. That is a look-dev question (it could mean smoothing below about 30 m), not an encoding one.
2. **8-bit fails the mountains or the plain.**
   - Linear 8-bit leaves 4 levels in the plain.
   - Sqrt 8-bit has 22.7 m steps at 2,500 m, which become visible 227 m terraces in hillshade at 10×.
   - Log 8-bit has 81 m steps at 2,500 m.
3. **KTX2 does not fit at this resolution.** KTX2 supercompression (zstd or zlib) has no spatial predictor. The float16 + zstd proxy is 3.10 MB at 2048 rows, about twice the budget. KTX2 with Basis is lossy, so it is unsuitable for heights.
4. **Mountain detail sets the file size, not plain precision.** Going from 1 m to 0.1 m linear adds 33%, and sqrt 10-bit to 12-bit adds 29%. No uniform 16-bit encoding at 2048 rows fits 1.5 MB as PNG. Uniform 16-bit PNGs under 1.5 MB need 1536 rows (1.51 km/px), for example linear 0.5 m at 1,360,220 bytes or sqrt-12 at 1,420,082.

### 1.8 Thailand-priority encoding

The design brief says neighbours are "desaturated and recede through aerial perspective". That suggests spending bits on Thailand:
- **Thailand:** full sqrt 11-bit precision inside Thailand, using the NE admin-0 polygon.
- **Fade band:** a linear blend over a band outside the border.
- **Neighbours:** smoothed with three passes of a box blur, then quantised to every 4th code, which is effectively 9 bits.

The smoothing strength is the radius r: three passes of a (2r+1)-wide box give σ = sqrt(3 × ((2r+1)² - 1) / 12) pixels.
- r = 1 gives σ = 1.41 px, about 1.6 km.
- r = 2 gives σ = 2.45 px, about 2.8 km.

All rows below use 2048 × 1815 and sqrt 11 bits. Thai land RMSE is 0.274 m in every row, because the quantiser inside Thailand is the same.

| Neighbours | Fade band | Neighbour land RMSE (m) | 16-bit PNG | RG WebP | MED + brotli 11 | MED + gzip 9 (planar) |
|---|---|---|---|---|---|---|
| uniform, no priority (reference) | none | 0.341 (all land) | 2,250,938 | 1,774,744 | 1,874,969 | 1,995,381 |
| every 4th code, no blur | 25 km | 1.57 | 1,882,860 | 1,439,680 | 1,531,051 | 1,742,991 |
| blur r = 1, every 4th code | 25 km | 68.73 | 1,371,567 | 1,070,142 | 1,060,656 | 1,347,963 |
| blur r = 1, every 4th code | 75 km | 69.12 | 1,454,568 | 1,121,266 | 1,127,166 | 1,404,145 |
| blur r = 2, every 4th code | 25 km | 98.56 | 1,174,213 | 911,224 | 925,956 | 1,203,769 |
| **blur r = 2, every 4th code** | **75 km** | 99.05 | **1,269,496** | 974,660 | 1,005,561 | 1,273,128 |

Other measured variants, all with r = 2 and a 25 km band:

| Grid | Precision | PNG | WebP | MED + brotli |
|---|---|---|---|---|
| 2048 rows | sqrt 12-bit | 1,282,216 | 1,042,092 | n/m |
| 2048 rows | sqrt 10-bit | 1,081,577 | 807,250 | 879,654 |
| 2560 rows | sqrt 11-bit | 1,772,782 | 1,368,938 | n/m |
| 2560 rows | sqrt 12-bit | n/m | 1,564,992 | n/m |
| 2560 rows | sqrt 10-bit | 1,629,375 | 1,214,428 | n/m |

Findings:
- **The savings come from smoothing, not from coarser codes.** Coarse codes alone save 16% of the PNG. Smoothing as well saves 39 to 48%.
- **The neighbour RMSE of 69 to 99 m is the smoothing itself:** ridges are softened, not shifted.
- **Container comparison:**
  - MED + brotli is within 3% of WebP for every priority variant, and about 21% smaller than PNG.
  - MED + gzip is no better than PNG. PNG is deflate with per-row predictors, so that is expected.
  - Interleaved little-endian residuals compress better with brotli than planar ones. For example, 925,956 against 1,086,269 bytes.
- **Visual check.** Hillshade previews at 10× exaggeration, looking straight down: `geodata/.cache/heightmap_candidates/preview_north_{full,blur1,blur2}.png`, cropped to 96 to 102°E and 17 to 22°N, and `preview_full_*.png`.
  - With a 25 km band, both radii show a clear step in sharpness at the Thai border.
  - Whether that reads as "neighbours recede" or as a seam is a look-dev call, made with the oblique camera, fog and desaturation in place.
  - The 75 km band softens the step for about 80 to 95 KB more. No preview was rendered for the 75 km band.

### 1.9 Recommendation and browser decode path

**Proposed heightmap:**
- 2048 × 1815, equirectangular, with the bbox from section 1.3.
- Sea = 0; land = 1 + round(sqrt(h / 3300) × 2046).
- Thailand priority with a 75 km fade band.
- Neighbours blurred with r = 2 as the starting point, and r = 1 as the look-dev alternative.
- Stored as a **16-bit grayscale PNG: 1,269,496 bytes with r = 2, or 1,454,568 bytes with r = 1**. Both are within 1.5 MB, which is read here as 1,500,000 bytes.

Why PNG rather than the smaller options:

- **Exact bits without canvas.** A small PNG reader (about 100 lines, no dependency) runs in a worker:
  - parse the chunks and join the IDAT data;
  - inflate it with `DecompressionStream('deflate')`, which per MDN browser-compat-data is in Chrome 80, Firefox 113 and Safari 16.4;
  - undo the five PNG filters at 2 bytes per pixel, giving a `Uint16Array` of codes.

  No image element, canvas or ImageBitmap is involved, so colour management and anti-fingerprinting readback noise cannot change a bit.
- **WebP is about 23% smaller but can only be decoded through the image pipeline.** That needs `createImageBitmap(..., {colorSpaceConversion: 'none', premultiplyAlpha: 'none'})` (Safari 15 and 17 respectively per MDN) and a canvas or WebGL readback for the CPU copy.
  - A changed high byte is an error of 256 codes.
  - Whether Safari's Advanced Fingerprinting Protection or Firefox's canvas randomisation touches these readbacks is **unverified**.
- **MED + brotli is about 21% smaller and deterministic.** It needs the host to serve the `.bin` with `Content-Encoding: br`, which is a deploy concern for the lead.
  - `DecompressionStream` cannot decode brotli in Chrome. MDN lists `brotli` as Firefox 147 and Safari 18.4 only.
  - The gzip fallback is no smaller than the PNG.
  - It stays the planned fallback if bytes ever matter more than simplicity.
- **The brief already names PNG,** and a 16-bit PNG opens in any image viewer for debugging.

On the client:
- **CPU copy:** keep the `Uint16Array` codes (7.4 MB) plus a 2,047-entry lookup table from code to metres. Height sampling for draping, gauge bases and the camera floor is then a table lookup and a bilinear blend.
- **GPU copy:** convert codes to half floats through a second lookup table and upload as an `R16F` texture (7.4 MB). Half floats are exact enough: the step is 1 m at 2,000 m and 1/128 m at 16 m.
  - That `R16F` is filterable in WebGL2 without extensions is **unverified** here, because the spec was not re-read. (Implementer's review note: the OpenGL ES 3.0 sized-format table that WebGL2 follows lists `R16F` as texture-filterable and `R32F` as not. That is from memory of the spec, not a browser test, so the Phase 3 probe should still confirm it.)
  - The alternative is `R16UI` with `texelFetch` and manual filtering.
- **Do not** load the 16-bit PNG through `<img>`, `Image` or `createImageBitmap`. Browsers are widely reported to reduce it to 8 bits, though this was **unverified** in this research.

Simpler fallback with no priority mask: 1536 × 1361 (1.51 km/px) sqrt 11-bit PNG, 1,300,913 bytes. It has a plain RMSE of 0.045 m and a land RMSE of 0.341 m.

## 2. Borders and provinces

### 2.1 Layers and versions

Downloaded from `https://naciscdn.org/naturalearth/10m/{cultural|physical}/<layer>.zip`. Versions come from each zip's `VERSION.txt`.

| Layer | Version | Use |
|---|---|---|
| `ne_10m_admin_0_countries` | 5.1.1 | Thailand polygon: the priority mask, neighbour desaturation and the national outline |
| `ne_10m_admin_0_boundary_lines_land` | 5.1.0 | international land boundaries, drawn |
| `ne_10m_admin_1_states_provinces` | 5.1.1 | the 77 Thai provinces |
| `ne_10m_admin_1_states_provinces_lines` | 5.1.0 | not needed: internal lines come from TopoJSON shared arcs |
| `ne_10m_land` | 5.1.1 | sea mask for the heightmap |
| `ne_10m_admin_0_disputed_areas`, `..._boundary_lines_disputed_areas`, `..._boundary_lines_maritime_indicator` | 5.1.1, 5.1.0, 5.1.0 | checked only (section 2.4) |

- **Latest release:** the GitHub API reports `v5.1.2`, published 2022-05-13.
  - The `admin_0_countries.zip` on naciscdn is dated the same day (`Last-Modified: 2022-05-13 05:20:49 GMT`), but its `VERSION.txt` says 5.1.1.
  - Whether v5.1.2 changed anything for Thailand is **unverified**.
  - `master` carries `VERSION` = `5.2.0-pre`. Pin the release that the pipeline uses.
- **Licence:** public domain. The terms page says:

  > All versions of Natural Earth raster + vector map data found on this website are in the public domain. … No permission is needed to use Natural Earth. Crediting the authors is unnecessary.

  It offers "Made with Natural Earth." as an optional credit. The page is https://www.naturalearthdata.com/about/terms-of-use/.

### 2.2 Point of view and Thai names

- **No Thailand point-of-view edition.**
  - Admin-0 v5.1.1 carries 33 `FCLASS_*` worldview fields: ISO, TLC and 31 countries (US, FR, RU, ES, CN, TW, IN, NP, PK, DE, GB, BR, IL, PS, SA, EG, MA, PT, AR, JP, KO, VN, TR, ID, PL, GR, IT, NL, SE, BD, UA).
  - None is for Thailand, so the default (de facto) lines are what we get.
  - All four Thai land boundaries (with Cambodia, Laos, Malaysia and Myanmar) have `FEATURECLA` "International boundary (verify)".
- **No Thai name fields.**
  - Admin-0 has `NAME_*` in 26 languages, and admin-1 and rivers in 23, but there is no `name_th` anywhere.
  - In the 77 Thai admin-1 rows, Thai script appears only in `name_local` and `name_alt`.
- **77 provinces,** including Bueng Kan (TH-38, created 2011) and Bangkok (TH-10). Wikidata has a Thai label for all 77 QIDs that NE links in `wikidataid`.
  - **68** have `name_local` equal to the Wikidata Thai label.
  - **3** differ only by a missing "จังหวัด" prefix: Chanthaburi (TH-22) is "จันทบุรี", Lop Buri (TH-16) is "ลพบุรี" and Roi Et (TH-45) is "ร้อยเอ็ด".
  - **6** are wrong:

| ISO | NE `name` | NE `name_local` | Wikidata Thai label |
|---|---|---|---|
| TH-10 | Bangkok Metropolis | จังหวัดเชียงใหม่ | กรุงเทพมหานคร |
| TH-38 | Bueng Kan | จังหวัดหนองคาย | จังหวัดบึงกาฬ |
| TH-36 | Chaiyaphum | จังหวัดชัยนาท | จังหวัดชัยภูมิ |
| TH-71 | Kanchanaburi | อำเภอเมืองกาญจนบุรี (a district) | จังหวัดกาญจนบุรี |
| TH-73 | Nakhon Pathom | อำเภอเมืองนครปฐม (a district) | จังหวัดนครปฐม |
| TH-67 | Phetchabun | อำเภอเมืองเพชรบูรณ์ (a district) | จังหวัดเพชรบูรณ์ |

  NE's English `name` also differs from `name_en` in places, for example "Bangkok Metropolis" against "Bangkok" and "Phangnga" against "Phang Nga".

**Proposal:** a curated `provinces.csv` with columns `iso_3166_2, wikidata, name_th, name_en`, keyed by ISO code.
- Seed it from the Wikidata labels, which are CC0 per Wikidata's copyright page.
- Store the short form without "จังหวัด". Copy rules decide when to show the prefix.
- Cross-check it against the Thai province names in ThaiWater's `geocode.province_name`, which search and station matching depend on. That check belongs to the data-quality work and was not done here.
- Never read `name_local` at runtime.

### 2.3 Output sizes

Built with mapshaper 0.7.101: `-simplify interval=<m> keep-shapes` (mapshaper's default method), TopoJSON `quantization=100000`, properties trimmed to the ISO code and English name. Script: `borders_sizes.sh`.

| File | Interval | Raw | gzip 9 | brotli 11 | Arcs / arc vertices |
|---|---|---|---|---|---|
| Thai provinces (77 polygons), TopoJSON | 250 m | 111,978 | 38,356 | 34,494 | 253 / 10,500 |
| | **500 m** | **85,557** | **29,586** | **26,580** | 252 / 7,440 |
| | 1000 m | 59,122 | 21,210 | 18,625 | 251 / 4,632 |
| Thai provinces, GeoJSON (no shared arcs), for comparison | 500 m | 227,573 | 65,629 | 39,293 | n/a |
| International land lines in the bbox, TopoJSON | 250 m | 48,601 | 18,217 | 15,435 | 472 / 4,386 |
| | **500 m** | **43,406** | **16,464** | **13,796** | 472 / 3,743 |
| | 1000 m | 34,217 | 13,353 | 10,926 | 472 / 2,688 |
| Country polygons clipped to the bbox, TopoJSON | 500 m | 110,672 | 38,231 | 34,758 | 158 / 11,365 |

- **Source vertices:** the Thai admin-1 source has 26,498 vertices, with shared edges stored twice.
- **Why 500 m:** it is about half a heightmap pixel, so lines stay within the terrain's own resolution.
- **The national outline is free:** it is the provinces' outer arcs (a TopoJSON mesh with `a === b`), so no separate country file is needed.
- **A Thailand raster mask** for desaturation, if the shader wants one rather than rasterising the outline at load, measured 23,334 bytes as PNG and 12,188 as WebP at 2048 rows.

### 2.4 Sensitive segments, especially Thailand and Cambodia

Facts measured on NE 10m (`ne_borders.py`):
- **No disputed polygons and no disputed boundary lines** intersect the bbox.
- **Eight maritime indicator lines** intersect it: 4 "Marine Indicator Median" and 4 "Marine Indicator Treaty".
- **Thailand-Cambodia land line:** 31 parts and 276 vertices, with a median segment length of 1.85 km, a 90th percentile of 6.31 km and a maximum of 18.27 km. It is generalised at kilometre scale.
- **Prasat Preah Vihear,** at its English Wikipedia coordinates (14.39056°N, 104.68028°E), falls inside NE's Cambodia polygon, 2.47 km from the line.
- **Prasat Ta Muen Thom** (14.3493°N, 103.2666°E) falls inside NE's Thailand polygon, 0.04 km from the line. That is on the line at NE's precision.
- **Ko Kut district** (Trat) falls inside Thailand, 39.8 km from the land line.
- **Recent context:** the 2025 Cambodian–Thai border crisis Wikipedia summary (revision of 2026-10-13) records armed confrontation from 24 July 2025, a ceasefire on 28 July 2025, and renewed hostilities in December.

Proposed approach. It is factual and does not take sides; the product is about rivers.
1. **One style for all international land lines.** Thailand-Cambodia gets no special stroke, colour, dash or label.
2. **No maritime lines at all.** The sea is a calm, empty plane, and the Gulf of Thailand claims are not the product's subject.
3. **No labels, markers or highlights at border temples** or other contested sites. Station labels are the only place names on the relief.
4. **Respect NE's precision.** The camera target is already bounded to Thailand (ADR 0004). Avoid close-ups where the 1.85 km median vertex spacing would show.
5. **Do not clip stations or rivers with the border polygon.** The Mekong is the boundary with Laos along long stretches, so a station on the Thai bank may fall on the far side of a generalised line. This was **not measured**; check it in Phase 3, when the pipeline runs, against the station list.
6. **Draft About-page wording, for the lead to approve:**
   - English: "Boundaries are approximate, from Natural Earth, and do not represent any position on sovereignty."
   - Thai: "เส้นเขตแดนเป็นเพียงเส้นโดยประมาณจาก Natural Earth ไม่ได้แสดงจุดยืนใดเรื่องอธิปไตย"

## 3. Rivers

The 13 rivers the product must show are เจ้าพระยา, ปิง, วัง, ยม, น่าน, ป่าสัก, ท่าจีน, แม่กลอง, มูล, ชี, โขง, ตาปี and บางปะกง.

### 3.1 Natural Earth 10m

The layers are `ne_10m_rivers_lake_centerlines` (v5.0.0) and its `_scale_rank` variant (v5.0.0, split by `strokeweig`).
- **Inventory in the bbox:** 53 and 129 features, with 40 distinct names. Fields include `scalerank`, `min_zoom`, `strokeweig` (scale_rank variant only) and `wikidataid`.
- **No Thai-script field.**

Matching by name alone (`ne_rivers.py`) finds 7 of the 13 names. Positional agreement with OSM (`river_agreement.py`) shows the real picture:
- OSM centre lines were sampled every 1 km, inside 96.8 to 106.2°E and 5.1 to 21.0°N.
- For each sample the script found the nearest NE line and its name.

| River | NE name on the matching line | OSM samples within 2 km of NE | NE `scalerank` | NE `strokeweig` |
|---|---|---|---|---|
| Mekong | Mekong (61%) and Lancang (32%) | 94% | 2 | 0.4 to 1.5 |
| Chao Phraya | Chao Phraya | 66% | 7 | 0.3 |
| Tha Chin | **"Chao Phraya"** | 100% | 7 | 0.3 |
| Nan | **"Yom"** | 98% | 7 | 0.15 to 0.3 |
| Ping | Ping | 98% | 7 | 0.2 to 0.3 |
| Mae Klong | Mae Klong | 95% | 7 | 0.15 to 0.25 |
| Mun | Mun | 99% | 8 | 0.15 to 0.3 |
| Chi | Chi | 93% | 8 | 0.15 to 0.25 |
| Yom | none | 0% | | |
| Wang | none (5% near "Ping") | 5% | | |
| Pa Sak | none | 0% | | |
| Tapi | none | 0% | | |
| Bang Pakong | none | 0% | | |

- **8 of the 13 are present as geometry, and 6 carry the right name.** One of those six, Chao Phraya, leaves the real river: NE's "Chao Phraya" line follows the Tha Chin downstream.
- **NE's "Yom" is the Nan.**
- **The rest are missing:** Yom, Wang, Pa Sak, Tapi and Bang Pakong.
- **Verdict:** NE is not usable for the named rivers.
- **Reservoirs:** NE 10m lakes in the bbox include "Bhumphol K. K. Nam" and "Siri Kit Dam" (English names only), but not the Pa Sak Jolasid reservoir.

### 3.2 OpenStreetMap via Overpass

Two queries were sent to `https://overpass-api.de/api/interpreter`. Both hit "server at capacity" (502) once and succeeded after a backoff. Their text is in `geodata/.cache/osm/*.overpassql`.
1. **Find the relations:** `waterway=river` relations whose `name:th` or `name` matches `^แม่น้ำ(เจ้าพระยา|…|บางปะกง)$` inside Thailand's bbox. This returned **exactly 13 relations, one per river.**
2. **Fetch geometry:** `out geom` for those 13 relations. OSM base timestamp 2026-10-29T11:17:02Z.

Measured with `osm_rivers.py`, clipped to the heightmap bbox. Centre line means members with role `main_stream` or no role.

| River | OSM relation | Ways | km (all) | km (centre line) | Pieces after merge | Largest gap | Vertices at DP 0.002° / 0.005° |
|---|---|---|---|---|---|---|---|
| Mekong | 215354 | 137 | 3,088 | 3,011 | 1 | 0 | 1,067 / 552 |
| Nan | 227210 | 34 | 838 | 838 | 1 | 0 | 659 / 325 |
| Chao Phraya | 227317 | 196 | 380 | 380 | 2 | 0.2 km | 227 / 121 |
| Ping | 227319 | 35 | 648 | 637 | 2 | **7.1 km** | 394 / 207 |
| Tha Chin | 228063 | 30 | 326 | 324 | 2 | 0.1 km | 256 / 133 |
| Pa Sak | 229599 | 85 | 743 | 743 | 2 | 0.4 km | 771 / 282 |
| Chi | 392516 | 62 | 1,007 | 1,007 | 1 | 0 | 1,004 / 455 |
| Mun | 392517 | 44 | 827 | 827 | 1 | 0 | 744 / 304 |
| Tapi | 2676664 | 22 | 325 | 272 | 2 | **5.1 km** | 252 / 112 |
| Wang | 11700349 | 12 | 325 | 325 | 5 | 0 (pieces touch) | 292 / 142 |
| Mae Klong | 12911358 | 10 | 138 | 138 | 1 | 0 | 80 / 47 |
| Yom | 14030694 | 28 | 940 | 735 | 5 | 0.1 km | 661 / 302 |
| Bang Pakong | 19264378 | 8 | 240 | 240 | 1 | 0 | 197 / 113 |

- **Gaps:** Ping and Tapi have real gaps in their centre line, of 7.1 km and 5.1 km. The pipeline must bridge them, either with the `side_stream` members or by hand, and then check them visually.
- **Size:** all 13 simplified at 0.002° (about 0.2 km) and written as GeoJSON with 4-decimal coordinates.

  | Simplification | Vertices | Raw | gzip 9 | brotli 11 |
  |---|---|---|---|---|
  | 0.002° | 6,604 | 124,201 | 38,496 | 22,931 |
  | **0.005° (about 0.55 km, half a DEM pixel)** | **3,095** | **59,076** | **19,379** | **11,923** |

- **Names:** all 13 have a `wikidata` tag.
  - 11 have `name:th`. Wang and Yom lack it, but their `name` is Thai. Wang also lacks `name:en`.
  - The Wikidata Thai label equals OSM `name:th`, or `name` where `name:th` is absent, **for all 13** (`thai_labels_check.py`).
- **Licence:** ODbL 1.0. The OSM copyright page says:

  > You are free to copy, distribute, transmit and adapt our data, as long as you credit OpenStreetMap and its contributors. If you alter or build upon our data, you may distribute the result only under the same license.

  It requires two things: "Provide credit to OpenStreetMap by displaying our attribution notice" and "Make clear that the data is available under the Open Database License."

  For us, the simplified river file is a derived database. It is published under ODbL, with the Overpass queries and the script beside it in the repo. The app shows "© OpenStreetMap contributors" linked to https://www.openstreetmap.org/copyright.

### 3.3 HydroRIVERS (HydroSHEDS)

- **Product:** HydroRIVERS v1.0, derived from HydroSHEDS at 15 arc-seconds, with streams starting at 10 km² upstream area or 0.1 m³/s.
  - Asia file `HydroRIVERS_v10_as_shp.zip`: 90,510,995 bytes, downloaded.
  - Global file: 544,388,154 bytes by HEAD, **not downloaded** because it is over the 500 MB limit.
- **Fields** (TechDoc v1.0 section 3.2):
  - `ORD_STRA`, the Strahler order;
  - `ORD_CLAS`, the classical order, where 1 is the main stem from sink to source;
  - `ORD_FLOW`, a logarithmic discharge class;
  - `UPLAND_SKM`, the upstream area;
  - `DIS_AV_CMS`, the long-term average discharge.
  - **No names.**
- **Coverage:** 242,926 reaches intersect the bbox.
- **Position:** against OSM, 84 to 100% of OSM samples lie within 2 km of a reach with an upstream area of 500 km² or more. Medians are 0.16 to 0.49 km (table in section 3.5).
- **Naming by tracing:** names would have to be attached by tracing. `hydrorivers_trace.py` follows maximum upstream area, guided by a 0.03° corridor around OSM.
  - It recovers the whole OSM length for 6 rivers: Ping, Nan, Pa Sak, Mae Klong, Mun and Bang Pakong.
  - The other 7 stop early at confluences or in braided reaches. Coverage was Chao Phraya 30%, Yom 7%, Wang 9%, Tha Chin 78%, Chi 23%, Mekong 10% and Tapi 79%. Each would need hand-placed waypoints.
- **Network sizes** if a background network were wanted. These use a custom int16 delta binary with brotli, after 0.005° simplification and 0.001° quantisation:

  | Subset | Reaches | km | Binary + brotli | GeoJSON + brotli |
  |---|---|---|---|---|
  | `ORD_STRA` ≥ 6 | 6,847 | 18,702 | 41,132 | 53,329 |
  | `UPLAND_SKM` ≥ 5,000 km² | 9,530 | 26,798 | 58,332 | 76,121 |
  | `ORD_FLOW` ≤ 4 (≥ 100 m³/s) | 11,307 | 29,737 | 67,725 | 88,843 |

- **Licence.** The product page says it is "freely available for scientific, educational and commercial use" under the HydroSHEDS licence, which is reproduced in HydroSHEDS TechDoc v1.4, Appendix A. Its terms that matter here:
  - **2.1.2:** distribution only "incorporated into any Derivative Works to End Users, subject to the terms and conditions of an end user license agreement with terms that are at least as protective"; and "In no event shall Licensee license or distribute the Licensed Materials as a stand-alone product."
  - **2.1.3:** the licensee "shall protect against unauthorized copying and/or distribution of Licensed Materials".
  - **Exhibit B:** a required notice must be "displayed with, attached to or embodied in (in a reasonably prominent manner) the documentation or metadata". It begins: "This product [insert Licensee Derivative Product name] incorporates data from the HydroSHEDS version 1 database which is © World Wildlife Fund, Inc. (2006-2022) …"

  A public web app that serves a static vector file anyone can download, from an open repository, does not clearly meet the end-user-licence and copy-protection terms. This is a legal judgement and **unverified**. It is the main reason not to ship HydroRIVERS.

### 3.4 Recommendation

**OSM alone for the 13 named rivers.**
- **Geometry:** the 13 relations, centre-line members, the Ping and Tapi gaps bridged, simplified at 0.005°. That is 11.9 KB brotli.
- **Names:** a curated `rivers.csv` keyed by OSM relation and Wikidata QID, with Thai names from Wikidata and OSM, which agree for all 13.
- **Line width by river order:**
  - The brief asks for width by order. OSM has no order attribute, so use the **classical order by topology**, which follows HydroRIVERS' `ORD_CLAS` definition and is public knowledge:
    - **1** = reaches the sea: Mekong, Chao Phraya, Tha Chin (a distributary), Mae Klong, Bang Pakong and Tapi.
    - **2** = flows into an order-1 river: Ping, Nan, Pa Sak and Mun.
    - **3** = flows into an order-2 river: Wang (into the Ping), Yom (into the Nan) and Chi (into the Mun).
  - Classical order puts the Tapi in the same class as the Mekong. Look-dev may special-case the Mekong or taper width with distance from the mouth, which is pure geometry. Record that in ADR 0006.
- **HydroRIVERS:** not shipped. It may still be used offline as a QA reference for OSM geometry.
- **Natural Earth rivers:** not used.

### 3.5 The 13 rivers

The English names are the Wikidata English labels. "HR within 2 km" is the share of OSM samples within 2 km of a HydroRIVERS reach with an upstream area of 500 km² or more, followed by the median distance.

| Thai | English | Wikidata | OSM relation | OSM | Natural Earth 10m | HydroRIVERS: within 2 km / median | Order |
|---|---|---|---|---|---|---|---|
| แม่น้ำเจ้าพระยา | Chao Phraya River | Q118850 | 227317 | yes, `name:th` | partly (66%) | 87% / 0.21 km | 1 |
| แม่น้ำปิง | Ping River | Q592383 | 227319 | yes, `name:th`, 7.1 km gap | yes | 100% / 0.16 km | 2 |
| แม่น้ำวัง | Wang River | Q666031 | 11700349 | yes, Thai `name` only | no | 100% / 0.28 km | 3 |
| แม่น้ำยม | Yom River | Q586733 | 14030694 | yes, Thai `name` only | no (NE "Yom" is the Nan) | 100% / 0.17 km | 3 |
| แม่น้ำน่าน | Nan River | Q1153522 | 227210 | yes, `name:th` | as "Yom" | 94% / 0.16 km | 2 |
| แม่น้ำป่าสัก | Pa Sak River | Q1152315 | 229599 | yes, `name:th` | no | 90% / 0.49 km | 2 |
| แม่น้ำท่าจีน | Tha Chin River | Q1359498 | 228063 | yes, `name:th` | as "Chao Phraya" | 96% / 0.33 km | 1 |
| แม่น้ำแม่กลอง | Mae Klong | Q1194112 | 12911358 | yes, `name:th` | yes | 97% / 0.17 km | 1 |
| แม่น้ำมูล | Mun River | Q1455933 | 392517 | yes, `name:th` | yes | 98% / 0.35 km | 2 |
| แม่น้ำชี | Chi River | Q1071554 | 392516 | yes, `name:th` | yes | 94% / 0.47 km | 3 |
| แม่น้ำโขง | Mekong River | Q41179 | 215354 | yes, `name:th` | yes (Mekong and Lancang) | 97% / 0.21 km | 1 |
| แม่น้ำตาปี | Tapi River | Q1195715 | 2676664 | yes, `name:th`, 5.1 km gap | no | 84% / 0.37 km | 1 |
| แม่น้ำบางปะกง | Bang Pakong River | Q806449 | 19264378 | yes, `name:th` | no | 99% / 0.27 km | 1 |

**Thai labels, reliably:**
- Never derive them at runtime.
- The pipeline reads the curated CSV and asserts that each OSM relation's `name:th` (or `name`) and the Wikidata Thai label equal the CSV value. A mismatch fails the build rather than silently changing a label.
- Labels are DOM overlays (design brief), so Thai shaping is the browser's job.

### 3.6 Reservoirs (optional)

ภูมิพล and สิริกิติ์ are in NE 10m lakes, with English names and approximate areas of 123 and 269 km² from a rough planar calculation. ป่าสักชลสิทธิ์ is not. OSM almost certainly maps all three as water polygons, but that is **unverified**: the two-query Overpass budget was used for rivers.

The DEM already shows reservoir surfaces as flat areas. If look-dev wants water fills, one more polite Overpass query in Phase 3 can fetch the three polygons, under the same ODbL terms.

## 4. Budget summary

| Asset | Source | Licence | Size (measured) | Attribution line |
|---|---|---|---|---|
| `heightmap.png`: 2048 × 1815, 16-bit gray, sqrt 11-bit, Thailand priority | Terrain Tiles z8 (SRTM, GMTED2010, ETOPO1), NE land for the sea mask, NE admin-0 for the priority mask | U.S. Government public domain data with requested credit; NE is public domain | **1,269,496 B** (r = 2, 75 km band); 1,454,568 B with r = 1 | the credit line in section 1.2 |
| `provinces.topo.json`: 77 polygons, 500 m | NE 10m admin-1 v5.1.1, names from the curated table | public domain; Wikidata labels CC0 | 85,557 raw / 29,586 gzip / **26,580 br** | optional: "Made with Natural Earth." |
| `borders.topo.json`: international land lines, 500 m | NE 10m admin-0 boundary lines land v5.1.0 | public domain | 43,406 raw / 16,464 gzip / **13,796 br** | optional: "Made with Natural Earth." |
| `rivers.geojson`: 13 rivers, DP 0.005° | OSM via Overpass, 2026-10-29 | **ODbL 1.0** | 59,076 raw / 19,379 gzip / **11,923 br** | "© OpenStreetMap contributors, ODbL", linked to https://www.openstreetmap.org/copyright |
| `provinces.csv` and `rivers.csv` | Wikidata labels checked against NE and OSM | CC0 for the labels; the ODbL file carries the OSM IDs | not built; roughly 90 rows, a few KB (**estimate**) | none required |
| optional `thailand_mask.png` at 2048 rows | NE admin-0 | public domain | 23,334 B as PNG, 12,188 B as WebP | none required |

The **total** initial geodata, with the heightmap at r = 2, is 1,269,496 + 26,580 + 13,796 + 11,923 = **1,321,795 bytes**, plus a few KB for names and 23 KB if the mask is shipped.

That the WebGL canvas is not an LCP candidate, so the heightmap delays the relief but not LCP, is **unverified**.

## 5. Open questions and unverified items

- **Browser behaviour** was not tested in a browser:
  - whether 16-bit PNGs are reduced to 8 bits through `<img>` or `createImageBitmap`;
  - whether canvas or WebGL readback is noised under Safari's Advanced Fingerprinting Protection or Firefox's canvas randomisation;
  - whether `R16F` is filterable in core WebGL2;
  - how long the worker takes to decode the PNG on a mid-range phone.

  Phase 3 should check all four with a 30-line probe page.
- **ThaiWater coordinates** are assumed to be WGS84.
- **Natural Earth v5.1.2** has not been checked against 5.1.1 for Thai features.
- **The HydroSHEDS licence**, as applied to a public web app, needs the lead's judgement. HydroRIVERS is proposed out of the shipped assets either way.
- **The border seam** in the Thailand-priority heightmap needs a look-dev call (section 1.8 previews). If it reads as an artefact, use the 75 km band and r = 1 (1.45 MB), or fall back to the uniform 1536-row PNG.
- **Smoothing the plain:** the plain's DEM noise (1.31 m standard deviation between neighbours) may read as texture at 6 to 15× exaggeration. Smoothing below about 30 m is a pipeline option for look-dev.
- **Stations and rivers near the Mekong border** were not checked against the generalised NE line.
- **OSM reservoirs** were not queried.

## Decisions (proposed)

1. **DEM source:** AWS Terrain Tiles, Terrarium, z8. The 224 tiles are fetched once with concurrency 4 and cached in `geodata/.cache/terrarium/`. Use the documented path, not `v2/`.
2. **Projection:** an equirectangular lon/lat grid over 92 to 111°E and 3 to 24°N, with pixels square at 13°N. One linear function maps lat/lon to scene coordinates, shared by the pipeline, client and tests.
   - Constants: 108.4847 km per degree of longitude and 110.6305 km per degree of latitude.
   - Record this in an ADR, because every client coordinate depends on it.
3. **Heightmap:**
   - 2048 × 1815, with sea code 0 and the land mask NE land ≥ 50% OR DEM above 1 m.
   - Land is 1 + round(sqrt(h / 3300) × 2046), with h clamped to 0 or more.
   - Thailand priority with a 75 km fade band and neighbours smoothed at r = 2 to start, chosen finally in look-dev.
   - Stored as a 16-bit grayscale PNG under 1.5 MB, decoded in a worker with `DecompressionStream('deflate')`.
   - Kept on the CPU as `Uint16` codes plus a lookup table, and on the GPU as `R16F`.
   - No bathymetry.
4. **Fallback encoding:** MED residual + brotli `.bin`, about 21% smaller, if the host can serve `Content-Encoding: br`.
5. **Borders:** NE 10m v5.1.x, pinned in the pipeline.
   - Province TopoJSON at 500 m, with the national outline from its outer arcs, and international land lines at 500 m.
   - No maritime or disputed lines, and one style for every international line.
   - No labels at contested sites, and a neutral "boundaries approximate" note on the About page.
6. **Thai names** come from curated CSVs (provinces keyed by ISO 3166-2, rivers by OSM relation and Wikidata QID), seeded from Wikidata and asserted against OSM in the build. Never use NE `name_local`.
7. **Rivers:** OSM only (ODbL), 13 relations, Ping and Tapi gaps bridged, simplified at 0.005°.
   - Width by classical order: 1 is Mekong, Chao Phraya, Tha Chin, Mae Klong, Bang Pakong and Tapi; 2 is Ping, Nan, Pa Sak and Mun; 3 is Wang, Yom and Chi.
   - Final widths are set in ADR 0006.
8. **HydroRIVERS** is not shipped, because of its licence. Offline QA use only.
9. **Credits** go on the About page and in `geodata/SOURCES.md`:
   - the elevation line from section 1.2;
   - "© OpenStreetMap contributors (ODbL)" with a link;
   - optionally, "Made with Natural Earth."

## 6. Measurement scripts and reports

All scripts are run from the repository root with the research venv: `.cache/venv/bin/python -I …`. The Python packages are pinned in `geodata/scripts/research/requirements.txt`. mapshaper 0.7.101 was installed with npm into the gitignored `.cache/node-scratch/`. Data paths are passed as arguments, and outputs go to `geodata/.cache/`.

| Script | What it measures | Command used |
|---|---|---|
| `geodata/scripts/research/fetch_terrarium.py` | tile ranges, a polite download, imagery-source headers | `--z 8 --bbox 92 3 111 24 --out geodata/.cache/terrarium` (also `--z 7`, `--heads`, `--dry-run`) |
| `geodata/scripts/research/heightmap_measure.py` | stitch, reproject, masks, every encoding in sections 1.6 to 1.8, previews | `--tiles geodata/.cache/terrarium --z 8 --bbox 92 3 111 24 --lat0 13 --land geodata/.cache/naturalearth/ne_10m_land/ne_10m_land.shp --heights 1024 1536 2048 2560 --out geodata/.cache/heightmap_candidates`, then reruns with `--skip-basic --sqrt-bits …`, `--tha <ne_10m_admin_0_countries.shp> --prio-bits … --nb-blur {0,1,2} --band-km {25,75}` |
| `geodata/scripts/research/ne_inspect.py` | NE fields, the 77 Thai rows, Thai-script fields, boundary classes | `geodata/.cache/naturalearth` |
| `geodata/scripts/research/thai_labels_check.py` | NE `name_local` against Wikidata; OSM against Wikidata river labels | `geodata/.cache/naturalearth geodata/.cache/wikidata/labels.json geodata/.cache/osm/rivers_relations.json` |
| `geodata/scripts/research/ne_borders.py` | disputed and maritime features, site containment, TH-KH segment stats | `geodata/.cache/naturalearth 92 3 111 24` |
| `geodata/scripts/research/borders_sizes.sh` | mapshaper TopoJSON and GeoJSON sizes | `.cache/node-scratch/node_modules/.bin/mapshaper geodata/.cache/naturalearth geodata/.cache/borders_candidates` |
| `geodata/scripts/research/border_vertices_mask.py` | arc and vertex counts, Thailand mask size | `geodata/.cache/naturalearth geodata/.cache/borders_candidates geodata/.cache/borders_candidates` |
| `geodata/scripts/research/ne_rivers.py` | NE river inventory, name matching, lakes | `geodata/.cache/naturalearth 92 3 111 24` |
| `geodata/scripts/research/osm_rivers.py` | per-relation ways, km, pieces, gaps, DP vertices, output sizes | `geodata/.cache/osm/rivers_geom.json 92 3 111 24 --out geodata/.cache/osm/candidates` |
| `geodata/scripts/research/hydrorivers_extract.py` | HydroRIVERS reaches in the bbox, to JSON | `<HydroRIVERS_v10_as.shp> 92 3 111 24 geodata/.cache/hydrosheds/hr_bbox.json` |
| `geodata/scripts/research/hydrorivers_trace.py` | tracing the 13 rivers, network subset sizes | `geodata/.cache/hydrosheds/hr_bbox.json geodata/.cache/osm/rivers_geom.json` |
| `geodata/scripts/research/river_agreement.py` | OSM against HydroRIVERS and NE positional agreement, NE names | `geodata/.cache/osm/rivers_geom.json geodata/.cache/hydrosheds/hr_bbox.json <ne_10m_rivers_lake_centerlines_scale_rank.shp>` |

The reports are under `geodata/.cache/`:
- heightmap: `heightmap_z8_report.json`, `heightmap_z8_sqrt_report.json`, `heightmap_z8_prio_report.json`, `heightmap_z8_med_report.json`, `heightmap_z8_noblur_report.json`, `heightmap_z8_blur1_report.json`, `heightmap_z8_band75_b1_report.json`, `heightmap_z8_band75_b2_report.json`;
- Natural Earth: `naturalearth/ne_inspect_report.txt` and `naturalearth/ne_rivers_report.txt`;
- borders: `borders_candidates/ne_borders_report.txt` and `borders_candidates/vertices_mask_report.txt`;
- OSM: `osm/osm_rivers_report.txt` and `osm/river_agreement_report.txt`;
- HydroRIVERS: `hydrosheds/trace_report.txt`.

## Sources

All accessed 2026-10-29.

- AWS Terrain Tiles, Terrarium tiles at z7 and z8, plus HEAD requests (one on `v2/`): https://s3.amazonaws.com/elevation-tiles-prod/terrarium/{z}/{x}/{y}.png, accessed 2026-10-29
- Registry of Open Data on AWS, Terrain Tiles: https://registry.opendata.aws/terrain-tiles/, accessed 2026-10-29
- Registry YAML: https://raw.githubusercontent.com/awslabs/open-data-registry/main/datasets/terrain-tiles.yaml, accessed 2026-10-29
- joerd attribution: https://github.com/tilezen/joerd/blob/master/docs/attribution.md, accessed 2026-10-29
- joerd formats (Terrarium encoding): https://github.com/tilezen/joerd/blob/master/docs/formats.md, accessed 2026-10-29
- joerd data sources (sources per zoom, imagery headers): https://github.com/tilezen/joerd/blob/master/docs/data-sources.md, accessed 2026-10-29
- Natural Earth downloads, the 11 layers in section 2.1: https://naciscdn.org/naturalearth/10m/cultural/ and https://naciscdn.org/naturalearth/10m/physical/, accessed 2026-10-29
- Natural Earth terms of use: https://www.naturalearthdata.com/about/terms-of-use/, accessed 2026-10-29
- Natural Earth latest release (v5.1.2): https://api.github.com/repos/nvkelso/natural-earth-vector/releases/latest, accessed 2026-10-29
- Natural Earth master `VERSION` (5.2.0-pre): https://raw.githubusercontent.com/nvkelso/natural-earth-vector/master/VERSION, accessed 2026-10-29
- HydroRIVERS product page and licence statement: https://www.hydrosheds.org/products/hydrorivers, accessed 2026-10-29
- HydroRIVERS Asia shapefile: https://data.hydrosheds.org/file/HydroRIVERS/HydroRIVERS_v10_as_shp.zip, accessed 2026-10-29
- HydroRIVERS global shapefile, HEAD only: https://data.hydrosheds.org/file/HydroRIVERS/HydroRIVERS_v10_shp.zip, accessed 2026-10-29
- HydroRIVERS Technical Documentation v1.0: https://data.hydrosheds.org/file/technical-documentation/HydroRIVERS_TechDoc_v10.pdf, accessed 2026-10-29
- HydroSHEDS Technical Documentation v1.4 (licence in Appendix A, Exhibit B): https://data.hydrosheds.org/file/technical-documentation/HydroSHEDS_TechDoc_v1_4.pdf, accessed 2026-10-29
- Overpass API, 2 queries: https://overpass-api.de/api/interpreter, accessed 2026-10-29
- OpenStreetMap copyright and licence: https://www.openstreetmap.org/copyright, accessed 2026-10-29
- Wikidata SPARQL endpoint, 1 query for 77 province and 13 river labels: https://query.wikidata.org/sparql, accessed 2026-10-29
- Wikidata copyright (CC0 for structured data): https://www.wikidata.org/wiki/Wikidata:Copyright, accessed 2026-10-29
- Wikipedia, Preah Vihear (summary and coordinates): https://en.wikipedia.org/api/rest_v1/page/summary/Preah_Vihear_Temple, accessed 2026-10-29
- Wikipedia, Prasat Ta Muen Thom (summary and coordinates): https://en.wikipedia.org/api/rest_v1/page/summary/Prasat_Ta_Muen_Thom, accessed 2026-10-29
- Wikipedia, Ko Kut district (summary and coordinates): https://en.wikipedia.org/api/rest_v1/page/summary/Ko_Kut_district, accessed 2026-10-29
- Wikipedia, 2025 Cambodian–Thai border crisis (summary): https://en.wikipedia.org/api/rest_v1/page/summary/2025_Cambodia%E2%80%93Thailand_border_conflict, accessed 2026-10-29
- MDN browser-compat-data, DecompressionStream: https://raw.githubusercontent.com/mdn/browser-compat-data/main/api/DecompressionStream.json, accessed 2026-10-29
- MDN browser-compat-data, createImageBitmap: https://raw.githubusercontent.com/mdn/browser-compat-data/main/api/_globals/createImageBitmap.json, accessed 2026-10-29
