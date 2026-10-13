package domain

import (
	"fmt"
	"maps"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func ValidateBindings(w *world.World, bindings map[string]Binding, requiredTargets, requiredSafeStops []string) error {
	if err := validateBindingCatalog(w, bindings); err != nil {
		return err
	}
	if err := requireBoundTargets(bindings, requiredTargets); err != nil {
		return err
	}
	return requireSafeStopTargets(bindings, requiredSafeStops)
}

func validateEffectorBinding(target, role, name string, effector *model.Effector, arguments map[string]bindingArgument) error {
	if name == "" || effector == nil {
		return fmt.Errorf("deviceworld: target %q %s references unknown effector %q", target, role, name)
	}
	properties, _ := effector.ArgsSchema["properties"].(map[string]any)
	if err := validateArgumentProperties(target, role, name, properties, arguments); err != nil {
		return err
	}
	required, _ := effector.ArgsSchema["required"].([]any)
	return validateRequiredArguments(target, role, name, required, arguments)
}

func validateBindingCatalog(w *world.World, bindings map[string]Binding) error {
	if w == nil || w.Spec == nil {
		return fmt.Errorf("deviceworld: world is required")
	}
	if len(bindings) == 0 {
		return fmt.Errorf("deviceworld: bindings declare no targets")
	}
	for _, target := range slices.Sorted(maps.Keys(bindings)) {
		if err := validateWorldBinding(w, target, bindings[target]); err != nil {
			return err
		}
	}
	return nil
}

func validateWorldBinding(w *world.World, target string, binding Binding) error {
	if binding.entity == "" || w.Entity(binding.entity) == nil {
		return fmt.Errorf("deviceworld: target %q references unknown entity %q", target, binding.entity)
	}
	if binding.valueState != "" && !slices.Contains(w.StateNamesOf(binding.entity), binding.valueState) {
		return fmt.Errorf("deviceworld: target %q references unknown value_state %q", target, binding.valueState)
	}
	effector := kindEffector(w, binding.entity, binding.effector)
	if err := validateEffectorBinding(target, "effector", binding.effector, effector, binding.arguments); err != nil {
		return err
	}
	return validateSafeStopBinding(w, target, binding)
}

func kindEffector(w *world.World, entityID, name string) *model.Effector {
	return w.EffectorOf(entityID, name)
}

func validateSafeStopBinding(w *world.World, target string, binding Binding) error {
	if binding.safeStopEffector == "" {
		return nil
	}
	effector := kindEffector(w, binding.entity, binding.safeStopEffector)
	return validateEffectorBinding(target, "safe_stop", binding.safeStopEffector, effector, binding.safeStopArgs)
}

func requireBoundTargets(bindings map[string]Binding, targets []string) error {
	for _, target := range targets {
		if _, ok := bindings[target]; !ok {
			return fmt.Errorf("deviceworld: required target %q has no world binding", target)
		}
	}
	return nil
}

func requireSafeStopTargets(bindings map[string]Binding, targets []string) error {
	for _, target := range targets {
		binding, ok := bindings[target]
		if !ok || binding.safeStopEffector == "" {
			return fmt.Errorf("deviceworld: required target %q has no explicit safe-stop binding", target)
		}
	}
	return nil
}

func validateArgumentProperties(target, role, effector string, properties map[string]any, arguments map[string]bindingArgument) error {
	for _, argument := range slices.Sorted(maps.Keys(arguments)) {
		if _, ok := properties[argument]; !ok {
			return fmt.Errorf("deviceworld: target %q %s argument %q is not declared by effector %q", target, role, argument, effector)
		}
	}
	return nil
}

func validateRequiredArguments(target, role, effector string, required []any, arguments map[string]bindingArgument) error {
	for _, raw := range required {
		field, ok := raw.(string)
		if !ok {
			return fmt.Errorf("deviceworld: target %q %s effector %q has invalid required argument declaration", target, role, effector)
		}
		if _, ok := arguments[field]; !ok {
			return fmt.Errorf("deviceworld: target %q %s binding omits required argument %q for effector %q", target, role, field, effector)
		}
	}
	return nil
}
