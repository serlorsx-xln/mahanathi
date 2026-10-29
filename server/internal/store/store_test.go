package store

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/serlorsx-xln/mahanathi/server/internal/domain"
	"github.com/serlorsx-xln/mahanathi/server/internal/normalize"
	"github.com/serlorsx-xln/mahanathi/server/internal/thaiwater"
)

var (
	ctx = context.Background()
	// t0 is 17:00 ICT on the day of the recording.
	t0 = time.Date(2026, 10, 29, 10, 0, 0, 0, time.UTC)
)

func open(t *testing.T) *Store {
	t.Helper()
	s, err := Open(ctx, filepath.Join(t.TempDir(), "mahanathi.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func fixture(t *testing.T) []normalize.Observation {
	t.Helper()
	b, err := os.ReadFile("../thaiwater/testdata/waterlevel_load.json")
	if err != nil {
		t.Fatal(err)
	}
	var load thaiwater.LoadResponse
	if err := json.Unmarshal(b, &load); err != nil {
		t.Fatal(err)
	}
	return normalize.Load(&load, nil).Observations
}

func obs(id int64, at time.Time, wl float64) normalize.Observation {
	bank, ground := 10.0, 2.0
	return normalize.Observation{
		Station: domain.Station{
			ID: id, Code: "T.1", Name: "ทดสอบ", Lat: 13.7, Lon: 100.5,
			ProvinceCode: "10", Province: "กรุงเทพมหานคร", Amphoe: "พระนคร", Tumbon: "พระบรมมหาราชวัง",
			Basin: "เจ้าพระยา", River: "แม่น้ำเจ้าพระยา", Agency: "RID",
			Refs: domain.RefsComplete, BankFull: bank, Ground: ground,
		},
		Reading:  domain.Reading{StationID: id, At: at, WL: wl},
		LoadRefs: normalize.LoadRefs{BankFull: &bank, Ground: &ground},
	}
}

func count(t *testing.T, s *Store, table string) int {
	t.Helper()
	var n int
	if err := s.r.QueryRowContext(ctx, "SELECT count(*) FROM "+table).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestOpenMigratesWALAndReopens(t *testing.T) {
	path := filepath.Join(t.TempDir(), "m.db")
	s, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	var mode string
	var version int
	if err := s.w.QueryRowContext(ctx, "PRAGMA journal_mode").Scan(&mode); err != nil || mode != "wal" {
		t.Errorf("journal_mode = %q %v, want wal", mode, err)
	}
	if err := s.w.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil || version != len(migrations) {
		t.Errorf("user_version = %d %v, want %d", version, err, len(migrations))
	}
	if _, err := s.SaveObservations(ctx, []normalize.Observation{obs(1, t0, 8.6)}, t0); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	// Reopening applies nothing and keeps the data.
	s, err = Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if n := count(t, s, "readings"); n != 1 {
		t.Errorf("readings after reopen = %d, want 1", n)
	}
	// Readers cannot write.
	if _, err := s.r.ExecContext(ctx, "DELETE FROM readings"); err == nil {
		t.Error("the read pool accepted a write")
	}
}

func TestRefusesANewerSchema(t *testing.T) {
	path := filepath.Join(t.TempDir(), "m.db")
	s, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.w.ExecContext(ctx, "PRAGMA user_version = 99"); err != nil {
		t.Fatal(err)
	}
	s.Close()
	if _, err := Open(ctx, path); !errors.Is(err, ErrNewerSchema) {
		t.Errorf("Open on a newer schema: %v, want ErrNewerSchema", err)
	}
}

func TestOpenFailsOnAnUnusablePath(t *testing.T) {
	if _, err := Open(ctx, filepath.Join(t.TempDir(), "missing", "dir", "m.db")); err == nil {
		t.Error("Open in a missing directory succeeded")
	}
}

func TestSaveObservationsIsIdempotent(t *testing.T) {
	s := open(t)
	all := fixture(t)
	n, err := s.SaveObservations(ctx, all, t0)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(all) {
		t.Errorf("first save wrote %d readings, want %d", n, len(all))
	}
	n, err = s.SaveObservations(ctx, all, t0.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Errorf("second save changed %d readings, want 0", n)
	}
	if got := count(t, s, "readings"); got != len(all) {
		t.Errorf("readings = %d, want %d", got, len(all))
	}
	if got := count(t, s, "stations"); got != len(all) {
		t.Errorf("stations = %d, want %d", got, len(all))
	}
	states, err := s.Stations(ctx)
	if err != nil {
		t.Fatal(err)
	}
	byCode := map[string]StationState{}
	for _, st := range states {
		byCode[st.Station.Code] = st
	}
	// Spot-check a full round trip against normalize's output.
	for _, o := range all {
		got := byCode[o.Station.Code]
		if got.Station != o.Station {
			t.Errorf("%s: station\n got %+v\nwant %+v", o.Station.Code, got.Station, o.Station)
		}
		if got.Latest == nil || !got.Latest.At.Equal(o.Reading.At) || got.Latest.WL != o.Reading.WL || got.Latest.At.Location() != time.UTC {
			t.Errorf("%s: latest %+v, want %+v", o.Station.Code, got.Latest, o.Reading)
		}
		if !got.FirstSeen.Equal(t0) || !got.LastSeen.Equal(t0.Add(time.Hour)) {
			t.Errorf("%s: seen %v to %v", o.Station.Code, got.FirstSeen, got.LastSeen)
		}
	}
}

func TestStationChangesAreKept(t *testing.T) {
	s := open(t)
	o := obs(1, t0, 8.6)
	if _, err := s.SaveObservations(ctx, []normalize.Observation{o}, t0); err != nil {
		t.Fatal(err)
	}
	o.Station.Name, o.Station.Refs, o.Station.Ground = "ทดสอบใหม่", domain.RefsBankOnly, 0
	o.LoadRefs.Ground = nil
	if _, err := s.SaveObservations(ctx, []normalize.Observation{o}, t0.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	st, ok, err := s.Station(ctx, 1)
	if err != nil || !ok {
		t.Fatal(ok, err)
	}
	if st.Station.Name != "ทดสอบใหม่" || st.Station.Refs != domain.RefsBankOnly || st.Station.Ground != 0 {
		t.Errorf("station not updated: %+v", st.Station)
	}
	if st.LoadRefs.Ground != nil || st.LoadRefs.BankFull == nil || *st.LoadRefs.BankFull != 10 {
		t.Errorf("load refs = %+v", st.LoadRefs)
	}
	if _, ok, err := s.Station(ctx, 404); ok || err != nil {
		t.Errorf("unknown station: ok %v err %v", ok, err)
	}
}

// The load rounds water levels to 2 decimals and the graph gives 3, so for
// the same minute the graph's value is kept, whichever arrives first.
func TestGraphBeatsLoadForTheSameMinute(t *testing.T) {
	s := open(t)
	at := t0.Add(40 * time.Minute)
	if _, err := s.SaveObservations(ctx, []normalize.Observation{obs(12479, at, 17.78)}, at); err != nil {
		t.Fatal(err)
	}
	n, err := s.SaveHistory(ctx, 12479, normalize.GraphRefs{}, []domain.Reading{{StationID: 12479, At: at, WL: 17.782}}, at)
	if err != nil || n != 1 {
		t.Fatalf("graph upsert changed %d, %v", n, err)
	}
	n, err = s.SaveObservations(ctx, []normalize.Observation{obs(12479, at, 17.78)}, at.Add(time.Hour))
	if err != nil || n != 0 {
		t.Fatalf("load after graph changed %d, %v", n, err)
	}
	st, _, _ := s.Station(ctx, 12479)
	if st.Latest == nil || st.Latest.WL != 17.782 {
		t.Errorf("latest = %+v, want the graph's 17.782", st.Latest)
	}
	// A corrected graph value replaces an earlier graph value.
	if n, _ := s.SaveHistory(ctx, 12479, normalize.GraphRefs{}, []domain.Reading{{StationID: 12479, At: at, WL: 17.783}}, at); n != 1 {
		t.Errorf("graph over graph changed %d, want 1", n)
	}
}

func TestSaveHistoryNeedsAKnownStation(t *testing.T) {
	s := open(t)
	if _, err := s.SaveHistory(ctx, 1, normalize.GraphRefs{}, nil, t0); err == nil {
		t.Error("history saved for a station that was never observed")
	}
}

func TestBackfillBookkeeping(t *testing.T) {
	s := open(t)
	if _, err := s.SaveObservations(ctx, []normalize.Observation{obs(1, t0, 8.6), obs(2, t0, 8.6), obs(3, t0, 8.6)}, t0); err != nil {
		t.Fatal(err)
	}
	need, err := s.NeedsHistory(ctx)
	if err != nil || len(need) != 3 {
		t.Fatalf("needs history = %v %v, want all three", need, err)
	}
	bank, ground := 10.004, 1.996
	readings := []domain.Reading{
		{StationID: 2, At: t0.Add(-2 * time.Hour), WL: 8.4},
		{StationID: 2, At: t0.Add(-time.Hour), WL: 8.5},
	}
	if n, err := s.SaveHistory(ctx, 2, normalize.GraphRefs{BankFull: &bank, Ground: &ground}, readings, t0); err != nil || n != 2 {
		t.Fatalf("SaveHistory = %d %v", n, err)
	}
	need, _ = s.NeedsHistory(ctx)
	if len(need) != 2 || need[0] != 1 || need[1] != 3 {
		t.Errorf("needs history = %v, want [1 3]", need)
	}
	refs, err := s.GraphRefs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(refs) != 1 || *refs[2].BankFull != bank || *refs[2].Ground != ground {
		t.Errorf("graph refs = %+v", refs)
	}
	// A refresh is requested by forgetting the history mark, as ingest does
	// when the load's references stop agreeing with the graph's.
	if err := s.ForgetHistory(ctx, 2); err != nil {
		t.Fatal(err)
	}
	need, _ = s.NeedsHistory(ctx)
	if len(need) != 3 {
		t.Errorf("after ForgetHistory needs = %v", need)
	}
	if refs, _ := s.GraphRefs(ctx); len(refs) != 1 {
		t.Error("ForgetHistory dropped the graph references it is meant to refresh")
	}
}

func TestStationsWithoutReadings(t *testing.T) {
	s := open(t)
	if _, err := s.SaveObservations(ctx, []normalize.Observation{obs(1, t0, 8.6)}, t0); err != nil {
		t.Fatal(err)
	}
	if _, err := s.w.ExecContext(ctx, "DELETE FROM readings"); err != nil {
		t.Fatal(err)
	}
	states, err := s.Stations(ctx)
	if err != nil || len(states) != 1 || states[0].Latest != nil {
		t.Errorf("states = %+v %v, want one station with no latest reading", states, err)
	}
}

func TestLatestHistoryAndWindow(t *testing.T) {
	s := open(t)
	var batch []domain.Reading
	for h := 0; h < 10; h++ {
		batch = append(batch, domain.Reading{StationID: 1, At: t0.Add(-time.Duration(h) * time.Hour), WL: 8 + float64(h)/10})
	}
	if _, err := s.SaveObservations(ctx, []normalize.Observation{obs(1, t0, 8), obs(2, t0.Add(-time.Hour), 5)}, t0); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SaveHistory(ctx, 1, normalize.GraphRefs{}, batch, t0); err != nil {
		t.Fatal(err)
	}
	states, err := s.Stations(ctx)
	if err != nil || len(states) != 2 {
		t.Fatal(states, err)
	}
	if !states[0].Latest.At.Equal(t0) || states[1].Latest.WL != 5 {
		t.Errorf("latest = %+v / %+v", states[0].Latest, states[1].Latest)
	}

	// History is (since, until], oldest first.
	h, err := s.History(ctx, 1, t0.Add(-3*time.Hour), t0.Add(-time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(h) != 2 || !h[0].At.Equal(t0.Add(-2*time.Hour)) || !h[1].At.Equal(t0.Add(-time.Hour)) || h[0].WL != 8.2 {
		t.Errorf("history = %+v", h)
	}

	w, err := s.ReadingsSince(ctx, t0.Add(-2*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(w[1]) != 2 || len(w[2]) != 1 || !w[1][0].At.Before(w[1][1].At) {
		t.Errorf("window = %+v", w)
	}
}

func TestRetentionDropsOldReadings(t *testing.T) {
	s := open(t)
	if _, err := s.SaveObservations(ctx, []normalize.Observation{obs(1, t0, 8)}, t0); err != nil {
		t.Fatal(err)
	}
	old := []domain.Reading{
		{StationID: 1, At: t0.Add(-Retention - time.Hour), WL: 1},
		{StationID: 1, At: t0.Add(-Retention), WL: 2}, // exactly 30 days is kept
	}
	if _, err := s.SaveHistory(ctx, 1, normalize.GraphRefs{}, old, t0); err != nil {
		t.Fatal(err)
	}
	n, err := s.Prune(ctx, t0)
	if err != nil || n != 1 {
		t.Errorf("Prune = %d %v, want 1", n, err)
	}
	if got := count(t, s, "readings"); got != 2 {
		t.Errorf("readings left = %d, want 2", got)
	}
}

func TestRunsAreRecorded(t *testing.T) {
	s := open(t)
	if _, ok, err := s.LastRun(ctx); ok || err != nil {
		t.Errorf("LastRun on an empty store: %v %v", ok, err)
	}
	r := Run{
		Started: t0, Finished: t0.Add(4 * time.Second), Rows: 804, Observations: 797, Dropped: 7,
		Warnings: 12, Upserted: 790, HistoryOK: 3, HistoryFailed: 1, RuleWebOK: true,
		OracleChecked: 789, OracleMismatches: 0,
	}
	if err := s.RecordRun(ctx, r); err != nil {
		t.Fatal(err)
	}
	r2 := r
	r2.Started, r2.Finished, r2.Err, r2.RuleWebOK = t0.Add(time.Hour), t0.Add(time.Hour+time.Second), "load: 503", false
	if err := s.RecordRun(ctx, r2); err != nil {
		t.Fatal(err)
	}
	got, ok, err := s.LastRun(ctx)
	if err != nil || !ok {
		t.Fatal(ok, err)
	}
	if got != r2 {
		t.Errorf("LastRun\n got %+v\nwant %+v", got, r2)
	}
	last, ok, err := s.LastGoodRun(ctx)
	if err != nil || !ok || last != r {
		t.Errorf("LastGoodRun = %+v %v %v", last, ok, err)
	}
}

func TestMeta(t *testing.T) {
	s := open(t)
	if _, ok, err := s.Meta(ctx, "rule_web"); ok || err != nil {
		t.Errorf("missing key: %v %v", ok, err)
	}
	for _, v := range []string{`[{"a":1}]`, `[{"a":2}]`} {
		if err := s.SetMeta(ctx, "rule_web", v); err != nil {
			t.Fatal(err)
		}
		if got, ok, err := s.Meta(ctx, "rule_web"); !ok || err != nil || got != v {
			t.Errorf("Meta = %q %v %v, want %q", got, ok, err, v)
		}
	}
}

// Readers and the single writer work at the same time under WAL.
func TestConcurrentReadsDuringWrites(t *testing.T) {
	s := open(t)
	var wg sync.WaitGroup
	errs := make(chan error, 64)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			at := t0.Add(time.Duration(i) * time.Hour)
			if _, err := s.SaveObservations(ctx, []normalize.Observation{obs(1, at, 8), obs(2, at, 9)}, at); err != nil {
				errs <- err
				return
			}
		}
	}()
	for g := 0; g < 4; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 20; i++ {
				if _, err := s.Stations(ctx); err != nil {
					errs <- err
					return
				}
				if _, err := s.ReadingsSince(ctx, t0); err != nil {
					errs <- err
					return
				}
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
	if got := count(t, s, "readings"); got != 40 {
		t.Errorf("readings = %d, want 40", got)
	}
}

func TestClosedStoreErrors(t *testing.T) {
	s, err := Open(ctx, filepath.Join(t.TempDir(), "m.db"))
	if err != nil {
		t.Fatal(err)
	}
	s.Close()
	checks := map[string]error{}
	_, checks["SaveObservations"] = s.SaveObservations(ctx, []normalize.Observation{obs(1, t0, 8)}, t0)
	_, checks["SaveHistory"] = s.SaveHistory(ctx, 1, normalize.GraphRefs{}, nil, t0)
	_, checks["Stations"] = s.Stations(ctx)
	_, _, checks["Station"] = s.Station(ctx, 1)
	_, checks["History"] = s.History(ctx, 1, t0, t0)
	_, checks["ReadingsSince"] = s.ReadingsSince(ctx, t0)
	_, checks["NeedsHistory"] = s.NeedsHistory(ctx)
	_, checks["GraphRefs"] = s.GraphRefs(ctx)
	checks["ForgetHistory"] = s.ForgetHistory(ctx, 1)
	_, checks["Prune"] = s.Prune(ctx, t0)
	checks["RecordRun"] = s.RecordRun(ctx, Run{Started: t0})
	_, _, checks["LastRun"] = s.LastRun(ctx)
	_, _, checks["LastGoodRun"] = s.LastGoodRun(ctx)
	_, _, checks["Meta"] = s.Meta(ctx, "k")
	checks["SetMeta"] = s.SetMeta(ctx, "k", "v")
	for name, err := range checks {
		if err == nil {
			t.Errorf("%s on a closed store returned no error", name)
		}
	}
}

func TestPathsThatWouldBreakTheDSN(t *testing.T) {
	for _, p := range []string{"/tmp/a?b.db", "/tmp/a#b.db"} {
		if _, err := Open(ctx, p); err == nil {
			t.Errorf("Open(%q) succeeded", p)
		}
	}
}

// A migration that fails is rolled back whole, version included, so the
// next start tries it again from a clean state.
func TestFailedMigrationRollsBack(t *testing.T) {
	saved := migrations
	t.Cleanup(func() { migrations = saved })
	migrations = append(append([]string{}, saved...), "CREATE TABLE extra (x INTEGER) STRICT; SELECT no_such_function();")
	path := filepath.Join(t.TempDir(), "m.db")
	if _, err := Open(ctx, path); err == nil {
		t.Fatal("a failing migration did not fail Open")
	}
	migrations = saved
	s, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	var version, extra int
	s.w.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version)
	s.w.QueryRowContext(ctx, "SELECT count(*) FROM sqlite_schema WHERE name = 'extra'").Scan(&extra)
	if version != len(saved) || extra != 0 {
		t.Errorf("after a failed migration: version %d, extra table %d", version, extra)
	}
}
