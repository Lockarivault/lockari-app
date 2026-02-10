package loggers

import (
	"log/slog"
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("JSON Handler", func(t *testing.T) {
		logger := New(Config{Handler: JSON, Level: slog.LevelInfo})
		if logger == nil {
			t.Fatal("expected logger to be initialized")
		}
	})

	t.Run("Text Handler", func(t *testing.T) {
		logger := New(Config{Handler: Text, Level: slog.LevelDebug})
		if logger == nil {
			t.Fatal("expected logger to be initialized")
		}
	})

	t.Run("Default Handler", func(t *testing.T) {
		logger := New(Config{}) // Should use Text/Info
		if logger == nil {
			t.Fatal("expected logger to be initialized with defaults")
		}
	})
}
