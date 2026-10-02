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
	if args == nil {
		return def
	}
	switch v := args[key].(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	case json.Number:
		if n, err := v.Int64(); err == nil {
			return n
		}
	case string:
		var n int64
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil {
			return n
		}
	}
	return def
}

func intArg(args map[string]any, key string) (int64, bool) {
	if args == nil {
		return 0, false
	}
	v, ok := args[key]
	if !ok {
		return 0, false
	}
	switch n := v.(type) {
	case float64:
		return int64(n), true
	case int64:
		return n, true
	case json.Number:
		value, err := n.Int64()
		return value, err == nil
	case string:
		var value int64
		if _, err := fmt.Sscanf(n, "%d", &value); err == nil {
			return value, true
		}
	}
	return 0, false
}
