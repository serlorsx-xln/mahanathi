# มหานที (Mahanathi)

**ระดับน้ำที่วัดได้จริง อ่านเป็นรูปทรง บอกเป็นภาษาไทย** — ทุกสถานีวัดระดับน้ำของประเทศไทย (804 สถานี) บนภูมิประเทศสามมิติ อัปเดตทุกชั่วโมง แตะที่สถานีใดก็ได้แล้วอ่านประโยคเดียวว่าแม่น้ำสายนั้นกำลังทำอะไร เช่น

> *"ต่ำกว่าตลิ่ง 1.40 เมตร สูงขึ้น 12 ซม. ใน 6 ชั่วโมงที่ผ่านมา · วัดเมื่อ 14:00 น."*

ภาพรวมทั้งประเทศก็เป็นประโยคเดียว คำนวณจากค่าที่วัดได้ เช่น "ตอนนี้ 1 สถานีน้ำล้นตลิ่ง อยู่ที่ประจวบคีรีขันธ์" — เราเล่าเฉพาะสิ่งที่วัดได้และเวลาที่วัด ไม่มีการพยากรณ์ งานนั้นเป็นของหน่วยงานที่รับผิดชอบ และขอบคุณข้อมูลเปิดจาก ThaiWater (HII · RID) ที่ทำให้เรื่องนี้เป็นไปได้

**เปิดใช้: <https://mahanathi.s3rx.net>**

---

## What this is

Every river gauge in Thailand, live, on a 3D relief of the country — measured water levels read as shapes (bank-full cap, fill, overflow blade, stale hollow), and one plain-Thai sentence per station and for the nation. Read-only, no accounts, no personal data; the API receives no coordinates. The design brief is `docs/brief/`.

## Architecture

One Go binary serves everything — the API and the embedded front end.

```
server/  Go 1.25, CGO_ENABLED=0
  internal/thaiwater  HTTP client for ThaiWater, with recorded fixtures
  internal/normalize  integer-millimetre arithmetic (ADR 0003)
  internal/domain     levels, percents, distances
  internal/humanize   the Thai sentence, banned-phrase rule test-enforced
  internal/store      SQLite (modernc.org/sqlite, pure Go, WAL), 30-day retention
  internal/ingest     hourly snapshot at :07 past the hour + 72 h backfill
  internal/api        /api/v1 endpoints, budgets, rate limits
  cmd/mahanathi       env config; embeds the web build via go:embed

web/     Vite + React 19 + MapLibre GL 6
  src/map        the globe relief, terrarium DEM, instanced WebGL gauge
                 layer (one draw call for all 804 stations, disc fallback)
  src/routes     the map surface and the evidence pages (/specimen, /color)
  src/design     tokens.ts, the single source (generated CSS, drift-tested)

geodata/ pinned, re-fetchable asset sources and build scripts
```

The API (`docs/api.md`): `GET /api/v1/now` (every station + the national
summary, 85 KB gzipped for 804 stations), `GET /api/v1/stations/{id}/history?hours=72`,
`GET /api/v1/timeline` (percent of bank-full per station per hour, 46 KB
gzipped against an 80 KB budget), plus `/healthz` and `/readyz`. ETags,
gzip variants and per-address rate limits. We compute percent-of-bank-full
ourselves from full-precision references; ThaiWater's own derived fields
are never served.

The decisions behind this shape (single binary, SQLite, bank-full reference,
MapLibre over R3F, typography, terrain exaggeration) are argued once in
`docs/adr/0001`–`0008` and not re-argued here.

## Running it

Needs Node 24, pnpm and Go 1.25.

```sh
make install    # pnpm install --frozen-lockfile
make dev        # the Go server and the Vite dev server together
```

Front end only: `pnpm --filter @mahanathi/web dev`. The full map surface
locally also needs the gitignored local basemap — the 160 MB Protomaps
PMTiles archive, built by the recorded commands in `geodata/SOURCES.md`
(the terrarium DEM ships with the repository).

```sh
make build                      # web build → go:embed → server/bin/mahanathi
PORT=3000 ./server/bin/mahanathi
```

The server listens on `:8080` by default (`PORT` or `MAHANATHI_ADDR`
override). On first start it ingests ThaiWater; the first ingest measured
33 s for 804 stations and 251,624 readings (`docs/PROGRESS.md`).

Also: `make test` (Go with the race detector + web unit), `make lint`,
`make e2e` (Playwright at 1440×900, 768×1024, 390×844), `make docker`
(the 21.4 MB distroless deploy image).

## How it was built

This repository is an honest record of an AI-built product, kept that way
on purpose.

- **The code was written by AI.** One implementing session worked through
  the phases (Claude Opus 5.5 through Phase 2, GLM-5.3 after — the switch
  is recorded in `docs/DEVLOG.md`). The product decisions are the owner's.
- **A tech lead reviewed every phase**, with adversarial multi-agent
  review. Its 25-finding review of the Phase 3b gate and the finding-by-
  finding disposition are in `docs/design/GATE-phase-3b.md`.
- **Design gates hold regenerable evidence.** Every screenshot and
  measurement regenerates from committed scripts
  (`e2e/tests/gate-phase3b.spec.ts`, `e2e/measure/` → `measurements.json`),
  not from one-off runs.
- **`docs/DEVLOG.md` records failures as they happened**, including AI
  mistakes caught by tests or review — e.g. a commit message that
  overclaimed its verification, and a critique score mis-summed, both
  corrected in the open.
- **Hard decisions have ADRs** (`docs/adr/0001`–`0008`).
- **No AI inside the product.** Every number on screen is computed by
  tested code; no prediction, no safety claims — the banned-phrase rule
  (`docs/brief/02-ENGINEERING.md`, humanize spec) is enforced by a test.

## Status: what ships, what's cut

Ships: the live globe relief with all 804 gauges, the Thai sentences
(station and national), the ใกล้ฉัน flight to a sample station (real
geolocation is the next unit — the button labels itself ตัวอย่างสถานี),
three camera presets, 2D/3D view modes, reduced-motion cuts, a no-WebGL
floor and the LINE in-app-browser offer.

Deliberately cut, planned rather than failed (`process/reports/phase-4.md`):

- the branded intro (built, then reverted before the interview — the
  honest choreography did not fit the minutes)
- the Tailwind v4 + shadcn/ui foundation (a half-hour unit done properly,
  scheduled post-interview)
- the shareable station deep-link (`?station=<id>` + Web Share)
- real-GPS ใกล้ฉัน (the geolocation twin of the current sample-station flight)
- the 72-hour timeline UI (the API serves it today; the chart is next)

## Data and credits

- **Live water levels**: ThaiWater, the National Hydroinformatics Data
  Center — HII · RID. The data provider, credited with thanks.
- **Basemap**: OpenStreetMap (ODbL), self-built as a Protomaps PMTiles
  archive with planetiler.
- **Terrain**: SRTM, GMTED2010 and ETOPO1 via AWS Terrain Tiles
  (USGS, NOAA); borders and provinces from Natural Earth (public domain).
- **Type**: IBM Plex Sans Thai (SIL OFL), self-hosted as Mahanathi Sans.

The full pinned-source and licence record, with rebuild commands, is
`geodata/SOURCES.md`.
