package repositoryauditdb

import (
	"context"
	"fmt"

	"github.com/Lockarivault/lockari-app/backend/libs/database/nosql"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(NewAuditMongoRepository),
	fx.Invoke(EnsureIndexes),
)

func EnsureIndexes(lc fx.Lifecycle, db nosql.DatabaseService) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			mongoDB, err := db.GetConnection()
			if err != nil {
				return fmt.Errorf("failed to get mongodb connection for indexing: %w", err)
			}
			collection := mongoDB.Collection("audit_logs")

			// Index by timestamp for chronological listing and TTL
			// Index by tenant_id for customer dashboard filtering
			indexes := []mongo.IndexModel{
				{
					Keys: bson.D{{Key: "timestamp", Value: -1}},
				},
				{
					Keys: bson.D{{Key: "context.tenant_id", Value: 1}},
				},
				{
					Keys: bson.D{{Key: "actor.id", Value: 1}},
				},
				// Activity 10: Retention Policy (TTL)
				// We can set an expireAfterSeconds here later
				{
					Keys:    bson.D{{Key: "timestamp", Value: 1}},
					Options: options.Index().SetExpireAfterSeconds(365 * 24 * 60 * 60), // 1 year default
				},
			}

			_, err = collection.Indexes().CreateMany(ctx, indexes)
			return err
		},
	})
}
