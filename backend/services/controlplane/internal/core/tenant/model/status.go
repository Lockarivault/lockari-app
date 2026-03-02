package tenantmodel

import (
	"errors"
	"time"
)

type StatusType string

const (
	StatusPending   StatusType = "pending"
	StatusActive    StatusType = "active"
	StatusInactive  StatusType = "inactive"
	StatusSuspended StatusType = "suspended"
	StatusFailed    StatusType = "failed"
)

func NewStatus(status string) (StatusType, error) {
	switch status {
	case "pending":
		return StatusPending, nil
	case "active":
		return StatusActive, nil
	case "inactive":
		return StatusInactive, nil
	case "suspended":
		return StatusSuspended, nil
	case "failed":
		return StatusFailed, nil
	default:
		return "", errors.New("invalid status")
	}
}

type TenantStatus struct {
	Status        StatusType
	FailureReason *string
	UpdatedAt     time.Time
}

func NewTenantStatus(status StatusType) *TenantStatus {
	return &TenantStatus{
		Status:    status,
		UpdatedAt: time.Now().UTC(),
	}
}

func (s StatusType) IsValid() bool {
	switch s {
	case StatusPending, StatusActive, StatusInactive, StatusSuspended, StatusFailed:
		return true
	default:
		return false
	}
}
