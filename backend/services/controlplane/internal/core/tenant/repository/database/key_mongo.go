package tenantdatabase

import (
	"context"
	"fmt"
	"time"

	"github.com/Lockarivault/lockari-app/backend/libs/database/nosql"
	"github.com/Lockarivault/lockari-app/backend/libs/uuid"
	repositorytenant "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/repository"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/fx"
)

type keyMongoRepository struct {
	collection *mongo.Collection
}

func NewKeyMongoRepository(lc fx.Lifecycle, db nosql.DatabaseService) (repositorytenant.RepositoryKey, error) {
	if db == nil {
		return nil, fmt.Errorf("database service is nil")
	}

	repo := &keyMongoRepository{}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			mongoDB, err := db.GetConnection()
			if err != nil {
				return fmt.Errorf("failed to get mongodb connection: %w", err)
			}
			repo.collection = mongoDB.Collection("keys")
			return nil
		},
	})

	return repo, nil
}

type keyEntity struct {
	ID         string    `bson:"_id"`
	TenantID   string    `bson:"tenant_id"`
	Ciphertext []byte    `bson:"ciphertext"`
	Nonce      []byte    `bson:"nonce"`
	Algorithm  string    `bson:"algorithm"`
	CreatedAt  time.Time `bson:"created_at"`
}

func (r *keyMongoRepository) Save(ctx context.Context, keyID string, tenantID uuid.UUID, encryptedKey []byte, nonce []byte, algorithm string) error {
	if r.collection == nil {
		return fmt.Errorf("repository not initialized - collection is nil")
	}

	entity := keyEntity{
		ID:         keyID,
		TenantID:   tenantID.String(),
		Ciphertext: encryptedKey,
		Nonce:      nonce,
		Algorithm:  algorithm,
		CreatedAt:  time.Now().UTC(),
	}

	_, err := r.collection.InsertOne(ctx, entity)
	return err
}

func (r *keyMongoRepository) Get(ctx context.Context, keyID string) ([]byte, []byte, string, error) {
	if r.collection == nil {
		return nil, nil, "", fmt.Errorf("repository not initialized - collection is nil")
	}

	var entity keyEntity
	err := r.collection.FindOne(ctx, map[string]interface{}{"_id": keyID}).Decode(&entity)
	if err != nil {
		return nil, nil, "", err
	}

	return entity.Ciphertext, entity.Nonce, entity.Algorithm, nil
}

func (r *keyMongoRepository) Delete(ctx context.Context, keyID string) error {
	if r.collection == nil {
		return fmt.Errorf("repository not initialized - collection is nil")
	}

	_, err := r.collection.DeleteOne(ctx, map[string]interface{}{"_id": keyID})
	return err
}
