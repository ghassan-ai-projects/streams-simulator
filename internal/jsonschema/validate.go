package jsonschema

import (
	"fmt"
	"strings"
)

func (s *Schema) Validate(v any) []Error {
	var errs []Error
	s.validate(v, "", &errs)
	return errs
}

func (s *Schema) validate(v any, path string, errs *[]Error) {
	fail := func(keyword, msg string) { *errs = append(*errs, Error{Path: path, Msg: msg, Keyword: keyword}) }
	if !s.validateValueKeywords(v, fail) {
		return
	}
	s.validatePrimitive(v, fail)
	s.validateCollections(v, path, errs)
	s.validateComposition(v, path, errs, fail)
}

func (s *Schema) validateValueKeywords(v any, fail failFunc) bool {
	if len(s.types) > 0 && !typeMatches(s.types, v) {
		fail("type", fmt.Sprintf("expected type %s, got %s", strings.Join(s.types, "|"), typeName(v)))
		return false
	}
	if len(s.enum) > 0 && !enumContains(s.enum, v) {
		fail("enum", "value not in enum")
	}
	if s.hasConst && !jsonEqual(s.constVal, v) {
		fail("const", "value does not match const")
	}
	return true
}

func (s *Schema) validateCollections(v any, path string, errs *[]Error) {
	switch x := v.(type) {
	case map[string]any:
		s.validateObject(x, path, errs)
	case []any:
		s.validateArray(x, path, errs)
	}
}
