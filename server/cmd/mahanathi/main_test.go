package main

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/serlorsx-xln/mahanathi/server/web"
)

// TestServeHealthzAndGracefulShutdown starts the real server on a free port,
// checks the liveness probe, then cancels the context the way SIGTERM would
// and expects a clean return.
func TestServeHealthzAndGracefulShutdown(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	site, err := web.NewHandler(web.Dist())
	if err != nil {
		t.Fatal(err)
	}
	go func() { done <- serve(ctx, newServer(logger, site), ln, logger, 2*time.Second) }()

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
