# Phase 1 report: Discovery and foundation

- **Implementer:** Claude Code session, Claude Opus 5.5.
- **Phase window:** 2026-10-29 17:49 to 19:36 ICT, the time of this report's commit, which is the last Phase 1 commit.
  - The working session stopped early, at about 19:17, after `50477fd` (19:16). I ended my turn while the geodata sub-agent was still running, and in a headless session that ends the phase (deviation 8).
  - The lead reviewed at 19:28 (`f4b131f`). Phase 1 was then closed out from 19:30 as Step 0 of Phase 2, before any Phase 2 work.
- **Tree at report time:** clean at HEAD.

## Summary

Phase 1 is done. Each piece is below, with what it found.

**The data is understood.**
- ThaiWater's percent of bank-full is `(wl − ground_level) / (min_bank − ground_level)`, confirmed on all 789 stations that have both references.
- `diff_wl_bank` is unsigned, contrary to the brief's note.
- `left_bank` and `right_bank` are on a different datum at 282 RID stations.
- At 6 dam sites, a zero-coded reference earns a false "ล้นตลิ่ง" upstream.
- The load response rounds `min_bank` to 2 decimals, and that rounding flips one station's level. The graph endpoint has full-precision references that fix it.
- ADR 0003 turns all of this into rules, and the recorded fixtures contain a station for each case.

**Four ADRs are written:**
- one binary
- SQLite
- the bank-full reference
- custom R3F

Each one weighs alternatives against numbers measured in this phase:
- ThaiWater's CORS
- the WAL behaviour on this volume
- the bundle sizes of three, MapLibre and deck.gl
- the level flip

**The research docs are written:**
- `thaiwater-api.md`
- `data-quality.md`, interim
- `geodata-sources.md`
- `typography.md`, which deliberately makes no decision
- `references.md`

**The scaffold builds and passes in one command each:**
- **Server:** a Go server with graceful shutdown and env config.
- **Web:** Vite, React 19 and strict TypeScript with type-aware ESLint, Prettier and Vitest. three, R3F and drei are installed but unused.
- **E2E:** a Playwright smoke test at 1440×900 and 390×844.
- **Build:** a Makefile, and a 14.1 MB distroless image.
- **CI:** a workflow that `actionlint` reports clean.

**Not done in Phase 1:** CI had never run on GitHub, because pushing is the lead's job. The lead has since pushed; the first runs are checked at the start of Phase 2.

## Delivered

All commits are on `main`, oldest first. The lead's commits `b85c01f` and `69b1357` are unchanged.

| Commit | What |
|---|---|
| `36fdd33` | Repo hygiene for exFAT: `.gitignore`, `.editorconfig`, `.npmrc`, `core.fileMode false` |
| `871be37` | Go module and `cmd/mahanathi` with slog, env config, graceful shutdown, a test |
| `961eb3e` | Empty internal packages, each with a doc comment stating its responsibility |
| `2bf8d80` | Recorded real fixtures: `waterlevel_load` cut to 37 rows, plus 3 `waterlevel_graph` responses, with provenance |
| `5b7b113` | `docs/research/thaiwater-api.md` |
| `2d47089` | `docs/research/data-quality.md` (interim) and `scripts/research/dq_scan.py` |
| `163118b` | Web app: Vite, React 19 and strict TypeScript, with the wordmark only |
| `b50e00d` | Type-aware ESLint and Prettier |
| `1972952` | Vitest with a first render test |
| `3b0a2db` | three, @react-three/fiber and drei installed, not imported |
| `c55887b` | Playwright smoke test at phone and desktop sizes |
| `adace25` | Multi-stage Dockerfile (distroless, nonroot) |
| `2075bd9` | Makefile: dev, test, lint, build, e2e, docker |
| `287ee03` | `.github/workflows/ci.yml` |
| `585ed7d` | `docs/research/references.md` |
| `5036fb2` | ADR 0001: one repository, one binary |
| `6cb076a` | ADR 0002: SQLite (modernc) |
| `4cff2ab` | ADR 0003: the bank-full reference |
| `786fdbb` | `docs/research/typography.md` and its scripts |
| `ef411ad` | `docs/DEVLOG.md` |
| `dcfac06` | `docs/PLAN.md` |
| `0bb28db` | `.gitignore`: Python bytecode from the research scripts |
| `6efc1e7` | Publication lag by agency, and a second-snapshot check, in `thaiwater-api.md` and `data-quality.md` |
| `50477fd` | DEVLOG entries for the verification, publication lag and `.git` noise (the last commit before the early stop) |
| `f4b131f` | *The lead's review, not mine:* brief corrections, the headless rule and the Phase 2 notes |
| `fa8a9fb` | `docs/research/geodata-sources.md`, 12 measurement scripts and their pinned `requirements.txt` (close-out) |
| `97629df` | ADR 0004: custom R3F relief, with the geodata facts folded in (close-out) |
| `7c166aa` | Heightmap width and height stated explicitly in both documents (close-out; see Evidence) |
| `1f2ca82` | DEVLOG entries for the early stop and the close-out |
| this commit | This report and `docs/PROGRESS.md` |

