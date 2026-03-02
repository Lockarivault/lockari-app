package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lockarivault/lockari-app/backend/libs/mensageria"
	"github.com/rabbitmq/amqp091-go"
)

type mensageriaAdapter struct {
	mq mensageria.MessageQueue
}

// NewMensageriaAdapter creates a new Publisher that uses the mensageria library.
func NewMensageriaAdapter(mq mensageria.MessageQueue) Publisher {
	return &mensageriaAdapter{
		mq: mq,
	}
}

// Publish serializes the event payload and delegates to the underlying message queue.
func (a *mensageriaAdapter) Publish(ctx context.Context, exchange, routingKey string, event Event) error {
	bodyBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal event payload: %w", err)
	}

	trace_id := ""
	span_id := ""
	headers := make(amqp091.Table)
	for k, v := range event.Headers {
		if k == "trace_id" {
			trace_id = v.(string)
		}
		if k == "span_id" {
			span_id = v.(string)
		}
		headers[k] = v
	}

	msg := amqp091.Publishing{
		ContentType:   "application/json",
		DeliveryMode:  amqp091.Persistent,
		Timestamp:     time.Now(),
		Type:          event.Type,
		Headers:       headers,
		Body:          bodyBytes,
		MessageId:     trace_id,
		CorrelationId: span_id,
	}

	if err := a.mq.Publish(ctx, exchange, routingKey, msg); err != nil {
		return fmt.Errorf("failed to publish message via mensageria: %w", err)
	}

	return nil
}
