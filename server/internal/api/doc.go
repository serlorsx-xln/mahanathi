// Package api serves the versioned JSON API (/api/v1/...) and the health
// probes on net/http with Go 1.22+ routing patterns.
//
// It is responsible for strict input validation, gzip, ETag and
// Cache-Control, security headers, and a per-IP token-bucket rate limit that
// answers 429 with Retry-After. It reads from the store and never calls
// ThaiWater directly.
package api
