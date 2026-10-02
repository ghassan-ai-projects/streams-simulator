package jsonschema

import "fmt"

func (s *Schema) validateComposition(v any, path string, errs *[]Error, fail func(string)) {
	for _, sub := range s.allOf {
		sub.validate(v, path, errs)
	}
	s.validateOneOf(v, fail)
	s.validateAnyOf(v, fail)
	s.validateNot(v, fail)
	s.validateCondition(v, path, errs)
}

func (s *Schema) validateOneOf(v any, fail func(string)) {
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
}

func (s *Schema) validateAnyOf(v any, fail func(string)) {
	if len(s.anyOf) > 0 && !anyBranchMatches(s.anyOf, v) {
		fail("value matches no anyOf branch")
	}
}

func anyBranchMatches(branches []*Schema, v any) bool {
	for _, sub := range branches {
		if len(sub.Validate(v)) == 0 {
			return true
		}
	}
	return false
}

func (s *Schema) validateNot(v any, fail func(string)) {
	if s.not != nil && len(s.not.Validate(v)) == 0 {
		fail("value must not match the not schema")
	}
}

func (s *Schema) validateCondition(v any, path string, errs *[]Error) {
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
