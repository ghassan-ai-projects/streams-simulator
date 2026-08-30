package deviceworld

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// bindingArgument is a data-defined source for one world effector argument.
// The adapter supports only generic sources; domain values live in the binding
// catalog, not in Go callbacks.
type bindingArgument struct {
	source    string
	parameter string
	value     any
}

type bindingsDocument struct {
	Bindings map[string]bindingDocument `json:"bindings"`
}

type bindingDocument struct {
	Effector   string                      `json:"effector"`
	ValueState string                      `json:"value_state,omitempty"`
	Arguments  map[string]argumentDocument `json:"arguments,omitempty"`
}

type argumentDocument struct {
	Source    string          `json:"source"`
	Parameter string          `json:"parameter,omitempty"`
	Value     json.RawMessage `json:"value,omitempty"`
}

// LoadBindings parses a strict device-target to world-effector binding catalog.
// Entity-source arguments resolve to the supplied runtime world entity; all
// other mapping and literal values come from the JSON catalog.
func LoadBindings(data []byte, entity string) (map[string]Binding, error) {
	var doc bindingsDocument
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&doc); err != nil {
		return nil, fmt.Errorf("deviceworld: decode bindings: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("deviceworld: bindings contain trailing JSON")
		}
		return nil, fmt.Errorf("deviceworld: decode trailing bindings: %w", err)
	}
	if len(doc.Bindings) == 0 {
		return nil, fmt.Errorf("deviceworld: bindings declare no targets")
	}

	bindings := make(map[string]Binding, len(doc.Bindings))
	for target, binding := range doc.Bindings {
		if target == "" || binding.Effector == "" {
			return nil, fmt.Errorf("deviceworld: target %q must declare an effector", target)
		}
		arguments, err := loadArguments(binding.Arguments)
		if err != nil {
			return nil, fmt.Errorf("deviceworld: target %q: %w", target, err)
		}
		if argumentsNeedEntity(arguments) && entity == "" {
			return nil, fmt.Errorf("deviceworld: target %q requires a world entity", target)
		}
		bindings[target] = Binding{
			effector:   binding.Effector,
			entity:     entity,
			valueState: binding.ValueState,
			arguments:  arguments,
		}
	}
	return bindings, nil
}

func loadArguments(doc map[string]argumentDocument) (map[string]bindingArgument, error) {
	arguments := make(map[string]bindingArgument, len(doc))
	for name, argument := range doc {
		if name == "" {
			return nil, fmt.Errorf("argument name must not be empty")
		}
		parsed, err := parseArgument(argument)
		if err != nil {
			return nil, fmt.Errorf("argument %q: %w", name, err)
		}
		arguments[name] = parsed
	}
	return arguments, nil
}

func parseArgument(doc argumentDocument) (bindingArgument, error) {
	argument := bindingArgument{source: doc.Source, parameter: doc.Parameter}
	switch doc.Source {
	case "entity":
		if doc.Parameter != "" || len(doc.Value) != 0 {
			return bindingArgument{}, fmt.Errorf("entity source cannot set parameter or value")
		}
	case "parameter":
		if doc.Parameter == "" || len(doc.Value) != 0 {
			return bindingArgument{}, fmt.Errorf("parameter source requires only parameter")
		}
	case "value":
		if doc.Parameter != "" || len(doc.Value) == 0 {
			return bindingArgument{}, fmt.Errorf("value source requires only value")
		}
		if err := json.Unmarshal(doc.Value, &argument.value); err != nil {
			return bindingArgument{}, fmt.Errorf("decode value: %w", err)
		}
	default:
		return bindingArgument{}, fmt.Errorf("unknown source %q", doc.Source)
	}
	return argument, nil
}

func argumentsNeedEntity(arguments map[string]bindingArgument) bool {
	for _, argument := range arguments {
		if argument.source == "entity" {
			return true
		}
	}
	return false
}

func (b Binding) args(params map[string]float64) (map[string]any, error) {
	args := make(map[string]any, len(b.arguments))
	for name, argument := range b.arguments {
		switch argument.source {
		case "entity":
			args[name] = b.entity
		case "parameter":
			value, ok := params[argument.parameter]
			if !ok {
				return nil, fmt.Errorf("parameter %q is not present", argument.parameter)
			}
			args[name] = value
		case "value":
			args[name] = argument.value
		default:
			return nil, fmt.Errorf("unknown argument source %q", argument.source)
		}
	}
	return args, nil
}
