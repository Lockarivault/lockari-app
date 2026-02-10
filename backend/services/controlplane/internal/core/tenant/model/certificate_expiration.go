package tenantmodel

import "strings"

type CertificateDefaultExpirationConfig struct {
	Value int    `json:"value" bson:"value"`
	Unit  string `json:"unit" bson:"unit"`
}

func defaultCertificateExpiration() CertificateDefaultExpirationConfig {
	return CertificateDefaultExpirationConfig{
		Value: defaultCertificateExpirationValue,
		Unit:  defaultCertificateExpirationUnit,
	}
}

func (c CertificateDefaultExpirationConfig) normalize() CertificateDefaultExpirationConfig {
	if c.Value <= 0 {
		c.Value = defaultCertificateExpirationValue
	}
	unit := strings.ToLower(strings.TrimSpace(c.Unit))
	switch unit {
	case supportedCertificateExpirationUnitDays,
		supportedCertificateExpirationUnitWeeks,
		supportedCertificateExpirationUnitMonths,
		supportedCertificateExpirationUnitYears:
		c.Unit = unit
	default:
		c.Unit = defaultCertificateExpirationUnit
	}
	return c
}

func parseCertificateExpiration(value interface{}) (CertificateDefaultExpirationConfig, error) {
	switch v := value.(type) {
	case nil:
		return defaultCertificateExpiration(), nil
	case CertificateDefaultExpirationConfig:
		return v.normalize(), nil
	case map[string]interface{}:
		cfg := CertificateDefaultExpirationConfig{}
		if val, ok := toInt(v["value"]); ok {
			cfg.Value = val
		}
		if unit, ok := v["unit"].(string); ok {
			cfg.Unit = unit
		}
		return cfg.normalize(), nil
	default:
		return defaultCertificateExpiration(), nil
	}
}
