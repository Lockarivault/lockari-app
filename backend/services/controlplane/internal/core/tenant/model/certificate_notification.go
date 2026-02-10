package tenantmodel

type CertificateNotificationConfig struct {
	DaysBefore int `json:"days_before" bson:"days_before"`
}

func defaultCertificateNotification() CertificateNotificationConfig {
	return CertificateNotificationConfig{DaysBefore: defaultCertificateNotificationDays}
}

func (c CertificateNotificationConfig) normalize() CertificateNotificationConfig {
	if c.DaysBefore <= 0 {
		c.DaysBefore = defaultCertificateNotificationDays
	}
	return c
}

func parseCertificateNotification(value interface{}) (CertificateNotificationConfig, error) {
	switch v := value.(type) {
	case nil:
		return defaultCertificateNotification(), nil
	case CertificateNotificationConfig:
		return v.normalize(), nil
	case map[string]interface{}:
		cfg := CertificateNotificationConfig{}
		if days, ok := toInt(v["days_before"]); ok {
			cfg.DaysBefore = days
		}
		return cfg.normalize(), nil
	default:
		return defaultCertificateNotification(), nil
	}
}
