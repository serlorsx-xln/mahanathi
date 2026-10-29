# Gate: Phase 3 — look-dev

- **Date:** 2026-10-30
- **Method:** two full iterations of `/impeccable critique` (dual-agent: a design-review assessment and an isolated detector + browser assessment), root-cause fixes between and after each, re-shoots of every screenshot, detector re-scan clean on all three served routes. Evidence: the screenshots and JSON in `docs/design/screens/phase-3/`, the e2e suite (22 passed, 8 project-skips by design), the web unit tests (51 passed), and the computed numbers cited below.
- **This session cannot view images** (DEVLOG 2026-10-29 23:25). Every visual claim below is grounded in a computed measurement, a test, or source; the lead's visual pass on real hardware is the named remaining step, as the workflow already requires.

## Scorecard

| Check | Verdict | Evidence |
|---|---|---|
| Iteration 1 run | done | Critique snapshot `web/.impeccable/critique/2026-10-29T17-47-03Z__src-app-tx.md`; findings and fixes in DEVLOG ~01:00; commit `049080a` |
| Iteration 2 run | done | Fresh assessment (31/40, zero banned items); fixes in DEVLOG ~01:15; commit `d72b03e` |
| Zero banned items | **pass** | Every item walked below, each with its evidence line |
| Every requirement findable in seconds | **pass** | The walk below; the two deferred requirements are named, not hidden |

## The anti-slop checklist (brief §10), item by item

### Page and layout tells
| Item | Verdict | Evidence |
|---|---|---|
| Hero section with pill badge / eyebrow label | pass | No hero; the first frame is the relief, the sentence and one button (`Lookdev.tsx` composition) |
| Identical icon-heading-text card grids | pass | No cards; /specimen and /color are one-column ruled sections |
| Hero-metric template (big number, small label, stats row) | pass | The panel is sentence-first; numbers are a subordinate live-count `<dl>` |
| Section numbers 01/02/03 | pass | None; sections use Thai headings |
| Nested cards | pass | None |
| Monotonous spacing | pass | Spacing steps by role (chrome space-4/5, panel space-5, sheet space-4/5) |
| Everything centred | pass | Panel left-anchored, sheet corner-anchored, Thailand composes off-centre; only the transient error card centres |
| Decorative grid backgrounds | pass | None; the only gradient is the skeleton's functional sweep (brief §9 allows skeletons) |
| Modals for tasks that need no interruption | pass | No modals |

