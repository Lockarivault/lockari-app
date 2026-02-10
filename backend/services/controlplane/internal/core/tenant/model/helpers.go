package tenantmodel

import (
	"fmt"
	"sort"
	"strings"
)

func normalizeStringSlice(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	// Sort deterministically: prefer base domains (fewer labels) before subdomains,
	// then lexicographically. This keeps e.g. example.com before app.example.com.
	sort.Slice(result, func(i, j int) bool {
		a := result[i]
		b := result[j]
		ai := strings.Count(a, ".")
		bi := strings.Count(b, ".")
		if ai != bi {
			return ai < bi
		}
		return a < b
	})
	return result
}

func toInt(value interface{}) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	case float32:
		return int(v), true
	case string:
		if strings.TrimSpace(v) == "" {
			return 0, false
		}
		var parsed int
		if _, err := fmt.Sscanf(v, "%d", &parsed); err == nil {
			return parsed, true
		}
	}
	return 0, false
}

func toBool(value interface{}) (bool, bool) {
	switch v := value.(type) {
	case bool:
		return v, true
	case string:
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "true", "1", "yes", "y":
			return true, true
		case "false", "0", "no", "n":
			return false, true
		}
	}
	return false, false
}

func interfaceToStringSlice(value interface{}) []string {
	switch v := value.(type) {
	case []string:
		return normalizeStringSlice(v)
	case []interface{}:
		result := make([]string, 0, len(v))
		for _, item := range v {
			result = append(result, fmt.Sprintf("%v", item))
		}
		return normalizeStringSlice(result)
	case string:
		if strings.TrimSpace(v) == "" {
			return []string{}
		}
		return normalizeStringSlice(strings.Split(v, ","))
	default:
		return []string{}
	}
}

func cloneValue(value interface{}) interface{} {
	switch v := value.(type) {
	case []string:
		cloned := make([]string, len(v))
		copy(cloned, v)
		return cloned
	case []interface{}:
		cloned := make([]interface{}, len(v))
		for i, item := range v {
			cloned[i] = cloneValue(item)
		}
		return cloned
	case map[string]interface{}:
		cloned := make(map[string]interface{}, len(v))
		for key, item := range v {
			cloned[key] = cloneValue(item)
		}
		return cloned
	case ProprietiesTypes:
		return v.Clone()
	default:
		return v
	}
}
