package jsonschema

import "fmt"

func (s *Schema) validatePrimitive(v any, fail func(string)) {
	s.validatePattern(v, fail)
	if str, ok := v.(string); ok {
		s.validateStringLimits(str, fail)
	}
	if f, ok := asFloat(v); ok {
		s.validateNumberLimits(f, fail)
	}
}

func (s *Schema) validatePattern(v any, fail func(string)) {
	if s.pattern != nil {
		if str, ok := v.(string); ok && !s.pattern.MatchString(str) {
			fail(fmt.Sprintf("string does not match pattern %s", s.pattern))
		}
	}
}

func (s *Schema) validateStringLimits(str string, fail func(string)) {
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

func (s *Schema) validateNumberLimits(f float64, fail func(string)) {
	if s.hasMin && f < s.min {
		fail(fmt.Sprintf("number %v below minimum %v", f, s.min))
	}
	if s.hasMax && f > s.max {
		fail(fmt.Sprintf("number %v above maximum %v", f, s.max))
	}
	s.validateExclusiveLimits(f, fail)
	s.validateMultiple(f, fail)
}

func (s *Schema) validateExclusiveLimits(f float64, fail func(string)) {
	if s.hasExMin && f <= s.exMin {
		fail(fmt.Sprintf("number %v not above exclusiveMinimum %v", f, s.exMin))
	}
	if s.hasExMax && f >= s.exMax {
		fail(fmt.Sprintf("number %v not below exclusiveMaximum %v", f, s.exMax))
	}
}

func (s *Schema) validateMultiple(f float64, fail func(string)) {
	if s.hasMult && s.multipleOf != 0 {
		q := f / s.multipleOf
		if q != float64(int64(q)) {
			fail(fmt.Sprintf("number %v is not a multiple of %v", f, s.multipleOf))
		}
	}
}
