# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

- **Primary (from the owner's brief):** a person in a flood-prone area (Ayutthaya, Pathum Thani, Nonthaburi, the southern provinces) checking at night on a phone, with rain outside, worried about the river or canal behind their house. They need a one-glance answer, trustworthy numbers and the time of measurement.
- **Secondary:** curious people and journalists wanting the national picture: where water is high now and how it changed over 72 hours.
- **Demo audience:** YWC21 judges, web programmers and designers, on a laptop and on their own phone for 20 seconds; experts at spotting template or AI-generic work.

## Product Purpose

Every river gauge in Thailand, live, on a 3D relief of the country. Tap **ใกล้ฉัน** and get one plain Thai sentence about the river nearest you. Success: a judge scans a QR code, taps ใกล้ฉัน and reads a correct sentence about a real river within 20 seconds.

## Positioning

ThaiWater's data is excellent but presented for hydrologists (tables, ม.รทก., station codes, five garish colours). This product answers the only question an ordinary person has: *what is the river near me doing right now?* — in plain Thai, with measured values and their trend, never predictions.

## Operating Context

- Flood season in Thailand; data publishes hourly from ~800 stations via ThaiWater (National Hydroinformatics Data Center).
- Used at night, one-handed, on a phone, possibly in the LINE in-app browser.
- Demoed in a bright interview room on a laptop.

## Capabilities and Constraints

- Core jobs (from the brief): see the country now (J1, < 2 s first frame), ใกล้ฉัน (J2, geolocation on-device), tolerant Thai search (J3), trust — measured time and source always shown (J4), a 72-hour scrubbing timeline (J5), works on phone/desktop/LINE/no-WebGL/reduced-motion (J6).
- **Explicit non-goals:** no prediction, no warnings, no safety claims (banned phrases are test-enforced); no accounts or personal data (location never leaves the device); no AI inside the product; no imitation of a government site.
- The server is a single Go binary with SQLite (Phases 1–2, done and live-verified); the API serves `/api/v1/now`, `/history`, `/timeline` (binary MNTL), `/readyz`.

## Brand Commitments

- Name: **มหานที** (Mahanathi), from the YWC21 theme; the theme is never preached in the UI.
- Visual world, committed in `docs/brief/01-DESIGN-DIRECTION.md`: a night-time hydrographic instrument — the national survey's plaster relief in a dark map room. Dark theme; cartographic truth; restraint.
- Copy voice: a calm knowledgeable friend. Banned: ปลอดภัย, จะท่วม, ไม่ท่วม, marketing words.
- Data source credited: ThaiWater / HII. Derived numbers only, clearly attributed.

## Evidence on Hand

- A live, tested data pipeline: 804 stations ingested and backfilled, oracle 789/789 levels in agreement (Phase 2 report, `process/reports/phase-2.md`).
- Recorded fixtures under `server/internal/thaiwater/testdata/` and research in `docs/research/` (data-quality, thaiwater-api, geodata-sources, typography).
- Real sentences from the humanize package (100 % coverage) — e.g. the national sentence and per-station sentences used in look-dev.
- No testimonials, invented numbers or marketing claims exist, and none may be created.

## Product Principles

1. **One glance, one sentence.** The answer to "what is the river near me doing" must be readable without training.
2. **Honest numbers.** Measured time and source on everything; stale and suspect data labelled, never passed off as current; no prediction, ever.
3. **The relief is the page.** Geography and data are the visual; chrome recedes.
4. **Works where the user is.** A phone at night in the rain is the primary device, not a special case.

*(Inferred by the implementing session from the owner's committed briefs: `docs/brief/00-PRODUCT.md` and `01-DESIGN-DIRECTION.md`; no interview mechanism was available this session. Every fact above traces to those briefs.)*

## Accessibility & Inclusion

- WCAG AA (4.5:1 body, 3:1 large text) by computation, evidenced in the look-dev gate.
- The situation scale must pass deuteranopia and protanopia simulation.
- `prefers-reduced-motion` honoured: cuts, not flights.
