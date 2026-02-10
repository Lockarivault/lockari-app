package tenantservice

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Lockarivault/lockari-app/backend/libs/database/cache"
	"github.com/Lockarivault/lockari-app/backend/libs/encryption"
	"github.com/Lockarivault/lockari-app/backend/libs/loggers"
	"github.com/Lockarivault/lockari-app/backend/libs/mensageria"
	"github.com/Lockarivault/lockari-app/backend/libs/uuid"
	"github.com/Lockarivault/lockari-app/backend/services/controlplane/config"
	tenantmodel "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/model"
	repositorytenant "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/repository"
	"github.com/rabbitmq/amqp091-go"
)

type ServiceTenant interface {
	PublishTenantCreated(ctx context.Context, tenant tenantmodel.TenantModel) error
	SubscribeTenantCreated(ctx context.Context, handler func(tenantmodel.TenantModel) error) error
}

type KeyProvider interface {
	GetTenantEncryptor(ctx context.Context, tenantID uuid.UUID) (encryption.Encryptor, error)
}

type tenant struct {
	logger   loggers.LoggerInterface
	cache    cache.RedisClientInterface
	mq       mensageria.MessageQueue
	repo     repositorytenant.RepositoryTenant
	keyRepo  repositorytenant.RepositoryKey
	vaultCfg config.VaultConfig
}

func InnicializeServiceTenant(
	logger loggers.LoggerInterface,
	cache cache.RedisClientInterface,
	mq mensageria.MessageQueue,
	repo repositorytenant.RepositoryTenant,
	keyRepo repositorytenant.RepositoryKey,
	cfg *config.Connections,
) (ServiceTenant, KeyProvider, error) {
	m := tenant{
		logger:   logger,
		cache:    cache,
		mq:       mq,
		repo:     repo,
		keyRepo:  keyRepo,
		vaultCfg: cfg.Vault,
	}
	return m, m, nil
}

func (t tenant) PublishTenantCreated(ctx context.Context, tenant tenantmodel.TenantModel) error {
	payload, err := json.Marshal(tenant)
	if err != nil {
		return fmt.Errorf("failed to marshal tenant: %w", err)
	}

	msg := amqp091.Publishing{
		ContentType: "application/json",
		Body:        payload,
		Timestamp:   time.Now(),
	}

	return t.mq.Publish(ctx, "tenant.events", "tenant.created", msg)
}

func (t tenant) SubscribeTenantCreated(ctx context.Context, handler func(tenantmodel.TenantModel) error) error {
	// The queue name should be unique for this consumer purpose (provisioning)
	queueName := "tenant.provisioning"

	return t.mq.Consume(ctx, queueName, true, "tenant-worker", func(msg amqp091.Delivery) error {
		var tenant tenantmodel.TenantModel
		if err := json.Unmarshal(msg.Body, &tenant); err != nil {
			t.logger.Error("failed to unmarshal tenant from queue", "error", err)
			return err
		}

		if err := handler(tenant); err != nil {
			t.logger.Error("failed to handle tenant created event", "error", err, "tenant_id", tenant.ID)
			return err
		}

		return nil
	})
}
func (t tenant) GetTenantEncryptor(ctx context.Context, tenantID uuid.UUID) (encryption.Encryptor, error) {
	t.logger.Info("retrieving tenant encryptor", "tenant_id", tenantID)

	// 1. Get Security Metadata to find the KeyID
	tenant, err := t.repo.GetByID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant security metadata: %w", err)
	}

	keyID := tenant.SecurityMetadata.GetKeyID()
	if keyID == "" {
		return nil, fmt.Errorf("tenant has no security metadata (KEK) provisioned")
	}

	// 2. Get Encrypted KEK from separate collection
	ciphertext, nonce, _, err := t.keyRepo.Get(ctx, keyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get encrypted KEK from repository: %w", err)
	}

	// 3. Decrypt KEK (L2) using Root Key (L1)
	parentKeyID := tenant.SecurityMetadata.GetParentKeyID()
	if parentKeyID == "" {
		parentKeyID = t.vaultCfg.RootKeyID
	}

	// Try to find the key in the RootKeys map first (new structure)
	rootKeyBase64, ok := t.vaultCfg.RootKeys[parentKeyID]
	if !ok {
		// Fallback for transition phase: if the ID matches the legacy one, use the legacy field
		if parentKeyID == t.vaultCfg.RootKeyID && t.vaultCfg.RootKey != "" {
			rootKeyBase64 = t.vaultCfg.RootKey
		} else {
			return nil, fmt.Errorf("root key %s not found in configuration", parentKeyID)
		}
	}

	rootKey, err := encryption.DecodeBase64(rootKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode root key %s: %w", parentKeyID, err)
	}

	rootEncryptor, err := encryption.NewEncryptor(rootKey, parentKeyID)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize root encryptor: %w", err)
	}

	// Envelope for KEK decryption
	envelope := &encryption.Envelope{
		Ciphertext: encryption.EncodeBase64(ciphertext),
		Nonce:      encryption.EncodeBase64(nonce),
		KeyID:      parentKeyID,
	}

	kek, err := rootEncryptor.DecryptDEK(ctx, envelope)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt tenant KEK with root key: %w", err)
	}

	// 4. Return Encryptor configured with Decrypted KEK
	return encryption.NewEncryptor(kek, keyID)
}
