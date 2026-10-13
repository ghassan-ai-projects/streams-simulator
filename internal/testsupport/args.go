package testsupport

import "strings"

func ArgsForSchema(schema map[string]any, entity string) map[string]any {
	out := map[string]any{}
	props, _ := schema["properties"].(map[string]any)
	required, _ := schema["required"].([]any)
	for _, r := range required {
		name, _ := r.(string)
		if name == "" {
			continue
		}
		p, _ := props[name].(map[string]any)
		out[name] = argFor(name, p["type"], entity)
	}
	return out
}

func argFor(name string, typ any, entity string) any {
	switch typ {
	case "string":
		if strings.HasSuffix(name, "_id") {
			return entity
		}
		return "x"
	case "number", "integer":
		return 1
	case "boolean":
		return true
	default:
		return ""
	}
}
