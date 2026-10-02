package jsonschema

import (
	"fmt"
)

// Compile builds a Schema from a decoded JSON Schema document. Recursive
// $defs references are supported: every top-level $def is pre-registered as
// a placeholder and filled after the root schema compiles, so self-referring
// definitions (like the adapter's valueExpr) compile fine and only
// terminate at validation time, walking finite values.
func Compile(doc any) (*Schema, error) {
	m, ok := doc.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("jsonschema: schema document must be an object")
	}
	defs := map[string]any{}
	if d, ok := m["$defs"].(map[string]any); ok {
		defs = d
	}
	reg := map[string]*Schema{}
	for name := range defs {
		reg[name] = &Schema{addAllowed: true}
	}
	root, err := compile(m, defs, reg, 0)
	if err != nil {
		return nil, fmt.Errorf("jsonschema: %w", err)
	}
	for name, raw := range defs {
		rm, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("jsonschema: $defs/%s must be an object", name)
		}
		if err := fill(reg[name], rm, defs, reg, 0); err != nil {
			return nil, fmt.Errorf("jsonschema: $defs/%s: %w", name, err)
		}
	}
	return root, nil
}

func compile(m map[string]any, defs map[string]any, reg map[string]*Schema, depth int) (*Schema, error) {
	if depth > 64 {
		return nil, fmt.Errorf("jsonschema: schema nesting too deep")
	}
	if r, ok := m["$ref"].(string); ok {
		name, err := refName(r)
		if err != nil {
			return nil, fmt.Errorf("jsonschema: %w", err)
		}
		if s, ok := reg[name]; ok {
			return s, nil
		}
		raw, ok := defs[name]
		if !ok {
			return nil, fmt.Errorf("jsonschema: unknown $defs entry %q", name)
		}
		rm, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("jsonschema: $defs/%s is not an object", name)
		}
		s := &Schema{addAllowed: true}
		reg[name] = s
		if err := fill(s, rm, defs, reg, depth+1); err != nil {
			return nil, fmt.Errorf("jsonschema: %w", err)
		}
		return s, nil
	}
	s := &Schema{addAllowed: true}
	if err := fill(s, m, defs, reg, depth+1); err != nil {
		return nil, fmt.Errorf("jsonschema: %w", err)
	}
	return s, nil
}
