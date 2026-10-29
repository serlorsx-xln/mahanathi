package humanize

// These tests run the real recorded ThaiWater responses through normalize
// and humanize, the way ingest and the API will. normalize is a test-only
// import; the package itself depends on domain and thaifmt alone.

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/serlorsx-xln/mahanathi/server/internal/domain"
	"github.com/serlorsx-xln/mahanathi/server/internal/normalize"
	"github.com/serlorsx-xln/mahanathi/server/internal/thaiwater"
)

var update = flag.Bool("update", false, "rewrite the golden files in testdata/")

const fixtures = "../thaiwater/testdata/"

// fetched is when waterlevel_load.json was recorded: 17:54 ICT.
var fetched = time.Date(2026, 10, 29, 10, 54, 0, 0, time.UTC)

type recorded struct {
	obs     []normalize.Observation // sorted by station id
	history map[int64][]domain.Reading
}

func loadRecorded(t testing.TB) recorded {
	t.Helper()
	var load thaiwater.LoadResponse
	readJSON(t, "waterlevel_load.json", &load)
	refs := map[int64]normalize.GraphRefs{}
	history := map[int64][]domain.Reading{}
	for _, id := range []int64{737, 1097520, 575568, 3441, 12479} {
		var g thaiwater.GraphResponse
		readJSON(t, fmt.Sprintf("waterlevel_graph_%d.json", id), &g)
		r, readings, _ := normalize.Graph(id, &g)
		refs[id], history[id] = r, readings
	}
	res := normalize.Load(&load, refs)
	if len(res.Observations) < 30 {
		t.Fatalf("only %d observations from the fixture", len(res.Observations))
	}
	sort.Slice(res.Observations, func(i, j int) bool {
		return res.Observations[i].Station.ID < res.Observations[j].Station.ID
	})
	return recorded{obs: res.Observations, history: history}
}

func readJSON(t testing.TB, name string, v any) {
	t.Helper()
	b, err := os.ReadFile(fixtures + name)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, v); err != nil {
		t.Fatal(err)
	}
}

// input assembles what ingest and the API will pass: the seven hourly
// buckets ending at the reading's own hour, staleness and suspicion at now.
func input(o normalize.Observation, readings []domain.Reading, now time.Time) Input {
	r := o.Reading
	hourly := normalize.Hourly(readings, domain.HourEnd(r.At), trendHours+1)
	var prev domain.Point
	if len(hourly) > 1 {
		prev = hourly[len(hourly)-2]
	}
	return Input{
		Station: o.Station,
		Latest:  &r,
		Hourly:  hourly,
		Stale:   normalize.IsStale(r.At, now),
		Suspect: normalize.SuspectReasons(o.Station, r, prev, now),
		Now:     now,
	}
}

// tally counts as the /now summary will: current, believable readings only.
func tally(outs []Output, obs []normalize.Observation) Tally {
	var t Tally
	for i, out := range outs {
		if out.Stale || out.Suspect {
			continue
		}
		t.Current++
		switch out.Level {
		case domain.LevelOverflow:
			t.Overflow = append(t.Overflow, obs[i].Station.Province)
		case domain.LevelHigh:
			t.High++
		}
	}
	return t
}

func TestGoldenRecordedFixtures(t *testing.T) {
	rec := loadRecorded(t)
	clocks := []struct {
		name string
		now  time.Time
	}{
		{"as recorded, 17:54 ICT", fetched},
		{"six hours later, 23:54 ICT", fetched.Add(6 * time.Hour)},
		{"the next day, 17:54 ICT", fetched.Add(24 * time.Hour)},
	}
	var buf bytes.Buffer
	buf.WriteString("# humanize over testdata/waterlevel_load.json; regenerate with go test ./internal/humanize -update\n")
	for _, c := range clocks {
		fmt.Fprintf(&buf, "\n## %s\n", c.name)
		outs := make([]Output, len(rec.obs))
		for i, o := range rec.obs {
			in := input(o, rec.history[o.Station.ID], c.now)
			outs[i] = Describe(in)
			checkOutput(t, in, outs[i])
			flags := "current"
			switch {
			case outs[i].Suspect:
				flags = "suspect"
			case outs[i].Stale:
				flags = "stale"
			}
			fmt.Fprintf(&buf, "%-8s L%d %-7s %s\n", o.Station.Code, outs[i].Level, flags, outs[i].Sentence)
		}
		national := National(tally(outs, rec.obs))
		checkSentence(t, national)
		fmt.Fprintf(&buf, "national %s\n", national)
	}
	golden(t, "recorded.golden", buf.Bytes())
}