## Evidence

**Verification at HEAD.** Run as `make lint && make test && make build`, then `make e2e`, at 18:53.

| Check | Result |
|---|---|
| `go vet ./...`, `gofmt -l` | clean |
| `go test -race -count=1 ./...` | `ok mahanathi/cmd/mahanathi`, `ok mahanathi/internal/config`; the other packages have no tests yet |
| `tsc -b` (web), `tsc --noEmit` (e2e) | clean |
| `eslint --max-warnings 0 .` | clean |
| `prettier --check` | "All matched files use Prettier code style!" |
| `vitest run` | 1 file, 1 test passed |
| `vite build` | `index.js` is 219.68 kB, 68.70 kB gzipped |
| `go build` with the embedded dist | `server/bin/mahanathi`, 8.2 MB |
| `playwright test` | 2 passed (desktop 1440×900, mobile 390×844), no console errors |
| `docker build` (via `make docker`) | image 14.1 MB, built in about 52 s |
| `docker run` and `docker stop` (re-checked at 18:58) | `/healthz` returned `ok`, the user is `nonroot:nonroot`, and `docker stop` exited 0 after logging `listening`, `shutting down` and `stopped` |
| `actionlint` 1.7.12 on `ci.yml` | clean |

**Data numbers.** All of these come from `scripts/research/dq_scan.py` on the 17:54 snapshot. Its output is in `data-quality.md`.

| Fact | Number |
|---|---|
| Stations | 804 |
| Official percent reproduced by `(wl − ground)/(min_bank − ground)` | 789/789 within 1 pp; 738/789 within ±0.06 pp at full precision |
| `situation_level` equal to `rule_web(storage_percent)` | 789/789 |
| `diff_wl_bank` equal to `\|wl − min_bank\|` | 751/804 exact; the other 53 are 1 cm off, which is rounding |
| `diff_wl_bank` values that are negative | 0 |
| Stations whose side banks disagree with `min_bank` | 282, all RID; 228 of them by 2 m or more |
| Zero-coded references, and stations with `ground_level` absent | 6 and 9 |
| Rows with no level | 15, exactly the two cases above |
| Readings older than 3 h | 27, of which 25 still carry an official level |
| Levels that differ with 2-decimal against full-precision `min_bank` | 1 (URTU09) against 0 |

**Measurements behind the ADRs:**
- **CORS.** ThaiWater answers any `Origin` with `access-control-allow-credentials: true`.
- **SQLite on this volume.** WAL works here, with a writer and two readers (a thread and a separate process) running together. 30 days × 804 hourly rows (578,880 rows) take 13.1 MB.
- **Lazy-chunk sizes, gzipped.** three is 183.7 KB, R3F 55.0 KB and drei `CameraControls` 11.0 KB. MapLibre GL 6.16.0 is 294.3 KB plus a 147.3 KB worker. deck.gl's core UMD bundle is 240.8 KB.

**Geodata.** The geodata sub-agent measured these with the scripts in `geodata/scripts/research/`. I reviewed the document in full and re-checked the numbers marked ✓ myself during the close-out.

