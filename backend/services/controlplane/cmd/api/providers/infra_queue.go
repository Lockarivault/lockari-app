package providers

import (
	"errors"

	"github.com/lockarivault/lockari-app/backend/libs/mensageria"
	"github.com/lockarivault/lockari-app/backend/services/controlplane/internal/infrastructure/messaging"
)

// ProvideInfraMessaging provides a publisher based on the application configuration.
func ProvideInfraMessaging(queue mensageria.MessageQueue) (messaging.Publisher, error) {
	if queue == nil {
		return nil, errors.New("queue is required")
	}
	return messaging.NewMensageriaAdapter(queue), nil
}
