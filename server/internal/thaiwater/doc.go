// Package thaiwater is the HTTP client for the public ThaiWater API
// (api-v3.thaiwater.net), published by the Hydro-Informatics Institute (HII).
//
// It owns the wire format only: typed DTOs that mirror the JSON (Num records
// whether a number came as a string, a number, null or not at all), request
// timeouts,
// retries with backoff and jitter, a named User-Agent, and a concurrency limit
// of at most four for history calls. It does not interpret values; that is
// the job of package normalize. Recorded real responses live in testdata/ so
// unit tests never touch the network.
package thaiwater
