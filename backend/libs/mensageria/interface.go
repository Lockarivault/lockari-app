package mensageria

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
)

type MessageQueue interface {
	Connect(ctx context.Context) error
	Disconnect() error
	Ping(ctx context.Context) error
	Publish(ctx context.Context, exchangeName string, routingKey string, msg amqp091.Publishing) error
	Consume(ctx context.Context, queueName string, autoAck bool, consumerTag string, handler func(amqp091.Delivery) error) error
}

type ConsumerInfo struct {
	Channel *amqp091.Channel
	Cancel  context.CancelFunc
}
