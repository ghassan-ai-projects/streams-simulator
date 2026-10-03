package jsonschema

import (
	"fmt"
	"maps"
	"slices"
	"sort"
)

func (s *Schema) validateObject(obj map[string]any, path string, errs *[]Error) {
	s.validateRequired(obj, path, errs)
	s.validateProperties(obj, path, errs)
	s.validateAdditional(obj, path, errs)
}

func (s *Schema) validateRequired(obj map[string]any, path string, errs *[]Error) {
	for _, name := range s.required {
		if _, ok := obj[name]; !ok {
			*errs = append(*errs, Error{Path: path, Msg: fmt.Sprintf("missing required property %q", name)})
		}
	}
}

func (s *Schema) validateProperties(obj map[string]any, path string, errs *[]Error) {
	// Sorted names preserve the stable diagnostic order.
	for _, name := range slices.Sorted(maps.Keys(s.properties)) {
		val, ok := obj[name]
		if !ok {
			continue
		}
		s.properties[name].validate(val, joinPath(path, name), errs)
	}
}

func (s *Schema) validateAdditional(obj map[string]any, path string, errs *[]Error) {
	for _, name := range s.extraPropertyNames(obj) {
		val := obj[name]
		switch {
		case s.additional != nil:
			s.additional.validate(val, joinPath(path, name), errs)
		case !s.addAllowed:
			*errs = append(*errs, Error{Path: path, Msg: fmt.Sprintf("additional property %q not allowed", name)})
		}
	}
}

func (s *Schema) extraPropertyNames(obj map[string]any) []string {
	var extra []string
	// determinism-safe: collected keys are sorted before validation.
	for name := range obj {
		if _, known := s.properties[name]; !known {
			extra = append(extra, name)
		}
	}
	sort.Strings(extra)
	return extra
}
