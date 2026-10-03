package mcp

import (
	"encoding/json"
	"fmt"
)

func str(args map[string]any, key string) string {
	if args == nil {
		return ""
	}
	if s, ok := args[key].(string); ok {
		return s
	}
	return ""
}

func num(args map[string]any, key string, def int64) int64 {
	value, ok := intArg(args, key)
	if ok {
		return value
	}
	return def
}

func intArg(args map[string]any, key string) (int64, bool) {
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
