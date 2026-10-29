# ADR 0002: SQLite (modernc, pure Go) in one file on the `/data` volume

- **Status:** accepted (Phase 1)
- **Date:** 2026-10-29
- **Deciders:** implementer, for review by the lead

## Context

The server polls ThaiWater every hour, keeps 72 hours of history for the product and 30 days in total, and must keep serving the last snapshot when ThaiWater is down. The product owner asked for an answer to "why SQLite" in the docs.

**Volume.** The snapshot has 804 stations. HII stations report every 10 minutes and most others hourly (`docs/research/thaiwater-api.md`). Storing readings as received, with the 10-minute series from backfill, gives a worst case of 804 × 144 × 30 ≈ 3.5 M rows in 30 days.

**Measured on this machine** (Python's `sqlite3`, SQLite 3.51.0, on the project's exFAT volume, in a throwaway directory):
- **Size.** A `WITHOUT ROWID` table keyed `(station_id, measured_at)` with one `REAL` column held 30 days of hourly rows for 804 stations (578,880 rows) in 13.1 MB. That is about 23 bytes a row, so the 3.5 M-row worst case is roughly 80 MB.
- **Concurrency.** `PRAGMA journal_mode=WAL` was accepted. A writer committed 70 batches of 804 upserts in 8 s while a reader thread and a separate reader process queried throughout, with no errors and with counts that only went up.

**Access pattern:**
- **Writes:** one batch an hour, plus a one-off backfill of about 800 stations.
- **Reads:** from the API, but `/now` and `/timeline` change only once an hour and can be served from memory.

**Constraints:**
- No secrets.
- One container on Coolify.
- A static binary with `CGO_ENABLED=0` on a distroless image (ADR 0001).

## Options

1. **SQLite through `modernc.org/sqlite`**, a pure-Go translation of SQLite. The latest version on the module proxy on 2026-10-29 was v1.61.0. This is the lead's recommendation.
2. **SQLite through `mattn/go-sqlite3`.** It is faster and long established, but it needs CGO.
3. **PostgreSQL**, either as a second Coolify service or a managed instance.
4. **An embedded key-value store** (bbolt or Pebble).
5. **No database:** an in-memory ring buffer plus a JSON snapshot file on `/data`.
6. **DuckDB**, which is columnar and suited to analytics.

| | Ops cost | Static binary | Query power | Survives restart | Fit for the volume |
|---|---|---|---|---|---|
| 1 modernc SQLite | none (one file) | yes | SQL | yes | ample |
| 2 mattn SQLite | none | no (C toolchain, libc) | SQL | yes | ample |
| 3 PostgreSQL | a second service, credentials, backups | yes | SQL | yes | far beyond need |
| 4 bbolt / Pebble | none | yes | hand-written scans | yes | ample |
| 5 memory + JSON | none | yes | hand-written | partly, and atomic writes are ours to get right | ample |
| 6 DuckDB | none | no (CGO) | SQL, columnar | yes | far beyond need |

- **Option 2** would rule out `distroless/static` and complicate cross-builds, for a speed-up this volume does not need.
- **Option 3** brings the first secret and a second thing to operate, back up and explain.
- **Options 4 and 5** move the bucketing, retention and counting queries into hand-written Go, which is more code for the owner to review, not less.
- **Option 5** also loses the "open the file and look" property that helps the owner check the data.
- **Option 6** is a large CGO dependency for a workload that is mostly small point writes.

## Decision

**SQLite through `modernc.org/sqlite`**, in one file on the `/data` volume.

**Location.** The file is `/data/mahanathi.db`. Phase 2 makes the path configurable through an environment variable and documents it in `.env.example`.

**Connection settings:**
- `journal_mode=WAL`, `synchronous=NORMAL`, `foreign_keys=ON` and a `busy_timeout`.
- One writer connection, used by ingest, and a small reader pool, used by the API.

**Migrations.** Migrations are plain SQL files embedded in the binary and applied in order at start-up, each in a transaction. The version is tracked in `PRAGMA user_version`.

**Readings:**
- Readings are stored as received, upserted on `(station_id, measured_at)`, where `measured_at` is a UTC instant.
- Upserts are idempotent, so re-polling or re-running a backfill is harmless.
- Hourly bucketing is applied when the timeline is built, not on write. This keeps the raw data and lets the bucketing rule change without a migration.

**Retention.** Readings older than 30 days are deleted after each ingest run.

**Serving.** After each ingest, the `/now` and `/timeline` payloads are rebuilt and held in memory with their ETags, so normal traffic does not touch the database. Station history reads from SQLite.

## Consequences

**Positive:**
- There is no second service and no credentials.
- A backup is one file (`VACUUM INTO` or the `.backup` command).
- The owner can open the database with any SQLite tool.
- The binary stays static and the image stays distroless.
- What the product shows can be rebuilt. If `/data` is lost, a 72-hour backfill restores it: about 800 graph calls at concurrency 4 and about 0.2 s each, so on the order of a minute plus our politeness delays. Only history older than the graph endpoint returns would be lost; we have tested 4 calendar days and not longer ranges.

**Negative:**
- **Speed.** modernc is slower than the CGO driver. At one batch an hour this does not matter, and Phase 2 logs ingest duration on every run so we can see it.
- **Binary size.** modernc makes the binary larger. Phase 2 records the image size after adding it, against the 14.1 MB baseline.
- **One writer, one node.** Two replicas must never share the volume. The deploy notes must tell the lead to run exactly one replica on Coolify.
- **Volume ownership.** The runtime user is distroless `nonroot` (uid 65532). A fresh named volume takes the ownership of the mount point in the image. The Dockerfile must therefore ship an empty `/data` owned by 65532: distroless has no shell, so this means `COPY --chown` from a builder stage. Otherwise the first write fails. This is recorded for Phase 6, which adds the volume and the HEALTHCHECK.
- **The exFAT result is indirect.** It was measured with Python's SQLite, not with modernc. The Phase 2 store tests run modernc on this volume and are the real check. If WAL misbehaves there, the reversible fallback is to point the database path at a temporary directory for tests and local runs.
