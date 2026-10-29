# ADR 0003: `min_bank` is bank-full, `ground_level` is the lower reference, and `left_bank`/`right_bank` are never read

- **Status:** accepted (Phase 1, on one snapshot; Phase 2 re-checks on more snapshots)
- **Date:** 2026-10-29
- **Deciders:** implementer, for review by the lead
- **Evidence:** `docs/research/data-quality.md`, produced by `scripts/research/dq_scan.py` on the `waterlevel_load` snapshot of 2026-10-29 17:54 ICT (804 stations)

## Context

Every claim the product makes rests on one number: how high the water is relative to the river's banks. The station record offers four candidate references: `left_bank`, `right_bank`, `min_bank` and `ground_level`. It also offers two derived values, `storage_percent` and `diff_wl_bank`. The lead's brief flagged Gt.11 as a lead, not a fact: its `left_bank` and `right_bank` sit about 44 m above `min_bank`.

What the snapshot shows:

1. **The official percent.** `storage_percent = (waterlevel_msl − ground_level) / (min_bank − ground_level) × 100` reproduces all 789 published values within 1 percentage point. 738 match within ±0.06 pp once `min_bank` is taken at full precision. The worst miss is 0.44 pp, which is at most 0.63 cm of water level and is explained by the source rounding `waterlevel_msl` to 2 decimals.
2. **The official level.** `situation_level` is `scale.data.rule_web` applied to `storage_percent` on 789 of 789 rows: >100 → 5, >70 → 4, >30 → 3, >10 → 2, ≤10 → 1.
3. **The bank distance.** `diff_wl_bank` equals `|waterlevel_msl − min_bank|`. It is **unsigned**: it is never negative, and the direction lives only in `diff_wl_bank_text`. The brief's Gt.11 check matched at 14:00 only because the water was above the bank then.
4. **The side banks.** `left_bank` and `right_bank` are inconsistent with `min_bank` at 282 stations, every one of them a Royal Irrigation Department station. At 228 of them the gap is 2 m or more, and often hundreds of metres. The pattern is consistent with a datum being added twice, but that cannot be confirmed from the API.
5. **Missing references.** 6 EGAT dam and weir sites code both `min_bank` and `ground_level` as `0`. ThaiWater still labels them "ล้นตลิ่ง" because `wl > 0`. A further 9 stations have no `ground_level` key at all. These 15 are exactly the rows without a `situation_level`.
6. **Water below the lower reference.** 18 stations sit below `ground_level`, so their published percent is negative (as low as −18.93 %).
7. **Rounding.** The load response rounds `min_bank` to 2 decimals. The graph response returns it at full precision: for BKC004 it gives 0.905 against 0.91. **The rounding changes one level in this snapshot.** With the 2-decimal value, URTU09 computes to 69.994 %, which is level 3, while ThaiWater publishes 70.04 %, which is level 4. With `min_bank` at full precision (recovered as described in `data-quality.md`, and checked against the graph for BKC004), 0 of 789 levels differ.

## Options

1. **`min_bank` as bank-full and `ground_level` as the lower reference**, which is the official arithmetic.
2. **`min(left_bank, right_bank)` as bank-full.** This is the natural reading of the field names.
3. **Take ThaiWater's derived fields as they are**: `storage_percent`, `situation_level` and `diff_wl_bank`.
4. **Option 1, with the references at full precision** when the graph endpoint supplies them.

- **Option 2 is wrong for 282 stations.** At P.64 it would put bank-full about 787 m above the water.
- **Option 3 has three problems.**
  - It cannot cover history, because the graph endpoint returns values only, not percents or levels.
  - It inherits the unsigned `diff_wl_bank`, which code could easily misread as "always above the bank".
  - It inherits the "ล้นตลิ่ง" label on zero-coded dams.
- **Option 1 at 2 decimals** disagrees with ThaiWater's level for 1 station in 789. A count that does not match the official site invites the question "which one is right?" during a demo.

## Decision

**Option 4.** In `internal/normalize`, the only place that knows these quirks, the rules are as follows.

**References:**
- `bankFull` is `station.min_bank` and `ground` is `station.ground_level`.
- When a `waterlevel_graph` response for the station is held, its `min_bank` and `ground_level` are used instead of the load response's values. Graph values are rounded to millimetres on parse, because RID values carry float32 noise such as `40.599998`.
- If the graph and load values differ by more than rounding (more than 0.005 m), the graph value still wins and normalize logs a warning. This rule is provisional until Phase 2 has measured how often it happens.

**Banks that are never read:**
- The product never reads `left_bank` or `right_bank`.
- They are not in the domain type, so they cannot be used by accident. The research script keeps using them, but only as evidence.

**The percent:**
- `pct = (wl − ground) / (bankFull − ground) × 100`, computed by us, stored unrounded, and served as is, so it may be negative.
- Clients clamp it to [0, 100] for the gauge fill only.

**The level:**
- `level` is computed from our `pct` using the thresholds in the table above.
- They are Go constants. Ingest compares them with the published `scale.data.rule_web` on every poll and logs a warning if ThaiWater changes them.
- The published `situation_level` is used only as a test oracle: on the recorded fixtures, our level must equal theirs on every row.

**The bank distance:**
- It is the signed `wl − bankFull`, computed by us.
- `diff_wl_bank` and `diff_wl_bank_text` are ignored except in tests.

**Treated as absent (no `pct`, no `level`):**
- `min_bank == 0 && ground_level == 0`.
- `ground_level ≥ min_bank`.
- A missing or unparsable `min_bank`.
- A missing or unparsable `ground_level`.

In those cases the sentence makes no overflow claim. It says "สถานีนี้ยังไม่มีข้อมูลระดับตลิ่ง" and gives the measured level, per the humanize spec.

**One case is still open:** `ground_level` absent with `min_bank` present (9 stations). The distance to the bank is well defined there, and ThaiWater publishes it. Whether we state it is open question 1 in `data-quality.md`. Until the lead answers, we follow the spec and make no distance claim.

## Consequences

**Why "above bank-full" can be trusted:**
- Our "above bank-full" is ThaiWater's level 5 by construction.
- On the 17:54 snapshot, with full-precision references, our counts by level would equal ThaiWater's, except for the deliberate exclusion of stale rows from "now" counts (`data-quality.md`, Staleness).
- That exclusion is a visible difference from the official site, and the about page explains it in one sentence.

**Test cases.** The fixtures already hold one station per case:
- P.64 and Gt.11: shifted side banks.
- BLGTU06 and MKSND01: zero-coded references.
- Gt.9 and I.17: `ground_level` absent.
- FOP045 and URTD03: water below `ground_level`.
- BSP002: needs full precision.
- URTU09: the level that flips at 2 decimals.

These become the first table cases in `normalize_test.go`.

**What the API serves.** `/api/v1/now` serves `bankFull` and `ground` as we resolved them, so a reviewer can reproduce any percent by hand from one response.

**Revisiting this decision.** If ThaiWater ever fixes the RID side banks, nothing changes for us. If it changes its own formula, the oracle tests on fresh fixtures will fail, and that failure is the signal to revisit this ADR.

**What we cannot know from the API:**
- What `ground_level` physically is. It behaves as a reference level, not the riverbed, since water sits below it at 18 stations.
- Why the RID side banks are shifted.

The about page states only what we measured. A question to HII is listed for the lead.
