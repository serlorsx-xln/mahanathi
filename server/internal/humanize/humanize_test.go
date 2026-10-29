package humanize

import (
	"math"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/serlorsx-xln/mahanathi/server/internal/domain"
)

// The clock for the table: 14:30 in Bangkok, with the reading at 14:00.
var (
	now    = time.Date(2026, 10, 29, 7, 30, 0, 0, time.UTC)
	at1400 = time.Date(2026, 10, 29, 7, 0, 0, 0, time.UTC)
	at0900 = time.Date(2026, 10, 29, 2, 0, 0, 0, time.UTC)  // 09:00 ICT, 5.5 h old
	at2000 = time.Date(2026, 10, 28, 13, 0, 0, 0, time.UTC) // 20:00 ICT the day before
	at1410 = time.Date(2026, 10, 29, 7, 10, 0, 0, time.UTC) // a 10-minute station
	nan    = math.NaN()                                     // an hour with no reading in hist()
	ground = 2.0                                            // the test station's references:
	bank   = 10.0                                           // 8 m of bank height
)

func station(refs domain.Refs) domain.Station {
	return domain.Station{ID: 1, Code: "T.1", Province: "ชุมพร", Refs: refs, BankFull: bank, Ground: ground}
}

// at is a current reading of wl metres at 14:00.
func at(st domain.Station, wl float64) Input {
	return Input{Station: st, Latest: &domain.Reading{StationID: st.ID, At: at1400, WL: wl}, Now: now}
}

// hist sets the hourly buckets for the six hours before the latest reading's
// own hour, oldest first; NaN is an hour with no reading.
func hist(in Input, wls ...float64) Input {
	h := domain.HourEnd(in.Latest.At)
	in.Hourly = nil
	for i, wl := range wls {
		p := domain.Point{At: h.Add(time.Duration(i-6) * time.Hour)}
		if !math.IsNaN(wl) {
			p.WL, p.OK = wl, true
		}
		in.Hourly = append(in.Hourly, p)
	}
	return in
}

// stale moves the reading to `when` and marks it stale, as normalize would.
func stale(in Input, when time.Time) Input {
	r := *in.Latest
	r.At = when
	in.Latest, in.Stale = &r, true
	return in
}

func suspect(in Input, rs ...domain.SuspectReason) Input {
	in.Suspect = rs
	return in
}

func moved(in Input, when time.Time) Input {
	r := *in.Latest
	r.At = when
	in.Latest = &r
	return in
}

