# The Mahanathi API, v1

The API is served by the same binary as the web app, on the same port. It is
read-only, same-origin in practice (no CORS headers are sent), sets no
cookies and receives no coordinates. Every response carries
`X-Content-Type-Options: nosniff`, `Referrer-Policy: no-referrer`,
`X-Frame-Options: DENY` and
`Content-Security-Policy: default-src 'none'; frame-ancestors 'none'`.

Success responses are cacheable: `Cache-Control: public, max-age=300`, a
strong `ETag`, and a gzip variant when the client accepts one. Errors are
`no-store`. `If-None-Match` is honoured (RFC 9110 weak comparison), so a
client that keeps its ETag pays nothing for unchanged data.

Under `/api/`, each client address gets a token bucket (default: burst 30,
one token per 200 ms). When it is exhausted the answer is `429` with
`Retry-After`. Behind our own proxy the key is the address the proxy saw
(the last `X-Forwarded-For` entry, trusted only when the direct peer is a
private address); otherwise it is the direct peer.

Times are UTC instants in RFC 3339. Percent of bank-full, distances and
levels are computed by us from the full-precision references (ADR 0003) —
ThaiWater's own `storage_percent`, `situation_level` and `diff_wl_bank` are
never served.

## GET /api/v1/now

Every station's current state in human terms, plus the national summary.
Built once per data change; `generatedAt` is the clock it was built against.

```json
{
  "generatedAt": "2026-10-29T12:55:00Z",
  "stations": [
    {
      "id": 12479, "code": "BSP002", "name": "คลองขนาน",
      "lat": 12.511667, "lon": 99.800833,
      "province": "ประจวบคีรีขันธ์", "amphoe": "ทับสะแก", "basin": "เพชรบุรี-ประจวบคีรีขันธ์",
      "river": "คลองขนาน", "agency": "HII",
      "wl": 16.727, "measuredAt": "2026-10-29T12:40:00Z",
      "bankFull": 16.475, "ground": 10.959,
      "pct": 104.57, "level": 5, "distanceMM": 252,
      "trend6h": -0.481, "stale": false, "suspect": [],
      "sentence": "…",
      "facts": [{"key": "measuredAt", "text": "19:40 น."}, …]
    }
  ],
  "summary": {
    "stations": 37,
    "levels": [0, 12, 20, 3, 1],
    "noLevel": 0,
    "staleCount": 0,
    "suspectCount": 0,
    "topProvinces": ["ประจวบคีรีขันธ์"],
    "national": "ตอนนี้ 1 สถานีน้ำล้นตลิ่ง อยู่ที่ประจวบคีรีขันธ์"
  }
}
```

Field rules:

- `null` means "not derivable", never zero: a station without references has
  no `pct`, not a `pct` of 0. `wl`/`measuredAt` are null when the station has
  no reading in retention.
- `pct` is the percent of bank-full rounded to hundredths
  (`(wl − ground)/(bankFull − ground) × 100`), and `level` follows
  rule_web's thresholds on that arithmetic: 1 น้อย, 2 ปกติ, 3 ระวัง, 4 น้ำมาก,
  5 น้ำล้นตลิ่ง.
- `distanceMM` is `wl − bankFull` in millimetres, positive above the bank.
  It exists whenever the bank reference does, with or without ground.
- `trend6h` is metres over the hourly buckets before the reading's own hour
  (positive is rising), null when omitted: stale, suspect, or fewer than
  four points.
- `stale` means the reading is older than 3 hours. `suspect` lists the
  plausibility reasons (`below_ground`, `implausible_pct`, `jump`, `future`).
  Suspect values are served with their flags; the sentence is what withholds
  them from presentation.
- `sentence` and `facts` are the humanize output (brief §3). The facts carry
  the numbers behind the sentence, already formatted, in a stable order.

