package normalize

import (
	"encoding/json"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/serlorsx-xln/mahanathi/server/internal/domain"
	"github.com/serlorsx-xln/mahanathi/server/internal/thaiwater"
)

const testdata = "../thaiwater/testdata/"

func loadFixture(t *testing.T) *thaiwater.LoadResponse {
	t.Helper()
	b, err := os.ReadFile(testdata + "waterlevel_load.json")
	if err != nil {
		t.Fatal(err)
	}
	var r thaiwater.LoadResponse
	if err := json.Unmarshal(b, &r); err != nil {
		t.Fatal(err)
	}
	return &r
}

func graphFixture(t *testing.T, id string) *thaiwater.GraphResponse {
	t.Helper()
	b, err := os.ReadFile(testdata + "waterlevel_graph_" + id + ".json")
	if err != nil {
		t.Fatal(err)
	}
	var g thaiwater.GraphResponse
	if err := json.Unmarshal(b, &g); err != nil {
		t.Fatal(err)
	}
	return &g
}

// allGraphRefs returns the full-precision references from every recorded
// graph response, keyed by station id.
func allGraphRefs(t *testing.T) map[int64]GraphRefs {
	t.Helper()
	out := map[int64]GraphRefs{}
	for id, name := range map[int64]string{737: "737", 1097520: "1097520", 575568: "575568", 3441: "3441", 12479: "12479"} {
		refs, _, _ := Graph(id, graphFixture(t, name))
		out[id] = refs
	}
	return out
}

func f(v float64) *float64 { return &v }

func rule(op, term string, level int) thaiwater.Rule {
	return thaiwater.Rule{Operator: op, Term: term, Level: level}
}

func near(a, b, tol float64) bool { return math.Abs(a-b) <= tol }

func TestParseICT(t *testing.T) {
	got, err := ParseICT("2026-10-29 17:40")
	if err != nil {
		t.Fatal(err)
	}
	if want := time.Date(2026, 10, 29, 10, 40, 0, 0, time.UTC); !got.Equal(want) || got.Location() != time.UTC {
		t.Errorf("ParseICT = %v, want %v in UTC", got, want)
	}
	// Midnight in Bangkok is the previous day in UTC.
	got, _ = ParseICT("2026-10-26 00:00")
	if want := time.Date(2026, 10, 25, 17, 0, 0, 0, time.UTC); !got.Equal(want) {
		t.Errorf("ParseICT(midnight) = %v, want %v", got, want)
	}
	for _, bad := range []string{"", "2026-10-29", "2026-10-29T17:40", "2026-10-29 17:40:00", "2026-13-01 00:00", "29/10/2026 17:40", "2026-10-29 24:00"} {
		if _, err := ParseICT(bad); err == nil {
			t.Errorf("ParseICT(%q) accepted", bad)
		}
	}
}

