package jsonschema

import (
	"regexp"
	"strings"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
)

func joinPath(base, name string) string {
	if base == "" {
		return "/" + escapePointer(name)
	}
	return base + "/" + escapePointer(name)
}

func escapePointer(s string) string {
	s = strings.ReplaceAll(s, "~", "~0")
	return strings.ReplaceAll(s, "/", "~1")
}

func typeMatches(types []string, v any) bool {
	for _, t := range types {
		if matchesType(t, v) {
			return true
		}
	}
	return false
}

func matchesType(t string, v any) bool {
	switch t {
	case "null", "boolean", "string":
		return matchesPrimitiveType(t, v)
	case "object", "array":
		return matchesCollectionType(t, v)
	case "number":
		_, ok := asFloat(v)
		return ok
	case "integer":
		return matchesInteger(v)
	}
	return false
}

func matchesPrimitiveType(t string, v any) bool {
	switch t {
	case "null":
		return v == nil
	case "boolean":
		_, ok := v.(bool)
		return ok
	case "string":
		_, ok := v.(string)
		return ok
	}
	return false
}

func matchesCollectionType(t string, v any) bool {
	switch t {
	case "object":
		_, ok := v.(map[string]any)
		return ok
	case "array":
		_, ok := v.([]any)
		return ok
	}
	return false
}

func matchesInteger(v any) bool {
	number, ok := asFloat(v)
	return ok && number == float64(int64(number))
}

type jsonNumber interface {
	Float64() (float64, error)
	String() string
}

func asFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint64:
		return float64(n), true
	default:
		return jsonFloat(v)
	}
}

func jsonFloat(v any) (float64, bool) {
	if number, ok := v.(jsonNumber); ok {
		value, err := number.Float64()
		return value, err == nil
	}
	return 0, false
}

func typeName(v any) string {
	switch v.(type) {
	case nil:
		return "null"
	case bool:
		return "boolean"
	case map[string]any, []any:
		return collectionTypeName(v)
	case string:
		return "string"
	default:
		return numericTypeName(v)
	}
}

func collectionTypeName(v any) string {
	if _, ok := v.(map[string]any); ok {
		return "object"
	}
	return "array"
}

func numericTypeName(v any) string {
	if _, ok := asFloat(v); ok {
		return "number"
	}
	return "unknown"
}

func enumContains(enum []any, v any) bool {
	for _, e := range enum {
		if jsonEqual(e, v) {
			return true
		}
	}
	return false
}

func jsonEqual(a, b any) bool {
	sa, err := canonical.MarshalString(a)
	if err != nil {
		return false
	}
	sb, err := canonical.MarshalString(b)
	if err != nil {
		return false
	}
	return sa == sb
}

func checkFormat(format, s string) string {
	switch format {
	case "date-time":
		return checkDateTime(s)
	case "regex":
		return checkRegex(s)
	case "uri":
		return checkURI(s)
	case "hostname":
		return checkHostname(s)
	}
	return ""
}

func checkDateTime(s string) string {
	if _, err := time.Parse(time.RFC3339Nano, s); err != nil {
		return "string is not a valid RFC 3339 date-time"
	}
	return ""
}

func checkRegex(s string) string {
	if _, err := regexp.Compile(s); err != nil {
		return "string is not a valid regular expression"
	}
	return ""
}

func checkURI(s string) string {
	if !strings.Contains(s, ":") {
		return "string is not a valid URI"
	}
	return ""
}

func checkHostname(s string) string {
	if s == "" || strings.ContainsAny(s, " /") {
		return "string is not a valid hostname"
	}
	return ""
}
