package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/likexian/whois"
	whoisparser "github.com/likexian/whois-parser"
)

var (
	// domainRegex follows RFC 1035 for domain name validation.
	// It allows for subdomains and ensures each label is 1-63 chars and the TLD is 2+ chars.
	domainRegex = regexp.MustCompile(`^(?i)[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?(\.[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?)*\.[a-z]{2,}$`)
)

type Origins string

func main() {
	address := []string{"vault.lockari.com", "www.myaddress.com.dsdfdsfdsfdffdsfsdfsdf"}
	for _, addr := range address {
		Origins(addr).IsValid()

	}
}

func (o Origins) IsValid() error {
	// Criamos um contexto com timeout de 2 segundos
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	v := strings.ToLower(strings.TrimSpace(string(o)))
	if v == "" {
		return errors.New("origin is required")
	}

	cleaned := Origins(v)

	// Try IP first
	if err := cleaned.isIP(); err == nil {
		return nil
	}

	// Try CIDR
	if err := cleaned.isCIDR(); err == nil {
		return nil
	}

	// Try Domain
	if err := cleaned.isDomain(ctx); err != nil {
		return err
	}

	return nil
}

func (o Origins) isIP() error {
	if net.ParseIP(string(o)) != nil {
		return nil
	}
	return errors.New("origin is not an IP")
}

func (o Origins) isCIDR() error {
	if _, _, err := net.ParseCIDR(string(o)); err != nil {
		return err
	}
	return nil
}

func (o Origins) isDomain(ctx context.Context) error {
	resolver := net.DefaultResolver
	_, err := resolver.LookupHost(ctx, string(o))
	if err != nil {
		return fmt.Errorf("Error querying WHOIS for %s: %v \n", string(o), err)
	}
	if domainRegex.MatchString(string(o)) {
		// 1. Query the raw WHOIS information
		domain := string(o)
		rawWhois, err := whois.Whois(domain)
		if err != nil {
			return fmt.Errorf("Error querying WHOIS for %s: %v \n", domain, err)
		}

		// 2. Parse the raw information into a structured format
		result, err := whoisparser.Parse(rawWhois)
		if err != nil {
			return fmt.Errorf("Error parsing WHOIS for %s: %v \n", domain, err)
		}
		if result.Domain == nil {
			return errors.New("domain not found")
		}

		return nil
	}
	return errors.New("origin is not a domain")
}
