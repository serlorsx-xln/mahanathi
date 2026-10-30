# DEVLOG

This log records real events only. Times are ICT (UTC+7). Exact times come from commits or file timestamps. A `~` marks a time reconstructed to the nearest few minutes from the surrounding commits.

**Authorship:**
- The implementing agent is a Claude Code session (Claude Opus 5.5). It wrote everything from commit `36fdd33` on.
- Those commits use this checkout's configured git identity as author, and each one ends with a `Co-Authored-By: Claude Opus 5.5` trailer. No human wrote that code.
- The lead's commits are `b85c01f`, `69b1357` and `f4b131f`.
- Three research documents were drafted by sub-agents that the implementing agent started (also Claude): `references.md`, `typography.md` and `geodata-sources.md`. The implementing agent reviewed each one before committing it. Those sub-agents' mistakes are logged here like any other.

**Entry tags:**
- **AI mistake:** something the agent got wrong, with what caught it.
- **Surprise:** a fact that changed a plan.
- **Decision:** a choice, with a pointer to its ADR or document.

## 2026-10-29, Phase 1 (discovery and foundation)

- **17:49 Process.** The lead committed the briefs, the phase prompts and the process scripts (`b85c01f`, `69b1357`). Phase 1 started straight after, with a read of all four briefs and `CLAUDE.md`.

- **17:52 Decision.** Repo hygiene for exFAT:
  - `core.fileMode false`.
  - `.npmrc` with `package-import-method=copy`.
  - `.gitignore` additions, including a root `.cache/` for raw downloads.

- **17:54 Data.** Recorded one full `waterlevel_load` snapshot: 804 stations, 1,395,146 bytes, and 241,911 bytes gzipped on the wire.

- **17:57 Data.** Recorded three `waterlevel_graph` responses: 737, 1097520 and 575568.

- **~17:58 AI mistake.** The first run of `scripts/research/dq_scan.py` crashed with `KeyError: 'situation_level'`.
  - The script assumed the key was always present, but it is absent on 15 rows rather than null.
  - **Caught by:** running the script.
  - **Fix:** `.get`. The absence itself became a documented finding.

- **17:58 Surprise.** The brief's Gt.11 note says `diff_wl_bank` equals `wl − min_bank`.
  - Across all 804 rows, `diff_wl_bank` is the **absolute** value. Gt.11 matched at 14:00 only because the water was then above the bank.
  - Code that trusted the sign would call every station overflowing.
  - **Written up in:** `docs/research/data-quality.md`.

- **17:58 Surprise.** `left_bank` and `right_bank` disagree with `min_bank` at 282 stations, all of them RID, and at 228 of them by 2 m or more.
  - `storage_percent = (wl − ground)/(min_bank − ground)` reproduces all 789 published percents within 1 pp.
  - **Led to:** ADR 0003.

- **~18:03 AI mistake.** `make_fixtures.py` stopped with "named station TTC06 not in snapshot".
  - The agent had keyed the station by its display code. Its `tele_station_oldcode` is `THA006`.
  - **Caught by:** the script's own guard, which refuses a named case that is missing.

- **~18:06 AI mistake.** The first fixture annotations were wrong for several rows:
  - Y.16 was labelled level 2. It is level 4.
  - T.1 was described without its level. It is level 5.
  - C.12 was called a canal. It is the Chao Phraya at Samsen.
  - A rule meant to pick a Mekong station matched nothing.
  - **Caught by:** reviewing the coverage table once the script printed each row's actual `situation_level` next to its label.
  - **Fix:** wrong labels corrected, the Mekong rule removed, and the level column kept in the printed table (commit `2bf8d80`).

- **~18:12 AI mistake.** A zsh loop that sampled four other public endpoints used `path` as its variable name.
  - In zsh, `path` is tied to `PATH`, so `curl` and `sleep` stopped resolving.
  - **Caught by:** "command not found".
  - **Fix:** the variable was renamed. Each endpoint was fetched once (18:12 to 18:13), to document it, not to use it.

- **18:13 Commit.** Fixtures committed (`2bf8d80`): 37 rows, covering every level, nulls, stale rows and Gt.11.

- **~18:16 AI mistake.** In a first draft of `data-quality.md`, the agent wrote that RID "added the datum twice" as if it were established.
  - The API cannot show that.
  - **Caught by:** the agent's own re-read before committing.
  - **Fix:** the claim was restated as "consistent with a datum added twice", with the test that supports it.

- **18:15 and 18:17 Commits.** Committed `thaiwater-api.md` (`5b7b113`) and the interim `data-quality.md` (`2d47089`).

- **~18:18 Decision.** TypeScript pinned to `~6.0.3`.
  - TypeScript 7.0.2 is published, but typescript-eslint 8.72.1 declares a peer range of `<6.1.0`.

