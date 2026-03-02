package providers

import (
	"fmt"

	"github.com/lockarivault/lockari-app/backend/libs/webserver"
	"github.com/lockarivault/lockari-app/backend/services/controlplane/config"
)

// ProvideServer provides the webserver instance.
func ProvideServer(cfg config.AppConfig) (webserver.ServerInterface, error) {
	// Example mapping from map[string]interface{} to webserver.Config
	// In a real scenario, we would use mapstructure or similar,
	// but for now, we'll use some default/standard names.

	certFile := "server.crt" // Default or from config
	keyFile := "server.key"  // Default or from config
	port := 8443

	if val, ok := cfg["webserver"].(map[string]interface{}); ok {
		if v, ok := val["cert_file"].(string); ok {
			certFile = v
		}
		if v, ok := val["key_file"].(string); ok {
			keyFile = v
		}
		if v, ok := val["port"].(int); ok {
			port = v
		}
	}

	server, err := webserver.New(webserver.Config{
		Port:        port,
		CertFile:    certFile,
		KeyFile:     keyFile,
		ServiceName: "control-plane-api",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create webserver: %w", err)
	}

	return server, nil
}
