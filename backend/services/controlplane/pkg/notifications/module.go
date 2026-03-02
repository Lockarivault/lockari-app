package notifications

import (
	"context"
)

// NotificationService defines the contract for sending user-facing notifications.
type NotificationService interface {
	SendEmail(ctx context.Context, to string, subject string, body string) error
}