- **~18:19 Surprise.** On this exFAT volume, `node_modules` is 517 MB of files that occupy 3.7 GB on disk, across about 26,400 files. The cause is the volume's 128 KiB allocation unit.

- **~18:20 AI mistake.** Three lint and test tools tripped over the AppleDouble `._*` files that macOS writes next to every file on exFAT:
  - **ESLint** parsed them.
  - **Prettier** read `.prettierignore` only from its working directory.
  - **Vitest:** setting `exclude` replaced Vitest's defaults instead of extending them.
  - **Caught by:** lint and test failures.
  - **Fixes:** a global ignore, `--ignore-path ../.prettierignore`, and `[...configDefaults.exclude, '**/._*']`.

- **18:23 Commit.** Playwright smoke test committed (`c55887b`). It passed at 1440×900 and at 390×844.

- **~18:24 Surprise.** `docker build .` failed with `failed to xattr ._.cache: operation not permitted`. BuildKit cannot read the extended attributes of AppleDouble files on exFAT.
  - **First fix:** stream the git-known files as a tar on stdin. Docker then read the plain tar stream as a Dockerfile.
  - **Final fix:** gzip the stream.
  - **Result:** the image builds in about 52 s, is 14.1 MB, runs as `nonroot` and answers `/healthz` (`adace25`).

- **~18:26 AI mistake.** `make help` did not list the `e2e` target. The agent's target regex had no digits.
  - **Caught by:** reading the help output.

- **18:28 Commit.** CI workflow committed (`287ee03`). `actionlint` 1.7.12 reports it clean. It has **not** run on GitHub, because pushing is the lead's job.

- **~18:28 Surprise.** ThaiWater's API reflects any `Origin` with `access-control-allow-credentials: true`, so a browser could call it directly. This was considered and rejected in ADR 0001.

- **~18:34 AI mistake.** The first SQLite WAL probe on exFAT reported a reader error.
  - The cause was a bug in the probe: one Python `sqlite3` connection was shared across threads, which Python does not allow.
  - **Caught by:** the error message.
  - **Fix:** the probe was re-run with one connection per thread plus a separate reader process. WAL worked, with no errors and monotonic reads (ADR 0002).

- **~18:36 Surprise.** The 2-decimal `min_bank` in the load response flips one station's level. URTU09 computes to 69.994 %, which is level 3, while ThaiWater publishes 70.04 %, level 4. This is why ADR 0003 prefers full-precision references.

- **~18:39 Measurement.** Gzipped sizes, for ADR 0004:
  - **Throwaway Vite build:** three 183.7 KB, `@react-three/fiber` 55.0 KB, drei `CameraControls` 11.0 KB, and the React shell 68.6 KB.
  - **Published bundles:** MapLibre GL 6.16.0 is 294.3 KB plus a 147.3 KB worker. The deck.gl 9.4.1 core UMD bundle is 240.8 KB.

- **~18:40 AI mistake (sub-agent).** The typography sub-agent assumed Bai Jamjuree has loopless Thai. It does not.
  - **Caught by:** its own counter-winding loop test and a render.

- **~18:40 Rule deviation (sub-agent).** The typography sub-agent downloaded `LINE_Seed_Sans_TH.zip` from seed.line.me.
  - `CLAUDE.md` allows font downloads only from Google Fonts or GitHub releases of OFL fonts.
  - The file is in the gitignored `.cache/fonts/` and was only read with fontTools.
  - Reported to the lead in the Phase 1 report.

- **~18:40 Sub-agent housekeeping.** While cleaning up, the same sub-agent also deleted other files' AppleDouble `._*` files. They are gitignored metadata that macOS recreates, so nothing was lost.

- **18:41 Commit.** `typography.md` committed with its scripts and pinned tool versions (`786fdbb`).

- **18:53 Verification.** Every check passed at HEAD:
  - `make lint`, `make test` and `make build`.
  - `make e2e`, with 2 passed.
  - The output is saved in the gitignored `.cache/verify-phase1.txt` and `.cache/verify-e2e.txt`.

- **~18:55 Surprise.** `git count-objects` reported "garbage found" for 316 AppleDouble `._*` files inside `.git/objects/`, which macOS writes on exFAT.
  - Git ignores files whose names are not valid objects, so this is noise.
  - Reported to the lead because `git gc` and `fsck` will warn about them.

- **18:58 Verification.** The container check was re-run, so the report would not rest on an earlier, less complete check:
  - `docker run` answered `/healthz` with `ok`.
  - The user is `nonroot:nonroot`.
  - `docker stop` exited 0 after logging `shutting down` and `stopped`.

- **19:00 Decision.** Added `__pycache__/` to `.gitignore` (`0bb28db`). The geodata research scripts import each other, so Python writes bytecode next to them even under `-I`.

