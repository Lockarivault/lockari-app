package tenantmodel

import (
	"fmt"
	"strings"
)

type ProprietiesTypes struct {
	FullyQualifiedDomain string
	AllowedOrigins       []string
	Keys                 map[string]interface{}
}

func NewProprieties(keys map[string]interface{}) ProprietiesTypes {
	if keys == nil {
		return ProprietiesTypes{Keys: make(map[string]interface{})}
	}
	cloned := make(map[string]interface{}, len(keys))
	for k, v := range keys {
		key := strings.TrimSpace(k)
		if key == "" {
			continue
		}
		cloned[key] = cloneValue(v)
	}
	return ProprietiesTypes{Keys: cloned}
}

func (p *ProprietiesTypes) EnsureDefaults() error {
	if p.Keys == nil {
		p.Keys = make(map[string]interface{})
	}

	policy, err := parsePasswordPolicy(p.Keys[proprietyPasswordPolicyKey])
	if err != nil {
		return err
	}
	p.Keys[proprietyPasswordPolicyKey] = policy

	notification, err := parseCertificateNotification(p.Keys[proprietyCertificateNotificationKey])
	if err != nil {
		return err
	}
	p.Keys[proprietyCertificateNotificationKey] = notification

	ips, err := parseAllowedClientIPs(p.Keys[proprietyAllowedClientIPsKey])
	if err != nil {
		return err
	}
	p.Keys[proprietyAllowedClientIPsKey] = ips

	expiration, err := parseCertificateExpiration(p.Keys[proprietyCertificateDefaultExpiration])
	if err != nil {
		return err
	}
	p.Keys[proprietyCertificateDefaultExpiration] = expiration

	sharing, err := parseVaultSharing(p.Keys[proprietyVaultSharingKey])
	if err != nil {
		return err
	}
	p.Keys[proprietyVaultSharingKey] = sharing

	return nil
}

func (p ProprietiesTypes) Validate() error {
	if _, err := parsePasswordPolicy(p.Keys[proprietyPasswordPolicyKey]); err != nil {
		return err
	}
	if _, err := parseCertificateNotification(p.Keys[proprietyCertificateNotificationKey]); err != nil {
		return err
	}
	if _, err := parseAllowedClientIPs(p.Keys[proprietyAllowedClientIPsKey]); err != nil {
		return err
	}
	if _, err := parseCertificateExpiration(p.Keys[proprietyCertificateDefaultExpiration]); err != nil {
		return err
	}
	if _, err := parseVaultSharing(p.Keys[proprietyVaultSharingKey]); err != nil {
		return err
	}
	return nil
}

func (p ProprietiesTypes) PasswordPolicy() PasswordPolicyConfig {
	policy, err := parsePasswordPolicy(p.Keys[proprietyPasswordPolicyKey])
	if err != nil {
		return defaultPasswordPolicy()
	}
	return policy
}

func (p ProprietiesTypes) CertificateNotification() CertificateNotificationConfig {
	notification, err := parseCertificateNotification(p.Keys[proprietyCertificateNotificationKey])
	if err != nil {
		return defaultCertificateNotification()
	}
	return notification
}

func (p ProprietiesTypes) AllowedClientIPs() []string {
	ips, err := parseAllowedClientIPs(p.Keys[proprietyAllowedClientIPsKey])
	if err != nil {
		return []string{}
	}
	return ips
}

func (p ProprietiesTypes) CertificateDefaultExpiration() CertificateDefaultExpirationConfig {
	expiration, err := parseCertificateExpiration(p.Keys[proprietyCertificateDefaultExpiration])
	if err != nil {
		return defaultCertificateExpiration()
	}
	return expiration
}

func (p ProprietiesTypes) VaultSharing() VaultSharingConfig {
	sharing, err := parseVaultSharing(p.Keys[proprietyVaultSharingKey])
	if err != nil {
		return defaultVaultSharing()
	}
	return sharing
}

