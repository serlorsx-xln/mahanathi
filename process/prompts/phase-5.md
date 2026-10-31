# Phase 5: one unit only — the repository README

Same session, one unit, then stop. The repository has no root `README.md`,
and the interview panel will open the GitHub repo before anything else. Write
`README.md` at the repo root.

## What it must carry

1. **The Thai voice first (3–5 lines, plain Thai).** What มหานที is, the live
   URL (<https://mahanathi.s3rx.net>), and the one honest claim the product
   makes: measured water levels, read as shapes, in plain Thai. The national
   sentence idea belongs here in one line.
2. **An English body for engineers** (repo docs are English): the architecture
   in one diagram or list — Go 1.25 single binary embedding the Vite+React
   front end (`go:embed`), SQLite (modernc, WAL) hourly snapshots + 72 h
   backfill, exact integer-millimetre arithmetic, MapLibre GL 6 globe with the
   terrarium DEM and the instanced WebGL gauge layer, `/api/v1` endpoints with
   budgets. Point at `docs/adr/0001..0008` instead of re-arguing decisions.
3. **How to run it.** Dev (`pnpm dev` on the web workspace) and production
   (`make build` → single binary, `PORT=3000`); note the first ingest takes
   ~33 s against ThaiWater. Do not invent commands — read the Makefile and
   `web/package.json` and cite what exists.
4. **How it was built — the honest AI process.** One implementer session
   across phases, a lead reviewing each phase with adversarial multi-agent
   review, design gates with committed evidence (`docs/design/GATE-*.md`),
   `docs/DEVLOG.md` recording failures including AI mistakes, ADRs for hard
   decisions. State plainly that the code was written by AI and the product
   decisions are the owner's. This is a feature of the entry, not a footnote.
5. **Status: what ships and what is deliberately cut.** Name the cut list from
   the latest phase reports (branded intro, shadcn/Tailwind migration,
   station deep-link, real-GPS ใกล้ฉัน, the 72 h UI chart) as planned work,
   not failures.
6. **Data and credits.** ThaiWater open data (HII · RID), basemap
   OpenStreetMap ODbL / Protomaps, terrain HydroSHEDS/AWS terrain tiles,
   IBM Plex Sans Thai (OFL, self-hosted as Mahanathi Sans). `geodata/SOURCES.md`
   has the full list — summarize and link it.

## Hard rules (they apply to the README too)

- The banned-phrase test governs product copy, and the README is product
  copy: never ปลอดภัย, จะท่วม, ไม่ท่วม, รับรอง, แน่นอน; never claim prediction.
- Every number must exist in committed evidence (DEVLOG, reports, gates,
  measurements.json). If you cannot cite it, do not write it.
- No negative comparison with Thailand's official flood systems — they are
  the data provider, and the credit is sincere.
- One commit (`docs: ...`), Conventional Commits, builds and passes tests.
- Then write `process/reports/phase-5.md` (short), update `docs/PROGRESS.md`
  and `docs/DEVLOG.md`, and stop. The lead pushes and decides everything else.