- **19:05 and 19:10 Data.** Two more `waterlevel_load` fetches, to measure publication lag.
  - None of RID's 19:00 readings was out by 19:10, while 217 of 315 17:00 readings had been out by 17:54. A poll at :07 therefore stores the previous hour for hourly stations.
  - A re-run of `dq_scan.py` on the 19:05 snapshot reproduced every Phase 1 finding.
  - Committed as `6efc1e7`.

- **~19:17 Process mistake.** The agent ended its turn while the geodata sub-agent was still running.
  - The session runs headless, so ending the turn ended Phase 1.
  - It left `geodata-sources.md`, its scripts, ADR 0004, the report and `PROGRESS.md` uncommitted, and the report still had four placeholders.
  - The sub-agent's last file write was at 19:17:44 (`osm_rivers_report.txt`).
  - **Caught by:** the lead's review (`f4b131f`, 19:28), which added the headless rule to `CLAUDE.md`.
  - **Fix:** Phase 1 was closed in Step 0 of Phase 2, before any Phase 2 work. From now on, sub-agents run in the foreground or are waited for.

- **19:28 Process.** The lead reviewed Phase 1 (`f4b131f`): brief corrections, answers to questions 1–12 and the Phase 2 notes.

## 2026-10-29, Phase 1 close-out (Step 0 of Phase 2)

- **19:30 Commit.** The geodata research and its scripts (`fa8a9fb`), after review.
  - The review fixed three references to "Phase 2" that meant the Phase 3 pipeline, removed a note about a stray `__pycache__` (now deleted), and pinned the scripts' Python packages in `requirements.txt`.

- **19:32 Commit.** ADR 0004 with the geodata facts folded in (`97629df`).

- **~19:34 AI mistake (sub-agent, copied by the agent).** The geodata research writes grid sizes rows first ("2048 × 1815"). The agent copied that into ADR 0004 as if it were width × height.
  - The PNG is 1815 px wide and 2048 px tall. A Phase 3 loader that trusted the ADR would have built the texture transposed.
  - **Caught by:** a spot-check of the cached heightmap with Pillow, run to verify the sizes before writing the report's evidence.
  - **Fix:** both documents now state width and height explicitly (`7c166aa`).

## 2026-10-29, Phase 2 (data core)

- **~19:37 CI.** The first push by the lead did not start CI. `gh` shows the GitHub account is locked for billing, so no job ran. A commit cannot fix that; it is reported in the Phase 2 report. Runners were pinned to `ubuntu-24.04` anyway (`5ac3cd4`), so the first real run is reproducible.

- **19:43 to 19:51 Commits.** The module path became `github.com/serlorsx-xln/mahanathi/server` (`840dd69`). The front end is embedded with a tracked `.gitkeep`, so a fresh clone builds and serves a notice page (`ceda0be`). The ThaiWater client landed with fixture tests and no live network (`8128bc7`).

- **19:52 Data.** Recorded two history responses, for URTU09 (3441) and BSP002 (12479), for the normalize oracle (`f84463c`). These were two requests, one at a time.

- **~19:55 AI mistake caught by a test.** `normalize.ResolveRefs` compared references without rounding them to the millimetre first. A 4-decimal ground level would have been kept at full precision on one path and rounded on the other.
  - **Caught by:** `TestResolveRefs/"values rounded to mm"`, written before the code was finished.
  - **Fix:** `ResolveRefs` rounds its inputs itself (`258ca3d`).

- **~20:10 AI mistake caught by review.** `domain.Level` (`39be5ae`) compared float64 percents with the thresholds. With ground −1.356, bank 5.044 and wl 3.124, an exact 70 % computes as 70.000000000000014 % and became level 4, though rule_web says "<=70" is level 3. 104.375 % computed as 104.37499999999999.
  - **Caught by:** working out the expected percent for a humanize test by hand (8.35 / 8 = 104.375, so 104.38), then a scratch search that found the 70 % cases.
  - **Fix:** levels, percents and distances are computed in whole millimetres (`96084d6`). The oracle still agrees on all 37 fixture rows.

- **~20:20 AI mistake caught while writing the store tests.** `go get modernc.org/sqlite@latest` resolved to v1.61.0, whose module requires Go 1.26. The local toolchain is Go 1.25.6, so the build broke the moment the dependency landed.
  - **Caught by:** `go build ./...` immediately after the `go get`.
  - **Fix:** pinned v1.59.0, the newest version that builds on Go 1.25, as part of `9cd5506`. Recorded late (21:07); the entry was planned before the outage and lost with it.

