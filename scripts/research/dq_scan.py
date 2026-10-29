"""Data-quality scan of a ThaiWater waterlevel_load snapshot.

Usage:
    python3 -I scripts/research/dq_scan.py <waterlevel_load.json> [fetched_at_ict]

fetched_at_ict is "YYYY-MM-DD HH:MM" in ICT and is used for the staleness
histogram. It defaults to the newest waterlevel_datetime in the file.

The goal is to establish, from evidence, which station fields ThaiWater uses
to compute storage_percent ("percent of bank-full") and diff_wl_bank, and to
count the stations where the bank and ground references are missing or look
inconsistent. Stdlib only, so it runs anywhere Python 3.9+ does.
"""

import json
import statistics
import sys
from collections import Counter
from datetime import datetime, timedelta

# How close a recomputed value must be to the published one to count as a
# match. storage_percent is published with 2 decimals and the inputs with 2-3,
# so a small tolerance absorbs rounding of the inputs.
PCT_TOL = 0.06  # percentage points
DIFF_TOL = 0.006  # metres

# Thailand's extent, padded slightly. Used only to flag impossible coordinates.
LAT_RANGE = (5.0, 21.0)
LON_RANGE = (97.0, 106.0)


def num(v):
    """Parse ThaiWater's mixed string/number/empty fields into float or None."""
    if v is None:
        return None
    if isinstance(v, (int, float)):
        return float(v)
    s = str(v).strip()
    if s == "":
        return None
    try:
        return float(s)
    except ValueError:
        return None


def pct(wl, ground, top):
    if None in (wl, ground, top) or top == ground:
        return None
    return (wl - ground) / (top - ground) * 100.0


def level_from_pct(p, rules):
    """Apply scale.rule_web exactly as published (first matching rule wins)."""
    for r in rules:
        t = float(r["term"])
        if r["operator"] == ">" and p > t:
            return r["level"]
        if r["operator"] == "<=" and p <= t:
            return r["level"]
    return None


def pctile(xs, q):
    xs = sorted(xs)
    if not xs:
        return None
    i = min(len(xs) - 1, max(0, int(round(q * (len(xs) - 1)))))
    return xs[i]


