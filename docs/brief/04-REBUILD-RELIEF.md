# Brief 04: Relief rebuild — MapLibre base, our gauge layer (decision 2026-10-30 ~02:00)

> The owner judged the Phase 3 R3F relief **not acceptable visually** (their words: ยังไม่สวยเลย).
> Lead research (four researchers + synthesis, 2026-10-30) concurs: the scene sits at ~3/10 of the
> premium standard, with 15 concrete quality bugs. The owner chose **Option A** from the decision table:
> adopt MapLibre GL JS as the terrain/basemap engine, keep our gauge data layer.

## What top-tier looks like (the target)
Normal-based multi-directional hillshade over a hypsometric ramp; ONE continuous aerial-perspective
fog unifying terrain AND data layers; dead-dark sea with a depth gradient and coast blend; one calm
animated data element; quiet DOM Thai typography; a choreographed intro flight. Materials and
lighting, not heavy postprocessing. Benchmarks: Mapbox Standard style, Mini Tokyo 3D, earth.nullschool.

## Why MapLibre
- BSD-3, mature, the shaded-relief/fog/pitch/DPR/mobile-perf stack is debugged by hundreds of devs
  (it moots most of the 15 bugs: terrain shading, camera, depth precision, world edge).
- Proven pattern for live data objects on a map: Mini Tokyo 3D (nagix, MIT) renders realtime
  objects as MapLibre custom layers. Protomaps PMTiles (BSD-3, ODbL data) gives a dark vector
  basemap with Thai names without any token or paid API.
- Our existing assets survive: the heightfield pipeline (re-encode once to terrarium256),
  the gauge instancing concept (ports to a custom layer), tokens (colors feed the style JSON),
  the recorded fixture, the e2e approach, DOM Thai labels (Markers).

## Execution order (de-risked, the lead's binding instruction)
1. **Base frame first, tonight.** MapLibre + terrarium DEM (re-encode our heightmap with a
   committed one-off script, or AWS terrain tiles as fallback) + hillshade layer + setFog +
   Protomaps dark basemap + maplibre-contour. Target: the national view makes the room go quiet
   WITHOUT any custom layer. Screenshot + commit. If globe projection (MapLibre v5+) is available,
   prefer it; otherwise Mercator is acceptable — record the choice in ADR 0007.
2. **Gauges as a custom layer, tomorrow morning.** Port the bank-full encoding (cap + fill rising
   past the cap, stale hollow, suspect marked) to a MapLibre custom layer using instanced boxes
   and our gauge shader logic, now WITH fog, ground contact (base ring), and a slim profile.
   Fix the encoding bugs listed below. **In parallel, build the disc encoding (MapLibre circle
   layer) behind a feature flag as the guaranteed fallback** — by midday the lead will pick
   between poles and discs on visual merit, no heroics.
3. **Interaction port:** presets (jumpTo), the intro flight (flyTo with curve), reduced-motion
   cuts, ใกล้ฉัน prep (flyTo around a lngLat), LINE browser handling.
4. **Then re-gate** (new GATE-phase-3b.md with the same checklist) and stop for the lead review.

