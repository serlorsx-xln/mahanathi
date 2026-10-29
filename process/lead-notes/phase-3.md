## Review of Phase 2 (lead)

The review was not mine alone: five independent reviewer agents went over the phase (humanize, domain/normalize, store/ingest, api/wiring, honesty/process), and every finding was then re-checked by a separate skeptic that tried to refute it. Nineteen agents in total. My own `go test -race -cover ./...` reproduces your coverage table exactly. Verdict: **Phase 2 passes.** The humanize package really is the crown jewel: 100% coverage, a real fuzz target, banned phrases enforced at three layers, trends from the bucketed series only, and every hand-checked sentence in the report recomputed correctly by an independent agent. Nine claimed findings were refuted — mostly deliberate, documented decisions, which is exactly how it should look.

Five confirmed findings. Fix 1–3 as **Step 0 of Phase 3**, each in its own commit with a test where one is possible:

1. **P1 — `suspect` serializes as `null`** (`server/internal/api/now.go`, stationNow): a nil slice marshals to `null`, but `docs/api.md` promises `"suspect": []` and the placeholder page does `s.suspect.length === 0`, which throws on `null`. Initialize to an empty slice. Add a wire-shape assertion to the API test.
2. **P2 — `docs/api.md` reference decoder skips the station ids**: the snippet starts reading the bitmap at offset 20, but the layout places 8×n id bytes there; the test decoder in `api_test.go` reads them correctly. The Phase 4 front end will implement the doc snippet, so the doc must be fixed now.
3. **P3 — 429 responses bypass the security headers**: in `api.go:102` swap the wrap order to `secured(h.limited(api))` so the reject branch also carries the headers (match the `/readyz` wiring), and assert it in the test.
4. **P2 — the Phase 2 report's rate-limit sentence misstates the recorded run**: the log shows 1 warm-up + 40 requests, 200 for the first 38, 429 from the 39th. Correct `process/reports/phase-2.md` in place; reports are part of the permanent record, so they must match what was witnessed.
5. **P3 — DEVLOG entry order contradicts the corrected timestamps**: move the flaky-tests entry so times ascend. Re-check its real time against the commit that introduced it.

One clarification for the record, not a defect: your DEVLOG calls the ~20:32 interruption "a power loss". The lead's own view is that the orchestration session ended while your process was a child of it; the reviewer found OS-level evidence for a power loss too. Both may be true in sequence — leave the DEVLOG entry as is, but if you ever state a single cause, state both accounts, not just one.

### Phase 3 direction

You are on GLM-5.3 now; taste is where weaker models show first, so the gates are stricter, not looser:

- **Read `docs/brief/01-DESIGN-DIRECTION.md` end to end again before anything.** The committed world is "a night-time hydrographic instrument". Every decision in this phase either serves that world or it is cut.
- **The two full gate iterations are mandatory**, and iteration 2 must change things that iteration 1's critique actually identified, not reshuffle.
- **Run `/impeccable critique` on real Playwright screenshots**, then score every checklist item in `GATE-phase-3.md` with the screenshot path as evidence. Zero banned items or the gate fails.
- **The situation colour scale needs the colour-blindness simulation screenshots** and computed contrast numbers, not adjectives.
- **The type specimen decides with rendered Thai**, not font names: the national sentence, three station sentences, "จังหวัดพระนครศรีอยุธยา", tabular figures against each other. Then ADR 0005.
- **The relief look is the demo moment.** Country view must make the room go quiet at 1440×900 and still read correctly at 390×844. Vertical exaggeration goes in ADR 0006 with before/after screenshots.
- Use the Phase 2 API for gauge data in the prototype (save a real `/api/v1/now` + `/api/v1/timeline` snapshot as a fixture under `web/src/design/` or `geodata/` — your call, state it in the ADR), so the look-dev is never fake data.
- Headless rule from `CLAUDE.md` applies doubly here: sub-agents run in the foreground or you wait for them. Nothing ends the phase while a sub-agent is alive.

Stop after the gate passes and the report is written. I must approve the look before Phase 4 opens.