- **~20:30 AI mistake caught while preparing the ingest tests.** `normalize.ResolveRefs` compared the load and graph references as float64 metres with `math.Abs(a-b) > 0.005`. For BKC004 the load's 0.91 and the graph's 0.905 differ by 0.0050000000000000044 in float64, so a value that is exactly the expected rounding was flagged as a disagreement.
  - **Caught by:** a table case for the ingest oracle that expected zero ref-mismatch warnings on the recorded fixtures.
  - **Fix:** the tolerance is now whole millimetres (`refsDiffer`, 5 mm), consistent with the mm arithmetic everywhere else (`96084d6`). Committed after the outage as `23a869f`; the edit itself was written before 20:31 and survived on disk.

- **20:32 Outage.** The machine lost power while Phase 2 was running; the implementer ran headless under the lead's orchestration session, so the run died mid-ingest. Uncommitted work survived on disk (the normalize fix above and the new `ingest_test.go`). No implementer process was left running. The lead verified the repo (`git fsck` clean apart from exFAT `._*` noise, build ok) and resumed this session.

- **Correction of a commit message (not history).** `5b86a32` says "51 table cases". The tables in `humanize_test.go` hold 46 `Describe` cases and 20 `National` cases. Per the lead's answer 11 the commit stays as written; this entry is the correction of record.

- **21:07 Model switch, recorded per the lead's continuation note.** From this resume the implementing session runs on **GLM-5.3** through the owner's router. The owner chose this after the Anthropic subscription was blocked for past-due payment. Everything through `fc26478` (Phases 1 and the first half of Phase 2) was written by Claude Opus 5.5 at effort xhigh; from here the DEVLOG is the source of truth for which model wrote what. Same session id, context intact. The briefs, test discipline and gates are unchanged.

- **~21:15 AI mistake caught by a test.** The ingest test expected BSP002's newest reading at the graph's 19:50, but that point's value is null; the newest valid point is 19:40. The expectation was written from the fixture's last row without checking its value.
  - **Caught by:** `TestFirstRunStoresTheSnapshotAndBackfillsEveryStation` on the first run of the new `ingest.go`.
  - **Fix:** the test now expects 19:40 and also checks the value (16.727).

- **~21:15 AI mistake caught by self-review before running.** `ingest.record` stored `formatRuleWeb(nil)` — an empty string — as the `rule_web` meta value, so every run would have looked like a change. The rules were not in scope in that function.
  - **Fix:** `SetMeta` moved into `RunOnce`, where the parsed rules are in scope; `record` only records the run.

- **~21:40 AI mistake caught by running the tests.** The timeline test expected the first bucket at 18:00 UTC; 13:00 minus 71 hours is 14:00. The expectation was written by mental arithmetic without checking.
  - **Fix:** the test now derives the expectation the way the code does and pins 14:00.

- **~21:44 Flaky tests under load, not code.** `TestPerAttemptTimeoutIsRetried` and `TestBodyLimit` in the thaiwater package failed once during a full-module run (the machine was running every package's tests at once) and passed in isolation and in the next full run, right before `e448700` (21:45). Both are timing tests around per-attempt timeouts. Recorded here in case CI sees it; no code was changed. (Entry was first logged at the end of the phase as "~22:10", out of order; moved and re-timed in Step 0 of Phase 3 per the lead's finding 5.)

- **~21:55 AI mistake found by debugging a failing round trip.** The test-side timeline decoder wrote `prev + v>>1 ^ -(v&1)` without parentheses, which Go parses as `(prev + v>>1) ^ -(v&1)`: every odd delta decoded as its bitwise complement. The encoder was correct; only the test decoder was wrong. An earlier fix for exactly this did not apply, because gofmt's spacing did not match the string it searched for.
  - **Caught by:** a hexdump of the encoded stream traced against the decoder, after hand-made round-trip data passed but real data failed.
  - **Fix:** parentheses, and a check that the fix applied.

- **~21:55 Design note (not a mistake, a discovery).** 800 × 72 hourly percents as JSON numbers are 96 KB compressed on the smoothest synthetic water and 120 KB on wild water, over the 80 KB budget however the JSON is shaped. The timeline is therefore a binary body: presence bitmap plus zigzag-varint deltas, precision kept at 0.01 % to agree with /now to the digit. Measured 61 KB compressed on rivers wilder than any real one. Documented in `docs/api.md` with a decoder.

- **~21:55 AI mistake caught by reading the running container's logs.** The Dockerfile declared `ARG VERSION` before the first FROM only, which scopes it to FROM lines; the build stage saw an empty version, so the User-Agent said `mahanathi/`. **Fix:** re-declare `ARG VERSION` inside the stage (`2df17fb`).

- **~21:58 AI mistake caught by looking at the served page.** The deploy image served a real web build, not the placeholder the lead asked for: the single server stage copied the web dist unconditionally, and a build from 18:55 was lying around. **Fix:** the Dockerfile now has a `deploy` target (server only, placeholder page) and a `full` target with the web build; `make docker` builds deploy, `make docker-full` builds full (`f6038ef`).

