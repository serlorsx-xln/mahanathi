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
	want := with(Defaults(), func(c *Config) {
		c.Addr, c.LogLevel, c.LogFormat, c.ShutdownTimeout = "127.0.0.1:9000", slog.LevelDebug, "text", 3*time.Second
	})
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

func TestPortAddrAndDataDirFromEnv(t *testing.T) {
	cases := []struct {
		name   string
		env    map[string]string
		want   Config
		expErr bool
	}{
		{name: "defaults", env: map[string]string{}, want: Defaults()},
		{name: "port only", env: map[string]string{"PORT": "8081"},
			want: with(Defaults(), func(c *Config) { c.Addr = ":8081" })},
		{name: "addr beats port", env: map[string]string{"PORT": "8081", "MAHANATHI_ADDR": "127.0.0.1:9000"},
			want: with(Defaults(), func(c *Config) { c.Addr = "127.0.0.1:9000" })},
		{name: "bad port", env: map[string]string{"PORT": "http"}, expErr: true},
		{name: "port out of range", env: map[string]string{"PORT": "70000"}, expErr: true},
		{name: "data dir", env: map[string]string{"MAHANATHI_DATA_DIR": "/data"},
			want: with(Defaults(), func(c *Config) { c.DataDir = "/data" })},
		{name: "poll schedule", env: map[string]string{"MAHANATHI_POLL_EVERY": "30m", "MAHANATHI_POLL_OFFSET": "3m"},
			want: with(Defaults(), func(c *Config) { c.PollEvery, c.PollOffset = 30*time.Minute, 3*time.Minute })},
		{name: "poll too fast", env: map[string]string{"MAHANATHI_POLL_EVERY": "30s"}, expErr: true},
		{name: "offset at every", env: map[string]string{"MAHANATHI_POLL_EVERY": "1h", "MAHANATHI_POLL_OFFSET": "1h"}, expErr: true},
		{name: "negative offset", env: map[string]string{"MAHANATHI_POLL_OFFSET": "-1m"}, expErr: true},
		{name: "bad duration", env: map[string]string{"MAHANATHI_POLL_EVERY": "hourly"}, expErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := FromEnv(func(k string) string { return tc.env[k] })
			if tc.expErr {
				if err == nil {
					t.Fatalf("accepted %v", tc.env)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("got %+v\nwant %+v", got, tc.want)
			}
		})
	}
}

func with(c Config, f func(*Config)) Config {
	f(&c)
	return c
}
