package tenantusecase

import (
	"context"

	"github.com/Lockarivault/lockari-app/backend/libs/encryption"
	"github.com/Lockarivault/lockari-app/backend/libs/uuid"
	tenantmodel "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/model"
	repositorytenant "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/repository"
	tenantservice "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/service"
)

var (
	tracerName = "tenant-usecase"
)

/*
PROXIMOS PASSOS PARA IMPLEMENTAÇÃO DO USECASE:

1. [ ] Implementar Create em `create.go`:
   - Validar dados de entrada.
   - Verificar unicidade do Slug no repositório.
   - Criar modelo com Status PENDING e KEK vazia.
   - Persistir no MongoDB via repo.
   - Publicar evento `tenant.created` no RabbitMQ via service.MQ.
   - Retornar o modelo criado para o Handler (202 Accepted).

2. [ ] Implementar GetByID e GetBySlug em `get.go`:
   - Recuperar do cache (Redis) primeiro.
   - Se miss, recuperar do DB (Mongo) e salvar no cache.

3. [ ] Implementar Update em `update.go`:
   - Permitir apenas atualização de metadados básicos (Nome, Descrição).
   - Invalidar cache após atualização.

4. [ ] Implementar Delete em `delete.go`:
   - Soft delete no DB.
   - Remover do cache.

5. [ ] Implementar List em `list.go`:
   - Filtragem básica de tenants.

6. [ ] Implementar Provisioning (Methods utilizados pelo Worker):
   - [ ] Implementar `UpdateSecurityMetadata`: Atualiza a KEK gerada pelo worker.
   - [ ] Implementar `ActivateTenant`: Altera status para ACTIVE.
   - [ ] Implementar `DeactivateTenant`: Altera status para INACTIVE.
   - [ ] Implementar `UpdateProprietiesTypes`: atualiza os tipos de propriedades.
*/

type UsecaseTenant interface {
	Create(ctx context.Context, tenant tenantmodel.TenantModel) (tenantmodel.TenantModel, error)
	GetByID(ctx context.Context, id uuid.UUID) (tenantmodel.TenantModel, error)
	GetBySlug(ctx context.Context, slug string) (tenantmodel.TenantModel, error)
	Update(ctx context.Context, tenant tenantmodel.TenantModel) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter map[string]interface{}) ([]tenantmodel.TenantModel, error)
	CheckSlugAvailability(ctx context.Context, slug string) (bool, error)

	// Métodos de orquestração assíncrona (chamados por consumers)
	UpdateSecurityMetadata(ctx context.Context, id uuid.UUID, metadata encryption.EncryptMetadata) error
	ActivateTenant(ctx context.Context, id uuid.UUID) error
	DeactivateTenant(ctx context.Context, id uuid.UUID) error
	UpdateProprietiesTypes(ctx context.Context, id uuid.UUID, properties tenantmodel.ProprietiesTypes) error
}

type tenant struct {
	repo            repositorytenant.RepositoryTenant
	service         tenantservice.ServiceTenant
	createTenant    CreateTetant
	getTenant       GetTenant
	manageTenant    ManageTenant
	checkSlugTenant CheckSlugTenant
}

func InnicializeUsecaseTenant(
	repo repositorytenant.RepositoryTenant,
	service tenantservice.ServiceTenant,
	create CreateTetant,
	get GetTenant,
	checkSlug CheckSlugTenant,
	manage ManageTenant,
) (UsecaseTenant, error) {
	return &tenant{
		repo:            repo,
		service:         service,
		createTenant:    create,
		getTenant:       get,
		manageTenant:    manage,
		checkSlugTenant: checkSlug,
	}, nil
}

func (u *tenant) Create(ctx context.Context, tenant tenantmodel.TenantModel) (tenantmodel.TenantModel, error) {
	output, err := u.createTenant.Execute(ctx, tenant)
	if err != nil {
		return tenantmodel.TenantModel{}, err
	}

	// Map output back to model
	tenant.ID = output.ID
	tenant.Name = output.Name
	tenant.Description = &output.Description
	tenant.Slug = output.Slug
	tenant.OwnerID = output.OwnerID
	tenant.Status = output.Status

	return tenant, nil
}

func (u *tenant) GetByID(ctx context.Context, id uuid.UUID) (tenantmodel.TenantModel, error) {
	return u.getTenant.ByID(ctx, id)
}

func (u *tenant) GetBySlug(ctx context.Context, slug string) (tenantmodel.TenantModel, error) {
	return u.getTenant.BySlug(ctx, slug)
}

func (u *tenant) Update(ctx context.Context, tenant tenantmodel.TenantModel) error {
	return u.manageTenant.Update(ctx, tenant)
}

func (u *tenant) Delete(ctx context.Context, id uuid.UUID) error {
	return u.manageTenant.Delete(ctx, id)
}

func (u *tenant) List(ctx context.Context, filter map[string]interface{}) ([]tenantmodel.TenantModel, error) {
	return u.manageTenant.List(ctx, filter)
}

func (u *tenant) UpdateSecurityMetadata(ctx context.Context, id uuid.UUID, metadata encryption.EncryptMetadata) error {
	if err := u.repo.UpdateSecurityMetadata(ctx, id, metadata); err != nil {
		return err
	}
	return u.invalidateCacheByID(ctx, id)
}

func (u *tenant) ActivateTenant(ctx context.Context, id uuid.UUID) error {
	if err := u.repo.ActivateTenant(ctx, id); err != nil {
		return err
	}
	return u.invalidateCacheByID(ctx, id)
}

func (u *tenant) DeactivateTenant(ctx context.Context, id uuid.UUID) error {
	if err := u.repo.DeactivateTenant(ctx, id); err != nil {
		return err
	}
	return u.invalidateCacheByID(ctx, id)
}

func (u *tenant) UpdateProprietiesTypes(ctx context.Context, id uuid.UUID, properties tenantmodel.ProprietiesTypes) error {
	if err := u.repo.UpdateProprietiesTypes(ctx, id, properties); err != nil {
		return err
	}
	return u.invalidateCacheByID(ctx, id)
}

func (u *tenant) invalidateCacheByID(ctx context.Context, id uuid.UUID) error {
	tenant, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return u.manageTenant.Update(ctx, tenant) // Update triggers invalidation, or just call invalidator
}

func (u *tenant) CheckSlugAvailability(ctx context.Context, slug string) (bool, error) {
	return u.checkSlugTenant.CheckSlugAvailability(ctx, slug)
}
