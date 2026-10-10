package jsonschema

import (
	"fmt"
	"strconv"
	"strings"
)

// refName accepts local definition references only.
func refName(ref string) (string, error) {
	if !strings.HasPrefix(ref, "#/$defs/") {
		return "", fmt.Errorf("jsonschema: unsupported $ref %q (local #/$defs refs only)", ref)
	}
	return strings.TrimPrefix(ref, "#/$defs/"), nil
}

// toInt reads a schema keyword that must be a non-negative integer. A value
// of any other kind is reported by the second result rather than read as 0.
func toInt(v any) (int, bool) {
	if number, ok := v.(jsonNumber); ok {
		return numberInteger(number)
	}
	return integerValue(v)
}

func integerValue(v any) (int, bool) {
	switch n := v.(type) {
	case float64:
		return int(n), n >= 0 && n == float64(int(n))
	case int:
		return n, n >= 0
	case int64:
		return int(n), n >= 0
	case uint64:
		// #nosec G115 -- schema integer keywords are small literals.
		return int(n), true
	default:
		return 0, false
	}
}

func toFloat(v any) (float64, bool) {
	return asFloat(v)
}

func numberInteger(n jsonNumber) (int, bool) {
	i, err := strconv.Atoi(n.String())
	return i, err == nil && i >= 0
}

// countKeyword reads the non-negative integer keyword key of m, if present.
func countKeyword(m map[string]any, key string) (value int, present bool, err error) {
	raw, ok := m[key]
	if !ok {
		return 0, false, nil
	}
	value, valid := toInt(raw)
	if !valid {
		return 0, true, fmt.Errorf("jsonschema: %s must be a non-negative integer", key)
	}
	return value, true, nil
}
