# ThaiWater API: what we use, what it really returns

Researched 2026-10-29, 17:54–18:20 ICT, against the live API. All numbers below come from responses recorded in that window (the raw files are in `.cache/thaiwater/raw/`, gitignored; trimmed copies are the fixtures in `server/internal/thaiwater/testdata/`). The data-quality findings that follow from these fields are in [`data-quality.md`](data-quality.md).

## Summary

- Both endpoints in the engineering brief work as described: no key, plain JSON, gzip supported.
- The brief's field list is right, with **five corrections that change the code**:
  1. `diff_wl_bank` is an **absolute** value. Its sign lives only in `diff_wl_bank_text`.
  2. Some keys are **absent**, not null: `situation_level` (15 rows), `river_name` (84), `station.ground_level` (9).
  3. Missing bank references are sometimes coded as **zero** (`min_bank = ground_level = 0`, 6 EGAT dam/weir stations).
  4. `waterlevel_load` **rounds** `min_bank` and water levels to 2 decimals, while ThaiWater computes `storage_percent` from 3-decimal values. `waterlevel_graph` returns the 3-decimal values.
  5. History cadence is **10 minutes for HII stations and hourly for most others**, not uniformly hourly. Values from some agencies carry float32 noise (`40.159998`).
- Each row also carries `river_name` (Thai, e.g. "แม่น้ำท่าจีน") and `river_gid`. The brief did not list these. They give us the river for the sentence and for search on 720 of 804 stations.
- There is **no gauge on the Mekong mainstem** in this feed. แม่น้ำโขง can be drawn and labelled, but it has no live data.
- **Licence:**
  - thaiwater.net states "Copyright © 2024 Hydro - Informatics Institute, All rights reserved."
  - We found no published open-data licence or API terms. The API is public and unauthenticated, and is presented as a public service.
  - This is an open question for the lead (see the end).

## 1. `waterlevel_load` (current levels)

`GET https://api-v3.thaiwater.net/api/v1/thaiwater30/public/waterlevel_load`

| Observed 2026-10-29 17:54 ICT | Value |
|---|---|
| HTTP status | 200 |
| Body | 1,395,146 bytes raw, 241,911 bytes gzipped (`Accept-Encoding: gzip`), 0.32 s |
| Response headers | `content-type: application/json; charset=utf-8`, `content-encoding: gzip`, `strict-transport-security`, `x-frame-options`, `x-content-type-options` |
| Caching headers | **None**: no `ETag`, `Last-Modified` or `Cache-Control`. Conditional requests are not possible, so we poll and compare content ourselves. |
| Top-level keys | `waterlevel_data`, `waterlevel_manual_data`, `basin`, `agency`, `station`, `scale`, `province`, each shaped `{result: "OK", data: …}` |
| `waterlevel_data.data` | 804 rows, 804 distinct station ids, all `station_type: "tele_waterlevel"` |
| Agencies in the rows | HII 327, RID 315, FOP 92, EGAT 70 |
| `station` block | `{tele_waterlevel: null, canal_waterlevel: null}`, empty, so station metadata comes only from the rows |

### Fields we will use

"Missing" counts rows out of 804 where the key is absent, null or the empty string. Units: metres; levels are above mean sea level (ม.รทก.).

