package repositoryauditdb

import (
	"time"

	auditmodel "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/audit/model"
)

type AuditLogEntity struct {
	ID        string                 `bson:"_id"`
	Timestamp time.Time              `bson:"timestamp"`
	Actor     ActorInfoEntity        `bson:"actor"`
	Action    string                 `bson:"action"`
	Target    TargetInfoEntity       `bson:"target"`
	Outcome   OutcomeInfoEntity      `bson:"outcome"`
	Context   AuditContextEntity     `bson:"context"`
	Metadata  map[string]interface{} `bson:"metadata,omitempty"`
}

type ActorInfoEntity struct {
	ID        string               `bson:"id"`
	Type      auditmodel.ActorType `bson:"type"`
	IP        string               `bson:"ip"`
	UserAgent string               `bson:"user_agent,omitempty"`
}

type TargetInfoEntity struct {
	ID   string                `bson:"id"`
	Type auditmodel.TargetType `bson:"type"`
	Name string                `bson:"name,omitempty"`
}

type OutcomeInfoEntity struct {
	Status    auditmodel.OutcomeType `bson:"status"`
	Reason    string                 `bson:"reason,omitempty"`
	ErrorCode string                 `bson:"error_code,omitempty"`
}

type AuditContextEntity struct {
	TraceID       string `bson:"trace_id,omitempty"`
	SpanID        string `bson:"span_id,omitempty"`
	TenantID      string `bson:"tenant_id,omitempty"`
	CorrelationID string `bson:"correlation_id,omitempty"`
}
