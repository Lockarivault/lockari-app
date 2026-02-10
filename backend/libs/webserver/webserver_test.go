package webserver

import (
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("Fails without Certificates", func(t *testing.T) {
		_, err := New(Config{})
		if err == nil {
			t.Error("expected error when no certificates are provided")
		}
	})

	t.Run("Defaults to Port 443 with Certificates", func(t *testing.T) {
		s, err := New(Config{
			CertFile: "test.crt",
			KeyFile:  "test.key",
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if s.Config.Port != 443 {
			t.Errorf("expected default port 443, got %d", s.Config.Port)
		}
	})

	t.Run("Custom Port", func(t *testing.T) {
		s, err := New(Config{
			Port:     9090,
			CertFile: "test.crt",
			KeyFile:  "test.key",
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if s.Config.Port != 9090 {
			t.Errorf("expected port 9090, got %d", s.Config.Port)
		}
	})
}
