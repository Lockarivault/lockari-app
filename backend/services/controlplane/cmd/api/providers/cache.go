package providers

import (
	"context"

	"github.com/Lockarivault/lockari-app/backend/libs/database/cache"
	"github.com/Lockarivault/lockari-app/backend/libs/loggers"
	"github.com/Lockarivault/lockari-app/backend/services/controlplane/config"
	"go.uber.org/fx"
)

// ProvideCache provides a Redis cache client based on the application configuration.
func ProvideCache(lc fx.Lifecycle, cfg config.AppConfig, logger loggers.LoggerInterface) cache.RedisClientInterface {
	addr := "localhost"
	port := "6379"
	password := ""
	user := ""
	db := 0
	tracing := true
	useTLS := true

	if val, ok := cfg["redis"].(map[string]interface{}); ok {
		if v, ok := val["address"].(string); ok {
			addr = v
		}
		if v, ok := val["port"].(string); ok {
			port = v
		}
		if v, ok := val["password"].(string); ok {
			password = v
		}
		if v, ok := val["user"].(string); ok {
			user = v
		}
		if v, ok := val["db"].(int); ok {
			db = v
		}
		if v, ok := val["use_tls"].(bool); ok {
			useTLS = v
		}
	}

	logger.Info("Initializing Redis cache",
		"address", addr,
		"port", port,
		"db", db,
		"user", user,
		"use_tls", useTLS,
	)

	client := cache.NewRedisClient(cache.CacheConfig{
		Addr:     addr,
		Port:     port,
		Password: password,
		Username: user,
		DB:       db,
		Tracing:  tracing,
		UseTLS:   useTLS,
	})

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := client.Connect(ctx); err != nil {
				return err
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return client.Disconnect(ctx)
		},
	})

	return client
}