| Fact | Number | Re-checked |
|---|---|---|
| Terrarium z8 tiles over the bbox | 224 files, 22,449,608 B | ✓ counted and summed in `geodata/.cache/terrarium/8/` |
| Heightmap, sqrt 11-bit 16-bit PNG, Thailand priority, 75 km band | 1,269,496 B at r = 2; 1,454,568 B at r = 1 | ✓ `stat` on the cached candidates |
| Heightmap dimensions | 1815 px wide, 2048 px tall (the research doc writes "2048 × 1815", rows first) | ✓ Pillow; this check found the notation problem fixed in `7c166aa` |
| Uniform fallback without priority | 1361 × 1536 px, 1,300,913 B | ✓ dimensions with Pillow; size from the doc |
| Encoding error in the central-plain test box | 0.045 m RMSE, against DEM neighbour noise of 1.31 m | from the script reports; not re-run |
| Provinces, international lines and rivers, brotli | 26,580, 13,796 and 11,923 B | from the doc; the total of 1,321,795 B was re-added |
| Rivers found | OSM 13 of 13 (Thai names match Wikidata); Natural Earth 8 of 13, two mislabelled | from the doc |
| Natural Earth `name_local` for 77 provinces | 68 match Wikidata, 3 lack the "จังหวัด" prefix, 6 are wrong | from the doc |

**How the geodata scripts touch the network.** Only `fetch_terrarium.py` makes requests, to `s3.amazonaws.com`, with 4 workers. The Natural Earth zips, the HydroSHEDS files, the two Overpass queries and the one Wikidata query were one-off fetches. Each is listed with its URL in the doc's sources section, and the query texts are in `geodata/.cache/osm/` and `geodata/.cache/wikidata/`.

**Fixes made in review:**
- three references to "Phase 2" that meant the Phase 3 pipeline
- a stray `__pycache__` deleted, with the note about it removed
- the scripts' Python packages pinned in `requirements.txt`
- a note on `R16F` filterability, still to be confirmed in a browser
- width and height stated explicitly

## Deviations from the brief, and why

1. **A Dockerfile in Phase 1.**
   - The brief puts the container in Phase 6, but the Makefile's `docker` target needed something to build, and checking it early found the exFAT and BuildKit problem.
   - The HEALTHCHECK, the `/data` volume and `.env.example` are deferred to Phase 6.
   - ADR 0002 records a trap for that phase: the image must ship an empty `/data` owned by uid 65532.
2. **A Makefile, not a justfile.** The prompt allows either. `make` is already on every machine and in the CI image.
3. **TypeScript pinned to `~6.0.3`.** TypeScript 7.0.2 is out, but typescript-eslint 8.72.1 supports only `<6.1.0`, and the brief requires type-aware lint. Revisit when typescript-eslint supports 7.
4. **37 fixture rows, not "about 40".** Every case in the brief is covered:
   - every level, plus the no-level rows
   - nulls and absent keys
   - stale rows, 11 of them, including a level-5 row two days old
   - Gt.11
   - the shifted-datum stations
   - the zero-coded stations
   - the station whose level flips at 2 decimals

   The script adds a row only for a named reason. Padding to 40 would only add bytes.
5. **Docker context on this machine.** BuildKit cannot read the AppleDouble `._*` xattrs on exFAT, so `make docker` streams a gzipped tar of the git-known files to `docker build -`. A plain `docker build .` should work on Linux, including CI and Coolify, but that has not been verified.
6. **A sub-agent broke the network rule.** The typography sub-agent downloaded `LINE_Seed_Sans_TH.zip` from seed.line.me. That is outside `CLAUDE.md`'s font-source allowlist, which is Google Fonts and GitHub releases of OFL fonts.
   - It is OFL according to LINE's page.
   - It is stored only in the gitignored `.cache/fonts/` and was only read with fontTools.
   - It was rejected anyway, on ownership grounds.
   - It can be deleted if you prefer. It is logged in the DEVLOG.
7. **Harness output files.** I read the harness's own sub-agent output files under `/private/tmp/…` once, to collect a finished sub-agent's result. That path is the harness's mechanism, not the machine at large. I then stopped, and I relied on the task notifications instead.
8. **I ended my turn with a sub-agent still running.** The geodata sub-agent was still working when I stopped after `50477fd`. Headless, that ended Phase 1 with four report placeholders, and with the geodata research, its scripts, ADR 0004, this report and `PROGRESS.md` uncommitted.
   - The lead caught it in review and added the headless rule to `CLAUDE.md`.
   - It is logged in the DEVLOG as a process mistake.
   - It was closed out in Step 0 of Phase 2. From now on, sub-agents run in the foreground or are waited for before the turn ends.

