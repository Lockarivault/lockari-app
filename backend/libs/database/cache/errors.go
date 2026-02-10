package cache

import "errors"

// ErrCacheMiss is the sentinel error returned when a key does not exist.
var ErrCacheMiss = errors.New("cache: miss")

// ErrNilClient returned when the redis client is not initialized.
var ErrNilClient = errors.New("cache: redis client is nil")
