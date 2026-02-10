package tenantdatabase

import (
	"context"
	"fmt"
	"time"

	"github.com/Lockarivault/lockari-app/backend/libs/database/nosql"
	"github.com/Lockarivault/lockari-app/backend/libs/encryption"
	"github.com/Lockarivault/lockari-app/backend/libs/telemetry"
	"github.com/Lockarivault/lockari-app/backend/libs/uuid"
	tenantmodel "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/model"
	repositorytenant "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/repository"
	tenanttools "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/tools"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/fx"
)

type tenantMongoRepository struct {
	collection *mongo.Collection
	obs        telemetry.OtelObservability
}

func NewTenantMongoRepository(lc fx.Lifecycle, db nosql.DatabaseService, obs telemetry.OtelObservability) (repositorytenant.RepositoryTenant, error) {
	if db == nil {
		return nil, tenanttools.ErrNilRepository
	}
	if obs == nil {
		return nil, tenanttools.ErrNilTelemetry
	}

	repo := &tenantMongoRepository{
		obs: obs,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			mongoDB, err := db.GetConnection()
			if err != nil {
				return fmt.Errorf("failed to get mongodb connection: %w", err)
			}
			repo.collection = mongoDB.Collection("tenants")

			// Ensure unique index for slug
			_, err = repo.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
				Keys:    map[string]interface{}{"slug": 1},
				Options: options.Index().SetUnique(true),
			})
			if err != nil {
				return fmt.Errorf("failed to create unique index for slug: %w", err)
			}

			return nil
		},
	})

	return repo, nil
}

func (r *tenantMongoRepository) Create(ctx context.Context, tenant tenantmodel.TenantModel) error {
	if r.collection == nil {
		return fmt.Errorf("repository not initialized - collection is nil")
	}
	entity := fromDomain(tenant)
	_, err := r.collection.InsertOne(ctx, entity)
	return err
}

func (r *tenantMongoRepository) GetByID(ctx context.Context, id uuid.UUID) (tenantmodel.TenantModel, error) {
	if r.collection == nil {
		return tenantmodel.TenantModel{}, fmt.Errorf("repository not initialized - collection is nil")
	}
	var entity tenantEntity
	err := r.collection.FindOne(ctx, map[string]interface{}{"_id": id.String()}).Decode(&entity)
	if err != nil {
		return tenantmodel.TenantModel{}, err
	}
	return toDomain(entity), nil
}

func (r *tenantMongoRepository) GetBySlug(ctx context.Context, slug string) (tenantmodel.TenantModel, error) {
	if r.collection == nil {
		return tenantmodel.TenantModel{}, fmt.Errorf("repository not initialized - collection is nil")
	}
	var entity tenantEntity
	err := r.collection.FindOne(ctx, map[string]interface{}{"slug": slug}).Decode(&entity)
	if err != nil {
		return tenantmodel.TenantModel{}, err
	}
	return toDomain(entity), nil
}

func (r *tenantMongoRepository) Update(ctx context.Context, tenant tenantmodel.TenantModel) error {
	if r.collection == nil {
		return fmt.Errorf("repository not initialized - collection is nil")
	}
	entity := fromDomain(tenant)
	_, err := r.collection.UpdateOne(ctx, map[string]interface{}{"_id": entity.ID}, map[string]interface{}{"$set": entity})
	return err
}

func (r *tenantMongoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if r.collection == nil {
		return fmt.Errorf("repository not initialized - collection is nil")
	}
	_, err := r.collection.DeleteOne(ctx, map[string]interface{}{"_id": id.String()})
	return err
}

func (r *tenantMongoRepository) List(ctx context.Context, filter map[string]interface{}) ([]tenantmodel.TenantModel, error) {
	if r.collection == nil {
		return nil, fmt.Errorf("repository not initialized - collection is nil")
	}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var entities []tenantEntity
	if err := cursor.All(ctx, &entities); err != nil {
		return nil, err
	}

	results := make([]tenantmodel.TenantModel, len(entities))
	for i, e := range entities {
		results[i] = toDomain(e)
	}
	return results, nil
}

func (r *tenantMongoRepository) UpdateSecurityMetadata(ctx context.Context, id uuid.UUID, metadata encryption.EncryptMetadata) error {
	if r.collection == nil {
		return fmt.Errorf("repository not initialized - collection is nil")
	}
	entity := fromDomainMetadata(metadata)
	_, err := r.collection.UpdateOne(ctx, map[string]interface{}{"_id": id.String()}, map[string]interface{}{
		"$set": map[string]interface{}{
			"security_metadata": entity,
			"updated_at":        time.Now().UTC(),
		},
	})
	return err
}

func (r *tenantMongoRepository) ActivateTenant(ctx context.Context, id uuid.UUID) error {
	if r.collection == nil {
		return fmt.Errorf("repository not initialized - collection is nil")
	}
	_, err := r.collection.UpdateOne(ctx, map[string]interface{}{"_id": id.String()}, map[string]interface{}{
		"$set": map[string]interface{}{
			"status":     tenantmodel.StatusActive,
			"updated_at": time.Now().UTC(),
		},
	})
	return err
}

func (r *tenantMongoRepository) DeactivateTenant(ctx context.Context, id uuid.UUID) error {
	if r.collection == nil {
		return fmt.Errorf("repository not initialized - collection is nil")
	}
	_, err := r.collection.UpdateOne(ctx, map[string]interface{}{"_id": id.String()}, map[string]interface{}{
		"$set": map[string]interface{}{
			"status":     tenantmodel.StatusInactive,
			"updated_at": time.Now().UTC(),
		},
	})
	return err
}

func (r *tenantMongoRepository) FailTenant(ctx context.Context, id uuid.UUID, reason string) error {
	if r.collection == nil {
		return fmt.Errorf("repository not initialized - collection is nil")
	}
	_, err := r.collection.UpdateOne(ctx, map[string]interface{}{"_id": id.String()}, map[string]interface{}{
		"$set": map[string]interface{}{
			"status":         tenantmodel.StatusFailed,
			"failure_reason": reason,
			"updated_at":     time.Now().UTC(),
		},
	})
	return err
}

func (r *tenantMongoRepository) UpdateProprietiesTypes(ctx context.Context, id uuid.UUID, properties tenantmodel.ProprietiesTypes) error {
	if r.collection == nil {
		return fmt.Errorf("repository not initialized - collection is nil")
	}
	_, err := r.collection.UpdateOne(ctx, map[string]interface{}{"_id": id.String()}, map[string]interface{}{
		"$set": map[string]interface{}{
			"properties": properties,
			"updated_at": time.Now().UTC(),
		},
	})
	return err
}