### Surface tells
| Item | Verdict | Evidence |
|---|---|---|
| Gradient text | pass | None |
| Glass and blur as decoration | pass | Exactly two `backdrop-filter: blur(14px)` — the panel and the station sheet, both for legibility over terrain (the brief's one allowed functional blur); dev chrome on solid panel since iteration 1 |
| Hairline border plus wide shadow | pass | Zero `box-shadow` declarations in shipped route CSS; iteration 1 dropped the panel shadow |
| Thick coloured side-stripes on cards | pass | None |
| Extreme border radius | pass | Max is the 12 px sheet; `radius-full` confined to swatch dots |
| Glow halos and spotlights | pass | None; no postprocessing at all (ADR 0006: flat pipeline by design) |
| Neon accents on dark | pass | Accents are muted teal `#7fb5b0` and ochre `#caa46a`; only levels 4/5 are saturated, both AA as text |
| Zero-blur offset "brutalist" shadows | pass | None |
| Sparklines / progress rings standing in for content | pass | None |

### Type tells
| Item | Verdict | Evidence |
|---|---|---|
| Inter, Geist or Roboto as the voice | pass | Mahanathi Sans (IBM Plex Sans Thai, OFL-renamed) + Noto Serif Thai (ADR 0005); self-hosted WOFF2, `font-display: swap` |
| Italic serif display | pass | Every `@font-face` is `font-style: normal` |
| Crushed tracking | pass | Zero `letter-spacing` declarations |
| Flat hierarchy | pass | Eight-step scale with real gaps; display caps at 4 rem (brief ceiling 6 rem) |
| One family doing everything | pass | Two families, distinct roles, 400/600 weights |

### Copy tells
| Item | Verdict | Evidence |
|---|---|---|
| Generic marketing claims | pass | Plain factual Thai; measured time and source always stated |
| Em-dash habit | pass | None in user-visible copy |
| "Not X. Y." slogans | pass | None |
| Invented numbers or testimonials | pass | Every number traces to the committed 804-station fixture; no testimonials exist |
| Banned water words (ปลอดภัย / จะท่วม / ไม่ท่วม / prediction) | pass | Grep-verified across web copy; generated sentences are governed by the humanize package's stricter banned-phrase test |

### 3D tells
| Item | Verdict | Evidence |
|---|---|---|
| Spinning / auto-rotating globe | pass | Camera moves only on input, presets, the one flight |
| Starfield, particles, floating dust | pass | None |
| Bloom and glow on everything | pass | `flat` Canvas, explicit output-encoding chunks in both custom shaders |
| Synthwave neon wireframe / grid floor | pass | Displaced hypsometrically tinted mesh |
| Arcs flying between points | pass | Only the border and 13 real rivers |
| Hex-bin globe | pass | N/a by construction |
| Hundreds of identical glowing pillars | pass | 804 gauges, level-based heights (44/34/17/12 km), bank-full cap, in-slot fill, overflow-above-cap, hollow stale — geometric encoding first |
| Free-spin controls that lose the horizon | pass | Polar 0.12–1.42 rad, distance 70–2800 km, target bounded to Thailand (iteration 1) |
| Default MeshStandardMaterial under stock HDRI | pass | Custom ShaderMaterial, NW key + soft fill (Imhof convention) |
| Stretched low-resolution textures | pass | 1815×2048 heightmap ≈ 1 km/texel; drapes sample real elevations |
| Jaggies or z-fighting | pass | `antialias: true`; exaggeration-scaled drape lift; e2e asserts zero console errors |
| 3D blocking first paint / phone jank | pass (caveat) | Lazy chunk behind a labelled skeleton, `frameloop="demand"`, gate waits on `data-scene-ready`; dpr is a `[1,2]` clamp, not adaptive — deferred, below |
| Camera wobble / parallax unrelated to data | pass | None |
| Broken Thai shaping in 3D text | pass | No Thai enters WebGL; the station label is a DOM `Html` projection |

## Requirements findable in seconds

- **Colour from the subject, single source:** `tokens.ts` → generated `tokens.css` with a drift test; the national screenshot samples back sea `#0d1a23` (34%) and ground `#0e161d` (21%) exactly.
- **Ordered situation scale, CVD evidence:** `scripts/research/cvd_scale_search.py`; min pairwise ΔE deutan 14.1 / protan 15.0 / normal 29.9; replayed live at `/color`; AA on every surface tested in `tokens.test.ts`.
- **Typography per §6:** ADR 0005 with `specimen-measurements.json`; body 1.6, sentence 1.65, tabular figures throughout; 64 KB of self-hosted fonts.
- **Relief per §4:** ADR 0006 — 12× default, slider 4–40× for the lead's pass; NW key; neighbour desaturation; aerial fog 900–2600 km; exaggeration evidence `lookdev-exagg6/30-*` (gradient energy 0.0149 / 0.0168 / 0.0193 at 6×/12×/30×).
- **Camera per §4:** oblique default, clamped orbit, target bounded, 1600 ms exponential ease-out interruptible flights, reduced motion cuts (e2e-tested).
- **Motion per §7:** one authored moment (ใกล้ฉัน flight, honestly annotated "· ตัวอย่างสถานี" until geolocation lands in Phase 4); UI transitions 160–260 ms.
- **States per §9:** loading skeleton with screen-reader status (resolves — regression caught and fixed in iteration 2); no-WebGL floor with the shell, sentence and data (e2e-tested); fetch failure names the problem and offers a working retry (e2e-tested through a real abort); reduced-motion cuts (e2e-tested).
- **Browser surfaces themed (§9):** selection, focus ring, caret, scrollbar, underline offset, `color-scheme: dark` in `base.css` — which iteration 1 discovered had never been imported; it now loads on every route.
- **Phone composition:** measured at 390×844 (`lookdev-mobile-coverage.json`): national panel 34.8% of viewport height, top chrome 8.6%, dev bar 12.7% — the relief keeps ~44% clear. Panel luminance AA over the brightest terrain: textPrimary 12.3:1, textSecondary 6.8:1, textFaint 4.7:1.

## Deferred to Phase 4, named (not hidden)

1. **Gauge LOD labels** (brief §4 "as you zoom in, gauges grow and gain labels"): country-zoom hierarchy is done; zoom-in labels are Phase 4 with the search/selection work they belong to.
2. **The phone bottom sheet with snap points** (brief §3): the look-dev panel is a top sheet; measured coverage above shows the relief keeps ~44% clear, and the real sheet lands with the product shell.
3. **The 72-hour chart against bank-full** in the answer panel (brief §3): Phase 4, on the timeline data path.
4. **Keyboard/screen-reader station access** (the brief's search field): Phase 4; the canvas is labelled and status changes announce.
5. **Adaptive DPR** (brief §4): the `[1,2]` clamp holds at this scene's weight; revisit with the perf pass.
6. **Relief shifts right on desktop when the panel appears** (brief §3): Phase 4 composition work with the real left-hand column.
7. **The exaggeration slider's drape recompute per input event** — a dev control that does not ship; noted for the Phase 4 perf pass if the pattern survives.

## Verdict

**Pass.** Zero banned items across all five checklist groups; every requirement either evidenced above or named in the deferral list with its reason. Two full iterations ran, each finding and fixing real defects (an unimported stylesheet serving grey-on-white on the contrast-evidence pages; an infinite loading announcement on the no-WebGL floor). The one thing this gate cannot certify is the rendered image itself — the lead's visual pass on real hardware, with ADR 0006's slider in hand, is the confirming step the workflow requires before Phase 4.
