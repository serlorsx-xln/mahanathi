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
