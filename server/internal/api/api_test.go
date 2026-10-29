package api

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/serlorsx-xln/mahanathi/server/internal/domain"
	"github.com/serlorsx-xln/mahanathi/server/internal/normalize"
	"github.com/serlorsx-xln/mahanathi/server/internal/store"
	"github.com/serlorsx-xln/mahanathi/server/internal/thaiwater"
)

const fixtures = "../thaiwater/testdata/"

// recorded is the same fake clock the ingest tests use: just after the
// history fixtures were written.
var recorded = time.Date(2026, 10, 29, 12, 55, 0, 0, time.UTC)

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

// populate fills a store the way a first ingest run would: the snapshot,
// then the five recorded histories, then the re-resolved snapshot.
func populate(t *testing.T) *store.Store {
	t.Helper()
	ctx := context.Background()
	st, err := store.Open(ctx, filepath.Join(t.TempDir(), "m.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })

	load := &thaiwater.LoadResponse{}
	readJSON(t, "waterlevel_load.json", load)
	res := normalize.Load(load, nil)
	if _, err := st.SaveObservations(ctx, res.Observations, recorded); err != nil {
		t.Fatal(err)
	}
	for _, id := range []int64{737, 1097520, 575568, 3441, 12479} {
		g := &thaiwater.GraphResponse{}
		readJSON(t, fmt.Sprintf("waterlevel_graph_%d.json", id), g)
		refs, readings, _ := normalize.Graph(id, g)
		if _, err := st.SaveHistory(ctx, id, refs, readings, recorded); err != nil {
			t.Fatal(err)
		}
	}
	refs, err := st.GraphRefs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.SaveObservations(ctx, normalize.Load(load, refs).Observations, recorded); err != nil {
		t.Fatal(err)
	}
	return st
}

// serve starts a test server over a populated store, with a settable clock.
func serve(t *testing.T, st *store.Store, now *time.Time, burst int, refill time.Duration) (*httptest.Server, *Handler) {
	t.Helper()
	h, err := New(Options{
		Store: st, Logger: slog.New(slog.DiscardHandler),
		Now:       func() time.Time { return *now },
		RateBurst: burst, RateRefill: refill,
	})
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	h.Register(mux)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, h
}

func get(t *testing.T, url string, headers map[string]string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func TestNowServesTheSnapshot(t *testing.T) {
	st := populate(t)
	now := recorded
	srv, h := serve(t, st, &now, 30, 200*time.Millisecond)
	if err := h.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}

	resp := get(t, srv.URL+"/api/v1/now", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	for k, want := range map[string]string{
		"Content-Type":            contentTypeJSON,
		"Cache-Control":           cacheControl,
		"X-Content-Type-Options":  "nosniff",
		"Referrer-Policy":         "no-referrer",
		"X-Frame-Options":         "DENY",
		"Content-Security-Policy": "default-src 'none'; frame-ancestors 'none'",
	} {
		if got := resp.Header.Get(k); got != want {
			t.Errorf("%s = %q, want %q", k, got, want)
		}
	}
	etag := resp.Header.Get("ETag")
	if etag == "" {
		t.Fatal("no ETag")
	}

	var body NowResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body.Stations) != 37 {
		t.Errorf("stations = %d, want 37", len(body.Stations))
	}
	if body.Summary.Stations != 37 || body.Summary.National == "" {
		t.Errorf("summary = %+v", body.Summary)
	}
	var bsp, urtu *StationNow
	for i := range body.Stations {
		switch body.Stations[i].ID {
		case 12479:
			bsp = &body.Stations[i]
		case 3441:
			urtu = &body.Stations[i]
		}
	}
	if bsp == nil || urtu == nil {
		t.Fatal("BSP002 or URTU09 missing")
	}

	// BSP002 (Khlong Khanan) is over its bank with full references.
	if bsp.Level == nil || *bsp.Level != 5 {
		t.Errorf("BSP002 level = %v, want 5", bsp.Level)
	}
	if !strings.Contains(bsp.Sentence, "ล้นตลิ่ง") {
		t.Errorf("BSP002 sentence = %q", bsp.Sentence)
	}
	if bsp.Trend6h == nil {
		t.Error("BSP002 has no trend, though its history is recorded")
	}
	// 16.727 m over bank 16.475, ground 10.959: (16727−10959)/(16475−10959)
	// is 104.57 % in hundredths-of-a-percent arithmetic.
	if bsp.Pct == nil || *bsp.Pct != 104.57 {
		t.Errorf("BSP002 pct = %v, want 104.57", bsp.Pct)
	}
	if bsp.BankFull == nil || *bsp.BankFull != 16.475 {
		t.Errorf("BSP002 bankFull = %v", bsp.BankFull)
	}

	// URTU09's level comes from its graph references (ADR 0003).
	if urtu.Level == nil || *urtu.Level != 4 {
		t.Errorf("URTU09 level = %v, want 4", urtu.Level)
	}
	if urtu.BankFull == nil || *urtu.BankFull != 225.386 {
		t.Errorf("URTU09 bankFull = %v, want 225.386", urtu.BankFull)
	}

	// The ETag is a real validator: the same body again is a 304.
	if resp := get(t, srv.URL+"/api/v1/now", map[string]string{"If-None-Match": etag}); resp.StatusCode != http.StatusNotModified {
		t.Errorf("If-None-Match status = %d", resp.StatusCode)
	}

	// gzip is negotiated, and the gzip body is the same JSON.
	resp = get(t, srv.URL+"/api/v1/now", map[string]string{"Accept-Encoding": "gzip"})
	if resp.Header.Get("Content-Encoding") != "gzip" || resp.Header.Get("Vary") != "Accept-Encoding" {
		t.Fatalf("content-encoding %q, vary %q", resp.Header.Get("Content-Encoding"), resp.Header.Get("Vary"))
	}
	zr, err := gzip.NewReader(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	var again NowResponse
	if err := json.NewDecoder(zr).Decode(&again); err != nil {
		t.Fatal(err)
	}
	if len(again.Stations) != len(body.Stations) || again.Summary.National != body.Summary.National {
		t.Error("the gzip variant is not the same document")
	}
}

func TestNowBeforeAnyRefreshIsServiceUnavailable(t *testing.T) {
	st := populate(t)
	now := recorded
	srv, _ := serve(t, st, &now, 30, 200*time.Millisecond)
	resp := get(t, srv.URL+"/api/v1/now", nil)
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if resp.Header.Get("Retry-After") == "" || resp.Header.Get("Cache-Control") != "no-store" {
		t.Errorf("retry-after %q, cache-control %q", resp.Header.Get("Retry-After"), resp.Header.Get("Cache-Control"))
	}
}

func TestNowGoesStaleTogether(t *testing.T) {
	st := populate(t)
	now := recorded.Add(4 * time.Hour)
	srv, h := serve(t, st, &now, 30, 200*time.Millisecond)
	if err := h.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	resp := get(t, srv.URL+"/api/v1/now", nil)
	var body NowResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Summary.StaleCount != 37 {
		t.Errorf("stale count = %d, want 37", body.Summary.StaleCount)
	}
	if body.Summary.National != "ยังไม่มีข้อมูลล่าสุดจากสถานีวัดระดับน้ำ" {
		t.Errorf("national = %q", body.Summary.National)
	}
	for _, s := range body.Stations {
		if !s.Stale {
			t.Fatalf("station %d not marked stale", s.ID)
		}
	}
}

func TestTimelineEncoding(t *testing.T) {
	st := populate(t)
	now := recorded
	srv, h := serve(t, st, &now, 30, 200*time.Millisecond)
	if err := h.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}

	resp := get(t, srv.URL+"/api/v1/timeline", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != timelineContentType {
		t.Fatalf("content-type %q", ct)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	tl := decodeTimeline(t, body)
	if tl.Hours != 72 || len(tl.Values) != len(tl.Stations) {
		t.Fatalf("hours %d, %d stations vs %d value rows", tl.Hours, len(tl.Stations), len(tl.Values))
	}
	// now is 12:55 UTC, so the buckets run from the 14:00 hour end two
	// days back to the 13:00 hour end.
	wantFrom := time.Date(2026, 10, 26, 14, 0, 0, 0, time.UTC)
	if !tl.From.Equal(wantFrom) {
		t.Errorf("from = %v, want %v", tl.From, wantFrom)
	}
	var bspAt int = -1
	for i, id := range tl.Stations {
		if len(tl.Values[i]) != 72 {
			t.Fatalf("station %d has %d values", id, len(tl.Values[i]))
		}
		if id == 12479 {
			bspAt = i
		}
	}
	if bspAt < 0 {
		t.Fatal("BSP002 missing from the timeline")
	}
	last := tl.Values[bspAt][71]
	if last == nil {
		t.Fatal("BSP002 last bucket is empty")
	}
	if *last != 10457 { // 104.57 %, as in the /now test
		t.Errorf("BSP002 last bucket = %d, want 10457", *last)
	}
}

// decodeTimeline reads the MNTL format back, the way docs/api.md describes
// it for the front end. If this cannot read what encodeTimeline wrote, the
// documentation is wrong.
func decodeTimeline(t *testing.T, b []byte) timelineData {
	t.Helper()
	if len(b) < 20 || string(b[:4]) != timelineMagic {
		t.Fatalf("not an MNTL body (first bytes %x)", b[:4])
	}
	if v := b[4]; v != timelineVersion {
		t.Fatalf("timeline version %d", v)
	}
	tl := timelineData{Hours: int(b[5])}
	tl.From = time.Unix(int64(binary.LittleEndian.Uint64(b[8:16])), 0).UTC()
	n := int(binary.LittleEndian.Uint32(b[16:20]))
	p := 20
	tl.Stations = make([]int64, n)
	for i := range tl.Stations {
		tl.Stations[i] = int64(binary.LittleEndian.Uint64(b[p:]))
		p += 8
	}
	bitmapLen := (tl.Hours + 7) / 8
	tl.Values = make([][]*uint16, n)
	for i := range tl.Values {
		bitmap := b[p : p+bitmapLen]
		p += bitmapLen
		vals := make([]*uint16, tl.Hours)
		prev, run := int64(0), false
		for j := 0; j < tl.Hours; j++ {
			if bitmap[j/8]&(1<<(7-j%8)) == 0 {
				run = false
				continue
			}
			v, used := binary.Uvarint(b[p:])
			if used <= 0 {
				t.Fatalf("station %d bucket %d: bad varint", tl.Stations[i], j)
			}
			p += used
			var x int64
			if run {
				x = prev + (int64(v>>1) ^ -int64(v&1)) // unzigzag
			} else {
				x = int64(v)
			}
			u := uint16(x)
			vals[j] = &u
			prev, run = x, true
		}
		tl.Values[i] = vals
	}
	if p != len(b) {
		t.Fatalf("decoded %d of %d bytes", p, len(b))
	}
	tl.Until = tl.From.Add(time.Duration(tl.Hours-1) * time.Hour)
	return tl
}

// The brief's budget: 800 stations × 72 hourly values in at most 80 KB
// compressed. The synthetic water moves ±0.15 m every hour on a 20 m bank,
// which is wilder than any real river sustains for three days, so the real
// payload has more headroom than this test measures.
func TestTimelineStaysWithinTheBudget(t *testing.T) {
	rng := rand.New(rand.NewSource(29))
	until := time.Date(2026, 10, 29, 13, 0, 0, 0, time.UTC)
	stations := make([]store.StationState, 800)
	readings := map[int64][]domain.Reading{}
	for i := range stations {
		id := int64(1000 + i)
		stations[i] = store.StationState{Station: stationComplete(id)}
		rs := make([]domain.Reading, 72)
		wl := 100.0 + rng.Float64()*20
		for j := range rs {
			wl += (rng.Float64() - 0.5) * 0.3
			rs[j] = domain.Reading{At: until.Add(time.Duration(j-71) * time.Hour), WL: wl}
		}
		readings[id] = rs
	}
	body := encodeTimeline(buildTimeline(stations, readings, until, 72))
	var buf bytes.Buffer
	zw, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	zw.Write(body)
	zw.Close()
	t.Logf("800×72 timeline: %d raw, %d compressed", len(body), buf.Len())
	if buf.Len() > 80*1024 {
		t.Errorf("800×72 timeline is %d bytes compressed, budget 81920", buf.Len())
	}
}

func TestHistoryEndpoint(t *testing.T) {
	st := populate(t)
	now := recorded
	srv, h := serve(t, st, &now, 30, 200*time.Millisecond)
	if err := h.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}

	resp := get(t, srv.URL+"/api/v1/stations/3441/history", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	var hist HistoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&hist); err != nil {
		t.Fatal(err)
	}
	if hist.Hours != 72 || len(hist.Points) != 72 {
		t.Fatalf("hours %d, points %d", hist.Hours, len(hist.Points))
	}
	if hist.Station.BankFull == nil || *hist.Station.BankFull != 225.386 {
		t.Errorf("bankFull = %v", hist.Station.BankFull)
	}
	for i := 1; i < len(hist.Points); i++ {
		if !hist.Points[i].At.After(hist.Points[i-1].At) {
			t.Fatalf("points not ascending at %d", i)
		}
	}

	for _, tc := range []struct {
		url  string
		code int
	}{
		{srv.URL + "/api/v1/stations/3441/history?hours=0", 400},
		{srv.URL + "/api/v1/stations/3441/history?hours=169", 400},
		{srv.URL + "/api/v1/stations/3441/history?hours=soon", 400},
		{srv.URL + "/api/v1/stations/999999/history", 404},
		{srv.URL + "/api/v1/stations/notanumber/history", 400},
		{srv.URL + "/api/v1/stations/-1/history", 400},
		{srv.URL + "/api/v1/nosuch", 404},
	} {
		if resp := get(t, tc.url, nil); resp.StatusCode != tc.code {
			t.Errorf("%s: status %d, want %d", tc.url, resp.StatusCode, tc.code)
		}
	}

	// The unknown endpoint is JSON, and carries the security headers too.
	resp = get(t, srv.URL+"/api/v1/nosuch", nil)
	if resp.Header.Get("Content-Type") != contentTypeJSON || resp.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("404 content-type %q, nosniff %q", resp.Header.Get("Content-Type"), resp.Header.Get("X-Content-Type-Options"))
	}
}

func TestReadyz(t *testing.T) {
	ctx := context.Background()
	st, err := store.Open(ctx, filepath.Join(t.TempDir(), "m.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	// The fixture's newest reading is around 19:40 ICT; recorded is 19:55,
	// which is already past the 2 h readiness bound, so the fresh checks run
	// at 18:55 ICT.
	fresh := recorded.Add(-time.Hour)
	now := fresh
	srv, _ := serve(t, st, &now, 30, 200*time.Millisecond)

	// Empty store: not ready, and not rate limited away from the probe.
	if resp := get(t, srv.URL+"/readyz", nil); resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("empty store: status %d", resp.StatusCode)
	}

	load := &thaiwater.LoadResponse{}
	readJSON(t, "waterlevel_load.json", load)
	res := normalize.Load(load, nil)
	if _, err := st.SaveObservations(ctx, res.Observations, recorded); err != nil {
		t.Fatal(err)
	}
	if resp := get(t, srv.URL+"/readyz", nil); resp.StatusCode != http.StatusOK {
		t.Fatalf("fresh data: status %d", resp.StatusCode)
	}

	now = fresh.Add(3 * time.Hour)
	if resp := get(t, srv.URL+"/readyz", nil); resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("3 h old data: status %d", resp.StatusCode)
	}
}

func TestRateLimit(t *testing.T) {
	st := populate(t)
	now := recorded
	srv, _ := serve(t, st, &now, 3, time.Hour)

	for i := 0; i < 3; i++ {
		if resp := get(t, srv.URL+"/api/v1/now", nil); resp.StatusCode != http.StatusServiceUnavailable {
			t.Fatalf("request %d: status %d (503 is the no-data answer; a 429 would mean the limit hit early)", i, resp.StatusCode)
		}
	}
	resp := get(t, srv.URL+"/api/v1/now", nil)
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("fourth request: status %d", resp.StatusCode)
	}
	if ra := resp.Header.Get("Retry-After"); ra == "" {
		t.Error("429 without Retry-After")
	} else if ra != "3600" {
		t.Errorf("Retry-After = %q, want 3600", ra)
	}
	if resp.Header.Get("Cache-Control") != "no-store" {
		t.Error("429 is cacheable")
	}

	// Half an hour later the bucket has half a token: still no.
	now = recorded.Add(30 * time.Minute)
	if resp := get(t, srv.URL+"/api/v1/now", nil); resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("after 30 min: status %d", resp.StatusCode)
	}
	// A full refill later, one request goes through again.
	now = recorded.Add(time.Hour)
	if resp := get(t, srv.URL+"/api/v1/now", nil); resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("after 1 h: status %d", resp.StatusCode)
	}
	now = recorded.Add(2 * time.Hour)
	if resp := get(t, srv.URL+"/api/v1/now", nil); resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("after 2 h: status %d", resp.StatusCode)
	}

	// /readyz is not rate limited: the probe always gets its real answer.
	if resp := get(t, srv.URL+"/readyz", nil); resp.StatusCode == http.StatusTooManyRequests {
		t.Fatal("readyz was rate limited")
	}
}

// stationComplete is a station with usable references for the budget test.
func stationComplete(id int64) domain.Station {
	return domain.Station{ID: id, Code: fmt.Sprintf("T.%d", id), Refs: domain.RefsComplete,
		BankFull: 120, Ground: 100, Lat: 16, Lon: 100}
}

func TestClientKeyTrustsOnlyOurProxy(t *testing.T) {
	r := &http.Request{RemoteAddr: "203.0.113.9:4444", Header: http.Header{}}
	r.Header.Set("X-Forwarded-For", "198.51.100.7")
	if k := clientKey(r); k != "203.0.113.9" {
		t.Errorf("public peer: key = %q", k)
	}
	r.RemoteAddr = "10.0.0.2:4444"
	if k := clientKey(r); k != "198.51.100.7" {
		t.Errorf("behind our proxy: key = %q", k)
	}
	r.Header.Set("X-Forwarded-For", "1.2.3.4, 198.51.100.7")
	if k := clientKey(r); k != "198.51.100.7" {
		t.Errorf("chained header: key = %q", k)
	}
}
