# Gate: Phase 3b — the relief rebuild on MapLibre

- **Date:** 2026-10-30 (re-gate), remediated 2026-10-30 afternoon (phase-3b-fix, the lead's 25-finding review — section at the end)
- **Method:** two full `/impeccable critique` iterations (dual-agent each: isolated design review + isolated detector/browser evidence). Iteration 1 artifact: `web/.impeccable/critique/2026-10-30T12-16-53+07-00__src-routes-lookdev-tsx.md` — **27/40** (the originally reported 28 was an arithmetic slip; corrected everywhere it propagated). Iteration 2, against the remediated build: `web/.impeccable/critique/2026-10-30T07-03-19Z__src-routes-lookdev-tsx.md` — 31/40, two P1s deferred to Phase 4 with reasons. Root-cause fixes for every priority finding, re-measurement of each fix. Every computed number below cites `docs/design/screens/phase-3b/measurements.json`, which regenerates from the committed script `e2e/measure/phase-3b.mjs` against `pnpm exec vite preview`. Evidence: `docs/design/screens/phase-3b/` — the 8 gate shots and 5 gauge-encoding shots are all taken by the committed spec `e2e/tests/gate-phase3b.spec.ts` (re-shot for real during remediation, after the first gate falsely claimed re-shoots it had not made; see the remediation section), the e2e suite (4 gate-evidence + 4 variant tests on desktop; the behaviour variants gained the extended timeout the relief tests already needed), the web unit tests (35 passed), the detector (clean, exit 0).
- **This session cannot view images** (unchanged since Phase 3). Every visual claim below is grounded in a computed measurement; the lead's visual pass on real hardware — including the poles-vs-discs pick the brief schedules for midday — is the confirming step this gate cannot substitute.

## What changed since GATE-phase-3.md

The engine is MapLibre GL 6.16.0 (ADR 0007): globe projection, our terrarium DEM as terrain+hillshade, the self-built Protomaps dark basemap retinted into the token world, our rivers/borders, one continuous night fog from the sky block, and the 804 gauges as a custom instanced WebGL layer (poles) with a circle-layer fallback (discs) behind `?gauges=discs`. The DOM contracts the e2e relies on are unchanged, which is why the Phase 3 evidence approach carried over.

## Scorecard

| Check | Verdict | Evidence |
|---|---|---|
| Critique run (dual-agent, isolated) | done | Two iterations; 27/40 then 31/40, zero banned items, 5 P1s then 2 (deferred, named) |
| Every P1 fixed and re-measured | done | Each fix's before/after number below |
| Zero banned items | **pass** | Item walk below |
| Every requirement findable in seconds | **pass** | Walk below; deferrals named, not hidden |

## The P1 fixes, with their measurements

| Finding (critique) | Fix | Before → after (measured) |
|---|---|---|
| Landing frame a red slab | blade screen width clamped ≤36 px | 24.45% → **6.72%** of canvas exact `#ee5d58` (`measurements.json` → `landingDesktop.alertRedPct`) |
| Phone stacking collapse (3 surfaces, 100% coverage, no dismiss) | a chosen station owns the phone: panel + dev chrome hidden, sheet gains an authored close | 0% relief → sheet 432–648 of 844; close returns the panel (DOM-verified) |
| Protomaps grey takeover at close zoom | retint extended to earth/landcover/buildings/roads, token-derived | 77.7% → **0%** Protomaps grey at the landing view (`measurements.json` → `landingDesktop.protomapsGreyPct`) |
| First-frame poles invisible (phone) | min width 3.5 px, fog starts at 0.55 of horizon | desktop national level-3/4/5 px: 513/633/72 → **742/1069/125** (`measurements.json` → `nationalDesktop`); phone-country poles measure zero exact-token pixels — the open question below |
| Keyboard focus invisible on the canvas | `:focus-visible` outline in the accent token | computed `outline-style: none` → visible ring |
| Hover contrast 4.44:1 on the ใกล้ฉัน note | note lifts to text-secondary on hover | 4.44 → **6.51:1** against the hover background (`measurements.json` → `contrast.nearmeNoteHover`) |
| Tap targets under floor | nearme 44 px, presets taller on phone, close ≥24 px | nearme 156×44, station-close 28×42, linehint-close 28×28 (`measurements.json` → `tapTargets`) |

## The anti-slop checklist, item by item

All Phase 3 items that live in the DOM chrome carry over unchanged (panel, sheet, type, copy — verified again by the detector and the critique's anti-slop sweep). The items this phase could have broken:

| Item | Verdict | Evidence |
|---|---|---|
| Gradient text / glass as decoration | pass | Exactly **one** functional blur (the answer panel's); the station sheet's own blur was removed in remediation — the design brief allows one, and the first gate's checklist wrongly scored two as a pass. The only gradient remains the skeleton sweep |
| Hairline border + wide shadow | pass | Zero `box-shadow` in the route CSS |
| Neon accents on dark | pass | Accents unchanged; the close zoom now shows token-derived surfaces, not the default dark map |
| Hundreds of identical glowing pillars | pass (one caveat) | 804 level-encoded blades, slim by the new clamp; one blade at landing zoom is 6.72% of the frame, not a slab; the phone-country floor is the open question below |
| Spinning globe / starfield / bloom | pass | Globe is static; no postprocessing; the atmosphere blend is the sky block's own |
| Emojis or Unicode glyphs as icons | pass | Both close controls are authored SVG strokes (the critique caught the first ✕; fixed) |
| Safety claims / flood prediction | pass | Grep-verified across the new copy (LINE hint, close labels); no banned water words |
| 3D blocking first paint | pass | Lazy chunk behind the labelled skeleton; `data-scene-ready` gate; e2e green |

## Requirements findable in seconds

- **Colour from the tokens, single source:** the style JSON, the gauge shader uniforms and the retint all read `tokens.ts` (the retint's mix ratios are the only hand-set numbers, like the shader's Lambert weights). The landing view samples back 0.03% foreign grey.
- **One continuous fog — with a disclosed limitation:** the sky block and the gauge shader share `relief.fog`, and the gauges' fog starts at 0.55 of the horizon distance so the country view keeps its colour. But **while the globe projection is active, MapLibre 6 ignores the style sky block's fog entirely** — the console warns verbatim `calculateFogMatrix is not supported on globe projection.` The terrain's aerial fade at the horizon therefore comes from the atmosphere blend, and the *gauges* are the only elements that carry our fog token explicitly (in the shader). The sky block still governs the horizon colour itself. Disclosed in ADR 0007's addendum and the phase report; if the lead wants true terrain fog on globe, the options are a mercator close-up or waiting on MapLibre upstream.
- **The encoding, geometric before coloured:** cap 0.78 + fill + overflow mapped onto CAP..1 + hollow stale + translucent suspect — mirrored in poles, discs and the DOM legend; suspect count (2) reaches the shader from the fixture.
- **The brief's ported bug fixes:** uploads keyed on inputs not frames (5), live reduced-motion listener (7), overflow clamped by construction (9), slim profile + darker base ring (10), suspect never fully trusted (8). Terrain-engine bugs (1, 2, 4, 11, 12, 14) died with MapLibre; the dead destructure (15) died with the route rewrite.
- **Interaction port:** presets jumpTo, the authored flyTo with curve, reduced-motion cuts (e2e), ใกล้ฉัน to the sample station, LINE webview offer with `?openExternalBrowser=1` (e2e-verified by `e2e/tests/variants.spec.ts` → 'the LINE webview gets the open-externally offer': link, dismiss, and never in a plain context — the test was written during remediation; the first gate's "e2e-verified" claim predates it and was false, recorded in the DEVLOG correction).
- **Budgets** (all from `measurements.json` → `bundleGzipKb`, gzip, from the committed build): initial shell **73.1 kB** (budget <300 kB ✓); lazy relief chunk **298 kB** + MapLibre worker **144.9 kB** = ~443 kB, against the ~250 kB the R3F stack weighed (ADR 0004) — heavier, recorded honestly; the engine swap bought the shaded-relief/fog/DPR stack that the hand-built scene could not reach.
- **Provenance:** ODbL/OSM + Protomaps credited on the page and in `geodata/SOURCES.md` (basemap + DEM are gitignored locals with recorded build commands).

## Open questions for the lead (named, not hidden)

1. **Poles vs discs** — the brief's midday pick. Measured: poles read beautifully from the plain view up (and at desktop national after the fixes) but are below the perceptual floor at phone-country zoom; discs register at every zoom but carry less encoding (no cap/fill geometry). Both ship: `/lookdev` vs `/lookdev?gauges=discs`.
2. **Basemap Thai labels** — the tiles carry `name:th`; no glyph set is served (fonts.protomaps.com is dead). Self-hosting a Thai glyph PBF set is the follow-up; the page credits OSM regardless.
3. **The exaggeration default (1.5×)** — MapLibre's terrain exaggeration is a different scale than the old rig's 4–40×; the slider spans 1–40× in the new units. The provisional 1.5× default is now recorded in ADR 0008; the lead confirms or changes it on hardware.
4. **The desktop national frame crops Thailand's tips** (critique iteration 2): at 1440×900 the northernmost land projects to y = −7 and Betong to y = 923 — outside the viewport despite `fitBounds` padding 72; the phone frame is clean. Likely a globe-projection/fitBounds interaction; Phase 4 with an e2e assertion on the four extreme land points.
5. **ใกล้ฉัน is a sample station** (critique iteration 2): the flight goes to fixture station 736 regardless of location, disclosed only by the button's 12 px note. Real geolocation is job J2 in Phase 4; until then the disclosure belongs in the sheet head, not the button.

## Remediation (phase-3b-fix, 2026-10-30 ~13:00–15:00)

The lead's adversarial review of the first gate found 25 findings (2 blockers, 6 majors, 17 minors). All 25 are addressed; the binding prompt is `process/prompts/phase-3b-fix.md`.

**Blocker 1 — the false re-shoot claim.** The first gate said "re-shoot of every screenshot" but the gauge-encoding shots were stale files from before the width-clamp fix. Every shot in `docs/design/screens/phase-3b/` is now taken by the committed spec `e2e/tests/gate-phase3b.spec.ts` and was re-shot for real during remediation (run at ~13:30 and again at ~14:00 after the copy fixes, all 4 tests green) — this paragraph is the correction of the record, per the truthfulness rule.

**Blocker 2 — two backdrop blurs.** The station sheet's blur is removed (opaque panel colour + border); the panel keeps the one blur the design brief allows. The checklist above now scores it correctly.

**Majors:** stale poles were exactly invisible (the `atCap` gate's dead `vStale < 0.5` clause — removed, with the stale close-up `gauges-poles-stale-closeup-1440.png` now shot by the spec, which also asserts the cap colour's pixels on the canvas); pre-load preset/exaggeration/flight requests are kept and applied on load; WebGL detection tests webgl2 only and a map-init failure reaches the same fallback message; every number now cites `measurements.json` (this document) regenerated by the committed script; the landcover green and the stale disc stroke come from the tokens; the LINE e2e exists (see the interaction-port row).

**Minors:** the init chain is guarded; flights sit in the brief's 1.2–2.0 s window; @types/three dropped; the skeleton sheen is a token; the critique arithmetic corrected (27, not 28 — the artifact carries a dated correction note); a shader compile failure degrades to the disc layer instead of aborting the frame loop; ADR 0008 records the exaggeration; the fog-on-globe no-op is disclosed above; the DEVLOG carries the outage history and the LINE-claim correction; the second critique iteration ran for real (artifact above).

## Verdict

**Pass, with open questions named.** Zero banned items; every critique finding from both iterations either fixed and re-measured or deferred to Phase 4 with a named reason; the e2e and unit suites green; the evidence committed and regenerable. What this gate cannot certify is the rendered image itself — the lead's visual pass (with the poles-vs-discs decision and the exaggeration confirmation on hardware) is the required next step before Phase 4, exactly as the workflow prescribes.
