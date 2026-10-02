package jsonschema

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// fill populates a Schema from its JSON form (the compile body, without the
// $ref shortcut).
func fill(s *Schema, m map[string]any, defs map[string]any, reg map[string]*Schema, depth int) error {
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

func refName(ref string) (string, error) {
	if !strings.HasPrefix(ref, "#/$defs/") {
		return "", fmt.Errorf("jsonschema: unsupported $ref %q (local #/$defs refs only)", ref)
	}
	return strings.TrimPrefix(ref, "#/$defs/"), nil
}

func toInt(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	case uint64:
		// #nosec G115 -- schema integer keywords are small literals.
		return int(n)
	case jsonNumber:
		i, err := strconv.Atoi(n.String())
		if err != nil {
			return 0
		}
		return i
	default:
		return 0
	}
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case jsonNumber:
		f, err := n.Float64()
		return f, err == nil
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint64:
		return float64(n), true
	}
	return 0, false
}
