# Phase 3 report: Look-dev (the visual world, before the product)

- **Implementer:** the same session as Phase 2 (GLM-5.3 via the owner's router; quality gates unchanged and, if anything, leaned on harder).
- **Phase window:** 2026-10-29 ~22:36 ICT (Step 0 fixes, `000b801`) to 2026-10-30 ~01:20 ICT.
- **Tree at report time:** clean at `900c020` (the report itself is `8a6ed5c`; `900c020` after it is only the Prettier pass over the phase's edits, no behaviour change).

## Summary

The visual world is built, evidenced and gated. Every colour lives in one tokens module that round-trips to the screen verbatim (the national screenshot samples the sea token back at 34% of the frame); the situation scale was found by constrained search under colour-vision-deficiency simulation; the type pairing was decided by measurement, not taste; the relief renders Thailand and all 804 real gauges with a geometric bank-full encoding; and the gate ran two full critique iterations, each of which caught and fixed real defects — including one that no build or unit test could see (the two custom shaders never compiled until the e2e console assertion caught it, and the contrast-evidence pages were serving grey-on-white because a stylesheet was never imported).

**This session cannot view images** (recorded in the DEVLOG at Phase 3 start). The gate therefore rests on computed evidence — pixel sampling against the tokens, canvas ink metrics, DOM measurements, CVD mathematics, deterministic tests — and the report names the lead's visual pass on real hardware as the remaining confirming step, which the workflow requires anyway before Phase 4.

## What was built

| Commit | Piece |
|---|---|
| `dd30cf9` … `000b801` | Step 0: the five Phase 2 review findings, each in its own commit |
| `75389cf` | geodata shipped: four assets, build scripts, SOURCES.md, byte-reproducible |
| `43a4ace` | design tokens: one TS source, generated CSS mirror, AA tested |
| `f6386ba` | specimen + colour evidence pages; fonts self-hosted; situation scale rebuilt by CVD search |
| `016f56f` | the type decision (ADR 0005): IBM Plex Sans Thai → "Mahanathi Sans" + Noto Serif Thai |
| `7c0e65b` | the `/lookdev` scene: terrain, gauges, drapes, rig, flights; PNG decode tested |
| `049080a` | gate iteration 1: critique root causes fixed, detector clean |
| `d72b03e` | gate iteration 2: a caught regression fixed, stale keyed, copy unified |
| `2409860` | DESIGN.md + sidecar recorded from the shipped artifact |
| `6fb793e` | ADR 0006 (relief look) + the scored gate document |

## The deliverables, against the prompt

1. **DESIGN.md and tokens** — `web/DESIGN.md` (impeccable format, frontmatter + sidecar `web/.impeccable/design.json`), derived from the shipped artifact with zero token drift. `tokens.ts` is the single source; `tokens.css` is generated (`pnpm tokens:css`) and a test fails on drift. Colour roles, the situation scale, type scale, spacing, radii, motion, z-layers — all covered.
2. **Situation colour scale** — `/color` renders the swatch strip with official labels, live deuteranopia/protanopia simulation (canonical Machado matrices), and computed contrast for text on every surface. The scale itself is the output of `scripts/research/cvd_scale_search.py`: min pairwise ΔE deutan 14.1 / protan 15.0 / normal 29.9. Screenshots: `color-{desktop,mobile,tablet}.png`.
3. **Type specimen (`/specimen`)** — three pairings × four sentence faces with real Phase 2 strings; decided by measurement (`specimen-measurements.json`): Plex's minus equals its plus where Noto Sans Thai's does not; Anuphan's ink rises above its own ascender. ADR `0005-typography.md`. Self-hosted WOFF2 subsets, 64 KB total, RFN rename verified.
4. **Geodata pipeline** — committed in Phase 3's first half (`75389cf`): `build_heightmap.py` and `build_vectors.py` rebuild all four assets byte-identically; `geodata/SOURCES.md` records licences; 1.46 MB total against the 1.5 MB budget.
5. **Relief prototype (`/lookdev`)** — R3F scene driven by the real Phase 2 fixture (804 stations): custom-shader terrain (NW key, hypsometric tint, paddy toning, neighbour desaturation, aerial fog), 804 instanced gauges with the bank-full cap encoding, draped borders and rivers, clamped camera with three presets and one interruptible 1.6 s flight, reduced motion cuts, no-WebGL floor, error state with working retry. Vertical exaggeration decided at 12× in ADR `0006-relief-look.md` with the slider left in for the lead's pass.
6. **Gate** — `docs/design/GATE-phase-3.md`: every checklist item scored pass with its evidence line, seven deferred requirements named rather than hidden, two full iterations recorded (below).

## The gate's two iterations

**Iteration 1** (dual-agent critique: design review + isolated detector/browser). The detector found 29 real contrast failures on the served `/specimen` and `/color` — root cause: `base.css` was imported nowhere, so the pages that prove AA contrast were the only pages failing it, on browser-default white. The design review found the unhandled fetch failure (infinite silent skeleton), the mislabelled ใกล้ฉัน, the unbounded camera target, and the border-plus-wide-shadow / blur-on-five-elements tells. All fixed at the root (`049080a`); detector re-scan: zero findings on all three routes.

**Iteration 2** (fresh assessment). Re-scored 31/40, zero banned items — and caught a regression iteration 1 had introduced: the skeleton's new `role="status"` announced "กำลังโหลด" forever on the no-WebGL floor because the branch gated on `!failed` instead of `!data`. Fixed, with the e2e assertion that would have caught it (`d72b03e`). Also: the panel's numbers row gained a hollow "ยังไม่อัปเดต" entry keyed to the hollow 3D gauges, missing-value markers unified, "% ของตลิ่ง" unified, `aria-pressed` on presets.

**Verdict: pass** — zero banned items, every requirement evidenced or named in the deferral list (gauge zoom labels, phone bottom sheet, 72-hour chart, keyboard station access, adaptive DPR, desktop panel shift, slider drape recompute — all Phase 4, all with reasons).

## Evidence index (all paths relative to the repo root)

The gate viewports are 390×844, 768×1024 and 1440×900, plus the reduced-motion and no-WebGL variants, all in `docs/design/screens/phase-3/`:

- National view: [desktop](../design/screens/phase-3/lookdev-national-desktop.png) · [tablet](../design/screens/phase-3/lookdev-national-tablet.png) · [mobile](../design/screens/phase-3/lookdev-national-mobile.png)
- Central plain: [desktop](../design/screens/phase-3/lookdev-plain-desktop.png) · [tablet](../design/screens/phase-3/lookdev-plain-tablet.png) · [mobile](../design/screens/phase-3/lookdev-plain-mobile.png)
- Station close-up: [desktop](../design/screens/phase-3/lookdev-station-desktop.png) · [tablet](../design/screens/phase-3/lookdev-station-tablet.png) · [mobile](../design/screens/phase-3/lookdev-station-mobile.png)
- The panel with the sentence: in every national/plain/station shot above; measured mobile coverage in [`lookdev-mobile-coverage.json`](../design/screens/phase-3/lookdev-mobile-coverage.json) (panel 34.8% of the viewport, relief keeps ~44% clear)
- The flight: [mid-flight frame (CDP screencast, motion asserted)](../design/screens/phase-3/lookdev-flight-mid-desktop.png) · [landed](../design/screens/phase-3/lookdev-flight-end-desktop.png)
- Exaggeration for ADR 0006: [6×](../design/screens/phase-3/lookdev-exagg6-desktop.png) · [30×](../design/screens/phase-3/lookdev-exagg30-desktop.png) (12× is the plain shot)
- Variants: [reduced motion](../design/screens/phase-3/lookdev-reduced-motion-desktop.png) · [no WebGL](../design/screens/phase-3/lookdev-no-webgl-desktop.png)
- Type specimen: [desktop](../design/screens/phase-3/specimen-desktop.png) · [tablet](../design/screens/phase-3/specimen-tablet.png) · [mobile](../design/screens/phase-3/specimen-mobile.png) · measurements in [`specimen-measurements.json`](../design/screens/phase-3/specimen-measurements.json)
- Colour scale + CVD simulation: [desktop](../design/screens/phase-3/color-desktop.png) · [tablet](../design/screens/phase-3/color-tablet.png) · [mobile](../design/screens/phase-3/color-mobile.png)

## Verification

- **Web:** `tsc -b` clean, `eslint --max-warnings 0` clean, **51 unit tests pass** (tokens AA, generated-mirror drift, PNG decode against a Pillow oracle, projection, the shipped heightmap's geography, SSR).
- **e2e:** **22 passed, 8 skipped** (project-scoped skips by design: variants are desktop-only, measurements desktop-only) across desktop 1440×900, mobile 390×844 and tablet 768×1024. Every relief spec asserts zero console errors; the flight test asserts streamed screencast frames differ (motion itself is tested); the error path is tested through a real aborted fetch; reduced motion and no-WebGL are tested as gate variants.
- **Detector:** `impeccable detect` on the served build — zero findings on `/lookdev`, `/specimen`, `/color` (verified against a self-test probe that fired correctly).
- **Computed colour evidence:** the national screenshot samples the sea token `#0d1a23` (34%) and ground `#0e161d` (21%) exactly; panel veil over the brightest terrain blends to `#212a32` with textPrimary 12.3:1, textSecondary 6.8:1, textFaint 4.7:1; exaggeration gradient energy 0.0149 / 0.0168 / 0.0193 at 6×/12×/30×.
- **Budgets:** lazy scene chunk 167.76 KB gzipped (ADR 0004 budget ~250 KB); fonts 64 KB; geodata 1.46 MB.

## Mistakes caught and recorded (DEVLOG)

All with timestamps in `docs/DEVLOG.md`: the two shaders that never compiled (undeclared varying, uninterpolated constant — caught only by the console assertion); the linear-vs-sRGB output encoding that rendered every token dark (caught by pixel-sampling the screenshots); the too-early exaggeration screenshots and the uncapturable mid-flight frame (both turned into deterministic methods: `data-scene-ready` and CDP screencast); `base.css` never imported (29 contrast findings, caught by the detector); and the iteration-1 regression that announced loading forever on the no-WebGL floor (caught by iteration 2).

## Open questions for the lead

1. **The rendered image itself.** This session cannot view it. The highest-value check is the national frame on real hardware in a bright room: does it read as a plaster relief lit at night, or as a dark void with dots? The two knobs are in hand — the moonlit Lambert floor in `Terrain.tsx` and the exaggeration slider on `/lookdev` (ADR 0006 defaults to 12×).
2. **The fog.** At the national zoom (camera ≈ 2,025 km) the 900–2600 km fog dissolves the far terrain 62–85% toward the sea colour. If the lead's pass finds the country's edges too absent, `uFogNear/uFogFar` are the numbers.
3. **Panel over relief.** Iteration 2's provocative question stands: if the panel needs an 88% veil plus 14 px blur, would raising the terrain's lit floor a step and reserving the blur be the stronger world?
4. **The deferral list** (gate doc, seven items) — confirm or reorder for Phase 4.

Phase 4 does not start until the lead's prompt. No pushes, no deploys — those are the lead's alone.
