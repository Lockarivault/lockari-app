package repositoryvaultdb

import (
	"context"
	"fmt"
	"time"

	"github.com/Lockarivault/lockari-app/backend/libs/database/nosql"
	"github.com/Lockarivault/lockari-app/backend/libs/uuid"
	vaultmodel "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/vault/model"
	repositoryvault "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/vault/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/fx"
)

type vaultEntity struct {
	ID          string               `bson:"_id"`
	TenantID    string               `bson:"tenant_id"`
	Name        string               `bson:"name"`
	Description string               `bson:"description,omitempty"`
	Type        vaultmodel.VaultType `bson:"type"`
	CreatedAt   time.Time            `bson:"created_at"`
	UpdatedAt   time.Time            `bson:"updated_at"`
}

type vaultMongoRepository struct {
	collection *mongo.Collection
}

func NewVaultMongoRepository(lc fx.Lifecycle, db nosql.DatabaseService) repositoryvault.RepositoryVault {
	repo := &vaultMongoRepository{}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			mongoDB, err := db.GetConnection()
			if err != nil {
				return fmt.Errorf("failed to get mongodb connection for vault: %w", err)
			}
			repo.collection = mongoDB.Collection("vaults")

			// Ensure indexes
			_, err = repo.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
				Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "name", Value: 1}},
			})
			return err
		},
	})

	return repo
}

func (r *vaultMongoRepository) Save(ctx context.Context, vault vaultmodel.Vault) error {
	if r.collection == nil {
		return fmt.Errorf("vault repository not initialized")
	}

	entity := vaultEntity{
		ID:          vault.ID.String(),
		TenantID:    vault.TenantID.String(),
		Name:        vault.Name,
		Description: vault.Description,
		Type:        vault.Type,
		CreatedAt:   vault.CreatedAt,
		UpdatedAt:   vault.UpdatedAt,
	}

	_, err := r.collection.InsertOne(ctx, entity)
	return err
}

func (r *vaultMongoRepository) GetByID(ctx context.Context, tenantID uuid.UUID, vaultID uuid.UUID) (vaultmodel.Vault, error) {
	if r.collection == nil {
		return vaultmodel.Vault{}, fmt.Errorf("vault repository not initialized")
	}

	var entity vaultEntity
	err := r.collection.FindOne(ctx, bson.M{"_id": vaultID.String(), "tenant_id": tenantID.String()}).Decode(&entity)
	if err != nil {
		return vaultmodel.Vault{}, err
	}

	id, _ := uuid.Parse(entity.ID)
	tID, _ := uuid.Parse(entity.TenantID)

	return vaultmodel.Vault{
		ID:          id,
		TenantID:    tID,
		Name:        entity.Name,
		Description: entity.Description,
		Type:        entity.Type,
		CreatedAt:   entity.CreatedAt,
		UpdatedAt:   entity.UpdatedAt,
	}, nil
}

func (r *vaultMongoRepository) List(ctx context.Context, tenantID uuid.UUID) ([]vaultmodel.Vault, error) {
	if r.collection == nil {
		return nil, fmt.Errorf("vault repository not initialized")
	}

	cursor, err := r.collection.Find(ctx, bson.M{"tenant_id": tenantID.String()})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var entities []vaultEntity
	if err := cursor.All(ctx, &entities); err != nil {
		return nil, err
	}

	vaults := make([]vaultmodel.Vault, len(entities))
	for i, entity := range entities {
		id, _ := uuid.Parse(entity.ID)
		tID, _ := uuid.Parse(entity.TenantID)
		vaults[i] = vaultmodel.Vault{
			ID:          id,
			TenantID:    tID,
			Name:        entity.Name,
			Description: entity.Description,
			Type:        entity.Type,
			CreatedAt:   entity.CreatedAt,
			UpdatedAt:   entity.UpdatedAt,
		}
	}
	return vaults, nil
}

var Module = fx.Options(
	fx.Provide(NewVaultMongoRepository),
)
