package repositoryauditdb

import (
	"github.com/Lockarivault/lockari-app/backend/libs/uuid"
	auditmodel "github.com/Lockarivault/lockari-app/backend/services/controlplane/internal/core/audit/model"
)

func toEntity(m auditmodel.AuditLog) AuditLogEntity {
	return AuditLogEntity{
		ID:        m.ID.String(),
		Timestamp: m.Timestamp,
		Actor: ActorInfoEntity{
			ID:        m.Actor.ID,
			Type:      m.Actor.Type,
			IP:        m.Actor.IP,
			UserAgent: m.Actor.UserAgent,
		},
		Action: m.Action,
		Target: TargetInfoEntity{
			ID:   m.Target.ID,
			Type: m.Target.Type,
			Name: m.Target.Name,
		},
		Outcome: OutcomeInfoEntity{
			Status:    m.Outcome.Status,
			Reason:    m.Outcome.Reason,
			ErrorCode: m.Outcome.ErrorCode,
		},
		Context: AuditContextEntity{
			TraceID:       m.Context.TraceID,
			SpanID:        m.Context.SpanID,
			TenantID:      m.Context.TenantID.String(),
			CorrelationID: m.Context.CorrelationID,
		},
		Metadata: m.Metadata,
	}
}

func fromEntity(e AuditLogEntity) auditmodel.AuditLog {
	id, _ := uuid.Parse(e.ID)
	tenantID, _ := uuid.Parse(e.Context.TenantID)

	return auditmodel.AuditLog{
		ID:        id,
		Timestamp: e.Timestamp,
		Actor: auditmodel.ActorInfo{
			ID:        e.Actor.ID,
			Type:      e.Actor.Type,
			IP:        e.Actor.IP,
			UserAgent: e.Actor.UserAgent,
		},
		Action: e.Action,
		Target: auditmodel.TargetInfo{
			ID:   e.Target.ID,
			Type: e.Target.Type,
			Name: e.Target.Name,
		},
		Outcome: auditmodel.OutcomeInfo{
			Status:    e.Outcome.Status,
			Reason:    e.Outcome.Reason,
			ErrorCode: e.Outcome.ErrorCode,
		},
		Context: auditmodel.AuditContext{
			TraceID:       e.Context.TraceID,
			SpanID:        e.Context.SpanID,
			TenantID:      tenantID,
			CorrelationID: e.Context.CorrelationID,
		},
		Metadata: e.Metadata,
	}
}
