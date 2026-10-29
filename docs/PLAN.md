# Plan

The workflow asks each phase to append its own plan here. This version was written at the end of Phase 1 (2026-10-29 ~18:45 ICT). It holds the Phase 1 plan as executed, an outline of Phases 2 to 6 with estimates, and the risks ranked.

## Time budget

- **Deadline:** live and rehearsed by **2026-10-31 12:00 ICT**. That is about 41 hours after this plan was written.
- **Working target:** deployed by **2026-10-31 06:00**. This leaves the owner six hours to rehearse on the real URL, plus a buffer for one redeploy.
- **Estimates:** the figures below are implementer working hours. They do not include the lead's reviews between phases. Allow about 0.5 h per review, and more for the Phase 3 look approval, which needs the lead and possibly the owner.

| Phase | Estimate | Earliest start if reviews are prompt | Gate that decides the next start |
|---|---|---|---|
| 1 Discovery and foundation | started 17:49; the end time is in the Phase 1 report | | ADRs 0001–0004, research, scaffold, CI |
| 2 Data core (TDD) | 4–5 h | 10-29 evening | `go test` green, coverage ≥ 90 %, a live ingest |
| 3 Look-dev | 5–7 h (two gate iterations) | 10-30 early | **the lead approves the look** |
| 4 Product build | 6–8 h | 10-30 midday | E2E green, budgets measured |
| 5 Polish and hardening | 3–4 h | 10-30 night | Lighthouse and audit gates |
| 6 Ship docs and handover | 2–3 h | 10-31 early | container verified, docs complete |
| **Total, Phases 2–6** | **20–27 h** | | |

- **Slack.** Back-to-back, 20 to 27 h of work plus reviews fits before 06:00 on 10-31 with about 6 to 12 h to spare. A second rejected look in Phase 3 would use most of that slack. This is why the riskiest item below is the look, not the code.

## Phase 1: Discovery and foundation (executed)

What was planned and how each item was verified. The full record is in `process/reports/phase-1.md`.

1. **Repo hygiene for exFAT.**
   - Covers `.gitignore`, `.editorconfig`, `.npmrc` and `core.fileMode false`.
   - Verified by `git status` staying clean of `._*` files.
2. **ThaiWater re-verification and fixtures.**
   - The two endpoints were fetched live, the responses cached, and fixtures trimmed by a script.
   - Verified by the fixture coverage table, which prints each row's real level.
3. **Data-quality scan.**
   - A stdlib script recomputes every derived field.
   - Verified by agreement counts against the published values.
4. **The server.**
   - A Go module with graceful shutdown.
   - Verified by a test that serves `/healthz` and then shuts down cleanly, and by a manual SIGTERM.
5. **The web app.**
   - Vite, React 19 and strict TypeScript, with ESLint, Prettier and Vitest, and three, R3F and drei installed but unused.
   - Verified by typecheck, lint and test passing, and a build.
6. **E2E.**
   - A Playwright smoke test.
   - Verified passing at 1440×900 and 390×844.
7. **Build and CI.**
   - A Makefile, a Dockerfile and the CI workflow.
   - Verified by `make` targets, a 14.1 MB image answering `/healthz`, and `actionlint`.
8. **Research.**
   - `references.md`, `typography.md` and `geodata-sources.md`, drafted by sub-agents and reviewed before commit.
9. **ADRs 0001–0004, this plan, `PROGRESS.md`, the DEVLOG and the report.**

## Phase 2: Data core (outline)

Tests come first for every pure package. The order follows `process/prompts/phase-2.md`. The estimates are the riskiest parts, not all the work.

1. **`thaiwater`** (0.5 h).
   - DTOs that keep numbers as strings. The load response's numbers are JSON strings, and the graph's are floats.
   - Tolerant decoding: keys can be absent, not just null.
   - A configurable User-Agent, timeouts, and backoff with jitter.
   - A semaphore of 4 for graph calls.
   - Tests on the recorded fixtures only.
