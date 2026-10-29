# ADR 0004: Render the relief with React Three Fiber and our own terrain material, not a map library

- **Status:** accepted (Phase 1). The look itself is tuned in Phase 3 and recorded in ADR 0006.
- **Date:** 2026-10-29
- **Deciders:** implementer, for review by the lead

## Context

The relief is the page. The design brief describes it as a plaster relief model of Thailand in a dark map room, lit only where water is measured. Here is what the brief asks of the renderer.

**Look:**
- Swiss-style shading from one north-west key light and a soft fill, graded by elevation.
- Vertical exaggeration of about 6 to 15×.
- Subtle fog or aerial perspective.
- No starfield, particles, halos or lens flare. Antialiasing on, and no z-fighting on borders or rivers.

**Interaction:**
- One authored moment: the ใกล้ฉัน flight, an exponential ease-out over 1.2 to 2.0 s that the user can interrupt.
- Thai labels as DOM overlays.

**Rendering mechanics:**
- `frameloop="demand"` and adaptive DPR.
- Instanced gauges in a single draw call, with shader materials that read the colour tokens.
- A designed 2D fallback without WebGL.

**Budgets:**
- Initial JS under 300 KB gzipped, *excluding the lazy 3D chunk*.
- LCP under 2.5 s on throttled Fast 4G on a mid-tier phone, and J1's first frame in under 2 s.
- 60 fps on desktop and 30 or more on a mid-range phone, with no long tasks over 200 ms.
- A heightmap of 1.5 MB or less.

**The scene is small and fixed.** It is one country, one static heightmap, about 800 points, a border and about a dozen rivers. Nothing has to stream, and nothing is tiled.

**Measured on 2026-10-29** (gzip -9 of minified output):

| Stack | Gzipped JS | How it was measured |
|---|---|---|
| three 0.186.1 | 183.7 KB | throwaway Vite 8 build of a minimal scene (displaced plane, `ShaderMaterial`, `InstancedMesh`) |
| @react-three/fiber 9.8.2 | 55.0 KB | same build |
| drei `CameraControls` | 11.0 KB | same build, imported on its own |
| **R3F stack, total** | **≈ 250 KB** | lazy chunk |
| MapLibre GL JS 6.16.0 | 294.3 KB main + 147.3 KB worker + 10.2 KB CSS | published `dist/` files |
| deck.gl 9.4.1 | core 240.8 KB, geo-layers 173.2 KB, mesh-layers 92.8 KB, layers 58.9 KB | published UMD bundles; an upper bound, since a tree-shaken ESM build is smaller |

Two caveats on these numbers:
- **R3F pulls in all of three.** Its catalogue imports `three` as a namespace, so tree-shaking does not help; the 183.7 KB is effectively the whole library.
- **The React shell is separate.** In the same build, React and React DOM are a separate 68.6 KB, inside the initial-JS budget.

**Geodata inputs, measured the same day** (detail and scripts in `docs/research/geodata-sources.md`):

| Asset | Source and licence | Shape | Bytes on the wire |
|---|---|---|---|
| Heightmap | AWS Terrain Tiles, Terrarium z8 (224 tiles, 22,449,608 B fetched); public domain with requested credit | 2048 × 1815 equirectangular, 92–111°E by 3–24°N, about 1.13 km per pixel; 16-bit grey PNG of sqrt-11-bit codes; sea is code 0 (no bathymetry); full precision inside a 75 km band around Thailand, neighbours blurred at r = 2 | **1,269,496** (1,454,568 with r = 1) |
| Provinces | Natural Earth 10m admin-1 v5.1.1; names from a curated Wikidata table, because 6 of NE's 77 `name_local` values are wrong | TopoJSON, simplified to 500 m | 26,580 (brotli) |
| International lines | Natural Earth 10m boundary lines (land) v5.1.0 | TopoJSON, 500 m; no maritime lines | 13,796 (brotli) |
| Rivers | OpenStreetMap relations, ODbL 1.0 | 13 rivers, Douglas–Peucker 0.005° (half a DEM pixel) | 11,923 (brotli) |
| **Total** | | | **1,321,795** |

