package mensageria

import (
	"context"
	"fmt"
	"sync"

	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type Client struct {
	config     Config
	fullConfig FullConfig
	conn       *amqp091.Connection
	channel    *amqp091.Channel
	mu         sync.RWMutex

	tracer     trace.Tracer
	propagator propagation.TextMapPropagator

	consumers   map[string]*ConsumerInfo
	consumersWg sync.WaitGroup
	shutdown    chan struct{}
}

func NewClient(cfg Config, fullCfg FullConfig, tracer trace.Tracer, propagator propagation.TextMapPropagator) *Client {
	return &Client{
		config:     cfg,
		fullConfig: fullCfg,
		tracer:     tracer,
		propagator: propagator,
		consumers:  make(map[string]*ConsumerInfo),
		shutdown:   make(chan struct{}),
	}
}

func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Prefer TLS if enabled or if port is 5671
	scheme := "amqp"
	if c.config.TLS || c.config.Port == "5671" {
		scheme = "amqps"
	}

	vhost := c.config.VHost
	if vhost != "" && vhost[0] != '/' {
		vhost = "/" + vhost
	}

	url := fmt.Sprintf("%s://%s:%s@%s:%s%s",
		scheme, c.config.User, c.config.Password, c.config.Addr, c.config.Port, vhost)

	conn, err := amqp091.Dial(url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	c.conn = conn

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}
	c.channel = ch

	// Declarative Setup
	if err := c.setupInfrastructure(); err != nil {
		return fmt.Errorf("failed to setup infrastructure: %w", err)
	}

	return nil
}

func (c *Client) setupInfrastructure() error {
	// 1. Declare Exchanges
	for _, ex := range c.fullConfig.Exchanges {
		err := c.channel.ExchangeDeclare(
			ex.Name, ex.Kind, ex.Durable, ex.AutoDeleted,
			ex.Internal, ex.NoWait, ex.Args,
		)
		if err != nil {
			return fmt.Errorf("failed to declare exchange %s: %w", ex.Name, err)
		}
	}

	// 2. Declare Queues
	for _, q := range c.fullConfig.Queues {
		args := amqp091.Table(q.Args)
		if q.DeadLetter != nil {
			if args == nil {
				args = make(amqp091.Table)
			}
			args["x-dead-letter-exchange"] = q.DeadLetter.Exchange
			args["x-dead-letter-routing-key"] = q.DeadLetter.Key
		}

		_, err := c.channel.QueueDeclare(
			q.Name, q.Durable, q.AutoDelete, q.Exclusive, q.NoWait, args,
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", q.Name, err)
		}
	}

	// 3. Declare Bindings
	for _, b := range c.fullConfig.Bindings {
		err := c.channel.QueueBind(
			b.QueueName, b.RoutingKey, b.ExchangeName, b.NoWait, amqp091.Table(b.Args),
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue %s to exchange %s: %w", b.QueueName, b.ExchangeName, err)
		}
	}

	return nil
}

func (c *Client) Ping(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.conn == nil || c.conn.IsClosed() {
		return fmt.Errorf("rabbit connection is closed")
	}

	if c.channel == nil || c.channel.IsClosed() {
		return fmt.Errorf("rabbit channel is closed")
	}

	return nil
}

func (c *Client) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	close(c.shutdown)
	c.consumersWg.Wait()

	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) Publish(ctx context.Context, exchangeName string, routingKey string, msg amqp091.Publishing) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.channel == nil {
		return fmt.Errorf("channel is nil")
	}

	// Inject tracing context
	if c.propagator != nil {
		if msg.Headers == nil {
			msg.Headers = make(amqp091.Table)
		}
		c.propagator.Inject(ctx, amqp091TableCarrier(msg.Headers))
	}

	return c.channel.PublishWithContext(ctx, exchangeName, routingKey, false, false, msg)
}

func (c *Client) Consume(ctx context.Context, queueName string, autoAck bool, consumerTag string, handler func(amqp091.Delivery) error) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	deliveries, err := c.channel.Consume(
		queueName, consumerTag, autoAck, false, false, false, nil,
	)
	if err != nil {
		return err
	}

	childCtx, cancel := context.WithCancel(ctx)
	c.consumers[consumerTag] = &ConsumerInfo{
		Channel: c.channel,
		Cancel:  cancel,
	}

	c.consumersWg.Add(1)
	go func() {
		defer c.consumersWg.Done()
		for {
			select {
			case <-c.shutdown:
				return
			case <-childCtx.Done():
				return
			case d, ok := <-deliveries:
				if !ok {
					return
				}

				// Extract tracing context
				var span trace.Span
				if c.propagator != nil {
					remoteCtx := c.propagator.Extract(childCtx, amqp091TableCarrier(d.Headers))
					_, span = c.tracer.Start(remoteCtx, fmt.Sprintf("Consume %s", queueName))
				} else {
					_, span = c.tracer.Start(childCtx, fmt.Sprintf("Consume %s", queueName))
				}

				err := handler(d)
				if err != nil {
					// Handle error, maybe nack
				}
				span.End()
			}
		}
	}()

	return nil
}

// amqp091TableCarrier adapt amqp091.Table for otel propagation
type amqp091TableCarrier amqp091.Table

func (c amqp091TableCarrier) Get(key string) string {
	if v, ok := c[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (c amqp091TableCarrier) Set(key string, value string) {
	c[key] = value
}

func (c amqp091TableCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return keys
}
