# Progress

Updated at the end of every phase. The detail lives in `process/reports/phase-N.md`, and the plan and risks in `docs/PLAN.md`.

## Phases

| # | Phase | Status | Gate | Report |
|---|---|---|---|---|
| 1 | Discovery and foundation | **done**, reviewed by the lead, closed out in Phase 2's Step 0 | research with sources ✓, ADRs for stack, storage and data reference ✓ (plus rendering), scaffold builds and tests pass ✓, CI defined ✓ | `process/reports/phase-1.md` |
| 2 | Data core (TDD) | **in progress** | | |
| 3 | Look-dev | not started | | |
| 4 | Product build | not started | | |
| 5 | Polish and hardening | not started | | |
| 6 | Ship docs and handover | not started | | |

## Current state (start of Phase 2, 2026-10-29)

**What exists:**
- **Server:** `cmd/mahanathi` serves only `/healthz`, with env config, slog and graceful shutdown. The internal packages exist as documented empty shells.
- **Web:** a blank page with the wordmark. The 3D libraries are installed but not imported.
- **E2E:** a Playwright smoke test.
- **Build:** a Makefile, a Dockerfile (a 14.1 MB distroless image) and a CI workflow.

**Data findings:**
- ThaiWater is understood, with recorded fixtures in `server/internal/thaiwater/testdata/` and the evidence in `docs/research/data-quality.md`.
- Bank-full is `min_bank` and ground is `ground_level`; `left_bank` and `right_bank` are never read (ADR 0003).

**Decisions:**
- ADR 0001: one binary that embeds the web build.
- ADR 0002: SQLite through modernc.
- ADR 0003: the bank-full reference.
- ADR 0004: custom R3F relief, with a fallback ladder and the geodata facts it depends on.

**Answered by the lead** (`process/lead-notes/phase-2.md`): questions 1–12 of the Phase 1 report. Questions 13–17, added at the close-out, are still open.

## How to check the current state

```sh
make install   # pnpm install --frozen-lockfile (Go has no dependencies yet)
make lint      # go vet, gofmt, tsc, eslint, prettier
make test      # go test -race, vitest
make build     # web build, embed, go build to server/bin/mahanathi
make e2e       # Playwright at 1440×900 and 390×844
make docker    # image mahanathi:dev (streams git-known files; see ADR 0001)
```
