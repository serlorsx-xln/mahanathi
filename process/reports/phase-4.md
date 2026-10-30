# Phase 4 report: product surface on the new base (cut at ~15:05 ICT)

- **Implementer:** the same session (GLM-5.3 via the owner's router).
- **Phase window:** 2026-10-30 ~14:28–~15:05 ICT, per the lead's hard-stop note (`process/lead-notes/phase-4.md`). Phase 3b-fix was fully committed and clean when the Phase 4 prompt was read, so this phase started on time.
- **Tree at report time:** `1e85e32` (view modes), `1cf1aea` (loopless), `8b37cb6` (copy rule), `f477249` (DEMO + credit), plus this report's commit.

## What landed (in the lead's priority order)

### P0 — the view modes (commit `1e85e32`)

The segmented pill (`2D MAP` / `3D TERRAIN`, English instrument labels, aria-pressed, keyboard focusable, 44 px tap floor) in the top chrome; state synced to the URL (`?view=2d` / `?view=3d`, deep-linkable) and remembered in localStorage (try/catch). 2D MAP: mercator, no terrain, pitch 0/bearing 0, hillshade softened to 0.25, **the full night fog** — the sky block's fog matrix only works in mercator, so ADR 0007's no-op warning became a designed trade (two token-derived skies; only the blend ratios are hand-set). 3D TERRAIN: the globe, the terrain, the authored pitches. The switch is sequenced exactly as specced — setProjection → idle → easeTo, terrain + `setStations` re-applied after the transition, pill and ใกล้ฉัน disabled for the window, the tile reload masked by a ~180 ms fade through the fog token (a cut under reduced motion). Presets and the flight flatten to pitch 0 in 2D.

**Verified in-browser, both directions, zero console errors:** `?view=2d` comes up mercator with `getTerrain()` false and pitch 0; the toggle returns the globe with terrain true; URL and storage follow (`?view=3d` / `mahanathi.view`). Note: pitch 0 after returning to 3D at the national view is correct — the national preset is a straight-down bounds fit.

The optional zoom-interpolated projection (the last optional item) was **not** taken: each manual set reloads tile managers, and the clock did not allow verifying the interpolate form against the installed dist.

### P0 — all Thai type loopless (commit `1cf1aea`)

Noto Serif Thai retires: the sentences and the wordmark render in Mahanathi Sans, hierarchy by size, leading and tone colour, never family (no letter-spacing — tracking breaks Thai combining marks). The `@font-face` leaves the build; the WOFF2 and OFL text stay as the specimen's comparison material. ADR 0005 carries the addendum (the lead's verified research + the Google Fonts reference links, the owner's rule, the consequences); DESIGN.md and the specimen note updated. **Verified in-browser:** the FontFaceSet holds only the two sans faces, no serif WOFF2 is requested, the computed sentence family is the sans stack. (`document.fonts.check('1em "Noto Serif Thai"')` returns true — the known fallback-resolution false positive from the Phase 3 integrity work; the FontFaceSet and the network log are the evidence that counts.)

### P1 — the copy rule (commit `8b37cb6`)

English instrument, Thai voice — written into DESIGN.md (Named Rules) and the design brief (§8). Applied: the station sheet's labels are `WATER LEVEL` / `BANK FULL` / `% BANKFULL` / `MEASURED`, timestamps render `18:40 ICT`, the view labels were already English. The datum caveat joins the sheet as a **Thai** note (a caveat the reader must understand is voice), and the unit stays ม.รทก. — its equality with m AMSL is not provable from the published data, so writing "m MSL" would claim an equivalence we cannot verify. The subset charset includes Basic Latin (U+0020-007E), so the new strings are covered without a rebuild.

### P1 — DEMO.md + the credit line (commit `f477249`)

`docs/DEMO.md`: the Thai opening script, the 20-second QR beat, the three why-us bullets, the honest closing line (the lead's own), and the "do not say" section headed by the banned-phrase rules. Every claim is one the deployed build can show — the 72-h timeline appears as the API capability it is (`docs/api.md`), not as a UI chart that does not exist yet. The panel's permanent credit now names ThaiWater's open data (HII · RID) over OSM/Protomaps.

## Suites (real numbers, this phase's runs)

tsc clean; eslint clean (`--max-warnings=0`); unit **35/35**; gate-evidence spec **4/4** (run three times — after each unit that changed the shots); the full desktop e2e run **14/15** with one failure of the known ReadPixels-contention class (`without WebGL` timed out in the pool after the WebGL-heavy gate spec, passed in 0.7 s on immediate isolated re-run — the same class recorded since Phase 3). The full three-project run was not repeated this phase; its last green record is the phase-3b-fix close (35 passed / 10 skipped / 0 failed at `--workers=2`).

## What did not land (the honest list)

- **P2 — shareable station deep-link** (`?station=<id>` + Web Share): not started. The lead's sequencing put it after the P1s and the clock won.
- **P2 — the loopless guard** (`e2e/tests/loopless.spec.ts`, hole-count raster test with the Looped sibling as positive control): not started. This is the one I would land first if any time is found before 16:00 — the loopless rule currently rests on the FontFaceSet evidence and the ADR, not on a committed raster test.
- **P3 — the about/ทำไมต้องมหานที section:** cut per the prompt's own instruction ("cut without ceremony"); its copy lives in DEMO.md.
- Critique iteration 2's two P1s (the ใกล้ฉัน sample-station disclosure in the sheet head; the desktop national frame cropping Thailand's tips — the globe/fitBounds interaction) also remain open; both are recorded in GATE-phase-3b.md's open questions and need Phase 5+ or a lead decision.

## Stop

Stopped for the lead at ~15:05 as instructed: pushed-and-deployed beats perfect, and the push/deploy are the lead's alone. The app is shippable at every commit above.

## Delta addendum (the owner's mid-phase directions, ~15:00)

The lead delivered `process/prompts/phase-4-delta.md` during the final minutes. Disposition under the hard stop:

1. **View labels renamed to DEFAULT / TERRAIN 3D** — landed. The owner's clarification (a plain default map vs the realistic view; ours is terrain relief, not satellite). The strings, DESIGN.md's Instrument/Voice rule and DEMO.md updated in the same commit; gate screenshots re-shot (the gate-shots test needed its isolated re-run — the same contention class).
2. **Loading screen + intro animation** — **cut**, named here: ~1.2 s of choreography did not fit the minutes left, and a rushed intro risks the never-delays-interactivity rule. The skeleton surface and motion tokens it would reuse are in place for the next round.
3. **Tailwind v4 + shadcn/ui foundation** — **cut**, named here: an install + init + the toggle/share rebuild is a half-hour unit done properly (the tokens.css drift test must keep passing, shadcn's variables scoped in their own layer), not a five-minute one. The two-stage decision (foundation + new components first, migration later with gates re-run) is the lead's to schedule post-interview.
4. **Template-first standing policy** — landed: brief §11.

Fewer, better — the clock won.
