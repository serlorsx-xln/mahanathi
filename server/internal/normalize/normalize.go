package normalize

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/serlorsx-xln/mahanathi/server/internal/domain"
	"github.com/serlorsx-xln/mahanathi/server/internal/thaiwater"
)

// ICT is Indochina Time, the zone of every ThaiWater timestamp. Thailand has
// no daylight saving, so a fixed offset is exact.
var ICT = time.FixedZone("ICT", 7*3600)

// sourceLayout is how ThaiWater writes times: local ICT, minutes, no offset.
const sourceLayout = "2006-01-02 15:04"

// The map extent shared with the geodata pipeline (docs/research/
// geodata-sources.md). A station outside it cannot be drawn, so it is
// dropped with a warning.
const (
	minLon, maxLon = 92.0, 111.0
	minLat, maxLat = 3.0, 24.0
)

// refMismatchToleranceMM is how far, in whole millimetres, the graph and
// load references may differ before it is more than rounding: the load gives
// 2 decimals, so up to 5 mm is expected. Comparing metres as floats instead
// flags 0.905 against 0.91, whose difference is 0.0050000000000000044.
const refMismatchToleranceMM = 5

// refsDiffer reports whether two references differ by more than rounding.
func refsDiffer(a, b float64) bool {
	d := domain.MM(a) - domain.MM(b)
	return d > refMismatchToleranceMM || d < -refMismatchToleranceMM
}

// WarningKind classifies a data problem found while normalising.
type WarningKind string

// The warning kinds. Each one is counted per ingest run and in
// docs/research/data-quality.md.
const (
	WarnBadTime      WarningKind = "bad_time"           // row dropped
	WarnNoWaterLevel WarningKind = "no_water_level"     // row dropped
	WarnBadPosition  WarningKind = "bad_position"       // row dropped
	WarnDuplicate    WarningKind = "duplicate_station"  // older copy dropped
	WarnJunkValue    WarningKind = "junk_value"         // value treated as absent
	WarnRefMismatch  WarningKind = "graph_ref_mismatch" // graph reference used anyway
)

// Warning is one data problem, attributable to a station.
type Warning struct {
	StationID int64
	Code      string
	Kind      WarningKind
	Detail    string
}

// LoadRefs are a station's references as the load endpoint gave them, after
// rounding to millimetres. nil means absent, null or unparsable.
type LoadRefs struct {
	BankFull *float64
	Ground   *float64
}

// GraphRefs are the full-precision references from a waterlevel_graph
// response, rounded to millimetres. nil means the response had none.
type GraphRefs struct {
	BankFull *float64
	Ground   *float64
}

// Published holds ThaiWater's own derived values, kept only as a test and
// monitoring oracle. Product values never come from here.
type Published struct {
	Level   *int
	Percent *float64
}

// Observation is one usable row of the national snapshot.
type Observation struct {
	Station   domain.Station
	Reading   domain.Reading
	LoadRefs  LoadRefs
	Published Published
}

// Result is a normalised snapshot.
type Result struct {
	Observations []Observation // one per station, in source order
	Warnings     []Warning
	Rows         int // rows received
	Dropped      int // rows that could not be used at all
}

// ParseICT reads a ThaiWater timestamp ("2026-10-29 17:40", ICT) and returns
// the UTC instant.
func ParseICT(s string) (time.Time, error) {
	t, err := time.ParseInLocation(sourceLayout, s, ICT)
	if err != nil {
		return time.Time{}, fmt.Errorf("normalize: time %q is not %q in ICT", s, sourceLayout)
	}
	return t.UTC(), nil
}

// Load turns the national snapshot into observations. graph holds the
// full-precision references already fetched for some stations; they win over
// the load's 2-decimal values (ADR 0003). Rows without a usable time, water
// level or position are dropped and counted, never guessed.
func Load(resp *thaiwater.LoadResponse, graph map[int64]GraphRefs) Result {
	rows := resp.WaterlevelData.Data
	res := Result{Rows: len(rows)}
	index := map[int64]int{} // station id → position in res.Observations

	for _, row := range rows {
		obs, warns, ok := observe(row, graph)
		res.Warnings = append(res.Warnings, warns...)
		if !ok {
			res.Dropped++
			continue
		}
		if i, seen := index[obs.Station.ID]; seen {
			older := res.Observations[i]
			res.Warnings = append(res.Warnings, Warning{
				StationID: obs.Station.ID, Code: obs.Station.Code, Kind: WarnDuplicate,
				Detail: fmt.Sprintf("two rows for one station (%s and %s); keeping the later", older.Reading.At.Format(time.RFC3339), obs.Reading.At.Format(time.RFC3339)),
			})
			if obs.Reading.At.After(older.Reading.At) {
				res.Observations[i] = obs
			}
			continue
		}
		index[obs.Station.ID] = len(res.Observations)
		res.Observations = append(res.Observations, obs)
	}
	return res
}

