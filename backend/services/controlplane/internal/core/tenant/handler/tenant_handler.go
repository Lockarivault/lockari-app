package handlertenant

import (
	"net/http"

	"github.com/Lockarivault/lockari-app/backend/libs/uuid"
	tenantmodel "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/tenant/model"
	"github.com/gin-gonic/gin"
)

// Create handles the synchronous phase of tenant creation.
// @Summary Create a new tenant
// @Description Creates a basic tenant identity and triggers asynchronous provisioning.
// @Tags Tenants
// @Accept json
// @Produce json
// @Param request body CreateTenantRequest true "Tenant creation data"
// @Success 202 {object} TenantResponse "Tenant created (pending provisioning)"
// @Failure 400 {object} ErrorResponse "Invalid request data"
// @Failure 409 {object} ErrorResponse "Slug already exists"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/tenants [post]
func (h *tenant) Create(c *gin.Context) {
	var req CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		id := uuid.New()
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: err.Error() + "\n " + id.String()})
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: err.Error()})
		return
	}

	ownerID, err := uuid.Parse(req.OwnerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_owner_id", Message: "invalid OwnerID format"})
		return
	}

	model := tenantmodel.NewTenantModel(req.Name, req.Slug, req.Description, ownerID)

	result, err := h.usecase.Create(c.Request.Context(), model)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, MapToResponse(result))
}

// GetByID retrieves a tenant by its unique ID.
// @Summary Get tenant by ID
// @Description Retrieves a tenant's details using its UUID.
// @Tags Tenants
// @Produce json
// @Param id path string true "Tenant ID"
// @Success 200 {object} TenantResponse "Tenant found"
// @Failure 404 {object} ErrorResponse "Tenant not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/tenants/id/{id} [get]
func (h *tenant) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "invalid UUID format"})
		return
	}

	result, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "tenant not found"})
		return
	}

	c.JSON(http.StatusOK, MapToResponse(result))
}

// GetBySlug retrieves a tenant by its unique slug.
// @Summary Get tenant by slug
// @Description Retrieves a tenant's details using its slug.
// @Tags Tenants
// @Produce json
// @Param slug path string true "Tenant Slug"
// @Success 200 {object} TenantResponse "Tenant found"
// @Failure 404 {object} ErrorResponse "Tenant not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/tenants/slug/{slug} [get]
func (h *tenant) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")

	result, err := h.usecase.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "tenant not found"})
		return
	}

	c.JSON(http.StatusOK, MapToResponse(result))
}

// List retrieves a list of tenants based on filters.
// @Summary List tenants
// @Description Retrieves a list of tenants.
// @Tags Tenants
// @Produce json
// @Success 200 {array} TenantResponse "List of tenants"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/tenants [get]
func (h *tenant) List(c *gin.Context) {
	// Simple list for now, no filters implemented in handler yet
	result, err := h.usecase.List(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, MapListToResponse(result))
}

// Delete removes a tenant.
// @Summary Delete a tenant
// @Description Soft deletes a tenant by its UUID.
// @Tags Tenants
// @Param id path string true "Tenant ID"
// @Success 204 "Tenant deleted"
// @Failure 404 {object} ErrorResponse "Tenant not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/tenants/{id} [delete]
func (h *tenant) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "invalid UUID format"})
		return
	}

	if err := h.usecase.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// Update updates an existing tenant's basic information.
// @Summary Update a tenant
// @Description Updates a tenant's name and description.
// @Tags Tenants
// @Accept json
// @Produce json
// @Param id path string true "Tenant ID"
// @Param request body UpdateTenantRequest true "Updated tenant data"
// @Success 200 {object} TenantResponse "Tenant updated"
// @Failure 400 {object} ErrorResponse "Invalid request data"
// @Failure 404 {object} ErrorResponse "Tenant not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/tenants/{id} [put]
func (h *tenant) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_id", Message: "invalid UUID format"})
		return
	}

	var req UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: err.Error()})
		return
	}

	// We need to fetch the existing tenant first to update it securely
	existing, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "not_found", Message: "tenant not found"})
		return
	}

	ownerID, err := uuid.Parse(req.OwnerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_owner_id", Message: "invalid OwnerID format"})
		return
	}

	existing.Name = req.Name
	existing.Description = req.Description
	existing.OwnerID = ownerID

	if err := h.usecase.Update(c.Request.Context(), existing); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, MapToResponse(existing))
}

// CheckSlugAvailability checks if a slug is available for a tenant.
// @Summary Check slug availability
// @Description Checks if a slug is available for a tenant.
// @Tags Tenants
// @Produce json
// @Param slug query string true "Slug to check"
// @Success 200 {object} bool "Slug availability"
// @Failure 400 {object} ErrorResponse "Invalid request data"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/tenants/check-slug [get]
func (h *tenant) CheckSlugAvailability(c *gin.Context) {
	slug := c.Query("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "slug is required"})
		return
	}

	available, err := h.usecase.CheckSlugAvailability(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, available)
}
