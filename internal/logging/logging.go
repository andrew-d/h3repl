// Package logging provides structured logging setup for h3repl.
package logging

import (
	"log/slog"
	"os"
)

// Setup configures and returns a structured logger
func Setup(verbose bool) *slog.Logger {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewTextHandler(os.Stdout, opts)
	return slog.New(handler)
}
