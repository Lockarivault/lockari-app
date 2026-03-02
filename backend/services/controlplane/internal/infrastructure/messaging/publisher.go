package messaging

import (
	"context"
)

// Event defines a generic representation of a message to be published.
type Event struct {
	Type    string
	Payload any
	Headers map[string]any
}

// Publisher defines the interface for publishing events to a message broker.
type Publisher interface {
	Publish(ctx context.Context, exchange, routingKey string, event Event) error
}
