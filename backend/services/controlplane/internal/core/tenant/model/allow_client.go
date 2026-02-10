package tenantmodel

import (
	"fmt"
	"strings"
)

func parseAllowedClientIPs(value interface{}) ([]string, error) {
	switch v := value.(type) {
	case nil:
		return []string{}, nil
	case []string:
		return normalizeStringSlice(v), nil
	case []interface{}:
		result := make([]string, 0, len(v))
		for _, item := range v {
			result = append(result, strings.TrimSpace(fmt.Sprintf("%v", item)))
		}
		return normalizeStringSlice(result), nil
	case string:
		if strings.TrimSpace(v) == "" {
			return []string{}, nil
		}
		parts := strings.Split(v, ",")
		return normalizeStringSlice(parts), nil
	default:
		return []string{}, nil
	}
}
