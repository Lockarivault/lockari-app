package tenantmodel

import "fmt"

type VaultSharingConfig struct {
	Enabled   bool     `json:"enabled" bson:"enabled"`
	Approvers []string `json:"approvers" bson:"approvers"`
}

func defaultVaultSharing() VaultSharingConfig {
	return VaultSharingConfig{
		Enabled:   false,
		Approvers: []string{},
	}
}

func (c VaultSharingConfig) normalize() (VaultSharingConfig, error) {
	c.Approvers = normalizeStringSlice(c.Approvers)
	if c.Enabled && len(c.Approvers) == 0 {
		return c, fmt.Errorf("vault sharing enabled requires at least one approver")
	}
	if !c.Enabled {
		c.Approvers = []string{}
	}
	return c, nil
}

func parseVaultSharing(value interface{}) (VaultSharingConfig, error) {
	switch v := value.(type) {
	case nil:
		return defaultVaultSharing(), nil
	case VaultSharingConfig:
		return v.normalize()
	case map[string]interface{}:
		cfg := VaultSharingConfig{}
		if enabled, ok := toBool(v["enabled"]); ok {
			cfg.Enabled = enabled
		}
		if approvers, ok := v["approvers"]; ok {
			cfg.Approvers = interfaceToStringSlice(approvers)
		}
		return cfg.normalize()
	default:
		return defaultVaultSharing(), nil
	}
}
