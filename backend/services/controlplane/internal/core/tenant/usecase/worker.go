package tenantusecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/Lockarivault/lockari-app/backend/libs/database/cache"
	"github.com/Lockarivault/lockari-app/backend/libs/encryption"
	"github.com/Lockarivault/lockari-app/backend/libs/loggers"
	"github.com/Lockarivault/lockari-app/backend/libs/telemetry"
	"github.com/Lockarivault/lockari-app/backend/libs/uuid"
	"github.com/Lockarivault/lockari-app/backend/services/controlplane/config"
	auditmodel "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/audit/model"
	auditusecase "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/audit/usecase"
	tenantmodel "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/model"
	repositorytenant "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/repository"
	tenantservice "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/service"
	"github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/pkg/notifications"
	"github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/pkg/webhooks"
	"go.uber.org/fx"
)

type workerTenant struct {
	repo           repositorytenant.RepositoryTenant
	keyRepo        repositorytenant.RepositoryKey
	service        tenantservice.ServiceTenant
	seedService    tenantservice.SeedService
	auditService   auditusecase.AuditService
	notifService   notifications.NotificationService
	webhookService webhooks.WebhookService
	cache          cache.RedisClientInterface
	logger         loggers.LoggerInterface
	telemetry      telemetry.OtelObservability
	cfg            *config.Connections
}

type WorkerTenant interface {
	Execute(ctx context.Context, input tenantmodel.TenantModel) (*tenantmodel.TenantModel, error)
}

func NewWorkerTenant(
	lc fx.Lifecycle,
	repo repositorytenant.RepositoryTenant,
	keyRepo repositorytenant.RepositoryKey,
	service tenantservice.ServiceTenant,
	seedService tenantservice.SeedService,
	auditService auditusecase.AuditService,
	notifService notifications.NotificationService,
	webhookService webhooks.WebhookService,
	cache cache.RedisClientInterface,
	logger loggers.LoggerInterface,
	telemetry telemetry.OtelObservability,
	cfg *config.Connections,
) (WorkerTenant, error) {
	if repo == nil {
		return nil, errors.New("repo is nil")
	}
	if keyRepo == nil {
		return nil, errors.New("keyRepo is nil")
	}
	if service == nil {
		return nil, errors.New("service is nil")
	}
	if seedService == nil {
		return nil, errors.New("seedService is nil")
	}
	if auditService == nil {
		return nil, errors.New("auditService is nil")
	}
	if notifService == nil {
		return nil, errors.New("notifService is nil")
	}
	if webhookService == nil {
		return nil, errors.New("webhookService is nil")
	}
	if cache == nil {
		return nil, errors.New("cache is nil")
	}
	if logger == nil {
		return nil, errors.New("logger is nil")
	}
	if telemetry == nil {
		return nil, errors.New("telemetry is nil")
	}
	if cfg == nil {
		return nil, errors.New("cfg is nil")
	}

	u := &workerTenant{
		repo:         repo,
		keyRepo:      keyRepo,
		service:      service,
		seedService:  seedService,
		auditService: auditService,
		cache:        cache,
		logger:       logger,
		telemetry:    telemetry,
		cfg:          cfg,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go u.worker(context.Background())
			return nil
		},
	})

	return u, nil
}

