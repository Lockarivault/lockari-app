package uuid

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	id1 := NewString()
	id2 := NewString()

	if id1 == id2 {
		t.Errorf("expected different UUIDs, got both %s", id1)
	}

	if !IsValid(id1) {
		t.Errorf("expected %s to be a valid UUID", id1)
	}
}

func TestNewV7Ordering(t *testing.T) {
	// UUID v7 is time-ordered
	id1 := NewString()
	time.Sleep(1 * time.Millisecond) // Ensure time moves forward
	id2 := NewString()

	if id1 >= id2 {
		t.Errorf("expected id1 < id2 (time ordered), but %s >= %s", id1, id2)
	}
}

func TestParse(t *testing.T) {
	original := NewString()
	parsed, err := Parse(original)
	if err != nil {
		t.Fatalf("failed to parse valid UUID: %v", err)
	}

	if parsed.String() != original {
		t.Errorf("expected %s, got %s", original, parsed.String())
	}

	_, err = Parse("invalid-uuid")
	if err == nil {
		t.Error("expected error parsing invalid UUID, got nil")
	}
}
