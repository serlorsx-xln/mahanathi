package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/serlorsx-xln/mahanathi/server/internal/domain"
	"github.com/serlorsx-xln/mahanathi/server/internal/httpx"
	"github.com/serlorsx-xln/mahanathi/server/internal/normalize"
)

// The history endpoint's bounds: anything a chart could reasonably ask of a
// 30-day store, while staying too small to matter for the size budget.
const (
	defaultHistoryHours = 72
	maxHistoryHours     = 168
)

// HistoryResponse is GET /api/v1/stations/{id}/history: one station's hourly
// buckets and the references they are read against.
type HistoryResponse struct {
	Station HistoryStation `json:"station"`
	From    time.Time      `json:"from"`
	Until   time.Time      `json:"until"`
	Hours   int            `json:"hours"`
	// Points[i] is the hour ending at From + i hours. WL is nil when that
	// hour has no reading.
	Points []HistoryPoint `json:"points"`
}

// HistoryStation is the station block of a history response.
type HistoryStation struct {
	ID       int64    `json:"id"`
	Code     string   `json:"code"`
	Name     string   `json:"name"`
	BankFull *float64 `json:"bankFull"`
	Ground   *float64 `json:"ground"`
}

// HistoryPoint is one hourly bucket.
type HistoryPoint struct {
	At time.Time `json:"at"`
	WL *float64  `json:"wl"`
}

// serveHistory answers GET /api/v1/stations/{id}/history?hours=72.
func (h *Handler) serveHistory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		h.error(w, http.StatusBadRequest, "the station id must be a positive number")
		return
	}
	hours := defaultHistoryHours
	if v := r.URL.Query().Get("hours"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > maxHistoryHours {
			h.error(w, http.StatusBadRequest, "hours must be a whole number from 1 to 168")
			return
		}
		hours = n
	}

	ctx := r.Context()
	st, ok, err := h.st.Station(ctx, id)
	if err != nil {
		h.error(w, http.StatusInternalServerError, "the store could not be read")
		h.log.Error("history: station lookup failed", "station", id, "error", err)
		return
	}
	if !ok {
		h.error(w, http.StatusNotFound, "no such station")
		return
	}

	now := h.now()
	until := domain.HourEnd(now)
	readings, err := h.st.History(ctx, id, until.Add(-time.Duration(hours)*time.Hour), until)
	if err != nil {
		h.error(w, http.StatusInternalServerError, "the store could not be read")
		h.log.Error("history: query failed", "station", id, "error", err)
		return
	}
	pts := normalize.Hourly(readings, until, hours)

	res := HistoryResponse{
		Station: HistoryStation{ID: st.Station.ID, Code: st.Station.Code, Name: st.Station.Name},
		From:    until.Add(-time.Duration(hours-1) * time.Hour),
		Until:   until,
		Hours:   hours,
		Points:  make([]HistoryPoint, hours),
	}
	if st.Station.HasBankFull() {
		b := st.Station.BankFull
		res.Station.BankFull = &b
	}
	if st.Station.HasGround() {
		g := st.Station.Ground
		res.Station.Ground = &g
	}
	for i, p := range pts {
		res.Points[i].At = p.At
		if p.OK {
			wl := p.WL
			res.Points[i].WL = &wl
		}
	}
	h.reply(w, r, httpx.NewPayload(marshal(res), contentTypeJSON, cacheControl))
}
