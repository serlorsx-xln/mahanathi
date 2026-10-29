# Phase 6: Ship docs and handover

Lead notes from the Phase 5 review are at the bottom; address them first.

**Goal:** a repo a judge can understand in 2 minutes and the owner can defend for 8.

## Deliverables
1. **Container**
   - Verify the Dockerfile with `docker build` and `docker run -p 8080:8080 -v mahanathi-data:/data`.
   - The health check passes and data appears within a few minutes.
   - Document the env vars in `.env.example`.
2. **`README.md`** (Thai first, with a short English section)
   - What it is in one sentence, plus a hero screenshot from a real render.
   - **Live URL**: leave a placeholder for the lead to fill in.
   - Why it exists, explained with the real data story.
   - How it works: one diagram (ThaiWater → ingest → SQLite → API → R3F) and the humanize rules.
   - Decisions, linking to the ADRs.
   - Data and licences.
   - Run locally.
   - **How this was built:** the owner directed it, a lead AI wrote the briefs and reviewed each phase, and an implementing AI built it in phases. Link to `docs/brief/`, `process/prompts/` and `process/reports/`. State this plainly and without apology.
3. **`docs/ARCHITECTURE.md`**: components, data flow, the timeline encoding, failure modes and their handling, and performance notes.
4. **`docs/REVIEW.md`** for the owner
   - The exact files and functions to read line by line before the interview: humanize, normalize (bank-full reference, stale and suspect), timeline encoding, rate limiter and nearest-station.
   - For each one: what it does, why it is written that way, and 3 questions a judge might ask with crisp answers.
   - Then the real AI mistakes from the DEVLOG, each with the test that caught it.
5. **`docs/DEMO.md`**
   - A 60-second demo script for the interview.
   - A 20-second "hand the judge the phone" flow.
   - Backup plans: no network (cached snapshot mode), no geolocation (search), and a 30-second screen recording to capture.

## Report and stop
Write `process/reports/phase-6.md` with a final checklist against `00-PRODUCT.md` success criteria. Stop.
