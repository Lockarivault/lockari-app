// Package uuid provides a standardized wrapper for UUID operations across the project.
// This package exists to ensure consistent use of UUID v7 throughout the backend,
// which provides time-ordered identifiers that improve database indexing performance
// compared to random UUIDs.
package uuid

import (
	"github.com/google/uuid"
)

// UUID represents a 128-bit Universal Unique Identifier.
// This type is aliased to provide a clean interface for other packages
// without exposing the underlying implementation details.
type UUID = uuid.UUID

// Nil represents an empty UUID (all zeros).
// Use this for comparison or as a default value when no UUID is available.
var Nil = uuid.Nil

// New generates a new UUID v7.
// It uses version 7 to provide time-ordered identifiers, which is crucial
// for maintaining database performance with large datasets.
func New() UUID {
	id, _ := uuid.NewV7()
	return id
}

// NewString generates a new UUID v7 as a string.
// This is a convenience function for cases where the string representation
// is needed directly, such as in API responses or logs.
func NewString() string {
	return New().String()
}

// Parse decodes s into a UUID.
// It accepts various UUID formats and returns an error if the input is invalid.
func Parse(s string) (UUID, error) {
	return uuid.Parse(s)
}

// IsValid checks if the provided string is a valid UUID.
// This is a non-allocating check useful for validation layers before
// attempting to parse or process the ID.
func IsValid(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}