func observe(row thaiwater.LoadRow, graph map[int64]GraphRefs) (Observation, []Warning, bool) {
	st := row.Station
	code := strings.TrimSpace(st.OldCode)
	var warns []Warning
	warn := func(kind WarningKind, format string, args ...any) {
		warns = append(warns, Warning{StationID: st.ID, Code: code, Kind: kind, Detail: fmt.Sprintf(format, args...)})
	}

	at, err := ParseICT(row.Datetime)
	if err != nil {
		warn(WarnBadTime, "waterlevel_datetime %q", row.Datetime)
		return Observation{}, warns, false
	}
	if !row.WaterlevelMSL.Valid {
		warn(WarnNoWaterLevel, "waterlevel_msl %q", row.WaterlevelMSL.Raw)
		return Observation{}, warns, false
	}
	if !st.Lat.Valid || !st.Lon.Valid || st.Lat.Value < minLat || st.Lat.Value > maxLat || st.Lon.Value < minLon || st.Lon.Value > maxLon {
		warn(WarnBadPosition, "lat %v lon %v is outside %v–%v°N, %v–%v°E", st.Lat.Value, st.Lon.Value, minLat, maxLat, minLon, maxLon)
		return Observation{}, warns, false
	}
	if st.MinBank.Junk() {
		warn(WarnJunkValue, "min_bank %q treated as absent", st.MinBank.Raw)
	}
	if st.GroundLevel.Junk() {
		warn(WarnJunkValue, "ground_level %q treated as absent", st.GroundLevel.Raw)
	}

	load := LoadRefs{BankFull: mmPtr(st.MinBank), Ground: mmPtr(st.GroundLevel)}
	var g *GraphRefs
	if gr, ok := graph[st.ID]; ok {
		g = &gr
	}
	refs, bank, ground, mismatch := ResolveRefs(load, g)
	if mismatch != "" {
		warn(WarnRefMismatch, "%s", mismatch)
	}

	obs := Observation{
		Station: domain.Station{
			ID:           st.ID,
			Code:         code,
			Name:         pick(st.Name),
			Lat:          st.Lat.Value,
			Lon:          st.Lon.Value,
			ProvinceCode: strings.TrimSpace(row.Geocode.ProvinceCode),
			Province:     pick(row.Geocode.Province),
			Amphoe:       pick(row.Geocode.Amphoe),
			Tumbon:       pick(row.Geocode.Tumbon),
			Basin:        pick(row.Basin.Name),
			River:        strings.TrimSpace(row.RiverName),
			Agency:       pickEN(row.Agency.Short),
			Refs:         refs,
			BankFull:     bank,
			Ground:       ground,
		},
		Reading:  domain.Reading{StationID: st.ID, At: at, WL: mm(row.WaterlevelMSL.Value)},
		LoadRefs: load,
		Published: Published{
			Level:   row.SituationLevel,
			Percent: row.StoragePercent.Ptr(),
		},
	}
	return obs, warns, true
}

// ResolveRefs applies ADR 0003 to one station. Each graph reference, when
// present, wins over the load's; a difference beyond rounding is returned as
// a warning text. The returned bank and ground are zero unless the state
// says they are usable.
func ResolveRefs(load LoadRefs, graph *GraphRefs) (refs domain.Refs, bankFull, ground float64, mismatch string) {
	// Rounding here as well as at parse time keeps the rule safe for
	// references that come back from the store or from a test.
	bank, gnd := mmOf(load.BankFull), mmOf(load.Ground)
	if graph != nil {
		var diffs []string
		if gb := mmOf(graph.BankFull); gb != nil {
			if bank != nil && refsDiffer(*gb, *bank) {
				diffs = append(diffs, fmt.Sprintf("min_bank graph %.3f load %.3f", *gb, *bank))
			}
			bank = gb
		}
		if gg := mmOf(graph.Ground); gg != nil {
			if gnd != nil && refsDiffer(*gg, *gnd) {
				diffs = append(diffs, fmt.Sprintf("ground_level graph %.3f load %.3f", *gg, *gnd))
			}
			gnd = gg
		}
		mismatch = strings.Join(diffs, "; ")
	}

	switch {
	case bank == nil:
		return domain.RefsMissing, 0, 0, mismatch
	case gnd == nil && *bank == 0:
		// Without a ground we cannot tell a real 0 m bank from the dams'
		// "none" code, so take the cautious reading.
		return domain.RefsZeroCoded, 0, 0, mismatch
	case gnd == nil:
		return domain.RefsBankOnly, *bank, 0, mismatch
	case *bank == 0 && *gnd == 0:
		return domain.RefsZeroCoded, 0, 0, mismatch
	case *gnd >= *bank:
		return domain.RefsInverted, 0, 0, mismatch
	default:
		return domain.RefsComplete, *bank, *gnd, mismatch
	}
}

