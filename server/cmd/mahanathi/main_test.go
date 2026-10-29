package main

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/serlorsx-xln/mahanathi/server/internal/api"
	"github.com/serlorsx-xln/mahanathi/server/internal/store"
	"github.com/serlorsx-xln/mahanathi/server/web"
)

// newTestServer wires the routes against a throwaway store, the way main
// does, without the ingester (nothing is fetched here).
func newTestServer(t *testing.T, logger *slog.Logger) *http.Server {
	t.Helper()
	st, err := store.Open(context.Background(), filepath.Join(t.TempDir(), "m.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	apiHandler, err := api.New(api.Options{Store: st, Logger: logger})
	if err != nil {
		t.Fatal(err)
	}
	site, err := web.NewHandler(web.Dist())
	if err != nil {
		t.Fatal(err)
	}
	return newServer(logger, site, apiHandler)
}

// TestServeHealthzAndGracefulShutdown starts the real server on a free port,
// checks the probes and the placeholder route, then cancels the context the
// way SIGTERM would and expects a clean return.
func TestServeHealthzAndGracefulShutdown(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- serve(ctx, newTestServer(t, logger), ln, logger, 2*time.Second) }()

	resp, err := http.Get("http://" + ln.Addr().String() + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK || string(body) != "ok\n" {
		t.Fatalf("GET /healthz = %d %q, want 200 \"ok\\n\"", resp.StatusCode, body)
	}

	// Whether or not a front-end build is embedded, / answers with HTML.
	resp, err = http.Get("http://" + ln.Addr().String() + "/")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET / = %d, want 200", resp.StatusCode)
	}

	// Before the first ingest: readyz says no, /now says retry later.
	resp, err = http.Get("http://" + ln.Addr().String() + "/readyz")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("GET /readyz = %d, want 503 on an empty store", resp.StatusCode)
	}
	resp, err = http.Get("http://" + ln.Addr().String() + "/api/v1/now")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable || resp.Header.Get("Retry-After") == "" {
		t.Fatalf("GET /api/v1/now = %d retry-after %q, want 503 with Retry-After", resp.StatusCode, resp.Header.Get("Retry-After"))
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("serve returned %v after shutdown, want nil", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("serve did not return after context cancel")
	}
}
