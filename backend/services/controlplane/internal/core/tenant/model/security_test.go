package tenantmodel

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTenantSecurity(t *testing.T) {
	ts := newTenantSecurity()
	assert.NotNil(t, ts)
	assert.Empty(t, ts.GetKekID())
	assert.Empty(t, ts.GetAlgorithm())
	assert.Empty(t, ts.GetProvider())
	assert.Nil(t, ts.GetRotatedAt())
	assert.Equal(t, 0, ts.GetVersion())
}

func TestTenantSecurity_SetKekID(t *testing.T) {
	ts := newTenantSecurity()

	tests := []struct {
		name    string
		kekID   string
		wantErr bool
	}{
		{
			name:    "valid kek id",
			kekID:   "kek-123",
			wantErr: false,
		},
		{
			name:    "empty kek id",
			kekID:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ts.SetKekID(tt.kekID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.kekID, ts.GetKekID())
			}
		})
	}
}

func TestTenantSecurity_SetAlgorithm(t *testing.T) {
	ts := newTenantSecurity()

	tests := []struct {
		name      string
		algorithm string
		wantErr   bool
	}{
		{
			name:      "valid algorithm",
			algorithm: "AES-256-GCM",
			wantErr:   false,
		},
		{
			name:      "empty algorithm",
			algorithm: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ts.SetAlgorithm(tt.algorithm)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.algorithm, ts.GetAlgorithm())
			}
		})
	}
}

func TestTenantSecurity_SetProvider(t *testing.T) {
	ts := newTenantSecurity()

	tests := []struct {
		name     string
		provider string
		wantErr  bool
	}{
		{
			name:     "valid provider",
			provider: "aws-kms",
			wantErr:  false,
		},
		{
			name:     "empty provider",
			provider: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ts.SetProvider(tt.provider)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.provider, ts.GetProvider())
			}
		})
	}
}

func TestTenantSecurity_SetRotatedAt(t *testing.T) {
	ts := newTenantSecurity()

	tests := []struct {
		name    string
		date    time.Time
		wantErr bool
	}{
		{
			name:    "valid date",
			date:    time.Now().UTC(),
			wantErr: false,
		},
		{
			name:    "zero date",
			date:    time.Time{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ts.SetRotatedAt(tt.date)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, ts.GetRotatedAt())
				assert.True(t, ts.GetRotatedAt().Equal(tt.date))
			}
		})
	}
}

func TestTenantSecurity_SetVersion(t *testing.T) {
	ts := newTenantSecurity()
	assert.Equal(t, 0, ts.GetVersion())

	ts.SetVersion()
	assert.Equal(t, 1, ts.GetVersion())

	ts.SetVersion()
	assert.Equal(t, 2, ts.GetVersion())
}
