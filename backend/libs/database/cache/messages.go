// Package cache provides a standardized Redis client wrapper for the Lockari app.
// It includes instrumentation for telemetry (metrics and tracing) and simplifies common operations.
package cache

// ClientMessage represents a standardized status message returned by cache operations.
// These messages are used for logging, telemetry tagging, and internal signaling.
type ClientMessage string

const (
	// ClientMessageCacheMiss indicates that the requested key was not found in the cache.
	ClientMessageCacheMiss ClientMessage = "cache_miss"
	// ClientMessageCacheHit indicates that the requested key was successfully retrieved from the cache.
	ClientMessageCacheHit ClientMessage = "cache_hit"
	// ClientMessageKeyNotFound is an alternative status for missing keys used in specific workflows.
	ClientMessageKeyNotFound ClientMessage = "key_not_found"
	// ClientMessageFailedToGetValue indicates an unexpected error during the retrieval process.
	ClientMessageFailedToGetValue ClientMessage = "failed_to_get_value"
)

// String returns the string representation of the cache message.
func (c ClientMessage) String() string {
	return string(c)
}
