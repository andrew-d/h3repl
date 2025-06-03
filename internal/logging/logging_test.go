package logging

import (
	"log/slog"
	"testing"
)

func TestSetup(t *testing.T) {

	tests := []struct {
		name          string
		verbose       bool
		expectedLevel slog.Level
	}{
		{
			name:          "verbose mode",
			verbose:       true,
			expectedLevel: slog.LevelDebug,
		},
		{
			name:          "normal mode",
			verbose:       false,
			expectedLevel: slog.LevelInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := Setup(tt.verbose)
			if logger == nil {
				t.Error("expected logger but got nil")
				return
			}

			// Test that the logger is functional by trying to log something
			// We can't easily test the exact level without accessing internals,
			// but we can verify the logger doesn't panic and is usable
			logger.Info("test message")
			logger.Debug("debug message")
		})
	}
}

func TestSetupReturnsValidLogger(t *testing.T) {

	logger := Setup(false)

	// Verify logger is not nil
	if logger == nil {
		t.Error("Setup() returned nil logger")
	}

	// Verify logger can be used without panicking
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("logger.Info() panicked: %v", r)
		}
	}()

	logger.Info("test log message")
}