| JSON path | Wire type | Missing | Meaning, units | Notes for the parser |
|---|---|---|---|---|
| `id` | int | 0 | id of this reading | Changes every reading. Do not key on it. |
| `waterlevel_datetime` | string `"YYYY-MM-DD HH:MM"` | 0 | measurement time, ICT, no offset | Parse as `Asia/Bangkok` (fixed UTC+7, no DST) and store UTC. Minutes seen: :00 (385), :40 (403), :10/:20/:30/:50 (16). |
| `waterlevel_msl` | **string** | 0 | water level, m MSL, 2 decimals | Parse as decimal. Rounded from the 3-decimal value in the graph endpoint. |
| `waterlevel_msl_previous` | string | 1 | the station's **previous reading** | 10 min earlier for HII, 1 h for hourly stations (checked against the graph for station 737 and Gt.11). Not a fixed interval, so **not usable for the 6 h trend**. |
| `storage_percent` | **string** | 15 | percent of bank-full, 2 decimals | `(wl − ground_level) / (min_bank − ground_level) × 100`; see data-quality.md. Can be **negative** (18 rows). |
| `situation_level` | int 1–5 | 15 (**key absent**) | official situation level | Equals `scale.rule_web` applied to `storage_percent` on 789/789 rows. |
| `diff_wl_bank` | **string** | 0 | **\|wl − min_bank\|**, m | Never negative (0 of 804). Matches `\|wl − min_bank\|` on 751/804; the rest are within 1 cm (rounding). |
| `diff_wl_bank_text` | string | 0 | `"ล้นตลิ่ง (ม.)"` or `"ต่ำกว่าตลิ่ง (ม.)"` | Agrees with sign(wl − min_bank) on 804/804. Says ล้นตลิ่ง for the 6 zero-coded stations, so it must not be trusted without a valid bank. |
| `river_name` | string | 84 (**key absent**) | river or canal name, Thai, e.g. "แม่น้ำเจ้าพระยา", "คลองบางบาล" | Not in the brief. 720 rows have it. |
| `river_gid` | int | 84 (key absent) | river geometry id | Unknown geometry source; not used yet. |
| `station.id` | int | 0 | stable station id | Our primary key. Also the `station_id` for the graph endpoint. |
| `station.tele_station_oldcode` | string | 0 | agency code, e.g. `C.2`, `Gt.11`, `THA006` | The HII display code can differ (`THA006` shows as "TTC06" in the name). |
| `station.tele_station_name.th` / `.en` | string | 0 | station name | Some names embed a code in brackets, e.g. "วัดท่าเจดีย์ (TTC06)". The longest name (KRP901) is 85 characters, so the layout must wrap. |
| `station.tele_station_lat` / `_long` | float | 0 | WGS84 degrees | All inside lat 5–21, lon 97–106. Two FOP stations are inside Myanmar. |
| `station.min_bank` | number | 0 (6 coded as 0) | bank-full level, m MSL, 2 decimals | **The bank-full reference ThaiWater uses** (ADR 0003). |
| `station.ground_level` | number | 9 (**key absent**), 6 coded as 0 | reference bottom level, m MSL | Lower reference in the percent formula. Water can be below it. |
| `station.left_bank` / `right_bank` | number | 0 | bank levels, m MSL **in theory** | On a different datum for 228 RID stations; see data-quality.md. **Not used for computation.** |
| `station.offset` | number | 569 | unknown datum offset | Explains the left/right shift for only 5 stations. Not used. |
| `station.is_key_station` | bool | 0 | key-station flag | Possibly useful for level-of-detail later. |
| `geocode.province_name.th` / `amphoe_name.th` / `tumbon_name.th` | string | 0 / 3 / 0 | admin names (also `.en` and codes) | The empty amphoe rows include the Myanmar stations, whose province is "สาธารณรัฐแห่งสหภาพเมียนมา". |
| `basin.basin_name.th` | string | 0 | basin, e.g. "ลุ่มน้ำท่าจีน" | |
| `agency.agency_shortname.th` / `.en`, `agency_name` | string | 0 | data owner | For the per-station credit line. |

Fields we will **not** use, for now:
- `waterlevel_m`: null on all 804 rows.
- `flow_rate`: 6 rows.
- `discharge`: 286 rows.
- `sort_order`.
- `station.qmax`, `warning_level_m` (2 rows), `critical_level_m` (35), `critical_level_msl` (106).
- `station.sponsor_by` (100 rows, a sponsor name such as a company or foundation).

### `scale` (the official thresholds)

`scale.data.rule_web` is the rule list that maps `storage_percent` to `situation_level`: `>100 → 5, >70 → 4, >30 → 3, >10 → 2, <=10 → 1`. This is the same table as in the engineering brief.

`scale.data.scale` holds ThaiWater's labels and colours:

| Level | Label | Colour |
|---|---|---|
| 5 | น้ำล้นตลิ่ง | `#FF0000` |
| 4 | น้ำมาก | `#003CFA` |
| 3 | น้ำปกติ | `#00B050` |
| 2 | น้ำน้อย | `#FFC000` |
| 1 | น้ำน้อยวิกฤติ | `#990000` |

Level 1's `colorname` is "gold", which does not match its colour. We use the labels and thresholds and redesign the colours (design brief §5).

### Freshness at 17:54 ICT

