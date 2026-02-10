package auditmodel

import (
	"time"

	"github.com/Lockarivault/lockari-app/backend/libs/uuid"
)

// ActorType identifies the source of the action.
type ActorType string

const (
	ActorUser   ActorType = "USER"
	ActorSystem ActorType = "SYSTEM"
	ActorAgent  ActorType = "AGENT"
)

// TargetType identifies the type of resource being audited.
type TargetType string

const (
	TargetTenant TargetType = "TENANT"
	TargetSecret TargetType = "SECRET"
	TargetKey    TargetType = "KEY"
	TargetPolicy TargetType = "POLICY"
)

// OutcomeType defines the result of the action.
type OutcomeType string

const (
	OutcomeSuccess OutcomeType = "SUCCESS"
	OutcomeFailure OutcomeType = "FAILURE"
	OutcomeDenied  OutcomeType = "DENIED"
)

// AuditLog represents a single auditable event in the system (AATO Pattern).
type AuditLog struct {
	ID        uuid.UUID `json:"id"`
	Timestamp time.Time `json:"timestamp"`

	// AATO Components
	Actor   ActorInfo   `json:"actor"`   // Who
	Action  string      `json:"action"`  // What (e.g., TENANT_CREATE)
	Target  TargetInfo  `json:"target"`  // Where
	Outcome OutcomeInfo `json:"outcome"` // Result

	// Contextual Metadata
	Context  AuditContext           `json:"context"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ActorInfo describes the entity performing the action.
type ActorInfo struct {
	ID        string    `json:"id"`
	Type      ActorType `json:"type"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent,omitempty"`
}

// TargetInfo describes the resource that was acted upon.
type TargetInfo struct {
	ID   string     `json:"id"`
	Type TargetType `json:"type"`
	Name string     `json:"name,omitempty"`
}

// OutcomeInfo describes the result and reason for the outcome.
type OutcomeInfo struct {
	Status    OutcomeType `json:"status"`
	Reason    string      `json:"reason,omitempty"`
	ErrorCode string      `json:"error_code,omitempty"`
}

// AuditContext provides tracing and relationship information.
type AuditContext struct {
	TraceID       string    `json:"trace_id,omitempty"`
	SpanID        string    `json:"span_id,omitempty"`
	TenantID      uuid.UUID `json:"tenant_id,omitempty"` // ID of the tenant where the action originated
	CorrelationID string    `json:"correlation_id,omitempty"`
}

// NewAuditLog creates a new audit log entry with default timestamp and unique ID.
func NewAuditLog(actor ActorInfo, action string, target TargetInfo) AuditLog {
	return AuditLog{
		ID:        uuid.New(),
		Timestamp: time.Now().UTC(),
		Actor:     actor,
		Action:    action,
		Target:    target,
		Metadata:  make(map[string]interface{}),
	}
}
