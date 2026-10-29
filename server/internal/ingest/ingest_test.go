package ingest

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/serlorsx-xln/mahanathi/server/internal/domain"
	"github.com/serlorsx-xln/mahanathi/server/internal/store"
	"github.com/serlorsx-xln/mahanathi/server/internal/thaiwater"
)

const fixtures = "../thaiwater/testdata/"

// recorded is just after the history fixtures were recorded (19:52 ICT), so
// none of their readings is in the future.
var recorded = time.Date(2026, 10, 29, 12, 55, 0, 0, time.UTC)

// fake is ThaiWater as recorded, with knobs for failure and slowness.
type fake struct {
	mu          sync.Mutex
	load        *thaiwater.LoadResponse
	loadErr     error
	graphs      map[int64]*thaiwater.GraphResponse
	graphErr    map[int64]error
	delay       time.Duration
	inFlight    int
	maxInFlight int
	graphCalls  []int64
	windows     [][2]time.Time
}

func newFake(t *testing.T) *fake {
	t.Helper()
	f := &fake{graphs: map[int64]*thaiwater.GraphResponse{}, graphErr: map[int64]error{}}
	f.load = &thaiwater.LoadResponse{}
	readJSON(t, "waterlevel_load.json", f.load)
	for _, id := range []int64{737, 1097520, 575568, 3441, 12479} {
		g := &thaiwater.GraphResponse{}
		readJSON(t, fmt.Sprintf("waterlevel_graph_%d.json", id), g)
		f.graphs[id] = g
	}
	return f
}

func readJSON(t *testing.T, name string, v any) {
	t.Helper()
	b, err := os.ReadFile(fixtures + name)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, v); err != nil {
		t.Fatal(err)
	}
}

func (f *fake) Load(ctx context.Context) (*thaiwater.LoadResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.loadErr != nil {
		return nil, f.loadErr
	}
	return f.load, nil
}

func (f *fake) Graph(ctx context.Context, id int64, from, to time.Time) (*thaiwater.GraphResponse, error) {
	f.mu.Lock()
	f.inFlight++
	f.maxInFlight = max(f.maxInFlight, f.inFlight)
	f.graphCalls = append(f.graphCalls, id)
	f.windows = append(f.windows, [2]time.Time{from, to})
	g, err, delay := f.graphs[id], f.graphErr[id], f.delay
	f.mu.Unlock()
	defer func() {
		f.mu.Lock()
		f.inFlight--
		f.mu.Unlock()
	}()
	if delay > 0 {
		time.Sleep(delay)
	}
	if err != nil {
		return nil, err
	}
	if g == nil {
		g = &thaiwater.GraphResponse{Result: "OK"} // a station with no history
	}
	return g, nil
}

func (f *fake) calls() []int64 {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]int64(nil), f.graphCalls...)
}

