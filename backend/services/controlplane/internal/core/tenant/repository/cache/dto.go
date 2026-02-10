package tenantcache

import (
	"time"
)

type securityMetadataDTO struct {
	KeyID       string    `json:"key_id"`
	ParentKeyID string    `json:"parent_key_id,omitempty"`
	KeyType     string    `json:"key_type"`
	Status      string    `json:"status"`
	Version     int       `json:"version"`
	Algorithm   string    `json:"algorithm"`
	Provider    string    `json:"provider"`
	Fingerprint string    `json:"fingerprint,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
	RotatedAt   time.Time `json:"rotated_at,omitempty"`
}

type propertiesDTO struct {
	Items map[string]interface{} `json:"items"`
}

type tenantCacheDTO struct {
	ID               string              `json:"id"`
	Name             string              `json:"name"`
	Description      *string             `json:"description,omitempty"`
	DisplayName      *string             `json:"display_name,omitempty"`
	Slug             string              `json:"slug"`
	OwnerID          string              `json:"owner_id"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
	DeletedAt        *time.Time          `json:"deleted_at,omitempty"`
	Status           string              `json:"status"`
	SecurityMetadata securityMetadataDTO `json:"security_metadata"`
	Properties       propertiesDTO       `json:"properties"`
}
