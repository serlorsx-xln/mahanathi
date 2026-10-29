package domain

import (
	"math"
	"testing"
	"time"
)

func TestLevelForPctMatchesRuleWeb(t *testing.T) {
	cases := []struct {
		pct  float64
		want Level
	}{
		{-18.93, LevelCriticallyLow}, // URTD03, water under the ground reference
		{0, LevelCriticallyLow},
		{10, LevelCriticallyLow}, // "<=10" is level 1
		{10.0001, LevelLow},
		{30, LevelLow},
		{30.01, LevelNormal},
		{69.994, LevelNormal}, // URTU09 with the 2-decimal min_bank
		{70, LevelNormal},
		{70.04, LevelHigh}, // URTU09 as ThaiWater publishes it
		{100, LevelHigh},   // exactly at the bank is not overflow
		{100.0001, LevelOverflow},
		{146.18, LevelOverflow},
	}
	for _, c := range cases {
		if got := LevelForPct(c.pct); got != c.want {
			t.Errorf("LevelForPct(%v) = %d, want %d", c.pct, got, c.want)
		}
	}
}

func TestLabels(t *testing.T) {
	want := map[Level]string{
		LevelUnknown:       "",
		LevelCriticallyLow: "น้ำน้อยวิกฤติ",
		LevelLow:           "น้ำน้อย",
		LevelNormal:        "น้ำปกติ",
		LevelHigh:          "น้ำมาก",
		LevelOverflow:      "น้ำล้นตลิ่ง",
	}
	for l, w := range want {
		if l.Label() != w {
			t.Errorf("Level(%d).Label() = %q, want %q", l, l.Label(), w)
		}
	}
}

func TestRefsString(t *testing.T) {
	for r, w := range map[Refs]string{RefsMissing: "missing", RefsComplete: "complete", RefsBankOnly: "bank-only", RefsZeroCoded: "zero-coded", RefsInverted: "inverted"} {
		if r.String() != w {
			t.Errorf("%d.String() = %q, want %q", r, r.String(), w)
		}
	}
}

func TestStationArithmetic(t *testing.T) {
	// Gt.11 at 14:00 in the brief: min_bank 40.60, ground 37.72, wl 41.93.
	gt11 := Station{Refs: RefsComplete, BankFull: 40.60, Ground: 37.72}
	pct, ok := gt11.Pct(41.93)
	if !ok || math.Abs(pct-146.18) > 0.005 {
		t.Errorf("Gt.11 pct = %v %v, want 146.18", pct, ok)
	}
	if d, ok := gt11.Distance(41.93); !ok || d != 1330 {
		t.Errorf("Gt.11 distance = %v %v, want +1330 mm", d, ok)
	}
	// At 17:54 the same station was below the bank, which the unsigned
	// diff_wl_bank ("0.44") hides; our distance keeps the sign.
	if d, _ := gt11.Distance(40.16); d != -440 {
		t.Errorf("Gt.11 distance at 17:54 = %v, want -440 mm", d)
	}
	if gt11.Level(40.16) != LevelHigh || gt11.Level(41.93) != LevelOverflow {
		t.Error("Gt.11 levels wrong")
	}

	bankOnly := Station{Refs: RefsBankOnly, BankFull: 10.9}
	if _, ok := bankOnly.Pct(11.01); ok {
		t.Error("bank-only station produced a percent")
	}
	if d, ok := bankOnly.Distance(11.01); !ok || d != 110 {
		t.Errorf("bank-only distance = %v %v, want +110 mm", d, ok)
	}
	if bankOnly.Level(11.01) != LevelUnknown || bankOnly.HasGround() || !bankOnly.HasBankFull() {
		t.Error("bank-only station claims a level or a ground")
	}

	for _, r := range []Refs{RefsMissing, RefsZeroCoded, RefsInverted} {
		s := Station{Refs: r, BankFull: 0, Ground: 0}
		if _, ok := s.Pct(331.73); ok {
			t.Errorf("%v: percent produced", r)
		}
		if _, ok := s.Distance(331.73); ok {
			t.Errorf("%v: distance produced", r)
		}
		if s.Level(331.73) != LevelUnknown {
			t.Errorf("%v: level produced", r)
		}
		if _, ok := s.PctHundredths(331.73); ok {
			t.Errorf("%v: rounded percent produced", r)
		}
	}
}

