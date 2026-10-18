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
		commands = append(commands, map[string]any{"name": eff.Name, "args_schema": validationOnly(eff.ArgsSchema),
			"risk_class": riskClass(eff.RiskClass), "safe_stop": eff.SafeStop})
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

var (
	valueKeywords = []string{"type", "minimum", "maximum", "exclusiveMinimum", "exclusiveMaximum", "multipleOf", "minLength",
		"maxLength", "pattern", "minItems", "maxItems", "uniqueItems", "minProperties", "maxProperties", "required", "enum", "const"}
	schemaKeywords = []string{"items", "additionalProperties", "not"}
	listKeywords   = []string{"anyOf", "oneOf", "allOf", "prefixItems"}
)

func validationOnly(value any) any {
	schema, isObject := value.(map[string]any)
	if !isObject {
		return value
	}
	out := map[string]any{}
	for _, key := range slices.Sorted(maps.Keys(schema)) {
		if kept, ok := keyword(key, schema[key]); ok {
			out[key] = kept
		}
	}
	return out
}

func keyword(key string, value any) (any, bool) {
	switch {
	case slices.Contains(valueKeywords, key):
		return value, true
	case slices.Contains(schemaKeywords, key):
		return validationOnly(value), true
	case slices.Contains(listKeywords, key):
		return each(value, validationOnly), true
	case key == "properties":
		return namedSchemas(value), true
	}
	return nil, false
}

func each(value any, apply func(any) any) any {
	list, isList := value.([]any)
	if !isList {
		return value
	}
	out := make([]any, len(list))
	for i, item := range list {
		out[i] = apply(item)
	}
	return out
}

func namedSchemas(value any) any {
	named, isObject := value.(map[string]any)
	if !isObject {
		return value
	}
	out := make(map[string]any, len(named))
	for _, name := range slices.Sorted(maps.Keys(named)) {
		out[name] = validationOnly(named[name])
	}
	return out
}