func TestLoadFixtureCases(t *testing.T) {
	res := Load(loadFixture(t), nil)
	if len(res.Observations) != 37 || res.Dropped != 0 || res.Rows != 37 {
		t.Fatalf("observations %d, dropped %d, rows %d; want 37, 0, 37", len(res.Observations), res.Dropped, res.Rows)
	}
	byCode := map[string]Observation{}
	for _, o := range res.Observations {
		byCode[o.Station.Code] = o
	}

	cases := []struct {
		code       string
		why        string
		refs       domain.Refs
		bank, gnd  float64
		pct        float64 // NaN when no percent
		level      domain.Level
		distance   float64 // NaN when no distance
		measuredAt string  // UTC
	}{
		{"THA006", "HII, every field present", domain.RefsComplete, 2.14, -2.065, 124.495, domain.LevelOverflow, 1.03, "2026-10-29T10:40:00Z"},
		{"Gt.11", "the brief's example; side banks 44 m off are ignored", domain.RefsComplete, 40.6, 37.72, 84.722, domain.LevelHigh, -0.44, "2026-10-29T10:00:00Z"},
		{"P.64", "side banks 787 m off are ignored", domain.RefsComplete, 793.35, 787.025, 18.735, domain.LevelLow, -5.14, "2026-10-28T23:00:00Z"},
		{"FOP045", "water below the ground reference", domain.RefsComplete, 25.13, 5.95, -0.782, domain.LevelCriticallyLow, -19.33, "2026-10-29T10:40:00Z"},
		{"URTD03", "most negative percent", domain.RefsComplete, 149.57, 139.867, -18.932, domain.LevelCriticallyLow, -11.54, "2026-10-29T10:00:00Z"},
		{"URTU09", "flips to level 3 at 2 decimals", domain.RefsComplete, 225.39, 221.824, 69.994, domain.LevelNormal, -1.07, "2026-10-29T10:00:00Z"},
		{"BKC004", "two days old", domain.RefsComplete, 0.91, -1.63, 105.512, domain.LevelOverflow, 0.14, "2026-10-27T12:40:00Z"},
		{"FOP038", "4-decimal ground is rounded to mm", domain.RefsComplete, 3.94, 0.289, 69.871, domain.LevelNormal, -1.10, "2026-10-29T10:40:00Z"},
		{"Gt.9", "ground absent: distance only", domain.RefsBankOnly, 10.9, math.NaN(), math.NaN(), domain.LevelUnknown, 0.11, "2026-10-29T10:00:00Z"},
		{"I.17", "ground absent: distance only", domain.RefsBankOnly, 389.77, math.NaN(), math.NaN(), domain.LevelUnknown, -2.24, "2026-10-28T23:00:00Z"},
		{"BLGTU06", "both coded 0 at a dam", domain.RefsZeroCoded, math.NaN(), math.NaN(), math.NaN(), domain.LevelUnknown, math.NaN(), "2026-10-29T10:00:00Z"},
		{"MKSND01", "both coded 0 at a dam", domain.RefsZeroCoded, math.NaN(), math.NaN(), math.NaN(), domain.LevelUnknown, math.NaN(), "2026-10-29T10:00:00Z"},
	}
	for _, c := range cases {
		t.Run(c.code, func(t *testing.T) {
			o, ok := byCode[c.code]
			if !ok {
				t.Fatalf("%s missing", c.code)
			}
			s := o.Station
			if s.Refs != c.refs {
				t.Fatalf("refs = %v, want %v (%s)", s.Refs, c.refs, c.why)
			}
			if !math.IsNaN(c.bank) && s.BankFull != c.bank {
				t.Errorf("bankFull = %v, want %v", s.BankFull, c.bank)
			}
			if !math.IsNaN(c.gnd) && s.Ground != c.gnd {
				t.Errorf("ground = %v, want %v", s.Ground, c.gnd)
			}
			pct, hasPct := s.Pct(o.Reading.WL)
			if math.IsNaN(c.pct) == hasPct {
				t.Errorf("has percent = %v, want %v", hasPct, !math.IsNaN(c.pct))
			} else if hasPct && !near(pct, c.pct, 0.001) {
				t.Errorf("pct = %.4f, want %.3f", pct, c.pct)
			}
			if lv := s.Level(o.Reading.WL); lv != c.level {
				t.Errorf("level = %d, want %d", lv, c.level)
			}
			d, hasD := s.Distance(o.Reading.WL)
			if math.IsNaN(c.distance) == hasD {
				t.Errorf("has distance = %v, want %v", hasD, !math.IsNaN(c.distance))
			} else if hasD && !near(d, c.distance, 1e-9) {
				t.Errorf("distance = %v, want %v", d, c.distance)
			}
			if got := o.Reading.At.Format(time.RFC3339); got != c.measuredAt {
				t.Errorf("measuredAt = %s, want %s", got, c.measuredAt)
			}
			if o.Reading.StationID != s.ID {
				t.Errorf("reading station %d != station %d", o.Reading.StationID, s.ID)
			}
		})
	}

	tha := byCode["THA006"].Station
	if tha.ID != 737 || tha.Name != "วัดท่าเจดีย์ (TTC06)" || tha.Province != "สุพรรณบุรี" || tha.ProvinceCode != "72" ||
		tha.Amphoe != "สองพี่น้อง" || tha.Tumbon != "บางตาเถร" || tha.Basin != "ลุ่มน้ำท่าจีน" || tha.River != "แม่น้ำท่าจีน" ||
		tha.Agency != "HII" || tha.Lat != 14.15657 || tha.Lon != 100.12744 {
		t.Errorf("THA006 station = %+v", tha)
	}
	if byCode["C.67"].Station.River != "" {
		t.Error("C.67 has no river_name key but got a river")
	}
	if o := byCode["THA006"]; o.LoadRefs.BankFull == nil || *o.LoadRefs.BankFull != 2.14 || o.Published.Level == nil || *o.Published.Level != 5 {
		t.Errorf("THA006 load refs or oracle = %+v %+v", o.LoadRefs, o.Published)
	}
	if o := byCode["Gt.9"]; o.LoadRefs.Ground != nil || o.Published.Level != nil {
		t.Errorf("Gt.9 load refs or oracle = %+v %+v", o.LoadRefs, o.Published)
	}
}

