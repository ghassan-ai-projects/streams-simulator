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
	if err := fillTypeKeyword(s, m); err != nil {
		return err
	}
	if err := fillEnum(s, m); err != nil {
		return err
	}
	if c, ok := m["const"]; ok {
		s.constVal, s.hasConst = c, true
	}
	return nil
}

func fillObjectKeywords(s *Schema, m map[string]any, defs map[string]any, reg map[string]*Schema, depth int) error {
	if err := fillProperties(s, m, defs, reg, depth); err != nil {
		return err
	}
	if err := fillRequired(s, m); err != nil {
		return err
	}
	return fillAdditional(s, m, defs, reg, depth)
}

func fillProperties(s *Schema, m, defs map[string]any, reg map[string]*Schema, depth int) error {
	properties, ok := m["properties"].(map[string]any)
	if !ok {
		return nil
	}
	s.properties = make(map[string]*Schema, len(properties))
	return fillPropertySchemas(s, properties, defs, reg, depth)
}

func fillPropertySchemas(s *Schema, properties, defs map[string]any, reg map[string]*Schema, depth int) error {
	for name, sub := range properties {
		cs, err := compileProperty(name, sub, defs, reg, depth)
		if err != nil {
			return err
		}
		s.properties[name] = cs
	}
	return nil
}

func fillRequired(s *Schema, m map[string]any) error {
	if r, ok := m["required"]; ok {
		arr, ok := r.([]any)
		if !ok {
			return fmt.Errorf("jsonschema: required must be an array")
		}
		if err := fillRequiredEntries(s, arr); err != nil {
			return err
		}
		sort.Strings(s.required)
	}
	return nil
}

func fillAdditional(s *Schema, m, defs map[string]any, reg map[string]*Schema, depth int) error {
	switch add := m["additionalProperties"].(type) {
	case bool:
		s.addAllowed = add
	case map[string]any:
		return fillSubschema(&s.additional, add, defs, reg, depth)
	case nil:
	default:
		return fmt.Errorf("jsonschema: additionalProperties must be a boolean or object")
	}
	return nil
}

func fillArrayKeywords(s *Schema, m map[string]any, defs map[string]any, reg map[string]*Schema, depth int) error {
	if err := fillSubschema(&s.items, m["items"], defs, reg, depth); err != nil {
		return err
	}
	fillArrayCardinality(s, m)
	return fillSubschema(&s.contains, m["contains"], defs, reg, depth)
}

func fillSubschema(dest **Schema, value any, defs map[string]any, reg map[string]*Schema, depth int) error {
	if m, ok := value.(map[string]any); ok {
		compiled, err := compile(m, defs, reg, depth+1)
		if err != nil {
			return fmt.Errorf("jsonschema: %w", err)
		}
		*dest = compiled
	}
	return nil
}

func fillTypeKeyword(s *Schema, m map[string]any) error {
	if t, ok := m["type"]; ok {
		switch tv := t.(type) {
		case string:
			s.types = []string{tv}
		case []any:
			return fillTypeArray(s, tv)
		default:
			return fmt.Errorf("jsonschema: type must be a string or array")
		}
	}
	return nil
}

func fillTypeArray(s *Schema, types []any) error {
	for _, e := range types {
		str, ok := e.(string)
		if !ok {
			return fmt.Errorf("jsonschema: type array must hold strings")
		}
		s.types = append(s.types, str)
	}
	return nil
}

func fillArrayCardinality(s *Schema, m map[string]any) {
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
}

func fillEnum(s *Schema, m map[string]any) error {
	if e, ok := m["enum"]; ok {
		arr, ok := e.([]any)
		if !ok {
			return fmt.Errorf("jsonschema: enum must be an array")
		}
		s.enum = arr
	}
	return nil
}

func compileProperty(name string, sub any, defs map[string]any, reg map[string]*Schema, depth int) (*Schema, error) {
	subm, ok := sub.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("jsonschema: property %q must be an object", name)
	}
	cs, err := compile(subm, defs, reg, depth+1)
	if err != nil {
		return nil, fmt.Errorf("jsonschema: property %q: %w", name, err)
	}
	return cs, nil
}

func fillRequiredEntries(s *Schema, arr []any) error {
	for _, e := range arr {
		str, ok := e.(string)
		if !ok {
			return fmt.Errorf("jsonschema: required entries must be strings")
		}
		s.required = append(s.required, str)
	}
	return nil
}
