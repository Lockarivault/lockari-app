package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Lockarivault/lockari-app/backend/libs/loggers"
	"go.uber.org/fx"
)

// WebhookService defines the contract for sending notifications to external systems.
type WebhookService interface {
	Send(ctx context.Context, url string, event string, payload interface{}) error
}

type webhookService struct {
	logger loggers.LoggerInterface
	client *http.Client
}

func NewWebhookService(logger loggers.LoggerInterface) WebhookService {
	return &webhookService{
		logger: logger,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type WebhookPayload struct {
	Event     string      `json:"event"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

func (s *webhookService) Send(ctx context.Context, url string, event string, payload interface{}) error {
	if url == "" {
		return nil // No webhook configured
	}

	data := WebhookPayload{
		Event:     event,
		Timestamp: time.Now().UTC(),
		Data:      payload,
	}

	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Lockari-ControlPlane-Webhook/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute webhook request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-success status: %d", resp.StatusCode)
	}

	s.logger.Info("webhook sent successfully", "url", url, "event", event)
	return nil
}

var Module = fx.Options(
	fx.Provide(NewWebhookService),
)
