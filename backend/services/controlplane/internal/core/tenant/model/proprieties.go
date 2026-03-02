package tenantmodel

import (
	"errors"
	"fmt"
	"strings"
)

var defaultFullyQualifiedDomain = "vault.lockari.com"

type ProprietiesKeys map[string]any

type ProprietiesTypes struct {
	FullyQualifiedDomain         string
	PlatformFullyQualifiedDomain string
	AllowedOrigins               []Origins
	ProprietiesKeys
}

func NewProprietiesTypes(keys map[string]any) ProprietiesTypes {

	cloned := keys
	if keys == nil {
		cloned = make(map[string]any)
	}

	tempKeys := make(map[string]any, len(cloned))
	for k, v := range cloned {
		t := strings.ToLower(strings.TrimSpace(k))
		if t == "" {
			continue
		}
		tempKeys[t] = v
	}

	return ProprietiesTypes{
		ProprietiesKeys: tempKeys,
	}
}

func (p *ProprietiesTypes) SetFullyQualifiedDomain(value string) error {
	return nil
}

func (p *ProprietiesTypes) SetAllowedOrigins(origins []Origins) error {
	if p.ProprietiesKeys == nil {
		p.ProprietiesKeys = make(map[string]interface{})
	}
	if origins == nil {
		origins = []Origins{}
	}

	for _, origin := range origins {
		if err := origin.IsValid(); err != nil {
			return err
		}
	}

	p.ProprietiesKeys[proprietyAllowedOriginsKey] = origins

	return nil
}

func (p *ProprietiesTypes) SetPlatformFullyQualifiedDomain(value string) error {
	if value == "" {
		return errors.New("platform fully qualified domain is required")
	}

	if len(value) < 3 || len(value) > 100 {
		return errors.New("platform fully qualified domain must be between 3 and 100 characters")
	}

	v := strings.ReplaceAll(strings.TrimSpace(value), " ", "")
	if v == "" {
		return errors.New("platform fully qualified domain is required")
	}

	if !strings.Contains(v, ".") {
		return errors.New("platform fully qualified domain must contain a dot")
	}

	p.PlatformFullyQualifiedDomain = fmt.Sprintf("%s.%s", v, defaultFullyQualifiedDomain)

	return nil
}