// TestOracleAgreesWithThaiWater is the test ADR 0003 promises: with
// references at full precision, our level equals the published
// situation_level on every row, and the rows without a published level are
// exactly the rows where we refuse to compute one.
func TestOracleAgreesWithThaiWater(t *testing.T) {
	resp := loadFixture(t)

	check := func(graph map[int64]GraphRefs) (mismatch []string, worstPP float64) {
		for _, o := range Load(resp, graph).Observations {
			ours := o.Station.Level(o.Reading.WL)
			pub := o.Published.Level
			if (pub == nil) != (ours == domain.LevelUnknown) {
				t.Errorf("%s: published level %v but ours %d", o.Station.Code, pub, ours)
				continue
			}
			if pub == nil {
				continue
			}
			if domain.Level(*pub) != ours {
				mismatch = append(mismatch, o.Station.Code)
			}
			if pct, ok := o.Station.Pct(o.Reading.WL); ok && o.Published.Percent != nil {
				worstPP = math.Max(worstPP, math.Abs(pct-*o.Published.Percent))
			}
		}
		return mismatch, worstPP
	}

	mismatch, worst := check(nil)
	if strings.Join(mismatch, ",") != "URTU09" {
		t.Errorf("with 2-decimal references, mismatches = %v, want exactly [URTU09]", mismatch)
	}
	if worst > 1 {
		t.Errorf("worst percent difference %.3f pp, want within 1 pp", worst)
	}

	mismatch, worst = check(allGraphRefs(t))
	if len(mismatch) != 0 {
		t.Errorf("with full-precision references, mismatches = %v, want none", mismatch)
	}
	if worst > 0.5 {
		t.Errorf("worst percent difference %.3f pp with full precision", worst)
	}
}

func TestResolveRefs(t *testing.T) {
	cases := []struct {
		name      string
		load      LoadRefs
		graph     *GraphRefs
		want      domain.Refs
		bank, gnd float64
		warn      bool
	}{
		{"load only", LoadRefs{f(40.6), f(37.72)}, nil, domain.RefsComplete, 40.6, 37.72, false},
		{"graph wins at full precision", LoadRefs{f(225.39), f(221.824)}, &GraphRefs{f(225.386), f(221.824)}, domain.RefsComplete, 225.386, 221.824, false},
		{"graph wins and a real difference warns", LoadRefs{f(2.0), f(-2.065)}, &GraphRefs{f(2.139), f(-2.065)}, domain.RefsComplete, 2.139, -2.065, true},
		{"graph fills an absent load ground", LoadRefs{f(10.9), nil}, &GraphRefs{f(10.9), f(8.0)}, domain.RefsComplete, 10.9, 8.0, false},
		{"graph without refs keeps load", LoadRefs{f(10.9), nil}, &GraphRefs{nil, nil}, domain.RefsBankOnly, 10.9, 0, false},
		{"ground absent", LoadRefs{f(10.9), nil}, nil, domain.RefsBankOnly, 10.9, 0, false},
		{"bank absent", LoadRefs{nil, f(3)}, nil, domain.RefsMissing, 0, 0, false},
		{"both absent", LoadRefs{}, nil, domain.RefsMissing, 0, 0, false},
		{"zero-coded", LoadRefs{f(0), f(0)}, nil, domain.RefsZeroCoded, 0, 0, false},
		{"bank 0 with no ground is treated as zero-coded", LoadRefs{f(0), nil}, nil, domain.RefsZeroCoded, 0, 0, false},
		{"bank 0 with a real ground below it is usable", LoadRefs{f(0), f(-3.5)}, nil, domain.RefsComplete, 0, -3.5, false},
		{"ground equal to bank", LoadRefs{f(5), f(5)}, nil, domain.RefsInverted, 0, 0, false},
		{"ground above bank", LoadRefs{f(5), f(6)}, nil, domain.RefsInverted, 0, 0, false},
		{"values rounded to mm", LoadRefs{f(40.599998), f(37.7200004)}, nil, domain.RefsComplete, 40.6, 37.72, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			refs, bank, gnd, warn := ResolveRefs(c.load, c.graph)
			if refs != c.want || bank != c.bank || gnd != c.gnd {
				t.Errorf("ResolveRefs = %v %v %v, want %v %v %v", refs, bank, gnd, c.want, c.bank, c.gnd)
			}
			if (warn != "") != c.warn {
				t.Errorf("warning = %q, want warning=%v", warn, c.warn)
			}
		})
	}
}

