## Review of Phase 1 (lead)

The foundation and the data investigation are excellent. The data-quality work (the unsigned `diff_wl_bank`, the 282 RID datum shifts, the zero-coded dam references and the precision flip) is precisely the depth the owner needs to defend. The commit history and DEVLOG are what I asked for.

**One process failure.** You ended your turn while the geodata sub-agent was still running. Headless, that ended the phase. As a result:
- `docs/research/geodata-sources.md`, `geodata/scripts/research/*`, ADR 0004, `docs/PROGRESS.md` and the report are uncommitted.
- The report still contains `__END__`, `__GEODATA_COMMITS__`, `__GEODATA_EVIDENCE__` and `__GEODATA_QUESTIONS__`.

I added a rule to `CLAUDE.md` about this. Please re-read it.

### Step 0: close Phase 1 before any Phase 2 work
1. Review the geodata sub-agent's `geodata-sources.md` and scripts the way you reviewed the other sub-agent docs. Delete `__pycache__`, ignore `*.pyc` (already done) and commit them.
2. Fold the DEM and river facts into ADR 0004 and commit it.
3. Fill every placeholder in `process/reports/phase-1.md`, set the end time to the time of your last Phase 1 commit, and commit the report and `docs/PROGRESS.md`.
4. Log the early stop in the DEVLOG as a process mistake.

### Answers to your open questions
1. **`diff_wl_bank`:** you are right. I corrected the engineering brief §1 (see the correction note). Never read its sign.
2. **Distance when ground is absent:** accepted. When `min_bank` is valid and not zero-coded, state the distance (`wl − min_bank`) but no percent and no level. Update the humanize spec table in your code docs and test it.
3. **"Now" counts:** keep excluding stale rows from our counts. Add `staleCount` to the `/api/v1/now` summary. The UI will show one quiet line such as "ไม่นับ 25 สถานีที่ข้อมูลเก่ากว่า 3 ชั่วโมง", and the about page explains the difference from thaiwater.net. Honest beats matching.
4. **Graph cadence:** accept your default. Bucket to the last non-null reading in `(HH−1:00, HH:00]`. The 6-hour trend uses the bucketed series, never `waterlevel_msl_previous`. Document it in `docs/api.md`.
5. **Tabular numerals:** noted. The gate will judge figures by rendering, not by the CSS declaration.
6. **`rule_web`:** correct, the brief is fixed. Ingest should warn loudly if `rule_web` ever changes from the known thresholds.
7. **Terms:** proceed with clear attribution, serve derived numbers only, and do not serve raw responses publicly. The small recorded fixtures are acceptable test data. Add a `testdata/README.md` stating their provenance and the HII copyright. I will tell the owner about permission.
8. **Bangkok canals:** out of scope for this release. The about page states that canals are not included yet.
9. **Mekong:** fine as you describe.
10. **Repository URL:** it is `https://github.com/serlorsx-xln/mahanathi` (public; I pushed `main` at the end of my review). Rename the Go module path to `github.com/serlorsx-xln/mahanathi/server` (or whatever your layout needs) early in Phase 2 as its own commit. User-Agent: `mahanathi/<version> (+https://github.com/serlorsx-xln/mahanathi)`.
11. **Author identity:** keep it as it is. The owner's identity plus the Co-Authored-By trailer, and the README states the process. Do not rewrite history.
12. **CI:** I pushed. Check the Actions run with `gh run list --repo serlorsx-xln/mahanathi --limit 3` and `gh run view <id> --log-failed`. Reading CI is allowed, but pushing is not. If it fails, fix it in a commit and I will push again at the next review.

### Additional direction for Phase 2
- **Make `go:embed` work on a fresh clone first** (your suggested step 3), so CI stays green.
- **Deploy after this phase.** I will deploy the binary to Coolify straight after this phase so the owner has a live URL early. Make sure:
  - the Dockerfile works with a plain `docker build .` on Linux (Coolify), with an empty `/data` owned by 65532 and a HEALTHCHECK. Bring that forward from Phase 6.
  - `PORT`, `DATA_DIR` and the poll schedule come from env.
  - `/readyz` behaves correctly before the first ingest finishes.
- **Placeholder page.** For the live URL before Phase 3, a single unstyled HTML page served by the binary is fine. It shows the national sentence and a list of the five most severe stations with their human sentences, fetched from `/api/v1/now`. Keep it plain semantic HTML with **no styling decisions**, because Phase 3 owns the look. Its job is to prove the pipeline is live.
- **Re-run `dq_scan.py`** on at least two more snapshots, one of them at night, and finish `data-quality.md` as you suggested.
