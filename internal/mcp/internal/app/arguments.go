package app

import (
	"encoding/json"
	"fmt"
)

// Str reads a string argument, "" when absent or not a string. The argument
// readers are shared with the protocol layer, which decodes the same maps.
func Str(args map[string]any, key string) string {
	if args == nil {
		return ""
	}
	if s, ok := args[key].(string); ok {
		return s
	}
	return ""
}

// Num reads an integer argument, def when absent or not an integer.
func Num(args map[string]any, key string, def int64) int64 {
	value, ok := IntArg(args, key)
	if ok {
		return value
	}
	return def
}

// IntArg reads an integer argument and reports whether it was present and valid.
func IntArg(args map[string]any, key string) (int64, bool) {
	value, ok := args[key]
	if !ok {
		return 0, false
	}
	return integerArgument(value)
}

func integerArgument(value any) (int64, bool) {
	switch n := value.(type) {
	case float64:
		return int64(n), true
	case int64:
		return n, true
	case json.Number:
		value, err := n.Int64()
		return value, err == nil
	case string:
		return parseIntegerArgument(n)
	}
	return 0, false
}

func parseIntegerArgument(text string) (int64, bool) {
	var value int64
	if _, err := fmt.Sscanf(text, "%d", &value); err == nil {
		return value, true
	}
	return 0, false
}
