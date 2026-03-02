package auditlog

import (
	"context"
	"testing"
)

type mockStore struct {
	saved bool
}

func (m *mockStore) Save(ctx context.Context, entry AuditEntry) error {
	m.saved = true
	return nil
}

func TestAuditService_CreateAuditLog_Validation(t *testing.T) {
	store := &mockStore{}
	service := New(Config{Store: store})

	tests := []struct {
		name    string
		entry   AuditEntry
		wantErr bool
	}{
		{
			name: "valid human entry",
			entry: AuditEntry{
				TenantID:     "tenant-1",
				UserID:       "user-1",
				ActorType:    UserHuman,
				ResourceType: "secret",
				Action:       ActionRead,
			},
			wantErr: false,
		},
		{
			name: "valid app entry",
			entry: AuditEntry{
				TenantID:     "tenant-1",
				UserID:       "app-1",
				ActorType:    UserApp,
				ResourceType: "secret",
				Action:       ActionRead,
			},
			wantErr: false,
		},
		{
			name: "missing actor type",
			entry: AuditEntry{
				TenantID:     "tenant-1",
				UserID:       "user-1",
				ResourceType: "secret",
				Action:       ActionRead,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.CreateAuditLog(context.Background(), tt.entry)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateAuditLog() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
