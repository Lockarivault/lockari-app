package tenantmodel

type StatusType string

const (
	StatusPending   StatusType = "pending"
	StatusActive    StatusType = "active"
	StatusInactive  StatusType = "inactive"
	StatusSuspended StatusType = "suspended"
	StatusFailed    StatusType = "failed"
)

func (s StatusType) IsValid() bool {
	switch s {
	case StatusPending, StatusActive, StatusInactive, StatusSuspended, StatusFailed:
		return true
	default:
		return false
	}
}
