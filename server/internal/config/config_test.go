package config

import (
	"log/slog"
	"testing"
	"time"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestFromEnvDefaults(t *testing.T) {
	got, err := FromEnv(env(nil))
	if err != nil {
		t.Fatal(err)
	}
	if got != Defaults() {
		t.Fatalf("FromEnv(empty) = %+v, want defaults %+v", got, Defaults())
	}
}

func TestFromEnvOverrides(t *testing.T) {
	got, err := FromEnv(env(map[string]string{
		"MAHANATHI_ADDR":             "127.0.0.1:9000",
		"MAHANATHI_LOG_LEVEL":        "debug",
		"MAHANATHI_LOG_FORMAT":       "TEXT",
		"MAHANATHI_SHUTDOWN_TIMEOUT": "3s",
	}))
	if err != nil {
		t.Fatal(err)
	}
	want := Config{Addr: "127.0.0.1:9000", LogLevel: slog.LevelDebug, LogFormat: "text", ShutdownTimeout: 3 * time.Second}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestFromEnvRejectsBadValues(t *testing.T) {
	for _, tc := range []struct{ key, val string }{
		{"MAHANATHI_LOG_LEVEL", "loud"},
		{"MAHANATHI_LOG_FORMAT", "xml"},
		{"MAHANATHI_SHUTDOWN_TIMEOUT", "ten"},
		{"MAHANATHI_SHUTDOWN_TIMEOUT", "-1s"},
		{"MAHANATHI_SHUTDOWN_TIMEOUT", "0s"},
	} {
		if _, err := FromEnv(env(map[string]string{tc.key: tc.val})); err == nil {
			t.Errorf("%s=%q: want error, got nil", tc.key, tc.val)
		}
	}
}
