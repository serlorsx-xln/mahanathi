package thaiwater

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// recorder captures the backoff sleeps instead of sleeping.
type recorder struct {
	mu     sync.Mutex
	sleeps []time.Duration
}

func (r *recorder) sleep(ctx context.Context, d time.Duration) error {
	r.mu.Lock()
	r.sleeps = append(r.sleeps, d)
	r.mu.Unlock()
	return ctx.Err()
}

func newTestClient(t *testing.T, url string, mod func(*Options)) (*Client, *recorder) {
	t.Helper()
	rec := &recorder{}
	o := Options{
		BaseURL:      url,
		UserAgent:    "mahanathi/test (+https://github.com/serlorsx-xln/mahanathi)",
		MaxAttempts:  4,
		BaseDelay:    time.Second,
		MaxDelay:     8 * time.Second,
		LoadTimeout:  2 * time.Second,
		GraphTimeout: 2 * time.Second,
		Sleep:        rec.sleep,
		Jitter:       func() float64 { return 0.5 },
	}
	if mod != nil {
		mod(&o)
	}
	c, err := New(o)
	if err != nil {
		t.Fatal(err)
	}
	return c, rec
}

func TestLoadParsesTheRecordedSnapshot(t *testing.T) {
	body := fixture(t, "waterlevel_load.json")
	var gotUA, gotAccept, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA, gotAccept, gotPath = r.UserAgent(), r.Header.Get("Accept"), r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer srv.Close()
	c, _ := newTestClient(t, srv.URL, nil)

	resp, err := c.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/v1/thaiwater30/public/waterlevel_load" {
		t.Errorf("path = %q", gotPath)
	}
	if gotUA != "mahanathi/test (+https://github.com/serlorsx-xln/mahanathi)" {
		t.Errorf("User-Agent = %q", gotUA)
	}
	if gotAccept != "application/json" {
		t.Errorf("Accept = %q", gotAccept)
	}

	rows := resp.WaterlevelData.Data
	if len(rows) != 37 {
		t.Fatalf("rows = %d, want 37", len(rows))
	}
	byCode := map[string]LoadRow{}
	for _, r := range rows {
		byCode[r.Station.OldCode] = r
	}

	// วัดท่าเจดีย์ (TTC06): the first row, every field present.
	r := byCode["THA006"]
	if r.Station.ID != 737 || r.Datetime != "2026-10-29 17:40" {
		t.Errorf("THA006 id/time = %d %q", r.Station.ID, r.Datetime)
	}
	if !r.WaterlevelMSL.Valid || r.WaterlevelMSL.Value != 3.17 {
		t.Errorf("THA006 waterlevel_msl = %+v", r.WaterlevelMSL)
	}
	if r.Station.MinBank.Value != 2.14 || r.Station.GroundLevel.Value != -2.065 {
		t.Errorf("THA006 refs = %+v %+v", r.Station.MinBank, r.Station.GroundLevel)
	}
	if r.SituationLevel == nil || *r.SituationLevel != 5 {
		t.Errorf("THA006 situation_level = %v", r.SituationLevel)
	}
	if r.Station.Name.TH != "วัดท่าเจดีย์ (TTC06)" || r.Geocode.Province.TH != "สุพรรณบุรี" ||
		r.Geocode.Amphoe.TH != "สองพี่น้อง" || r.Basin.Name.TH != "ลุ่มน้ำท่าจีน" ||
		r.RiverName != "แม่น้ำท่าจีน" || r.Agency.Short.EN != "HII" {
		t.Errorf("THA006 names = %+v %+v %+v %q %+v", r.Station.Name, r.Geocode, r.Basin, r.RiverName, r.Agency)
	}
	if r.Station.Lat.Value != 14.15657 || r.Station.Lon.Value != 100.12744 {
		t.Errorf("THA006 position = %v %v", r.Station.Lat, r.Station.Lon)
	}

	// Gt.9: ground_level is absent, not null.
	if g := byCode["Gt.9"].Station.GroundLevel; g.Present || g.Valid {
		t.Errorf("Gt.9 ground_level = %+v, want absent", g)
	}
	// BLGTU06: both references coded as 0, no situation_level key, null percent.
	b := byCode["BLGTU06"]
	if !b.Station.MinBank.Valid || b.Station.MinBank.Value != 0 || b.SituationLevel != nil || b.StoragePercent.Valid {
		t.Errorf("BLGTU06 = %+v %v %+v", b.Station.MinBank, b.SituationLevel, b.StoragePercent)
	}
	// C.67: no river_name key.
	if byCode["C.67"].RiverName != "" {
		t.Errorf("C.67 river = %q", byCode["C.67"].RiverName)
	}

	want := []Rule{{">", "100", 5}, {">", "70", 4}, {">", "30", 3}, {">", "10", 2}, {"<=", "10", 1}}
	got := resp.Scale.Data.RuleWeb
	if len(got) != len(want) {
		t.Fatalf("rule_web = %+v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("rule_web[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestGraphSendsTheDocumentedQuery(t *testing.T) {
	body := fixture(t, "waterlevel_graph_1097520.json")
	var q map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/thaiwater30/public/waterlevel_graph" {
			http.NotFound(w, r)
			return
		}
		q = map[string]string{}
		for k, v := range r.URL.Query() {
			q[k] = strings.Join(v, ",")
		}
		_, _ = w.Write(body)
	}))
	defer srv.Close()
	c, _ := newTestClient(t, srv.URL, nil)

	// 2026-10-26 17:30 UTC is already 2026-10-27 00:30 in Bangkok, so the
	// start date must be the 27th: the endpoint takes ICT calendar dates.
	from := time.Date(2026, 10, 26, 17, 30, 0, 0, time.UTC)
	to := time.Date(2026, 10, 29, 10, 54, 0, 0, time.UTC)
	g, err := c.Graph(context.Background(), 1097520, from, to)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"station_type": "tele_waterlevel",
		"station_id":   "1097520",
		"start_date":   "2026-10-27",
		"end_date":     "2026-10-29",
	}
	for k, v := range want {
		if q[k] != v {
			t.Errorf("query %s = %q, want %q", k, q[k], v)
		}
	}
	if len(q) != len(want) {
		t.Errorf("query = %v, want exactly %v", q, want)
	}

	if len(g.Data.GraphData) != 90 {
		t.Fatalf("graph points = %d, want 90", len(g.Data.GraphData))
	}
	if p := g.Data.GraphData[0]; p.Datetime != "2026-10-26 00:00" || p.Value.Valid {
		t.Errorf("first point = %+v, want a null at 00:00", p)
	}
	if p := g.Data.GraphData[1]; !p.Value.Valid || p.Value.Value != 38.269998 {
		t.Errorf("second point = %+v", p)
	}
	if g.Data.MinBank.Value != 40.599998 || g.Data.GroundLevel.Value != 37.72 {
		t.Errorf("graph refs = %+v %+v", g.Data.MinBank, g.Data.GroundLevel)
	}
}