- **21:57 Live verification.** The deploy image (21.4 MB, distroless, healthcheck healthy, 0 warnings in the log) ran its first ingest in 33 s: 804 rows, 804 observations, 251,624 readings upserted, 804/804 histories fetched, oracle 789/789 levels in agreement. `/api/v1/now` (765 KB raw, 85 KB gzipped) served the national sentence "ตอนนี้ 18 สถานีน้ำล้นตลิ่ง ส่วนใหญ่อยู่ที่นครปฐม กาญจนบุรี และสุพรรณบุรี"; the timeline body is 74 KB raw, 46 KB gzipped. The two suspect stations (MKSND01, BKK005) were hand-checked against the raw graph responses; both are datum alternation in the logger (details in data-quality.md). The live container and its volume were removed after the checks; the image is on the local daemon.

## 2026-10-29, Phase 3 (look-dev)

- **22:36 Step 0.** The five findings from the lead's Phase 2 review were fixed, each in its own commit: the `suspect` wire shape (`dd30cf9`), the timeline decoder doc (`23ede73`), the 429 security headers (`4f370c9`), the Phase 2 report's rate-limit record (`c10b3c3`), and the DEVLOG ordering above (`000b801`).

- **22:36 Process mistake.** While reordering the DEVLOG entry, a commit command was mistyped and an empty-message commit `7cdc970` ("amend") landed with the reorder diff in it, and the intended message never did. The file content is correct; only the commit message is wrong. Recorded here rather than rewritten (the lead's answer 11: no history rewrites).

- **22:55 Geodata shipped (`75389cf`).** All four assets are committed with build scripts and `geodata/SOURCES.md` (licences and credit lines); 1.46 MB total, inside the 1.5 MB budget. Both scripts are deterministic: every rebuild from the cache is byte-identical to the shipped file — the heightmap against the research run's `heightmap_candidates_band75_b2/` candidate, the three vectors against their candidates. A false alarm during this: the first heightmap rebuild was compared against `heightmap_candidates/h2048_prio_sqrt11_gray16.png`, which is the *25 km-band* run (1,174,213 B), not the shipped 75 km configuration; the 211,564 differing pixels were the different band, not nondeterminism. `geodata/SOURCES.md` initially recorded that wrong size; corrected before committing.

- **23:05 Design tokens shipped (`43a4ace`).** `web/src/design/tokens.ts` is the single source; `tokens.css` is generated from it (`pnpm tokens:css`) and a test fails if they drift. Writing the AA test caught two real things before any pixel was drawn: the stale swatch is a 3:1 graphic stroke, not a 4.5:1 text colour (text about stale data uses `textFaint`), and the first level-3 green was slightly too saturated for "only 4 and 5 get strong colour" (`#55a06c` → `#5c9e72`).

- **23:20 The situation scale was rebuilt by search (`scripts/research/cvd_scale_search.py`).** While building the colour-evidence page, the first CVD check turned out to use mistyped Machado matrices (rows mixed between sources); with the canonical matrices that first scale failed badly: reds and greens both collapse toward olive under deuteranopia, and the AA contrast floor pins any red at simulated L\* ≈ 62, exactly where the quiet greens sit. Hand-tuning could not solve that, so the scale is now the output of a constrained search over the five levels (AA on every surface, quiet saturation below level 4, the amber pinned by eye) maximising the minimum pairwise delta-E under both simulations. Result: deutan 14.1, protan 15.0, normal vision 29.9 — every pair clearly distinct. Level 1 moved to a muted slate blue because the blue axis is the only hue direction that survives both simulations; level 2 is near-grey. The tokens and their tests were updated to the searched values.

- **23:25 Environment fact that shapes the gate.** This implementing session cannot view images: the Read tool blocks PNGs ("this session's model cannot accept images, the vision fallback is unconfigured"), and a subagent run on a different model is blocked the same way. The Phase 3 gate therefore cannot rest on this session's eyes. It will rest on computed evidence instead — pixel sampling of the screenshots against the tokens, canvas ink metrics for the Thai clipping questions, DOM measurements, and the deterministic tests — and the phase report will name the lead's visual pass as the remaining step before Phase 4, which the workflow already requires.


## 2026-10-30, Phase 3 (look-dev, continued)

- **~00:09 AI mistake caught by the relief e2e's console assertion.** The scene's two custom shaders never compiled: the terrain fragment referenced a `vCameraPos` that was never declared (the uniform is `uCameraPos`), and the gauges fragment interpolated `${CAP}` into some lines but not the `waterTop` line, leaving a bare `CAP` identifier in GLSL. WebGL logged `VALIDATE_STATUS false` and drew nothing for those meshes: with the program invalid every draw call failed, so the screenshots taken before the fix show only the sea-coloured clear, the panel and the chrome — no terrain, no gauges. **Caught by:** `expect(errors).toEqual([])` on the page console, after a diagnostic run showed `drawElements: no valid shader program in use` for the instanced gauges. **Fix:** the varying/uniform name and the missing template interpolation. The lesson recorded: "the build passes and the scene mounts" says nothing about whether the GPU accepted the shaders; only the console does.
  - The same diagnostic run explained the relief-spec timeouts: Playwright's WebGL screenshots stall on `ReadPixels` (driver warning: "GPU stall due to ReadPixels"), several seconds each, so three tests legitimately need more than the 30 s default — `test.setTimeout(180_000)` on the describe.

- **~00:20 AI mistake caught by sampling the gate screenshots against the tokens.** The scene's colours were all wrong in a way no test reported: a raw `ShaderMaterial` gets none of three's automatic output transforms, so the terrain, sea, gauges and fog were writing linear working-space values straight to an sRGB framebuffer — the sea `#0d1a23` reached the screen as `#010304`. Compounding it, R3F's Canvas defaults to ACES tone mapping, which would have silently rewritten every token colour even with correct encoding. **Caught by:** pixel-sampling the screenshots and finding a 35% colour that is no token, then recognising it as `#0e161d`'s linear value displayed raw. **Fix:** `flat` + `alpha: false` on the Canvas, `#include <tonemapping_fragment>` / `<colorspace_fragment>` at both custom fragments' single exit, and a moonlit Lambert scale in the shader (0.22 + 0.55 key + 0.16 fill) since with NoToneMapping that scale is the only thing between the tokens and the screen. The sea's early `return` initially bypassed the new chunks — restructured to one exit. After the fix the screenshots sample back the tokens exactly: sea `#0d1a23` 34%, ground `#0e161d` 21% of the national view.

- **~00:30 Two evidence-method corrections, both found by the measurements.** (1) The exaggeration screenshots were being taken before the lazy scene chunk had painted its first frame (the panel renders before the canvas does), so `lookdev-exagg6` was 82% empty background; the page now exposes `data-scene-ready` from the Canvas `onCreated` and the specs wait for it plus two rAFs. (2) A mid-flight screenshot is not capturable with `page.screenshot` — WebGL captures stall seconds on ReadPixels, longer than the 1.6 s flight, and the first attempt unknowingly clipped only the fixed DOM panel (mean diff 0.01). The mid-flight frame now comes from CDP `Page.startScreencast`, and the test asserts the streamed frames differ, so motion itself is tested.

- **~01:00 The gate's first critique iteration (dual-agent, impeccable).** Assessment A (design review) and Assessment B (detector + browser) ran as isolated sub-agents. B caught what no build or unit test had: `web/src/design/base.css` was imported nowhere, so `/specimen` and `/color` served textSecondary and textFaint on browser-default white (2.1:1 and 3.1:1, measured on the served build) — the pages that exist to prove AA contrast were the only pages failing it, and the browser-surface theming (selection, focus ring) was unapplied everywhere. A's findings: the unhandled fetch-failure path (infinite silent skeleton, contra brief §9), "ใกล้ฉัน" flying to a fixture station while labelled as near-me, the camera target not bounded to Thailand, and the hairline-border-plus-wide-shadow / blur-on-five-elements judgement calls. **Fixes:** base.css imported plus caret, scrollbar and underline-offset theming; fetch failure caught with named-problem copy and a working retry (e2e-tested); the button now reads "ใกล้ฉัน · ตัวอย่างสถานี"; the rig clamps its target to Thailand's extent; the wide panel shadow dropped (veil + border + the one allowed blur separate the panel) and the dev chrome sits on solid panel instead of veiled blur; the skeleton carries a screen-reader status; the canvas div is labelled; `situations[4]` became a named `OVERFLOW`. Detector re-scan after the fixes: zero findings on all three routes (the one residual, a 95-character line on the specimen note, fixed with a 60ch measure and re-verified clean).

- **~01:15 Gate iteration 2 (fresh assessment, 31/40, zero banned items).** The re-review verified every iteration-1 fix in source and tests, walked the full anti-slop checklist (all PASS with evidence), and found one regression introduced by iteration 1 itself: adding `role="status"` to the skeleton had made the no-WebGL floor announce "กำลังโหลด" forever after the data loaded, because the skeleton branch gated on `!failed` rather than `!data`. Fixed, and the no-WebGL e2e now asserts the skeleton is gone once the panel is visible — the assertion that would have caught the regression. Also taken: the retry returns to the loading status while re-fetching, the panel's numbers row gains a sixth hollow "ยังไม่อัปเดต" entry keyed to the hollow 3D gauges, missing-value markers unified to '·', "% ของตลิ่ง" unified across pages, `aria-pressed` on the presets, and the shadow token annotated as reserved for Phase 4. Deferred with reasons (recorded in the gate doc): gauge zoom labels (Phase 4 LOD), the phone bottom sheet, the 72-hour chart, keyboard station access, adaptive DPR.

- **~01:25 Phase 3 closed.** `process/reports/phase-3.md` written with the evidence index; PROGRESS updated. The implementing session stops here for the lead's visual pass on real hardware (the one thing the computed evidence cannot substitute for) and the Phase 4 prompt.

- **~02:05 Plan change: Phase 3b, the relief rebuild.** The owner rejected the Phase 3 relief visually (ยังไม่สวยเลย) and chose Option A from the lead's research: MapLibre GL JS as the terrain/basemap engine, our gauge data layer kept. ADR `0007-relief-rebuild-maplibre.md` records the decision, the alternative (repairing the hand-built scene against 15 quality bugs and the clock), and the consequences. Also recorded there and here, for the project's honesty standard: at ~01:24 the lead reported launching a remediation round that was in fact never launched (no prompt written, nothing entered the session) — a lead process error, acknowledged in the Phase 3b prompt and corrected by this phase actually running.

- **~02:35 Step 0 finding that flips the premise: the committed JSON was the liar, not the ADR.** The lead's review held that ADR 0005's figure table contradicted `specimen-measurements.json` (which showed Noto Sans Thai's minus equalling its plus, 2.53 = 2.53). Re-measuring with the candidate fonts restored proves the opposite: the fresh, load-verified numbers give Noto Sans Thai minus 2.306 vs plus 2.556 — the ADR's original claim — and every other value in the ADR's table matches the fresh run exactly. The committed JSON had been silently corrupted: the e2e re-run that produced it ran **after** the candidates were trimmed from the build, so its Anuphan and Noto Sans Thai rows measured the system fallback (the tell: two different families with bit-identical floats, digit 55.615234375 for both), while `document.fonts.check` still reported them "loaded" because an unknown family resolves to the already-loaded default. **Fixes:** the committed JSON is now the verified full-candidate measurement; `e2e/tests/measure.spec.ts` verifies every family with `document.fonts.load` (empty result = not in build), refuses to write a partial measurement (the committed record stands; re-measuring is a deliberate promote→run→trim act), and throws if all UI families measure the same digit advance (a fallback collapse cannot pass silently); ADR 0005 carries an evidence-integrity note; the phase-3 report and GATE-phase-3.md carry dated correction notes (no silent history edits); the fonts are trimmed back to the pairing. The decision itself needed no flip — it stands on now-verified evidence.

- **~02:55 Phase 3b step 1: the MapLibre base frame renders (terrain-only; basemap still building).** The engine swap is in: MapLibre GL 6.16.0 (`setProjection({type:'globe'})` confirmed in the shipped .d.ts), our terrarium DEM as `raster-dem` + hillshade, borders and rivers as GeoJSON, the sky block carrying the one continuous night fog, all colours from the tokens. The hand-built `src/scene/` is deleted (history keeps it, per ADR 0007). Two engine-integration bugs were caught by the first headless screenshot, both worth their DEVLOG lines:
  - **MapLibre 6's worker never loaded in the Vite build** ("Worker failed to load", the map's `load` never fired, `data-scene-ready` never appeared). MapLibre 6 resolves `./maplibre-gl-worker.mjs` against `import.meta.url`, which points into our lazy chunk after bundling; the file is not there. **Fix:** the worker is emitted as its own asset (`maplibre-gl/dist/maplibre-gl-worker.mjs?url`) and `setWorkerUrl()` points at it — the worker bundle is self-contained, so that is the whole fix. Without the worker fix the page looks fine in the DOM and never draws a pixel; the only witnesses are the console and the missing `load` event.
  - **The basemap HEAD check was a false positive:** dev/preview servers answer missing paths with the SPA fallback (index.html, 200, `text/html`), so the check believed the archive existed and the PMTiles reader choked on HTML ("Wrong magic number"). **Fix:** the check now requires a non-HTML content type.
  - The terrain-only frame verifies by pixel sampling (this session cannot view images, per the 23:25 entry above): 79.8% ground token at the national view, 22 distinct quantized shades from the hillshade (lit ridges, not a flat fill), zero console errors. Evidence: `docs/design/screens/phase-3b-base-frame-terrain-only-1440.png`. The sea reads as ground for now — the water surface comes from the basemap, which is still building in planetiler (osm_pass2). Bundle, measured gzip: index 75.1 kB + lazy LookdevMap 303.2 kB + worker 143.9 kB — over ADR 0004's ~250 kB budget; recorded for the re-gate, not hidden.

- **~11:55 Phase 3b step 2: the gauges render on MapLibre, in both encodings.** The custom layer (`web/src/map/gauges.ts`) draws all 804 stations as one instanced call - crossed slim blades, bank-full cap + water fill + overflow above the cap, ground-contact base ring, screen-constant minimum width, fog from the relief token; the disc fallback (`web/src/map/discs.ts`) sits behind `?gauges=discs`. The brief's bug list is addressed where it survived the port: no per-frame buffer uploads (5 - instance buffers rebuild only on stations/exaggeration change; camera-dependent values are uniforms), live reduced-motion listener (7), overflow clamped to the blade top by construction (9 - the last 70% of the clamped fill maps exactly onto CAP..1), suspect readings render hollow-capped or translucent, never fully trusted (8), and the base ring is darker than the track (10). Three engine lessons cost real debugging time and are recorded so they are not relearned:
  - **The projection prelude owns `#version`.** The custom-layer contract puts `shaderData.vertexShaderPrelude` *after* your own `#version 300 es` line, not before - my first assembly had it first, so the version directive landed on line 11 and the compiler rejected every `in`/`out`. **The prelude's uniforms are ours to set.** MapLibre declares `u_projection_matrix` (and globe's four friends) but does not upload them into a custom program; with zeroed matrices the draw executes silently and nothing appears anywhere. Both were caught by the pixel sampler, not by any console error - a custom layer that compiles, links and draws can still be invisible.
  - **Mercator projection takes mercator z, globe takes metres.** At national zoom (globe) the blades rendered; at station zoom (mercator, after the globe→mercator transition) they vanished: `projectTileFor3D`'s elevation parameter is metres under globe but conformal mercator units under mercator, so a 44 km pole interpreted as 44,000 world-units clipped out of the frame. One uniform (`uElevScale` = 1 or `meterInMercatorCoordinateUnits()`) switches the unit.
  - **Canvas readback needs `preserveDrawingBuffer`.** The evidence pipeline samples the canvas pixel-by-pixel; with the default `preserveDrawingBuffer: false` the readback races the compositor and returns one flat shade while the composited page (and `page.screenshot`) shows full detail. The map now sets `canvasContextAttributes: { preserveDrawingBuffer: true }`.
  - Evidence: `docs/design/screens/phase-3b/gauges-{poles,discs}-{national,station}-1440.png`; pixel sampling confirms every situation colour on the canvas in both modes (e.g. level-3 green 513 px, level-4 amber 633 px, level-5 red 72 px at the national poles view; the สองพี่น้อง overflow blade dominates the station view at ~79k red pixels). The full e2e suite (10 tests: presets, flight, exaggeration, reduced motion, no-WebGL, failed fetch, specimen, colour, smoke, measurements) passes against the MapLibre build - the DOM contracts carried over unchanged. The basemap PMTiles (160 MB, planetiler from the Geofabrik Thailand extract) is built and served with range requests; it is a gitignored local like the DEM tiles, regenerated by the recorded commands.

