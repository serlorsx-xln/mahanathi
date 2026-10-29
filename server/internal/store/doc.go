// Package store persists stations and hourly readings in SQLite
// (modernc.org/sqlite, pure Go, WAL mode).
//
// It owns the schema and its embedded migrations, idempotent upserts keyed by
// (station_id, measured_at), the queries behind the API (now, history,
// timeline) and the 30-day retention sweep. See ADR 0002 for why SQLite.
package store
