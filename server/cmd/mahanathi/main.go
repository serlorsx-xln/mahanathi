// Command mahanathi serves the Mahanathi web app and its JSON API from a
// single binary.
//
// Configuration comes from the environment: PORT (or MAHANATHI_ADDR),
// MAHANATHI_DATA_DIR, MAHANATHI_POLL_EVERY and MAHANATHI_POLL_OFFSET, plus
// the logging and shutdown variables (see internal/config). The process
// shuts down gracefully on SIGINT or SIGTERM: it stops accepting connections
// and gives in-flight requests up to MAHANATHI_SHUTDOWN_TIMEOUT to finish.
//
// The one subcommand, `mahanathi healthcheck`, asks /readyz on the loopback
// interface and exits 0 or 1, so a distroless container can probe itself.
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
	"path/filepath"
	"syscall"
	"time"

	"github.com/serlorsx-xln/mahanathi/server/internal/api"
	"github.com/serlorsx-xln/mahanathi/server/internal/config"
	"github.com/serlorsx-xln/mahanathi/server/internal/ingest"
	"github.com/serlorsx-xln/mahanathi/server/internal/store"
	"github.com/serlorsx-xln/mahanathi/server/internal/thaiwater"
	"github.com/serlorsx-xln/mahanathi/server/web"
)

// version is injected at build time (-ldflags "-X main.version=…"); "dev"
// means it was not.
var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		os.Exit(runHealthcheck())
	}

	cfg, err := config.FromEnv(os.Getenv)
	if err != nil {
		fmt.Fprintln(os.Stderr, "mahanathi: config:", err)
		os.Exit(2)
	}
	logger := newLogger(cfg, os.Stdout)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		logger.Error("data directory unusable", "dir", cfg.DataDir, "err", err)
		os.Exit(1)
	}
	st, err := store.Open(ctx, filepath.Join(cfg.DataDir, "mahanathi.db"))
	if err != nil {
		logger.Error("opening the store failed", "dir", cfg.DataDir, "err", err)
		os.Exit(1)
	}
	defer st.Close()

	client, err := thaiwater.New(thaiwater.Options{
		BaseURL:   thaiwater.DefaultBaseURL,
		UserAgent: "mahanathi/" + version + " (+https://github.com/serlorsx-xln/mahanathi)",
	})
	if err != nil {
		logger.Error("building the ThaiWater client failed", "err", err)
		os.Exit(1)
	}

	apiHandler, err := api.New(api.Options{Store: st, Logger: logger})
	if err != nil {
		logger.Error("building the API failed", "err", err)
		os.Exit(1)
	}
	// Data from a previous run, if any, is served while the first ingest
	// runs; before that the API answers 503 with Retry-After.
	if err := apiHandler.Refresh(ctx); err != nil {
		logger.Warn("serving the previous data failed; the first ingest will retry", "err", err)
	}

	ing, err := ingest.New(ingest.Options{
		Source: client, Store: st, Logger: logger,
		Every: cfg.PollEvery, Offset: cfg.PollOffset,
		OnChange: func() {
			if err := apiHandler.Refresh(ctx); err != nil {
				logger.Warn("refreshing the API failed; serving the last good payloads", "err", err)
			}
		},
	})
	if err != nil {
		logger.Error("building the ingester failed", "err", err)
		os.Exit(1)
	}
	go func() {
		if err := ing.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			logger.Error("the ingester stopped", "err", err)
		}
	}()

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
		logger.Info("no web build embedded; / serves the placeholder page (run make build)")
	}
	logger.Info("starting", "version", version, "addr", cfg.Addr, "data", cfg.DataDir,
		"poll_every", cfg.PollEvery.String(), "poll_offset", cfg.PollOffset.String())
	if err := serve(ctx, newServer(logger, site, apiHandler), ln, logger, cfg.ShutdownTimeout); err != nil {
		logger.Error("server stopped with error", "err", err)
		os.Exit(1)
	}
}

// runHealthcheck asks this process's own /readyz. It exits 0 when the
// server is ready, 1 otherwise.
func runHealthcheck() int {
	cfg, err := config.FromEnv(os.Getenv)
	if err != nil {
		fmt.Fprintln(os.Stderr, "mahanathi: config:", err)
		return 1
	}
	_, port, err := net.SplitHostPort(cfg.Addr)
	if err != nil {
		port = "8080"
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("http://127.0.0.1:" + port + "/readyz")
	if err != nil {
		fmt.Fprintln(os.Stderr, "mahanathi: healthcheck:", err)
		return 1
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintln(os.Stderr, "mahanathi: healthcheck: /readyz:", resp.Status)
		return 1
	}
	return 0
}

func newLogger(cfg config.Config, w io.Writer) *slog.Logger {
	opts := &slog.HandlerOptions{Level: cfg.LogLevel}
	if cfg.LogFormat == "text" {
		return slog.New(slog.NewTextHandler(w, opts))
	}
	return slog.New(slog.NewJSONHandler(w, opts))
}

// newServer wires the HTTP routes: the probes, the JSON API, and the
// embedded web app at every other path.
func newServer(logger *slog.Logger, site http.Handler, apiHandler *api.Handler) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = io.WriteString(w, "ok\n")
	})
	apiHandler.Register(mux)
	mux.Handle("/", site)

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
