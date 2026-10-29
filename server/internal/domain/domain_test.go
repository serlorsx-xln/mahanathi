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
	if d, ok := gt11.Distance(41.93); !ok || math.Abs(d-1.33) > 1e-9 {
		t.Errorf("Gt.11 distance = %v %v, want +1.33", d, ok)
	}
	// At 17:54 the same station was below the bank, which the unsigned
	// diff_wl_bank ("0.44") hides; our distance keeps the sign.
	if d, _ := gt11.Distance(40.16); math.Abs(d+0.44) > 1e-9 {
		t.Errorf("Gt.11 distance at 17:54 = %v, want -0.44", d)
	}
	if gt11.Level(40.16) != LevelHigh || gt11.Level(41.93) != LevelOverflow {
		t.Error("Gt.11 levels wrong")
	}

	bankOnly := Station{Refs: RefsBankOnly, BankFull: 10.9}
	if _, ok := bankOnly.Pct(11.01); ok {
		t.Error("bank-only station produced a percent")
	}
	if d, ok := bankOnly.Distance(11.01); !ok || math.Abs(d-0.11) > 1e-9 {
		t.Errorf("bank-only distance = %v %v, want +0.11", d, ok)
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