func TestDescribe(t *testing.T) {
	complete := station(domain.RefsComplete)
	bankOnly := domain.Station{ID: 2, Refs: domain.RefsBankOnly, BankFull: 10.9}
	missing := domain.Station{ID: 3, Refs: domain.RefsMissing}
	zero := domain.Station{ID: 4, Refs: domain.RefsZeroCoded}
	inverted := domain.Station{ID: 5, Refs: domain.RefsInverted, BankFull: 5, Ground: 6}

	// Everything rising from 8.48 to 8.60 over six hours.
	rising := hist(at(complete, 8.6), 8.48, 8.5, 8.52, 8.55, 8.57, 8.59)

	cases := []struct {
		name  string
		in    Input
		want  string
		level domain.Level
	}{
		// Distance to bank-full.
		{"below", at(complete, 8.6), "ต่ำกว่าตลิ่ง 1.40 เมตร", domain.LevelHigh},
		{"below-low", at(complete, 3), "ต่ำกว่าตลิ่ง 7.00 เมตร", domain.LevelLow},
		{"below-ground-reference", at(complete, 1.5), "ต่ำกว่าตลิ่ง 8.50 เมตร", domain.LevelCriticallyLow},
		{"below-normal", at(complete, 6), "ต่ำกว่าตลิ่ง 4.00 เมตร", domain.LevelNormal},
		{"half-cm", at(complete, 8.605), "ต่ำกว่าตลิ่ง 1.40 เมตร", domain.LevelHigh}, // 1.395 m, not 1.39
		{"above", at(complete, 10.35), "สูงกว่าตลิ่ง 0.35 เมตร (น้ำล้นตลิ่ง)", domain.LevelOverflow},
		{"above-half-cm", at(complete, 12.345), "สูงกว่าตลิ่ง 2.35 เมตร (น้ำล้นตลิ่ง)", domain.LevelOverflow},
		{"above-thousands", at(domain.Station{Refs: domain.RefsComplete, BankFull: 2, Ground: 1}, 1236.5), "สูงกว่าตลิ่ง 1,234.50 เมตร (น้ำล้นตลิ่ง)", domain.LevelOverflow},
		{"at-bank", at(complete, 10), "ระดับน้ำเท่ากับตลิ่ง", domain.LevelHigh},
		{"tiny-above", at(complete, 10.004), "สูงกว่าตลิ่งไม่ถึง 1 ซม. (น้ำล้นตลิ่ง)", domain.LevelOverflow},
		{"tiny-below", at(complete, 9.996), "ต่ำกว่าตลิ่งไม่ถึง 1 ซม.", domain.LevelHigh},
		{"tiny-above-5mm", at(complete, 10.005), "สูงกว่าตลิ่ง 0.01 เมตร (น้ำล้นตลิ่ง)", domain.LevelOverflow},

		// Ground absent: the distance, but no level and so no level label
		// (lead's answer 2). Gt.9 บ้านกลาง: wl 11.01, min_bank 10.90.
		{"bank-only-above", at(bankOnly, 11.01), "สูงกว่าตลิ่ง 0.11 เมตร", domain.LevelUnknown},
		{"bank-only-below", at(bankOnly, 10), "ต่ำกว่าตลิ่ง 0.90 เมตร", domain.LevelUnknown},
		{"bank-only-at-bank", at(bankOnly, 10.9), "ระดับน้ำเท่ากับตลิ่ง", domain.LevelUnknown},

		// No usable bank-full: no distance claim at all. BLGTU06 ฝายละแอ
		// (zero-coded) would otherwise read as 331.73 m over the bank.
		{"no-bank-missing", at(missing, 5.12), "สถานีนี้ยังไม่มีข้อมูลระดับตลิ่ง ระดับน้ำ 5.12 ม.รทก.", domain.LevelUnknown},
		{"no-bank-zero-coded", at(zero, 331.73), "สถานีนี้ยังไม่มีข้อมูลระดับตลิ่ง ระดับน้ำ 331.73 ม.รทก.", domain.LevelUnknown},
		{"no-bank-inverted", at(inverted, 5.5), "สถานีนี้ยังไม่มีข้อมูลระดับตลิ่ง ระดับน้ำ 5.50 ม.รทก.", domain.LevelUnknown},
		{"no-bank-negative-msl", at(missing, -0.44), "สถานีนี้ยังไม่มีข้อมูลระดับตลิ่ง ระดับน้ำ -0.44 ม.รทก.", domain.LevelUnknown},
		{"no-bank-with-trend", hist(at(missing, 5.12), 5, 5.02, 5.05, 5.07, 5.1, 5.11), "สถานีนี้ยังไม่มีข้อมูลระดับตลิ่ง ระดับน้ำ 5.12 ม.รทก. สูงขึ้น 12 ซม. ใน 6 ชั่วโมงที่ผ่านมา", domain.LevelUnknown},

		// Trend over the hourly buckets.
		{"rising", rising, "ต่ำกว่าตลิ่ง 1.40 เมตร สูงขึ้น 12 ซม. ใน 6 ชั่วโมงที่ผ่านมา", domain.LevelHigh},
		{"steady", hist(at(complete, 8.62), 8.6, 8.61, 8.6, 8.6, 8.62, 8.62), "ต่ำกว่าตลิ่ง 1.38 เมตร ทรงตัว", domain.LevelHigh},
		{"steady-49mm", hist(at(complete, 8.649), 8.6, 8.6, 8.6, 8.6, 8.6, 8.6), "ต่ำกว่าตลิ่ง 1.35 เมตร ทรงตัว", domain.LevelHigh},
		{"rising-50mm", hist(at(complete, 8.65), 8.6, 8.6, 8.6, 8.6, 8.6, 8.6), "ต่ำกว่าตลิ่ง 1.35 เมตร สูงขึ้น 5 ซม. ใน 6 ชั่วโมงที่ผ่านมา", domain.LevelHigh},
		{"steady-falling-49mm", hist(at(complete, 8.551), 8.6, 8.6, 8.6, 8.6, 8.6, 8.6), "ต่ำกว่าตลิ่ง 1.45 เมตร ทรงตัว", domain.LevelHigh},
		{"rising-under-a-metre", hist(at(complete, 9.594), 8.6, 8.8, 9, 9.2, 9.4, 9.5), "ต่ำกว่าตลิ่ง 0.41 เมตร สูงขึ้น 99 ซม. ใน 6 ชั่วโมงที่ผ่านมา", domain.LevelHigh},
		{"rising-rounds-to-a-metre", hist(at(complete, 9.596), 8.6, 8.8, 9, 9.2, 9.4, 9.5), "ต่ำกว่าตลิ่ง 0.40 เมตร สูงขึ้น 1.00 เมตร ใน 6 ชั่วโมงที่ผ่านมา", domain.LevelHigh},
		{"falling-metres", hist(at(complete, 8.6), 9.8, 9.6, 9.3, 9.1, 8.9, 8.7), "ต่ำกว่าตลิ่ง 1.40 เมตร ลดลง 1.20 เมตร ใน 6 ชั่วโมงที่ผ่านมา", domain.LevelHigh},
		{"falling-half-cm", hist(at(complete, 8.6), 8.675, 8.66, 8.64, 8.63, 8.62, 8.61), "ต่ำกว่าตลิ่ง 1.40 เมตร ลดลง 8 ซม. ใน 6 ชั่วโมงที่ผ่านมา", domain.LevelHigh},
		{"overflow-rising", hist(at(complete, 10.35), 10, 10.1, 10.2, 10.25, 10.3, 10.33), "สูงกว่าตลิ่ง 0.35 เมตร (น้ำล้นตลิ่ง) สูงขึ้น 35 ซม. ใน 6 ชั่วโมงที่ผ่านมา", domain.LevelOverflow},
		{"short-span-says-so", hist(at(complete, 8.6), nan, nan, 8.5, 8.52, 8.55, nan), "ต่ำกว่าตลิ่ง 1.40 เมตร สูงขึ้น 10 ซม. ใน 4 ชั่วโมงที่ผ่านมา", domain.LevelHigh},
		{"sparse-three-points", hist(at(complete, 8.6), nan, nan, nan, 8.5, nan, 8.55), "ต่ำกว่าตลิ่ง 1.40 เมตร", domain.LevelHigh},
		{"sparse-no-history", at(complete, 8.6), "ต่ำกว่าตลิ่ง 1.40 เมตร", domain.LevelHigh},
		{"ten-minute-latest", hist(moved(at(complete, 8.6), at1410), 8.48, 8.5, 8.52, 8.55, 8.57, 8.59), "ต่ำกว่าตลิ่ง 1.40 เมตร สูงขึ้น 12 ซม. ใน 6 ชั่วโมงที่ผ่านมา", domain.LevelHigh},

		// Stale: lead with the age, then the last known state, never a trend.
		{"stale-same-day", stale(rising, at0900), "ข้อมูลล่าสุดเมื่อ 09:00 น. (ยังไม่อัปเดต) ตอนนั้นต่ำกว่าตลิ่ง 1.40 เมตร", domain.LevelHigh},
		{"stale-other-day", stale(at(complete, 10.35), at2000), "ข้อมูลล่าสุดเมื่อ 28 ต.ค. 69 20:00 น. (ยังไม่อัปเดต) ตอนนั้นสูงกว่าตลิ่ง 0.35 เมตร (น้ำล้นตลิ่ง)", domain.LevelOverflow},
		{"stale-at-bank", stale(at(complete, 10), at0900), "ข้อมูลล่าสุดเมื่อ 09:00 น. (ยังไม่อัปเดต) ตอนนั้นระดับน้ำเท่ากับตลิ่ง", domain.LevelHigh},
		{"stale-bank-only", stale(at(bankOnly, 11.01), at0900), "ข้อมูลล่าสุดเมื่อ 09:00 น. (ยังไม่อัปเดต) ตอนนั้นสูงกว่าตลิ่ง 0.11 เมตร", domain.LevelUnknown},
		{"stale-no-bank", stale(at(zero, 331.73), at0900), "ข้อมูลล่าสุดเมื่อ 09:00 น. (ยังไม่อัปเดต) สถานีนี้ยังไม่มีข้อมูลระดับตลิ่ง ตอนนั้นระดับน้ำ 331.73 ม.รทก.", domain.LevelUnknown},

		// Suspect: say so plainly, show no level, no distance, no trend.
		{"suspect-below-ground", suspect(at(complete, -30), domain.SuspectBelowGround), "ค่าที่รายงานล่าสุดผิดปกติ (ต่ำกว่าระดับอ้างอิงของสถานีเกิน 20 เมตร) จึงไม่นำมาแสดงเป็นระดับน้ำปัจจุบัน", domain.LevelUnknown},
		{"suspect-two-reasons", suspect(rising, domain.SuspectImplausiblePct, domain.SuspectJump), "ค่าที่รายงานล่าสุดผิดปกติ (สูงเกิน 4 เท่าของความสูงตลิ่ง และเปลี่ยนไปเกิน 3 เมตรจากชั่วโมงก่อนหน้า) จึงไม่นำมาแสดงเป็นระดับน้ำปัจจุบัน", domain.LevelUnknown},
		{"suspect-future", suspect(at(complete, 8.6), domain.SuspectFuture), "ค่าที่รายงานล่าสุดผิดปกติ (เวลาที่รายงานอยู่ในอนาคต) จึงไม่นำมาแสดงเป็นระดับน้ำปัจจุบัน", domain.LevelUnknown},
		{"suspect-unknown-reason", suspect(at(missing, 8.6), domain.SuspectReason(99)), "ค่าที่รายงานล่าสุดผิดปกติ (ไม่ผ่านการตรวจสอบ) จึงไม่นำมาแสดงเป็นระดับน้ำปัจจุบัน", domain.LevelUnknown},
		{"suspect-and-stale", stale(suspect(at(complete, 50), domain.SuspectImplausiblePct), at0900), "ข้อมูลล่าสุดเมื่อ 09:00 น. (ยังไม่อัปเดต) ค่าที่รายงานล่าสุดผิดปกติ (สูงเกิน 4 เท่าของความสูงตลิ่ง) จึงไม่นำมาแสดงเป็นระดับน้ำปัจจุบัน", domain.LevelUnknown},

		// Nothing to describe.
		{"no-reading", Input{Station: complete, Now: now}, "ยังไม่มีข้อมูลระดับน้ำจากสถานีนี้", domain.LevelUnknown},
		{"non-finite-reading", at(complete, math.Inf(1)), "ยังไม่มีข้อมูลระดับน้ำจากสถานีนี้", domain.LevelUnknown},
	}
	if len(cases) < 30 {
		t.Fatalf("the spec asks for at least 30 table cases, have %d", len(cases))
	}
	seen := map[string]bool{}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if seen[c.name] {
				t.Fatalf("duplicate case name %q", c.name)
			}
			seen[c.name] = true
			out := Describe(c.in)
			if out.Sentence != c.want {
				t.Errorf("sentence\n got %q\nwant %q", out.Sentence, c.want)
			}
			if out.Level != c.level {
				t.Errorf("level = %d, want %d", out.Level, c.level)
			}
			checkOutput(t, c.in, out)
		})
	}
}