def main():
    path = sys.argv[1]
    d = json.load(open(path, encoding="utf-8"))
    rows = d["waterlevel_data"]["data"]
    rules = d["scale"]["data"]["rule_web"]

    out = []
    p = out.append

    p(f"file: {path}")
    p(f"rows: {len(rows)}")
    ids = Counter(r["station"]["id"] for r in rows)
    p(f"distinct station ids: {len(ids)}; duplicated ids: {sum(1 for c in ids.values() if c > 1)}")
    p(f"station_type: {dict(Counter(r.get('station_type') for r in rows))}")
    p(f"agency: {dict(Counter(r['agency']['agency_shortname']['en'] for r in rows))}")

    # ---- field presence -------------------------------------------------
    p("\n## nulls / empty values")
    fields = {
        "waterlevel_msl": lambda r: r.get("waterlevel_msl"),
        "waterlevel_msl_previous": lambda r: r.get("waterlevel_msl_previous"),
        "waterlevel_m": lambda r: r.get("waterlevel_m"),
        "storage_percent": lambda r: r.get("storage_percent"),
        "situation_level": lambda r: r.get("situation_level"),
        "diff_wl_bank": lambda r: r.get("diff_wl_bank"),
        "discharge": lambda r: r.get("discharge"),
        "flow_rate": lambda r: r.get("flow_rate"),
        "station.left_bank": lambda r: r["station"].get("left_bank"),
        "station.right_bank": lambda r: r["station"].get("right_bank"),
        "station.min_bank": lambda r: r["station"].get("min_bank"),
        "station.ground_level": lambda r: r["station"].get("ground_level"),
        "station.offset": lambda r: r["station"].get("offset"),
        "station.qmax": lambda r: r["station"].get("qmax"),
        "station.warning_level_m": lambda r: r["station"].get("warning_level_m"),
        "station.critical_level_m": lambda r: r["station"].get("critical_level_m"),
        "station.critical_level_msl": lambda r: r["station"].get("critical_level_msl"),
        "station.tele_station_lat": lambda r: r["station"].get("tele_station_lat"),
        "river_name": lambda r: r.get("river_name") or None,
        "basin": lambda r: r.get("basin"),
        "geocode.amphoe_name.th": lambda r: (r.get("geocode") or {}).get("amphoe_name", {}).get("th") or None,
    }
    for name, get in fields.items():
        vals = [get(r) for r in rows]
        missing = sum(1 for v in vals if v is None or v == "")
        types = Counter(type(v).__name__ for v in vals if v is not None and v != "")
        p(f"  {name:32s} missing {missing:4d}  types {dict(types)}")

    # ---- per-row derived values ----------------------------------------
    recs = []
    for r in rows:
        s = r["station"]
        recs.append(
            dict(
                id=s["id"],
                code=s.get("tele_station_oldcode"),
                name=s["tele_station_name"].get("th"),
                prov=(r.get("geocode") or {}).get("province_name", {}).get("th"),
                dt=r.get("waterlevel_datetime"),
                wl=num(r.get("waterlevel_msl")),
                wlp=num(r.get("waterlevel_msl_previous")),
                sp=num(r.get("storage_percent")),
                lvl=r.get("situation_level"),
                dwb=num(r.get("diff_wl_bank")),
                dwb_text=r.get("diff_wl_bank_text"),
                left=num(s.get("left_bank")),
                right=num(s.get("right_bank")),
                minb=num(s.get("min_bank")),
                ground=num(s.get("ground_level")),
                lat=num(s.get("tele_station_lat")),
                agency=r["agency"]["agency_shortname"]["en"],
                lon=num(s.get("tele_station_long")),
            )
        )

    # ---- storage_percent formula ---------------------------------------
    p("\n## storage_percent: which bank-full reference reproduces it?")
    with_sp = [x for x in recs if x["sp"] is not None and x["wl"] is not None]
    p(f"rows with wl and storage_percent: {len(with_sp)}")
    candidates = {
        "A (wl-ground)/(min_bank-ground)": lambda x: pct(x["wl"], x["ground"], x["minb"]),
        "B (wl-ground)/(min(left,right)-ground)": lambda x: pct(
            x["wl"], x["ground"], min(v for v in (x["left"], x["right"]) if v is not None)
        )
        if (x["left"] is not None or x["right"] is not None)
        else None,
        "C (wl-ground)/(left_bank-ground)": lambda x: pct(x["wl"], x["ground"], x["left"]),
        "D (wl-ground)/(right_bank-ground)": lambda x: pct(x["wl"], x["ground"], x["right"]),
        "E wl/min_bank (no ground)": lambda x: (x["wl"] / x["minb"] * 100) if x["minb"] else None,
    }
    best = {}
    for label, f in candidates.items():
        match = near = computable = 0
        for x in with_sp:
            c = f(x)
            if c is None:
                continue
            computable += 1
            if abs(c - x["sp"]) <= PCT_TOL:
                match += 1
            elif abs(c - x["sp"]) <= 1.0:
                near += 1
        best[label] = match
        p(f"  {label:42s} computable {computable:4d}  match(±{PCT_TOL}) {match:4d}  within 1pp {near:4d}")

    # The load endpoint rounds min_bank to 2 decimals, while left/right_bank
    # keep 3. Where min(left,right) rounds to min_bank, it is the same number
    # at full precision, so recompute A with it.
    def precise_min(x):
        lr = [v for v in (x["left"], x["right"]) if v is not None]
        if lr and abs(min(lr) - x["minb"]) <= 0.005:
            return min(lr)
        return x["minb"]

    a_precise = lambda x: pct(x["wl"], x["ground"], precise_min(x))
    m = sum(1 for x in with_sp if a_precise(x) is not None and abs(a_precise(x) - x["sp"]) <= PCT_TOL)
    within1 = sum(1 for x in with_sp if a_precise(x) is not None and abs(a_precise(x) - x["sp"]) <= 1.0)
    p(f"  A' as A, with min_bank at full precision      computable {len(with_sp):4d}  match(±{PCT_TOL}) {m:4d}  within 1pp {within1:4d}")
    resid = [abs(a_precise(x) - x["sp"]) for x in with_sp if a_precise(x) is not None]
    p(f"  A' residual |recomputed - published| pp: median {statistics.median(resid):.3f}  max {max(resid):.3f}")
    # What the remaining residual is worth in water level: invert A' for wl.
    wl_err = [
        abs(x["sp"] / 100 * (precise_min(x) - x["ground"]) + x["ground"] - x["wl"])
        for x in with_sp
        if a_precise(x) is not None and abs(a_precise(x) - x["sp"]) > PCT_TOL
    ]
    if wl_err:
        p(f"  for rows A' misses, the implied water level differs from the published one by: max {max(wl_err)*100:.2f} cm, median {statistics.median(wl_err)*100:.2f} cm")

    # Rows where formula A does not match: describe them.
    a = candidates["A (wl-ground)/(min_bank-ground)"]
    a_miss = []
    for x in with_sp:
        c = a(x)
        if c is None or abs(c - x["sp"]) > PCT_TOL:
            a_miss.append((x, c))
    p(f"  rows where A does not reproduce storage_percent: {len(a_miss)}")
    for x, c in a_miss[:8]:
        p(
            f"    {x['code']!s:10s} {x['name']!s:28s} wl={x['wl']} ground={x['ground']} min={x['minb']} "
            f"L={x['left']} R={x['right']} sp={x['sp']} A={None if c is None else round(c, 2)}"
        )

    # ---- diff_wl_bank formula -------------------------------------------
    p("\n## diff_wl_bank: which reference reproduces it?")
    with_d = [x for x in recs if x["dwb"] is not None and x["wl"] is not None]
    p(f"rows with wl and diff_wl_bank: {len(with_d)}")
    for label, ref in {
        "wl - min_bank": lambda x: x["minb"],
        "wl - min(left,right)": lambda x: min(v for v in (x["left"], x["right"]) if v is not None)
        if (x["left"] is not None or x["right"] is not None)
        else None,
        "wl - left_bank": lambda x: x["left"],
    }.items():
        m = mabs = comp = 0
        for x in with_d:
            rv = ref(x)
            if rv is None:
                continue
            comp += 1
            if abs((x["wl"] - rv) - x["dwb"]) <= DIFF_TOL:
                m += 1
            if abs(abs(x["wl"] - rv) - x["dwb"]) <= DIFF_TOL:
                mabs += 1
        p(f"  {label:24s} computable {comp:4d}  signed match(±{DIFF_TOL} m) {m:4d}  |abs| match {mabs:4d}")
    p(f"  diff_wl_bank < 0: {sum(1 for x in with_d if x['dwb'] < 0)}")
    over_text = lambda x: bool(x["dwb_text"]) and "ล้น" in x["dwb_text"]
    p(
        "  text says ล้นตลิ่ง exactly when wl > min_bank: "
        f"{sum(1 for x in with_d if over_text(x) == (x['wl'] > x['minb']))}/{len(with_d)}"
    )
    p(
        "  rows the text calls ล้นตลิ่ง: "
        f"{sum(1 for x in with_d if over_text(x))}, by situation_level "
        f"{dict(Counter(x['lvl'] for x in with_d if over_text(x)))}"
    )
    p(f"  diff_wl_bank_text values: {dict(Counter(x['dwb_text'] for x in recs))}")

    # ---- bank references ------------------------------------------------
    p("\n## bank references")
    have_all = [x for x in recs if None not in (x["left"], x["right"], x["minb"])]
    eq_min = [x for x in have_all if abs(x["minb"] - min(x["left"], x["right"])) <= 0.011]
    p(f"stations with left, right and min_bank: {len(have_all)}")
    p(f"  min_bank == min(left,right) within 1 cm: {len(eq_min)}")
    gaps = sorted(
        ((min(x["left"], x["right"]) - x["minb"], x) for x in have_all if x not in eq_min),
        key=lambda t: -abs(t[0]),
    )
    p(f"  min_bank differs from min(left,right): {len(gaps)}")
    buckets = Counter()
    for g, _ in gaps:
        ag = abs(g)
        buckets["<0.5 m" if ag < 0.5 else "0.5-2 m" if ag < 2 else "2-10 m" if ag < 10 else ">=10 m"] += 1
    p(f"    |min(left,right) - min_bank| buckets: {dict(buckets)}")
    for g, x in gaps[:6]:
        p(
            f"    {x['code']!s:10s} {x['name']!s:28s} L={x['left']} R={x['right']} min={x['minb']} "
            f"ground={x['ground']} wl={x['wl']} gap={g:+.2f}"
        )
    agencies_big = Counter(x["agency"] for g, x in gaps if abs(g) >= 2)
    p(f"    gaps >= 2 m by agency: {dict(agencies_big)} (agency totals: {dict(Counter(x['agency'] for x in recs))})")
    # Hypothesis: for these stations left/right_bank had a datum near the
    # ground level added a second time. If so, subtracting ground_level should
    # land them near min_bank.
    resid = [
        (min(x["left"], x["right"]) - x["ground"]) - x["minb"]
        for g, x in gaps
        if abs(g) >= 2 and x["ground"] is not None
    ]
    if resid:
        rb = Counter()
        for r_ in resid:
            ar = abs(r_)
            rb["<1 m" if ar < 1 else "1-3 m" if ar < 3 else "3-10 m" if ar < 10 else ">=10 m"] += 1
        p(
            "    double-datum test, |(min(left,right) - ground) - min_bank|: "
            f"n {len(resid)}  median {statistics.median([abs(v) for v in resid]):.2f} m  buckets {dict(rb)}"
        )
        p(f"    (for comparison, median |min(left,right) - min_bank| in the same rows: {statistics.median([abs(g) for g, x in gaps if abs(g) >= 2 and x['ground'] is not None]):.2f} m)")
    zero = [x for x in recs if x["minb"] == 0 and x["ground"] == 0]
    p(f"  min_bank == ground_level == 0 (missing bank data coded as zero): {len(zero)}")
    for x in zero:
        p(f"    {x['code']!s:10s} {x['name']!s:28s} agency={x['agency']} wl={x['wl']} text={x['dwb_text']} level={x['lvl']}")
    no_min = [x for x in recs if x["minb"] is None]
    no_ground = [x for x in recs if x["ground"] is None]
    p(f"  min_bank missing: {len(no_min)}; ground_level missing: {len(no_ground)}")
    p(f"  min_bank missing but left/right present: {sum(1 for x in no_min if x['left'] is not None or x['right'] is not None)}")
    bad_order = [x for x in recs if None not in (x["minb"], x["ground"]) and x["ground"] >= x["minb"]]
    p(f"  ground_level >= min_bank (bank-full height <= 0): {len(bad_order)}")
    for x in bad_order[:10]:
        p(f"    {x['code']!s:10s} {x['name']!s:28s} ground={x['ground']} min={x['minb']} sp={x['sp']}")
    depth = [x["minb"] - x["ground"] for x in recs if None not in (x["minb"], x["ground"]) and x["minb"] > x["ground"]]
    if depth:
        p(
            "  channel depth (min_bank - ground) m: "
            f"min {min(depth):.2f}  p10 {pctile(depth, .1):.2f}  median {statistics.median(depth):.2f}  "
            f"p90 {pctile(depth, .9):.2f}  max {max(depth):.2f}"
        )

    # ---- situation_level ------------------------------------------------
    p("\n## situation_level")
    p(f"  counts: {dict(sorted(Counter(x['lvl'] for x in recs).items(), key=lambda t: (t[0] is None, t[0])))}")
    agree = disagree = 0
    dis = []
    for x in recs:
        if x["sp"] is None or x["lvl"] is None:
            continue
        lv = level_from_pct(x["sp"], rules)
        if lv == x["lvl"]:
            agree += 1
        else:
            disagree += 1
            dis.append((x, lv))
    p(f"  level == rule_web(storage_percent): {agree}; disagree: {disagree}")
    for x, lv in dis[:10]:
        p(f"    {x['code']} sp={x['sp']} level={x['lvl']} rule_web={lv}")
    p(f"  level null but storage_percent present: {sum(1 for x in recs if x['lvl'] is None and x['sp'] is not None)}")
    p(f"  level present but storage_percent null: {sum(1 for x in recs if x['lvl'] is not None and x['sp'] is None)}")

    # ---- plausibility ---------------------------------------------------
    p("\n## plausibility (the brief's suspect rules, applied to the published numbers)")
    low = [x for x in recs if None not in (x["wl"], x["ground"]) and x["wl"] < x["ground"] - 20]
    p(f"  wl < ground - 20 m: {len(low)}")
    hi = [x for x in recs if x["sp"] is not None and x["sp"] > 400]
    p(f"  storage_percent > 400: {len(hi)}")
    for x in hi[:10]:
        p(f"    {x['code']} {x['name']} sp={x['sp']} wl={x['wl']} ground={x['ground']} min={x['minb']}")
    neg = [x for x in recs if x["sp"] is not None and x["sp"] < 0]
    p(f"  storage_percent < 0 (water below the channel bed reference): {len(neg)}")
    for x in neg[:10]:
        p(f"    {x['code']} {x['name']} sp={x['sp']} wl={x['wl']} ground={x['ground']} min={x['minb']}")
    jump = [x for x in recs if None not in (x["wl"], x["wlp"]) and abs(x["wl"] - x["wlp"]) > 3]
    p(f"  |wl - wl_previous| > 3 m: {len(jump)}")
    same = sum(1 for x in recs if None not in (x["wl"], x["wlp"]) and x["wl"] == x["wlp"])
    p(f"  wl == wl_previous exactly: {same} of {sum(1 for x in recs if None not in (x['wl'], x['wlp']))}")
    coords = [x for x in recs if x["lat"] is None or not (LAT_RANGE[0] <= x["lat"] <= LAT_RANGE[1]) or not (LON_RANGE[0] <= x["lon"] <= LON_RANGE[1])]
    p(f"  coordinates missing or outside Thailand's padded bbox: {len(coords)}")
    for x in coords[:10]:
        p(f"    {x['code']} {x['name']} lat={x['lat']} lon={x['lon']}")

    # ---- timestamps -----------------------------------------------------
    p("\n## timestamps (ICT, no offset in the source)")
    dts = [datetime.strptime(x["dt"], "%Y-%m-%d %H:%M") for x in recs if x["dt"]]
    ref = datetime.strptime(sys.argv[2], "%Y-%m-%d %H:%M") if len(sys.argv) > 2 else max(dts)
    p(f"  reference time: {ref:%Y-%m-%d %H:%M} ICT; newest {max(dts):%Y-%m-%d %H:%M}; oldest {min(dts):%Y-%m-%d %H:%M}")
    minutes = Counter(dt.minute for dt in dts)
    p(f"  minute-of-hour of readings: {dict(sorted(minutes.items()))}")
    ages = Counter()
    for dt in dts:
        h = (ref - dt) / timedelta(hours=1)
        ages["<=1h" if h <= 1 else "1-3h" if h <= 3 else "3-24h" if h <= 24 else "1-30d" if h <= 720 else ">30d"] += 1
    p(f"  age buckets: {dict(ages)}")
    stale = sorted((x for x in recs if x["dt"] and ref - datetime.strptime(x["dt"], "%Y-%m-%d %H:%M") > timedelta(hours=3)), key=lambda x: x["dt"])
    p(f"  stale (> 3 h old): {len(stale)}; of which still carry a situation_level: {sum(1 for x in stale if x['lvl'] is not None)}")
    for x in stale[:12]:
        p(f"    {x['dt']} {x['code']} {x['name']} level={x['lvl']} sp={x['sp']}")

    # ---- named example ----------------------------------------------------
    p("\n## the brief's example station (Gt.11 บ้านช้างแรก)")
    for x in recs:
        if (x["name"] or "").find("ช้างแรก") >= 0 or (x["code"] or "").upper() in ("GT.11", "GT11"):
            for k in ("id", "code", "name", "prov", "dt", "wl", "sp", "lvl", "dwb", "left", "right", "minb", "ground"):
                p(f"    {k}: {x[k]}")
            p(f"    A recomputed: {pct(x['wl'], x['ground'], x['minb'])}")

    print("\n".join(out))


if __name__ == "__main__":
    main()
