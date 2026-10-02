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

func toInt(v any) int {
	if number, ok := v.(jsonNumber); ok {
		return numberInteger(number)
	}
	return integerValue(v)
}

func integerValue(v any) int {
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
	default:
		return 0
	}
}

func toFloat(v any) (float64, bool) {
	return asFloat(v)
}

func numberInteger(n jsonNumber) int {
	i, err := strconv.Atoi(n.String())
	if err != nil {
		return 0
	}
	return i
}