- **~12:20 Phase 3b steps 3–4: interaction port complete, re-gate passed with fixes.** The LINE in-app browser offer (J6) landed with an authored-SVG close and `?openExternalBrowser=1`; the e2e that verified it caught a real defect first — on phones the hint sat on top of the answer panel and the panel intercepted the close button's pointer — fixed by moving the hint below the panel. The re-gate ran as the prescribed dual-agent critique (artifact `2026-10-30T12-16-53+07-00__src-routes-lookdev-tsx.md`, 28/40, zero banned items, 5 P1s) and every P1 was fixed and re-measured before commit:
  - **Landing frame red slab:** 24.45% of the canvas was exact alert red — the blade had a minimum screen width but no maximum. Clamped ≤36 px → 6.72%.
  - **Phone stacking collapse:** panel + sheet + footer stacked to 100% coverage with 12/19 px overlaps after selecting a station. A chosen station now owns the phone (panel and dev chrome hide; the sheet gains a close that returns to national). 
  - **Protomaps grey takeover:** 77.7% of the landing frame was the dark theme's own #1f1f1f/#3d3d3d — the retint now maps earth/landcover/buildings/roads to token-derived colours → 0.03%.
  - **First-frame poles on phone:** below the perceptual floor at country zoom (a 17 km pole is ~1.6 px tall there — wide-but-flat smears). Min width 3.5 px and a later fog start improved the desktop national field by ~50–70%; the phone-country floor remains the open poles-vs-discs question the brief routes to the lead's midday pick.
  - **A11y batch:** the canvas focus outline was suppressed (keyboard entry began invisible); the ใกล้ฉัน note failed AA on hover (4.44:1); three tap targets sat under their floors. All fixed and re-measured.
  - The critique's Assessment A also reported the LINE-hint overlap from an early pass against the pre-fix CSS; it did not reproduce on the current build (re-measured by B and the parent). Recorded in the artifact: a finding that dies on re-measurement is still evidence of how easily phone stacking breaks.
  - One test-infrastructure fix: the relief's ReadPixels screenshot stall reached variants.spec (30 s timeout) once /lookdev gained terrain+basemap+gauges; it gained the extended timeout lookdev.spec has carried since Phase 3.
