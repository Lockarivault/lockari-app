package tenantmodel

import (
	"errors"
	"time"
)

type SecurityProvider string

type TenantSecurity struct {
	KekID     string
	Algorithm string
	Provider  SecurityProvider // ex: "aws-kms", "internal"
	RotatedAt *time.Time
	Version   int
}

func newTenantSecurity() *TenantSecurity {
	return &TenantSecurity{}
}

func (t *TenantSecurity) SetKekID(kekID string) error {
	if kekID == "" {
		return errors.New("kek id is required")
	}
	t.KekID = kekID

	return nil
}

func (t *TenantSecurity) SetAlgorithm(algorithm string) error {
	if algorithm == "" {
		return errors.New("algorithm is required")
	}
	t.Algorithm = algorithm

	return nil
}

func (t *TenantSecurity) SetProvider(provider string) error {
	if provider == "" {
		return errors.New("provider is required")
	}
	t.Provider = SecurityProvider(provider)
	return nil
}

func (t *TenantSecurity) SetRotatedAt(rotatedAt time.Time) error {
	if rotatedAt.IsZero() {
		return errors.New("rotated at is required")
	}

	t.RotatedAt = &rotatedAt

	return nil
}

func (t *TenantSecurity) SetVersion() {
	t.Version = t.Version + 1
}

func (t *TenantSecurity) GetKekID() string {
	return t.KekID
}

func (t *TenantSecurity) GetAlgorithm() string {
	return t.Algorithm
}

func (t *TenantSecurity) GetProvider() string {
	return string(t.Provider)
}

func (t *TenantSecurity) GetRotatedAt() *time.Time {
	return t.RotatedAt
}

func (t *TenantSecurity) GetVersion() int {
	return t.Version
}
