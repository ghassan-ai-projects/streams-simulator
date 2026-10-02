package device

import (
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
)

func strictParams(raw any) (map[string]any, bool) {
	out := map[string]any{}
	m, ok := raw.(map[string]any)
	if !ok {
		return nil, false
	}
	for name, v := range m {
		if name == "" {
			return nil, false
		}
		switch value := v.(type) {
		case float64:
			if !finite(value) {
				return nil, false
			}
		case string:
			if value == "" {
				return nil, false
			}
		default:
			return nil, false
		}
		out[name] = v
	}
	return out, true
}

func numericParams(raw any) (map[string]float64, bool) {
	params, ok := strictParams(raw)
	if !ok {
		return nil, false
	}
	out := map[string]float64{}
	for name, value := range params {
		number, ok := value.(float64)
		if ok {
			out[name] = number
		}
	}
	return out, true
}

func semanticCommandDigest(command map[string]any) (string, error) {
	identity := make(map[string]any, len(command))
	for key, value := range command {
		if key != "command_id" {
			identity[key] = value
		}
	}
	digest, err := canonical.DigestDomain("situation-runtime/device-command/v1\n", identity)
	if err != nil {
		return "", fmt.Errorf("canonicalize semantic command identity: %w", err)
	}
	return digest, nil
}

func orDefault(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

func repeat(b byte, n int) string {
	out := make([]byte, n)
	for i := range out {
		out[i] = b
	}
	return string(out)
}
