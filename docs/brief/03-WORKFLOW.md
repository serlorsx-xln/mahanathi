# Workflow: how this project is built

## Roles
- **Product owner:** ชานน. Sets goals and is the final say on direction.
- **Tech lead / reviewer:** the orchestrating Claude session. Writes briefs and phase prompts, reviews every phase report, screenshots and diff, and performs outward actions (GitHub, deploy).
- **Implementer:** this Claude Code session (Opus 5.5, effort xhigh), resumed once per phase.

## Phase loop (every phase)
1. **Orient:** re-read the briefs, `docs/PROGRESS.md` and the last report.
2. **Plan:** append the phase plan to `docs/PLAN.md`. List the tasks, the order and how each one will be verified.
3. **Build in small commits:** use Conventional Commits (`feat(humanize): …`, `test(api): …`, `docs(adr): …`). Each commit builds and passes its tests.
4. **Verify:** run the tests and linters. For UI work, run the design gate (screenshots, then `/impeccable critique`, then the checklist, then fix and re-shoot).
5. **Log:**
   - Update `docs/PROGRESS.md` with the phase table and the current state.
   - Append real events to `docs/DEVLOG.md`: decisions, surprises, and AI mistakes caught, with the test that caught each one.
6. **Report:** write `process/reports/phase-N.md` with these sections:
   - Summary
   - Delivered (with commit hashes)
   - Evidence (test output summary, screenshot paths)
   - Deviations from brief and why
   - Open questions for the lead
   - Risks
   - Suggested next steps
7. **Stop.** Wait for the next prompt.

## Phases
| # | Name | Goal | Gate (must be true to pass) |
|---|---|---|---|
| 1 | Discovery and foundation | Research, ADRs, scaffold, tooling, CI, plan | Research docs exist with sources. ADRs are written for stack, data reference and storage. Scaffold builds, an empty test passes, and CI is defined. |
| 2 | Data core (TDD) | ThaiWater client, normalize, humanize, store, ingest and backfill, API, minimal page | `go test ./...` is green and humanize/normalize coverage is 90% or more. Live ingest has run locally. The data-quality report is written. `/api/v1/now` serves real data. |
| 3 | Look-dev | Tokens, type specimen, colour scale, geodata pipeline, relief prototype | Screenshots pass the anti-slop gate. The lead approves the look before Phase 4. |
| 4 | Product build | Live data in 3D, ใกล้ฉัน, search, panel, timeline, states, fallbacks | E2E is green. The gate is re-run on the real product. Performance budgets are measured. |
| 5 | Polish and hardening | impeccable polish/audit, a11y, mobile, LINE, copy, OG/meta, headers, rate limit | Lighthouse a11y is 95 or more and performance is 85 or more (mobile). The audit has no P0/P1 issues. |
| 6 | Ship docs | Dockerfile, README, ARCHITECTURE, REVIEW guide for the owner, demo script | The container runs locally, the docs are complete and honest, and the owner's review guide lists the files to read line by line. |

## Commit history is part of the product
The history should read like a careful engineer's: steady and small, with the tests landing alongside (or before) the code, and messages that explain intent. Do not squash and do not rewrite history.

## Honesty about AI
- This project is built by an AI agent under the direction of the owner and a lead. The README states this plainly, along with the process that was used.
- The DEVLOG records real AI mistakes caught by tests or review. It is evidence that the process works, so never pad it.
- In Phase 6, write `docs/REVIEW.md`. It lists the code the owner must personally read and understand before presenting (humanize, normalize, the stale and suspect logic, the timeline encoding and the rate limiter), with the questions a judge might ask about each.
