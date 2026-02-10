package repositoryaudit

import (
	"context"

	auditmodel "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/audit/model"
)

// RepositoryAudit defines the storage contract for audit logs.
// This interface is shared but implementations may differ between SaaS (MongoDB)
// and Agent (Local Storage/Relay).
type RepositoryAudit interface {
	// Save persists an audit log entry. This is an append-only operation.
	Save(ctx context.Context, log auditmodel.AuditLog) error

	// List retrieves audit logs based on filters (primarily for SaaS).
	List(ctx context.Context, filter map[string]interface{}) ([]auditmodel.AuditLog, error)

	// GetByID retrieves a specific log for deep inspection.
	GetByID(ctx context.Context, id string) (auditmodel.AuditLog, error)
}
