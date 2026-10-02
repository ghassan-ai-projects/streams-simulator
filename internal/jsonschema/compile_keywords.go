package jsonschema

import (
	"fmt"
	"sort"
)

// fill populates a Schema from its JSON form (the compile body, without the
// $ref shortcut).
func fill(s *Schema, m map[string]any, defs map[string]any, reg map[string]*Schema, depth int) error {
	for _, fillKeywords := range []func(*Schema, map[string]any, map[string]any, map[string]*Schema, int) error{fillValueKeywords, fillObjectKeywords, fillArrayKeywords, fillStringKeywords, fillNumberKeywords, fillCompositionKeywords, fillConditionalKeywords} {
		if err := fillKeywords(s, m, defs, reg, depth); err != nil {
			return err
		}
	}
	return nil
}

func fillValueKeywords(s *Schema, m map[string]any, defs map[string]any, reg map[string]*Schema, depth int) error {
	if t, ok := m["type"]; ok {
		switch tv := t.(type) {
		case string:
			s.types = []string{tv}
		case []any:
			for _, e := range tv {
				str, ok := e.(string)
				if !ok {
					return fmt.Errorf("jsonschema: type array must hold strings")
				}
				s.types = append(s.types, str)
			}
		default:
			return fmt.Errorf("jsonschema: type must be a string or array")
		}
	}
	if e, ok := m["enum"]; ok {
		arr, ok := e.([]any)
		if !ok {
			return fmt.Errorf("jsonschema: enum must be an array")
		}
		s.enum = arr
	}
	if c, ok := m["const"]; ok {
		s.constVal, s.hasConst = c, true
	}
	return nil
}

func fillObjectKeywords(s *Schema, m map[string]any, defs map[string]any, reg map[string]*Schema, depth int) error {
	if p, ok := m["properties"].(map[string]any); ok {
		s.properties = make(map[string]*Schema, len(p))
		for name, sub := range p {
			subm, ok := sub.(map[string]any)
			if !ok {
				return fmt.Errorf("jsonschema: property %q must be an object", name)
			}
			cs, err := compile(subm, defs, reg, depth+1)
			if err != nil {
				return fmt.Errorf("jsonschema: property %q: %w", name, err)
			}
			s.properties[name] = cs
		}
	}
	if r, ok := m["required"]; ok {
		arr, ok := r.([]any)
		if !ok {
			return fmt.Errorf("jsonschema: required must be an array")
		}
		for _, e := range arr {
			str, ok := e.(string)
			if !ok {
				return fmt.Errorf("jsonschema: required entries must be strings")
			}
			s.required = append(s.required, str)
		}
		sort.Strings(s.required)
	}
	switch add := m["additionalProperties"].(type) {
	case bool:
		s.addAllowed = add
	case map[string]any:
		cs, err := compile(add, defs, reg, depth+1)
		if err != nil {
			return fmt.Errorf("jsonschema: %w", err)
		}
		s.additional = cs
	case nil:
	default:
		return fmt.Errorf("jsonschema: additionalProperties must be a boolean or object")
	}
	return nil
}

func fillArrayKeywords(s *Schema, m map[string]any, defs map[string]any, reg map[string]*Schema, depth int) error {
	if it, ok := m["items"].(map[string]any); ok {
		cs, err := compile(it, defs, reg, depth+1)
		if err != nil {
			return fmt.Errorf("jsonschema: %w", err)
		}
		s.items = cs
	}
	if v, ok := m["minItems"]; ok {
		s.minItems = toInt(v)
		s.hasMinItems = true
	}
	if v, ok := m["maxItems"]; ok {
		s.maxItems = toInt(v)
		s.hasMaxItems = true
	}
	if v, ok := m["uniqueItems"]; ok {
		s.uniqueItems = v.(bool)
	}
	if c, ok := m["contains"].(map[string]any); ok {
		cs, err := compile(c, defs, reg, depth+1)
		if err != nil {
			return fmt.Errorf("jsonschema: %w", err)
		}
		s.contains = cs
	}
	return nil
}
