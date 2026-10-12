package domain

import (
	"maps"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func KindTemplates(kinds []*model.Kind) []any {
	templates := make([]any, 0, len(kinds))
	for _, kind := range kinds {
		templates = append(templates, map[string]any{"kind": kind.Name, "metrics": metricTemplates(kind, kind.Channels), "commands": commandTemplates(kind)})
	}
	return templates
}

func metricTemplates(kind *model.Kind, channels []model.Channel) []any {
	metrics := []any{}
	for i := range channels {
		if template := metricTemplate(kind, &channels[i]); template != nil {
			metrics = append(metrics, template)
		}
	}
	return metrics
}

func metricTemplate(kind *model.Kind, ch *model.Channel) map[string]any {
	kindOfData := dataType(ch)
	if kindOfData == "" {
		return nil
	}
	template := map[string]any{"name": ch.Name, "data_type": kindOfData, "resolution": ch.Resolution, "writable": assigned(kind, ch.Observes)}
	addMetricFacts(template, ch)
	return template
}

func addMetricFacts(template map[string]any, ch *model.Channel) {
	if ch.Unit != "" {
		template["unit"] = ch.Unit
	}
	if ch.Range != nil {
		template["eu_range"] = map[string]any{"min": ch.Range.Min, "max": ch.Range.Max}
	}
	if len(ch.EnumValues) > 0 {
		template["enum_values"] = toAny(ch.EnumValues)
	}
}

var dataTypes = map[string]string{"number": "number", "counter": "integer", "boolean": "boolean"}

func dataType(ch *model.Channel) string {
	if ch.ValueType != "string" {
		return dataTypes[ch.ValueType]
	}
	if len(ch.EnumValues) > 0 {
		return "enum"
	}
	return "string"
}

func assigned(kind *model.Kind, state string) bool {
	for i := range kind.Effectors {
		for _, delta := range kind.Effectors[i].Effect.StateDeltas {
			if state != "" && delta.State == state && delta.AssignFromArg != "" {
				return true
			}
		}
	}
	return false
}

func commandTemplates(kind *model.Kind) []any {
	commands := make([]any, 0, len(kind.Effectors))
	for i := range kind.Effectors {
		eff := &kind.Effectors[i]
		commands = append(commands, map[string]any{"name": eff.Name, "args_schema": withoutDescriptions(eff.ArgsSchema), "risk_class": riskClass(eff.RiskClass), "safe_stop": false})
	}
	return commands
}

func riskClass(declared string) string {
	switch declared {
	case "low", "medium", "high", "critical":
		return declared
	}
	return "high"
}

func withoutDescriptions(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return objectWithoutDescriptions(typed)
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = withoutDescriptions(item)
		}
		return out
	}
	return value
}

func objectWithoutDescriptions(object map[string]any) map[string]any {
	out := make(map[string]any, len(object))
	for _, key := range slices.Sorted(maps.Keys(object)) {
		if key != "description" {
			out[key] = withoutDescriptions(object[key])
		}
	}
	return out
}