// A reading exactly on a threshold takes the rule's side of it. In float64
// (3.124 + 1.356) / (5.044 + 1.356) × 100 is 70.000000000000014, which would
// make an exact 70 % look like "> 70" and give level 4 instead of 3.
func TestLevelsAreExactAtTheThresholds(t *testing.T) {
	cases := []struct {
		ground, bank, wl float64
		want             Level
	}{
		{-1.356, 5.044, 3.124, LevelNormal}, // exactly 70 %
		{-1.356, 5.044, 3.125, LevelHigh},   // 1 mm more
		{-3, -2, -2.3, LevelNormal},         // exactly 70 % below sea level
		{2, 10, 4.4, LevelLow},              // exactly 30 %
		{2, 10, 2.8, LevelCriticallyLow},    // exactly 10 %
		{2, 10, 10, LevelHigh},              // exactly at the bank
		{2, 10, 10.001, LevelOverflow},      // 1 mm over it
		{37.72, 40.599998, 49.24, LevelOverflow},
	}
	for _, c := range cases {
		s := Station{Refs: RefsComplete, Ground: c.ground, BankFull: c.bank}
		if got := s.Level(c.wl); got != c.want {
			t.Errorf("ground %v bank %v wl %v: level %d, want %d", c.ground, c.bank, c.wl, got, c.want)
		}
	}
	// A hand-built station with no bank height has no level, rather than a
	// division by zero.
	if (Station{Refs: RefsComplete, Ground: 5, BankFull: 5}).Level(6) != LevelUnknown {
		t.Error("zero bank height produced a level")
	}
}

func TestPctHundredthsRoundsHalfAwayFromZero(t *testing.T) {
	st := Station{Refs: RefsComplete, Ground: 2, BankFull: 10}
	cases := []struct {
		wl   float64
		want int64
	}{
		{10.35, 10438},   // 104.375 %, which float64 makes 104.37499999999999
		{8.6, 8250},      // 82.50 %
		{2, 0},           // at the ground reference
		{1.9999, 0},      // rounds to the same millimetre
		{1.999, -1},      // −0.0125 % is −0.01 %
		{1.9996, 0},      // 2.000 after rounding to mm
		{0.5, -1875},     // −18.75 %
		{10, 10000},      // at the bank
		{34.0004, 40000}, // 400 %
	}
	for _, c := range cases {
		got, ok := st.PctHundredths(c.wl)
		if !ok || got != c.want {
			t.Errorf("PctHundredths(%v) = %d %v, want %d", c.wl, got, ok, c.want)
		}
	}
	// URTU09 at 17:00 with its full-precision references: 2.496 / 3.562 is
	// 70.07 %. ThaiWater publishes 70.04 %; both are level 4.
	urtu09 := Station{Refs: RefsComplete, Ground: 221.824, BankFull: 225.386}
	if got, _ := urtu09.PctHundredths(224.32); got != 7007 || urtu09.Level(224.32) != LevelHigh {
		t.Errorf("URTU09 = %d, level %d", got, urtu09.Level(224.32))
	}
}

func TestMM(t *testing.T) {
	for in, want := range map[float64]int64{1.395: 1395, 40.599998: 40600, -0.44: -440, 0: 0, -1.0005: -1001} {
		if got := MM(in); got != want {
			t.Errorf("MM(%v) = %d, want %d", in, got, want)
		}
	}
}

func TestSuspectReasonNames(t *testing.T) {
	for r, w := range map[SuspectReason]string{SuspectBelowGround: "below-ground", SuspectImplausiblePct: "implausible-percent", SuspectJump: "jump", SuspectFuture: "future-time", 0: "unknown"} {
		if r.String() != w {
			t.Errorf("SuspectReason(%d).String() = %q, want %q", r, r.String(), w)
		}
	}
}

func TestHourEnd(t *testing.T) {
	at := func(h, m int) time.Time { return time.Date(2026, 10, 29, h, m, 0, 0, time.UTC) }
	cases := []struct{ in, want time.Time }{
		{at(10, 0), at(10, 0)},
		{at(10, 10), at(11, 0)},
		{at(10, 59), at(11, 0)},
		{at(23, 30), time.Date(2026, 10, 30, 0, 0, 0, 0, time.UTC)},
		{at(10, 0).Add(time.Second), at(11, 0)},
	}
	for _, c := range cases {
		if got := HourEnd(c.in); !got.Equal(c.want) {
			t.Errorf("HourEnd(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}
