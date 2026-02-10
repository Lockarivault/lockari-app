package handlertenant

import (
	"fmt"
	"strings"

	"github.com/Lockarivault/lockari-app/backend/libs/uuid"
	tenantmodel "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/model"
)

// CreateTenantRequest represents the payload to create a new tenant.
// @Description Tenant creation request data
type CreateTenantRequest struct {
	Name        string  `json:"name" binding:"required" example:"My Awesome Corp"`            // Tenant display name
	Slug        string  `json:"slug" example:"my-awesome-corp"`                               // Unique slug, will be generated from name if empty
	Description *string `json:"description" example:"This is the main tenant for my company"` // Optional description
	OwnerID     string  `json:"owner_id" binding:"required,uuid" example:"018d867c-3f41-71fb-89c5-34821cc6bd6d"`
}

// UpdateTenantRequest represents the payload to update an existing tenant.
// @Description Tenant update request data
type UpdateTenantRequest struct {
	Name        string  `json:"name" binding:"required" example:"Updated Corp Name"`
	Description *string `json:"description" example:"Updated description"`
	OwnerID     string  `json:"owner_id" binding:"required,uuid" example:"018d867c-3f41-71fb-89c5-34821cc6bd6d"`
}

// TenantResponse represents the standard tenant data returned by the API.
// @Description Tenant response data
type TenantResponse struct {
	ID          string                 `json:"id" example:"018d867c-3f41-71fb-89c5-34821cc6bd6d"`
	Name        string                 `json:"name" example:"My Awesome Corp"`
	Slug        string                 `json:"slug" example:"my-awesome-corp"`
	Description *string                `json:"description" example:"This is the main tenant for my company"`
	Status      tenantmodel.StatusType `json:"status" example:"pending"`
	CreatedAt   string                 `json:"created_at" example:"2024-02-07T12:00:00Z"`
	UpdatedAt   string                 `json:"updated_at" example:"2024-02-07T12:00:00Z"`
}

func MapToResponse(m tenantmodel.TenantModel) TenantResponse {
	return TenantResponse{
		ID:          m.ID.String(),
		Name:        m.Name,
		Slug:        m.Slug,
		Description: m.Description,
		Status:      m.Status,
		CreatedAt:   m.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   m.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// MapListToResponse converts a list of Domain Models to a list of Response DTOs.
func MapListToResponse(list []tenantmodel.TenantModel) []TenantResponse {
	res := make([]TenantResponse, len(list))
	for i, v := range list {
		res[i] = MapToResponse(v)
	}
	return res
}

// ErrorResponse represents a standard error response.
// @Description Error details
type ErrorResponse struct {
	Error   string `json:"error" example:"invalid request data"`
	Message string `json:"message" example:"the name field is required"`
}

func (t *CreateTenantRequest) Validate() error {
	if t.Name == "" {
		return fmt.Errorf("name is required")
	}
	if !uuid.IsValid(t.OwnerID) {
		return fmt.Errorf("owner_id is required")
	}
	if t.Slug != "" {
		t.Slug = strings.TrimSpace(strings.ReplaceAll(strings.ToLower(t.Slug), " ", "-"))
	}

	return nil
}
