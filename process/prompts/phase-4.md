# Phase 4: Product build

Lead notes from the Phase 3 review are at the bottom; address them first.

**Goal:** turn the approved look into the product described in `docs/brief/00-PRODUCT.md` (jobs J1–J6), on live data, without diluting the look.

## Build order
1. **App shell:** the shell and sentence paint first, and the 3D chunk lazy-loads. Data loads from `/api/v1/now`, with last-good-snapshot caching.
2. **J1, national view:** the relief and live gauges with LOD and hierarchy, plus the national sentence.
3. **J2, ใกล้ฉัน:**
   - Geolocation runs on the device. Find the nearest valid station with haversine and unit tests.
   - Fly the camera there, then show the panel with the sentence.
   - Design the denied, timeout and unavailable states.
4. **J3, search:** Thai-normalised and ranked, with unit tests. Keyboard accessible.
5. **Answer panel:**
   - The sentence, facts in tabular figures, the measured time and source.
   - A 72-hour chart against bank-full, drawn with care: no chart-library defaults, a labelled bank-full line, and honest gaps for nulls.
   - A bottom sheet with snap points on mobile and a column on desktop.
6. **J5, timeline:** scrubbing animates every gauge, keyboard and touch work, and the hour is labelled in Thai time.
7. **J4 and J6, states and fallbacks:** stale, suspect and API down. The no-WebGL 2D fallback, reduced motion and LINE in-app handling.
8. **About page:** short and honest, with credits and the licence.

## Verification
- E2E covers every scenario in engineering brief §5, run against the real binary serving fixtures. Add a test-only flag that serves fixtures.
- Measure the performance budgets (bundle sizes, Lighthouse mobile, fps on desktop and a throttled run) and record them in the report.
- Re-run the design gate on the real product into `docs/design/screens/phase-4/` and `docs/design/GATE-phase-4.md`.

## Report and stop
Write `process/reports/phase-4.md`. Stop.
