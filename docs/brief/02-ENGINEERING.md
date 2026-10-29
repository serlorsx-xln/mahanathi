# Engineering brief: มหานที

## 1. Data source (verified by the lead on 2026-10-29 ~14:00 ICT; re-verify everything)

### Current levels
`GET https://api-v3.thaiwater.net/api/v1/thaiwater30/public/waterlevel_load` (no key, ~1.4 MB JSON)
- Top-level keys: `waterlevel_data{result,data[]}`, `waterlevel_manual_data`, `basin`, `agency`, `station`, `scale`, `province`.
- `waterlevel_data.data[]` has about 804 rows. Each row has:
  - `id`
  - `waterlevel_datetime` (`"YYYY-MM-DD HH:MM"`, local ICT, no offset)
  - `waterlevel_msl` (string, metres above MSL)
  - `waterlevel_msl_previous`
  - `storage_percent` (string)
  - `situation_level` (1–5 or null)
  - `diff_wl_bank`
  - `station{ id, tele_station_name.th, tele_station_lat, tele_station_long, tele_station_oldcode, left_bank, right_bank, min_bank, ground_level, offset, qmax, is_key_station, warning_level_m, critical_level_m, critical_level_msl, … }`
  - `geocode{ area_name, province_name, amphoe_name, tumbon_name (th/en) }`
  - `basin{ basin_name th/en }`
  - `agency{ agency_name, agency_shortname }`
- `scale.data.scale` gives the official thresholds on `storage_percent`:

  | % of bank-full | Level | Label |
  |---|---|---|
  | >100 | 5 | น้ำล้นตลิ่ง |
  | >70 | 4 | น้ำมาก |
  | >30 | 3 | น้ำปกติ |
  | >10 | 2 | น้ำน้อย |
  | ≤10 | 1 | น้ำน้อยวิกฤติ |

- Snapshot at 14:00: level counts were 5→22, 4→145, 3→348, 2→211, 1→63 and null→15. Timestamps were mostly from the current or previous hour, with a few rows hours old, so **stale rows exist**.

> **Correction (lead, after the Phase 1 review):** `diff_wl_bank` is **unsigned** (`|wl − min_bank|`). The direction lives only in `diff_wl_bank_text`. Our code computes the signed difference itself and never reads the sign of this field. The thresholds are in `scale.data.rule_web`; `scale.data.scale` holds labels and colours. See ADR 0003 and `docs/research/data-quality.md`, which supersede the observation below.

### Data-quality observation (a lead, not a fact; verify across all stations)
Sample station Gt.11 บ้านช้างแรก (Prachuap Khiri Khan):
- `left_bank` 84.71, `right_bank` 84.34, `min_bank` 40.60, `ground_level` 37.72, `waterlevel_msl` 41.93
- `storage_percent` 146.18 equals `(wl − ground)/(min_bank − ground)×100`
- `diff_wl_bank` 1.33 equals `wl − min_bank`

So `left_bank` and `right_bank` look inconsistent with `min_bank` for this station (a different datum?). Do not compute from left and right blindly.
- Establish the formula ThaiWater uses and confirm it across every station.
- Write `docs/research/data-quality.md`: how many stations are consistent, the anomalies, missing bank values and what the product does in each case.
- This finding is a real engineering story for the interview, so document it precisely.

### History (for backfill and the 72-hour timeline)
`GET https://api-v3.thaiwater.net/api/v1/thaiwater30/public/waterlevel_graph?station_type=tele_waterlevel&station_id={station.id}&start_date=YYYY-MM-DD&end_date=YYYY-MM-DD`
- Returns `{result:"OK", data:{graph_data:[{datetime, value, value_out, discharge}]}}` with values that can be null. Values are every 10 minutes for HII stations and hourly for RID stations (corrected after Phase 1), so bucket them to hours (the last non-null reading in each hour).
- Use it to backfill 72 hours on first start and for newly seen stations.
- Concurrency must stay at 4 or lower, with exponential backoff and jitter.

### Good citizenship and licensing
- Set a User-Agent that names the project and its repo URL.
- Poll the load endpoint once per hour, a few minutes after the hour, and retry with backoff.
- Find the correct attribution and terms of use for ThaiWater/HII (สถาบันสารสนเทศทรัพยากรน้ำ, องค์การมหาชน) and put the credit in the UI and README.
- If the terms forbid redistribution, write it down as an open question and continue.

