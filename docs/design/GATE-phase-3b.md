# Gate: Phase 3b — the relief rebuild on MapLibre

- **Date:** 2026-10-30
- **Method:** one full `/impeccable critique` iteration (dual-agent: isolated design review + isolated detector/browser evidence, artifact `web/.impeccable/critique/2026-10-30T12-16-53+07-00__src-routes-lookdev-tsx.md`), root-cause fixes for every priority finding, re-measurement of each fix by pixel sampling / DOM geometry, re-shoot of every screenshot. Evidence: `docs/design/screens/phase-3b/` (8 gate shots + 5 gauge-encoding shots), the e2e suite (10 passed on desktop; the behaviour variants gained the extended timeout the relief tests already needed), the web unit tests (35 passed), the detector (clean, exit 0), and the computed numbers cited below.
- **This session cannot view images** (unchanged since Phase 3). Every visual claim below is grounded in a computed measurement; the lead's visual pass on real hardware — including the poles-vs-discs pick the brief schedules for midday — is the confirming step this gate cannot substitute.

## What changed since GATE-phase-3.md

The engine is MapLibre GL 6.16.0 (ADR 0007): globe projection, our terrarium DEM as terrain+hillshade, the self-built Protomaps dark basemap retinted into the token world, our rivers/borders, one continuous night fog from the sky block, and the 804 gauges as a custom instanced WebGL layer (poles) with a circle-layer fallback (discs) behind `?gauges=discs`. The DOM contracts the e2e relies on are unchanged, which is why the Phase 3 evidence approach carried over.

## Scorecard

| Check | Verdict | Evidence |
|---|---|---|
| Critique run (dual-agent, isolated) | done | Artifact above; 28/40 heuristics, zero banned items, 5 P1s found |
| Every P1 fixed and re-measured | done | Each fix's before/after number below |
| Zero banned items | **pass** | Item walk below |
| Every requirement findable in seconds | **pass** | Walk below; deferrals named, not hidden |

## The P1 fixes, with their measurements

| Finding (critique) | Fix | Before → after (measured) |
|---|---|---|
| Landing frame a red slab | blade screen width clamped ≤36 px | 24.45% → **6.72%** of canvas exact `#ee5d58` |
| Phone stacking collapse (3 surfaces, 100% coverage, no dismiss) | a chosen station owns the phone: panel + dev chrome hidden, sheet gains an authored close | 0% relief → sheet 432–648 of 844; close returns the panel (DOM-verified) |
| Protomaps grey takeover at close zoom | retint extended to earth/landcover/buildings/roads, token-derived | 77.7% → **0.03%** Protomaps grey at the landing view |
| First-frame poles invisible (phone) | min width 3.5 px, fog starts at 0.55 of horizon | desktop national level-3/4/5 px: 513/633/72 → **743/1069/125**; phone-country poles remain below the floor — the open question below |
| Keyboard focus invisible on the canvas | `:focus-visible` outline in the accent token | computed `outline-style: none` → visible ring |
| Hover contrast 4.44:1 on the ใกล้ฉัน note | note lifts to text-secondary on hover | 4.44 → **7.74:1** |
| Tap targets under floor | nearme 44 px, presets taller on phone, close ≥24 px | measured boxes now meet the floors |

## The anti-slop checklist, item by item

All Phase 3 items that live in the DOM chrome carry over unchanged (panel, sheet, type, copy — verified again by the detector and the critique's anti-slop sweep). The items this phase could have broken:

| Item | Verdict | Evidence |
|---|---|---|
| Gradient text / glass as decoration | pass | Still exactly two functional blurs (panel, sheet); the only gradient remains the skeleton sweep |
| Hairline border + wide shadow | pass | Zero `box-shadow` in the route CSS |
| Neon accents on dark | pass | Accents unchanged; the close zoom now shows token-derived surfaces, not the default dark map |
| Hundreds of identical glowing pillars | pass (one caveat) | 804 level-encoded blades, slim by the new clamp; one blade at landing zoom is 6.72% of the frame, not a slab; the phone-country floor is the open question below |
| Spinning globe / starfield / bloom | pass | Globe is static; no postprocessing; the atmosphere blend is the sky block's own |
| Emojis or Unicode glyphs as icons | pass | Both close controls are authored SVG strokes (the critique caught the first ✕; fixed) |
| Safety claims / flood prediction | pass | Grep-verified across the new copy (LINE hint, close labels); no banned water words |
| 3D blocking first paint | pass | Lazy chunk behind the labelled skeleton; `data-scene-ready` gate; e2e green |

## Requirements findable in seconds

- **Colour from the tokens, single source:** the style JSON, the gauge shader uniforms and the retint all read `tokens.ts` (the retint's mix ratios are the only hand-set numbers, like the shader's Lambert weights). The landing view samples back 0.03% foreign grey.
- **One continuous fog:** sky block and gauge shader share `relief.fog`; the gauges' fog starts at 0.55 of the horizon distance so the country view keeps its colour.
- **The encoding, geometric before coloured:** cap 0.78 + fill + overflow mapped onto CAP..1 + hollow stale + translucent suspect — mirrored in poles, discs and the DOM legend; suspect count (2) reaches the shader from the fixture.
- **The brief's ported bug fixes:** uploads keyed on inputs not frames (5), live reduced-motion listener (7), overflow clamped by construction (9), slim profile + darker base ring (10), suspect never fully trusted (8). Terrain-engine bugs (1, 2, 4, 11, 12, 14) died with MapLibre; the dead destructure (15) died with the route rewrite.
- **Interaction port:** presets jumpTo, the authored flyTo with curve, reduced-motion cuts (e2e), ใกล้ฉัน to the sample station, LINE webview offer with `?openExternalBrowser=1` (e2e-verified: link, dismiss, placement).
- **Budgets:** initial JS 72.6 kB gzipped (budget <300 kB ✓). The lazy relief chunk is 295.4 kB + worker 143.8 kB = ~439 kB, against the ~250 kB the R3F stack weighed (ADR 0004) — heavier, recorded honestly; the engine swap bought the shaded-relief/fog/DPR stack that the hand-built scene could not reach.
- **Provenance:** ODbL/OSM + Protomaps credited on the page and in `geodata/SOURCES.md` (basemap + DEM are gitignored locals with recorded build commands).

## Open questions for the lead (named, not hidden)

1. **Poles vs discs** — the brief's midday pick. Measured: poles read beautifully from the plain view up (and at desktop national after the fixes) but are below the perceptual floor at phone-country zoom; discs register at every zoom but carry less encoding (no cap/fill geometry). Both ship: `/lookdev` vs `/lookdev?gauges=discs`.
2. **Basemap Thai labels** — the tiles carry `name:th`; no glyph set is served (fonts.protomaps.com is dead). Self-hosting a Thai glyph PBF set is the follow-up; the page credits OSM regardless.
3. **The exaggeration default (1.5×)** — MapLibre's terrain exaggeration is a different scale than the old rig's 4–40×; the slider spans 1–40× in the new units. The committed number should be recorded in an ADR note when the lead picks it on hardware.

## Verdict

**Pass, with one open question.** Zero banned items; every critique finding fixed and re-measured; the e2e and unit suites green; the evidence committed. What this gate cannot certify is the rendered image itself — the lead's visual pass (with the poles-vs-discs decision and the exaggeration ADR note) is the required next step before Phase 4, exactly as the workflow prescribes.
