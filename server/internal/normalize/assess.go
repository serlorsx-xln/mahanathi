package normalize

import (
	"math"
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

// SuspectReason is why a reading is not believable as a current value.
type SuspectReason int

// The reasons, in the order they are reported.
const (
	SuspectBelowGround SuspectReason = iota + 1
	SuspectImplausiblePct
	SuspectJump
	SuspectFuture
)

// String names the reason for logs and the API.
func (r SuspectReason) String() string {
	switch r {
	case SuspectBelowGround:
		return "below-ground"
	case SuspectImplausiblePct:
		return "implausible-percent"
	case SuspectJump:
		return "jump"
	case SuspectFuture:
		return "future-time"
	default:
		return "unknown"
	}
}

// IsStale reports whether a reading taken at `at` is older than StaleAfter
// at `now`. The clock is always passed in, never read.
func IsStale(at, now time.Time) bool { return now.Sub(at) > StaleAfter }

// SuspectReasons checks one reading against the plausibility bounds. prev is
// the hourly bucket just before the reading's own bucket; pass a zero Point
// when it is unknown, and the jump check is skipped.
func SuspectReasons(st domain.Station, r domain.Reading, prev domain.Point, now time.Time) []SuspectReason {
	var out []SuspectReason
	wl := toMM(r.WL)
	if st.HasGround() {
		ground, bank := toMM(st.Ground), toMM(st.BankFull)
		if wl < ground-maxBelowGroundMM {
			out = append(out, SuspectBelowGround)
		}
		if wl-ground > maxPctFactor*(bank-ground) {
			out = append(out, SuspectImplausiblePct)
		}
	}
	if prev.OK {
		if d := wl - toMM(prev.WL); d > maxJumpMM || d < -maxJumpMM {
			out = append(out, SuspectJump)
		}
	}
	if r.At.After(now.Add(futureTolerance)) {
		out = append(out, SuspectFuture)
	}
	return out
}

// HourEnd is the end of the hourly bucket that contains t: buckets are
// (HH−1:00, HH:00], so 10:00 belongs to the 10:00 bucket and 10:10 to 11:00.
func HourEnd(t time.Time) time.Time {
	h := t.Truncate(time.Hour)
	if h.Equal(t) {
		return h
	}
	return h.Add(time.Hour)
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
		i := int(HourEnd(r.At).Sub(first) / time.Hour)
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

func toMM(v float64) int64 { return int64(math.Round(v * 1000)) }