// The levels humanize reports for the recorded rows are ThaiWater's own
// published levels wherever ThaiWater publishes one (ADR 0003).
func TestRecordedLevelsMatchThePublishedOnes(t *testing.T) {
	rec := loadRecorded(t)
	checked := 0
	for _, o := range rec.obs {
		if o.Published.Level == nil {
			continue
		}
		out := Describe(Input{Station: o.Station, Latest: &o.Reading, Now: fetched})
		if int(out.Level) != *o.Published.Level {
			t.Errorf("%s: level %d, ThaiWater publishes %d", o.Station.Code, out.Level, *o.Published.Level)
		}
		checked++
	}
	if checked < 30 {
		t.Errorf("only %d rows had a published level", checked)
	}
}

// Every recorded row, at several clocks, with real and synthetic histories
// and every kind of suspicion, must stay inside the rules: no banned words,
// no stray formatting, sentence and level in agreement.
func TestBannedPhrasesAcrossFixtures(t *testing.T) {
	rec := loadRecorded(t)
	clocks := []time.Duration{0, 2 * time.Hour, 3*time.Hour + time.Minute, 4 * time.Hour, 30 * time.Hour, -48 * time.Hour}
	shapes := map[string]func(l float64, at time.Time) []domain.Reading{
		"none": func(float64, time.Time) []domain.Reading { return nil },
		"steady": func(l float64, at time.Time) []domain.Reading {
			return ramp(at, l, 0, 6, 1)
		},
		"rise-30cm": func(l float64, at time.Time) []domain.Reading { return ramp(at, l, 0.3, 6, 1) },
		"rise-2.5m": func(l float64, at time.Time) []domain.Reading { return ramp(at, l, 2.5, 6, 1) },
		"fall-4m":   func(l float64, at time.Time) []domain.Reading { return ramp(at, l, -4, 6, 1) },
		"gappy":     func(l float64, at time.Time) []domain.Reading { return ramp(at, l, 0.12, 6, 2) },
		"sparse":    func(l float64, at time.Time) []domain.Reading { return ramp(at, l, 0.5, 2, 1) },
	}
	forced := [][]domain.SuspectReason{
		nil,
		{domain.SuspectBelowGround},
		{domain.SuspectImplausiblePct, domain.SuspectJump},
		{domain.SuspectFuture},
		{domain.SuspectBelowGround, domain.SuspectImplausiblePct, domain.SuspectJump, domain.SuspectFuture},
	}
	n := 0
	for _, d := range clocks {
		now := fetched.Add(d)
		outs := make([]Output, len(rec.obs))
		for i, o := range rec.obs {
			histories := map[string][]domain.Reading{"recorded": rec.history[o.Station.ID]}
			for name, shape := range shapes {
				histories[name] = shape(o.Reading.WL, o.Reading.At)
			}
			for _, readings := range histories {
				for _, rs := range forced {
					in := input(o, readings, now)
					if rs != nil {
						in.Suspect = rs
					}
					out := Describe(in)
					checkOutput(t, in, out)
					n++
					if rs == nil && readings == nil {
						outs[i] = out
					}
				}
			}
		}
		checkSentence(t, National(tally(outs, rec.obs)))
	}
	// National sentences for every prefix of the recorded provinces, so that
	// 0, 1, a few and many overflowing stations all appear.
	var provinces []string
	for _, o := range rec.obs {
		provinces = append(provinces, o.Station.Province)
	}
	for k := 0; k <= len(provinces); k++ {
		for _, high := range []int{0, 1, 2, 145} {
			checkSentence(t, National(Tally{Current: len(provinces), High: high, Overflow: provinces[:k]}))
			n++
		}
	}
	t.Logf("%d outputs checked", n)
}

