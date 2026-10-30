# Progress

Updated at the end of every phase. The detail lives in `process/reports/phase-N.md`, and the plan and risks in `docs/PLAN.md`.

## Phases

| # | Phase | Status | Gate | Report |
|---|---|---|---|---|
| 1 | Discovery and foundation | **done**, reviewed by the lead, closed out in Phase 2's Step 0 | research with sources ✓, ADRs for stack, storage and data reference ✓ (plus rendering), scaffold builds and tests pass ✓, CI defined ✓ | `process/reports/phase-1.md` |
| 2 | Data core (TDD) | **done**, awaiting the lead's review | pipeline live-verified on 804 stations ✓, oracle 789/789 ✓, timeline within budget (46 KB gzipped of 80) ✓, humanize coverage 100 % ✓, deploy image builds and self-healthchecks ✓ | `process/reports/phase-2.md` |
| 3 | Look-dev | **rebuilt in 3b**, awaiting the lead's visual pass + poles-vs-discs pick | the R3F relief was rejected visually and replaced on MapLibre (ADR 0007): globe, terrarium DEM, retinted Protomaps basemap, 804 gauges as a custom layer + disc fallback, gate re-run with fixes — see `docs/design/GATE-phase-3b.md` | `process/reports/phase-3.md`, `process/reports/phase-3b.md` |
| 4 | Product build | not started | | |
| 5 | Polish and hardening | not started | | |
| 6 | Ship docs and handover | not started | | |

## Current state (end of Phase 3b + 3b-fix, 2026-10-30)

**What exists:**
- **Server:** the whole data pipeline runs from one binary. `thaiwater` (client with fixtures), `normalize` (ADR 0003 arithmetic, mm-exact), `domain` (levels, percents, distances), `humanize` (the Thai sentence, 100 % coverage), `store` (SQLite WAL, migrations, retention), `ingest` (hourly at :07, 72 h backfill, oracle), `api` (`/now`, `/history`, `/timeline`, `/readyz`) and `main` wiring with env config and a `healthcheck` subcommand.
- **Web:** the visual world is built and gated on MapLibre (ADR 0007). `/lookdev` renders the globe relief (our terrarium DEM, self-built Protomaps basemap retinted to the tokens, one continuous night fog) with all 804 real gauges as a custom instanced WebGL layer (bank-full cap + fill + overflow, stale hollow, suspect translucent) and a disc fallback behind `?gauges=discs`; three camera presets, the ใกล้ฉัน flight, reduced-motion cuts, a no-WebGL floor and the LINE in-app-browser offer. `/specimen` and `/color` are the evidence pages. The product shell is Phase 4. The basemap PMTiles (160 MB, ODbL) and DEM tiles are gitignored locals with recorded build commands (`geodata/SOURCES.md`).
- **Design system:** `web/src/design/tokens.ts` is the single source (generated CSS mirror, drift-tested); DESIGN.md + sidecar recorded from the artifact. Type: Mahanathi Sans (renamed IBM Plex Sans Thai) + Noto Serif Thai, 64 KB self-hosted. Situation scale: CVD-searched, AA on every surface. ADRs 0005 (type) and 0006 (relief look, 12x).
- **Build:** `make docker` builds the 21.4 MB deploy image (distroless, `/data` for 65532, self-healthcheck); `make docker-full` is the whole pipeline once Phase 3 ships.

**Live evidence (21:57 ICT):** first ingest in 33 s — 804 rows, 251,624 readings, 804/804 histories, oracle 789/789. `/now` 85 KB gzipped; the binary timeline body 46 KB gzipped against the 80 KB budget. Details and the five hand-checked sentences: `process/reports/phase-2.md`.

**Data findings:** finalised in `docs/research/data-quality.md` — three snapshots agree (17:54, 19:05, 21:49 night), the two live suspects are hand-checked logger datum alternation.

**Decisions:** ADR 0001–0004 as before. The timeline wire format (binary MNTL, delta varints) is documented in `docs/api.md`, the contract for Phase 4.

**Model note:** from `23a869f` the implementing session runs on GLM-5.3 (the owner's decision, recorded in the DEVLOG and `process/README.md`); everything before is Opus 5.5 at xhigh.

**Phase 3b-fix (2026-10-30 afternoon):** the lead's 25-finding review of the Phase 3b gate, fully remediated — the false re-shoot claim corrected (all evidence now regenerates from the committed spec `e2e/tests/gate-phase3b.spec.ts` and script `e2e/measure/phase-3b.mjs` → `measurements.json`), one functional blur, stale poles visible, load-order races fixed, webgl2-only detection, token-sourced colours, the real LINE e2e, honest no-WebGL copy and HII attribution, ADR 0008 (exaggeration) and ADR 0007's fog-on-globe addendum, the second critique iteration (31/40). Suites: tsc/eslint clean, unit 35/35, full e2e 35 passed / 10 skipped / 0 failed (all projects, `--workers=2`; at default full parallelism two WebGL tests hit the recorded ReadPixels-contention timeouts and pass on re-run). The report's remediation appendix has the finding-by-finding disposition; critique iteration 2 added two Phase-4 open questions (the desktop national frame cropping Thailand's tips; the ใกล้ฉัน sample-station disclosure).

**Phase 4 (2026-10-30 ~14:30–15:05, cut at the lead's hard stop):** four units landed and committed — the 2D MAP / 3D TERRAIN view modes (sequenced projection switching, mercator gets the real fog back, verified in-browser both directions), the loopless retirement of Noto Serif Thai (ADR 0005 addendum; sentences and wordmark in Mahanathi Sans, hierarchy by size/leading/tone), the English-instrument/Thai-voice copy rule (station labels + ICT timestamps, the datum caveat stays Thai), and DEMO.md + the HII·RID credit. Not landed (report's honest list): the station deep-link and loopless raster guard (P2), the about section (P3, cut per the prompt), critique iteration 2's two P1s. Suites: tsc/eslint clean, unit 35/35, gate spec 4/4, desktop e2e 14/15 (the one failure the known ReadPixels-contention class, passing isolated). The report: `process/reports/phase-4.md`.

## How to check the current state

```sh
make install   # pnpm install --frozen-lockfile (Go has no dependencies yet)
make lint      # go vet, gofmt, tsc, eslint, prettier
make test      # go test -race, vitest
make build     # web build, embed, go build to server/bin/mahanathi
make e2e       # Playwright at 1440×900, 768×1024 and 390×844
make docker    # image mahanathi:dev (streams git-known files; see ADR 0001)
```
