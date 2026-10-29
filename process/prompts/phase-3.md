# Phase 3: Look-dev (the visual world, before the product)

Lead notes from the Phase 2 review are at the bottom; address them first.

**Goal:** prove the visual world in isolation before building features. This phase decides whether the project feels premium. Re-read `docs/brief/01-DESIGN-DIRECTION.md` end to end before starting. Then invoke the `impeccable` skill and follow its new-work and craft guidance, with our design brief as the pinned brief.

## Deliverables
1. **`DESIGN.md` and tokens**
   - Use the impeccable document format if it helps.
   - Tokens go in `web/src/design/tokens.css` and `web/src/design/tokens.ts`, with one source and a generated or mirrored second form. Cover colour roles, the situation scale, type scale, spacing, radii, motion durations and easings, and z-layers.
2. **Situation colour scale**
   - Render it as a swatch strip with the official labels.
   - Show a deuteranopia and protanopia simulation as screenshots.
   - Include contrast calculations for text on every surface.
3. **Type specimen page (`/specimen`, dev only)**
   - Test 2–3 shortlisted pairings with real Thai strings: the national sentence, three station sentences, long province names, and numbers in tabular figures.
   - Pick one, and record the choice in ADR `0005-typography.md` with screenshots.
   - Self-host the chosen WOFF2 subsets.
4. **Geodata pipeline (`geodata/`)**
   - Make it reproducible from scripts (Node or Go): DEM tiles, then a stitched heightmap at the chosen resolution, then an encoded asset.
   - Borders and rivers go to compact files.
   - Write `geodata/SOURCES.md` with the licences.
   - Commit only the final assets if they are under budget; the cache is gitignored.
5. **Relief prototype (`/lookdev`, dev only)**
   - The R3F scene with the relief, lighting, material, border line, rivers and atmosphere.
   - Instanced gauges driven by a **static snapshot of real data** from the Phase 2 API, saved as a fixture.
   - Camera presets: national, a central plain close-up and one station close-up, with one sample flight.
   - Tune the vertical exaggeration and record it in ADR `0006-relief-look.md`.
6. **Gate**
   - Take Playwright screenshots into `docs/design/screens/phase-3/`: national view, central plain, station close-up, the panel mock with the sentence, type specimen and colour scale, at 390×844 and 1440×900.
   - Run `/impeccable critique` on them.
   - Write `docs/design/GATE-phase-3.md` with every checklist item marked pass or fail and evidence.
   - Iterate until it passes. **At least two full iterations are expected.** Record what changed and why in each one.

## Taste notes from the lead
- The first view should feel like a museum relief model lit at night, not a video game or a SaaS dashboard.
- Restraint is the luxury: few colours, one light and exact type.
- If a choice looks "cool" but encodes nothing, delete it.
- Thai text must look typeset, not merely rendered. Check the vertical rhythm with stacked marks.

## Report and stop
Write `process/reports/phase-3.md` with the final screenshots embedded as relative links. Stop. The lead must approve the look before Phase 4.