func TestLoadUsesGraphRefsAndWarnsOnRealDifferences(t *testing.T) {
	resp := loadFixture(t)
	graph := map[int64]GraphRefs{
		3441: {f(225.386), f(221.824)}, // URTU09 at full precision: no warning
		2659: {f(10.9), f(9.0)},        // Gt.9 gains a ground from history
		737:  {f(2.5), f(-2.065)},      // 0.36 m away from the load's 2.14: warn
	}
	res := Load(resp, graph)
	byCode := map[string]Observation{}
	for _, o := range res.Observations {
		byCode[o.Station.Code] = o
	}
	if lv := byCode["URTU09"].Station.Level(byCode["URTU09"].Reading.WL); lv != domain.LevelHigh {
		t.Errorf("URTU09 level = %d with full-precision refs, want 4", lv)
	}
	if byCode["Gt.9"].Station.Refs != domain.RefsComplete {
		t.Errorf("Gt.9 refs = %v, want complete once history supplies a ground", byCode["Gt.9"].Station.Refs)
	}
	var warned []int64
	for _, w := range res.Warnings {
		if w.Kind == WarnRefMismatch {
			warned = append(warned, w.StationID)
		}
	}
	if len(warned) != 1 || warned[0] != 737 {
		t.Errorf("mismatch warnings for %v, want only 737", warned)
	}
}

func TestLoadDropsRowsItCannotPlaceOrDate(t *testing.T) {
	resp := loadFixture(t)
	rows := resp.WaterlevelData.Data[:6]
	mod := make([]thaiwater.LoadRow, len(rows))
	copy(mod, rows)
	mod[0].Datetime = "2026-10-29T17:40"                            // bad time
	mod[1].WaterlevelMSL = thaiwater.Num{Present: true}             // null level
	mod[2].Station.Lat = thaiwater.Num{Present: true}               // no position
	mod[3].Station.Lon = thaiwater.Num{Value: 2, Valid: true}       // far outside the map
	mod[4].Station.MinBank = thaiwater.Num{Present: true, Raw: "-"} // junk reference: kept, refs missing
	dup := mod[5]
	dup.Datetime = "2026-10-29 18:00" // a later copy of row 5 wins
	dup.WaterlevelMSL = thaiwater.Num{Value: 9.99, Valid: true, Present: true}
	mod = append(mod, dup)
	resp.WaterlevelData.Data = mod

	res := Load(resp, nil)
	if res.Rows != 7 || res.Dropped != 4 || len(res.Observations) != 2 {
		t.Fatalf("rows %d dropped %d kept %d; want 7, 4, 2", res.Rows, res.Dropped, len(res.Observations))
	}
	kinds := map[WarningKind]int{}
	for _, w := range res.Warnings {
		kinds[w.Kind]++
		if w.Detail == "" {
			t.Errorf("warning %v has no detail", w.Kind)
		}
	}
	for k, n := range map[WarningKind]int{WarnBadTime: 1, WarnNoWaterLevel: 1, WarnBadPosition: 2, WarnJunkValue: 1, WarnDuplicate: 1} {
		if kinds[k] != n {
			t.Errorf("%s warnings = %d, want %d (all: %v)", k, kinds[k], n, kinds)
		}
	}
	for _, o := range res.Observations {
		if o.Station.ID == rows[4].Station.ID && o.Station.Refs != domain.RefsMissing {
			t.Errorf("junk min_bank gave refs %v", o.Station.Refs)
		}
		if o.Station.ID == rows[5].Station.ID && o.Reading.WL != 9.99 {
			t.Errorf("duplicate kept the older row (wl %v)", o.Reading.WL)
		}
	}
}

