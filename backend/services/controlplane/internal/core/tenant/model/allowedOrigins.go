package tenantmodel

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/likexian/whois"
	whoisparser "github.com/likexian/whois-parser"
)

type Origins string

func (o Origins) IsValid() error {

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
