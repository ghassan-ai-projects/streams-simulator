package jsonschema

import (
	"fmt"
	"strconv"
	"strings"
)

// fill populates a Schema from its JSON form (the compile body, without the
// $ref shortcut).
func refName(ref string) (string, error) {
	if !strings.HasPrefix(ref, "#/$defs/") {
		return "", fmt.Errorf("jsonschema: unsupported $ref %q (local #/$defs refs only)", ref)
	}
	return strings.TrimPrefix(ref, "#/$defs/"), nil
}

func toInt(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	case uint64:
		// #nosec G115 -- schema integer keywords are small literals.
		return int(n)
	case jsonNumber:
		i, err := strconv.Atoi(n.String())
		if err != nil {
			return 0
		}
		return i
	default:
		return 0
	}
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case jsonNumber:
		f, err := n.Float64()
		return f, err == nil
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint64:
		return float64(n), true
	}
	return 0, false
}
