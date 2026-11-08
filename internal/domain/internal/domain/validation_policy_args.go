package domain

import (
	"maps"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (checks domainChecks) policyArguments(e *model.Effector, args map[string]any) error {
	for _, name := range slices.Sorted(maps.Keys(args)) {
		if err := checks.policyArgument(e, name, args[name]); err != nil {
			return err
		}
	}
	for _, name := range requiredArguments(e) {
		if _, given := args[name]; !given && !slices.Contains(e.EntityArguments(), name) {
			return checks.bad("/args", "the call omits argument %q, which effector %q requires", name, e.Name)
		}
	}
	return nil
}

func (checks domainChecks) policyArgument(e *model.Effector, name string, value any) error {
	at := "/args" + memberPointer(name)
	if slices.Contains(e.EntityArguments(), name) {
		return checks.bad(at, "argument %q of %q is the call's entity, which its target supplies", name, e.Name)
	}
	properties, _ := e.ArgsSchema["properties"].(map[string]any)
	property, declared := properties[name]
	if !declared {
		return checks.bad(at, "effector %q declares no argument %q", e.Name, name)
	}
	if violation := valueViolation(property, value); violation != "" {
		return checks.bad(at, "argument %q of %q: %s", name, e.Name, violation)
	}
	return nil
}

func valueViolation(property, value any) string {
	schema, err := jsonschema.Compile(property)
	if err != nil {
		return err.Error()
	}
	if violations := schema.Validate(value); len(violations) > 0 {
		return violations[0].Msg
	}
	return ""
}

func requiredArguments(e *model.Effector) []string {
	required, _ := e.ArgsSchema["required"].([]any)
	names := make([]string, 0, len(required))
	for _, item := range required {
		if name, ok := item.(string); ok {
			names = append(names, name)
		}
	}
	return names
}