func TestStationNamesFallBackToEnglish(t *testing.T) {
	resp := loadFixture(t)
	resp.WaterlevelData.Data = resp.WaterlevelData.Data[:1]
	r := &resp.WaterlevelData.Data[0]
	r.Station.Name.TH = "  "
	r.Agency.Short.EN = ""
	o := Load(resp, nil).Observations[0]
	if o.Station.Name != "Tha Chedi Temple (TTC06)" {
		t.Errorf("name = %q, want the English fallback", o.Station.Name)
	}
	if o.Station.Agency != "สสน." {
		t.Errorf("agency = %q, want the Thai short name as fallback", o.Station.Agency)
	}
}

func TestGraphParsesRefsAndReadings(t *testing.T) {
	refs, readings, warns := Graph(1097520, graphFixture(t, "1097520"))
	if refs.BankFull == nil || *refs.BankFull != 40.6 || refs.Ground == nil || *refs.Ground != 37.72 {
		t.Errorf("refs = %v %v, want 40.6 and 37.72 after rounding float32 noise", refs.BankFull, refs.Ground)
	}
	if len(warns) != 0 {
		t.Errorf("warnings = %v", warns)
	}
	nonNull := 0
	for _, p := range graphFixture(t, "1097520").Data.GraphData {
		if p.Value.Valid {
			nonNull++
		}
	}
	if len(readings) != nonNull {
		t.Fatalf("readings = %d, want the %d non-null points", len(readings), nonNull)
	}
	first := readings[0]
	if first.StationID != 1097520 || first.WL != 38.27 || !first.At.Equal(time.Date(2026, 10, 25, 18, 0, 0, 0, time.UTC)) {
		t.Errorf("first reading = %+v, want 38.27 at 2026-10-26 01:00 ICT", first)
	}
	for i := 1; i < len(readings); i++ {
		if !readings[i].At.After(readings[i-1].At) {
			t.Fatalf("readings not strictly ascending at %d", i)
		}
	}

	// BKC004 stopped reporting: most points are null and must simply be absent.
	_, readings, _ = Graph(575568, graphFixture(t, "575568"))
	if len(readings) != 540-497 {
		t.Errorf("BKC004 readings = %d, want 43", len(readings))
	}
}

func TestGraphSkipsBadTimesAndDuplicates(t *testing.T) {
	g := &thaiwater.GraphResponse{Result: "OK"}
	v := func(x float64) thaiwater.Num { return thaiwater.Num{Value: x, Valid: true, Present: true} }
	g.Data.GraphData = []thaiwater.GraphPoint{
		{Datetime: "2026-10-29 10:10", Value: v(1.01)},
		{Datetime: "bad", Value: v(1.02)},
		{Datetime: "2026-10-29 10:00", Value: v(1.00)},
		{Datetime: "2026-10-29 10:10", Value: v(1.05)}, // same instant again: last one wins
	}
	refs, readings, warns := Graph(1, g)
	if refs.BankFull != nil || refs.Ground != nil {
		t.Errorf("refs = %+v from a response without them", refs)
	}
	if len(readings) != 2 || readings[0].WL != 1.00 || readings[1].WL != 1.05 {
		t.Errorf("readings = %+v", readings)
	}
	if len(warns) != 1 || warns[0].Kind != WarnBadTime {
		t.Errorf("warnings = %+v, want one bad_time", warns)
	}
}