// ramp is `hours` hourly readings before `at`, every `step` hours, ending
// `rise` metres below the latest level l.
func ramp(at time.Time, l, rise float64, hours, step int) []domain.Reading {
	var out []domain.Reading
	for i := hours; i >= 1; i -= step {
		out = append(out, domain.Reading{At: at.Add(-time.Duration(i) * time.Hour), WL: l - rise*float64(i)/float64(hours)})
	}
	return out
}

// FuzzDescribe throws arbitrary references, levels, histories, ages and
// suspicions at Describe. Values are whole millimetres, as normalize
// produces them.
func FuzzDescribe(f *testing.F) {
	f.Add(int32(8600), int32(10000), int32(2000), uint8(domain.RefsComplete), int32(120), uint8(0x3f), uint16(30), uint8(0))
	f.Add(int32(10350), int32(10000), int32(2000), uint8(domain.RefsComplete), int32(-1200), uint8(0x15), uint16(300), uint8(0))
	f.Add(int32(11010), int32(10900), int32(0), uint8(domain.RefsBankOnly), int32(40), uint8(0x3f), uint16(10), uint8(0))
	f.Add(int32(331730), int32(0), int32(0), uint8(domain.RefsZeroCoded), int32(0), uint8(0), uint16(2000), uint8(0))
	f.Add(int32(-30000), int32(10000), int32(2000), uint8(domain.RefsComplete), int32(5000), uint8(0x3f), uint16(0), uint8(0x0f))
	f.Add(int32(10004), int32(10000), int32(2000), uint8(domain.RefsComplete), int32(49), uint8(0x07), uint16(181), uint8(0))
	f.Fuzz(func(t *testing.T, wlMM, bankMM, groundMM int32, refs uint8, deltaMM int32, mask uint8, ageMin uint16, suspectMask uint8) {
		st := domain.Station{
			Refs:     domain.Refs(refs % 5),
			BankFull: float64(bankMM) / 1000,
			Ground:   float64(groundMM) / 1000,
			Province: "ชุมพร",
		}
		if st.Refs == domain.RefsComplete && groundMM >= bankMM {
			st.Refs = domain.RefsInverted // normalize never produces this as complete
		}
		at := time.Date(2026, 10, 29, 7, 0, 0, 0, time.UTC)
		now := at.Add(time.Duration(ageMin) * time.Minute)
		latest := domain.Reading{At: at, WL: float64(wlMM) / 1000}
		in := Input{Station: st, Latest: &latest, Stale: normalize.IsStale(at, now), Now: now}
		for i := 1; i <= trendHours; i++ {
			if mask&(1<<(i-1)) == 0 {
				continue
			}
			wl := float64(int64(wlMM)-int64(deltaMM)*int64(i)/trendHours) / 1000
			in.Hourly = append(in.Hourly, domain.Point{At: at.Add(-time.Duration(i) * time.Hour), WL: wl, OK: true})
		}
		for r := domain.SuspectBelowGround; r <= domain.SuspectFuture; r++ {
			if suspectMask&(1<<(r-1)) != 0 {
				in.Suspect = append(in.Suspect, r)
			}
		}
		out := Describe(in)
		checkOutput(t, in, out)
		if out.Trend != nil && math.Abs(float64(out.Trend.DeltaMM)) > math.Abs(float64(deltaMM))+1 {
			t.Errorf("trend %+v larger than the history's %d mm", *out.Trend, deltaMM)
		}
	})
}

// golden compares got with testdata/name, or rewrites it under -update.
func golden(t *testing.T, name string, got []byte) {
	t.Helper()
	path := filepath.Join("testdata", name)
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%v (run go test ./internal/humanize -update to create it)", err)
	}
	if bytes.Equal(got, want) {
		return
	}
	g, w := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	for i := 0; i < len(g) || i < len(w); i++ {
		var gl, wl string
		if i < len(g) {
			gl = g[i]
		}
		if i < len(w) {
			wl = w[i]
		}
		if gl != wl {
			t.Fatalf("%s differs at line %d\n got %s\nwant %s\n(if the change is intended, review it and run with -update)", path, i+1, gl, wl)
		}
	}
}
