// Command mahanathi serves the Mahanathi web app and its JSON API from a
// single binary.
//
// Configuration comes from MAHANATHI_* environment variables (see
// internal/config). The process shuts down gracefully on SIGINT or SIGTERM:
// it stops accepting connections and gives in-flight requests up to
// MAHANATHI_SHUTDOWN_TIMEOUT to finish.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/serlorsx-xln/mahanathi/server/internal/config"
	"github.com/serlorsx-xln/mahanathi/server/web"
)

func main() {
	cfg, err := config.FromEnv(os.Getenv)
	if err != nil {
		fmt.Fprintln(os.Stderr, "mahanathi: config:", err)
		os.Exit(2)
	}
	logger := newLogger(cfg, os.Stdout)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ln, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		logger.Error("listen failed", "addr", cfg.Addr, "err", err)
		os.Exit(1)
	}
	site, err := web.NewHandler(web.Dist())
	if err != nil {
		logger.Error("embedded web build unreadable", "err", err)
		os.Exit(1)
	}
	if !site.Built() {
		logger.Warn("no web build embedded; / serves a notice (run make build)")
	}
	if err := serve(ctx, newServer(logger, site), ln, logger, cfg.ShutdownTimeout); err != nil {
		logger.Error("server stopped with error", "err", err)
		os.Exit(1)
	}
}

func newLogger(cfg config.Config, w io.Writer) *slog.Logger {
	opts := &slog.HandlerOptions{Level: cfg.LogLevel}
	if cfg.LogFormat == "text" {
		return slog.New(slog.NewTextHandler(w, opts))
	}
	return slog.New(slog.NewJSONHandler(w, opts))
}

// newServer wires the HTTP routes: the liveness probe and the embedded web
// app at every other path.
func newServer(logger *slog.Logger, site http.Handler) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/", site)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = io.WriteString(w, "ok\n")
	})

	return &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelWarn),
	}
}

// serve runs srv on ln until ctx is cancelled, then shuts it down, waiting up
// to timeout for open requests. It returns nil after a clean shutdown.
func serve(ctx context.Context, srv *http.Server, ln net.Listener, logger *slog.Logger, timeout time.Duration) error {
	errc := make(chan error, 1)
	go func() {
		logger.Info("listening", "addr", ln.Addr().String())
		errc <- srv.Serve(ln)
	}()

	select {
	case err := <-errc:
		// Serve only returns early on a real failure.
		return err
	case <-ctx.Done():
	}

	logger.Info("shutting down", "timeout", timeout.String())
	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	if err := <-errc; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	logger.Info("stopped")
	return nil
}
