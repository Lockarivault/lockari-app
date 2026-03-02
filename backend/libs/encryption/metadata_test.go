package encryption

import (
	"testing"
)

func TestNewEncryptMetadata(t *testing.T) {
	keyID := "key-123"
	keyType := KeyTypeDEK
	algorithm := "AES_GCM_256"
	provider := "AWS_KMS"

	m := NewEncryptMetadata(keyID, keyType, algorithm, provider)

	if m.GetKeyID() != keyID {
		t.Errorf("expected KeyID %s, got %s", keyID, m.GetKeyID())
	}
	if m.GetKeyType() != keyType {
		t.Errorf("expected KeyType %s, got %s", keyType, m.GetKeyType())
	}
	if m.GetAlgorithm() != algorithm {
		t.Errorf("expected Algorithm %s, got %s", algorithm, m.GetAlgorithm())
	}
	if m.GetProvider() != provider {
		t.Errorf("expected Provider %s, got %s", provider, m.GetProvider())
	}
	if m.GetStatus() != StatusActive {
		t.Errorf("expected status %s, got %s", StatusActive, m.GetStatus())
	}
	if m.GetVersion() != 1 {
		t.Errorf("expected version 1, got %d", m.GetVersion())
	}
	if m.GetCreatedAt().IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestEncryptMetadata_Validate(t *testing.T) {
	tests := []struct {
		name    string
		m       *EncryptMetadata
		wantErr bool
	}{
		{
			name:    "Valid Metadata",
			m:       NewEncryptMetadata("id", KeyTypeKEK, "alg", "prov"),
			wantErr: false,
		},
		{
			name: "Missing ID",
			m: &EncryptMetadata{
				keyType:   KeyTypeKEK,
				algorithm: "alg",
				provider:  "prov",
				status:    StatusActive,
				version:   1,
			},
			wantErr: true,
		},
		{
			name: "Invalid Key Type",
			m: &EncryptMetadata{
				keyID:     "id",
				keyType:   "INVALID",
				algorithm: "alg",
				provider:  "prov",
				status:    StatusActive,
				version:   1,
			},
			wantErr: true,
		},
		{
			name: "Zero Version",
			m: &EncryptMetadata{
				keyID:     "id",
				keyType:   KeyTypeKEK,
				algorithm: "alg",
				provider:  "prov",
				status:    StatusActive,
				version:   0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.m.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEncryptMetadata_WithMethods(t *testing.T) {
	m := NewEncryptMetadata("id", KeyTypeDEK, "alg", "prov")

	parentID := "parent-456"
	m.WithParentKeyID(parentID)
	if m.GetParentKeyID() != parentID {
		t.Errorf("expected ParentKeyID %s, got %s", parentID, m.GetParentKeyID())
	}

	m.WithStatus(StatusRotated)
	if m.GetStatus() != StatusRotated {
		t.Errorf("expected status %s, got %s", StatusRotated, m.GetStatus())
	}

	fingerprint := "hash123"
	m.WithFingerprint(fingerprint)
	if m.GetFingerprint() != fingerprint {
		t.Errorf("expected fingerprint %s, got %s", fingerprint, m.GetFingerprint())
	}

	m.MarkAsUsed()
	if m.GetLastUsedAt().IsZero() {
		t.Error("expected LastUsedAt to be set")
	}

	m.MarkAsRotated()
	if m.GetRotatedAt().IsZero() {
		t.Error("expected RotatedAt to be set")
	}
}
