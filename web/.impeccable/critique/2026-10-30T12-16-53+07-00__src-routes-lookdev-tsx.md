---
run: 2026-10-30T12:16:53+07:00
committed: 2026-10-30T12:16:53+07:00
target: src/routes/Lookdev.tsx
note: >-
  The Phase 3b re-gate critique, run as two isolated sub-agents
  (Assessment A: design review; Assessment B: detector + browser
  evidence) against build 187ef97, synthesized in the parent, then the
  priority findings fixed and re-measured before commit. Committed in
  the same sitting as the run.
total_score: 28
max_score: 40
p0_count: 0
p1_count: 5
---

# `/impeccable critique` — Phase 3b re-gate (MapLibre relief, both gauge encodings)

**Method:** dual-agent. Assessment A (design review: DOM, computed styles, measured layout, canvas pixel sampling at four viewports in poles and discs modes, LINE-UA / no-WebGL / reduced-motion variants) and Assessment B (detector CLI + console/a11y/contrast/tap-target/overflow measurements) ran as isolated sub-agents; synthesis happened only after both finished. Detector: clean (`[]`, exit 0) on all seven changed files. Neither session can view images; every claim below is a computed measurement.

## Assessment A verdict

**Specific in its stillness, generic at its moments of motion.** The resting national frame is unmistakably this product; the two authored moments (first frame, ใกล้ฉัน landing) dissolved into generic material at the extremes. Heuristic scores (0–4): 2.5, 4, 2, 3, 3, 2.5, 2, 2.5, 3.5, 2 → 28/40.

**Priority issues found (P1, all fixed after synthesis):**

1. *Landing frame a monolithic red slab:* 24.45% of the desktop canvas exact `#ee5d58` (phone 20.55%) — the blade had a minimum screen width but no maximum. **Fixed:** width clamped to ≤36 px; re-measured 6.72%.
2. *Phone stacking collapse:* after selecting a station, panel + sheet + footer stacked to 100% viewport coverage with 12/19 px overlaps — 0% relief visible at the journey's peak; no way to dismiss the sheet. **Fixed:** on phones a chosen station owns the screen (`data-station` hides panel and dev chrome; the sheet gains an authored-SVG close that returns to the national view). Re-measured: panel and footer hidden, sheet 432–648 of 844, close works, panel returns.
3. *First-frame gauge invisibility on phones:* zero detectable situation-colour pixels at the national view on any 390-px viewport (poles are ~1.6 px tall at that zoom — wide-but-flat smears below the perceptual floor); discs at the same view do register. **Partially fixed:** minimum blade width 2.5→3.5 px, fog start 0.35→0.55 of horizon (desktop national re-measured: level-3 513→743, level-4 633→1069, level-5 72→125 px). Phone-country poles remain below the floor — recorded in the gate as the open poles-vs-discs question the brief routes to the lead's midday visual pick.
4. *Committed world dissolves at close zoom:* 77.7% of the landing frame was Protomaps DARK's own `earth` #1f1f1f with #3d3d3d roads and #111 buildings. **Fixed:** the retint now maps earth/landcover/buildings/roads to token-derived colours (only mix ratios hand-set). Re-measured: 0.03% Protomaps grey.
5. *Keyboard entry invisible:* the MapLibre canvas is tab-focusable but its outline was suppressed. **Fixed:** `:focus-visible` outline in the accent token. (The wrapper's `role="img"` on an interactive surface is recorded as a Phase 4 deferral with keyboard station access.)

Also: the LINE-hint covering the answer panel on phones — reported by A's early pass against the pre-fix CSS, **not reproducible** on the current build (B and the parent re-measured: hint at bottom, no overlap, close clickable 4/4); recorded here because a finding that dies on re-measurement is still evidence about how easily phone stacking breaks.

## Assessment B measurements

- **Contrast:** 19 elements measured; one failure — `.lookdev__nearme-note` at 4.44:1 in the button's hover state. **Fixed:** the note lifts to text-secondary on hover.
- **Tap targets (390×844):** nearme 40 px tall, presets 29 px, range input 16 px, linehint-close 20×20. **Fixed:** nearme min-height 44 px, presets taller on phones, close ≥24 px. The range input is dev chrome (recorded as a deferral).
- **Geometry:** station sheet overlapped the presets row by 3.19 px — resolved by the phone station-owns-screen rule.
- **Zero console/page errors** in all five scenarios; no horizontal overflow; a11y tree correct for every spot-checked control; the detector is clean.

## Strengths A recorded

1. The answer panel is the brief realised (sentence-first hierarchy, count-dl as legend, full source chain).
2. The encoding is genuinely geometric before coloured — cap/fill/overflow/hollow-stale/suspect-alpha survive monochrome and CVD by construction, mirrored across poles, discs and DOM.
3. Token and browser-surface discipline holds under measurement; anti-slop sweep clean (the skeleton sweep is the one functional gradient exception).

## Deferred (named, with reasons)

- Phone-country poles below the perceptual floor → the lead's poles-vs-discs decision (both encodings shipped behind `?gauges=`).
- Basemap Thai labels (glyph set unresolved) — Phase 4 with the self-hosted font work.
- Keyboard/screen-reader station access and the wrapper role — Phase 4 with search.
- The exaggeration slider's native 16 px thumb — dev chrome; the slider does not ship.
- Gauge zoom labels, phone snap-point sheet, 72 h chart — unchanged from the Phase 3 deferral list.
