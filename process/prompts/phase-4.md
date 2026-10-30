# Phase 4: product surface on the new base (compressed for the 16:00 interview)

If any Phase 3b-fix item is still outstanding when you read this, finish it first — then this phase, in the priority order below. **The lead will cut this phase at ~15:05 ICT regardless of where it stands; every commit must leave the app shippable.** Budget your units accordingly.

The owner's three new directions, researched and specified:

## P0 — the view modes: '2D MAP' / '3D TERRAIN' (Google-Maps pattern)
A segmented pill in the map chrome, English labels `2D MAP` / `3D TERRAIN` (aria-pressed, keyboard focusable), state synced to the URL as `?view=3d` (deep-linkable for the demo) and remembered in localStorage (try/catch).

- **2D MAP**: `map.setProjection({ type: 'mercator' })`, `setTerrain(null)`, pitch 0 / bearing 0, hillshade softened to 0.25, **full night fog** (the sky block only works in mercator — this mode gets the fog back).
- **3D TERRAIN**: globe + terrain + globe atmosphere; presets keep their authored pitch (55/60).
- Two sky specs exported from `style.ts`, both built only from tokens (the no-hand-hex rule holds): mercator keeps today's fog-style sky; globe uses a higher `atmosphere-blend` (Rayleigh shader draws only on globe) — this converts the console warning into a designed trade. Update the ADR 0007 note/DEVLOG accordingly.
- **Never switch projection and fly the camera in the same frame**: sequence setProjection → `projectiontransition` → easeTo. Re-apply terrain and `setStations` after the transition (the gauges layer is already projection-aware). Guard toggle clicks while mid-switch.
- Mask the tile reload with a fade-through-fog overlay (~180 ms to ~0.4 opacity in the fog token, removed on next `idle`) — under reduced motion, cut. Presets flatten in 2D (pitch forced 0) and compose with the existing preset buttons.
- Optional last: in 3D mode use the zoom interpolation `['interpolate',['linear'],['zoom'],4,'globe',8,'mercator']` so zooming in flattens like Google's. Do NOT tween projection manually (each set change reloads tile managers).

## P0 — all Thai type loopless (no หัว), the serif retires
IBM Plex Sans Thai ("Mahanathi Sans") is confirmed loopless; **Noto Serif Thai — the sentence face — is the looped one.** Retire it: the national/station sentences render in Mahanathi Sans, differentiated by size/tracking/the sentence tone colour, not family. Update `tokens.ts` (`sentenceStack`), DESIGN.md, and **append an ADR 0005 addendum** recording the external confirmation (Google Fonts links) and the retirement decision. Re-shoot the affected screenshots.

## P1 — the copy rule: English instrument, Thai voice
Write the rule down (DESIGN.md + a line in the design brief): **Thai is the voice (the sentence, situation labels, buttons the user reads as speech); English is the instrument (measurement labels, units, credits, timestamps, controls).** Then apply it: panel measurement labels → `WATER LEVEL` / `BANK FULL` / `% BANKFULL` / `MEASURED`, view labels `2D MAP`/`3D TERRAIN`, timestamps/units in English. Keep the datum footnote honest (ม.รทก. is not provably identical to m AMSL — keep the Thai note). Verify the WOFF2 subsets cover the new Latin strings (`scripts/research/subset_sizes.sh`; rebuild if short).

## P1 — docs/DEMO.md (the interview script) + the credit line
- The verified comparison, honest, no disparaging: ThaiWater's new portal is a clickable-map dashboard requiring station-by-station clicks; our differentiators (researched 2026-10-30): the plain-Thai national sentence, ใกล้ฉัน computed on-device (position never leaves the phone), the 72-h timeline, works in the LINE webview, dark night-first design, and being **built on their open data** (315 RID stations of the 804 — say so as a feature).
- Opening script (3–5 lines Thai), the 20-second QR-and-phone beat, the "why us" three bullets, and the honest closing line ("ระบบทางการลึกกว่ามาก — เราอ่านข้อมูลเปิดของเขามาเล่าให้คนทั่วไปฟังในหนึ่งประโยค"). **Never claim sub-second paint** (brief says <2 s first frame); no prediction/safety language (the banned-phrase test still governs every string).
- Add the permanent credit line to the map attribution area: built on ThaiWater open data (HII/RID) + OSM/Protomaps.

## P2 — shareable station deep-link
On station select: `history.replaceState` → `?station=<id>`; restore on load. A small share button (Web Share API, clipboard fallback). Cheap, demo-visible.

## P2 — the loopless guard
`e2e/tests/loopless.spec.ts`: canvas hole-count on ก/จ at ≥220 px, calibrated against the IBM Plex Sans Thai **Looped** sibling loaded test-only as a positive control (control assertion first). Raster test — keep it tolerant (large size, single family).

## P3 — cut without ceremony if the clock wins
The about/"ทำไมต้องมหานที" UI section (its copy lives in DEMO.md anyway).

## Gates
Same anti-slop checklist for every new pixel (the segmented control and share button are UI — they inherit the tokens, the motion rules, the evidence discipline). Run the full suite; record real numbers in the report. Stop for the lead at 15:05 sharp — pushed and deployed beats perfect.
