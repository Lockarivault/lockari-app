package providers

import (
	"github.com/Lockarivault/lockari-app/backend/libs/loggers"
)

// ProvideLogger provides a logger based on the application configuration.
func ProvideLogger() loggers.LoggerInterface {
	// For now, using default JSON/Info config.
	// In the future, this can pull from config.AppConfig if needed.
	return loggers.New(loggers.Config{
		Handler: loggers.JSON,
		Level:   loggers.LevelInfo,
	})
}
