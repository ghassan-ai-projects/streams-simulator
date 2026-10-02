package jsonschema

import (
	"fmt"
	"regexp"
)

// fill populates a Schema from its JSON form (the compile body, without the
// $ref shortcut).

func fillStringKeywords(s *Schema, m map[string]any, defs map[string]any, reg map[string]*Schema, depth int) error {
	if v, ok := m["minLength"]; ok {
		s.minLen = toInt(v)
		s.hasMinLen = true
	}
	if v, ok := m["maxLength"]; ok {
		s.maxLen = toInt(v)
		s.hasMaxLen = true
	}
	if p, ok := m["pattern"].(string); ok {
		re, err := regexp.Compile(p)
		if err != nil {
			return fmt.Errorf("jsonschema: bad pattern %q: %w", p, err)
		}
		s.pattern = re
	}
	return nil
}

func fillNumberKeywords(s *Schema, m map[string]any, defs map[string]any, reg map[string]*Schema, depth int) error {
	if v, ok := m["minimum"]; ok {
		s.min, s.hasMin = toFloat(v)
	}
	if v, ok := m["maximum"]; ok {
		s.max, s.hasMax = toFloat(v)
	}
	if v, ok := m["exclusiveMinimum"]; ok {
		s.exMin, s.hasExMin = toFloat(v)
	}
	if v, ok := m["exclusiveMaximum"]; ok {
		s.exMax, s.hasExMax = toFloat(v)
	}
	if v, ok := m["multipleOf"]; ok {
		s.multipleOf, s.hasMult = toFloat(v)
	}
	return nil
}

func fillCompositionKeywords(s *Schema, m map[string]any, defs map[string]any, reg map[string]*Schema, depth int) error {
	for _, kw := range []string{"oneOf", "anyOf", "allOf"} {
		if arr, ok := m[kw].([]any); ok {
			var out []*Schema
			for _, sub := range arr {
				subm, ok := sub.(map[string]any)
				if !ok {
					return fmt.Errorf("jsonschema: %s entries must be objects", kw)
				}
				cs, err := compile(subm, defs, reg, depth+1)
				if err != nil {
					return fmt.Errorf("jsonschema: %w", err)
				}
				out = append(out, cs)
			}
			switch kw {
			case "oneOf":
				s.oneOf = out
			case "anyOf":
				s.anyOf = out
			case "allOf":
				s.allOf = out
			}
		}
	}
	return nil
}

func fillConditionalKeywords(s *Schema, m map[string]any, defs map[string]any, reg map[string]*Schema, depth int) error {
	if n, ok := m["not"].(map[string]any); ok {
		cs, err := compile(n, defs, reg, depth+1)
		if err != nil {
			return fmt.Errorf("jsonschema: %w", err)
		}
		s.not = cs
	}
	if cond, ok := m["if"].(map[string]any); ok {
		cs, err := compile(cond, defs, reg, depth+1)
		if err != nil {
			return fmt.Errorf("jsonschema: %w", err)
		}
		s.ifS = cs
		if then, ok := m["then"].(map[string]any); ok {
			tcs, err := compile(then, defs, reg, depth+1)
			if err != nil {
				return fmt.Errorf("jsonschema: %w", err)
			}
			s.thenS = tcs
		}
		if els, ok := m["else"].(map[string]any); ok {
			ecs, err := compile(els, defs, reg, depth+1)
			if err != nil {
				return fmt.Errorf("jsonschema: %w", err)
			}
			s.elseS = ecs
		}
	}
	if f, ok := m["format"].(string); ok {
		s.format = f
	}
	return nil
}
