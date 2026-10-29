# Data quality: bank-full, ground and the percent ThaiWater publishes

**Status: interim (Phase 1).** It is built from one full snapshot: `waterlevel_load` fetched 2026-10-29 17:54 ICT, 804 stations. Phase 2 finishes the investigation on further snapshots and records the product behaviour as implemented.

**Method:**
- `scripts/research/dq_scan.py` (stdlib Python, no dependencies) recomputes every derived field from the station references and counts agreements.
- Re-run it with `python3 -I scripts/research/dq_scan.py <waterlevel_load.json> "<fetched_at ICT>"`.
- Every number below is from its output for this snapshot, except where a separate check is named.
- A match means within ±0.06 percentage points for percentages and ±0.6 cm for levels, which absorbs the source's own rounding.

## The answer so far

1. **ThaiWater's percent of bank-full is**
   `storage_percent = (waterlevel_msl − ground_level) / (min_bank − ground_level) × 100`.
   This reproduces **all 789** published values within 1 percentage point, and 738 of them within ±0.06 pp, once `min_bank` is taken at full precision. The worst miss is 0.44 pp, worth at most 0.63 cm of water level, which is the source's own 2-decimal rounding of the level.
2. **The distance to the bank is**
   `diff_wl_bank = |waterlevel_msl − min_bank|`. It is unsigned. The direction is carried only by `diff_wl_bank_text` ("ล้นตลิ่ง" when wl > min_bank, otherwise "ต่ำกว่าตลิ่ง"), and the text agrees on 804/804 rows.
3. **`situation_level` is `scale.rule_web` applied to `storage_percent`** on 789/789 rows: >100 → 5, >70 → 4, >30 → 3, >10 → 2, ≤10 → 1.
4. **`left_bank` and `right_bank` are not a usable bank-full reference.**
   - For all 282 inconsistent stations, they disagree with `min_bank`, and every one of them is a Royal Irrigation Department (RID) station.
   - For 228 of them, the disagreement is ≥ 2 m, and typically hundreds of metres.
   - The pattern is consistent with a datum added twice (details below). This is a hypothesis we cannot confirm from the API alone.
5. **The product therefore uses `min_bank` as bank-full and `ground_level` as the lower reference.** This is the same arithmetic as the official percent, so our "above bank-full" agrees with ThaiWater's level 5 by construction. (ADR 0003.)

## Evidence

### Which reference reproduces `storage_percent`?

There are 789 rows with both a level and a published percent.

| Candidate formula | Exact match (±0.06 pp) | Within 1 pp |
|---|---|---|
| A `(wl − ground) / (min_bank − ground)` | 713 | 789 |
| **A′ as A, with `min_bank` at full precision** | **738** | **789** |
| B `(wl − ground) / (min(left, right) − ground)` | 466 | 521 |
| C `(wl − ground) / (left_bank − ground)` | 291 | 376 |
| D `(wl − ground) / (right_bank − ground)` | 252 | 328 |
| E `wl / min_bank` (no ground) | 27 | 35 |

Why A′:
- `waterlevel_load` publishes `min_bank` rounded to 2 decimals, while `left_bank` and `right_bank` keep 3.
- Where min(left, right) rounds to `min_bank`, it is the same number at full precision, so A′ uses it.
- Independently, the `waterlevel_graph` endpoint returns `min_bank` at full precision. For BKC004 it gives 0.905 against 0.91 in the load response, and with 0.905 the published 105.72 % reproduces exactly; with 0.91 it gives 105.51 %.
- The remaining 51 inexact rows all differ by less than 0.44 pp. Inverting the formula, this is a difference of at most 0.63 cm of water level, so it is explained by `waterlevel_msl` being rounded to 2 decimals. The graph endpoint shows the 3-decimal values; for example, station 737 read 3.167 at 17:40, published as "3.17".

### `diff_wl_bank`

| Reference | Signed match | Match as an absolute value |
|---|---|---|
| `wl − min_bank` | 23 / 804 | **751 / 804** |
| `wl − min(left, right)` | 17 / 804 | 494 / 804 |
| `wl − left_bank` | 13 / 804 | 309 / 804 |

- No `diff_wl_bank` is negative. The 53 absolute-value misses are all exactly 1 cm off, which is rounding.
- The brief's note that `diff_wl_bank` equals `wl − min_bank` was true for Gt.11 at 14:00 only because the water was above the bank at that moment. At 17:54 Gt.11 read wl 40.16, min_bank 40.60, `diff_wl_bank` "0.44" and "ต่ำกว่าตลิ่ง".
- Code that trusts the sign of `diff_wl_bank` would report every station as above its bank.