func TestRetriesTransientFailuresWithJitteredBackoff(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch calls.Add(1) {
		case 1:
			http.Error(w, "busy", http.StatusBadGateway)
		case 2:
			http.Error(w, "busy", http.StatusServiceUnavailable)
		default:
			_, _ = w.Write(fixture(t, "waterlevel_graph_737.json"))
		}
	}))
	defer srv.Close()
	c, rec := newTestClient(t, srv.URL, nil)

	if _, err := c.Graph(context.Background(), 737, time.Now(), time.Now()); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 3 {
		t.Errorf("calls = %d, want 3", calls.Load())
	}
	// Full jitter with the test's jitter fixed at 0.5: half of 1 s, then half
	// of 2 s.
	want := []time.Duration{500 * time.Millisecond, time.Second}
	if len(rec.sleeps) != len(want) {
		t.Fatalf("sleeps = %v, want %v", rec.sleeps, want)
	}
	for i := range want {
		if rec.sleeps[i] != want[i] {
			t.Errorf("sleep %d = %v, want %v", i, rec.sleeps[i], want[i])
		}
	}
}

func TestBackoffIsCappedAndJitterSpansTheWindow(t *testing.T) {
	c, _ := newTestClient(t, "http://unused", func(o *Options) { o.Jitter = func() float64 { return 0.999 } })
	for attempt, ceiling := range map[int]time.Duration{0: time.Second, 1: 2 * time.Second, 2: 4 * time.Second, 3: 8 * time.Second, 10: 8 * time.Second} {
		got := c.backoff(attempt)
		if got > ceiling || got < ceiling*99/100 {
			t.Errorf("backoff(%d) = %v, want just under %v", attempt, got, ceiling)
		}
	}
	c, _ = newTestClient(t, "http://unused", func(o *Options) { o.Jitter = func() float64 { return 0 } })
	if got := c.backoff(3); got != 0 {
		t.Errorf("backoff with zero jitter = %v, want 0 (full jitter)", got)
	}
}

func TestDoesNotRetryClientErrors(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		http.Error(w, "no", http.StatusNotFound)
	}))
	defer srv.Close()
	c, _ := newTestClient(t, srv.URL, nil)

	_, err := c.Load(context.Background())
	var he *HTTPError
	if !errors.As(err, &he) || he.Status != http.StatusNotFound {
		t.Fatalf("err = %v, want an HTTPError 404", err)
	}
	if calls.Load() != 1 {
		t.Errorf("calls = %d, want 1", calls.Load())
	}
}

func TestGivesUpAfterMaxAttempts(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		http.Error(w, "down", http.StatusInternalServerError)
	}))
	defer srv.Close()
	c, rec := newTestClient(t, srv.URL, nil)

	_, err := c.Load(context.Background())
	if err == nil || !strings.Contains(err.Error(), "4 attempts") {
		t.Fatalf("err = %v, want it to report 4 attempts", err)
	}
	if calls.Load() != 4 || len(rec.sleeps) != 3 {
		t.Errorf("calls = %d, sleeps = %d; want 4 and 3", calls.Load(), len(rec.sleeps))
	}
}

