// Package ingest keeps the store current.
//
// It fetches the national snapshot once at start and then hourly a few
// minutes past the hour, backfills 72 hours of history for stations that lack
// it (at most four requests in flight), and records counts, duration and
// errors for every run. The ThaiWater client is an interface so tests can use
// a fake.
package ingest
