package domain

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

var annotationKeywords = map[string]bool{"description": true, "title": true, "$comment": true, "examples": true}

func OperatorEffector(effector *model.Effector) EffectorInfo {
	risk := effector.RiskClass
	if risk == "" {
		risk = "medium"
	}
	schema, _ := withoutAnnotations(effector.ArgsSchema).(map[string]any)
	return EffectorInfo{Name: effector.Name, ArgsSchema: schema, RiskClass: risk}
}

func withoutAnnotations(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return objectWithoutAnnotations(typed)
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = withoutAnnotations(item)
		}
		return out
	}
	return value
}

func objectWithoutAnnotations(object map[string]any) map[string]any {
	out := make(map[string]any, len(object))
	// determinism-safe: copies a map into a map; no order reaches output.
	for key, item := range object {
		if !annotationKeywords[key] {
			out[key] = withoutAnnotations(item)
		}
	}
	return out
}
