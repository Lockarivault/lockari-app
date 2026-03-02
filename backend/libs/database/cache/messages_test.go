package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestClientMessageString validates that the ClientMessage string representation
// matches its underlying value, which is important for logging and telemetry.
func TestClientMessageString(t *testing.T) {
	tests := []struct {
		name     string
		message  ClientMessage
		expected string
	}{
		{
			name:     "Cache Miss",
			message:  ClientMessageCacheMiss,
			expected: "cache_miss",
		},
		{
			name:     "Cache Hit",
			message:  ClientMessageCacheHit,
			expected: "cache_hit",
		},
		{
			name:     "Key Not Found",
			message:  ClientMessageKeyNotFound,
			expected: "key_not_found",
		},
		{
			name:     "Failed To Get Value",
			message:  ClientMessageFailedToGetValue,
			expected: "failed_to_get_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.message.String())
		})
	}
}