func TestHourlyTakesTheLastReadingInEachHour(t *testing.T) {
	// วัดท่าเจดีย์ reports every 10 minutes. 2026-10-26 01:00 ICT is
	// 2026-10-25 18:00 UTC.
	_, readings, _ := Graph(737, graphFixture(t, "737"))
	end := time.Date(2026, 10, 25, 20, 0, 0, 0, time.UTC)
	pts := Hourly(readings, end, 3)
	if len(pts) != 3 {
		t.Fatalf("points = %d", len(pts))
	}
	// Each bucket (HH−1:00, HH:00] ends on the hour reading itself when
	// present, so bucket 19:00 UTC is the 02:00 ICT value.
	byICT := map[string]float64{}
	for _, p := range graphFixture(t, "737").Data.GraphData {
		if p.Value.Valid {
			byICT[p.Datetime] = p.Value.Value
		}
	}
	for i, ict := range []string{"2026-10-26 01:00", "2026-10-26 02:00", "2026-10-26 03:00"} {
		want, ok := byICT[ict]
		if !ok {
			t.Fatalf("fixture lacks %s", ict)
		}
		if !pts[i].OK || pts[i].WL != math.Round(want*1000)/1000 {
			t.Errorf("bucket %s = %+v, want %v", ict, pts[i], want)
		}
		if !pts[i].At.Equal(end.Add(time.Duration(i-2) * time.Hour)) {
			t.Errorf("bucket %d At = %v", i, pts[i].At)
		}
	}
}

func TestHourlyGapsOrderAndBounds(t *testing.T) {
	base := time.Date(2026, 10, 29, 0, 0, 0, 0, time.UTC)
	r := func(min int, wl float64) domain.Reading {
		return domain.Reading{StationID: 1, At: base.Add(time.Duration(min) * time.Minute), WL: wl}
	}
	readings := []domain.Reading{
		r(170, 3.0), // 02:50 → bucket 03:00
		r(60, 1.0),  // exactly 01:00 → bucket 01:00, not 02:00
		r(130, 2.5), // 02:10 → bucket 03:00, but earlier than 02:50
		r(-30, 9.9), // before the window
		r(400, 9.9), // after the window
	}
	pts := Hourly(readings, base.Add(4*time.Hour), 4) // buckets 01:00 … 04:00
	want := []struct {
		ok bool
		wl float64
	}{{true, 1.0}, {false, 0}, {true, 3.0}, {false, 0}}
	for i, w := range want {
		if pts[i].OK != w.ok || (w.ok && pts[i].WL != w.wl) {
			t.Errorf("bucket %d = %+v, want %+v", i, pts[i], w)
		}
	}
	if pts := Hourly(readings, base.Add(4*time.Hour+15*time.Minute), 2); !pts[1].At.Equal(base.Add(4 * time.Hour)) {
		t.Errorf("an end off the hour was not truncated: %v", pts[1].At)
	}
	if len(Hourly(nil, base, 0)) != 0 {
		t.Error("zero hours gave points")
	}
}

func TestIsStale(t *testing.T) {
	now := time.Date(2026, 10, 29, 10, 54, 0, 0, time.UTC) // 17:54 ICT
	cases := []struct {
		age  time.Duration
		want bool
	}{
		{0, false},
		{3 * time.Hour, false}, // "older than 3 h"
		{3*time.Hour + time.Second, true},
		{3*time.Hour + 24*time.Minute, true}, // AIT009 at 14:30
		{-5 * time.Minute, false},
	}
	for _, c := range cases {
		if got := IsStale(now.Add(-c.age), now); got != c.want {
			t.Errorf("IsStale(age %v) = %v, want %v", c.age, got, c.want)
		}
	}
}