2. **`normalize`** (1 h). This implements ADR 0003 exactly.
   - ICT to UTC with `time.FixedZone`.
   - Millimetre rounding of RID float32 noise.
   - The zero-coded and absent reference rules.
   - Stale against an injected clock.
   - The suspect rules.
   - Table cases named after the fixture stations: P.64, Gt.11, BLGTU06, Gt.9, FOP045, BSP002, URTU09, BKC004 and others.
3. **`humanize`** (1.5 h).
   - The spec in brief §3: at least 30 table cases, golden files for sentences, and a banned-phrase fuzz over every fixture and every generated case.
   - National-sentence grammar for 0, 1 and many.
   - The Thai formatter module (numbers, `14:00 น.`, `29 ต.ค. 69`).
   - Coverage of 90 % or more.
4. **`store`** (0.5 h).
   - modernc, WAL, embedded migrations and the upsert per ADR 0002.
   - Bucketing for the timeline: the last non-null value in `(HH−1:00, HH:00]`.
   - Tests on a temporary database **on this exFAT volume**, which is the real check of the WAL probe.
5. **`ingest`** (0.5 h).
   - A run at start, then at :07 past each hour. Measure first when the hourly stations publish, and move the minute if :07 is too early.
   - A 72-hour backfill for stations without history.
   - Run metrics logged.
   - Tested with a fake client.
6. **`api`** (1 h).
   - `/now`, history, `/timeline`, `/healthz` and `/readyz`.
   - ETag, gzip, security headers, and a token bucket with `Retry-After`.
   - The timeline encoding documented in `docs/api.md` and measured on real data against the 80 KB budget. 800 × 72 values as `uint8` with 255 meaning null is 57.6 KB before compression.
7. **`main` wiring.**
   - The `go:embed` placeholder from ADR 0001, so a fresh clone still compiles.
8. **Finish `data-quality.md`.**
   - At least two more snapshots, including a night one, plus the backfilled graph references compared with the load references.

## Phase 3: Look-dev (outline)

1. **Tokens and `DESIGN.md`.**
   - One source with a generated mirror (CSS custom properties and TS).
2. **The situation scale.**
   - Lightness-ordered.
   - Deuteranopia and protanopia simulations.
   - Contrast computed for every text and surface pair.
3. **The type specimen (`/specimen`).**
   - The three pairings from `typography.md`, with the sentence face switchable.
   - Leading set per face from the measured no-touch values.
   - The decision goes in ADR 0005, and the subsets are self-hosted with the licence name IDs kept and Plex renamed.
4. **The geodata pipeline.**
   - Scripts in `geodata/` following `geodata-sources.md`: DEM tiles, a stitched heightmap, the encoded asset, and borders and rivers.
   - `geodata/SOURCES.md` with the licences.
5. **The relief prototype (`/lookdev`).**
   - Displaced relief, baked or computed hillshade, flat-area toning, one north-west key light and fog from tokens.
   - The border and rivers on the surface, and instanced gauges from a saved real snapshot.
   - Three camera presets and one flight.
   - Exaggeration and smoothing recorded in ADR 0006.
6. **The gate.**
   - Screenshots at 390×844 and 1440×900, then `/impeccable critique`, then the checklist, at least twice.
   - The ten checks at the end of `references.md` make "premium" testable.

## Phase 4: Product build (outline)

1. **The shell and the sentence paint first.**
   - The 3D chunk is about 250 KB gzipped (three, R3F and `CameraControls`, measured in Phase 1). It loads lazily after the shell.
   - The last good snapshot is cached in the browser.
2. **J1, the national view.** The national sentence and level-of-detail hierarchy.
3. **J2, ใกล้ฉัน.**
   - On-device haversine, tested.
   - The flight follows the terrain.
   - Designed states for denied, timeout and unavailable.
   - Geolocation needs a secure context, so it only works on the HTTPS live URL or on `localhost`.