| Reading age | Rows |
|---|---|
| ≤ 1 h | 686 |
| 1–3 h | 91 |
| 3–24 h | 21 |
| 1–30 days | 6 (oldest 2026-10-22 14:00) |

27 rows are older than 3 h, and **25 of them still carry a `situation_level`**, so ThaiWater's headline counts include stale stations. At 17:54, one of the twenty level-5 stations (BKC004) had last reported on 2026-10-27 19:40. Our "now" counts exclude stale rows, so **our numbers will differ from ThaiWater's, and the about page should say why**.

The newest reading in the response was 17:40, while the graph endpoint already had 17:50 for the same station, so `waterlevel_load` lags by at least one 10-minute step.

**Publication lag.** Two more snapshots, at 19:05 and 19:10, show how far behind the hour each agency's newest reading is. The two responses differ (different sha256), so this is not a server cache:

| Fetched (ICT) | HII (327), newest reading | RID (315) | EGAT (70) | FOP (92) |
|---|---|---|---|---|
| 17:54 | 315 at 17:40 | 217 at **17:00**, 80 at 16:00 | 59 at 17:00, 8 at 16:00 | 86 at 17:40 |
| 19:05 | 315 at 18:50 | 298 at **18:00**, 0 at 19:00 | 6 at 19:00, 47 at 18:00, 13 at 17:00 | 83 at 18:50 |
| 19:10 | 315 at 18:50 | 298 at **18:00**, 0 at 19:00 | 6 at 19:00, 47 at 18:00, 13 at 17:00 | 83 at 18:50 |

- **RID's HH:00 readings appear between HH:10 and HH:54.** None of the 19:00 readings was there at 19:10, and 217 of 315 17:00 readings were there by 17:54.
- **What a poll at :07 gets.** For most hourly stations it gets the *previous* hour's reading. That reading is complete (298 of 315 at 18:00) but about an hour old.
- **What a poll at :55 would get.** It would catch about two thirds of the current hour, and leave the rest one hour behind.
- **10-minute stations** trail by 10 to 20 minutes, whatever the minute.
- **No minute makes everything fresh.** The stale threshold is 3 h, so either minute keeps the hourly stations "current". Phase 2 should still measure this across a few more hours before choosing.

## 2. `waterlevel_graph` (history)

`GET https://api-v3.thaiwater.net/api/v1/thaiwater30/public/waterlevel_graph?station_type=tele_waterlevel&station_id={station.id}&start_date=YYYY-MM-DD&end_date=YYYY-MM-DD`

Recorded for three stations with `start_date=2026-10-26&end_date=2026-10-29`:

| Station | Agency | Points | Cadence | Nulls | Size (gzip) | Time |
|---|---|---|---|---|---|---|
| 737 วัดท่าเจดีย์ (TTC06) | HII | 540 | 10 min | 14 in 9 runs | 2,786 B | 0.22 s |
| 1097520 Gt.11 บ้านช้างแรก | RID | 90 | 1 h | 7 in 4 runs (around midnight each day) | 895 B | 0.17 s |
| 575568 BKC004 ปตร. คลองลัดบางยอ 2 | HII | 540 | 10 min | 497 (silent since 2026-10-27 19:40) | 2,168 B | 0.20 s |

The response is shaped:

```json
{"result":"OK","data":{"graph_data":[{"datetime":"2026-10-26 00:00","value":3.178,"value_out":null,"discharge":null}],"min_bank":2.139,"warning_level":null,"critical_level":null,"ground_level":-2.065,"qmax":null}}
```

Behaviour:
- **Range:** `start_date` and `end_date` are ICT calendar dates and both are inclusive. The series runs from `start_date 00:00` up to the latest reading on `end_date`, and future slots are not returned. For 72 h of history, request the last 4 calendar days.
- **Values:** `value` is a float or null. HII values have 3 decimals; RID values carry float32 artefacts (`40.159998`, `min_bank` 40.599998) and must be rounded to millimetres on parse.
- **Full-precision references:** the graph returns `min_bank` and `ground_level` at full precision. BKC004's `min_bank` is 0.905 here and 0.91 in `waterlevel_load`. With 0.905, the published 105.72 % reproduces exactly.
- **Bucketing:** the timeline is hourly, so 10-minute series must be reduced to one value per hour. The proposal for Phase 2 is the last non-null reading in `(HH−1:00, HH:00]`. That makes a 10-minute station and an hourly station mean the same thing at the same hour. The rule is documented in `docs/api.md` when it is written.
- **Trend:** the 6 h trend in the humanize spec should come from this series, not from `waterlevel_msl_previous`.

