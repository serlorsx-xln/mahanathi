# 0008. Terrain exaggeration default on the MapLibre rig

Date: 2026-10-30 (phase 3b-fix) · Status: provisional (the lead confirms on hardware)

## Context

The hand-built R3F rig (ADR 0004/0006) ran vertical exaggeration 4–40×
on a metre-scale heightfield we shaded ourselves. MapLibre's terrain
exaggeration multiplies the DEM's own metres against a camera that
renders the whole country at once, so the same number produces a
completely different relief. The look-dev slider spans 1–40× in the
new units, and the landing view needed a default.

The design brief (01, §4) wants "the national survey's plaster relief
in a dark map room" — readable landform, not drama. The gate critique
measured the landing frame against token-colour dominance; the
exaggeration changes hillshade contrast, not token shares.

## Options

1. **1.0× (true scale).** Honest to the metre, but Thailand's central
   plain reads dead flat at national zoom — the relief stops being the
   page, which the brief makes a principle.
2. **1.5×.** The plain gains readable shape while the ranges stay
   recognisably Thailand to anyone who knows the country's outline;
   close-up station views don't yet distort valley floors enough to
   mislead the eye about bank-full geometry.
3. **≥2×.** Dramatic at national zoom, but station close-ups begin to
   exaggerate exactly the geometry the gauge encoding depends on —
   the eye would read the pole-to-valley relationship as it never is.

## Decision

**1.5×, provisional.** Chosen during the Phase 3b look-dev as the
lowest exaggeration at which the central plain reads as relief at the
national fit. Recorded now (the lead's remediation finding 16) because
the number had been living only in the slider default.

## Consequences

- The slider stays 1–40× so the lead can confirm or change the default
  on hardware at the scheduled visual pass; if changed, this ADR gets
  a dated addendum, not a rewrite.
- Any change to the default re-seats the gauges too (their ground
  contact follows `queryTerrainElevation`, which is
  exaggeration-aware) — the gate's stale close-up and station shots
  must be re-shot after a change.
- True-scale remains available at 1.0× for any future "surveyor mode"
  where metre-honesty is the point.
