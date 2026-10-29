// Package api serves the JSON API: the current state of every station in
// human terms, one station's history, and the national timeline.
//
// The /now and /timeline responses are built once per data change and then
// served as prepared payloads with a strong ETag and a gzip variant, so
// hashing and compression never run per request. The per-station history is
// small and built per request. Every response carries the security headers;
// /api/* is rate limited per client address.
//
// The wire formats are documented in docs/api.md, which is the contract the
// Phase 4 front end is written against.
package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/serlorsx-xln/mahanathi/server/internal/domain"
	"github.com/serlorsx-xln/mahanathi/server/internal/httpx"
	"github.com/serlorsx-xln/mahanathi/server/internal/store"
)

const (
	contentTypeJSON = "application/json; charset=utf-8"
	cacheControl    = "public, max-age=300"
	// readyAfter is how old the newest reading may be for /readyz: the
	// brief's "has data no older than 2 h".
	readyAfter = 2 * time.Hour
	// timelineHours is the one timeline window the product serves.
	timelineHours = 72
)

// Options configures the API. Everything except Store has a default; the
// clock, the logger and the rate limit knobs are for tests.
type Options struct {
	Store  *store.Store
	Logger *slog.Logger
	Now    func() time.Time
	// RateBurst is how many requests a client may send at once (default 30).
	RateBurst int
	// RateRefill is how often a client earns one request back
	// (default 200ms, five per second).
	RateRefill time.Duration
}

// Handler serves the API. Construct it, Register it on the server's mux,
// and call Refresh whenever the data changes.
type Handler struct {
	st  *store.Store
	log *slog.Logger
	now func() time.Time

	refreshMu sync.Mutex

	nowPayload      atomic.Pointer[httpx.Payload]
	timelinePayload atomic.Pointer[httpx.Payload]

	limiter *limiter
}

// New validates the options and returns a Handler.
func New(o Options) (*Handler, error) {
	if o.Store == nil {
		return nil, errNoStore
	}
	burst, refill := o.RateBurst, o.RateRefill
	if burst == 0 {
		burst = defaultBurst
	}
	if refill == 0 {
		refill = defaultRefill
	}
	if burst < 1 || refill < 0 {
		return nil, errBadRate
	}
	h := &Handler{st: o.Store, log: o.Logger, now: o.Now}
	if h.log == nil {
		h.log = slog.Default()
	}
	if h.now == nil {
		h.now = time.Now
	}
	h.limiter = newLimiter(burst, refill, h.now)
	return h, nil
}

// Register adds the API's routes to mux. The rest of the mux is not the
// API's business.
func (h *Handler) Register(mux *http.ServeMux) {
	api := http.NewServeMux()
	api.HandleFunc("GET /api/v1/now", h.serveNow)
	api.HandleFunc("GET /api/v1/timeline", h.serveTimeline)
	api.HandleFunc("GET /api/v1/stations/{id}/history", h.serveHistory)
	api.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		h.error(w, http.StatusNotFound, "no such endpoint")
	})
	// secured wraps limited (not the other way round) so the 429 reject branch
	// carries the security headers too, like every other API response.
	mux.Handle("/api/", secured(h.limited(api)))
	mux.Handle("GET /readyz", secured(http.HandlerFunc(h.serveReadyz)))
}

// Refresh rebuilds the /now and /timeline payloads from the store. It is
// safe to call concurrently; the last call wins.
func (h *Handler) Refresh(ctx context.Context) error {
	h.refreshMu.Lock()
	defer h.refreshMu.Unlock()

	now := h.now()
	stations, err := h.st.Stations(ctx)
	if err != nil {
		return err
	}
	readings, err := h.st.ReadingsSince(ctx, domain.HourEnd(now).Add(-timelineHours*time.Hour))
	if err != nil {
		return err
	}

	nowBody := marshal(BuildNow(stations, readings, now))
	h.nowPayload.Store(httpx.NewPayload(nowBody, contentTypeJSON, cacheControl))

	tl := encodeTimeline(buildTimeline(stations, readings, now, timelineHours))
	h.timelinePayload.Store(httpx.NewPayload(tl, timelineContentType, cacheControl))
	return nil
}

func (h *Handler) serveNow(w http.ResponseWriter, r *http.Request) {
	if p := h.nowPayload.Load(); p != nil {
		h.reply(w, r, p)
		return
	}
	w.Header().Set("Retry-After", "30")
	h.error(w, http.StatusServiceUnavailable, "no data yet")
}

func (h *Handler) serveTimeline(w http.ResponseWriter, r *http.Request) {
	if p := h.timelinePayload.Load(); p != nil {
		h.reply(w, r, p)
		return
	}
	w.Header().Set("Retry-After", "30")
	h.error(w, http.StatusServiceUnavailable, "no data yet")
}

// serveReadyz answers whether the store has a reading no older than 2 h.
// It reads the store, not the cached payloads, so it stays honest when the
// ingester has stopped.
func (h *Handler) serveReadyz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	stations, err := h.st.Stations(r.Context())
	if err != nil {
		h.log.Error("readyz: reading stations failed", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("store error\n"))
		return
	}
	var newest time.Time
	for _, st := range stations {
		if st.Latest != nil && st.Latest.At.After(newest) {
			newest = st.Latest.At
		}
	}
	if newest.IsZero() || h.now().Sub(newest) > readyAfter {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("stale data\n"))
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ready\n"))
}

// reply writes a prepared payload, or 304 when the client already has it.
func (h *Handler) reply(w http.ResponseWriter, r *http.Request, p *httpx.Payload) {
	p.ServeHTTP(w, r)
}

// error writes a small JSON error with no-store, so a failure is never
// cached anywhere.
func (h *Handler) error(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", contentTypeJSON)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(code)
	_, _ = w.Write(marshal(map[string]string{"error": msg}))
}

// marshal is json.Marshal with the panic path made explicit: every value the
// API marshals is a plain struct of numbers, strings and pointers.
func marshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

// secured adds the security headers every API response carries.
func secured(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}

// limited applies the per-client token bucket to a handler and answers 429
// with Retry-After when it is exhausted.
func (h *Handler) limited(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ok, retry := h.limiter.allow(clientKey(r)); !ok {
			secs := int(retry.Seconds() + 0.999)
			w.Header().Set("Retry-After", itoa(secs))
			h.error(w, http.StatusTooManyRequests, "too many requests")
			return
		}
		next.ServeHTTP(w, r)
	})
}
