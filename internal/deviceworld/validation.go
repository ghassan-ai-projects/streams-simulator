package deviceworld

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

// ValidateBindings checks a complete device/world composition before a
// listener is opened. The mapping remains data-defined; this function only
// verifies that its names and argument shape agree with the loaded world.
func ValidateBindings(w *world.World, bindings map[string]Binding, requiredTargets, requiredSafeStops []string) error {
	if w == nil || w.Spec == nil {
		return fmt.Errorf("deviceworld: world is required")
	}
	if len(bindings) == 0 {
		return fmt.Errorf("deviceworld: bindings declare no targets")
	}
	states := make(map[string]struct{}, len(w.Spec.StateNames()))
	for _, name := range w.Spec.StateNames() {
		states[name] = struct{}{}
	}
	for target, binding := range bindings {
		if binding.entity == "" || w.Entity(binding.entity) == nil {
			return fmt.Errorf("deviceworld: target %q references unknown entity %q", target, binding.entity)
		}
		if _, ok := states[binding.valueState]; binding.valueState != "" && !ok {
			return fmt.Errorf("deviceworld: target %q references unknown value_state %q", target, binding.valueState)
		}
		effector := w.Spec.Effector(binding.effector)
		if err := validateEffectorBinding(target, "effector", binding.effector, effector, binding.arguments); err != nil {
			return err
		}
		if binding.safeStopEffector != "" {
			safeStop := w.Spec.Effector(binding.safeStopEffector)
			if err := validateEffectorBinding(target, "safe_stop", binding.safeStopEffector, safeStop, binding.safeStopArgs); err != nil {
				return err
			}
		}
	}
	for _, target := range requiredTargets {
		if _, ok := bindings[target]; !ok {
			return fmt.Errorf("deviceworld: required target %q has no world binding", target)
		}
	}
	for _, target := range requiredSafeStops {
		binding, ok := bindings[target]
		if !ok || binding.safeStopEffector == "" {
			return fmt.Errorf("deviceworld: required target %q has no explicit safe-stop binding", target)
		}
	}
	return nil
}

func validateEffectorBinding(target, role, name string, effector *model.Effector, arguments map[string]bindingArgument) error {
	if name == "" || effector == nil {
		return fmt.Errorf("deviceworld: target %q %s references unknown effector %q", target, role, name)
	}
	properties, _ := effector.ArgsSchema["properties"].(map[string]any)
	for argument := range arguments {
		if _, ok := properties[argument]; !ok {
			return fmt.Errorf("deviceworld: target %q %s argument %q is not declared by effector %q", target, role, argument, name)
		}
	}
	if required, _ := effector.ArgsSchema["required"].([]any); len(required) > 0 {
		for _, raw := range required {
			field, ok := raw.(string)
			if !ok {
				return fmt.Errorf("deviceworld: target %q %s effector %q has invalid required argument declaration", target, role, name)
			}
			if _, ok := arguments[field]; !ok {
				return fmt.Errorf("deviceworld: target %q %s binding omits required argument %q for effector %q", target, role, field, name)
			}
		}
	}
	return nil
}