What these facts mean for the renderer:
- **Precision is enough for the plain.** In a test box in the lower Chao Phraya plain the encoding error is 0.045 m RMSE, against the DEM's own neighbour-to-neighbour noise of 1.31 m. The noise, not the encoding, is the risk: at 10× it becomes 13 m bumps, so smoothing below about 30 m is a look-dev option.
- **The projection is fixed by the pipeline.** `u = (lon − 92)/19`, `v = (24 − lat)/21`, `X_km = (lon − 101.5) × 108.4847`, `Z_km = (lat − 13.5) × −110.6305` and `Y = height(u, v) × exaggeration`. Pixels are square at 13°N. The scene, the CPU sampler and the geodata pipeline must share this one function, so it lives in one tested module.
- **Decoding needs no canvas.** A worker inflates the PNG with `DecompressionStream('deflate')` (Chrome 80, Firefox 113, Safari 16.4). The CPU keeps the `Uint16` codes (7.4 MB) and a lookup table; the GPU gets an `R16F` texture (7.4 MB). That `R16F` filters linearly in core WebGL2 is from the ES 3.0 format table, not a browser test, so the Phase 3 probe confirms it.
- **Rivers come from OSM, not Natural Earth.** NE has only 8 of the 13 rivers, labels the Nan as the Yom, and its "Chao Phraya" follows the Tha Chin. OSM's Ping and Tapi centre lines have gaps of 7.1 km and 5.1 km that the pipeline must bridge. HydroRIVERS is not shipped, because its licence limits redistribution.
- **One line style at every border.** The Thailand–Cambodia line is drawn like every other land border, with no labels at the disputed temples. The about page says the lines are approximate, from Natural Earth, and take no position on sovereignty.
- **A simpler fallback exists.** If the priority mask causes trouble, a uniform 1536 × 1361 (1.51 km per pixel) sqrt-11 PNG is 1,300,913 B, with the same plain error.

## Options

1. **A custom scene in R3F.** One displaced terrain mesh with our own `ShaderMaterial`: tint, shading, flat-area toning and fog, all from tokens. Borders and rivers are draped lines. Gauges are one `InstancedMesh`, and labels are DOM. This is the lead's recommendation.
2. **MapLibre GL JS** with a `raster-dem` terrain source and a hillshade layer, a style with no basemap, and a three.js custom layer for the gauges.
3. **deck.gl `TerrainLayer`**, optionally under MapLibre for the camera, with a mesh or column layer for the gauges.
4. **CesiumJS.** A globe engine. It is the largest of all, its defaults assume imagery and a globe, and it is not considered further.
5. **A pre-rendered relief.** A high-quality offline shaded-relief image at a fixed or near-fixed camera, with SVG or canvas gauges on top. This is close to the 2D fallback that J6 requires anyway.

## Evaluation