func TestHonoursRetryAfter(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.Header().Set("Retry-After", "3")
			http.Error(w, "slow down", http.StatusTooManyRequests)
			return
		}
		_, _ = w.Write(fixture(t, "waterlevel_graph_737.json"))
	}))
	defer srv.Close()
	c, rec := newTestClient(t, srv.URL, nil)

	if _, err := c.Graph(context.Background(), 737, time.Now(), time.Now()); err != nil {
		t.Fatal(err)
	}
	if len(rec.sleeps) != 1 || rec.sleeps[0] != 3*time.Second {
		t.Errorf("sleeps = %v, want [3s] from Retry-After", rec.sleeps)
	}
}

func TestRetriesWhenTheBodyIsNotOK(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch calls.Add(1) {
		case 1:
			_, _ = w.Write([]byte(`{"result":"NO","data":null}`))
		case 2:
			_, _ = w.Write([]byte(`<html>maintenance</html>`))
		default:
			_, _ = w.Write(fixture(t, "waterlevel_graph_737.json"))
		}
	}))
	defer srv.Close()
	c, _ := newTestClient(t, srv.URL, nil)
	if _, err := c.Graph(context.Background(), 737, time.Now(), time.Now()); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 3 {
		t.Errorf("calls = %d, want 3", calls.Load())
	}
}

func TestLoadRejectsAnEmptyOrFailedSnapshot(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"waterlevel_data":{"result":"OK","data":[]}}`))
	}))
	defer srv.Close()
	c, _ := newTestClient(t, srv.URL, func(o *Options) { o.MaxAttempts = 1 })
	if _, err := c.Load(context.Background()); err == nil {
		t.Fatal("an empty snapshot was accepted; it would mark every station as gone")
	}
}

func TestPerAttemptTimeoutIsRetried(t *testing.T) {
	var calls atomic.Int32
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			select {
			case <-release:
			case <-r.Context().Done():
			}
			return
		}
		_, _ = w.Write(fixture(t, "waterlevel_graph_737.json"))
	}))
	defer srv.Close()
	defer close(release)
	c, _ := newTestClient(t, srv.URL, func(o *Options) { o.GraphTimeout = 100 * time.Millisecond })

	if _, err := c.Graph(context.Background(), 737, time.Now(), time.Now()); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 2 {
		t.Errorf("calls = %d, want 2", calls.Load())
	}
}

func TestCancelStopsRetrying(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		http.Error(w, "down", http.StatusBadGateway)
	}))
	defer srv.Close()
	ctx, cancel := context.WithCancel(context.Background())
	c, _ := newTestClient(t, srv.URL, func(o *Options) {
		o.Sleep = func(context.Context, time.Duration) error { cancel(); return context.Canceled }
	})
	_, err := c.Load(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	if calls.Load() != 1 {
		t.Errorf("calls = %d, want 1", calls.Load())
	}
}

func TestGraphConcurrencyIsLimited(t *testing.T) {
	var inFlight, peak atomic.Int32
	body := fixture(t, "waterlevel_graph_737.json")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := inFlight.Add(1)
		for {
			p := peak.Load()
			if n <= p || peak.CompareAndSwap(p, n) {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
		inFlight.Add(-1)
		_, _ = w.Write(body)
	}))
	defer srv.Close()
	c, _ := newTestClient(t, srv.URL, func(o *Options) { o.GraphConcurrency = 3 })

	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := c.Graph(context.Background(), 737, time.Now(), time.Now()); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if peak.Load() > 3 {
		t.Errorf("peak concurrent graph requests = %d, want at most 3", peak.Load())
	}
	if peak.Load() < 2 {
		t.Errorf("peak = %d; the limiter serialised everything", peak.Load())
	}
}

func TestOptionsAreValidated(t *testing.T) {
	bad := []Options{
		{BaseURL: "http://x", UserAgent: "ua", GraphConcurrency: 5},
		{BaseURL: "", UserAgent: "ua"},
		{BaseURL: "http://x", UserAgent: ""},
		{BaseURL: "ftp://x", UserAgent: "ua"},
	}
	for _, o := range bad {
		if _, err := New(o); err == nil {
			t.Errorf("New(%+v) accepted invalid options", o)
		}
	}
	c, err := New(Options{BaseURL: "https://api-v3.thaiwater.net/", UserAgent: "ua"})
	if err != nil {
		t.Fatal(err)
	}
	if c.opts.GraphConcurrency != 4 || c.opts.MaxAttempts != 4 || c.opts.BaseURL != "https://api-v3.thaiwater.net" {
		t.Errorf("defaults = %+v", c.opts)
	}
}

func TestBodyLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(fixture(t, "waterlevel_load.json"))
	}))
	defer srv.Close()
	c, _ := newTestClient(t, srv.URL, func(o *Options) { o.MaxBodyBytes = 1000; o.MaxAttempts = 1 })
	if _, err := c.Load(context.Background()); err == nil || !strings.Contains(err.Error(), "larger than") {
		t.Fatalf("err = %v, want a body-size error", err)
	}
}