func (u *workerTenant) Execute(ctx context.Context, input tenantmodel.TenantModel) (*tenantmodel.TenantModel, error) {
	u.logger.Info("provisioning tenant", "id", input.ID, "slug", input.Slug)

	// Idempotency: Check if tenant is already active or has security metadata
	if input.Status == tenantmodel.StatusActive && input.SecurityMetadata.GetKeyID() != "" {
		u.logger.Info("tenant already provisioned, skipping", "id", input.ID)
		return &input, nil
	}

	var keyID string
	var kek []byte

	// Step 7: Idempotency - check if we already have a key assigned (from a previous failed run)
	if input.SecurityMetadata.GetKeyID() != "" {
		keyID = input.SecurityMetadata.GetKeyID()
		u.logger.Info("tenant already has a key assigned, resuming provisioning", "tenant_id", input.ID, "key_id", keyID)
	} else {
		// 1. Generate KEK (Key Encryption Key) - 32 bytes for AES-256
		var err error
		kek, err = encryption.GenerateRandomKey(32)
		if err != nil {
			return nil, fmt.Errorf("failed to generate KEK: %w", err)
		}

		// 2. Encrypt KEK with Root Key (L1 -> L2)
		rootKeyID := u.cfg.Vault.RootKeyID
		if rootKeyID == "" {
			rootKeyID = "root-key-internal" // Fallback
		}

		rootKeyBase64, ok := u.cfg.Vault.RootKeys[rootKeyID]
		if !ok {
			// Fallback to legacy field if not in map
			rootKeyBase64 = u.cfg.Vault.RootKey
		}

		if rootKeyBase64 == "" {
			return nil, fmt.Errorf("root key %s material not found in configuration", rootKeyID)
		}

		rootKey, err := encryption.DecodeBase64(rootKeyBase64)
		if err != nil {
			return nil, fmt.Errorf("failed to decode root key: %w", err)
		}

		encryptor, err := encryption.NewEncryptor(rootKey, rootKeyID)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize root encryptor: %w", err)
		}

		envelope, err := encryptor.EncryptDEKWithKEK(ctx, kek)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt KEK with root key: %w", err)
		}

		// 3. Save Encrypted KEK in 'keys' collection
		keyID = uuid.New().String()
		ciphertext, _ := encryption.DecodeBase64(envelope.Ciphertext)
		nonce, _ := encryption.DecodeBase64(envelope.Nonce)
		if err := u.keyRepo.Save(ctx, keyID, input.ID, ciphertext, nonce, "AES-256-GCM"); err != nil {
			return nil, fmt.Errorf("failed to save encrypted KEK: %w", err)
		}

		// 4. Update Security Metadata in 'tenants' collection
		metadata := encryption.NewEncryptMetadata(
			keyID,
			encryption.KeyTypeKEK,
			"AES-256-GCM",
			"InternalVault",
		).WithParentKeyID(rootKeyID)

		if err := u.repo.UpdateSecurityMetadata(ctx, input.ID, *metadata); err != nil {
			// Step 4 & 5: Cleanup Orphan Key if Metadata fails
			u.logger.Error("failed to update security metadata, cleaning up orphan key", "error", err, "tenant_id", input.ID, "key_id", keyID)
			_ = u.keyRepo.Delete(ctx, keyID)

			auditLogCleanup := auditmodel.NewAuditLog(
				auditmodel.ActorInfo{ID: "system-provisioner", Type: auditmodel.ActorSystem},
				"CLEANUP_SUCCESS",
				auditmodel.TargetInfo{ID: keyID, Type: auditmodel.TargetKey, Name: "Orphan Key Cleanup"},
			)
			_ = u.auditService.Log(ctx, auditLogCleanup)

			return nil, fmt.Errorf("failed to update security metadata: %w", err)
		}
	}

	// 5. Apply default Quotas and Properties from configuration
	input.Properties.EnsureDefaults()
	input.Properties.SetMaxSecrets(u.cfg.Quotas.MaxSecrets)
	input.Properties.SetMaxUsers(u.cfg.Quotas.MaxUsers)
	input.Properties.SetMaxStorageBytes(u.cfg.Quotas.MaxStorageBytes)

	// Step 6: DNS for the new tenant (Item 6)
	baseDomain := u.cfg.App.BaseDomain
	if baseDomain == "" {
		baseDomain = "lockari.com" // Safety fallback
	}
	fqdn := fmt.Sprintf("%s.%s", input.Slug, baseDomain)
	input.Properties.SetFullyQualifiedDomain(fqdn)

	if err := u.repo.UpdateProprietiesTypes(ctx, input.ID, input.Properties); err != nil {
		u.logger.Warn("failed to apply default quotas", "error", err, "tenant_id", input.ID)
		// We don't fail the whole provisioning for quotas, as they can be set manually later
	}

	// 6. Activate Tenant
	if err := u.repo.ActivateTenant(ctx, input.ID); err != nil {
		return nil, fmt.Errorf("failed to activate tenant: %w", err)
	}

	u.logger.Info("tenant provisioned and activated successfully", "id", input.ID, "slug", input.Slug, "key_id", keyID)

	// Audit Log (Activity 6)
	auditLog := auditmodel.NewAuditLog(
		auditmodel.ActorInfo{
			ID:   "system-provisioning-worker",
			Type: auditmodel.ActorSystem,
			IP:   "internal",
		},
		"TENANT_PROVISION_SUCCESS",
		auditmodel.TargetInfo{
			ID:   input.ID.String(),
			Type: auditmodel.TargetTenant,
			Name: input.Name,
		},
	)
	auditLog.Outcome.Status = auditmodel.OutcomeSuccess
	auditLog.Context.TenantID = input.ID

	if err := u.auditService.Log(ctx, auditLog); err != nil {
		u.logger.Warn("failed to save audit log for tenant provisioning", "error", err, "tenant_id", input.ID)
	}

	// Audit Log for DNS (Item 6)
	auditLogDNS := auditmodel.NewAuditLog(
		auditmodel.ActorInfo{ID: "system-provisioner", Type: auditmodel.ActorSystem},
		"DNS_CONFIGURED",
		auditmodel.TargetInfo{ID: input.ID.String(), Type: auditmodel.TargetTenant, Name: fqdn},
	)
	auditLogDNS.Outcome.Status = auditmodel.OutcomeSuccess
	auditLogDNS.Context.TenantID = input.ID
	_ = u.auditService.Log(ctx, auditLogDNS)

	// 7. Seed Initial Data (Step 8)
	if err := u.seedService.ProvisionSeedData(ctx, input.ID); err != nil {
		u.logger.Warn("failed to provision seed data", "error", err, "tenant_id", input.ID)
		// We don't fail the whole provisioning for seed data, can be fixed manually
	}

	// 8. Notifications & Onboarding (Step 4)
	u.triggerNotifications(ctx, input)

	input.Status = tenantmodel.StatusActive
	return &input, nil
}

