# Progress

Updated at the end of every phase. The detail lives in `process/reports/phase-N.md`, and the plan and risks in `docs/PLAN.md`.

## Phases

| # | Phase | Status | Gate | Report |
|---|---|---|---|---|
| 1 | Discovery and foundation | **done**, reviewed by the lead, closed out in Phase 2's Step 0 | research with sources ✓, ADRs for stack, storage and data reference ✓ (plus rendering), scaffold builds and tests pass ✓, CI defined ✓ | `process/reports/phase-1.md` |
| 2 | Data core (TDD) | **done**, awaiting the lead's review | pipeline live-verified on 804 stations ✓, oracle 789/789 ✓, timeline within budget (46 KB gzipped of 80) ✓, humanize coverage 100 % ✓, deploy image builds and self-healthchecks ✓ | `process/reports/phase-2.md` |
| 3 | Look-dev | not started | | |
| 4 | Product build | not started | | |
| 5 | Polish and hardening | not started | | |
| 6 | Ship docs and handover | not started | | |

## Current state (end of Phase 2, 2026-10-29)

**What exists:**
- **Server:** the whole data pipeline runs from one binary. `thaiwater` (client with fixtures), `normalize` (ADR 0003 arithmetic, mm-exact), `domain` (levels, percents, distances), `humanize` (the Thai sentence, 100 % coverage), `store` (SQLite WAL, migrations, retention), `ingest` (hourly at :07, 72 h backfill, oracle), `api` (`/now`, `/history`, `/timeline`, `/readyz`) and `main` wiring with env config and a `healthcheck` subcommand.
- **Web:** the placeholder page the lead asked for — the national sentence and the five most severe stations from `/api/v1/now`, plain semantic HTML, no styling decisions. The real UI is Phase 3.
- **Build:** `make docker` builds the 21.4 MB deploy image (distroless, `/data` for 65532, self-healthcheck); `make docker-full` is the whole pipeline once Phase 3 ships.

**Live evidence (21:57 ICT):** first ingest in 33 s — 804 rows, 251,624 readings, 804/804 histories, oracle 789/789. `/now` 85 KB gzipped; the binary timeline body 46 KB gzipped against the 80 KB budget. Details and the five hand-checked sentences: `process/reports/phase-2.md`.

**Data findings:** finalised in `docs/research/data-quality.md` — three snapshots agree (17:54, 19:05, 21:49 night), the two live suspects are hand-checked logger datum alternation.

**Decisions:** ADR 0001–0004 as before. The timeline wire format (binary MNTL, delta varints) is documented in `docs/api.md`, the contract for Phase 4.

**Model note:** from `23a869f` the implementing session runs on GLM-5.3 (the owner's decision, recorded in the DEVLOG and `process/README.md`); everything before is Opus 5.5 at xhigh.

## How to check the current state

```sh
make install   # pnpm install --frozen-lockfile (Go has no dependencies yet)
make lint      # go vet, gofmt, tsc, eslint, prettier
make test      # go test -race, vitest
make build     # web build, embed, go build to server/bin/mahanathi
make e2e       # Playwright at 1440×900 and 390×844
make docker    # image mahanathi:dev (streams git-known files; see ADR 0001)
```
