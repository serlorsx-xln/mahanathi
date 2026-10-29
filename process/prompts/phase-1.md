# Phase 1: Discovery and foundation

You are the implementing engineer on มหานที (Mahanathi). The briefs in `docs/brief/` were written by the tech lead and are your source of truth. Read all four, plus `CLAUDE.md`, before doing anything else.

**Goal of this phase:** de-risk the project and lay a foundation a senior engineer would be proud of. No product UI yet.

## Tasks

### 1. Research (write each as `docs/research/<topic>.md` with dated sources)
- **`thaiwater-api.md`**
  - Re-verify both endpoints in the engineering brief and document the fields you will use, their types, nulls and units.
  - Find the official attribution and terms of use for ThaiWater/HII data.
  - Check whether better public endpoints exist for station metadata or rainfall. Note them and do not use them yet.
  - Save trimmed real responses as fixtures under `server/internal/thaiwater/testdata/`: one `waterlevel_load` subset of about 40 diverse stations covering every situation level, nulls, a stale row and the Gt.11 anomaly, plus 3 `waterlevel_graph` responses.
- **`data-quality.md`**
  - Start the investigation of `min_bank`, `left_bank`, `right_bank`, `ground_level`, `storage_percent` and `diff_wl_bank` across all ~800 stations, with a small script and its numbers.
  - The formula should be confirmed by evidence, not assumed.
  - Phase 2 finishes this, so put your interim conclusions here.
- **`geodata-sources.md`**
  - Choose the DEM source (AWS Terrain Tiles, Terrarium), the zoom level and the bbox for Thailand plus its neighbours. Estimate the heightmap resolution and file size.
  - Choose border and river sources, with their licences. Find out whether Natural Earth rivers are detailed enough for Thailand's main rivers (เจ้าพระยา, ปิง, วัง, ยม, น่าน, ป่าสัก, ท่าจีน, แม่กลอง, มูล, ชี, โขง, ตาปี, บางปะกง) and how to label them in Thai.
- **`typography.md`**
  - Shortlist Thai and Latin OFL pairings per the design brief and record where to download each one.
  - Do not decide yet. The specimen in Phase 3 decides.
- **`references.md`**
  - Study the design references named in the design brief, plus 3–5 more premium 3D, cartographic or data pieces you find, such as Reuters, NYT or Pudding terrain and flood graphics.
  - For each, write one paragraph on *what specifically makes it premium* (type, light, camera, restraint, annotation) and how it applies here.
  - Also write a short section on 3D slop patterns you observed in the wild, beyond the brief's list.

### 2. Decisions (ADRs in `docs/adr/`)
- `0001-monorepo-single-binary.md`: Go binary embedding the web build. Consider the alternatives.
- `0002-storage-sqlite.md`
- `0003-bank-full-reference.md`: which fields define bank-full and ground, based on your data-quality evidence so far.
- `0004-rendering-r3f-custom-relief.md`: a custom R3F relief compared with MapLibre/deck.gl.
  - Decide on quality, performance and the demo goals.
  - Include a fallback plan if the custom relief fails the look-dev gate.

### 3. Scaffold and tooling (one concern per commit)
- Initialise git, run `git config core.fileMode false`, and add `.gitignore`, `.editorconfig` and `.npmrc` (`package-import-method=copy`).
- **`server/`:**
  - Go module, `cmd/mahanathi/main.go` with graceful shutdown, `slog` and config from env.
  - Empty internal packages, each with a doc comment.
  - `go vet` and one trivial test.
- **`web/`:**
  - Vite + React 19 + TS strict, with ESLint and Prettier.
  - Vitest with one trivial test.
  - three, @react-three/fiber and drei installed but not used yet.
  - A blank page containing the wordmark text only. No styling decisions yet, because tokens come in Phase 3.
- `e2e/`: a Playwright config and one smoke test that the page loads.
- **Root:**
  - A `Makefile` or `justfile` with `dev`, `test`, `lint`, `build` and `docker` targets.
  - `.github/workflows/ci.yml` running Go tests and the web typecheck, test and build.
- `docs/PLAN.md`: the task breakdown for phases 2–6, with your estimates and the riskiest items called out.
- Create `docs/PROGRESS.md` and `docs/DEVLOG.md`.

### 4. Report
Write `process/reports/phase-1.md` per the workflow. In **Open questions**, list anything in the briefs you disagree with and why. The lead wants pushback where you have evidence.

## Constraints and reminders
- Follow `CLAUDE.md` hard rules: stay in this directory, no pushing or deploying, and make small truthful commits.
- The first commit in this repo was made by the lead and contains the briefs and process files. Keep it as is.
- Make commits as you go. A phase with one big commit at the end fails review.
- Stop after the report. Do not start Phase 2.
