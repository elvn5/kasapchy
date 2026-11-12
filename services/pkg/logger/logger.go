// Package logger builds the structured slog logger used by every service.
package logger

import (
	"log/slog"
	"os"
)

// New returns a logger tagged with the service name. Production uses JSON output
// (for log collectors); other environments use human-readable text.
func New(service, appEnv, level string) *slog.Logger {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		lvl = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: lvl}
	var handler slog.Handler
	if appEnv == "production" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler).With(slog.String("service", service))
	slog.SetDefault(logger)
	return logger
}
