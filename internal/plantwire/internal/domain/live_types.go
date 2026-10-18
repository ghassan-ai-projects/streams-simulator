package domain

import (
	"math"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonvalue"
)

func plausibleType(value any, template map[string]any) bool {
	switch text(template,"data_type") {
	case "number","integer":
		v,ok := jsonvalue.Number(value)
		return ok && !math.IsNaN(v) && !math.IsInf(v,0) && (text(template,"data_type") != "integer" || math.Trunc(v) == v)
	case "boolean":
		_,ok := value.(bool)
		return ok
	case "string":
		_,ok := value.(string)
		return ok
	default:
		return enumValue(value,template["enum_values"])
	}
}

func enumValue(value, choices any) bool {
	list,ok := choices.([]any)
	return ok && slices.Contains(list,value)
}