func TestTrend(t *testing.T) {
	complete := station(domain.RefsComplete)
	cases := []struct {
		name string
		in   Input
		want *Trend
	}{
		{"six hours", hist(at(complete, 8.6), 8.48, 8.5, 8.52, 8.55, 8.57, 8.59), &Trend{DeltaMM: 120, Hours: 6, Points: 7}},
		{"four points", hist(at(complete, 8.6), nan, nan, 8.5, 8.52, 8.55, nan), &Trend{DeltaMM: 100, Hours: 4, Points: 4}},
		{"three points", hist(at(complete, 8.6), nan, nan, nan, 8.5, nan, 8.55), nil},
		{"steady still has a trend", hist(at(complete, 8.6), 8.6, 8.6, 8.6, 8.6, 8.6, 8.6), &Trend{DeltaMM: 0, Hours: 6, Points: 7}},
		{"float noise", hist(at(complete, 40.159998), 40.6, nan, 40.4, nan, 40.2, nan), &Trend{DeltaMM: -440, Hours: 6, Points: 4}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Describe(c.in).Trend
			switch {
			case c.want == nil && got != nil:
				t.Errorf("trend = %+v, want none", *got)
			case c.want != nil && got == nil:
				t.Errorf("no trend, want %+v", *c.want)
			case c.want != nil && *got != *c.want:
				t.Errorf("trend = %+v, want %+v", *got, *c.want)
			}
		})
	}
}

