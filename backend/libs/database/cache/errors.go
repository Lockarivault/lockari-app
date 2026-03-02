// Package cache provides a standardized Redis client wrapper for the Lockari app.
// This file defines common sentinel errors used to handle cache-specific edge cases
// without leaking underlying implementation details (like redis.Nil).
package cache

import "errors"

// ErrCacheMiss is the sentinel error returned when a key does not exist.
// This allows callers to distinguish between a missing value and a transient network error.
var ErrCacheMiss = errors.New("cache: miss")

// ErrNilClient returned when the redis client is not initialized.
// This exists to prevent nil pointer panics when executing operations before Connect().
var ErrNilClient = errors.New("cache: redis client is nil")
