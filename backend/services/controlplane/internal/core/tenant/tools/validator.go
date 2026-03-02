package tenanttools

import (
	"errors"
	"net"
	"strings"
)

func ValidateAllowedOrigins(origins []string) error {
	return nil
}

func DomainIsValid(domain string) error {
	if domain == "" {
		return errors.New("domain is required")
	}

	if len(domain) < 3 || len(domain) > 100 {
		return errors.New("domain must be between 3 and 100 characters")
	}

	if !strings.Contains(domain, ".") {
		return errors.New("domain must contain a dot")
	}

	_, err := net.LookupAddr(domain)
	if err != nil {
		return err
	}

	return nil
}