func (u *workerTenant) triggerNotifications(ctx context.Context, tenant tenantmodel.TenantModel) {
	// 1. Send Welcome Email
	emailBody := fmt.Sprintf("Welcome to Lockari, %s! Your tenant is now active and ready to use.", tenant.Name)
	if err := u.notifService.SendEmail(ctx, "admin@"+tenant.Slug+".lockari.com", "Welcome to Lockari", emailBody); err != nil {
		u.logger.Warn("failed to send welcome email", "error", err, "tenant_id", tenant.ID)
	}

	// 2. Trigger Webhook
	webhookURL := "" // This would come from dynamic config or subscription data
	if err := u.webhookService.Send(ctx, webhookURL, "TENANT_ACTIVATED", tenant); err != nil {
		u.logger.Warn("failed to trigger onboarding webhook", "error", err, "tenant_id", tenant.ID)
	}

	// 3. Audit Notification
	notificationLog := auditmodel.NewAuditLog(
		auditmodel.ActorInfo{ID: "system-provisioner", Type: auditmodel.ActorSystem},
		"NOTIFICATION_SENT",
		auditmodel.TargetInfo{ID: tenant.ID.String(), Type: auditmodel.TargetTenant, Name: tenant.Name},
	)
	notificationLog.Outcome.Status = auditmodel.OutcomeSuccess
	notificationLog.Context.TenantID = tenant.ID
	_ = u.auditService.Log(ctx, notificationLog)
}

func (u *workerTenant) worker(ctx context.Context) {
	u.logger.Info("starting tenant provisioning worker")

	err := u.service.SubscribeTenantCreated(ctx, func(tenant tenantmodel.TenantModel) error {
		u.logger.Info("received tenant created event", "id", tenant.ID, "slug", tenant.Slug)

		_, err := u.Execute(ctx, tenant)
		if err != nil {
			u.logger.Error("failed to provision tenant, marking as failed", "error", err, "tenant_id", tenant.ID)
			// Mark as failed in DB for UI feedback
			if failErr := u.repo.FailTenant(ctx, tenant.ID, err.Error()); failErr != nil {
				u.logger.Error("failed to mark tenant as failed in repository", "error", failErr, "tenant_id", tenant.ID)
			}
			return err // Return error to let MQ handle retries/DLQ
		}
		return nil
	})

	if err != nil {
		u.logger.Error("failed to start tenant provisioning subscription", "error", err)
	}
}
