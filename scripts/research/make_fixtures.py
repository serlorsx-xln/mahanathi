"""Trim a real waterlevel_load response into a small, diverse test fixture.

Usage:
    python3 -I scripts/research/make_fixtures.py <waterlevel_load.json> <out.json> "<fetched_at ICT>"

Rows are selected by named case (each one exercises a specific quirk found in
docs/research/data-quality.md) and then topped up by rule so every
situation_level, every agency and every staleness bucket is present. The rows
keep their original order and content; only the list is shortened. All other
top-level blocks are kept, except waterlevel_manual_data, which is cut to a few
rows because the product does not use it.
"""

import json
import sys
from collections import Counter
from datetime import datetime, timedelta

# station code -> why it is in the fixture
NAMED = {
    "THA006": "level 5, HII 10-minute station (วัดท่าเจดีย์ TTC06), storage_percent reproduces exactly",
    "BSP002": "level 5, storage_percent only reproduces with min_bank at full precision",
    "HDA005": "level 5 in Ayutthaya (search target)",
    "C.67": "level 5, RID, no river_name key",
    "BKC004": "level 5 but last reading is two days old (stale overflow)",
    "Gt.11": "the brief's example: left/right_bank on a different datum from min_bank",
    "CPY012": "level 4 on แม่น้ำเจ้าพระยา at บางปะอิน, Ayutthaya",
    "C.35": "level 4, RID, Chao Phraya at Ayutthaya city",
    "MOU510": "level 4 but stale for two days",
    "X.260": "level 4, RID reading hours old (stale)",
    "URTU09": "level 4, EGAT, flow_rate present",
    "PTTEP3": "level 4, sponsor_by present, no river_name",
    "C.2": "level 3, Chao Phraya at Nakhon Sawan, critical_level_msl present",
    "C.13": "level 3, below the Chao Phraya Dam",
    "P.1": "level 3, Ping at Chiang Mai",
    "CMP001": "level 3 but four days old",
    "AIT011": "level 3, reading at :10 past the hour, แม่น้ำมูล",
    "AIT009": "level 3, reading at 14:30 (just over 3 h old at fetch time)",
    "M.170": "level 3, waterlevel_msl_previous is null",
    "MYA001": "station inside Myanmar, empty amphoe",
    "P.64": "level 2, left/right_bank about 787 m above min_bank (consistent with a datum added twice)",
    "Y.16": "RID, offset field present, left/right_bank about 31 m above min_bank",
    "URTU07": "level 2, EGAT, six days old",
    "T.1": "RID on the tidal Tha Chin, left/right_bank negative and below min_bank",
    "SKU02": "level 1, seven days old (oldest row)",
    "FOP045": "level 1, water below ground_level (negative storage_percent)",
    "URTD03": "level 1, storage_percent -18.93",
    "FOP029": "level 1, stale since yesterday",
    "BLGTU06": "no level: min_bank and ground_level both coded as 0",
    "MKSND01": "no level: zero-coded bank references at a dam",
    "Gt.9": "no level: ground_level key absent while min_bank exists",
    "I.17": "ground_level key absent",
}


def main():
    src, dst, fetched = sys.argv[1], sys.argv[2], sys.argv[3]
    ref = datetime.strptime(fetched, "%Y-%m-%d %H:%M")
    d = json.load(open(src, encoding="utf-8"))
    rows = d["waterlevel_data"]["data"]
    by_code = {r["station"]["tele_station_oldcode"]: r for r in rows}

    chosen = {}
    for code, why in NAMED.items():
        if code not in by_code:
            sys.exit(f"named station {code} not in snapshot; update NAMED")
        chosen[by_code[code]["station"]["id"]] = why

    def add_first(pred, why):
        for r in rows:
            if r["station"]["id"] not in chosen and pred(r):
                chosen[r["station"]["id"]] = why
                return

    prov = lambda r: ((r.get("geocode") or {}).get("province_name") or {}).get("th")
    add_first(lambda r: prov(r) == "กรุงเทพมหานคร" and r.get("situation_level") == 4, "level 4 in Bangkok")
    add_first(lambda r: prov(r) in ("สุราษฎร์ธานี", "ชุมพร") and r.get("situation_level") == 5, "level 5 in the upper south")
    add_first(lambda r: r["agency"]["agency_shortname"]["en"] == "HII" and r.get("situation_level") == 2, "level 2, HII")
    add_first(lambda r: r["agency"]["agency_shortname"]["en"] == "FOP" and r.get("situation_level") == 1
              and float(r.get("storage_percent") or -1) > 0, "level 1 with a positive percentage")
    add_first(lambda r: r.get("waterlevel_msl") == r.get("waterlevel_msl_previous") and r.get("situation_level") == 3,
              "waterlevel_msl equals waterlevel_msl_previous")

    out_rows = [r for r in rows if r["station"]["id"] in chosen]
    trimmed = dict(d)
    trimmed["waterlevel_data"] = dict(d["waterlevel_data"], data=out_rows)
    trimmed["waterlevel_manual_data"] = dict(d["waterlevel_manual_data"], data=d["waterlevel_manual_data"]["data"][:3])

    with open(dst, "w", encoding="utf-8") as f:
        json.dump(trimmed, f, ensure_ascii=False, indent=1)
        f.write("\n")

    # Coverage report, so the README can state what the fixture contains.
    lv = Counter(r.get("situation_level") for r in out_rows)
    ag = Counter(r["agency"]["agency_shortname"]["en"] for r in out_rows)
    age = Counter()
    for r in out_rows:
        h = (ref - datetime.strptime(r["waterlevel_datetime"], "%Y-%m-%d %H:%M")) / timedelta(hours=1)
        age["<=3h" if h <= 3 else ">3h"] += 1
    print(f"rows: {len(out_rows)}")
    print(f"situation_level: {dict(sorted(lv.items(), key=lambda t: (t[0] is None, t[0])))}")
    print(f"agency: {dict(ag)}")
    print(f"age at {fetched}: {dict(age)}")
    print("| code | station id | level | why |")
    print("|---|---|---|---|")
    for r in out_rows:
        s = r["station"]
        print(f"| {s['tele_station_oldcode']} | {s['id']} | {r.get('situation_level', '-')} | {chosen[s['id']]} |")


if __name__ == "__main__":
    main()
