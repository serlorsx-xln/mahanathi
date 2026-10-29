package normalize

import (
	"time"

	"github.com/serlorsx-xln/mahanathi/server/internal/domain"
)

// StaleAfter is how old a reading may be and still count as "now" (humanize
// spec: older than 3 h is stale).
const StaleAfter = 3 * time.Hour

// futureTolerance absorbs clock skew between ThaiWater's loggers and us.
const futureTolerance = 10 * time.Minute

// Plausibility bounds from the humanize spec, in millimetres so that the
// comparisons are exact: 49.24 − 37.72 is 11.520000000000003 in float64,
// which would make exactly 400 % look like more.
const (
	maxBelowGroundMM = 20_000 // wl more than 20 m under the ground reference
	maxJumpMM        = 3_000  // more than 3 m between consecutive hours
	maxPctFactor     = 4      // more than 400 % of bank-full
)

// IsStale reports whether a reading taken at `at` is older than StaleAfter
// at `now`. The clock is always passed in, never read.
func IsStale(at, now time.Time) bool { return now.Sub(at) > StaleAfter }

// SuspectReasons checks one reading against the plausibility bounds. prev is
// the hourly bucket just before the reading's own bucket; pass a zero Point
// when it is unknown, and the jump check is skipped.
func SuspectReasons(st domain.Station, r domain.Reading, prev domain.Point, now time.Time) []domain.SuspectReason {
	var out []domain.SuspectReason
	wl := domain.MM(r.WL)
	if st.HasGround() {
		ground, bank := domain.MM(st.Ground), domain.MM(st.BankFull)
		if wl < ground-maxBelowGroundMM {
			out = append(out, domain.SuspectBelowGround)
		}
		if wl-ground > maxPctFactor*(bank-ground) {
			out = append(out, domain.SuspectImplausiblePct)
		}
	}
	if prev.OK {
		if d := wl - domain.MM(prev.WL); d > maxJumpMM || d < -maxJumpMM {
			out = append(out, domain.SuspectJump)
		}
	}
	if r.At.After(now.Add(futureTolerance)) {
		out = append(out, domain.SuspectFuture)
	}
	return out
}

// Hourly buckets readings into `hours` consecutive hourly points ending at
// `end` (truncated to the hour). Each point holds the last reading in
// (At − 1 h, At]; hours without one have OK false. This is how 10-minute HII
// series and hourly RID series become one shape (lead's answer to Phase 1
// question 4). Input order does not matter.
func Hourly(readings []domain.Reading, end time.Time, hours int) []domain.Point {
	if hours <= 0 {
		return nil
	}
	end = end.Truncate(time.Hour)
	first := end.Add(-time.Duration(hours-1) * time.Hour)
	pts := make([]domain.Point, hours)
	for i := range pts {
		pts[i].At = first.Add(time.Duration(i) * time.Hour)
	}
	best := make([]time.Time, hours)
	for _, r := range readings {
		i := int(domain.HourEnd(r.At).Sub(first) / time.Hour)
		if i < 0 || i >= hours {
			continue
		}
		if !pts[i].OK || r.At.After(best[i]) {
			pts[i].WL, pts[i].OK = r.WL, true
			best[i] = r.At
		}
	}
	return pts
}