// FullyQualifiedDomain returns the fully qualified domain stored in the proprieties.
func (p ProprietiesTypes) FullyQualifiedDomainValue() string {
	if p.Keys == nil {
		return ""
	}
	switch v := p.Keys[proprietyFullyQualifiedDomainKey].(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(strings.ToLower(v))
	default:
		return strings.TrimSpace(strings.ToLower(fmt.Sprintf("%v", v)))
	}
}

func (p ProprietiesTypes) MaxSecrets() int {
	if v, ok := p.Keys[proprietyMaxSecretsKey].(int); ok {
		return v
	}
	if v, ok := p.Keys[proprietyMaxSecretsKey].(float64); ok { // JSON unmarshals to float64
		return int(v)
	}
	return 0
}

func (p ProprietiesTypes) MaxUsers() int {
	if v, ok := p.Keys[proprietyMaxUsersKey].(int); ok {
		return v
	}
	if v, ok := p.Keys[proprietyMaxUsersKey].(float64); ok {
		return int(v)
	}
	return 0
}

func (p ProprietiesTypes) MaxStorageBytes() int64 {
	if v, ok := p.Keys[proprietyMaxStorageBytesKey].(int64); ok {
		return v
	}
	if v, ok := p.Keys[proprietyMaxStorageBytesKey].(float64); ok {
		return int64(v)
	}
	return 0
}

func (p *ProprietiesTypes) SetMaxSecrets(v int) {
	if p.Keys == nil {
		p.Keys = make(map[string]interface{})
	}
	p.Keys[proprietyMaxSecretsKey] = v
}

func (p *ProprietiesTypes) SetMaxUsers(v int) {
	if p.Keys == nil {
		p.Keys = make(map[string]interface{})
	}
	p.Keys[proprietyMaxUsersKey] = v
}

func (p *ProprietiesTypes) SetMaxStorageBytes(v int64) {
	if p.Keys == nil {
		p.Keys = make(map[string]interface{})
	}
	p.Keys[proprietyMaxStorageBytesKey] = v
}

// AllowedOrigins returns allowed origins from proprieties.
func (p ProprietiesTypes) AllowedOriginsValue() []string {
	if p.Keys == nil {
		return []string{}
	}
	val := p.Keys[proprietyAllowedOriginsKey]
	switch v := val.(type) {
	case nil:
		return []string{}
	case []string:
		return normalizeStringSlice(v)
	case []interface{}:
		res := make([]string, 0, len(v))
		for _, it := range v {
			res = append(res, strings.TrimSpace(fmt.Sprintf("%v", it)))
		}
		return normalizeStringSlice(res)
	case string:
		if strings.TrimSpace(v) == "" {
			return []string{}
		}
		parts := strings.Split(v, ",")
		return normalizeStringSlice(parts)
	default:
		return []string{}
	}
}

// SetFullyQualifiedDomain stores the domain into proprieties map.
func (p *ProprietiesTypes) SetFullyQualifiedDomain(domain string) {
	if p.Keys == nil {
		p.Keys = make(map[string]interface{})
	}
	p.Keys[proprietyFullyQualifiedDomainKey] = strings.TrimSpace(strings.ToLower(domain))
}

// SetAllowedOrigins stores allowed origins into proprieties map.
func (p *ProprietiesTypes) SetAllowedOrigins(origins []string) {
	if p.Keys == nil {
		p.Keys = make(map[string]interface{})
	}
	p.Keys[proprietyAllowedOriginsKey] = normalizeStringSlice(origins)
}

func (p ProprietiesTypes) Clone() ProprietiesTypes {
	return NewProprieties(p.Keys)
}

func (p ProprietiesTypes) AsMap() map[string]interface{} {
	if p.Keys == nil {
		return map[string]interface{}{}
	}
	cloned := make(map[string]interface{}, len(p.Keys))
	for k, v := range p.Keys {
		cloned[k] = cloneValue(v)
	}
	return cloned
}

func (p ProprietiesTypes) GetItems() map[string]interface{} {
	return p.Keys
}
