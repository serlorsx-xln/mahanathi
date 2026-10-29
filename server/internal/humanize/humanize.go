package humanize

import (
	"fmt"
	"math"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/serlorsx-xln/mahanathi/server/internal/domain"
	"github.com/serlorsx-xln/mahanathi/server/internal/humanize/thaifmt"
)

const (
	trendHours     = 6  // the trend looks back six hourly buckets
	minTrendPoints = 4  // the spec's minimum, counting the latest reading
	steadyMM       = 50 // |Δ| under 5 cm is ทรงตัว
	tinyMM         = 5  // under half a centimetre, which would print as 0.00
)

// Input is one station's latest state. The caller decides staleness and
// suspicion (normalize.IsStale, normalize.SuspectReasons); humanize only
// words them.
type Input struct {
	Station domain.Station
	// Latest is the newest reading, or nil if the station has none.
	Latest *domain.Reading
	// Hourly are hourly buckets around the latest reading, in any order.
	// Only the six before the latest reading's own hour are used.
	Hourly  []domain.Point
	Stale   bool
	Suspect []domain.SuspectReason
	// Now decides whether a time needs its date. It is never read from the
	// clock here.
	Now time.Time
}

// Trend is the change over the hourly buckets before the latest reading.
type Trend struct {
	DeltaMM int64 // latest minus the earliest bucket, in mm
	Hours   int   // hours between them: 6, or fewer when the start is missing
	Points  int   // buckets with a reading, the latest included
}

// Steady reports a change of under 5 cm either way.
func (t Trend) Steady() bool { return t.DeltaMM > -steadyMM && t.DeltaMM < steadyMM }

// Phrase is the trend in words: "ทรงตัว", "สูงขึ้น 12 ซม. ใน 6 ชั่วโมงที่ผ่านมา".
func (t Trend) Phrase() string {
	if t.Steady() {
		return "ทรงตัว"
	}
	verb, d := "สูงขึ้น", t.DeltaMM
	if d < 0 {
		verb, d = "ลดลง", -d
	}
	return fmt.Sprintf("%s %s ใน %d ชั่วโมงที่ผ่านมา", verb, length(d), t.Hours)
}

// FactKey names a fact so the client can lay facts out without parsing text.
type FactKey string

// The facts, in the order Describe lists them.
const (
	FactMeasuredAt FactKey = "measuredAt" // "14:00 น.", with the date when not today
	FactWaterLevel FactKey = "wl"         // "8.60 ม.รทก."
	FactBankFull   FactKey = "bankFull"   // "10.00 ม.รทก."
	FactGround     FactKey = "ground"     // "2.00 ม.รทก."
	FactPercent    FactKey = "pct"        // "82.50%"
	FactDistance   FactKey = "distance"   // "ต่ำกว่าตลิ่ง 1.40 เมตร"
	FactTrend      FactKey = "trend"      // "สูงขึ้น 12 ซม. ใน 6 ชั่วโมงที่ผ่านมา"
	FactSuspect    FactKey = "suspect"    // why the value is not shown
)

// Fact is one number behind the sentence, already formatted in Thai.
type Fact struct {
	Key  FactKey
	Text string
}

// Output is the sentence and everything a client needs to show its working.
type Output struct {
	Sentence string
	Facts    []Fact
	Level    domain.Level
	Stale    bool
	Suspect  bool
	Reasons  []domain.SuspectReason
	Trend    *Trend // nil when omitted: stale, suspect or too few points
}