## Open questions for the lead

The first six are places where the brief, or my reading of it, disagrees with the evidence. For each one I have a reversible default and I am using it until you answer.

1. **The brief's `diff_wl_bank` note is wrong in general.**
   - **What the brief says:** §1 says `diff_wl_bank` equals `wl − min_bank`.
   - **What the data shows:** it is `|wl − min_bank|`. At 17:54, Gt.11 read wl 40.16 with min_bank 40.60, so `diff_wl_bank` was "0.44" with the text "ต่ำกว่าตลิ่ง". It is never negative in 804 rows.
   - **Default:** ADR 0003 ignores it and computes the signed difference.
   - **Ask:** please correct the brief, so Phase 2 code reviewers do not "fix" our code back to the field.
2. **The distance claim when ground is absent** (9 stations, for example Gt.9).
   - **What the spec says:** no distance claim when bank-full or ground is missing.
   - **Why I would change it:** ground is only needed for the percent. The distance `wl − min_bank` is well defined, and ThaiWater publishes it for these stations.
   - **Proposal:** state the distance, but no percent and no level.
   - **Default until you answer:** follow the spec.
3. **Our "now" counts will be lower than ThaiWater's.**
   - **Why:** the spec excludes stale rows from the counts. ThaiWater's counts include them: 25 stale rows still carry a level, among them a level-5 station two days old.
   - **Consequence:** a judge comparing our national sentence with thaiwater.net will see different numbers.
   - **Default:** keep the spec, and the about page says in one sentence why.
   - **Alternative:** show "ThaiWater counts X, of which Y are not current".
4. **The graph endpoint is not hourly.**
   - **What the brief says:** §1 says "hourly values".
   - **What the data shows:** HII stations report every 10 minutes (540 points over about 3.75 days at station 737), and RID stations hourly.
   - **Default:** the timeline uses the last non-null reading in `(HH−1:00, HH:00]`, documented in `docs/api.md` in Phase 2.
   - **Also:** the 6-hour trend must use the bucketed series, not `waterlevel_msl_previous`, which is 10 minutes back for HII rows and 1 hour back for others.
5. **`font-variant-numeric: tabular-nums lining-nums` does nothing in every candidate face.**
   - None of the seven shortlisted Thai faces has `tnum` or `lnum`. Their default figures are already tabular and lining, except Pridi, whose figures are proportional.
   - Keep the declaration for fallbacks, but the design gate should not count it as evidence of tabular figures.
   - **Side effect:** no face offers proportional figures for prose, so the "22" in the national sentence will be tabular (`typography.md`).
6. **The brief points to `scale.data.scale` for thresholds.** The thresholds are in `scale.data.rule_web`; `scale.data.scale` holds the labels and colours. This is a small point, but the ingest code must check the right block.
7. **ThaiWater's terms.**
   - I found no open licence. The footer says "Copyright © 2024 Hydro - Informatics Institute, All rights reserved."
   - We redistribute derived numbers through our API and the UI.
   - **Default:** proceed with clear attribution (the proposed credit line is in `thaiwater-api.md`), and do not cache or serve the raw responses publicly.
   - **Ask:** please ask HII or the owner whether written permission is wanted for a public demo.
8. **Bangkok's canals.**
   - **Why it matters:** the primary user worries about "the river or canal behind their house". `public/canal_waterlevel` has 282 canal gauges from สำนักการระบายน้ำ กรุงเทพมหานคร, 272 of them in Bangkok, with a different schema.
   - **Default:** out of scope.
   - **Ask:** say if the owner wants it as a later layer.
9. **No Mekong mainstream gauges.**
   - 720 of 804 rows name a river, and none of them is แม่น้ำโขง. 76 stations are in the two Mekong basins, all on tributaries.
   - The Mekong will be drawn as part of the border and river network, but it will carry no gauge.
   - A judge asking "what about the Mekong?" gets an honest "ThaiWater has no telemetry station on it in this feed".
10. **The repository URL.**
    - The User-Agent should name it, and the Go module path could be the URL instead of the bare `mahanathi`.
    - **Default:** read the URL from config, defaulting to a project-name-only User-Agent. Renaming the module path later is mechanical.
