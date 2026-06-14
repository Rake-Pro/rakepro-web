// Package config loads runtime configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all tunable settings for the web server. Every field maps to an
// environment variable so the same binary runs identically in Docker and k8s.
type Config struct {
	// Addr is the host:port the HTTP server binds to.
	Addr string
	// Env is the deployment environment label (development, staging, production).
	// It controls log formatting: development uses a human-readable console writer.
	Env string
	// LogLevel is the minimum zerolog level to emit (trace..fatal).
	LogLevel string
	// ReadTimeout caps the time to read a full request, headers and body.
	ReadTimeout time.Duration
	// WriteTimeout caps the time to write the response.
	WriteTimeout time.Duration
	// IdleTimeout caps how long an idle keep-alive connection is held open.
	IdleTimeout time.Duration
	// ShutdownTimeout bounds graceful shutdown before connections are forced closed.
	ShutdownTimeout time.Duration
}

// Load reads configuration from the environment, applying sane defaults. It
// returns an error only when a provided value cannot be parsed.
func Load() (Config, error) {
	c := Config{
		Addr:            getEnv("RAKEPRO_ADDR", ":8080"),
		Env:             getEnv("RAKEPRO_ENV", "development"),
		LogLevel:        getEnv("RAKEPRO_LOG_LEVEL", "info"),
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     120 * time.Second,
		ShutdownTimeout: 15 * time.Second,
	}

	var err error
	if c.ReadTimeout, err = getEnvDuration("RAKEPRO_READ_TIMEOUT", c.ReadTimeout); err != nil {
		return c, err
	}
	if c.WriteTimeout, err = getEnvDuration("RAKEPRO_WRITE_TIMEOUT", c.WriteTimeout); err != nil {
		return c, err
	}
	if c.IdleTimeout, err = getEnvDuration("RAKEPRO_IDLE_TIMEOUT", c.IdleTimeout); err != nil {
		return c, err
	}
	if c.ShutdownTimeout, err = getEnvDuration("RAKEPRO_SHUTDOWN_TIMEOUT", c.ShutdownTimeout); err != nil {
		return c, err
	}

	return c, nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) (time.Duration, error) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback, nil
	}
	// Accept either a Go duration ("30s") or a bare integer count of seconds.
	if d, err := time.ParseDuration(v); err == nil {
		return d, nil
	}
	secs, err := strconv.Atoi(v)
	if err != nil {
		return fallback, fmt.Errorf("config: %s=%q is not a valid duration", key, v)
	}
	return time.Duration(secs) * time.Second, nil
}
