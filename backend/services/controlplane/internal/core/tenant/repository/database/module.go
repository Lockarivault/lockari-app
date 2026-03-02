package tenantrepositorydatabase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/lockarivault/lockari-app/backend/libs/database/nosql"
	tenantmodel "github.com/lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/model"
	tenantrepository "github.com/lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TenantDatabase defines the interface for tenant database operations.
// It embeds tenantrepository.TenantRepository to ensure it implements
// the common repository interface used by the core repository.
type TenantDatabase interface {
	tenantrepository.TenantDBRepository
}

// tenant provides an implementation of TenantDatabase.
// This exists to handle persistent storage of tenant data, ensuring
// data integrity and durability across system restarts.
type tenant struct {
	db *mongo.Collection
}

// NewTenantDatabaseRepository creates a new instance of TenantDatabase.
func NewTenantDatabaseRepository(db nosql.DatabaseService) (TenantDatabase, error) {
	if db == nil {
		return nil, errors.New("database service is required")
	}

	err := db.Connect(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	collection, err := db.Collection("tenants")
	if err != nil {
		return nil, fmt.Errorf("error getting collection: %w", err)
	}

	return &tenant{db: collection}, nil
}

// GetByID retrieves a tenant from the database by its ID.
func (t *tenant) GetByID(ctx context.Context, id string) (*tenantmodel.TenantModel, error) {
	if id == "" {
		return nil, errors.New("id is required")
	}

	var p tenantrepository.TenantPersistence
	err := t.db.FindOne(ctx, bson.M{"_id": id, "deleted_at": nil}).Decode(&p)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("error finding tenant: %w", err)
	}

	return tenantrepository.ConvertToTenantModel(&p), nil
}

// Create persists a new tenant in the database.
func (t *tenant) Create(ctx context.Context, tenant *tenantmodel.TenantModel) error {
	if tenant == nil {
		return errors.New("tenant is required")
	}

	p := tenantrepository.ConvertToTenantStorage(tenant)
	_, err := t.db.InsertOne(ctx, p)
	if err != nil {
		return fmt.Errorf("error inserting tenant: %w", err)
	}

	return nil
}

// Update modifies an existing tenant in the database.
func (t *tenant) Update(ctx context.Context, tenant *tenantmodel.TenantModel) error {
	if tenant == nil {
		return errors.New("tenant is required")
	}

	p := tenantrepository.ConvertToTenantStorage(tenant)
	_, err := t.db.ReplaceOne(ctx, bson.M{"_id": p.ID}, p)
	if err != nil {
		return fmt.Errorf("error updating tenant: %w", err)
	}

	return nil
}

// Delete performs a soft delete of a tenant in the database.
func (t *tenant) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id is required")
	}

	_, err := t.db.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"deleted_at": time.Now().UTC()}},
	)
	if err != nil {
		return fmt.Errorf("error deleting tenant: %w", err)
	}

	return nil
}

// List retrieves all non-deleted tenants from the database with pagination.
func (t *tenant) List(ctx context.Context, p tenantrepository.Pagination) ([]*tenantmodel.TenantModel, error) {
	findOptions := options.Find()
	if p.Limit > 0 {
		findOptions.SetLimit(p.Limit)
	}
	if p.Skip > 0 {
		findOptions.SetSkip(p.Skip)
	}

	cursor, err := t.db.Find(ctx, bson.M{"deleted_at": nil}, findOptions)
	if err != nil {
		return nil, fmt.Errorf("error listing tenants: %w", err)
	}
	defer cursor.Close(ctx)

	var tenants []*tenantmodel.TenantModel
	for cursor.Next(ctx) {
		var p tenantrepository.TenantPersistence
		if err := cursor.Decode(&p); err != nil {
			return nil, fmt.Errorf("error decoding tenant: %w", err)
		}
		tenants = append(tenants, tenantrepository.ConvertToTenantModel(&p))
	}

	return tenants, nil
}

// Filter retrieves tenants based on a filter map with pagination.
func (t *tenant) Filter(ctx context.Context, filter map[string]any, p tenantrepository.Pagination) ([]*tenantmodel.TenantModel, error) {
	if filter == nil {
		filter = make(map[string]any)
	}
	// Always exclude deleted tenants unless explicitly requested
	if _, ok := filter["deleted_at"]; !ok {
		filter["deleted_at"] = nil
	}

	findOptions := options.Find()
	if p.Limit > 0 {
		findOptions.SetLimit(p.Limit)
	}
	if p.Skip > 0 {
		findOptions.SetSkip(p.Skip)
	}

	cursor, err := t.db.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("error filtering tenants: %w", err)
	}
	defer cursor.Close(ctx)

	var tenants []*tenantmodel.TenantModel
	for cursor.Next(ctx) {
		var p tenantrepository.TenantPersistence
		if err := cursor.Decode(&p); err != nil {
			return nil, fmt.Errorf("error decoding tenant: %w", err)
		}
		tenants = append(tenants, tenantrepository.ConvertToTenantModel(&p))
	}

	return tenants, nil
}
