# ADR 0007: Relief rebuild — MapLibre GL JS base, our gauge layer

- **Status:** accepted (Phase 3b, owner's decision)
- **Date:** 2026-10-30 ~02:00
- **Deciders:** the owner (ชานน), from the lead's research (four researchers + synthesis, 2026-10-30); recorded by the implementing agent.

## Context

The Phase 3 relief — a hand-built R3F scene with custom terrain, gauge
and drape shaders (ADR 0004/0006, commits `7c0e65b`…`d72b03e`) — passed
its computed gate but the owner rejected it visually: **ยังไม่สวยเลย**
("not beautiful at all"). The lead's research concurs, scoring the
scene ~3/10 of the premium standard and listing 15 concrete quality
bugs (brief 04, "The 15 bugs"): river/border lines tunnelling through
terrain, unfogged gauges over fogged terrain, a flat unshaded sea,
dashboard-confetti poles without depth cues or ground contact, a
shading ceiling that reads as matte clay, a missing terrain skirt,
depth-buffer z-fighting, and more. Most of these are engine-level
problems that a small team should not re-solve by hand under a
2026-10-31 16:00 interview deadline.

The full options table lives in the lead's research (not committed
here); the brief's own framing of the decision space:

- **Option A (chosen): adopt MapLibre GL JS as the terrain/basemap
  engine and keep our gauge data layer.** The engine's
  shaded-relief/fog/pitch/DPR/mobile-perf stack is debugged by hundreds
  of developers; the proven pattern for live data objects on a MapLibre
  map is a custom layer (Mini Tokyo 3D); the dark basemap with Thai
  labels comes from Protomaps PMTiles (BSD-3, ODbL data) with no token
  or paid API.
- **Repair the hand-built R3F relief** (the implicit alternative): keep
  full control of every pixel but fix 15 bugs, several of them deep
  engine work (multi-directional hillshade, depth precision, mobile
  DPR), against the clock.

## Decision

**Adopt MapLibre GL JS as the terrain/basemap engine; port our gauge
data layer to it as a custom layer.** Execution order (binding, from
the brief): base frame first, gauges as a custom layer, the disc
encoding behind `?gauges=discs` as the guaranteed fallback built in
parallel, then the interaction port (presets, intro flight,
reduced-motion cuts, ใกล้ฉัน prep), then a fresh gate
(`GATE-phase-3b.md`) and a stop for the lead review.

Engine specifics are recorded as they are verified and pinned: the
MapLibre version and licence, the globe-projection availability (prefer
it for the national view if present in the pinned version; Mercator
acceptable otherwise), the DEM encoding (our heightmap re-encoded to
terrarium256 by a committed one-off script under `geodata/scripts/`,
AWS terrain tiles as fallback), and the basemap (Protomaps dark PMTiles
with `name:th` labels, credited ODbL/OSM + Protomaps in the about page
and `geodata/SOURCES.md`).

## Consequences

- **The hand-built relief is removed from `src/`** once the MapLibre
  base renders, and lives only in git history (`7c0e65b`…`d72b03e`).
  Its concepts that earned their place carry over: the bank-full
  cap+fill gauge encoding, the situation scale, the tokens (which feed
  the MapLibre style), the recorded fixture, the DOM-Thai-labels rule,
  the e2e evidence approach.
- The 15 bugs die with the engine where they were engine problems
  (terrain shading, camera clamps, depth precision, world edge); the
  encoding-level ones (5, 6, 7, 9, 10: per-frame re-uploads, label
  pinning, reduced-motion handling, overflow clipping, pole profile)
  are fixed in the port.
- A map engine brings its own world: style JSON, tile caching, and an
  unfamiliar failure surface. The no-WebGL fallback becomes MapLibre's
  own raster fallback path.
- Bundle budget re-checked against ADR 0004's ~250 KB with the real
  numbers recorded when the base frame lands.
- **Process record, for honesty:** at ~01:24 on 2026-10-30 the lead
  reported launching a remediation round for the Phase 3 findings; no
  prompt was written and nothing entered the session. That was a lead
  process error, acknowledged by the lead in the Phase 3b prompt and
  recorded here: the project's honesty standard applies to the lead
  too. The Phase 3b prompt (02:05) is the round that actually ran.
