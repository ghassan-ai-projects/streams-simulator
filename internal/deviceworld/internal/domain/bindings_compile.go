package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"slices"
)

func decodeBindings(data []byte) (*bindingsDocument, error) {
	var document bindingsDocument
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&document); err != nil {
		return nil, fmt.Errorf("deviceworld: decode bindings: %w", err)
	}
	if err := rejectTrailingBindings(decoder); err != nil {
		return nil, err
	}
	return &document, nil
}

func rejectTrailingBindings(decoder *json.Decoder) error {
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return fmt.Errorf("deviceworld: bindings contain trailing JSON")
		}
		return fmt.Errorf("deviceworld: decode trailing bindings: %w", err)
	}
	return nil
}

func compileBindings(documents map[string]bindingDocument, entity string) (map[string]Binding, error) {
	bindings := make(map[string]Binding, len(documents))
	for _, target := range slices.Sorted(maps.Keys(documents)) {
		binding, err := compileBinding(target, documents[target], entity)
		if err != nil {
			return nil, err
		}
		bindings[target] = binding
	}
	return bindings, nil
}

func compileBinding(target string, document bindingDocument, entity string) (Binding, error) {
	if target == "" || document.Effector == "" {
		return Binding{}, fmt.Errorf("deviceworld: target %q must declare an effector", target)
	}
	binding := Binding{effector: document.Effector, entity: entity, valueState: document.ValueState}
	if err := binding.loadSources(target, document); err != nil {
		return Binding{}, err
	}
	if err := binding.admitEntity(target); err != nil {
		return Binding{}, err
	}
	return binding, nil
}

func (binding *Binding) loadSources(target string, document bindingDocument) error {
	arguments, err := loadArguments(document.Arguments)
	if err != nil {
		return fmt.Errorf("deviceworld: target %q: %w", target, err)
	}
	binding.arguments = arguments
	if document.SafeStop != nil {
		return binding.loadSafeStop(target, document.SafeStop)
	}
	return nil
}

func (binding *Binding) loadSafeStop(target string, document *safeStopDocument) error {
	if document.Effector == "" {
		return fmt.Errorf("deviceworld: target %q safe_stop must declare an effector", target)
	}
	binding.safeStopEffector = document.Effector
	arguments, err := loadArguments(document.Arguments)
	if err != nil {
		return fmt.Errorf("deviceworld: target %q safe_stop: %w", target, err)
	}
	binding.safeStopArgs = arguments
	return nil
}

func (binding Binding) admitEntity(target string) error {
	if argumentsNeedEntity(binding.arguments) && binding.entity == "" {
		return fmt.Errorf("deviceworld: target %q requires a world entity", target)
	}
	if argumentsNeedEntity(binding.safeStopArgs) && binding.entity == "" {
		return fmt.Errorf("deviceworld: target %q safe_stop requires a world entity", target)
	}
	return nil
}
