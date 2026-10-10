package jsonschema

import (
	"fmt"
	"maps"
	"slices"
)

// Compile builds a Schema from a decoded JSON Schema document. Definitions are
// registered before the root is filled so recursive references remain valid.
func Compile(doc any) (*Schema, error) {
	m, ok := doc.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("jsonschema: schema document must be an object")
	}
	return compileDocument(m)
}

func compileDocument(m map[string]any) (*Schema, error) {
	defs := schemaDefinitions(m)
	reg := definitionRegistry(defs)
	root, err := compile(m, defs, reg, 0)
	if err != nil {
		return nil, fmt.Errorf("jsonschema: %w", err)
	}
	if err := fillDefinitions(defs, reg); err != nil {
		return nil, err
	}
	return root, nil
}

func schemaDefinitions(m map[string]any) map[string]any {
	if defs, ok := m["$defs"].(map[string]any); ok {
		return defs
	}
	return map[string]any{}
}

func definitionRegistry(defs map[string]any) map[string]*Schema {
	reg := map[string]*Schema{}
	// determinism-safe: fills a map; no order is observable.
	for name := range defs {
		reg[name] = &Schema{addAllowed: true}
	}
	return reg
}

func fillDefinitions(defs map[string]any, reg map[string]*Schema) error {
	for _, name := range slices.Sorted(maps.Keys(defs)) {
		raw := defs[name]
		m, ok := raw.(map[string]any)
		if !ok {
			return fmt.Errorf("jsonschema: $defs/%s must be an object", name)
		}
		if err := fill(reg[name], m, defs, reg, 0); err != nil {
			return fmt.Errorf("jsonschema: $defs/%s: %w", name, err)
		}
	}
	return nil
}

func compile(m map[string]any, defs map[string]any, reg map[string]*Schema, depth int) (*Schema, error) {
	if depth > 64 {
		return nil, fmt.Errorf("jsonschema: schema nesting too deep")
	}
	if ref, ok := m["$ref"].(string); ok {
		return compileReference(ref, defs, reg, depth)
	}
	s := &Schema{addAllowed: true}
	if err := fill(s, m, defs, reg, depth+1); err != nil {
		return nil, fmt.Errorf("jsonschema: %w", err)
	}
	return s, nil
}

func compileReference(ref string, defs map[string]any, reg map[string]*Schema, depth int) (*Schema, error) {
	name, err := refName(ref)
	if err != nil {
		return nil, fmt.Errorf("jsonschema: %w", err)
	}
	if s, ok := reg[name]; ok {
		return s, nil
	}
	m, err := referencedDefinition(name, defs)
	if err != nil {
		return nil, err
	}
	return registerDefinition(name, m, defs, reg, depth)
}

func referencedDefinition(name string, defs map[string]any) (map[string]any, error) {
	raw, ok := defs[name]
	if !ok {
		return nil, fmt.Errorf("jsonschema: unknown $defs entry %q", name)
	}
	m, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("jsonschema: $defs/%s is not an object", name)
	}
	return m, nil
}

func registerDefinition(name string, m, defs map[string]any, reg map[string]*Schema, depth int) (*Schema, error) {
	s := &Schema{addAllowed: true}
	reg[name] = s
	if err := fill(s, m, defs, reg, depth+1); err != nil {
		return nil, fmt.Errorf("jsonschema: %w", err)
	}
	return s, nil
}