func TestSuspectReasons(t *testing.T) {
	now := time.Date(2026, 10, 29, 11, 0, 0, 0, time.UTC)
	complete := domain.Station{Refs: domain.RefsComplete, BankFull: 40.6, Ground: 37.72}
	bankOnly := domain.Station{Refs: domain.RefsBankOnly, BankFull: 10.9}
	at := now.Add(-time.Hour)
	prev := func(wl float64) domain.Point { return domain.Point{At: at.Add(-time.Hour), WL: wl, OK: true} }
	none := domain.Point{}

	cases := []struct {
		name string
		st   domain.Station
		wl   float64
		at   time.Time
		prev domain.Point
		want []domain.SuspectReason
	}{
		{"ordinary", complete, 40.16, at, prev(40.20), nil},
		{"20 m below ground is still believable", complete, 17.72, at, none, nil},
		{"more than 20 m below ground", complete, 17.71, at, none, []domain.SuspectReason{domain.SuspectBelowGround}},
		{"pct of exactly 400 is allowed", complete, 37.72 + 4*2.88, at, none, nil},
		{"pct above 400", complete, 37.72 + 4*2.88 + 0.01, at, none, []domain.SuspectReason{domain.SuspectImplausiblePct}},
		{"a 3 m jump in an hour is allowed", complete, 43.0, at, prev(40.0), nil},
		{"a jump over 3 m in an hour", complete, 40.0, at, prev(43.01), []domain.SuspectReason{domain.SuspectJump}},
		{"no previous hour, no jump check", complete, 40.0, at, none, nil},
		{"timestamp ten minutes ahead is clock skew", complete, 40.16, now.Add(10 * time.Minute), none, nil},
		{"timestamp in the future", complete, 40.16, now.Add(11 * time.Minute), none, []domain.SuspectReason{domain.SuspectFuture}},
		{"bank-only has no ground or percent check", bankOnly, -500, at, none, nil},
		{"several at once", complete, 10.0, now.Add(time.Hour), prev(40.0), []domain.SuspectReason{domain.SuspectBelowGround, domain.SuspectJump, domain.SuspectFuture}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := SuspectReasons(c.st, domain.Reading{At: c.at, WL: c.wl}, c.prev, now)
			if len(got) != len(c.want) {
				t.Fatalf("SuspectReasons = %v, want %v", got, c.want)
			}
			for i := range got {
				if got[i] != c.want[i] {
					t.Fatalf("SuspectReasons = %v, want %v", got, c.want)
				}
			}
		})
	}
}

func TestCheckRuleWeb(t *testing.T) {
	if err := CheckRuleWeb(loadFixture(t).Scale.Data.RuleWeb); err != nil {
		t.Fatalf("the recorded rule_web was rejected: %v", err)
	}
	known := []thaiwater.Rule{rule(">", "100", 5), rule(">", "70", 4), rule(">", "30", 3), rule(">", "10", 2), rule("<=", "10", 1)}
	changed := func(mut func([]thaiwater.Rule) []thaiwater.Rule) []thaiwater.Rule {
		c := make([]thaiwater.Rule, len(known))
		copy(c, known)
		return mut(c)
	}
	bad := map[string][]thaiwater.Rule{
		"empty":            nil,
		"threshold moved":  changed(func(r []thaiwater.Rule) []thaiwater.Rule { r[1].Term = "80"; return r }),
		"operator changed": changed(func(r []thaiwater.Rule) []thaiwater.Rule { r[0].Operator = ">="; return r }),
		"level renumbered": changed(func(r []thaiwater.Rule) []thaiwater.Rule { r[2].Level = 2; return r }),
		"step removed":     changed(func(r []thaiwater.Rule) []thaiwater.Rule { return r[:4] }),
		"step added":       changed(func(r []thaiwater.Rule) []thaiwater.Rule { return append(r, rule(">", "150", 6)) }),
		"term not numeric": changed(func(r []thaiwater.Rule) []thaiwater.Rule { r[3].Term = "ten"; return r }),
	}
	for name, rules := range bad {
		if err := CheckRuleWeb(rules); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	// Order and spacing are not meaningful; values are.
	shuffled := []thaiwater.Rule{rule("<=", "10", 1), rule(" > ", "100.0", 5), rule(">", "30", 3), rule(">", "70", 4), rule(">", "10", 2)}
	if err := CheckRuleWeb(shuffled); err != nil {
		t.Errorf("reordered rules rejected: %v", err)
	}
}
