package tenanttools

import "errors"

var (
	ErrNilRepository = errors.New("repository cannot be nil")
	ErrNilTelemetry  = errors.New("telemetry cannot be nil")
)

type ToolsTenant interface {
}

type tenant struct {
}

func InnicializeToolsTenant() (ToolsTenant, error) {
	m := tenant{}
	return m, nil
}
