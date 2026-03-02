package encryption

import (
	"fmt"
	"time"
)

// Key types
const (
	KeyTypeKEK  = "KEK"
	KeyTypeDEK  = "DEK"
	KeyTypeROOT = "ROOT"
)

// Key statuses
const (
	StatusActive          = "ACTIVE"
	StatusDisabled        = "DISABLED"
	StatusRotated         = "ROTATED"
	StatusPendingDeletion = "PENDING_DELETION"
	StatusCompromised     = "COMPROMISED"
)

// EncryptMetadata defines the cryptographic metadata for KEKs and DEKs.
type EncryptMetadata struct {
	keyID       string
	parentKeyID string
	keyType     string
	status      string
	version     int
	createdAt   time.Time
	updatedAt   time.Time
	expiresAt   time.Time
	rotatedAt   time.Time
	algorithm   string
	provider    string
	fingerprint string
	lastUsedAt  time.Time
	createdBy   string
}

// NewEncryptMetadata creates a new EncryptMetadata with default values.
func NewEncryptMetadata(keyID, keyType, algorithm, provider string) *EncryptMetadata {
	now := time.Now()
	return &EncryptMetadata{
		keyID:     keyID,
		keyType:   keyType,
		status:    StatusActive,
		version:   1,
		createdAt: now,
		updatedAt: now,
		algorithm: algorithm,
		provider:  provider,
	}
}

// Setters and logic methods

// Validate checks if the metadata is consistent and has required fields.
func (m *EncryptMetadata) Validate() error {
	if m == nil {
		return fmt.Errorf("metadata is nil")
	}

	if m.keyID == "" {
		return fmt.Errorf("key_id is required")
	}

	switch m.keyType {
	case KeyTypeKEK, KeyTypeDEK, KeyTypeROOT:
		// valid
	default:
		return fmt.Errorf("invalid key_type: %s", m.keyType)
	}

	if m.algorithm == "" {
		return fmt.Errorf("algorithm is required")
	}

	if m.provider == "" {
		return fmt.Errorf("provider is required")
	}

	if m.status == "" {
		return fmt.Errorf("status is required")
	}

	if m.version <= 0 {
		return fmt.Errorf("version must be greater than 0")
	}

	return nil
}

// Getters

func (m *EncryptMetadata) GetKeyID() string {
	return m.keyID
}

func (m *EncryptMetadata) GetParentKeyID() string {
	return m.parentKeyID
}

func (m *EncryptMetadata) GetKeyType() string {
	return m.keyType
}

func (m *EncryptMetadata) GetStatus() string {
	return m.status
}

func (m *EncryptMetadata) GetVersion() int {
	return m.version
}

func (m *EncryptMetadata) GetCreatedAt() time.Time {
	return m.createdAt
}

func (m *EncryptMetadata) GetUpdatedAt() time.Time {
	return m.updatedAt
}

func (m *EncryptMetadata) GetExpiresAt() time.Time {
	return m.expiresAt
}

func (m *EncryptMetadata) GetRotatedAt() time.Time {
	return m.rotatedAt
}

func (m *EncryptMetadata) GetAlgorithm() string {
	return m.algorithm
}

func (m *EncryptMetadata) GetProvider() string {
	return m.provider
}

func (m *EncryptMetadata) GetFingerprint() string {
	return m.fingerprint
}

func (m *EncryptMetadata) GetLastUsedAt() time.Time {
	return m.lastUsedAt
}

func (m *EncryptMetadata) GetCreatedBy() string {
	return m.createdBy
}

// Set methods (fluents for ease of use)

func (m *EncryptMetadata) WithParentKeyID(id string) *EncryptMetadata {
	m.parentKeyID = id
	return m
}

func (m *EncryptMetadata) WithStatus(status string) *EncryptMetadata {
	m.status = status
	m.updatedAt = time.Now()
	return m
}

func (m *EncryptMetadata) WithVersion(v int) *EncryptMetadata {
	m.version = v
	return m
}

func (m *EncryptMetadata) WithExpiresAt(t time.Time) *EncryptMetadata {
	m.expiresAt = t
	return m
}

func (m *EncryptMetadata) WithFingerprint(f string) *EncryptMetadata {
	m.fingerprint = f
	return m
}

func (m *EncryptMetadata) WithCreatedBy(id string) *EncryptMetadata {
	m.createdBy = id
	return m
}

func (m *EncryptMetadata) MarkAsUsed() *EncryptMetadata {
	m.lastUsedAt = time.Now()
	return m
}

func (m *EncryptMetadata) MarkAsRotated() *EncryptMetadata {
	m.rotatedAt = time.Now()
	m.updatedAt = m.rotatedAt
	return m
}