func TestTrendOnlyUsesTheSixHoursBefore(t *testing.T) {
	complete := station(domain.RefsComplete)
	in := hist(at(complete, 8.6), 8.48, 8.5, 8.52, 8.55, 8.57, 8.59)
	h := domain.HourEnd(in.Latest.At)
	in.Hourly = append(in.Hourly,
		domain.Point{At: h.Add(-7 * time.Hour), WL: 1, OK: true}, // before the window
		domain.Point{At: h, WL: 50, OK: true},                    // the latest's own hour: the reading wins
		domain.Point{At: h.Add(time.Hour), WL: 70, OK: true},     // after the reading
		domain.Point{At: h.Add(-6 * time.Hour), WL: 3, OK: true}, // a duplicate hour: the first one wins
		domain.Point{At: h.Add(-3 * time.Hour), WL: nan, OK: true},
	)
	got := Describe(in).Trend
	if got == nil || *got != (Trend{DeltaMM: 120, Hours: 6, Points: 7}) {
		t.Errorf("trend = %+v, want Δ 120 mm over 6 h from 7 points", got)
	}
}

func TestTrendPhrase(t *testing.T) {
	cases := []struct {
		tr   Trend
		want string
	}{
		{Trend{DeltaMM: 0, Hours: 6}, "ทรงตัว"},
		{Trend{DeltaMM: 49, Hours: 6}, "ทรงตัว"},
		{Trend{DeltaMM: -49, Hours: 6}, "ทรงตัว"},
		{Trend{DeltaMM: 50, Hours: 6}, "สูงขึ้น 5 ซม. ใน 6 ชั่วโมงที่ผ่านมา"},
		{Trend{DeltaMM: -50, Hours: 5}, "ลดลง 5 ซม. ใน 5 ชั่วโมงที่ผ่านมา"},
		{Trend{DeltaMM: 994, Hours: 6}, "สูงขึ้น 99 ซม. ใน 6 ชั่วโมงที่ผ่านมา"},
		{Trend{DeltaMM: 995, Hours: 6}, "สูงขึ้น 1.00 เมตร ใน 6 ชั่วโมงที่ผ่านมา"},
		{Trend{DeltaMM: -1770, Hours: 4}, "ลดลง 1.77 เมตร ใน 4 ชั่วโมงที่ผ่านมา"},
	}
	for _, c := range cases {
		if got := c.tr.Phrase(); got != c.want {
			t.Errorf("%+v.Phrase() = %q, want %q", c.tr, got, c.want)
		}
		if got := c.tr.Steady(); got != (c.want == "ทรงตัว") {
			t.Errorf("%+v.Steady() = %v", c.tr, got)
		}
	}
}

