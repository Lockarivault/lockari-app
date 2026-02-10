package repositoryidentitydb

import (
	"context"
	"fmt"
	"time"

	"github.com/Lockarivault/lockari-app/backend/libs/database/nosql"
	"github.com/Lockarivault/lockari-app/backend/libs/uuid"
	identitymodel "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/identity/model"
	repositoryidentity "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/identity/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/fx"
)

type groupEntity struct {
	ID          string    `bson:"_id"`
	TenantID    string    `bson:"tenant_id"`
	Name        string    `bson:"name"`
	Description string    `bson:"description,omitempty"`
	CreatedAt   time.Time `bson:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at"`
}

type identityMongoRepository struct {
	collection *mongo.Collection
}

func NewIdentityMongoRepository(lc fx.Lifecycle, db nosql.DatabaseService) repositoryidentity.RepositoryIdentity {
	repo := &identityMongoRepository{}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			mongoDB, err := db.GetConnection()
			if err != nil {
				return fmt.Errorf("failed to get mongodb connection for identity: %w", err)
			}
			repo.collection = mongoDB.Collection("groups")

			// Ensure indexes
			_, err = repo.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
				Keys: bson.D{{Key: "tenant_id", Value: 1}, {Key: "name", Value: 1}},
			})
			return err
		},
	})

	return repo
}

func (r *identityMongoRepository) SaveGroup(ctx context.Context, group identitymodel.Group) error {
	if r.collection == nil {
		return fmt.Errorf("identity repository not initialized")
	}

	entity := groupEntity{
		ID:          group.ID.String(),
		TenantID:    group.TenantID.String(),
		Name:        group.Name,
		Description: group.Description,
		CreatedAt:   group.CreatedAt,
		UpdatedAt:   group.UpdatedAt,
	}

	_, err := r.collection.InsertOne(ctx, entity)
	return err
}

func (r *identityMongoRepository) GetGroupByID(ctx context.Context, tenantID uuid.UUID, groupID uuid.UUID) (identitymodel.Group, error) {
	if r.collection == nil {
		return identitymodel.Group{}, fmt.Errorf("identity repository not initialized")
	}

	var entity groupEntity
	err := r.collection.FindOne(ctx, bson.M{"_id": groupID.String(), "tenant_id": tenantID.String()}).Decode(&entity)
	if err != nil {
		return identitymodel.Group{}, err
	}

	id, _ := uuid.Parse(entity.ID)
	tID, _ := uuid.Parse(entity.TenantID)

	return identitymodel.Group{
		ID:          id,
		TenantID:    tID,
		Name:        entity.Name,
		Description: entity.Description,
		CreatedAt:   entity.CreatedAt,
		UpdatedAt:   entity.UpdatedAt,
	}, nil
}

func (r *identityMongoRepository) ListGroups(ctx context.Context, tenantID uuid.UUID) ([]identitymodel.Group, error) {
	if r.collection == nil {
		return nil, fmt.Errorf("identity repository not initialized")
	}

	cursor, err := r.collection.Find(ctx, bson.M{"tenant_id": tenantID.String()})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var entities []groupEntity
	if err := cursor.All(ctx, &entities); err != nil {
		return nil, err
	}

	groups := make([]identitymodel.Group, len(entities))
	for i, entity := range entities {
		id, _ := uuid.Parse(entity.ID)
		tID, _ := uuid.Parse(entity.TenantID)
		groups[i] = identitymodel.Group{
			ID:          id,
			TenantID:    tID,
			Name:        entity.Name,
			Description: entity.Description,
			CreatedAt:   entity.CreatedAt,
			UpdatedAt:   entity.UpdatedAt,
		}
	}
	return groups, nil
}

var Module = fx.Options(
	fx.Provide(NewIdentityMongoRepository),
)
