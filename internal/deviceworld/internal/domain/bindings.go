package domain

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
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
	SafeStop   *safeStopDocument           `json:"safe_stop,omitempty"`
}

type safeStopDocument struct {
	Effector  string                      `json:"effector"`
	Arguments map[string]argumentDocument `json:"arguments,omitempty"`
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
	document, err := decodeBindings(data)
	if err != nil {
		return nil, err
	}
	if len(document.Bindings) == 0 {
		return nil, fmt.Errorf("deviceworld: bindings declare no targets")
	}
	return compileBindings(document.Bindings, entity)
}

func loadArguments(doc map[string]argumentDocument) (map[string]bindingArgument, error) {
	arguments := make(map[string]bindingArgument, len(doc))
	for _, name := range slices.Sorted(maps.Keys(doc)) {
		argument := doc[name]
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
	if err := validateArgumentSource(doc); err != nil {
		return bindingArgument{}, err
	}
	if doc.Source == "value" {
		if err := json.Unmarshal(doc.Value, &argument.value); err != nil {
			return bindingArgument{}, fmt.Errorf("decode value: %w", err)
		}
	}
	return argument, nil
}

func argumentsNeedEntity(arguments map[string]bindingArgument) bool {
	// determinism-safe: an existence test; the answer ignores order.
	for _, argument := range arguments {
		if argument.source == "entity" {
			return true
		}
	}
	return false
}

func (b Binding) args(params map[string]float64) (map[string]any, error) {
	return resolveArgs(b.arguments, b.entity, params)
}

func (b Binding) safeStopArgsForEntity() (map[string]any, error) {
	return resolveArgs(b.safeStopArgs, b.entity, nil)
}

func resolveArgs(arguments map[string]bindingArgument, entity string, params map[string]float64) (map[string]any, error) {
	args := make(map[string]any, len(arguments))
	for _, name := range slices.Sorted(maps.Keys(arguments)) {
		value, err := arguments[name].resolve(entity, params)
		if err != nil {
			return nil, err
		}
		args[name] = value
	}
	return args, nil
}

func validateArgumentSource(doc argumentDocument) error {
	switch doc.Source {
	case "entity":
		return requireArgumentShape(doc.Parameter == "" && len(doc.Value) == 0, "entity source cannot set parameter or value")
	case "parameter":
		return requireArgumentShape(doc.Parameter != "" && len(doc.Value) == 0, "parameter source requires only parameter")
	case "value":
		return requireArgumentShape(doc.Parameter == "" && len(doc.Value) != 0, "value source requires only value")
	default:
		return fmt.Errorf("unknown source %q", doc.Source)
	}
}

func requireArgumentShape(valid bool, message string) error {
	if !valid {
		return fmt.Errorf("%s", message)
	}
	return nil
}

func (argument bindingArgument) resolve(entity string, params map[string]float64) (any, error) {
	switch argument.source {
	case "entity":
		return entity, nil
	case "parameter":
		return parameterArgument(params, argument.parameter)
	case "value":
		return argument.value, nil
	default:
		return nil, fmt.Errorf("unknown argument source %q", argument.source)
	}
}

func parameterArgument(params map[string]float64, name string) (any, error) {
	value, ok := params[name]
	if !ok {
		return nil, fmt.Errorf("parameter %q is not present", name)
	}
	return value, nil
}
