package repositoryauditdb

import (
	"context"
	"fmt"

	"github.com/Lockarivault/lockari-app/backend/libs/database/nosql"
	auditmodel "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/audit/model"
	repositoryaudit "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/audit/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/fx"
)

type auditMongoRepository struct {
	collection *mongo.Collection
}

func NewAuditMongoRepository(lc fx.Lifecycle, db nosql.DatabaseService) repositoryaudit.RepositoryAudit {
	repo := &auditMongoRepository{}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			mongoDB, err := db.GetConnection()
			if err != nil {
				return fmt.Errorf("failed to get mongodb connection for audit: %w", err)
			}
			repo.collection = mongoDB.Collection("audit_logs")
			return nil
		},
	})

	return repo
}

func (r *auditMongoRepository) Save(ctx context.Context, log auditmodel.AuditLog) error {
	if r.collection == nil {
		return fmt.Errorf("audit repository not initialized")
	}
	entity := toEntity(log)
	_, err := r.collection.InsertOne(ctx, entity)
	if err != nil {
		return fmt.Errorf("failed to insert audit log: %w", err)
	}
	return nil
}

func (r *auditMongoRepository) List(ctx context.Context, filter map[string]interface{}) ([]auditmodel.AuditLog, error) {
	// Simple BSON map for now
	query := bson.M{}
	for k, v := range filter {
		query[k] = v
	}

	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}})
	cursor, err := r.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs: %w", err)
	}
	defer cursor.Close(ctx)

	var entities []AuditLogEntity
	if err := cursor.All(ctx, &entities); err != nil {
		return nil, fmt.Errorf("failed to decode audit logs: %w", err)
	}

	logs := make([]auditmodel.AuditLog, len(entities))
	for i, e := range entities {
		logs[i] = fromEntity(e)
	}
	return logs, nil
}

func (r *auditMongoRepository) GetByID(ctx context.Context, id string) (auditmodel.AuditLog, error) {
	var entity AuditLogEntity
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&entity)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return auditmodel.AuditLog{}, fmt.Errorf("audit log not found")
		}
		return auditmodel.AuditLog{}, fmt.Errorf("failed to get audit log: %w", err)
	}
	return fromEntity(entity), nil
}