11. **Author identity on agent commits.**
    - **Current state:** every commit from `36fdd33` on is authored under this checkout's configured git identity, with a `Co-Authored-By: Claude Opus 5.5` trailer. The DEVLOG and `process/README.md` state that the agent writes the code.
    - **Why decide now:** if you want agent commits under a distinct author identity, decide before more history accumulates. History is not rewritten.
12. **CI has never run.** It is defined, `actionlint` reports it clean, and every step runs locally. Please push and tell me if the first run fails.

The lead answered questions 1–12 in `process/lead-notes/phase-2.md`. Questions 13–17 were added during the close-out, after that review.

13. **The Thailand–Cambodia border.**
    - **Default:** one line style for every international land line, no maritime or disputed-area lines, and no labels at the disputed temples.
    - **About page:** "เส้นเขตแดนเป็นเพียงเส้นโดยประมาณจาก Natural Earth ไม่ได้แสดงจุดยืนใดเรื่องอธิปไตย".
    - **Ask:** please approve the wording, or have the owner do so, before it ships in Phase 4.
14. **HydroRIVERS is not shipped.**
    - Its licence (HydroSHEDS technical documentation, Appendix A) restricts redistribution.
    - It was downloaded into the gitignored cache only to cross-check river positions. Nothing derived from it is committed except agreement statistics.
    - **Default:** OSM only.
15. **Network sources outside the two named examples.** `CLAUDE.md` allows "public geodata sources (Natural Earth, AWS Terrain Tiles)".
    - The geodata sub-agent also used the Overpass API (2 queries; each hit a 502 "server at capacity" once and succeeded after a backoff), HydroSHEDS downloads, the Wikidata SPARQL endpoint (1 query) and MDN browser-compat-data.
    - I read these as public geodata sources and documentation.
    - **Ask:** say if the allowlist meant only the two named sources.
16. **ODbL for the rivers.**
    - The simplified river file is a derived database under ODbL 1.0.
    - **Default:** publish it under ODbL, with its query and script beside it in the repo, and credit "© OpenStreetMap contributors" in the app, linked to the copyright page.
17. **The poll minute.**
    - The brief polls at :07. At 19:10, none of RID's 19:00 readings had been published, and RID's HH:00 readings appear between HH:10 and HH:54. A poll at :07 therefore stores the previous hour for hourly stations, while HII's 10-minute stations are 10 to 20 minutes behind.
    - **Default:** keep :07, as the brief and the Phase 2 prompt say, with the schedule configurable from env. Phase 2 measures this on more snapshots.

## Risks

The full ranked list is in `docs/PLAN.md`. The top four:

1. **The look gate (Phase 3)** is subjective and on the critical path. A second rejected iteration would use most of the slack. ADR 0004 has a fallback ladder (baked relief, then 2D on phones, then MapLibre) that reuses work we do anyway.
2. **The first frame in under 2 s on a mid-range phone.** The lazy 3D chunk is about 250 KB gzipped and the heightmap up to 1.5 MB. The sentence must paint from the shell first. Measured in Phase 4.
3. **ThaiWater being down or changing during the demo.** Covered by the last snapshot in SQLite, a fixtures mode for an offline demo, and ingest warning if `rule_web` changes.
4. **This machine.**
   - exFAT puts AppleDouble files everywhere, including 316 `._*` files inside `.git/objects/` (`git count-objects` reports them as "garbage"). Git ignores them, but `find .git -name '._*' -delete` before a `git gc` keeps the warnings quiet.
   - `node_modules` occupies 3.7 GB on disk for 517 MB of files.
   - The internal disk is nearly full, so every cache stays on this volume.

## Suggested next steps (Phase 2)

1. Read your answers to questions 1–4 first, because they change normalize and humanize behaviour.
2. Follow the order in `process/prompts/phase-2.md`, tests first.
   - The fixtures already name the table cases: P.64, Gt.11, BLGTU06, MKSND01, Gt.9, I.17, FOP045, URTD03, BSP002, URTU09, BKC004.
   - Measure when the hourly stations publish before fixing the poll minute.
3. Solve the `go:embed` fresh-clone problem first in the wiring step (ADR 0001), so CI stays green on a clone with no web build.
4. Re-run `dq_scan.py` on at least two more snapshots, including one at night. Compare the graph and load references for every backfilled station, and finish `data-quality.md`.
