// Package domain is the vocabulary every other package shares: stations,
// readings, hourly points and the five official levels.
//
// It holds the arithmetic of ADR 0003 (percent of bank-full, the signed
// distance to the bank, and the level thresholds) because those three
// formulas are the product's core claims and must exist exactly once. It has
// no dependencies and does no I/O.
package domain

import "time"

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

// LevelForPct applies the thresholds. Everything at or below 10 % (including
// negative percents, where the water is under the ground reference) is
// LevelCriticallyLow.
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

// Pct is the percent of bank-full, (wl − ground)/(bankFull − ground) × 100,
// unrounded and possibly negative or above 100.
func (s Station) Pct(wl float64) (float64, bool) {
	if s.Refs != RefsComplete {
		return 0, false
	}
	return (wl - s.Ground) / (s.BankFull - s.Ground) * 100, true
}

// Level is the level for a water level, or LevelUnknown without a percent.
func (s Station) Level(wl float64) Level {
	pct, ok := s.Pct(wl)
	if !ok {
		return LevelUnknown
	}
	return LevelForPct(pct)
}

// Distance is the signed distance from bank-full in metres: positive means
// the water is above the bank. We compute it ourselves because ThaiWater's
// diff_wl_bank is unsigned.
func (s Station) Distance(wl float64) (float64, bool) {
	if !s.HasBankFull() {
		return 0, false
	}
	return wl - s.BankFull, true
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
