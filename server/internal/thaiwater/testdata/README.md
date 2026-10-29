# ThaiWater fixtures

Real responses from the public ThaiWater API, recorded on **2026-10-29 at 17:54 ICT** (10:54 UTC) by the research for `docs/research/thaiwater-api.md`. Unit tests read these files; they never call the network.

| File | Source request | Notes |
|---|---|---|
| `waterlevel_load.json` | `GET /api/v1/thaiwater30/public/waterlevel_load` | Trimmed from 804 rows to 37 by `scripts/research/make_fixtures.py`. The row contents and their order are untouched, and every other top-level block is kept whole except `waterlevel_manual_data` (cut to 3 rows; the product does not use it). Re-serialised with one-space indentation. |
| `waterlevel_graph_737.json` | `GET .../waterlevel_graph?station_type=tele_waterlevel&station_id=737&start_date=2026-10-26&end_date=2026-10-29` | วัดท่าเจดีย์ (TTC06), HII, 10-minute cadence, above bank-full, scattered nulls. Byte-for-byte as received. |
| `waterlevel_graph_1097520.json` | same, `station_id=1097520` | Gt.11 บ้านช้างแรก, RID, hourly, float32 noise (`40.159998`), nightly null runs, falling 1.77 m in four hours. Byte-for-byte. |
| `waterlevel_graph_575568.json` | same, `station_id=575568` | BKC004 ปตร. คลองลัดบางยอ 2, HII, stopped reporting at 2026-10-27 19:40, so 497 of 540 values are null. Byte-for-byte. |

SHA-256 of the untrimmed load response: `8956900d604654e724eefd3795addf634a8352919383ad1cb9eaca9cec15887a` (1,395,146 bytes, kept out of git in `.cache/thaiwater/raw/`).

To rebuild the trimmed file from a fresh snapshot:

```sh
python3 -I scripts/research/make_fixtures.py <raw waterlevel_load.json> server/internal/thaiwater/testdata/waterlevel_load.json "YYYY-MM-DD HH:MM"
```

The script stops if a named station is missing, so a fresh snapshot can change the selection; the table below describes this recording.

## Coverage of `waterlevel_load.json`

- rows: 37
- situation_level: {1: 5, 2: 3, 3: 9, 4: 9, 5: 7, None: 4}
- agency: {'HII': 10, 'RID': 15, 'FOP': 6, 'EGAT': 6}
- age at 2026-10-29 17:54: {'<=3h': 26, '>3h': 11}

| code | station id | level | why |
|---|---|---|---|
| THA006 | 737 | 5 | level 5, HII 10-minute station (วัดท่าเจดีย์ TTC06), storage_percent reproduces exactly |
| BSP002 | 12479 | 5 | level 5, storage_percent only reproduces with min_bank at full precision |
| X.248 | 999395 | 5 | level 5 in the upper south |
| C.67 | 1095849 | 5 | level 5, RID, no river_name key |
| T.1 | 2676 | 5 | RID on the tidal Tha Chin, left/right_bank negative and below min_bank |
| BKC004 | 575568 | 5 | level 5 but last reading is two days old (stale overflow) |
| HDA005 | 471584 | 5 | level 5 in Ayutthaya (search target) |
| PTTEP3 | 1373691 | 4 | level 4, sponsor_by present, no river_name |
| C.12 | 2599 | 4 | level 4 in Bangkok |
| CPY012 | 49 | 4 | level 4 on แม่น้ำเจ้าพระยา at บางปะอิน, Ayutthaya |
| Y.16 | 2941 | 4 | RID, offset field present, left/right_bank about 31 m above min_bank |
| C.35 | 2609 | 4 | level 4, RID, Chao Phraya at Ayutthaya city |
| MOU510 | 1130137 | 4 | level 4 but stale for two days |
| Gt.11 | 1097520 | 4 | the brief's example: left/right_bank on a different datum from min_bank |
| X.260 | 726667 | 4 | level 4, RID reading hours old (stale) |
| URTU09 | 3441 | 4 | level 4, EGAT, flow_rate present |
| FOP038 | 740739 | 3 | waterlevel_msl equals waterlevel_msl_previous |
| C.13 | 2744 | 3 | level 3, below the Chao Phraya Dam |
| P.1 | 3226 | 3 | level 3, Ping at Chiang Mai |
| C.2 | 2795 | 3 | level 3, Chao Phraya at Nakhon Sawan, critical_level_msl present |
| CMP001 | 835 | 3 | level 3 but four days old |
| MYA001 | 1225528 | 3 | station inside Myanmar, empty amphoe |
| AIT011 | 740540 | 3 | level 3, reading at :10 past the hour, แม่น้ำมูล |
| M.170 | 2707 | 3 | level 3, waterlevel_msl_previous is null |
| AIT009 | 740538 | 3 | level 3, reading at 14:30 (just over 3 h old at fetch time) |
| PLG001 | 1126506 | 2 | level 2, HII |
| URTU07 | 3519 | 2 | level 2, EGAT, six days old |
| P.64 | 3112 | 2 | level 2, left/right_bank about 787 m above min_bank (consistent with a datum added twice) |
| MOU322 | 1476456 | 1 | level 1 with a positive percentage |
| SKU02 | 1113290 | 1 | level 1, seven days old (oldest row) |
| FOP029 | 740730 | 1 | level 1, stale since yesterday |
| FOP045 | 740746 | 1 | level 1, water below ground_level (negative storage_percent) |
| URTD03 | 3526 | 1 | level 1, storage_percent -18.93 |
| BLGTU06 | 562804 | - | no level: min_bank and ground_level both coded as 0 |
| I.17 | 3296 | - | ground_level key absent |
| Gt.9 | 2659 | - | no level: ground_level key absent while min_bank exists |
| MKSND01 | 700551 | - | no level: zero-coded bank references at a dam |

"Stale" means older than 3 hours at the recording time, per the humanize spec. Note that most stale rows still carry a `situation_level`; ThaiWater counts them, and we must not.

Source: ThaiWater, คลังข้อมูลน้ำแห่งชาติ (National Hydroinformatics Data Center), operated by สถาบันสารสนเทศทรัพยากรน้ำ (องค์การมหาชน) / Hydro-Informatics Institute (HII). The readings themselves come from the agencies named in each row (HII, RID, EGAT and the Friend in Need (of "Pa") Volunteers Foundation in this file). thaiwater.net states "Copyright © 2024 Hydro - Informatics Institute, All rights reserved." and publishes no explicit data licence that we could find; these trimmed copies exist only to test our parser. See `docs/research/thaiwater-api.md` for the attribution research and the open question about terms.
