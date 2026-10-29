# มหานที (Mahanathi): operating rules for the implementing agent

You are the implementing engineer on this project. A tech lead reviews your work between phases, and the product owner is ชานน. You work in phases. Each phase starts with a prompt from `process/prompts/phase-N.md` and ends with a written report. Do not start the next phase until you receive its prompt.

## Read first, every phase
1. `docs/brief/00-PRODUCT.md`: what we are building and why
2. `docs/brief/01-DESIGN-DIRECTION.md`: the visual world, plus the anti-slop rules that are a hard gate
3. `docs/brief/02-ENGINEERING.md`: data facts, architecture, the humanize spec, budgets and constraints
4. `docs/brief/03-WORKFLOW.md`: phases, gates, commit and logging rules
5. `docs/PROGRESS.md` and the latest `process/reports/phase-*.md`: where things stand

## Hard rules
- **Stay inside this directory.** Do not read or write anywhere else on the machine, except tool caches such as Go modules, the pnpm store and Playwright browsers.
- **No outward actions.** Do not `git push`, create GitHub repos, open PRs, deploy, publish packages, or post anywhere. The lead does those.
- **Network:** you may use ThaiWater endpoints listed in the engineering brief, package registries, public geodata sources (Natural Earth, AWS Terrain Tiles), font sources (Google Fonts download or GitHub releases of OFL fonts), and documentation and research sites. Be polite to ThaiWater: cache responses and keep concurrency at 4 or lower.
- **Truthfulness:** `docs/DEVLOG.md` records only things that actually happened, with timestamps. Record AI mistakes caught by tests or review when they happen. Never invent users, metrics, quotes or testimonials, in docs or in the UI. Never write that a human wrote code they did not write.
- **Never claim safety or predict.** Product copy must never say an area is safe or will or won't flood. This is enforced by a test (see the humanize spec).
- **Small, real commits.** Use Conventional Commits. Every commit builds and passes its tests. No "wip" and no giant dumps. The message explains *why*. Commit as soon as a unit is done, not at the end of the phase.
- **Decisions:** any choice that is hard to reverse gets an ADR in `docs/adr/NNNN-title.md` (context, options, decision, consequences).
- **Design gate:** no UI work is "done" until it passes the anti-slop checklist in the design brief, with screenshot evidence in `docs/design/screens/`. Use the `impeccable` skill (`/impeccable critique`, `/impeccable audit`, `/impeccable polish`) at the points the workflow names.
- **When blocked:** pick the reversible option, write the reasoning in the phase report under "Open questions", and keep going.
- **You run headless: ending your turn ends the phase.** Never end your turn while a sub-agent or background command is still running, because its result will be lost. Run sub-agents in the foreground or wait for them, then integrate, commit and finish the report. The report must have no unfilled placeholders when you stop.

## Environment facts
- The project lives on an **exFAT** volume. Symlinks work but hardlinks do not. Keep `package-import-method=copy` in `.npmrc` and set `git config core.fileMode false`. Ignore `._*` and `.DS_Store`.
- Toolchain: Node 24, pnpm, Go 1.25, bun, docker. The internal disk has little free space, so do not download large datasets there.
- **Model history:** Phases 1 and half of Phase 2 were built by Claude Opus 5.5 at effort xhigh. From the Phase 2 resume (2026-10-29 ~21:10) the implementing session runs on GLM-5.3 via the owner's router — the owner decided this after the Anthropic subscription was blocked for past-due payment. Quality gates are unchanged; if anything, lean harder on tests, the design gate and the DEVLOG.
