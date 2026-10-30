# Phase 3b-fix: remediation of the lead's adversarial review

The lead's multi-agent review of your Phase 3b confirmed 25 findings (0 refuted, verified against the repo). The same session continues — do not repeat completed work; fix the list below. The two blockers first, then majors, then the quick minors. Re-gate evidence rules from `docs/brief/04-REBUILD-RELIEF.md` (the honest scope note) are **binding**: no computed-evidence claim without a committed script or JSON; screenshots must be actually re-shot, not relabelled.

## Blockers
1. **The gate claims "re-shoot of every screenshot" but only one 1440 gauge shot was re-taken.** `gauges-poles/discs-national/station-1440.png` are still the `88195ee` versions — one shows the red slab the gate says was fixed. Re-shoot the named viewports for real, and correct the GATE wording to describe exactly what was re-shot and when.
2. **Two backdrop blurs — the brief allows one.** `lookdev.css:193` (panel) and `:273` (station sheet). Brief 01: one functional blur for the panel over the relief. The sheet's comment self-granting "the second surface allowed the blur" must go, along with one of the blurs — keep the panel's, restyle the sheet with an opaque/veiled surface per the tokens. Fix the anti-slop checklist scoring to match reality.

## Majors
3. **Stale stations are invisible in poles mode.** `gauges.ts:148`: `atCap` is gated on `vStale < 0.5`, so every fragment of a stale pole is discarded — the hollow cap stroke never draws and `uSteel * 0.75` is dead code. Drop the `&& vStale < 0.5` clause (the `vStale > 0.5` branch already runs before branch 2). Then re-shoot a stale-station close-up as evidence.
4. **Preset/exaggeration/flight fired before `load` are dropped.** `LookdevMap.tsx`: the load handler uses first-render closure values (exaggeration captured forever), and the preset/exaggeration/flight effects no-op when `readyRef` is still false — user clicks ใกล้ฉัน during load → stuck spinner. Keep pending values (refs or a pending-state) and apply them once on load.
5. **`webglAvailable()` accepts WebGL1 but MapLibre 6 requires WebGL2.** `Lookdev.tsx:47`. A WebGL1-only device (exactly the LINE-webview population you special-case) passes the gate then gets a dead map. Test `webgl2` only, with a graceful floor for the rest.
6. **Every pixel-sampling/contrast/budget number in GATE-3b and the report has no committed script or JSON** (24.45% → 6.72%, 77.7% → 0.03%, budget figures, contrast ratios). Brief 04's rule is binding. Write the measurement script under `e2e/` or `scripts/`, commit it, run it, commit its JSON output, and cite the file path at every number. Where a number cannot be regenerated, remove it rather than leave it untraceable.
7. **Hand-written `#599c5b` in `style.ts:45`** contradicts the file's own "no colour written by hand" claim. Import the situation colour from `tokens.ts` (the ratios stay hand-set).
8. **The LINE-offer "e2e-verified" claim is false** — no such test exists anywhere; DEVLOG, gate and the `187ef97` message tell three conflicting stories. Either write the real e2e (LINE UA → hint visible, dismiss works, link carries `?openExternalBrowser=1`) and cite it, or correct all three documents to say what actually checked it (the critique's browser measurement). The truthfulness rule is not negotiable.

## Minors (fix the quick ones now; the rest go to the deferral list explicitly)
9. Map init runs in an unguarded promise chain — add a `.catch` that surfaces the error state (Lookdev.tsx already has one).
10. Discs fallback: a stale station's stroke can take a level colour (`discs.ts:55`) — force the stale steel stroke like the poles claim to.
11. `flyTo` moves have no duration; brief 01 wants 1.2–2.0 s ease-out. Set explicit durations.
12. Remove `@types/three` from devDependencies.
13. `lookdev.css:47` hard-coded `rgba(255,255,255,0.015)` — token or named var.
14. The critique artifact's scores sum to 27, not 28/40 — recheck the sheet arithmetic everywhere it propagated (gate, DEVLOG, report) and correct.
15. The shader compile throw inside `render()` aborts MapLibre's frame loop — catch and degrade to the discs layer instead.
16. Record the 1.5× vertical exaggeration decision in an ADR (brief 01 §4).
17. The sky/fog block is a no-op while globe projection is active (MapLibre 6 limitation, warned in console). Disclose in the gate/report and ADR 0007 rather than claim "one continuous fog" unconditionally; note the mercator-transition behaviour.
18. DEVLOG jumps ~02:55 → ~11:55: add the real history — the USB outage, the session's idle-to-hang, the repo fsck, and the `lookdev.css` specificity fix made by the lead's session during the outage (the report mentions it; the DEVLOG must too).
19. Only one dated critique iteration ran (disclosed) against the mandatory-two rule — run a second real one now against the post-fix build and commit its dated artifact, or state plainly in the gate that the rule was cut for the deadline with the lead's consent.

## Then
- Re-run the full web suite (tsc, eslint, unit, e2e) and record real numbers.
- Update `docs/PROGRESS.md`, `process/reports/phase-3b.md` (append a remediation section — do not rewrite history), DEVLOG.
- Commit in small honest units. Stop for the lead when done — Phase 4 does not open until the lead and the owner approve.