// Describe words one station's latest reading, following the table in the
// package documentation.
func Describe(in Input) Output {
	if in.Latest == nil || !finite(in.Latest.WL) {
		return Output{Sentence: "ยังไม่มีข้อมูลระดับน้ำจากสถานีนี้"}
	}
	r, st := *in.Latest, in.Station
	when := timeText(r.At, in.Now)
	out := Output{
		Stale:   in.Stale,
		Suspect: len(in.Suspect) > 0,
		Facts:   []Fact{{FactMeasuredAt, when}},
	}
	var parts []string
	if in.Stale {
		parts = append(parts, "ข้อมูลล่าสุดเมื่อ "+when+" (ยังไม่อัปเดต)")
	}

	if out.Suspect {
		out.Reasons = slices.Clone(in.Suspect)
		why := reasonsText(in.Suspect)
		out.Facts = append(out.Facts, Fact{FactSuspect, why})
		parts = append(parts, "ค่าที่รายงานล่าสุดผิดปกติ ("+why+") จึงไม่นำมาแสดงเป็นระดับน้ำปัจจุบัน")
		out.Sentence = strings.Join(parts, " ")
		return out
	}

	// A stale state is the state at that time, not now.
	then := ""
	if in.Stale {
		then = "ตอนนั้น"
	}
	out.Facts = append(out.Facts, Fact{FactWaterLevel, msl(r.WL)})
	if d, ok := st.Distance(r.WL); ok {
		out.Facts = append(out.Facts, Fact{FactBankFull, msl(st.BankFull)})
		if pct, ok := st.PctHundredths(r.WL); ok {
			out.Level = st.Level(r.WL)
			out.Facts = append(out.Facts, Fact{FactGround, msl(st.Ground)}, Fact{FactPercent, percentText(pct)})
		}
		dist := distanceText(d)
		out.Facts = append(out.Facts, Fact{FactDistance, dist})
		if out.Level == domain.LevelOverflow {
			dist += " (" + domain.LevelOverflow.Label() + ")"
		}
		parts = append(parts, then+dist)
	} else {
		parts = append(parts, "สถานีนี้ยังไม่มีข้อมูลระดับตลิ่ง", then+"ระดับน้ำ "+msl(r.WL))
	}

	if !in.Stale {
		if tr, ok := trend(r, in.Hourly); ok {
			out.Trend = &tr
			phrase := tr.Phrase()
			out.Facts = append(out.Facts, Fact{FactTrend, phrase})
			parts = append(parts, phrase)
		}
	}
	out.Sentence = strings.Join(parts, " ")
	return out
}

// trend compares the latest reading with the earliest of the six hourly
// buckets before its own hour. The latest reading stands for its own hour,
// so a bucket for that hour in hourly is ignored.
func trend(latest domain.Reading, hourly []domain.Point) (Trend, bool) {
	h := domain.HourEnd(latest.At)
	var seen [trendHours + 1]bool
	var startWL float64
	start := 0 // hours back of the earliest bucket found
	n := 1     // the latest reading is a point
	for _, p := range hourly {
		if !p.OK || !finite(p.WL) {
			continue
		}
		back := int(h.Sub(domain.HourEnd(p.At)) / time.Hour)
		if back < 1 || back > trendHours || seen[back] {
			continue
		}
		seen[back] = true
		n++
		if back > start {
			start, startWL = back, p.WL
		}
	}
	if n < minTrendPoints {
		return Trend{}, false
	}
	return Trend{DeltaMM: domain.MM(latest.WL) - domain.MM(startWL), Hours: start, Points: n}, true
}

// distanceText words a signed distance from bank-full in mm.
func distanceText(d int64) string {
	switch {
	case d == 0:
		return "ระดับน้ำเท่ากับตลิ่ง"
	case d > 0 && d < tinyMM:
		return "สูงกว่าตลิ่งไม่ถึง 1 ซม."
	case d < 0 && d > -tinyMM:
		return "ต่ำกว่าตลิ่งไม่ถึง 1 ซม."
	case d > 0:
		return "สูงกว่าตลิ่ง " + thaifmt.Metres(d) + " เมตร"
	default:
		return "ต่ำกว่าตลิ่ง " + thaifmt.Metres(-d) + " เมตร"
	}
}

// length is a positive change: centimetres below 1 m after rounding, metres
// with two decimals from there.
func length(mm int64) string {
	if cm := thaifmt.Centimetres(mm); cm < 100 {
		return thaifmt.Int(int(cm)) + " ซม."
	}
	return thaifmt.Metres(mm) + " เมตร"
}

// msl is a level above mean sea level: "8.60 ม.รทก.".
func msl(m float64) string { return thaifmt.Metres(domain.MM(m)) + " ม.รทก." }

