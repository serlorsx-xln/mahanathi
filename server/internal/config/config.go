// Package config reads the server's settings from environment variables.
//
// Every setting has a default that works for local development, so the binary
// runs with no environment at all. Invalid values are an error at start-up
// rather than a silent fallback, because a typo in a deploy should be loud.
package config

import (
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// Config is the validated server configuration.
type Config struct {
	// Addr is the TCP address the HTTP server listens on (MAHANATHI_ADDR).
	Addr string
	// LogLevel is the minimum slog level that is written (MAHANATHI_LOG_LEVEL:
	// debug, info, warn or error).
	LogLevel slog.Level
	// LogFormat is "json" for machines or "text" for a terminal
	// (MAHANATHI_LOG_FORMAT).
	LogFormat string
	// ShutdownTimeout bounds how long in-flight requests may take to finish
	// after SIGINT or SIGTERM (MAHANATHI_SHUTDOWN_TIMEOUT, a Go duration).
	ShutdownTimeout time.Duration
}

// Defaults returns the configuration used when no variables are set.
func Defaults() Config {
	return Config{
		Addr:            ":8080",
		LogLevel:        slog.LevelInfo,
		LogFormat:       "json",
		ShutdownTimeout: 10 * time.Second,
	}
}

// FromEnv builds a Config from getenv, which is os.Getenv in production and a
// map lookup in tests. Empty variables keep their defaults.
func FromEnv(getenv func(string) string) (Config, error) {
	c := Defaults()

	if v := getenv("MAHANATHI_ADDR"); v != "" {
		c.Addr = v
	}

	if v := getenv("MAHANATHI_LOG_LEVEL"); v != "" {
		if err := c.LogLevel.UnmarshalText([]byte(v)); err != nil {
			return Config{}, fmt.Errorf("MAHANATHI_LOG_LEVEL: %q is not one of debug, info, warn, error", v)
		}
	}

	if v := getenv("MAHANATHI_LOG_FORMAT"); v != "" {
		switch f := strings.ToLower(v); f {
		case "json", "text":
			c.LogFormat = f
		default:
			return Config{}, fmt.Errorf("MAHANATHI_LOG_FORMAT: %q is not json or text", v)
		}
	}

	if v := getenv("MAHANATHI_SHUTDOWN_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil || d <= 0 {
			return Config{}, fmt.Errorf("MAHANATHI_SHUTDOWN_TIMEOUT: %q is not a positive duration such as 10s", v)
		}
		c.ShutdownTimeout = d
	}

	return c, nil
}
