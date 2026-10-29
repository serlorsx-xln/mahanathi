package api

import (
	"sort"
	"time"

	"github.com/serlorsx-xln/mahanathi/server/internal/domain"
	"github.com/serlorsx-xln/mahanathi/server/internal/humanize"
	"github.com/serlorsx-xln/mahanathi/server/internal/normalize"
	"github.com/serlorsx-xln/mahanathi/server/internal/store"
)

// The clocks and windows the summary is built around. StaleAfter comes from
// the humanize spec; the trend needs six buckets before the latest reading's
// own, plus that bucket itself.
const (
	trendBuckets = 7
)

// NowResponse is GET /api/v1/now: every station's current state in human
// terms, plus the national summary. The encoding is documented in
// docs/api.md.
type NowResponse struct {
	GeneratedAt time.Time    `json:"generatedAt"`
	Stations    []StationNow `json:"stations"`
	Summary     Summary      `json:"summary"`
}

// StationNow is one station's current state. Nil pointers mean "not
// derivable", never zero: a station without references has no percent, not a
// percent of zero.
type StationNow struct {
	ID         int64           `json:"id"`
	Code       string          `json:"code"`
	Name       string          `json:"name"`
	Lat        float64         `json:"lat"`
	Lon        float64         `json:"lon"`
	Province   string          `json:"province"`
	Amphoe     string          `json:"amphoe"`
	Basin      string          `json:"basin"`
	River      string          `json:"river"`
	Agency     string          `json:"agency"`
	WL         *float64        `json:"wl"` // metres MSL
	MeasuredAt *time.Time      `json:"measuredAt"`
	BankFull   *float64        `json:"bankFull"`   // metres MSL
	Ground     *float64        `json:"ground"`     // metres MSL
	Pct        *float64        `json:"pct"`        // percent of bank-full, 2 decimals
	Level      *int            `json:"level"`      // 1..5 per rule_web
	DistanceMM *int64          `json:"distanceMM"` // wl − bank-full
	Trend6h    *float64        `json:"trend6h"`    // metres, positive is rising
	Stale      bool            `json:"stale"`
	Suspect    []string        `json:"suspect"` // reason codes; empty when fine
	Sentence   string          `json:"sentence"`
	Facts      []humanize.Fact `json:"facts"`
}

// Summary is the national picture, built from the stations that count as
// current: a reading no older than normalize.StaleAfter. Suspect readings
// still count as data, but never as high water.
type Summary struct {
	Stations     int      `json:"stations"`
	Levels       [5]int   `json:"levels"`     // counts by level 1..5
	NoLevel      int      `json:"noLevel"`    // current, trustworthy, no level
	StaleCount   int      `json:"staleCount"` // not counted anywhere above
	SuspectCount int      `json:"suspectCount"`
	TopProvinces []string `json:"topProvinces"` // provinces holding level 4–5
	National     string   `json:"national"`
}

// BuildNow assembles the /now response. readings holds each station's
// readings of the last days (the trend and the jump check read the same
// window; see docs/api.md).
func BuildNow(stations []store.StationState, readings map[int64][]domain.Reading, now time.Time) NowResponse {
	res := NowResponse{GeneratedAt: now, Stations: make([]StationNow, 0, len(stations))}
	sum := Summary{TopProvinces: []string{}}
	provinceCount := map[string]int{}
	var overflow []string
	tally := humanize.Tally{}

	for _, st := range stations {
		s := stationNow(st, readings[st.Station.ID], now)
		res.Stations = append(res.Stations, s)
		sum.Stations++
		if s.MeasuredAt == nil {
			continue
		}
		if s.Stale {
			sum.StaleCount++
			continue
		}
		tally.Current++
		suspect := len(s.Suspect) > 0
		if suspect {
			sum.SuspectCount++
		}
		if s.Level == nil {
			if !suspect {
				sum.NoLevel++
			}
			continue
		}
		if !suspect {
			sum.Levels[*s.Level-1]++
			if *s.Level >= 4 && s.Province != "" {
				provinceCount[s.Province]++
			}
			if *s.Level == 5 {
				overflow = append(overflow, s.Province)
				tally.Overflow = overflow
			}
			if *s.Level == 4 {
				tally.High++
			}
		}
	}

	for p, n := range provinceCount {
		if n > 0 {
			sum.TopProvinces = append(sum.TopProvinces, p)
		}
	}
	sortNamesByCount(sum.TopProvinces, provinceCount)
	sum.National = humanize.National(tally)
	res.Summary = sum
	return res
}

func stationNow(st store.StationState, readings []domain.Reading, now time.Time) StationNow {
	s := StationNow{
		ID:       st.Station.ID,
		Code:     st.Station.Code,
		Name:     st.Station.Name,
		Lat:      st.Station.Lat,
		Lon:      st.Station.Lon,
		Province: st.Station.Province,
		Amphoe:   st.Station.Amphoe,
		Basin:    st.Station.Basin,
		River:    st.Station.River,
		Agency:   st.Station.Agency,
	}
	if st.Station.HasBankFull() {
		b := st.Station.BankFull
		s.BankFull = &b
	}
	if st.Station.HasGround() {
		g := st.Station.Ground
		s.Ground = &g
	}
	if st.Latest == nil {
		s.Sentence, s.Facts = noData(st)
		return s
	}

	latest := *st.Latest
	wl := latest.WL
	s.WL, s.MeasuredAt = &wl, &latest.At
	if d, ok := st.Station.Distance(wl); ok {
		s.DistanceMM = &d
	}
	if h, ok := st.Station.PctHundredths(wl); ok {
		p := float64(h) / 100
		s.Pct = &p
		lvl := int(st.Station.Level(wl))
		if st.Station.Level(wl) != domain.LevelUnknown {
			s.Level = &lvl
		}
	}

	stale := normalize.IsStale(latest.At, now)
	hourly := normalize.Hourly(readings, domain.HourEnd(latest.At), trendBuckets)
	var prev domain.Point
	if len(hourly) == trendBuckets {
		prev = hourly[trendBuckets-2] // the bucket just before the latest's own
	}
	suspect := normalize.SuspectReasons(st.Station, latest, prev, now)
	// An empty slice, not nil: the wire shape in docs/api.md is "suspect": [],
	// and clients index it without a null check.
	s.Suspect = make([]string, 0, len(suspect))
	for _, r := range suspect {
		s.Suspect = append(s.Suspect, r.String())
	}

	out := humanize.Describe(humanize.Input{
		Station: st.Station, Latest: &latest, Hourly: hourly,
		Stale: stale, Suspect: suspect, Now: now,
	})
	s.Stale, s.Sentence, s.Facts = stale, out.Sentence, out.Facts
	if out.Trend != nil {
		t := float64(out.Trend.DeltaMM) / 1000
		s.Trend6h = &t
	}
	return s
}

func noData(st store.StationState) (string, []humanize.Fact) {
	out := humanize.Describe(humanize.Input{Station: st.Station, Now: time.Time{}})
	return out.Sentence, out.Facts
}

// sortNamesByCount orders names by their count, highest first, then by name,
// so the order is stable across runs.
func sortNamesByCount(names []string, count map[string]int) {
	sort.Slice(names, func(i, j int) bool {
		if count[names[i]] != count[names[j]] {
			return count[names[i]] > count[names[j]]
		}
		return names[i] < names[j]
	})
}