### Time
Source timestamps are ICT (UTC+7) without an offset. Store them as UTC instants and display them in ICT as "14:00 น.". Dates use the Thai short form, "29 ต.ค. 69".

## 2. Architecture (lead recommendation; confirm or change with an ADR)
```
mahanathi/
  server/            Go 1.25, single binary `cmd/mahanathi`
    internal/thaiwater   HTTP client + DTOs + recorded fixtures (testdata/)
    internal/normalize   DTO → domain Station/Reading, validation, suspect flags
    internal/humanize    PURE: domain → Thai sentence + structured facts. No I/O. Exhaustively tested.
    internal/store       SQLite (modernc.org/sqlite, pure Go, WAL), migrations, retention 30d
    internal/ingest      scheduler (hourly), backfill (72h), idempotent upserts
    internal/api         net/http with Go 1.22+ patterns, gzip, ETag, rate limit, security headers
    web/                 embed of ../web/dist (go:embed) so one binary serves app + API
  web/               Vite + React 19 + TypeScript (strict) + three + @react-three/fiber + drei (+ GSAP for timelines)
  geodata/           reproducible scripts → static assets (heightmap, borders, rivers) + SOURCES.md (licences)
  e2e/               Playwright
  docs/              briefs, research, adr, design, DEVLOG, PROGRESS
  process/           phase prompts, runner, reports (logs are gitignored)
```

**Why one binary plus SQLite:**
- 800 rows an hour is about 19k rows a day, so SQLite is plenty.
- One container is simple to deploy and simple to explain.
- There is no external DB to operate.

**UI stack:**
- No component kit (no shadcn, MUI or Chakra). Components are hand-built on the design tokens.
- Use plain CSS with custom properties (CSS Modules are fine).
- If Tailwind is used at all, its default theme must be fully replaced by our tokens, so no default palette or default font stack survives. Plain CSS is preferred.

### API (v1, JSON, gzip/brotli, ETag, `Cache-Control: public, max-age=300`)
- `GET /api/v1/now` returns:
  - stations as a compact array: id, name, lat, lon, province, amphoe, basin, river (if derivable), wl, bankFull, ground, pct, level, trend6h (m), measuredAt, stale, suspect
  - a summary: counts by level, top provinces at levels 4–5, the national sentence and generatedAt
- `GET /api/v1/stations/{id}/history?hours=72` returns that station's series plus its bank-full and ground references.
- `GET /api/v1/timeline?hours=72` returns the percent of bank-full per station per hour for scrubbing.
  - Budget: 800 × 72 values in ≤ 80 KB compressed.
  - Quantise percentages (for example, uint8 or uint16 steps) and document the encoding. Null means no data.
- `GET /healthz` (liveness) and `GET /readyz` (has data no older than 2 h).
- Validate inputs strictly. Rate limit per IP with a token bucket and answer 429 with `Retry-After`.

### Privacy
Geolocation and nearest-station search run in the browser. The server never receives coordinates. Use no cookies and no third-party trackers.

## 3. Humanize spec (the part a human must be able to defend line by line)
The input is a normalised reading plus its recent history. The output is `{ sentence, facts[], level, stale, suspect }`.

| Case | Behaviour |
|---|---|
| Normal | Distance to bank-full in metres with 2 decimals: "ต่ำกว่าตลิ่ง 1.40 เมตร". Above bank-full: "สูงกว่าตลิ่ง 0.35 เมตร (น้ำล้นตลิ่ง)". |
| Trend (6 h) | \|Δ\| < 0.05 m gives "ทรงตัว". Rising gives "สูงขึ้น 12 ซม. ใน 6 ชั่วโมงที่ผ่านมา". Falling gives "ลดลง …". Use cm below 1 m and m with 2 decimals at or above 1 m. If fewer than 4 hourly points exist, omit the trend rather than guess. |
| Stale | `measuredAt` older than 3 h means stale = true. The sentence leads with the age: "ข้อมูลล่าสุดเมื่อ 09:00 น. (ยังไม่อัปเดต)". Stale readings are excluded from "now" counts. |
| Missing bank-full or ground | Make no distance claim. Use the situation label if present, otherwise "สถานีนี้ยังไม่มีข้อมูลระดับตลิ่ง". |
| Suspect | Flag values outside plausible bounds (for example wl < ground − 20 m, a jump over 3 m/h, or pct > 400) as suspect. Do not present them as current, and say so plainly. |
| National sentence | Built from counts and top provinces, for example "ตอนนี้ 22 สถานีน้ำล้นตลิ่ง ส่วนใหญ่อยู่ที่พระนครศรีอยุธยา สุราษฎร์ธานี และชุมพร". It must be grammatical when counts are 0 or 1. |

