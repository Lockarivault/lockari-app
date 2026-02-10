package providers

import (
	"fmt"

	"github.com/Lockarivault/lockari-app/backend/services/controlplane/config"
)

func ProvideConfig() (*config.Connections, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return cfg, nil
}

func ProvideAppSettings(cfg *config.Connections) config.AppConfig {
	return cfg.Settings
}
