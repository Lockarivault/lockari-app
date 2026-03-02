package providers

import (
	"context"
	"fmt"

	"github.com/lockarivault/lockari-app/backend/libs/database/nosql"
	"github.com/lockarivault/lockari-app/backend/services/controlplane/config"
	"go.uber.org/fx"
)

// ProvideNoSQL initializes and provides the NoSQL database service with Fx lifecycle management.
func ProvideNoSQL(lc fx.Lifecycle, cfg config.AppConfig) (nosql.DatabaseService, error) {
	host := "localhost"
	user := ""
	password := ""
	database := "lockari"
	port := "27017"
	useSRV := false

	if val, ok := cfg["mongodb"].(map[string]interface{}); ok {
		if v, ok := val["host"].(string); ok {
			host = v
		}
		if v, ok := val["user"].(string); ok {
			user = v
		}
		if v, ok := val["password"].(string); ok {
			password = v
		}
		if v, ok := val["database"].(string); ok {
			database = v
		}
		if v, ok := val["port"].(string); ok {
			port = v
		}
		if v, ok := val["use_srv"].(bool); ok {
			useSRV = v
		}
	}

	mongoCfg := nosql.MongoDBConnect{
		Host:     host,
		User:     user,
		Password: password,
		Database: database,
		Port:     port,
		UseSRV:   useSRV,
	}

	if err := mongoCfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid mongodb config: %w", err)
	}

	db, err := nosql.NewMongoDB(mongoCfg.GetURI(), mongoCfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to create mongodb instance: %w", err)
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return db.Connect(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return db.Close(ctx)
		},
	})

	return db, nil
}
