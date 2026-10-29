// Package domain is the vocabulary every other package shares: stations,
// readings, hourly points and the five official levels.
//
// It holds the arithmetic of ADR 0003 (percent of bank-full, the signed
// distance to the bank, and the level thresholds) because those three
// formulas are the product's core claims and must exist exactly once. It has
// no dependencies and does no I/O.
//
// The claims are computed in whole millimetres, the finest unit ThaiWater
// publishes, so that they are exact. In float64 an exact 70 % can come out
// as 70.000000000000014 and cross the "> 70" threshold, and 104.375 % can
// come out as 104.37499999999999 and round the wrong way.
package domain

import (
	"math"
	"time"
)

// MM converts metres to the nearest whole millimetre.
func MM(m float64) int64 { return int64(math.Round(m * 1000)) }

// Level is ThaiWater's five-step situation scale, computed by us from the
// percent of bank-full. LevelUnknown means the station's references do not
// allow a percent.
type Level int

// The levels, in the order of scale.data.rule_web.
const (
	LevelUnknown       Level = 0
	LevelCriticallyLow Level = 1 // น้ำน้อยวิกฤติ, ≤ 10 %
	LevelLow           Level = 2 // น้ำน้อย, > 10 %
	LevelNormal        Level = 3 // น้ำปกติ, > 30 %
	LevelHigh          Level = 4 // น้ำมาก, > 70 %
	LevelOverflow      Level = 5 // น้ำล้นตลิ่ง, > 100 %
)

// Threshold is one step of the scale: a percent strictly above Above earns
// Level.
type Threshold struct {
	Above float64
	Level Level
}

// Thresholds are the known values of ThaiWater's scale.data.rule_web, highest
// first. Ingest compares the live rule_web against these on every poll and
// warns loudly if ThaiWater changes them.
var Thresholds = []Threshold{
	{100, LevelOverflow},
	{70, LevelHigh},
	{30, LevelNormal},
	{10, LevelLow},
}

// LevelForPct applies the thresholds to a percent that is already known,
// such as ThaiWater's published storage_percent. Everything at or below 10 %
// (including negative percents, where the water is under the ground
// reference) is LevelCriticallyLow. Station.Level does the same comparison
// exactly, without forming the percent.
func LevelForPct(pct float64) Level {
	for _, t := range Thresholds {
		if pct > t.Above {
			return t.Level
		}
	}
	return LevelCriticallyLow
}

// Label is the Thai label ThaiWater uses for the level (scale.data.scale).
func (l Level) Label() string {
	switch l {
	case LevelCriticallyLow:
		return "น้ำน้อยวิกฤติ"
	case LevelLow:
		return "น้ำน้อย"
	case LevelNormal:
		return "น้ำปกติ"
	case LevelHigh:
		return "น้ำมาก"
	case LevelOverflow:
		return "น้ำล้นตลิ่ง"
	default:
		return ""
	}
}

// Refs says which of a station's two references can be used (ADR 0003).
type Refs int

// The reference states. Only RefsComplete allows a percent and a level;
// RefsComplete and RefsBankOnly allow a distance to the bank.
const (
	// RefsMissing: min_bank is absent or unparsable. No claim relative to
	// the bank is possible.
	RefsMissing Refs = iota
	// RefsComplete: bank-full and ground are both usable and ground is below
	// bank-full.
	RefsComplete
	// RefsBankOnly: ground is absent or unparsable but bank-full is valid
	// and not zero-coded. The distance is stated; percent and level are not
	// (lead's answer to Phase 1 question 2).
	RefsBankOnly
	// RefsZeroCoded: both references are 0, which EGAT uses at dams to mean
	// "none". Treated as absent, so the dam is never called overflowing.
	RefsZeroCoded
	// RefsInverted: ground is at or above bank-full. One of the two is wrong
	// and we cannot tell which, so neither is used.
	RefsInverted
)

// String names the state for logs and the data-quality report.
func (r Refs) String() string {
	switch r {
	case RefsComplete:
		return "complete"
	case RefsBankOnly:
		return "bank-only"
	case RefsZeroCoded:
		return "zero-coded"
	case RefsInverted:
		return "inverted"
	default:
		return "missing"
	}
}

