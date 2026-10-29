# Phase 2: Data core (test-driven)

The lead has reviewed Phase 1. Lead notes are appended at the bottom of this prompt; address them first.

**Goal:** a correct, defensible data pipeline from ThaiWater to a JSON API, with the humanize package as the crown jewel. It must be the most carefully tested code in the repo, because the owner will defend it line by line in the interview.

## Order of work (tests first for every pure package)
1. **`thaiwater` client**
   - Typed DTOs, context timeouts, retries with backoff and jitter, the User-Agent, and a concurrency limiter for graph calls.
   - Fixture-based tests with no live network in unit tests.
2. **`normalize`**
   - Turn DTOs into domain `Station` and `Reading` values.
   - Parse ICT into UTC, parse numbers from strings, and apply the bank-full and ground reference from ADR 0003.
   - Flag suspect values and mark stale ones (stale is relative to an injected clock).
   - Use table tests over the fixtures.
3. **`humanize`**
   - Implement exactly the spec in the engineering brief §3, using table tests and golden files.
   - Banned-phrase fuzz across all fixtures.
   - National sentence grammar for counts of 0, 1 and many.
   - Thai number, time and date formatting module.
   - Coverage must be 90% or more.
4. **`store`**
   - SQLite in WAL mode, embedded SQL migrations and idempotent upserts keyed by (station_id, measured_at).
   - Queries for now, history and the timeline, plus 30-day retention.
   - Tests against a temp DB.
5. **`ingest`**
   - Run once at start, then hourly at :07 past.
   - Backfill 72 hours for stations with missing history, with concurrency of 4 or lower.
   - Record run metrics, and test with a fake client.
6. **`api`**
   - The endpoints in brief §2 with compact JSON, ETag, gzip, security headers and a rate limit.
   - The `/api/v1/timeline` encoding must be documented in `docs/api.md` and stay within the size budget. Measure it on real data.
   - Use `httptest` tests.
7. **Wire up `main`:** config, store, ingest and api, with `web/dist` embedded and a placeholder page.
8. **Finish `docs/research/data-quality.md`** with numbers from the full live dataset: stations consistent with ADR 0003, anomalies by type, and the product behaviour for each type.

## Verification (paste the summaries into the report)
- `go test -race -cover ./...` output.
- One live run: start the binary, let the ingest and backfill complete, `curl /api/v1/now | head`, and record the national sentence it produced. Also record the timeline payload size compressed and uncompressed.
- Five real human sentences for five different real stations (an overflow, a high, a normal, a stale and a missing-bank case). Check them against the raw numbers by hand in the report.

## Report and stop
Write `process/reports/phase-2.md`. Record any AI mistakes the tests caught in DEVLOG as they happen. Stop after the report.