// percentText formats hundredths of a percent: 10438 is "104.38%".
func percentText(h int64) string {
	sign := ""
	if h < 0 {
		sign, h = "-", -h
	}
	return fmt.Sprintf("%s%s.%02d%%", sign, thaifmt.Int(int(h/100)), h%100)
}

// timeText is the clock time, with the date when it is not today in Bangkok.
func timeText(at, now time.Time) string {
	if thaifmt.SameDay(at, now) {
		return thaifmt.Time(at)
	}
	return thaifmt.DateTime(at)
}

// reasonsText says why a value is not believable. The bounds are the ones in
// normalize/assess.go, which come from the humanize spec.
func reasonsText(rs []domain.SuspectReason) string {
	words := make([]string, 0, len(rs))
	for _, r := range rs {
		switch r {
		case domain.SuspectBelowGround:
			words = append(words, "ต่ำกว่าระดับอ้างอิงของสถานีเกิน 20 เมตร")
		case domain.SuspectImplausiblePct:
			words = append(words, "สูงเกิน 4 เท่าของความสูงตลิ่ง")
		case domain.SuspectJump:
			words = append(words, "เปลี่ยนไปเกิน 3 เมตรจากชั่วโมงก่อนหน้า")
		case domain.SuspectFuture:
			words = append(words, "เวลาที่รายงานอยู่ในอนาคต")
		default:
			words = append(words, "ไม่ผ่านการตรวจสอบ")
		}
	}
	return thaifmt.List(words)
}

func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

// Tally is what the national sentence is built from. Count current,
// believable readings only: not stale and not suspect.
type Tally struct {
	Current  int      // stations with a current, believable reading
	High     int      // of those, stations at น้ำมาก
	Overflow []string // the province of each overflowing station; "" if unknown
}

// maxNamed is how many provinces the national sentence will name.
const maxNamed = 3

// National is the one-line national summary, following the table in the
// package documentation.
func National(t Tally) string {
	if t.Current == 0 {
		return "ยังไม่มีข้อมูลล่าสุดจากสถานีวัดระดับน้ำ"
	}
	n := len(t.Overflow)
	if n == 0 {
		s := "ตอนนี้ไม่มีสถานีที่น้ำล้นตลิ่ง"
		if t.High > 0 {
			s += " มี " + thaifmt.Int(t.High) + " สถานีที่น้ำมาก"
		}
		return s
	}
	s := "ตอนนี้ " + thaifmt.Int(n) + " สถานีน้ำล้นตลิ่ง"
	if w := where(t.Overflow); w != "" {
		s += " " + w
	}
	return s
}

// where says where the overflowing stations are, claiming "most" only when
// the named provinces hold more than half of them and each has more stations
// than any province left unnamed.
func where(provinces []string) string {
	count := map[string]int{}
	unknown := 0
	for _, p := range provinces {
		if p = strings.TrimSpace(p); p == "" {
			unknown++
		} else {
			count[p]++
		}
	}
	ranked := make([]string, 0, len(count))
	for p := range count {
		ranked = append(ranked, p)
	}
	sort.Slice(ranked, func(i, j int) bool {
		if count[ranked[i]] != count[ranked[j]] {
			return count[ranked[i]] > count[ranked[j]]
		}
		return ranked[i] < ranked[j]
	})

	total := len(provinces)
	switch {
	case total == 1 && unknown == 0:
		return "อยู่ที่" + ranked[0]
	case unknown == 0 && len(ranked) <= maxNamed:
		return "ทั้งหมดอยู่ที่" + thaifmt.List(ranked)
	}
	sum := 0
	for k := 0; k < len(ranked) && k < maxNamed; k++ {
		sum += count[ranked[k]]
		clean := k+1 == len(ranked) || count[ranked[k]] > count[ranked[k+1]]
		if 2*sum > total && clean {
			return "ส่วนใหญ่อยู่ที่" + thaifmt.List(ranked[:k+1])
		}
	}
	if unknown == 0 {
		return "กระจายอยู่ใน " + thaifmt.Int(len(ranked)) + " จังหวัด"
	}
	return ""
}