func TestFactsShowTheWorking(t *testing.T) {
	complete := station(domain.RefsComplete)
	bankOnly := domain.Station{Refs: domain.RefsBankOnly, BankFull: 10.9}
	cases := []struct {
		name string
		in   Input
		want []Fact
	}{
		{"complete with a trend", hist(at(complete, 8.6), 8.48, 8.5, 8.52, 8.55, 8.57, 8.59), []Fact{
			{FactMeasuredAt, "14:00 น."},
			{FactWaterLevel, "8.60 ม.รทก."},
			{FactBankFull, "10.00 ม.รทก."},
			{FactGround, "2.00 ม.รทก."},
			{FactPercent, "82.50%"},
			{FactDistance, "ต่ำกว่าตลิ่ง 1.40 เมตร"},
			{FactTrend, "สูงขึ้น 12 ซม. ใน 6 ชั่วโมงที่ผ่านมา"},
		}},
		{"bank-only", at(bankOnly, 11.01), []Fact{
			{FactMeasuredAt, "14:00 น."},
			{FactWaterLevel, "11.01 ม.รทก."},
			{FactBankFull, "10.90 ม.รทก."},
			{FactDistance, "สูงกว่าตลิ่ง 0.11 เมตร"},
		}},
		{"no bank", at(domain.Station{Refs: domain.RefsZeroCoded}, 331.73), []Fact{
			{FactMeasuredAt, "14:00 น."},
			{FactWaterLevel, "331.73 ม.รทก."},
		}},
		{"stale from yesterday", stale(at(complete, 10.35), at2000), []Fact{
			{FactMeasuredAt, "28 ต.ค. 69 20:00 น."},
			{FactWaterLevel, "10.35 ม.รทก."},
			{FactBankFull, "10.00 ม.รทก."},
			{FactGround, "2.00 ม.รทก."},
			{FactPercent, "104.38%"},
			{FactDistance, "สูงกว่าตลิ่ง 0.35 เมตร"},
		}},
		// The suspect value itself is not offered as a fact, so no client
		// can show it as the water level by accident.
		{"suspect", suspect(at(complete, 50), domain.SuspectImplausiblePct), []Fact{
			{FactMeasuredAt, "14:00 น."},
			{FactSuspect, "สูงเกิน 4 เท่าของความสูงตลิ่ง"},
		}},
		// 23:00 yesterday is current at 00:30, and its date is shown.
		{"current across midnight", Input{
			Station: domain.Station{Refs: domain.RefsMissing},
			Latest:  &domain.Reading{At: time.Date(2026, 10, 28, 16, 0, 0, 0, time.UTC), WL: 1},
			Now:     time.Date(2026, 10, 28, 17, 30, 0, 0, time.UTC),
		}, []Fact{
			{FactMeasuredAt, "28 ต.ค. 69 23:00 น."},
			{FactWaterLevel, "1.00 ม.รทก."},
		}},
		{"no reading", Input{Station: complete, Now: now}, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Describe(c.in).Facts
			if len(got) != len(c.want) {
				t.Fatalf("facts = %v, want %v", got, c.want)
			}
			for i := range got {
				if got[i] != c.want[i] {
					t.Errorf("fact %d = %+v, want %+v", i, got[i], c.want[i])
				}
			}
		})
	}
}

