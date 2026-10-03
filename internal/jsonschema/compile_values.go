package jsonschema

import (
	"fmt"
	"regexp"
)

func fillStringKeywords(s *Schema, m map[string]any, defs map[string]any, reg map[string]*Schema, depth int) error {
	if v, ok := m["minLength"]; ok {
		s.minLen = toInt(v)
		s.hasMinLen = true
	}
	if v, ok := m["maxLength"]; ok {
		s.maxLen = toInt(v)
		s.hasMaxLen = true
	}
	return fillPattern(s, m)
}

func fillNumberKeywords(s *Schema, m map[string]any, defs map[string]any, reg map[string]*Schema, depth int) error {
	if v, ok := m["minimum"]; ok {
		s.min, s.hasMin = toFloat(v)
	}
	if v, ok := m["maximum"]; ok {
		s.max, s.hasMax = toFloat(v)
	}
	fillExclusiveLimits(s, m)
	return nil
}

func fillCompositionKeywords(s *Schema, m map[string]any, defs map[string]any, reg map[string]*Schema, depth int) error {
	for _, keyword := range []string{"oneOf", "anyOf", "allOf"} {
		if branches, ok := m[keyword].([]any); ok {
			compiled, err := compileBranches(keyword, branches, defs, reg, depth)
			if err != nil {
				return err
			}
			assignComposition(s, keyword, compiled)
		}
	}
	return nil
}

func compileBranches(keyword string, branches []any, defs map[string]any, reg map[string]*Schema, depth int) ([]*Schema, error) {
	var out []*Schema
	for _, branch := range branches {
		m, ok := branch.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("jsonschema: %s entries must be objects", keyword)
		}
		compiled, err := compile(m, defs, reg, depth+1)
		if err != nil {
			return nil, fmt.Errorf("jsonschema: %w", err)
		}
		out = append(out, compiled)
	}
	return out, nil
}

func assignComposition(s *Schema, keyword string, branches []*Schema) {
	switch keyword {
	case "oneOf":
		s.oneOf = branches
	case "anyOf":
		s.anyOf = branches
	case "allOf":
		s.allOf = branches
	}
}

func fillConditionalKeywords(s *Schema, m map[string]any, defs map[string]any, reg map[string]*Schema, depth int) error {
	if err := fillSubschema(&s.not, m["not"], defs, reg, depth); err != nil {
		return err
	}
	if err := fillCondition(s, m, defs, reg, depth); err != nil {
		return err
	}
	if format, ok := m["format"].(string); ok {
		s.format = format
	}
	return nil
}

func fillCondition(s *Schema, m, defs map[string]any, reg map[string]*Schema, depth int) error {
	if condition, ok := m["if"].(map[string]any); ok {
		if err := fillSubschema(&s.ifS, condition, defs, reg, depth); err != nil {
			return err
		}
		if err := fillSubschema(&s.thenS, m["then"], defs, reg, depth); err != nil {
			return err
		}
		return fillSubschema(&s.elseS, m["else"], defs, reg, depth)
	}
	return nil
}

func fillPattern(s *Schema, m map[string]any) error {
	if p, ok := m["pattern"].(string); ok {
		re, err := regexp.Compile(p)
		if err != nil {
			return fmt.Errorf("jsonschema: bad pattern %q: %w", p, err)
		}
		s.pattern = re
	}
	return nil
}

func fillExclusiveLimits(s *Schema, m map[string]any) {
	if v, ok := m["exclusiveMinimum"]; ok {
		s.exMin, s.hasExMin = toFloat(v)
	}
	if v, ok := m["exclusiveMaximum"]; ok {
		s.exMax, s.hasExMax = toFloat(v)
	}
	if v, ok := m["multipleOf"]; ok {
		s.multipleOf, s.hasMult = toFloat(v)
	}
}