The summary counts only current stations (a reading no older than 3 hours):
`levels[i]` counts level `i+1` among trustworthy current stations;
`noLevel` counts those without a level; `staleCount` is excluded from all of
the above; `topProvinces` lists provinces holding level 4–5 stations, most
first. `national` is the national sentence, grammatical at every count.

The trend and the suspect jump check read the last 72 hours of readings, so
a station stale for longer than that gets no trend (its value is historical
anyway).

## GET /api/v1/stations/{id}/history?hours=72

One station's hourly buckets, oldest first. `hours` is a whole number from
1 to 168 (default 72); anything else is a 400. An unknown station is a 404.

```json
{
  "station": {"id": 3441, "code": "URTU09", "name": "…", "bankFull": 225.386, "ground": 220.11},
  "from": "2026-10-26T14:00:00Z", "until": "2026-10-29T13:00:00Z", "hours": 72,
  "points": [{"at": "2026-10-26T14:00:00Z", "wl": 224.931}, …, {"at": "…", "wl": null}]
}
```

Buckets are `(t − 1h, t]`, holding the last reading in the hour
(`waterlevel_msl_previous` is never read). Buckets without a reading are
present with `wl: null`, so a chart can keep its axis. This response is
built per request (it is small); its ETag is computed the same way as the
big ones.

## GET /api/v1/timeline

The percent of bank-full per station per hour, for scrubbing. This body is
**binary**, not JSON: 800 stations × 72 hourly values as JSON numbers are
96 KB compressed even on the smoothest data, and the budget is 80 KB
(brief §2). The measured worst case on synthetic rivers wilder than any
real one is 61 KB compressed.

- Content-Type: `application/octet-stream`; served with the same ETag, gzip
  and cache headers as everything else.
- The percent is the same arithmetic as `pct` above, in hundredths, clamped
  to [0, 65535]. Only stations with complete references (ADR 0003) and at
  least one value are included.

Layout, little-endian, integers unsigned:

```
offset  size  field
0       4     magic "MNTL"
4       1     version = 1
5       1     hours                  buckets in the window (72)
6       2     reserved = 0
8       8     fromUnix               the first bucket end, UTC seconds
16      4     stationCount
20      8×n   station ids, ascending
then, per station, in the same order:
        ⌈h/8⌉ presence bitmap       bit 7−i%8 of byte i/8 is set when
                                      bucket i has a value
        varint per present bucket   the first of each run (the start of the
                                      window, or the bucket after a gap) is
                                      the absolute value; the rest are
                                      zigzag varints of the difference from
                                      the previous present bucket
```

Varints are unsigned LEB128. Zigzag maps a signed delta to an unsigned
varint (`−1 → 1, 1 → 2, −63 → 125, 63 → 126`), so small deltas of either
sign are one byte.

Bucket `i` of a station is the hour ending at `fromUnix + i` hours; its
value divided by 100 is the percent of bank-full in that hour. `until` is
`fromUnix + (hours − 1)` hours.

A decoder (Go, from the tests; TypeScript will follow the same shape):

```go
p := 20
for each station {
    bitmap := body[p : p+bitmapLen]; p += bitmapLen
    prev, inRun := 0, false
    for i := 0; i < hours; i++ {
        if bitmap[i/8]&(1<<(7-i%8)) == 0 { inRun = false; continue }
        v, n := binary.Uvarint(body[p:]); p += n
        if inRun {
            prev += int64(v>>1) ^ -int64(v&1) // unzigzag
        } else {
            prev = int64(v)
        }
        value[i] = uint16(prev)
        inRun = true
    }
}
```

## GET /healthz, GET /readyz

Plain text. `/healthz` answers `ok` while the process is up. `/readyz`
answers `ready` when the store's newest reading is no older than 2 hours,
`stale data` otherwise — it reads the store, not the cached payloads, so it
stays honest when the ingester has stopped. Neither is rate limited; before
the first ingest finishes, `/api/v1/now` and `/api/v1/timeline` answer
`503` with `Retry-After: 30`.
