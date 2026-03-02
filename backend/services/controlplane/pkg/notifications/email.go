package notifications

import (
	"context"

	"github.com/lockarivault/lockari-app/backend/libs/loggers"
	"go.uber.org/fx"
)

type emailNotificationService struct {
	logger loggers.LoggerInterface
}

func NewEmailNotificationService(logger loggers.LoggerInterface) NotificationService {
	return &emailNotificationService{
		logger: logger,
	}
}

func (s *emailNotificationService) SendEmail(ctx context.Context, to string, subject string, body string) error {
	// Simulation: Log the email instead of sending it
	s.logger.Info("sending email (simulated)",
		"to", to,
		"subject", subject,
		"body_preview", body[:min(len(body), 50)]+"...",
	)
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

var Module = fx.Options(
	fx.Provide(NewEmailNotificationService),
)