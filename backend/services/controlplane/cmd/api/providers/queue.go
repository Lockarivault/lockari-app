package providers

import (
	"context"

	"github.com/lockarivault/lockari-app/backend/libs/mensageria"
	"github.com/lockarivault/lockari-app/backend/libs/telemetry"
	"github.com/lockarivault/lockari-app/backend/services/controlplane/config"
	"go.uber.org/fx"
)

func ProvideQueue(lc fx.Lifecycle, cfg *config.Connections, obs telemetry.OtelObservability) (mensageria.MessageQueue, error) {
	// Convert service config to library config
	libCfg := mensageria.Config{
		Addr:     cfg.MessageQueues.Addr,
		User:     cfg.MessageQueues.User,
		Password: cfg.MessageQueues.Password,
		VHost:    cfg.MessageQueues.VHost,
		Port:     cfg.MessageQueues.Port,
		TLS:      cfg.MessageQueues.TLS,
	}

	fullCfg := mensageria.FullConfig{
		RabbitMQ: libCfg,
	}

	// Map Exchanges
	for _, ex := range cfg.MessageQueues.Data.Exchanges {
		fullCfg.Exchanges = append(fullCfg.Exchanges, mensageria.ExchangeConfig{
			Name:        ex.Name,
			Kind:        ex.Kind,
			Durable:     ex.Durable,
			AutoDeleted: ex.AutoDeleted,
			Internal:    ex.Internal,
			NoWait:      ex.NoWait,
			Args:        ex.Args,
		})
	}

	// Map Queues
	for _, q := range cfg.MessageQueues.Data.Queues {
		var dl *mensageria.DeadLetterConfig
		if q.DeadLetter != nil {
			dl = &mensageria.DeadLetterConfig{
				Exchange: q.DeadLetter.Exchange,
				Key:      q.DeadLetter.Key,
			}
		}
		fullCfg.Queues = append(fullCfg.Queues, mensageria.QueueConfig{
			Name:               q.Name,
			Durable:            q.Durable,
			AutoDelete:         q.AutoDelete,
			Exclusive:          q.Exclusive,
			NoWait:             q.NoWait,
			Args:               q.Args,
			DeadLetter:         dl,
			PrefetchCount:      q.PrefetchCount,
			MaxDeliveryRetries: q.MaxDeliveryRetries,
		})
	}

	// Map Bindings
	for _, b := range cfg.MessageQueues.Data.Bindings {
		fullCfg.Bindings = append(fullCfg.Bindings, mensageria.BindingConfig{
			QueueName:    b.Queue,
			ExchangeName: b.Exchange,
			RoutingKey:   b.RoutingKey,
			NoWait:       b.NoWait,
			Args:         b.Args,
		})
	}

	client := mensageria.NewClient(libCfg, fullCfg, obs.Tracer("rabbitmq-client"), obs.TextMapPropagator())

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return client.Connect(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return client.Disconnect()
		},
	})

	return client, nil
}