func TestPercentText(t *testing.T) {
	cases := map[int64]string{
		8250: "82.50%", 0: "0.00%", -1: "-0.01%", -1893: "-18.93%", 5: "0.05%",
		10000: "100.00%", 10572: "105.72%", 40000: "400.00%", 123456: "1,234.56%",
	}
	for in, want := range cases {
		if got := percentText(in); got != want {
			t.Errorf("percentText(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestReasonsAreCopied(t *testing.T) {
	rs := []domain.SuspectReason{domain.SuspectJump}
	out := Describe(suspect(at(station(domain.RefsComplete), 20), rs...))
	rs[0] = domain.SuspectFuture
	if len(out.Reasons) != 1 || out.Reasons[0] != domain.SuspectJump {
		t.Errorf("Reasons = %v; the output must not alias the input", out.Reasons)
	}
}

func TestNational(t *testing.T) {
	rep := func(p string, n int) []string {
		out := make([]string, n)
		for i := range out {
			out[i] = p
		}
		return out
	}
	join := func(groups ...[]string) []string {
		var out []string
		for _, g := range groups {
			out = append(out, g...)
		}
		return out
	}
	ten := []string{"เชียงใหม่", "ลำพูน", "ลำปาง", "แพร่", "น่าน", "พะเยา", "เชียงราย", "ตาก", "สุโขทัย", "พิษณุโลก"}
	cases := []struct {
		name string
		in   Tally
		want string
	}{
		{"national-none", Tally{}, "ยังไม่มีข้อมูลล่าสุดจากสถานีวัดระดับน้ำ"},
		{"national-none-ignores-counts", Tally{High: 3, Overflow: []string{"ชุมพร"}}, "ยังไม่มีข้อมูลล่าสุดจากสถานีวัดระดับน้ำ"},
		{"national-zero", Tally{Current: 700}, "ตอนนี้ไม่มีสถานีที่น้ำล้นตลิ่ง"},
		{"national-zero-one-high", Tally{Current: 700, High: 1}, "ตอนนี้ไม่มีสถานีที่น้ำล้นตลิ่ง มี 1 สถานีที่น้ำมาก"},
		{"national-zero-many-high", Tally{Current: 700, High: 145}, "ตอนนี้ไม่มีสถานีที่น้ำล้นตลิ่ง มี 145 สถานีที่น้ำมาก"},
		{"national-one", Tally{Current: 700, High: 9, Overflow: []string{"ชุมพร"}}, "ตอนนี้ 1 สถานีน้ำล้นตลิ่ง อยู่ที่ชุมพร"},
		{"national-one-unknown-province", Tally{Current: 700, Overflow: []string{" "}}, "ตอนนี้ 1 สถานีน้ำล้นตลิ่ง"},
		{"national-one-province", Tally{Current: 700, Overflow: rep("ชุมพร", 3)}, "ตอนนี้ 3 สถานีน้ำล้นตลิ่ง ทั้งหมดอยู่ที่ชุมพร"},
		{"national-all-two", Tally{Current: 700, Overflow: []string{"ชุมพร", "พระนครศรีอยุธยา", "พระนครศรีอยุธยา"}}, "ตอนนี้ 3 สถานีน้ำล้นตลิ่ง ทั้งหมดอยู่ที่พระนครศรีอยุธยา และชุมพร"},
		{"national-all-three-ties-by-name", Tally{Current: 700, Overflow: []string{"สุราษฎร์ธานี", "ชุมพร", "ระนอง"}}, "ตอนนี้ 3 สถานีน้ำล้นตลิ่ง ทั้งหมดอยู่ที่ชุมพร ระนอง และสุราษฎร์ธานี"},
		// The spec's own example: 6 + 4 + 2 = 12 of 22 is more than half.
		{"national-majority-spec-example", Tally{Current: 700, Overflow: join(rep("พระนครศรีอยุธยา", 6), rep("สุราษฎร์ธานี", 4), rep("ชุมพร", 2), ten)}, "ตอนนี้ 22 สถานีน้ำล้นตลิ่ง ส่วนใหญ่อยู่ที่พระนครศรีอยุธยา สุราษฎร์ธานี และชุมพร"},
		{"national-majority-one", Tally{Current: 700, Overflow: join(rep("พระนครศรีอยุธยา", 6), ten[:5])}, "ตอนนี้ 11 สถานีน้ำล้นตลิ่ง ส่วนใหญ่อยู่ที่พระนครศรีอยุธยา"},
		// Exactly half is not "most": 4 of 8 needs a second province.
		{"national-majority-half-is-not-most", Tally{Current: 700, Overflow: join(rep("ชุมพร", 4), rep("ระนอง", 2), ten[:2])}, "ตอนนี้ 8 สถานีน้ำล้นตลิ่ง ส่วนใหญ่อยู่ที่ชุมพร และระนอง"},
		// A tie at the cut-off would make the choice of names arbitrary
		// (why ลำปาง and not ลำพูน?), so no province is named past it. This
		// is the shape of the 17:54 recording: 2, then four provinces at 1.
		{"national-tie-at-the-cutoff", Tally{Current: 700, Overflow: join(rep("พระนครศรีอยุธยา", 2), ten[:4])}, "ตอนนี้ 6 สถานีน้ำล้นตลิ่ง กระจายอยู่ใน 5 จังหวัด"},
		{"national-tie-inside-the-named", Tally{Current: 700, Overflow: join(rep("ชุมพร", 3), rep("ระนอง", 3), ten[:2])}, "ตอนนี้ 8 สถานีน้ำล้นตลิ่ง ส่วนใหญ่อยู่ที่ชุมพร และระนอง"},
		{"national-spread", Tally{Current: 700, Overflow: ten}, "ตอนนี้ 10 สถานีน้ำล้นตลิ่ง กระจายอยู่ใน 10 จังหวัด"},
		{"national-majority-with-unknown", Tally{Current: 700, Overflow: join(rep("ชุมพร", 4), []string{"", "ตาก"})}, "ตอนนี้ 6 สถานีน้ำล้นตลิ่ง ส่วนใหญ่อยู่ที่ชุมพร"},
		// With an unknown province "ทั้งหมด" and a province count would
		// both be guesses, so the location is left out.
		{"national-few-with-unknown", Tally{Current: 700, Overflow: []string{"ชุมพร", ""}}, "ตอนนี้ 2 สถานีน้ำล้นตลิ่ง"},
		{"national-spread-with-unknown", Tally{Current: 700, Overflow: join(ten[:4], []string{"", ""})}, "ตอนนี้ 6 สถานีน้ำล้นตลิ่ง"},
		{"national-thousands", Tally{Current: 1200, Overflow: rep("ชุมพร", 1001)}, "ตอนนี้ 1,001 สถานีน้ำล้นตลิ่ง ทั้งหมดอยู่ที่ชุมพร"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := National(c.in)
			if got != c.want {
				t.Errorf("\n got %q\nwant %q", got, c.want)
			}
			checkSentence(t, got)
		})
	}
}

// banned are words that predict, reassure or speak of flooding an area. The
// product reports water against the bank and says nothing about the future.
var banned = []string{
	"ปลอดภัย",                // safe
	"จะ",                     // will: any future claim, including จะท่วม and จะล้น
	"ท่วม",                   // flood, including ไม่ท่วม; we never say an area floods
	"รับรอง",                 // guarantee
	"แน่นอน",                 // certainly
	"กังวล", "ห่วง", "วางใจ", // worry, concern, rest assured
	"สบายใจ",       // at ease
	"คาด",          // expect or forecast (คาดว่า, คาดการณ์)
	"น่าจะ", "อาจ", // probably, might
	"พยากรณ์", "ทำนาย", // forecast, predict
	"แนวโน้ม",          // tendency: the trend is only ever the past six hours
	"เตือนภัย", "อพยพ", // warnings and evacuation are not ours to issue
}

// checkSentence holds for every sentence the package can produce.
func checkSentence(t *testing.T, s string) {
	t.Helper()
	switch {
	case s == "":
		t.Error("empty sentence")
	case !utf8.ValidString(s):
		t.Errorf("invalid UTF-8: %q", s)
	case strings.TrimSpace(s) != s, strings.Contains(s, "  "):
		t.Errorf("stray spaces in %q", s)
	}
	for _, junk := range []string{"NaN", "Inf", "%!", "-0.00", "<nil>", "\n"} {
		if strings.Contains(s, junk) {
			t.Errorf("%q contains %q", s, junk)
		}
	}
	for _, b := range banned {
		if strings.Contains(s, b) {
			t.Errorf("%q contains the banned %q", s, b)
		}
	}
	if (strings.Contains(s, "สูงขึ้น") || strings.Contains(s, "ลดลง")) && !strings.Contains(s, "ชั่วโมงที่ผ่านมา") {
		t.Errorf("a change without its past window: %q", s)
	}
}

// checkOutput holds between every input and its output.
func checkOutput(t *testing.T, in Input, out Output) {
	t.Helper()
	checkSentence(t, out.Sentence)
	for _, f := range out.Facts {
		checkSentence(t, f.Text)
	}
	s := out.Sentence
	if out.Stale != in.Stale {
		t.Errorf("Stale = %v, want %v", out.Stale, in.Stale)
	}
	if out.Suspect != (len(in.Suspect) > 0) {
		t.Errorf("Suspect = %v with reasons %v", out.Suspect, in.Suspect)
	}
	if (out.Level == domain.LevelOverflow) != strings.Contains(s, "(น้ำล้นตลิ่ง)") {
		t.Errorf("level %d and sentence %q disagree about overflow", out.Level, s)
	}
	if out.Level == domain.LevelOverflow && !strings.Contains(s, "สูงกว่าตลิ่ง") {
		t.Errorf("overflow without saying above the bank: %q", s)
	}
	if out.Level != domain.LevelUnknown && in.Station.Refs != domain.RefsComplete {
		t.Errorf("level %d from refs %v", out.Level, in.Station.Refs)
	}
	if in.Latest == nil {
		return
	}
	if out.Trend != nil && (in.Stale || out.Suspect || out.Trend.Points < minTrendPoints || out.Trend.Hours < 3 || out.Trend.Hours > trendHours) {
		t.Errorf("trend %+v for stale=%v suspect=%v", *out.Trend, in.Stale, out.Suspect)
	}
	if in.Stale && !strings.HasPrefix(s, "ข้อมูลล่าสุดเมื่อ ") {
		t.Errorf("stale sentence does not lead with its age: %q", s)
	}
	if out.Suspect {
		if out.Level != domain.LevelUnknown || !strings.HasSuffix(s, "จึงไม่นำมาแสดงเป็นระดับน้ำปัจจุบัน") || strings.Contains(s, "ม.รทก.") || strings.Contains(s, "กว่าตลิ่ง ") {
			t.Errorf("suspect reading presented as a value: level %d, %q", out.Level, s)
		}
		return
	}
	finite := !math.IsNaN(in.Latest.WL) && !math.IsInf(in.Latest.WL, 0)
	if finite && !in.Station.HasBankFull() && !strings.Contains(s, "สถานีนี้ยังไม่มีข้อมูลระดับตลิ่ง ") {
		t.Errorf("a station without bank-full made a claim: %q", s)
	}
	if finite && in.Station.HasBankFull() && !strings.Contains(s, "ตลิ่ง") {
		t.Errorf("a station with bank-full gave no distance: %q", s)
	}
}
