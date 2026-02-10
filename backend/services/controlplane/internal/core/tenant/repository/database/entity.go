package tenantdatabase

import (
	"time"
)

type securityMetadataEntity struct {
	KeyID       string    `bson:"key_id"`
	ParentKeyID string    `bson:"parent_key_id,omitempty"`
	KeyType     string    `bson:"key_type"`
	Status      string    `bson:"status"`
	Version     int       `bson:"version"`
	Algorithm   string    `bson:"algorithm"`
	Provider    string    `bson:"provider"`
	Fingerprint string    `bson:"fingerprint,omitempty"`
	CreatedAt   time.Time `bson:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at"`
	ExpiresAt   time.Time `bson:"expires_at,omitempty"`
	RotatedAt   time.Time `bson:"rotated_at,omitempty"`
}

type propertiesEntity struct {
	Items map[string]interface{} `bson:"items"`
}

type tenantEntity struct {
	ID               string                 `bson:"_id"`
	Name             string                 `bson:"name"`
	Description      *string                `bson:"description,omitempty"`
	DisplayName      *string                `bson:"display_name,omitempty"`
	Slug             string                 `bson:"slug"`
	OwnerID          string                 `bson:"owner_id"`
	CreatedAt        time.Time              `bson:"created_at"`
	UpdatedAt        time.Time              `bson:"updated_at"`
	DeletedAt        *time.Time             `bson:"deleted_at,omitempty"`
	Status           string                 `bson:"status"`
	FailureReason    *string                `bson:"failure_reason,omitempty"`
	SecurityMetadata securityMetadataEntity `bson:"security_metadata"`
	Properties       propertiesEntity       `bson:"properties"`
}
