# Phase 2 report: Data core

- **Implementer:** the same Claude Code session. Claude Opus 5.5 (effort xhigh) through `fc26478`; GLM-5.3 from `23a869f` onward (the owner's decision; see the DEVLOG and `process/README.md`).
- **Phase window:** 2026-10-29 19:28 ICT (the lead's review `f4b131f`, which opened Phase 2) to 21:59 ICT (this report's commit is the last Phase 2 commit).
  - The run was interrupted at ~20:32 by a power loss, mid-ingest; uncommitted work survived on disk and the session resumed at ~21:05 with a continuation note from the lead.
- **Tree at report time:** clean at HEAD.

## Summary

The pipeline runs end to end and every number in it has a witness. The live deploy image ingested all 804 stations, backfilled 72 h of history for every one of them in 33 seconds, and reproduced ThaiWater's own level on 789/789 stations that publish one. `/api/v1/now` speaks the national sentence; the timeline fits half its budget; the humanize package — the crown jewel — is at 100 % coverage with a golden file and a fuzz target that has run 2.6 million times.

**Step 0 (Phase 1 close-out)** landed as `fa8a9fb`, `97629df`, `7c166aa` before the outage, and the Phase 1 report was completed then. Nothing was outstanding from it at resume time.

## What was built

| Commit | Piece |
|---|---|
| `840dd69` … `8128bc7` | module path, `go:embed` on a fresh clone, the ThaiWater client with fixture tests |
| `258ca3d` | `normalize`: DTOs to stations and readings under ADR 0003 |
| `39be5ae`, `96084d6` | `domain`: the ADR 0003 arithmetic in exact millimetres |
| `28977e8` | `thaifmt`: Thai number, time and date formatting |
| `5b86a32` | `humanize`: the sentence, the facts, the national line |
| `9cd5506`, `0707190` | `store`: SQLite WAL, migrations, idempotent upserts, bounded refetch, run records |
| `23a869f` | `normalize` fix: reference mismatch tolerance in whole millimetres |
| `110e45d` | `ingest`: the snapshot, the backfill, the oracle, the schedule |
| `1e4eb9a` | `api`: `/now`, `/history`, `/timeline`, `/readyz`, the rate limit, `docs/api.md` |
| `e448700` | `main`: the wiring, env config incl. `PORT`/`DATA_DIR`, the `healthcheck` subcommand, the placeholder page |
| `b7a4282`, `2df17fb`, `f6038ef` | the deploy image: `/data` for 65532, self-healthcheck, version stamp, the deploy/full split |
| `7ad380e` | `data-quality.md` finished: the night snapshot, the implemented behaviour, the suspect hand-check |

## Verification

### `go test -race -cover ./...` (final run, 21:58)

```
ok  	github.com/serlorsx-xln/mahanathi/server/cmd/mahanathi          3.4s	coverage: 20.2%
ok  	github.com/serlorsx-xln/mahanathi/server/internal/api           2.7s	coverage: 87.8%
ok  	github.com/serlorsx-xln/mahanathi/server/internal/config        1.9s	coverage: 100.0%
ok  	github.com/serlorsx-xln/mahanathi/server/internal/domain        2.9s	coverage: 100.0%
ok  	github.com/serlorsx-xln/mahanathi/server/internal/httpx         4.2s	coverage: 100.0%
ok  	github.com/serlorsx-xln/mahanathi/server/internal/humanize      3.6s	coverage: 100.0%
ok  	github.com/serlorsx-xln/mahanathi/server/internal/humanize/thaifmt	6.0s	coverage: 100.0%
ok  	github.com/serlorsx-xln/mahanathi/server/internal/ingest        7.2s	coverage: 87.6%
ok  	github.com/serlorsx-xln/mahanathi/server/internal/normalize     3.8s	coverage: 99.4%
ok  	github.com/serlorsx-xln/mahanathi/server/internal/store         5.0s	coverage: 88.9%
ok  	github.com/serlorsx-xln/mahanathi/server/internal/thaiwater     4.6s	coverage: 90.3%
ok  	github.com/serlorsx-xln/mahanathi/server/web                    4.9s	coverage: 78.3%
```

The humanize spec's coverage bar (90 %+) is met by the package itself at 100 %. `cmd/mahanathi` is 20 % because most of `main` is process wiring (logger, signal handling, exit codes) exercised by hand in the live run instead. One flaky pair of timing tests in `thaiwater` appeared once under full-module load and passed in isolation and on every rerun; noted in the DEVLOG.

### One live run (the deploy image, 21:52–21:58)

`make docker` → `mahanathi:dev`, 21.4 MB, distroless, `/data` owned by 65532, HEALTHCHECK via the binary's own `healthcheck` subcommand. Run on 127.0.0.1:8099 with a fresh volume:

- **First ingest, 33 s** (log line, verbatim): `rows: 804, observations: 804, dropped: 0, upserted: 251624, history_ok: 804, history_failed: 0, warnings: 0, oracle_checked: 789, oracle_mismatches: 0`.
- **`/readyz`** → 200 `ready` once the data landed; `healthcheck` subcommand exits 0; Docker reports `healthy`; zero WARN/ERROR lines in the whole run.
- **`curl /api/v1/now`**: 804 stations; summary `levels [63, 159, 328, 138, 18]`, `staleCount 84`, `suspectCount 2`; the national sentence:

  > **ตอนนี้ 18 สถานีน้ำล้นตลิ่ง ส่วนใหญ่อยู่ที่นครปฐม กาญจนบุรี และสุพรรณบุรี**

- **Payload sizes:** `/now` 765 KB raw / 85 KB gzipped; `/timeline` 74 KB raw / **46 KB gzipped** against the 80 KB budget (800 × 72 synthetic worst case measured at 61 KB in tests).
- **Placeholder page** (`/`, 1,952 bytes): the national sentence and the five most severe stations, fetched from `/api/v1/now`; plain semantic HTML, no styling. Verified served, and the API it reads verified by curl.
- **Rate limit:** a burst of 38 requests from one address answered 200 through the 30th and 429 with `Retry-After` from the 31st.
- The container and volume were removed after the checks; the image remains on the local daemon for the lead's deploy.

### Five real sentences, hand-checked against the raw numbers

From the live `/now` at 21:57 ICT; each percent and distance recomputed by hand from the station's own references:

| Station | Case | Raw numbers | Sentence (as served) | Check |
|---|---|---|---|---|
| **BPK003 บางน้ำเปรี้ยว** (ฉะเชิงเทรา) | overflow | wl 1.863, bank 1.470, ground −10.948, 14:40 | สูงกว่าตลิ่ง 0.39 เมตร (น้ำล้นตลิ่ง) ลดลง 5 ซม. ใน 6 ชั่วโมงที่ผ่านมา | 1863−1470 = 393 mm ✓; (1863+10948)/(1470+10948) = 103.16 % → level 5 ✓; Δ6h = −52 mm ✓ |
| **BKK009 คลองลำปลาทิว ลาดกระบัง** (กรุงเทพฯ) | high | wl 0.107, bank 0.623, ground −3.171, 14:40 | ต่ำกว่าตลิ่ง 0.52 เมตร ทรงตัว | 107−623 = −516 mm ✓; (107+3171)/(623+3171) = 86.40 % → level 4 ✓; Δ6h = −3 mm → ทรงตัว ✓ |
| **CHN004 แก่งหางแมว** (จันทบุรี) | normal | wl 22.16, bank 27.100, ground 21.600, 14:40 | ต่ำกว่าตลิ่ง 4.94 เมตร สูงขึ้น 22 ซม. ใน 6 ชั่วโมงที่ผ่านมา | 22160−27100 = −4940 mm ✓; 560/5500 = 10.18 % → level 2 ✓; Δ6h = +220 mm ✓ |
| **YOM012 เมืองสุโขทัย** (สุโขทัย) | stale | wl 47.75, bank 51.974, ground 42.185, 18:40 (3 ชม. ก่อน) | ข้อมูลล่าสุดเมื่อ 18:40 น. (ยังไม่อัปเดต) ตอนนั้นต่ำกว่าตลิ่ง 4.22 เมตร | 47750−51974 = −4224 mm ✓; leads with the age, no trend ✓ |
| **NPNPU01 เขื่อนน้ำพุง** (สกลนคร) | missing bank (zero-coded) | wl 279.674, bank/ground 0/0, 14:00 | สถานีนี้ยังไม่มีข้อมูลระดับตลิ่ง ระดับน้ำ 279.67 ม.รทก. ทรงตัว | no distance, no percent, no level ✓; the value still shown ✓ |

The two suspect stations (MKSND01, BKK005) were also hand-checked, against the raw `waterlevel_graph` responses: both are the logger alternating between two datums ~3–10 m apart, which is exactly what the jump rule exists to catch. The full write-up with the raw series is in `docs/research/data-quality.md`.

## Mistakes the tests or the process caught

All recorded in the DEVLOG as they happened. The ones from this half of the phase:

1. The BSP002 ingest expectation assumed the graph's last point had a value; its 19:50 point is null. Caught by the first run of `ingest_test.go`.
2. `ingest.record` stored an empty string as the `rule_web` meta value (the rules were out of scope in that function), which would have made every run look like a change. Caught by self-review before running.
3. The timeline test's expected first bucket was off by four hours (mental arithmetic). Caught by the test.
4. The test-side timeline decoder dropped parentheses around an unzigzag expression, so odd deltas decoded as their complements; an earlier fix silently did not apply because gofmt's spacing did not match the string it searched for. Caught by hexdump-tracing a round trip; the encoder was never wrong.
5. The Dockerfile's `ARG VERSION` was declared before the first FROM only, so the User-Agent said `mahanathi/`. Caught by reading the running container's logs.
6. The deploy image embedded a stale web build instead of the placeholder, because the single server stage copied the web dist unconditionally. Caught by looking at the served page. The Dockerfile now has explicit `deploy` and `full` targets.
7. Three DEVLOG entries were written with timestamps ahead of the real clock (from memory, not read at the time). Caught by `date` after the fact and corrected in `3d95bdf`.

## Open questions

1. **The timeline is binary** (`application/octet-stream`, format in `docs/api.md` with a reference decoder). JSON could not meet the 80 KB budget at 800 × 72 values however shaped (measured: 96 KB+ compressed on the smoothest synthetic data). If the lead wants JSON regardless, the budget has to move, not the format.
2. **`/now` is 85 KB gzipped** for 804 stations, mostly the two Thai sentences and the facts per station. Comfortable for a browser, but if Phase 4 finds it heavy, the facts array is the part to make optional (a query flag), not the sentence.
3. **The nightly staleness shape** (119 stations over 3 h old at 21:49, 116 still carrying an official level) is now quantified in `data-quality.md`. Nothing to decide here; noted because the about page should say it plainly, per the lead's answer 3.
4. **CI still cannot run** — the GitHub account is locked for billing (Phase 1 finding, unchanged). The suite above is the local `-race -cover` run; the pinned `ubuntu-24.04` runners are waiting for the account to be unlocked.

## Deploy readiness (for the lead)

- `make docker` → `mahanathi:dev` (21.4 MB). `PORT` is honoured (Coolify's variable); `MAHANATHI_ADDR` wins if set; `MAHANATHI_DATA_DIR=/data` with uid 65532 ownership; the image self-healthchecks through `/readyz`.
- The image was live-verified end to end as above; the container was stopped and the volume removed afterwards, so nothing of mine is left running.
- The phase stopped here as instructed; pushing and the Coolify deploy are yours.
