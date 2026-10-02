package jsonschema

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
)

// Validate checks v against the schema and returns every failure found.
func (s *Schema) Validate(v any) []Error {
	var errs []Error
	s.validate(v, "", &errs)
	return errs
}

func (s *Schema) validate(v any, path string, errs *[]Error) {
	fail := func(msg string) {
		*errs = append(*errs, Error{Path: path, Msg: msg})
	}
	if len(s.types) > 0 && !typeMatches(s.types, v) {
		fail(fmt.Sprintf("expected type %s, got %s", strings.Join(s.types, "|"), typeName(v)))
		return // type failure short-circuits value keywords
	}
	if len(s.enum) > 0 && !enumContains(s.enum, v) {
		fail("value not in enum")
	}
	if s.hasConst && !jsonEqual(s.constVal, v) {
		fail("value does not match const")
	}
	s.validatePrimitive(v, fail)
	switch x := v.(type) {
	case map[string]any:
		s.validateObject(x, path, errs)
	case []any:
		s.validateArray(x, path, errs)
	}
	s.validateComposition(v, path, errs, fail)
}

func (s *Schema) validatePrimitive(v any, fail func(string)) {
	if s.pattern != nil {
		if str, ok := v.(string); ok && !s.pattern.MatchString(str) {
			fail(fmt.Sprintf("string does not match pattern %s", s.pattern))
		}
	}
	if str, ok := v.(string); ok {
		if s.hasMinLen && len([]rune(str)) < s.minLen {
			fail(fmt.Sprintf("string shorter than minLength %d", s.minLen))
		}
		if s.hasMaxLen && len([]rune(str)) > s.maxLen {
			fail(fmt.Sprintf("string longer than maxLength %d", s.maxLen))
		}
		if s.format != "" {
			if msg := checkFormat(s.format, str); msg != "" {
				fail(msg)
			}
		}
	}
	if f, ok := asFloat(v); ok {
		if s.hasMin && f < s.min {
			fail(fmt.Sprintf("number %v below minimum %v", f, s.min))
		}
		if s.hasMax && f > s.max {
			fail(fmt.Sprintf("number %v above maximum %v", f, s.max))
		}
		if s.hasExMin && f <= s.exMin {
			fail(fmt.Sprintf("number %v not above exclusiveMinimum %v", f, s.exMin))
		}
		if s.hasExMax && f >= s.exMax {
			fail(fmt.Sprintf("number %v not below exclusiveMaximum %v", f, s.exMax))
		}
		if s.hasMult && s.multipleOf != 0 {
			q := f / s.multipleOf
			if q != float64(int64(q)) {
				fail(fmt.Sprintf("number %v is not a multiple of %v", f, s.multipleOf))
			}
		}
	}
}

func (s *Schema) validateComposition(v any, path string, errs *[]Error, fail func(string)) {
	for _, sub := range s.allOf {
		sub.validate(v, path, errs)
	}
	if len(s.oneOf) > 0 {
		matches := 0
		for _, sub := range s.oneOf {
			if len(sub.Validate(v)) == 0 {
				matches++
			}
		}
		if matches != 1 {
			fail(fmt.Sprintf("value must match exactly one oneOf branch (matched %d)", matches))
		}
	}
	if len(s.anyOf) > 0 {
		matched := false
		for _, sub := range s.anyOf {
			if len(sub.Validate(v)) == 0 {
				matched = true
				break
			}
		}
		if !matched {
			fail("value matches no anyOf branch")
		}
	}
	if s.not != nil {
		if len(s.not.Validate(v)) == 0 {
			fail("value must not match the not schema")
		}
	}
	if s.ifS != nil {
		if len(s.ifS.Validate(v)) == 0 {
			if s.thenS != nil {
				s.thenS.validate(v, path, errs)
			}
		} else if s.elseS != nil {
			s.elseS.validate(v, path, errs)
		}
	}
}

func (s *Schema) validateObject(obj map[string]any, path string, errs *[]Error) {
	for _, name := range s.required {
		if _, ok := obj[name]; !ok {
			*errs = append(*errs, Error{Path: path, Msg: fmt.Sprintf("missing required property %q", name)})
		}
	}
	// Properties validate in sorted order so the error report is stable
	// across runs (map iteration order is not).
	var names []string
	// determinism-safe: collected here, sorted below before any output.
	for name := range s.properties {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		sub := s.properties[name]
		val, ok := obj[name]
		if !ok {
			continue
		}
		sub.validate(val, joinPath(path, name), errs)
	}
	// Additional-property checks iterate in sorted order so validation
	// output is deterministic across runs (map order is not).
	var extra []string
	// determinism-safe: collected here, sorted below before any output.
	for name := range obj {
		if _, known := s.properties[name]; !known {
			extra = append(extra, name)
		}
	}
	sort.Strings(extra)
	for _, name := range extra {
		val := obj[name]
		switch {
		case s.additional != nil:
			s.additional.validate(val, joinPath(path, name), errs)
		case !s.addAllowed:
			*errs = append(*errs, Error{Path: path, Msg: fmt.Sprintf("additional property %q not allowed", name)})
		}
	}
}

func (s *Schema) validateArray(arr []any, path string, errs *[]Error) {
	if s.hasMinItems && len(arr) < s.minItems {
		*errs = append(*errs, Error{Path: path, Msg: fmt.Sprintf("array shorter than minItems %d", s.minItems)})
	}
	if s.hasMaxItems && len(arr) > s.maxItems {
		*errs = append(*errs, Error{Path: path, Msg: fmt.Sprintf("array longer than maxItems %d", s.maxItems)})
	}
	if s.uniqueItems {
		seen := map[string]bool{}
		for _, e := range arr {
			key, err := canonical.MarshalString(e)
			if err == nil && seen[key] {
				*errs = append(*errs, Error{Path: path, Msg: "array items must be unique"})
				break
			}
			seen[key] = true
		}
	}
	if s.contains != nil {
		matched := false
		for i, e := range arr {
			if len(s.contains.Validate(e)) == 0 {
				matched = true
				break
			}
			_ = i
		}
		if !matched {
			*errs = append(*errs, Error{Path: path, Msg: "no array item matches contains"})
		}
	}
	if s.items != nil {
		for i, e := range arr {
			s.items.validate(e, joinPath(path, fmt.Sprintf("%d", i)), errs)
		}
	}
}
