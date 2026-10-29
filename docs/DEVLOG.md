# DEVLOG

This log records real events only. Times are ICT (UTC+7). Exact times come from commits or file timestamps. A `~` marks a time reconstructed to the nearest few minutes from the surrounding commits.

**Authorship:**
- The implementing agent is a Claude Code session (Claude Opus 5.5). It wrote everything from commit `36fdd33` on.
- Those commits use this checkout's configured git identity as author, and each one ends with a `Co-Authored-By: Claude Opus 5.5` trailer. No human wrote that code.
- The lead's commits are `b85c01f` and `69b1357`.
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
