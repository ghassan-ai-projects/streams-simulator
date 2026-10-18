package jsonschema

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonvalue"
)

func (s *Schema) validatePrimitive(v any, fail failFunc) {
	s.validatePattern(v, fail)
	if str, ok := v.(string); ok {
		s.validateStringLimits(str, fail)
	}
	if f, ok := jsonvalue.Number(v); ok {
		s.validateNumberLimits(f, fail)
	}
}

func (s *Schema) validatePattern(v any, fail failFunc) {
	str, ok := v.(string)
	if s.pattern == nil || !ok || s.pattern.MatchString(str) {
		return
	}
	if s.expected != "" {
		fail("pattern", fmt.Sprintf("%q is not %s", str, s.expected))
		return
	}
	fail("pattern", fmt.Sprintf("string does not match pattern %s", s.pattern))
}

func (s *Schema) validateStringLimits(str string, fail failFunc) {
	if s.hasMinLen && len([]rune(str)) < s.minLen {
		fail("minLength", fmt.Sprintf("string shorter than minLength %d", s.minLen))
	}
	if s.hasMaxLen && len([]rune(str)) > s.maxLen {
		fail("maxLength", fmt.Sprintf("string longer than maxLength %d", s.maxLen))
	}
	if s.format != "" {
		if msg := checkFormat(s.format, str); msg != "" {
			fail("format", msg)
		}
	}
}

func (s *Schema) validateNumberLimits(f float64, fail failFunc) {
	if s.hasMin && f < s.min {
		fail("minimum", fmt.Sprintf("number %v below minimum %v", f, s.min))
	}
	if s.hasMax && f > s.max {
		fail("maximum", fmt.Sprintf("number %v above maximum %v", f, s.max))
	}
	s.validateExclusiveLimits(f, fail)
	s.validateMultiple(f, fail)
}

func (s *Schema) validateExclusiveLimits(f float64, fail failFunc) {
	if s.hasExMin && f <= s.exMin {
		fail("exclusiveMinimum", fmt.Sprintf("number %v not above exclusiveMinimum %v", f, s.exMin))
	}
	if s.hasExMax && f >= s.exMax {
		fail("exclusiveMaximum", fmt.Sprintf("number %v not below exclusiveMaximum %v", f, s.exMax))
	}
}

func (s *Schema) validateMultiple(f float64, fail failFunc) {
	if s.hasMult && s.multipleOf != 0 {
		q := f / s.multipleOf
		if q != float64(int64(q)) {
			fail("multipleOf", fmt.Sprintf("number %v is not a multiple of %v", f, s.multipleOf))
		}
	}
}