## The 15 bugs (fix where they survive the port; most terrain ones die with MapLibre)
1 | Rivers/border lines tunnel through terrain: drape() samples height only at input GeoJSON vertices (often tens of km apart) and draws straight segments; at 12x exaggeration the ~3.24 km lift is nearly constant while ridge-to-valley deltas are 10-30 km, so segments dive inside mountains. Fix: densify to ~2 km spacing against the h
2 | Inconsistent vertical datums: draped lines sit at ground + ~3.2 km, gauge pole bases at ground - 0.5 km, terrain at ground exactly — rivers visibly float above the floodplain and poles sink. Fix: one datum (lines ground + ~0.5 km, poles seated ground - 0.2 km); after densification the lift can shrink to near zero. web/src/scene/
3 | Gauges and lines are unfogged while terrain fogs: terrain applies aerial fog smoothstep(900,2600) toward #0d1a23 at 85% but the gauge shader and drei <Line> rivers/borders have no fog term — far gauges render as crisp saturated bars over hazed terrain. Fix: pass uFog/uFogNear/uFogFar/uCameraPos into both shaders. web/src/scene/G
4 | Sea is one flat fill (colour = uSea, zero shading/depth gradient/shoreline blend) with a hard coast clip; the file-header comment claims 'a whisper of the depth gradient toward the shore' but that code was never written. Fix: toward-shore gradient (2-tap blur of height code or distance-from-land) + 1-2 texel coast smoothstep. we
5 | Gauges re-upload all 804 instance matrices and allocate a new THREE.Matrix4 every rendered frame (~51 KB GPU upload + heap alloc per frame even under frameloop=demand with nothing changed). Fix: delete the useFrame; set matrices once in a layout effect keyed on [stations, heightfield, exaggeration], needsUpdate only when the mem
6 | Station label pinned to hardcoded y=34 km, unrelated to pole top or exaggeration — at 40x the label is buried in terrain, at 12x it hovers detached. Fix: compute y from the same formula as Gauges (ground*exag + pole height + ~4 km margin), pass exaggeration through. web/src/scene/LookdevScene.tsx:212-215
7 | prefers-reduced-motion read once at mount (useMemo []) so a mid-session OS toggle is ignored; and the 'snap to first preset' mount effect calls setLookAt with CameraControls' default animated transition instead of cutting. Fix: live matchMedia 'change' listener + setLookAt(..., false) on mount. web/src/scene/LookdevScene.tsx:70-
8 | Suspect flag dropped: fixture marks suspect readings (stations carry "suspect":["jump"], summary suspectCount:2) but StationState has no suspect field, so suspect stations render as fully trusted — for an 'instrument' brand this is the opposite of premium. Fix: hollow/striped cap or dashed outline reusing the stale-stroke langua
9 | Gauge overflow encoding clips: waterTop = CAP + 0.7*(1-CAP)/0.6 ~ 1.037 exceeds the 1.0 box top, so every station above ~163% of bank-full renders identically and the alert fill ends flush with the pole top looking like a clipping bug; and the vLocalY > 0.004 guard discards the pole base (~48 m at 12 km poles) so poles visibly f
10 | 804 flat-coloured 2.6 km square-section poles (POLE_WIDTH_KM=2.6, BoxGeometry, heights 12/17/34/44 km, 5 quantised flat hues) overlap into 'dashboard confetti' in the central plain: no depth cue, no ground contact (no ring/AO/disc), level only by hue. Fix: slim blade or cylinder, darker base ring, distance-scaled alpha/saturatio
11 | Terrain shading ceiling: single Lambert key+fill with hardcoded directions, no AO/curvature/slope texture, so all relief reads as uniform matte clay; 'paddy toning' mixes at most +4% luminance (invisible); the 512px country mask is LinearFilter with no mipmaps so the edge is a soft wash that shimmers during motion. Fix: AO appro
12 | No terrain skirt: bare PlaneGeometry with no side walls; at maxPolarAngle 1.42 rad low-angle views reveal the zero-thickness sheet edge with clear colour behind it (diorama tell). Fix: skirt ring extruded to y = -30 km in sea colour (robust) and/or clamp maxPolarAngle to ~1.25 rad. web/src/scene/Terrain.tsx:142-149, web/src/scen
13 | Two backdrop-filter: blur(14px) surfaces (panel and station sheet) composite over the WebGL canvas while the brief allows one; per-frame backdrop blur over canvas is a known mid-range-phone jank source. Fix: keep blur on the panel only, replace the station-sheet blur with the pre-darkened veil (panelVeil 0.88 alpha passes AA per
14 | Depth-buffer ratio ~12000:1 from camera near=1 / far=12000 while minDistance is 70 km: 99% of depth precision wasted, causing line-vs-terrain z-fighting shimmer exactly where the lifted lines are marginal. Fix: near ~ 10 (or logarithmicDepthBuffer, small mobile cost) and re-verify. web/src/scene/LookdevScene.tsx:254 vs :198
15 | Dead destructure: the loading effect's Promise.all has four elements but destructures five ([fixture, heightfield, borders, rivers, provinces]); provinces is permanently undefined, stored into dead state. One-line fix (or wire it properly if provinces topo should reach the heightfield loader). web/src/routes/Lookdev.tsx:80-89

## Honest scope note
Phase 3's gate evidence-integrity findings (lead review 2026-10-30 01:20) still stand and are in
the lead notes for this phase: ADR 0005's numbers must be corrected from the committed JSON, a
dated critique artifact must exist, and no computed-evidence claim may be cited without a
committed script or JSON. The old relief code is removed from src and lives in git history
(record this in ADR 0007).