| | 1 Custom R3F | 2 MapLibre | 3 deck.gl | 5 Pre-rendered |
|---|---|---|---|---|
| **Control of the look** | every pixel is our shader: tint, shade, flat toning, fog | hillshade paint properties plus sky and fog properties; terrain shading follows the library's model | `TerrainLayer` is lit through deck.gl's lighting effects; custom shading means writing deck.gl shader modules | highest, offline |
| **Risk of library defaults bringing slop back** (references.md patterns 5, 8, 9) | none, since nothing is on by default | real: the sky, fog and hillshade defaults must each be set from tokens and tested (pattern 9 is Mapbox's star default; MapLibre has no stars, but its sky and fog still have defaults) | moderate | none |
| **Camera:** authored flight, polar clamp, terrain-following, interruptible | ours to write; `CameraControls` gives smooth, interruptible transitions | `flyTo` and `easeTo` are good; following the terrain on a custom path is ours to write | view-state transitions are fine | fixed or near-fixed |
| **Lazy 3D JS (gzip)** | ≈ 250 KB | ≈ 452 KB (main, worker and CSS) | the UMD bundles for core, geo-layers and mesh-layers total about 507 KB; a tree-shaken build is smaller but was not measured | about 0 |
| **Assets** | one static heightmap and compact vectors; no tile requests | `raster-dem` expects tiles; one static image is possible but goes against the library's grain | `TerrainLayer` decodes elevation images per tile or as one image | one large image per camera |
| **Gauges in one draw call with token shaders** | `InstancedMesh` with our material | a custom layer with three.js inside MapLibre: two render loops to keep in sync | yes, through layers | SVG or canvas |
| **Idle at zero frames** | `frameloop="demand"` | renders on demand already | renders on demand already | yes |
| **Labels** | DOM, as the brief asks | DOM, so the library's glyph pipeline would go unused | DOM | DOM |
| **What the demo needs** (the first frame quiets the room; the flight lands on *your* river) | full control of both | a good flight, a weaker look | adequate | the strongest first frame, but no flight |
| **Work we own** | projection, draping, picking, camera limits, context loss | style, custom-layer sync, fighting defaults | layer configuration and shader modules | the image pipeline |

Map libraries pay off when a scene streams tiles across zoom levels, projects onto a globe or shows a basemap. Ours does none of these. What they would bring us is a camera and a terrain mesh. In exchange we would pay roughly 200 KB more JavaScript, accept a lighting model we do not control, and take on several defaults that `references.md` names as the source of common 3D slop. Option 5 has the best possible still image, but it cannot fly. It is kept as the fallback.

## Decision

**Option 1: a custom R3F scene.** The intended shape is below. The parameters are tuned in Phase 3 and recorded in ADR 0006.

**Terrain:**
- **Mesh.** One terrain mesh in the projection above, shared with the geodata pipeline. Its heights come from the 2048 × 1815 heightmap as an `R16F` texture, and the vertex shader applies the exaggeration. The mesh can have fewer vertices than the heightmap has pixels; Phase 3 measures the trade.
- **CPU copy.** The same `Uint16` codes are sampled bilinearly on the CPU to drape the border and rivers, to seat the gauge bases, and to keep the camera above the terrain during flights.

**Shading:**
- One north-west key light and a soft fill.
- Flat-area toning on the plain.
- Tint graded by elevation.
- Fog from tokens.
- No shadow maps, no ambient occlusion and no post-processing bloom.
- Normals or a hillshade term can be precomputed in the geodata pipeline if computing them per pixel costs too much on phones.

**Lines:**
- The border and rivers are drawn a fixed small lift above the sampled surface, with a depth bias. Neither is ever tested for depth against a coplanar surface, which prevents z-fighting.

**Gauges:**
- One `InstancedMesh` with a token-driven shader material.
- Picking is screen-space: project the about 800 positions and take the nearest within a touch radius of at least 22 px. This is cheaper and more forgiving than ray-casting thin instances, and it gives the 44 px touch targets.

**Camera:**
- drei `CameraControls`, clamped in polar angle, distance and pan to Thailand's extent.
- Flights use its transitions with our easing, and any input stops them.

**Rendering:**
- `frameloop="demand"`. Controls, flights, timeline scrubs and data changes call `invalidate()`.
- Adaptive DPR.
- On `webglcontextlost`, the app switches to the 2D fallback rather than showing a dead canvas.

**Loading:**
- The 3D chunk loads with `import()` after the shell and the national sentence have painted.

## Fallback plan if look-dev fails

**When to decide.** The plan is decided at the end of the second Phase 3 gate iteration. If the relief still fails the anti-slop checklist or the lead's look review, or if a throttled phone measurement shows under 30 fps, take the first step that fixes the failure.

1. **Bake the look and keep the flight.** The geodata pipeline renders the full shaded relief offline (tint, hillshade and flat toning, at more resolution than real-time lighting allows). It is draped as an unlit texture on a lower-polygon displaced mesh. This is the classic Swiss and Patterson workflow, and it is easier to make beautiful than real-time lighting. We keep the 3D camera and the ใกล้ฉัน flight. We lose only light that responds to the view.
2. **Make 2D the default on phones.** If 3D cannot hold 30 fps on a mid-range phone, the designed 2D view (the baked relief image with SVG gauges, which J6 requires anyway) becomes the phone default. Desktop keeps 3D.
3. **Fall back to MapLibre, as a last resort.** If our own camera or terrain code proves unworkable, switch to MapLibre with `raster-dem` and hillshade, a style written from our tokens, and every sky and fog property set explicitly, with a test that checks the style's sky block against the tokens. The gauges go in a custom layer. This costs about 200 KB more lazy JavaScript and most of the shader control.

Steps 1 and 2 reuse work the project does anyway: the geodata pipeline and the J6 fallback. That is why they come first.

## Consequences

**Positive:**
- The tokens drive every visible colour, the light and the fog.
- No third-party style can bring back a starfield, a satellite drape or a stock basemap.
- The lazy 3D chunk is about 250 KB against about 450 KB for MapLibre.
- No tile requests are made, so first paint never waits on streaming, and the CSP stays at `'self'`.

**Negative: work that is ours to own.**
- The projection, line draping and z-fighting, picking, camera limits, label placement and decluttering, DPR and resize, and context loss.
- Each of these gets a small pure function with unit tests where it can: projection, nearest-in-screen-space and camera clamps.

**Negative: costs and limits.**
- **The whole of three is in the lazy chunk.** That is acceptable, because the budget excludes the 3D chunk, but it is the main reason the shell must paint without it.
- **Station close-ups are limited by the DEM.** With one static heightmap, the finest terrain detail is about 1.13 km per pixel. A valley 3 km wide is under three pixels across. The product shows no streets or inundation, so a close-up only needs the valley's shape, and the camera's closest distance is clamped to what the heightmap can show.
- **About 15 MB of client memory for heights**, split between the CPU codes and the GPU texture. That is acceptable on a mid-range phone, but it is one more reason the 3D chunk loads only after the shell.
- **The canvas needs a text alternative**, listing the national sentence and the selected station. This is done in Phase 5.