Other observations:
- `waterlevel_graph_oldcode?station_id=&agency_id=&start_date=&end_date=` also exists, addressing by agency code instead of id. Not needed.
- `waterlevel_graph_year?station_type=tele_waterlevel&station_id=&year=` exists for whole-year series. Not needed.

## 3. Other public endpoints (noted, not used)

Found in the ThaiWater web app's bundle (`https://www.thaiwater.net/dist/js/app.chunk.js`, fetched 2026-10-29, 1.14 MB). Each one listed with a size was fetched once, at concurrency 1.

| Endpoint | Observed | Possible use |
|---|---|---|
| `public/waterlevel?province_code=10,11,12,13,73,74` | 31 rows for Bangkok and its neighbours, same row shape as `waterlevel_load`; all 31 ids are in the load response. 7.8 KB. | A cheap per-province refresh. We do not need it, because one load call per hour is already polite. |
| `public/canal_waterlevel` | 282 canal gauges of สำนักการระบายน้ำ กรุงเทพมหานคร: 272 in Bangkok, the rest in Nonthaburi, Samut Prakan, Chachoengsao and Nakhon Pathom. Different schema (`canal_datetime`, `canal_value`, `station.canal_name`, `station.bank`). 222 readings were < 3 h old and 54 were > 30 days old. `bank` is present on 250. 28 KB. | **Relevant to the primary user's "canal behind the house" in Bangkok.** Out of scope for the brief, so it is listed as an open question. |
| `public/rain_24h` | 4,805 rain stations with `rain_24h` and `rain_1h`. 612 KB. | A rainfall layer. Out of scope. |
| `public/thaiwater_main` | `rain`, `temperature`, `dam` (35 large dams plus `summary4dam`) and `storm`. 570 KB. | Dam storage context. Out of scope. |
| `public/watergate_load`, `public/flow`, `public/flood_road`, `public/thailand`, `public/rain7day_forecast`, `public/storm_data` | Not fetched. | `rain7day_forecast` is a forecast, which the product deliberately does not show. |

We found no public endpoint that serves station metadata alone (the `station` block in `waterlevel_load` is empty). The metadata embedded in every row is enough.

## 4. Attribution and terms of use

What we could verify, each item fetched 2026-10-29:
- **Footer copyright:** the ThaiWater web app's footer reads **"Copyright © 2024 Hydro - Informatics Institute,All rights reserved."** (string in `app.chunk.js`). A second footer variant reads "Copyright © 2024 Power by National Hydroinformatics Data Center, All rights reserved."
- **Site title:** "คลังข้อมูลน้ำแห่งชาติ - Thaiwater.net | National Hydroinformatics Data Center" (`https://www.thaiwater.net/`).
- **Public access statement:** the app's about text says the national water data warehouse connects data from 58 agencies in 13 ministries and "เปิดให้บริการข้อมูลผ่านเว็บไซต์ www.Thaiwater.net และ แอพพลิเคชั่นบนโทรศัพท์มือถือ … ให้ประชาชนทั่วไปสามารถเข้าถึงข้อมูล". In other words, it opens data to the general public through the site and the apps.
- **Policies linked from the site:**
  - privacy policy: `https://pdpa.hii.or.th/privacy-policy/thaiwater-privacy-policy`
  - privacy notice: `https://pdpa.hii.or.th/privacy-notice/thaiwater-privacy-notice`
  - cookie policy: `https://pdpa.hii.or.th/cookies-policy/thaiwater-cookies-policy`

  The privacy notice covers personal data only. It says nothing about reuse of the water data.
- **No licence found:** we found no open-data licence (such as CC BY or the Thai open government licence), no API terms, and no data.go.th dataset for these readings. Searches are listed under Sources.
- **Original owners:** the readings belong to the agency named on each row. In this snapshot that is HII (สสน.), the Royal Irrigation Department (กรมชลประทาน), EGAT (กฟผ.) and the Friend in Need (of "Pa") Volunteers Foundation (มูลนิธิอาสาเพื่อนพึ่ง (ภาฯ) ยามยาก สภากาชาดไทย).

