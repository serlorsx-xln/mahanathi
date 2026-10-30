# Phase 3b report: the relief rebuild on MapLibre

- **Implementer:** the same session throughout (GLM-5.3 via the owner's router; quality gates unchanged).
- **Phase window:** 2026-10-30 ~02:05 ICT (Step 0) to ~12:25 ICT, interrupted twice by the owner's router and once — for about four hours — by the UNITEK volume dropping off USB mid-planetiler-run (it recovered; see Environment below).
- **Tree at report time:** the phase's commits are `0098253` (Step 0), `37da366`, `c71e88b` (base frame), `88195ee` (gauges + discs), `187ef97` (interaction port), plus this report's gate commit.

## Summary

The relief the owner rejected is replaced by the engine the owner chose: MapLibre GL 6.16.0 renders the terrain, the hillshade, the basemap and the fog; our data — borders, rivers and all 804 gauges — renders on top as a custom instanced WebGL layer, with the disc encoding behind `?gauges=discs` for the lead's visual pick. The re-gate ran as the prescribed dual-agent critique, found five real P1s (a 24%-of-canvas red slab at the landing zoom, a phone stacking collapse, a Protomaps-grey takeover at close zoom, sub-floor tap targets and a suppressed keyboard focus), and every one was fixed and re-measured before commit. The e2e suite (10 tests) and unit tests (35) are green; the detector is clean; the evidence is committed under `docs/design/screens/phase-3b/`.

**This session cannot view images** (unchanged). The gate rests on computed evidence, and the lead's visual pass on real hardware — including the poles-vs-discs decision and the exaggeration ADR note — is the required confirming step before Phase 4.

## What was built

| Commit | Piece |
|---|---|
| `0098253`, `37da366` | Step 0: the evidence corrections, ADR 0007, the superseded-gate note |
| `c71e88b` | The base frame: MapLibre globe + terrarium DEM (committed `build_terrarium.py`) + hillshade + fog + retinted Protomaps basemap; `src/scene/` deleted (history keeps it) |
| `88195ee` | The gauges: custom instanced layer (cap+fill+overflow, stale hollow, suspect translucent, base ring, fog) + the disc fallback; the ported bug fixes 5/7/8/9/10 |
| `187ef97` | The interaction port's last item: the LINE in-app-browser offer; basemap provenance in SOURCES.md |
| (this commit) | The re-gate: critique fixes (width clamp, phone station-owns-screen, extended retint, a11y batch), GATE-phase-3b.md, the dated critique artifact |

## Engine lessons (all DEVLOG-recorded)

1. MapLibre 6 resolves its worker against `import.meta.url`, which breaks under bundling — `setWorkerUrl` with the worker emitted as its own asset.
2. The basemap HEAD check must test the content type: dev servers answer missing paths with the SPA fallback (200, text/html), which the PMTiles reader then chokes on.
3. The custom-layer projection prelude owns `#version` and its uniforms are ours to upload — a layer that compiles, links and draws can still be silently invisible.
4. `projectTileFor3D` takes metres under globe but mercator z under mercator — one uniform switches the unit.
5. Canvas pixel evidence needs `preserveDrawingBuffer`.

## The 15 bugs (disposition)

Died with the engine (1, 2, 4, 11, 12, 14 — terrain shading, datums, sea, AO, skirts, depth precision are MapLibre's own); died with the route rewrite (15); ported and fixed in the custom layer (5 no per-frame uploads, 7 live reduced-motion listener, 8 suspect encoding, 9 overflow clamped by construction, 10 slim profile + base ring + width clamp); superseded by the a11y batch (the label-height bug 6 has no 3D labels — the DOM sheet carries the name).

## Budgets (measured, gzip)

Initial JS 72.6 kB (budget <300 kB ✓). Lazy relief chunk 295.4 kB + worker 143.8 kB ≈ 439 kB, against the ~250 kB the R3F stack weighed — heavier, recorded in the gate, not hidden. Fonts unchanged (64 kB).

## Open questions (for the lead)

1. **Poles vs discs** — the brief's midday pick; both ship behind the flag. Measured trade-off in the gate: poles carry the full encoding but sit below the phone-country perceptual floor; discs register at every zoom with less encoding.
2. **Thai basemap labels** — needs a self-hosted glyph PBF set; credited regardless.
3. **The exaggeration default (1.5×)** — MapLibre units differ from the old rig's; an ADR note when the lead picks on hardware.

## Environment (the four-hour gap)

At ~03:02 the UNITEK volume dropped off USB mid-write (kernel: transaction error, "hardware connection lost"; a scheduled restic job and an overnight Remotion render were hammering the same USB2 link) and re-enumerated 4 s later, but diskarbitration stalled on the dead node until the owner's watchdog force-recovered it. Costs: the planetiler run died reading its jar off the dying transport (rebuilt isolated on the internal disk, 10 min), and the implementing session idled ~07:00–11:00 waiting for the volume. The repo itself was never corrupted — the last commit before the drop (`c71e88b`'s tree) verified clean, and every file this phase touched was re-checked after remount. The lookdev.css specificity fix that appears in `88195ee` was made by another session on this machine during the outage window; it was reviewed, found correct, and committed with credit in the message.

## Stop

The phase stops here for the lead's review: the visual pass on real hardware, the poles-vs-discs decision, and the Phase 4 prompt. No push, no deploy — those are the lead's alone.