### `left_bank` / `right_bank` against `min_bank`

| Group | Stations | Agencies |
|---|---|---|
| `min_bank` = min(left, right) within 1 cm | 522 | HII 327, FOP 92, EGAT 70, RID 33 |
| differ by 1 cm to 2 m | 54 | RID 54 (in 49 of them `min_bank` is the lower value) |
| differ by ≥ 2 m | 228 | RID 228 (by ≥ 10 m in 185) |

Every station from HII, FOP and EGAT is consistent. RID is consistent for only 33 of its 315 stations.

Examples (m MSL):

| Station | left | right | min_bank | ground | wl at 17:54 | min(L,R) − ground |
|---|---|---|---|---|---|---|
| P.64 บ้านหลวง | 1580.785 | 1580.871 | 793.35 | 787.025 | 788.21 | 793.76 |
| Gt.11 บ้านช้างแรก | 84.709 | 84.345 | 40.60 | 37.72 | 40.16 | 46.63 |
| Y.16 บางระกำ | 70.526 | 71.553 | 38.93 | 30.84 | 38.24 | 39.69 |
| T.1 นครชัยศรี | −7.586 | −7.578 | 1.50 | −9.10 | 2.36 | 1.51 |

**Hypothesis: a datum added twice.**
- If the left and right values had a datum near the ground level added a second time, then subtracting `ground_level` should land them near `min_bank`.
- Across the 222 shifted stations that have a ground level, the median |min(left, right) − min_bank| is **126.78 m**. After subtracting `ground_level`, the median |(min(left, right) − ground) − min_bank| drops to **2.64 m**:

  | Remaining difference | Stations |
  |---|---|
  | < 1 m | 46 |
  | 1–3 m | 80 |
  | 3–10 m | 82 |
  | ≥ 10 m | 14 |

- The station's `offset` field explains the shift for only 5 stations.
- So the datum used was close to, but not exactly, the ground level. A gauge-zero datum would fit that description.
- We cannot see the gauge-zero datum in the API, so this stays a hypothesis.
- Whatever the cause, `waterlevel_msl` is on the same datum as `min_bank`, because the official percent comes out plausible for these stations: a median of 36.9 % and a p10–p90 range of 15.6–73.8 %. The left and right values are not on that datum: wl − min(left, right) has a median of −131.5 m.

Both P.64 and Gt.11 are in the fixtures.

### Missing and placeholder references

| Case | Stations | What the source does | Example |
|---|---|---|---|
| `ground_level` key absent | 9 | no `storage_percent`, no `situation_level`; `diff_wl_bank` still published | Gt.9 บ้านกลาง: wl 11.01, min_bank 10.90 → text "ล้นตลิ่ง", no level |
| `min_bank = ground_level = 0` | 6 (all EGAT dam or weir sites, e.g. เขื่อนบางลาง, ฝายละแอ) | no percent and no level, but `diff_wl_bank_text` says **"ล้นตลิ่ง"** because wl > 0 | BLGTU06 ฝายละแอ: wl 331.73 → "overflowing by 331.73 m" if taken literally |
| `ground_level ≥ min_bank` (zero height) | the same 6 | | |

These 15 rows are exactly the 15 with no `situation_level`.

### Water below the lower reference

- **18 stations** publish a negative `storage_percent`, as low as −18.93 % (URTD03, แม่น้ำชี ฝายมหาสารคาม). Their water level is below `ground_level`.
- So `ground_level` is a reference level for the percent scale. It is not literally the riverbed: water can sit below it, as at a weir or where the bed has scoured.
- ThaiWater classifies these rows as level 1, and that is correct by its rule.

### Bank-full height

`min_bank − ground_level`, over the 789 stations with both references:

| Statistic | m |
|---|---|
| minimum | 1.15 |
| p10 | 3.34 |
| median | 7.14 |
| p90 | 12.81 |
| maximum | 41.14 |

The gauge glyph encodes "fill relative to the banks", so this spread matters. A 3 m channel at 90 % and a 12 m channel at 90 % look the same, which is the intent.

### Staleness

| Fact | Count |
|---|---|
| readings older than 3 h | 27 |
| … older than a day | 6 |
| oldest reading | 2026-10-22 14:00 (SKU02), seven days old |
| stale rows that still carry a `situation_level` | 25 |

Stale level-5 example: BKC004 ปตร. คลองลัดบางยอ 2, last reading 2026-10-27 19:40, still counted as น้ำล้นตลิ่ง.