**Proposed credit**, in the UI next to every number, on the about page and in the README:

> ข้อมูลระดับน้ำจาก ThaiWater (คลังข้อมูลน้ำแห่งชาติ) โดยสถาบันสารสนเทศทรัพยากรน้ำ (องค์การมหาชน) และหน่วยงานเจ้าของสถานี · thaiwater.net

In the station panel, also name the station's own agency, e.g. "สถานีของกรมชลประทาน".

**How we stay a good citizen without a licence:**
- Link back to thaiwater.net.
- Never imitate ThaiWater's look or imply we are official.
- Fetch the national snapshot once an hour, plus a one-off 72 h backfill at concurrency ≤ 4.
- Send a User-Agent that names the project and its repo.
- Serve **derived** values (percent, distance, trend and our own sentence) with attribution, rather than mirroring the raw API.

## 5. Implications for the Phase 2 client and normaliser

- **DTO typing:**
  - Decode numbers that arrive as strings into a type that accepts a JSON string, a JSON number, `""` and null, for example a `json.Number`-backed `NullDecimal`.
  - Treat absent keys and null identically.
- **Bank references:**
  - Treat `min_bank == 0 && ground_level == 0` as "no reference".
  - Treat a missing `ground_level` as "no percent".
  - Neither case gets a level or an overflow claim, whatever `diff_wl_bank_text` says.
- **Rounding:**
  - Round every level to millimetres on parse, to remove float32 noise.
  - Prefer the graph's full-precision `min_bank` and `ground_level` when we have them. Otherwise use the load values, which are within 0.5 cm.
- **Timestamps:** parse `"2006-01-02 15:04"` in a fixed `UTC+7` zone (`time.FixedZone("ICT", 7*3600)`) so the binary does not depend on tzdata.
- **Polling:**
  - Poll `waterlevel_load` hourly.
  - The lag was partly measured on 2026-10-29: RID's HH:00 readings appear between HH:10 and HH:54 (see Freshness). A poll at :07 therefore stores the previous hour for hourly stations. The prompt suggests :07; Phase 2 should confirm over more hours before choosing the minute.
  - Upstream has no ETag, so hash the body to skip identical snapshots.
- **User-Agent:** `Mahanathi/<version> (+<repo URL>)`. The repo URL is not known yet, so it should come from config with a sensible default.

## Sources

All accessed 2026-10-29.

**Endpoints and site:**
- `https://api-v3.thaiwater.net/api/v1/thaiwater30/public/waterlevel_load` (live response, recorded)
- `https://api-v3.thaiwater.net/api/v1/thaiwater30/public/waterlevel_graph?station_type=tele_waterlevel&station_id={737,1097520,575568}&start_date=2026-10-26&end_date=2026-10-29` (live responses, recorded)
- `https://api-v3.thaiwater.net/api/v1/thaiwater30/public/waterlevel?province_code=10,11,12,13,73,74`, `…/public/canal_waterlevel`, `…/public/rain_24h`, `…/public/thaiwater_main` (live responses, each fetched once)
- `https://www.thaiwater.net/`, page HTML and `https://www.thaiwater.net/dist/js/app.chunk.js`: endpoint list, footer copyright, about text and policy links
- `https://pdpa.hii.or.th/privacy-notice/thaiwater-privacy-notice`: privacy notice, with no data-reuse terms

**Searches that found no licence or API terms:**
- "ThaiWater thaiwater.net เงื่อนไขการใช้งาน ข้อมูล สถาบันสารสนเทศทรัพยากรน้ำ terms of use"
- "api-v3.thaiwater.net thaiwater30 public API endpoints"
- "data.go.th สถาบันสารสนเทศทรัพยากรน้ำ ระดับน้ำ ชุดข้อมูล สัญญาอนุญาต"
- "สสน. hii.or.th บริการข้อมูล API ThaiWater ขอใช้ข้อมูล open data"
- "\"ThaiWater Challenge\" 2569 API ข้อมูล"

**Background on the platform (news results seen in search, not relied on for facts above):**
- `https://www.infoquest.co.th/?p=239329`
- `https://www.bangkokbiznews.com/tech/1243777`
- `https://twa.thaiwater.net/en/website-guide/Mainpage/homepage-guide`