// Graph extracts a station's full-precision references and its readings from
// a history response. Null values are absent, not zero. Readings come back in
// ascending time with one value per instant.
func Graph(stationID int64, resp *thaiwater.GraphResponse) (GraphRefs, []domain.Reading, []Warning) {
	refs := GraphRefs{BankFull: mmPtr(resp.Data.MinBank), Ground: mmPtr(resp.Data.GroundLevel)}
	var warns []Warning
	byTime := map[int64]domain.Reading{}
	for _, p := range resp.Data.GraphData {
		if !p.Value.Valid {
			continue
		}
		at, err := ParseICT(p.Datetime)
		if err != nil {
			warns = append(warns, Warning{StationID: stationID, Kind: WarnBadTime, Detail: fmt.Sprintf("graph datetime %q", p.Datetime)})
			continue
		}
		byTime[at.Unix()] = domain.Reading{StationID: stationID, At: at, WL: mm(p.Value.Value)}
	}
	readings := make([]domain.Reading, 0, len(byTime))
	for _, r := range byTime {
		readings = append(readings, r)
	}
	sort.Slice(readings, func(i, j int) bool { return readings[i].At.Before(readings[j].At) })
	return refs, readings, warns
}

// CheckRuleWeb compares ThaiWater's live scale.data.rule_web with the
// thresholds the product hard-codes (domain.Thresholds). Order and spacing
// do not matter; every operator, term and level must match.
func CheckRuleWeb(rules []thaiwater.Rule) error {
	type step struct {
		op    string
		term  float64
		level int
	}
	want := []step{}
	for _, t := range domain.Thresholds {
		want = append(want, step{">", t.Above, int(t.Level)})
	}
	last := domain.Thresholds[len(domain.Thresholds)-1]
	want = append(want, step{"<=", last.Above, int(domain.LevelCriticallyLow)})

	got := make([]step, 0, len(rules))
	for _, r := range rules {
		term, err := strconv.ParseFloat(strings.TrimSpace(r.Term), 64)
		if err != nil {
			return fmt.Errorf("rule_web term %q is not a number", r.Term)
		}
		got = append(got, step{strings.TrimSpace(r.Operator), term, r.Level})
	}
	sort.Slice(got, func(i, j int) bool { return got[i].level > got[j].level })
	if len(got) != len(want) {
		return fmt.Errorf("rule_web has %d steps, the product expects %d: %v", len(got), len(want), rules)
	}
	for i := range want {
		if got[i] != want[i] {
			return fmt.Errorf("rule_web step for level %d is %s %v, the product expects %s %v (full rule_web: %v)",
				want[i].level, got[i].op, got[i].term, want[i].op, want[i].term, rules)
		}
	}
	return nil
}

// pick returns the Thai text, or the English when the Thai is empty.
func pick(t thaiwater.Text) string {
	if s := strings.TrimSpace(t.TH); s != "" {
		return s
	}
	return strings.TrimSpace(t.EN)
}

// pickEN returns the English text, or the Thai when the English is empty.
// Agency short names read best in English (HII, RID, EGAT).
func pickEN(t thaiwater.Text) string {
	if s := strings.TrimSpace(t.EN); s != "" {
		return s
	}
	return strings.TrimSpace(t.TH)
}

// mm rounds metres to the nearest millimetre. River levels are not measured
// more finely, and this removes float32 noise such as 40.599998.
func mm(v float64) float64 { return math.Round(v*1000) / 1000 }

func mmPtr(n thaiwater.Num) *float64 {
	if !n.Valid {
		return nil
	}
	v := mm(n.Value)
	return &v
}

func mmOf(p *float64) *float64 {
	if p == nil {
		return nil
	}
	v := mm(*p)
	return &v
}