Additional rules:
- **Banned phrases test:** no output may contain ปลอดภัย, จะท่วม, ไม่ท่วม, รับรอง, แน่นอน, or anything else that predicts or reassures. Fuzz the generator over all fixtures and assert this.
- **Golden tests:** at least 30 table cases, including real recorded fixtures and every edge above, plus golden files for sentences.
- Thai number and time formatting lives in one small, tested module that both Go and TS follow. Duplicate the formatter tests in TS if the client formats anything.

## 4. Front-end engineering
- **Geodata:** build the assets reproducibly with scripts in `geodata/`.
  - Elevation: AWS Terrain Tiles (Terrarium PNG, open data).
  - Borders and admin1: Natural Earth 10m (public domain).
  - Rivers: Natural Earth rivers, or OSM major waterways if better (ODbL needs attribution).
  - Output a heightmap of 1.5 MB or less (PNG or KTX2) plus compact binary or TopoJSON vectors.
  - Record every source and licence in `geodata/SOURCES.md`.
- **Rendering:**
  - R3F with `frameloop="demand"` and adaptive DPR.
  - Instanced gauges in a single draw call, with custom shader materials reading colour tokens.
  - Thai labels as DOM overlays.
  - Lazy-load the 3D chunk so the UI shell and sentence paint first.
- **Performance budgets:**
  - Initial JS under 300 KB gzipped, excluding the lazy 3D chunk.
  - LCP under 2.5 s on throttled "Fast 4G, mid-tier mobile".
  - 60 fps on desktop and 30 fps or better on a mid-range phone.
  - No long tasks over 200 ms after load.
- **Search:**
  - Normalise Thai input: strip tone marks for matching, collapse spaces, accept "อ." and "จ." prefixes.
  - Rank province, then amphoe, then station, then river.
  - Unit-test it.
- **Fallbacks:**
  - If WebGL is unavailable or a low-power heuristic fires, show a designed 2D map (SVG or canvas) using the same tokens.
  - Detect the LINE in-app browser and offer to open externally.

## 5. Testing and quality gates
- **Go:**
  - `go test ./...` with the race detector and `go vet`.
  - Coverage of `humanize` and `normalize` at 90% or more.
  - API tests with `httptest`.
  - Ingest tests with a fake client and recorded fixtures (trimmed real responses in `testdata/`).
- **Web:**
  - `tsc --noEmit`, ESLint, Vitest for pure logic (nearest station, search, timeline decode, formatters).
- **E2E (Playwright):**
  - load, then the national sentence is visible
  - search "อยุธยา", then the panel shows a sentence
  - mocked geolocation for ใกล้ฉัน
  - geolocation denied
  - reduced motion
  - WebGL disabled, then the fallback renders
  - API down, then the last snapshot message shows
- **Visual:** screenshots per the design gate.
- **CI:** a GitHub Actions workflow that runs Go tests and the web typecheck, tests and build on push. The lead pushes; you only write the workflow.

## 6. Security and operations
- Security headers: CSP (allow only self plus required font and tile origins, preferably none), HSTS, `X-Content-Type-Options`, `Referrer-Policy` and `Permissions-Policy` (geolocation=self).
- No secrets are needed. If one ever is, it comes from env and never goes into git.
- Use structured logs with `slog`. Log each ingest run with counts, duration and errors.
- **Dockerfile:**
  - Multi-stage build: pnpm build, then go build with embed, then a minimal runtime as a non-root user.
  - A `/data` volume holds SQLite, and there is a HEALTHCHECK.
  - Config comes from env vars documented in `.env.example`.
- **Deploy target:** the lead deploys to the owner's Coolify server. Your job is a Dockerfile that runs with `docker run -p 8080:8080 -v data:/data`.

## 7. Environment constraints
- The repo is on an **exFAT** volume. Hardlinks are not supported, so `.npmrc` needs `package-import-method=copy`. Run `git config core.fileMode false`. `.gitignore` must include `._*`, `.DS_Store`, `node_modules`, `dist`, `process/logs/`, `*.db*` and `/data`.
- Keep large downloads such as DEM tiles under `geodata/.cache/` (gitignored) on this volume. The internal disk is nearly full.