So **ThaiWater's own counts include stale stations**. At 17:54 the published counts were:

| Level | 5 | 4 | 3 | 2 | 1 | none |
|---|---|---|---|---|---|---|
| Stations | 20 | 148 | 345 | 207 | 69 | 15 |

Our "now" counts will exclude the 27 stale rows and will therefore be lower. The about page must say this in one plain sentence.

### Plausibility (the brief's suspect rules, on published numbers)

| Rule | Rows in this snapshot |
|---|---|
| `wl < ground − 20 m` | 0 |
| `storage_percent > 400` | 0 |
| \|wl − wl_previous\| > 3 m | 0 |
| coordinates outside Thailand's padded bbox | 0 |

The previous reading is 10 minutes or 1 hour back depending on the agency, so it is not a per-hour rate. Phase 2 applies the jump rule to the hourly history instead.

For scale: Gt.11 fell 0.90 m between 14:00 and 15:00 and 1.77 m in four hours (from the recorded graph), which is fast but real. The 3 m/h threshold leaves headroom over this.

### Stability: a second snapshot at 19:05

The same script was re-run on a `waterlevel_load` fetched at 19:05 ICT, 71 minutes later. Most hourly stations had moved on by one reading. The output is saved alongside the first one in the gitignored `.cache/thaiwater/`.

| Check | 17:54 | 19:05 |
|---|---|---|
| A′ exact (±0.06 pp) / within 1 pp | 738 / 789 | 746 / 789 |
| A′ misses, max implied level difference | 0.63 cm | 0.62 cm |
| `situation_level` = `rule_web(storage_percent)` | 789 / 789 | 789 / 789 |
| `\|wl − min_bank\|` = `diff_wl_bank` | 751 / 804 | 752 / 804 |
| Side banks consistent / 1 cm–2 m off / ≥ 2 m off | 522 / 54 / 228 | unchanged |
| Zero-coded / ground absent | 6 / 9 | unchanged |
| Levels flipped by 2-decimal `min_bank` | 1 (URTU09) | 1 (URTU09) |
| Older than 3 h (of which with a level) | 27 (25) | 28 (27) |
| Level 5 count | 20 | 19 |

The formula, the datum split and the zero codes held across the two snapshots. That is still only 71 minutes of evidence, so the night snapshot in Phase 2 remains on the list.

## Product behaviour per case (proposed, finalised in Phase 2)

| Case | Detected by | Percent and level | Sentence | Gauge glyph |
|---|---|---|---|---|
| Both references usable (789 stations) | `ground` present, `min_bank > ground` | ours = official formula | distance to bank from `wl − min_bank`, signed by us | fill vs cap |
| RID shifted left/right (282) | always: we never read left/right | unaffected, uses `min_bank` | unaffected | unaffected |
| Zero-coded references (6) | `min_bank == 0 && ground == 0` | none | "สถานีนี้ยังไม่มีข้อมูลระดับตลิ่ง" + the measured level, no overflow claim | hollow "no reference" mark |
| `ground_level` absent (9) | key absent | none (no lower reference) | see open question 1 | hollow "no reference" mark |
| Water below `ground_level` (18) | `wl < ground` | percent clamps to 0 for drawing; the level stays the official 1 | distance below the bank, as normal | empty fill |
| Stale (> 3 h, 27) | `now − measuredAt > 3 h` (injected clock) | kept for history, excluded from "now" counts | leads with the age: "ข้อมูลล่าสุดเมื่อ …" | hollow, quiet |
| Suspect (0 today) | brief's rules on hourly history | excluded from "now" counts | says plainly that the value looks wrong | marked, not filled |

### Open question 1

When `ground_level` is absent but `min_bank` is present (9 stations), the distance to bank-full is still well defined. ThaiWater itself publishes it in `diff_wl_bank`.

- **The humanize spec says** "missing bank-full or ground: make no distance claim".
- **I propose** making the distance claim in this one case, because ground is needed only for the percent, not for the distance.
- **Reversible default if the lead disagrees:** follow the spec and make no distance claim.

## What Phase 2 must still do

1. Re-run the scan on at least two more snapshots, including one at night, and confirm the formula and the RID pattern are stable over time. A second daytime snapshot (19:05) already agrees; see Stability.
2. Check whether `min_bank` and `ground_level` from `waterlevel_graph` differ from the load values beyond rounding, for every backfilled station. If they do, decide which one wins and record it in ADR 0003.
3. Measure how many stations a full 72 h backfill marks suspect under the jump rule, and look at each one by hand.
4. Put the final counts per case and the implemented behaviour into this file.
