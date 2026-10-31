# Phase 5 report: the repository README (one unit)

- **Implementer:** the same session (GLM-5.3 via the owner's router).
- **Phase window:** 2026-10-31 ~10:45–11:05 ICT. Commit: `3ecabaf`.

## What landed

`README.md` at the repo root (`3ecabaf`), the first the interview panel
will see. Structure, per the prompt:

1. **Thai voice first** — what มหานที is, the live URL, the one honest
   claim (measured levels, read as shapes, in plain Thai), the station
   sentence quoted from the product brief and the national sentence from
   `docs/api.md` (both committed evidence), and the no-prediction line.
2. **English architecture** — the package map as verified against
   `server/internal/` and `web/src/` (not from memory), the `/api/v1`
   endpoints with their measured budgets (85 KB / 46 KB vs 80 KB,
   `process/reports/phase-2.md`), pointers to ADR 0001–0008.
3. **Run commands** — read from the Makefile and `config.go`: `make
   install` / `make dev`, `PORT=3000 ./server/bin/mahanathi` (default
   `:8080`, `MAHANATHI_ADDR` overrides), the 33 s first ingest
   (`docs/PROGRESS.md`), the gitignored basemap/DEM note from
   `geodata/SOURCES.md`.
4. **The honest AI process** — AI-written code (with the model switch
   recorded), the lead's adversarial reviews, regenerable gate evidence,
   the DEVLOG's recorded AI mistakes, no AI inside the product.
5. **The cut list** — all five items from the prompt, verified against
   the tree (the intro is in history as landed-then-reverted `c6b8449`;
   ใกล้ฉัน's geolocation twin is absent from `LookdevMap.tsx`), framed
   as planned work.
6. **Credits** — ThaiWater (HII · RID), OSM/Protomaps, terrain, fonts,
   linking `geodata/SOURCES.md`.

## Corrections made during the unit (the check-the-file rule earning its keep)

- **Terrain credit:** the prompt said "HydroSHEDS/AWS terrain tiles";
  `geodata/SOURCES.md` says SRTM + GMTED2010 (USGS) and ETOPO1 (NOAA)
  via AWS Terrain Tiles. The README follows the committed record.
- **Two overclaims caught against the code before commit:** "the ใกล้ฉัน
  button computes distance on the device" (real GPS is cut — the button
  flies to a sample station) and a garbled "search-free browse" line
  (there is no search UI). Both reworded to what ships.

## Verification

Banned-phrase scan of the README: clean (the only "พยากรณ์" occurrence
is the allowed "ไม่มีการพยากรณ์" framing). Prettier ignores `*.md`, so
the CI format step is unaffected. Suites at commit time: `make test`
green — Go with the race detector, web unit **35/35**. A docs-only
change; tsc/eslint/e2e not re-run (no code touched).

## Stop

Stopped after the closing docs commit, per the prompt. The lead pushes
and decides everything else. Interview is today 16:00 ICT.