4. **J3, search.** Thai normalisation, ranking and keyboard access, tested.
5. **The answer panel.**
   - The sentence, facts, measured time and source.
   - A hand-built 72-hour chart with a labelled bank-full line and honest gaps.
   - A bottom sheet on phones.
6. **J5, the timeline scrubber.**
7. **J4 and J6, states and fallbacks.**
   - States: stale, suspect, and API down.
   - Fallbacks: the 2D map without WebGL, reduced motion, and the LINE in-app browser.
8. **The about page.** It includes the stale-count sentence from `data-quality.md`.
9. **Verification.**
   - E2E for every scenario in brief §5, against the binary in a fixtures mode.
   - The budgets measured.
   - The gate re-run.

## Phase 5: Polish and hardening (outline)

- `/impeccable audit` and `/impeccable polish`.
- Keyboard path, focus, landmarks and the canvas text alternative.
- Phones at 390 and 360 wide, safe areas and 44 px targets.
- A Thai copy pass.
- Meta: an OG image from a real render, and a designed 404.
- Headers and CSP checked with curl.
- A rate-limit test.
- ThaiWater-down behaviour.
- Before and after performance numbers.
- Lighthouse mobile: accessibility ≥ 95, best practices ≥ 95, performance ≥ 85.

## Phase 6: Ship docs (outline)

**The container:**
- An empty `/data` owned by `nonroot`. Without it, the first write to a fresh named volume fails (ADR 0002).
- A HEALTHCHECK.
- `.env.example`.
- `docker run -p 8080:8080 -v mahanathi-data:/data` verified.

**Docs:**
- `README.md`, Thai first.
- `docs/ARCHITECTURE.md`.
- `docs/REVIEW.md`: the files the owner must read, plus the AI mistakes from this DEVLOG with the check that caught each one.
- `docs/DEMO.md`, with the backup plans.

## Risks, ranked

| # | Risk | Likelihood × impact | Mitigation, and what would trigger it |
|---|---|---|---|
| 1 | **The relief does not read as premium** within two gate iterations. A subjective gate sits on the critical path. | medium × high | The checks in `references.md`; Swiss-style shading with no cast shadows; the fallback ladder in ADR 0004. Trigger: the second iteration still fails the checklist. |
| 2 | **First frame over 2 s on a mid-range phone.** The 3D chunk is about 250 KB gzipped and the heightmap is up to 1.5 MB. | medium × high | Text-first shell (the sentence is the LCP element); a small first heightmap with a sharper one after; `frameloop="demand"`. Measured in Phase 4, not assumed. |
| 3 | **ThaiWater is down or changes shape during the demo.** | low × high | Last good snapshot from SQLite with an honest timestamp; a fixtures mode for an offline demo; ingest warns when `rule_web` changes. |
| 4 | **Phase 3 approval waits on people.** The lead, and possibly the owner, must look. | medium × medium | Send the screenshots as soon as the first iteration is ready, and keep the geodata pipeline moving while waiting. |
| 5 | **Thai rendering differs across browsers.** The research used HarfBuzz only; Safari uses Core Text. | medium × medium | The specimen is checked in WebKit through Playwright, and the leading comes from measured no-touch values with margin. |
| 6 | **The LINE in-app browser** has limited WebGL or no geolocation. | medium × medium | Detect it and offer "open in browser"; the 2D fallback; search as the path when geolocation is denied. |
| 7 | **Data-quality surprises on the live set**: new zero codes, a new datum shift, references changing between snapshots. | medium × low | ADR 0003 rules are table-tested; `data-quality.md` is re-run on more snapshots in Phase 2. |
| 8 | **This machine.** exFAT means AppleDouble files, slow `node_modules` and a Docker context that needs the tar workaround; the internal disk is nearly full. | high × low | Workarounds are in place; caches stay on this volume. CI on Linux is the clean check once the lead pushes. |
| 9 | **CI has never run.** It is defined but not pushed. | medium × low | `actionlint` is clean, and every command in it runs locally. The first push may still need a fix. |
