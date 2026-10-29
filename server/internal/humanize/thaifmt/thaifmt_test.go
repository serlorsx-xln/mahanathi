package thaifmt

import (
	"testing"
	"time"
)

func TestCentimetresRoundsHalfAwayFromZero(t *testing.T) {
	cases := []struct {
		mm   int64
		want int64
	}{
		{0, 0}, {4, 0}, {5, 1}, {14, 1}, {15, 2},
		{1395, 140}, // 1.395 m is 140 cm, not 139
		{-4, 0}, {-5, -1}, {-1395, -140},
		{996, 100},
	}
	for _, c := range cases {
		if got := Centimetres(c.mm); got != c.want {
			t.Errorf("Centimetres(%d) = %d, want %d", c.mm, got, c.want)
		}
	}
}

func TestMetres(t *testing.T) {
	cases := []struct {
		mm   int64
		want string
	}{
		{0, "0.00"},
		{1400, "1.40"},
		{1395, "1.40"},
		{1394, "1.39"},
		{350, "0.35"},
		{5, "0.01"},
		{4, "0.00"},
		{-1400, "-1.40"},
		{-4, "0.00"}, // never "-0.00"
		{12345678, "12,345.68"},
		{224320, "224.32"},
	}
	for _, c := range cases {
		if got := Metres(c.mm); got != c.want {
			t.Errorf("Metres(%d) = %q, want %q", c.mm, got, c.want)
		}
	}
}

func TestMM(t *testing.T) {
	cases := []struct {
		m    float64
		want int64
	}{
		{1.395, 1395}, // 1.395 is 1.39499… in float64; rounding to mm first keeps it
		{0.1 + 0.2, 300},
		{-0.44, -440},
		{40.599998, 40600},
		{3.22 - 3.17, 50}, // 0.04999999999999982 in float64
	}
	for _, c := range cases {
		if got := MM(c.m); got != c.want {
			t.Errorf("MM(%v) = %d, want %d", c.m, got, c.want)
		}
	}
}

func TestInt(t *testing.T) {
	cases := map[int]string{0: "0", 7: "7", 22: "22", 804: "804", 1000: "1,000", 12345: "12,345", 1234567: "1,234,567", -1234: "-1,234"}
	for n, want := range cases {
		if got := Int(n); got != want {
			t.Errorf("Int(%d) = %q, want %q", n, got, want)
		}
	}
}

func TestTimeAndDateAreICT(t *testing.T) {
	// 07:00 UTC is 14:00 in Bangkok.
	at := time.Date(2026, 10, 29, 7, 0, 0, 0, time.UTC)
	if got := Time(at); got != "14:00 น." {
		t.Errorf("Time = %q, want 14:00 น.", got)
	}
	if got := Date(at); got != "29 ต.ค. 69" {
		t.Errorf("Date = %q, want 29 ต.ค. 69", got)
	}
	if got := DateTime(at); got != "29 ต.ค. 69 14:00 น." {
		t.Errorf("DateTime = %q", got)
	}
	// 17:30 UTC on the 28th is already 00:30 on the 29th in Bangkok.
	late := time.Date(2026, 10, 28, 17, 30, 0, 0, time.UTC)
	if got := DateTime(late); got != "29 ต.ค. 69 00:30 น." {
		t.Errorf("DateTime(late) = %q", got)
	}
	// The input zone does not matter, only the instant.
	ny := time.FixedZone("EST", -5*3600)
	if got := Time(at.In(ny)); got != "14:00 น." {
		t.Errorf("Time from another zone = %q", got)
	}
}

func TestEveryMonthAndTheBuddhistYear(t *testing.T) {
	want := []string{"ม.ค.", "ก.พ.", "มี.ค.", "เม.ย.", "พ.ค.", "มิ.ย.", "ก.ค.", "ส.ค.", "ก.ย.", "ต.ค.", "พ.ย.", "ธ.ค."}
	for i, w := range want {
		at := time.Date(2026, time.Month(i+1), 1, 5, 0, 0, 0, time.UTC)
		if got := Date(at); got != "1 "+w+" 69" {
			t.Errorf("month %d: Date = %q, want %q", i+1, got, "1 "+w+" 69")
		}
	}
	// 2027 is BE 2570; the short year is its last two digits.
	if got := Date(time.Date(2027, 1, 5, 5, 0, 0, 0, time.UTC)); got != "5 ม.ค. 70" {
		t.Errorf("Date(2027) = %q", got)
	}
	if got := Date(time.Date(2057, 1, 5, 5, 0, 0, 0, time.UTC)); got != "5 ม.ค. 00" {
		t.Errorf("Date(BE 2600) = %q, want a zero-padded year", got)
	}
}

func TestSameDay(t *testing.T) {
	a := time.Date(2026, 10, 29, 16, 59, 0, 0, time.UTC) // 23:59 ICT on the 29th
	b := time.Date(2026, 10, 29, 17, 0, 0, 0, time.UTC)  // 00:00 ICT on the 30th
	if SameDay(a, b) {
		t.Error("23:59 and 00:00 Bangkok time were treated as one day")
	}
	if !SameDay(a, a.Add(-23*time.Hour)) {
		t.Error("00:59 and 23:59 on the same Bangkok day were treated as two days")
	}
}

func TestList(t *testing.T) {
	cases := []struct {
		in   []string
		want string
	}{
		{nil, ""},
		{[]string{"ชุมพร"}, "ชุมพร"},
		{[]string{"พระนครศรีอยุธยา", "ชุมพร"}, "พระนครศรีอยุธยา และชุมพร"},
		{[]string{"พระนครศรีอยุธยา", "สุราษฎร์ธานี", "ชุมพร"}, "พระนครศรีอยุธยา สุราษฎร์ธานี และชุมพร"},
	}
	for _, c := range cases {
		if got := List(c.in); got != c.want {
			t.Errorf("List(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}