// Station is a telemetry water-level station with its references resolved.
type Station struct {
	ID           int64
	Code         string // tele_station_oldcode, such as "C.35"
	Name         string // Thai name
	Lat, Lon     float64
	ProvinceCode string
	Province     string
	Amphoe       string
	Tumbon       string
	Basin        string
	River        string // empty when ThaiWater gives none
	Agency       string // short English name: HII, RID, EGAT, …

	Refs     Refs
	BankFull float64 // metres MSL; meaningful when Refs is RefsComplete or RefsBankOnly
	Ground   float64 // metres MSL; meaningful when Refs is RefsComplete
}

// HasBankFull reports whether a distance to the bank can be stated.
func (s Station) HasBankFull() bool { return s.Refs == RefsComplete || s.Refs == RefsBankOnly }

// HasGround reports whether the ground reference can be used.
func (s Station) HasGround() bool { return s.Refs == RefsComplete }

// Fill is how far up the bank the water stands, as an exact fraction in
// whole millimetres: the percent of bank-full is 100 × above / height, where
// above is wl − ground and height is bankFull − ground (always positive).
func (s Station) Fill(wl float64) (above, height int64, ok bool) {
	if s.Refs != RefsComplete {
		return 0, 0, false
	}
	g := MM(s.Ground)
	above, height = MM(wl)-g, MM(s.BankFull)-g
	if height <= 0 {
		return 0, 0, false
	}
	return above, height, true
}

// Pct is the percent of bank-full, (wl − ground)/(bankFull − ground) × 100,
// unrounded and possibly negative or above 100. It is for display and
// sorting; levels and served percents come from Level and PctHundredths.
func (s Station) Pct(wl float64) (float64, bool) {
	above, height, ok := s.Fill(wl)
	if !ok {
		return 0, false
	}
	return float64(above) / float64(height) * 100, true
}

// PctHundredths is the percent of bank-full in hundredths of a percent,
// rounded half away from zero: 10438 is 104.38 %. It is what the API serves
// and what humanize prints, so the two agree to the digit.
func (s Station) PctHundredths(wl float64) (int64, bool) {
	above, height, ok := s.Fill(wl)
	if !ok {
		return 0, false
	}
	return divRound(above*10000, height), true
}

// Level is the level for a water level, or LevelUnknown without a percent.
// "pct > t" is tested as 100 × above > t × height, which is exact for the
// whole-number thresholds of rule_web (ingest stops trusting them if
// ThaiWater changes them).
func (s Station) Level(wl float64) Level {
	above, height, ok := s.Fill(wl)
	if !ok {
		return LevelUnknown
	}
	for _, t := range Thresholds {
		if float64(above*100) > t.Above*float64(height) {
			return t.Level
		}
	}
	return LevelCriticallyLow
}

// Distance is the signed distance from bank-full in millimetres: positive
// means the water is above the bank. We compute it ourselves because
// ThaiWater's diff_wl_bank is unsigned.
func (s Station) Distance(wl float64) (int64, bool) {
	if !s.HasBankFull() {
		return 0, false
	}
	return MM(wl) - MM(s.BankFull), true
}

// divRound is n / d rounded half away from zero, for d > 0.
func divRound(n, d int64) int64 {
	if n < 0 {
		return -((-2*n + d) / (2 * d))
	}
	return (2*n + d) / (2 * d)
}

// Reading is one measured water level.
type Reading struct {
	StationID int64
	At        time.Time // UTC instant
	WL        float64   // metres above mean sea level
}

// Point is one hourly bucket: the last reading in (At − 1 h, At]. OK is false
// when the hour has no reading.
type Point struct {
	At time.Time
	WL float64
	OK bool
}

// HourEnd is the end of the hourly bucket that contains t: buckets are
// (HH−1:00, HH:00], so 10:00 belongs to the 10:00 bucket and 10:10 to 11:00.
// Every hourly series in the product (history, timeline, the 6 h trend) uses
// this one rule (lead's answer to Phase 1 question 4).
func HourEnd(t time.Time) time.Time {
	h := t.Truncate(time.Hour)
	if h.Equal(t) {
		return h
	}
	return h.Add(time.Hour)
}

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
