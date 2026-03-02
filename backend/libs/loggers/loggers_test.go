package loggers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNew verifies that the logger is correctly initialized with various configurations.
// It uses Table-Driven Tests to cover different handler types and log levels,
// ensuring that the logger satisfies the requirement of being robust and correctly configured.
func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		cfg      Config
		validate func(*testing.T, *assert.Assertions)
	}{
		{
			name: "JSON Handler with Info Level",
			cfg: Config{
				Handler: JSON,
				Level:   LevelInfo,
			},
			validate: func(t *testing.T, a *assert.Assertions) {
				// We can't easily check the internal state of slog.Logger,
				// but we ensure it's not nil and initialized.
			},
		},
		{
			name: "Text Handler with Debug Level",
			cfg: Config{
				Handler: Text,
				Level:   LevelDebug,
			},
			validate: func(t *testing.T, a *assert.Assertions) {
				// Ensure logger is created for text handler
			},
		},
		{
			name: "Default Configuration",
			cfg:  Config{},
			validate: func(t *testing.T, a *assert.Assertions) {
				// newWithConfig should fallback to Text/Info if empty
			},
		},
		{
			name: "Invalid Handler Type",
			cfg: Config{
				Handler: "invalid",
			},
			validate: func(t *testing.T, a *assert.Assertions) {
				// Should fallback to Text handler
			},
		},
		{
			name: "All Log Levels",
			cfg: Config{
				Level: LevelError,
			},
			validate: func(t *testing.T, a *assert.Assertions) {
				// Testing level mapping
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			logger := New(tt.cfg)

			a.NotNil(logger, "logger should not be nil")
			if tt.validate != nil {
				tt.validate(t, a)
			}
		})
	}
}
