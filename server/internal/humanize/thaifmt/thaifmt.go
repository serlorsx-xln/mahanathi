// Package thaifmt formats numbers, times and dates the way the product
// writes them in Thai: "1.40 เมตร", "12 ซม.", "14:00 น.", "29 ต.ค. 69".
//
// Lengths are handled as whole millimetres. Rounding a float straight to
// centimetres goes wrong at the half: 1.395 is 1.39499… in float64 and
// would print as 1.39. Rounding to the nearest millimetre first (which
// undoes the float error), then rounding half away from zero to
// centimetres, gives 1.40 as a person would expect.
//
// Times are always shown in Indochina Time, whatever zone the input carries.
// If the client ever formats anything itself, it must follow these tests.
package thaifmt

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

var ict = time.FixedZone("ICT", 7*3600)

var months = [...]string{"ม.ค.", "ก.พ.", "มี.ค.", "เม.ย.", "พ.ค.", "มิ.ย.", "ก.ค.", "ส.ค.", "ก.ย.", "ต.ค.", "พ.ย.", "ธ.ค."}

// MM converts metres to the nearest whole millimetre.
func MM(m float64) int64 { return int64(math.Round(m * 1000)) }

// Centimetres rounds millimetres to centimetres, half away from zero.
func Centimetres(mm int64) int64 {
	if mm < 0 {
		return -((-mm + 5) / 10)
	}
	return (mm + 5) / 10
}

// Metres formats millimetres as metres with exactly two decimals and
// thousands separators. A value that rounds to zero is "0.00", never
// "-0.00".
func Metres(mm int64) string {
	cm := Centimetres(mm)
	sign := ""
	if cm < 0 {
		sign, cm = "-", -cm
	}
	return fmt.Sprintf("%s%s.%02d", sign, Int(int(cm/100)), cm%100)
}

// Int formats an integer with comma thousands separators, as Thai news and
// government sites do.
func Int(n int) string {
	s := strconv.Itoa(n)
	neg := strings.HasPrefix(s, "-")
	if neg {
		s = s[1:]
	}
	var b strings.Builder
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(r)
	}
	if neg {
		return "-" + b.String()
	}
	return b.String()
}

// Time is the clock time in Bangkok, "14:00 น.".
func Time(t time.Time) string { return t.In(ict).Format("15:04") + " น." }

// Date is the short Thai date with the Buddhist-era year, "29 ต.ค. 69".
func Date(t time.Time) string {
	l := t.In(ict)
	return fmt.Sprintf("%d %s %02d", l.Day(), months[l.Month()-1], (l.Year()+543)%100)
}

// DateTime is Date followed by Time, "29 ต.ค. 69 14:00 น.".
func DateTime(t time.Time) string { return Date(t) + " " + Time(t) }

// SameDay reports whether two instants fall on the same Bangkok calendar day.
func SameDay(a, b time.Time) bool {
	ay, am, ad := a.In(ict).Date()
	by, bm, bd := b.In(ict).Date()
	return ay == by && am == bm && ad == bd
}

// List joins names the Thai way: "ก ข และค". The last item takes "และ"
// with no space, as in "พระนครศรีอยุธยา สุราษฎร์ธานี และชุมพร".
func List(items []string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	default:
		return strings.Join(items[:len(items)-1], " ") + " และ" + items[len(items)-1]
	}
}
