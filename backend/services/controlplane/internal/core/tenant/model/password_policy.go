package tenantmodel

type PasswordPolicyConfig struct {
	MinLength        int  `json:"min_length" bson:"min_length"`
	RequireNumbers   bool `json:"require_numbers" bson:"require_numbers"`
	RequireUppercase bool `json:"require_uppercase" bson:"require_uppercase"`
	RequireLowercase bool `json:"require_lowercase" bson:"require_lowercase"`
	RequireSpecial   bool `json:"require_special" bson:"require_special"`
}

func defaultPasswordPolicy() PasswordPolicyConfig {
	return PasswordPolicyConfig{
		MinLength:        defaultPasswordMinLength,
		RequireNumbers:   true,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireSpecial:   true,
	}
}

func (c PasswordPolicyConfig) normalize() PasswordPolicyConfig {
	if c.MinLength <= 0 {
		c.MinLength = defaultPasswordMinLength
	}
	if !c.RequireNumbers && !c.RequireUppercase && !c.RequireLowercase && !c.RequireSpecial {
		// Ensure at least lowercase characters are required if nothing is configured.
		c.RequireLowercase = true
	}
	return c
}

func parsePasswordPolicy(value interface{}) (PasswordPolicyConfig, error) {
	switch v := value.(type) {
	case nil:
		return defaultPasswordPolicy(), nil
	case PasswordPolicyConfig:
		return v.normalize(), nil
	case map[string]interface{}:
		cfg := PasswordPolicyConfig{}
		if length, ok := toInt(v["min_length"]); ok {
			cfg.MinLength = length
		}
		if b, ok := toBool(v["require_numbers"]); ok {
			cfg.RequireNumbers = b
		}
		if b, ok := toBool(v["require_uppercase"]); ok {
			cfg.RequireUppercase = b
		}
		if b, ok := toBool(v["require_lowercase"]); ok {
			cfg.RequireLowercase = b
		}
		if b, ok := toBool(v["require_special"]); ok {
			cfg.RequireSpecial = b
		}
		return cfg.normalize(), nil
	default:
		return defaultPasswordPolicy(), nil
	}
}
