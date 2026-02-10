package uuid

import (
	"github.com/google/uuid"
)

// UUID is a wrapper for uuid.UUID
type UUID = uuid.UUID

// Nil is the zero value for UUID
var Nil = uuid.Nil

// New generates a new UUID v7
func New() UUID {
	id, _ := uuid.NewV7()
	return id
}

// NewString generates a new UUID v7 as a string
func NewString() string {
	return New().String()
}

// Parse decodes s into a UUID
func Parse(s string) (UUID, error) {
	return uuid.Parse(s)
}

// IsValid checks if s is a valid UUID
func IsValid(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}