type harness struct {
	in      *Ingester
	src     *fake
	st      *store.Store
	logs    *bytes.Buffer
	now     time.Time
	changes int
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	st, err := store.Open(context.Background(), filepath.Join(t.TempDir(), "m.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	h := &harness{src: newFake(t), st: st, logs: &bytes.Buffer{}, now: recorded}
	h.in, err = New(Options{
		Source:   h.src,
		Store:    st,
		Logger:   slog.New(slog.NewJSONHandler(h.logs, &slog.HandlerOptions{Level: slog.LevelDebug})),
		Now:      func() time.Time { return h.now },
		OnChange: func() { h.changes++ },
	})
	if err != nil {
		t.Fatal(err)
	}
	return h
}

type logLine struct {
	Level string `json:"level"`
	Msg   string `json:"msg"`
	Error string `json:"error"`
	Old   string `json:"old"`
	New   string `json:"new"`
}

func (h *harness) lines(t *testing.T) []logLine {
	t.Helper()
	var out []logLine
	sc := bufio.NewScanner(bytes.NewReader(h.logs.Bytes()))
	for sc.Scan() {
		var l logLine
		if err := json.Unmarshal(sc.Bytes(), &l); err != nil {
			t.Fatalf("log line %q: %v", sc.Text(), err)
		}
		out = append(out, l)
	}
	return out
}

func (h *harness) has(t *testing.T, level, msg string) bool {
	t.Helper()
	for _, l := range h.lines(t) {
		if l.Level == level && l.Msg == msg {
			return true
		}
	}
	return false
}

func TestFirstRunStoresTheSnapshotAndBackfillsEveryStation(t *testing.T) {
	h := newHarness(t)
	run, err := h.in.RunOnce(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if run.Rows != 37 || run.Observations != 37 || run.Dropped != 0 {
		t.Errorf("rows %d, observations %d, dropped %d", run.Rows, run.Observations, run.Dropped)
	}
	if run.HistoryOK != 37 || run.HistoryFailed != 0 {
		t.Errorf("history ok %d, failed %d", run.HistoryOK, run.HistoryFailed)
	}
	if !run.RuleWebOK || run.Err != "" {
		t.Errorf("rule_web ok %v, err %q", run.RuleWebOK, run.Err)
	}
	// With full-precision references from the graph, every published level
	// is reproduced (ADR 0003).
	if run.OracleChecked != 33 || run.OracleMismatches != 0 {
		t.Errorf("oracle %d checked, %d mismatched", run.OracleChecked, run.OracleMismatches)
	}
	if run.Upserted < 37+500 {
		t.Errorf("upserted %d; expected the snapshot plus five recorded histories", run.Upserted)
	}
	if !run.Started.Equal(recorded) || run.Finished.Before(run.Started) {
		t.Errorf("run times %v to %v", run.Started, run.Finished)
	}
	if h.changes == 0 {
		t.Error("OnChange was not called")
	}

	// The history window asked for is the last 72 hours.
	for _, w := range h.src.windows {
		if !w[0].Equal(recorded.Add(-72*time.Hour)) || !w[1].Equal(recorded) {
			t.Fatalf("history window %v to %v", w[0], w[1])
		}
	}

	ctx := context.Background()
	urtu09, _, _ := h.st.Station(ctx, 3441)
	if urtu09.Station.BankFull != 225.386 || urtu09.Station.Level(urtu09.Latest.WL) != domain.LevelHigh {
		t.Errorf("URTU09 not re-resolved with its graph references: %+v", urtu09.Station)
	}
	// BSP002's newest reading is the graph's 19:40 — its 19:50 point is
	// null — not the load's 17:40.
	bsp, _, _ := h.st.Station(ctx, 12479)
	if bsp.Latest == nil || !bsp.Latest.At.Equal(time.Date(2026, 10, 29, 12, 40, 0, 0, time.UTC)) || bsp.Latest.WL != 16.727 {
		t.Errorf("BSP002 latest = %+v", bsp.Latest)
	}
	if need, _ := h.st.NeedsHistory(ctx); len(need) != 0 {
		t.Errorf("stations still needing history: %v", need)
	}
	last, ok, _ := h.st.LastRun(ctx)
	if !ok || last != run {
		t.Errorf("recorded run\n got %+v\nwant %+v", last, run)
	}
	if !h.has(t, "INFO", "ingest run") {
		t.Error("no run summary in the log")
	}
}

func TestSecondRunChangesNothingAndFetchesNoHistory(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	if _, err := h.in.RunOnce(ctx); err != nil {
		t.Fatal(err)
	}
	calls, changes := len(h.src.calls()), h.changes
	h.now = recorded.Add(time.Hour)
	run, err := h.in.RunOnce(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if run.Upserted != 0 || run.HistoryOK != 0 || len(h.src.calls()) != calls {
		t.Errorf("repeat run: upserted %d, history %d, graph calls %d → %d", run.Upserted, run.HistoryOK, calls, len(h.src.calls()))
	}
	if h.changes != changes {
		t.Error("OnChange called when nothing changed")
	}
}

func TestHistoryRequestsStayWithinTheLimit(t *testing.T) {
	h := newHarness(t)
	h.src.delay = 5 * time.Millisecond
	in, err := New(Options{Source: h.src, Store: h.st, Logger: slog.New(slog.DiscardHandler), Now: func() time.Time { return h.now }, Concurrency: 16})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := in.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if h.src.maxInFlight > thaiwater.MaxGraphConcurrency || h.src.maxInFlight < 2 {
		t.Errorf("max history requests in flight = %d, want 2..%d", h.src.maxInFlight, thaiwater.MaxGraphConcurrency)
	}
}

func TestFailedHistoryIsRetriedNextRun(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	h.src.graphErr[3441] = errors.New("graph: 503")
	run, err := h.in.RunOnce(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if run.HistoryOK != 36 || run.HistoryFailed != 1 || run.Err != "" {
		t.Errorf("history ok %d, failed %d, err %q", run.HistoryOK, run.HistoryFailed, run.Err)
	}
	// Without its graph references URTU09 computes 69.994 %, level 3, where
	// ThaiWater publishes 4. The run says so.
	if run.OracleMismatches != 1 || !h.has(t, "WARN", "levels differ from ThaiWater's published levels") {
		t.Errorf("oracle mismatches %d", run.OracleMismatches)
	}
	if !h.has(t, "WARN", "history fetch failed") {
		t.Error("the failure was not logged")
	}
	delete(h.src.graphErr, 3441)
	before := len(h.src.calls())
	h.now = recorded.Add(time.Hour)
	run, err = h.in.RunOnce(ctx)
	if err != nil {
		t.Fatal(err)
	}
	calls := h.src.calls()[before:]
	if len(calls) != 1 || calls[0] != 3441 || run.HistoryOK != 1 || run.OracleMismatches != 0 {
		t.Errorf("retry: calls %v, history ok %d, mismatches %d", calls, run.HistoryOK, run.OracleMismatches)
	}
}

func TestRuleWebChangeIsLoud(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	if _, err := h.in.RunOnce(ctx); err != nil {
		t.Fatal(err)
	}
	if h.has(t, "ERROR", "rule_web changed since the last run") {
		t.Fatal("the first run reported a change")
	}
	// ThaiWater moves level 4 from > 70 to > 75.
	h.src.load.Scale.Data.RuleWeb = append([]thaiwater.Rule(nil), h.src.load.Scale.Data.RuleWeb...)
	for i, r := range h.src.load.Scale.Data.RuleWeb {
		if r.Level == 4 {
			h.src.load.Scale.Data.RuleWeb[i].Term = "75"
		}
	}
	h.now = recorded.Add(time.Hour)
	run, err := h.in.RunOnce(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if run.RuleWebOK {
		t.Error("RuleWebOK with a changed rule_web")
	}
	var changed, differs bool
	for _, l := range h.lines(t) {
		if l.Level == "ERROR" && l.Msg == "rule_web changed since the last run" && l.Old != l.New && l.New != "" {
			changed = true
		}
		if l.Level == "ERROR" && l.Msg == "rule_web differs from the thresholds the product uses" && l.Error != "" {
			differs = true
		}
	}
	if !changed || !differs {
		t.Errorf("logs lack the loud warnings (changed %v, differs %v):\n%s", changed, differs, h.logs)
	}
	// The data is still stored; the thresholds in use do not change silently.
	if run.Observations != 37 {
		t.Errorf("observations = %d", run.Observations)
	}
}

func TestLoadFailureIsRecordedAndReturned(t *testing.T) {
	h := newHarness(t)
	h.src.loadErr = errors.New("thaiwater: 503 after 4 attempts")
	run, err := h.in.RunOnce(context.Background())
	if err == nil || run.Err == "" {
		t.Fatalf("RunOnce = %+v, %v; want the error", run, err)
	}
	last, ok, _ := h.st.LastRun(context.Background())
	if !ok || last.Err != run.Err {
		t.Errorf("recorded run %+v", last)
	}
	if _, ok, _ := h.st.LastGoodRun(context.Background()); ok {
		t.Error("a failed run counted as good")
	}
	if !h.has(t, "ERROR", "ingest run failed") {
		t.Error("the failure was not logged")
	}
}

// A store that fails mid-run fails the run; it must not panic or be recorded
// as good.
func TestStoreErrorFailsTheRun(t *testing.T) {
	h := newHarness(t)
	h.st.Close() // the t.Cleanup close is a no-op on an already-closed pool
	run, err := h.in.RunOnce(context.Background())
	if err == nil || run.Err == "" {
		t.Fatalf("RunOnce = %+v, %v; want the store error", run, err)
	}
	if !h.has(t, "ERROR", "ingest run failed") {
		t.Error("the failure was not logged")
	}
}

func TestGraphReferenceDisagreementRefetchesOncePerDay(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	if _, err := h.in.RunOnce(ctx); err != nil {
		t.Fatal(err)
	}
	// The load now publishes a bank 1 m lower than the stored graph value.
	for i, r := range h.src.load.WaterlevelData.Data {
		if r.Station.ID == 3441 {
			h.src.load.WaterlevelData.Data[i].Station.MinBank = thaiwater.Num{Value: 224.39, Valid: true, Present: true}
		}
	}
	h.now = recorded.Add(time.Hour)
	if _, err := h.in.RunOnce(ctx); err != nil {
		t.Fatal(err)
	}
	if need, _ := h.st.NeedsHistory(ctx); len(need) != 0 {
		t.Errorf("refetch asked for within a day of the last fetch: %v", need)
	}
	before := len(h.src.calls())
	h.now = recorded.Add(25 * time.Hour)
	if _, err := h.in.RunOnce(ctx); err != nil {
		t.Fatal(err)
	}
	if calls := h.src.calls()[before:]; len(calls) != 1 || calls[0] != 3441 {
		t.Errorf("after a day, history calls = %v, want [3441]", calls)
	}
}

func TestOldReadingsArePruned(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	if _, err := h.in.RunOnce(ctx); err != nil {
		t.Fatal(err)
	}
	h.now = recorded.Add(store.Retention + 5*24*time.Hour)
	if _, err := h.in.RunOnce(ctx); err != nil {
		t.Fatal(err)
	}
	left, err := h.st.ReadingsSince(ctx, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	for id, rs := range left {
		for _, r := range rs {
			if r.At.Before(h.now.Add(-store.Retention)) {
				t.Fatalf("station %d kept a reading from %v", id, r.At)
			}
		}
	}
}

func TestNextRun(t *testing.T) {
	ict := time.FixedZone("ICT", 7*3600)
	at := func(h, m, s int) time.Time { return time.Date(2026, 10, 29, h, m, s, 0, ict) }
	cases := []struct {
		now, want     time.Time
		every, offset time.Duration
	}{
		{at(14, 0, 0), at(14, 7, 0), time.Hour, 7 * time.Minute},
		{at(14, 6, 59), at(14, 7, 0), time.Hour, 7 * time.Minute},
		{at(14, 7, 0), at(15, 7, 0), time.Hour, 7 * time.Minute}, // never "now" again
		{at(14, 30, 0), at(15, 7, 0), time.Hour, 7 * time.Minute},
		{at(23, 30, 0), time.Date(2026, 10, 30, 0, 7, 0, 0, ict), time.Hour, 7 * time.Minute},
		{at(14, 10, 0), at(14, 37, 0), 30 * time.Minute, 7 * time.Minute},
		{at(14, 10, 0), at(14, 15, 0), 5 * time.Minute, 0},
	}
	for _, c := range cases {
		if got := NextRun(c.now, c.every, c.offset); !got.Equal(c.want) {
			t.Errorf("NextRun(%v, %v, %v) = %v, want %v", c.now, c.every, c.offset, got, c.want)
		}
	}
}

func TestRunLoopRunsAtStartThenOnSchedule(t *testing.T) {
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var waits []time.Duration
	in, err := New(Options{
		Source: h.src, Store: h.st, Logger: slog.New(slog.DiscardHandler),
		Now:   func() time.Time { return h.now },
		Every: time.Hour, Offset: 7 * time.Minute,
		Wait: func(ctx context.Context, d time.Duration) error {
			waits = append(waits, d)
			if len(waits) == 3 {
				cancel()
				return ctx.Err()
			}
			h.now = h.now.Add(d)
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := in.Run(ctx); !errors.Is(err, context.Canceled) {
		t.Errorf("Run = %v, want context.Canceled", err)
	}
	// 19:55 → 20:07, then hourly.
	want := []time.Duration{12 * time.Minute, time.Hour, time.Hour}
	if fmt.Sprint(waits) != fmt.Sprint(want) {
		t.Errorf("waits = %v, want %v", waits, want)
	}
	if n := countRuns(t, h.st); n != 3 {
		t.Errorf("runs recorded = %d, want 3 (start plus two scheduled)", n)
	}
}

// A failed run does not stop the loop.
func TestRunLoopSurvivesAFailedRun(t *testing.T) {
	h := newHarness(t)
	h.src.loadErr = errors.New("down")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	n := 0
	in, _ := New(Options{
		Source: h.src, Store: h.st, Logger: slog.New(slog.DiscardHandler), Now: func() time.Time { return h.now },
		Wait: func(ctx context.Context, d time.Duration) error {
			if n++; n == 2 {
				cancel()
				return ctx.Err()
			}
			return nil
		},
	})
	in.Run(ctx)
	if got := countRuns(t, h.st); got != 2 {
		t.Errorf("runs = %d, want 2", got)
	}
}

func countRuns(t *testing.T, st *store.Store) int {
	t.Helper()
	n, err := st.RunCount(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func TestOptionsAreValidated(t *testing.T) {
	h := newHarness(t)
	bad := []Options{
		{Store: h.st},
		{Source: h.src},
		{Source: h.src, Store: h.st, Every: time.Hour, Offset: time.Hour},
		{Source: h.src, Store: h.st, Every: time.Hour, Offset: -time.Minute},
		{Source: h.src, Store: h.st, Every: 30 * time.Second},
	}
	for i, o := range bad {
		if _, err := New(o); err == nil {
			t.Errorf("options %d accepted: %+v", i, o)
		}
	}
}
