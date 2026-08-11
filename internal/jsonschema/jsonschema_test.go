package jsonschema

import (
	"encoding/json"
	"strings"
	"testing"
)

func compileDoc(t *testing.T, doc string) *Schema {
	t.Helper()
	var v any
	if err := json.Unmarshal([]byte(doc), &v); err != nil {
		t.Fatalf("bad test schema: %v", err)
	}
	s, err := Compile(v)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	return s
}

func validateDoc(t *testing.T, s *Schema, doc string) []Error {
	t.Helper()
	var v any
	if err := json.Unmarshal([]byte(doc), &v); err != nil {
		t.Fatalf("bad test value: %v", err)
	}
	return s.Validate(v)
}

func TestBasicTypesAndRequired(t *testing.T) {
	s := compileDoc(t, `{"type": "object", "required": ["id"], "properties": {"id": {"type": "string", "pattern": "^[a-z]+$"}, "count": {"type": "integer", "minimum": 0}}}`)
	if errs := validateDoc(t, s, `{"id": "abc", "count": 3}`); len(errs) != 0 {
		t.Fatalf("valid doc rejected: %v", errs)
	}
	errs := validateDoc(t, s, `{"count": 3}`)
	if len(errs) != 1 || !strings.Contains(errs[0].Msg, "missing required") {
		t.Fatalf("missing required not caught: %v", errs)
	}
	errs = validateDoc(t, s, `{"id": "ABC"}`)
	if len(errs) != 1 || !strings.Contains(errs[0].Msg, "pattern") {
		t.Fatalf("pattern not caught: %v", errs)
	}
	errs = validateDoc(t, s, `{"id": "abc", "count": -1}`)
	if len(errs) != 1 || !strings.Contains(errs[0].Msg, "minimum") {
		t.Fatalf("minimum not caught: %v", errs)
	}
}

func TestAdditionalProperties(t *testing.T) {
	s := compileDoc(t, `{"type": "object", "additionalProperties": false, "properties": {"a": {"type": "number"}}}`)
	errs := validateDoc(t, s, `{"a": 1, "b": 2}`)
	if len(errs) != 1 || !strings.Contains(errs[0].Msg, "additional property") {
		t.Fatalf("additionalProperties=false not enforced: %v", errs)
	}
	// additionalProperties as a schema.
	s2 := compileDoc(t, `{"type": "object", "properties": {"a": {"type": "number"}}, "additionalProperties": {"type": "string"}}`)
	if errs := validateDoc(t, s2, `{"x": "ok"}`); len(errs) != 0 {
		t.Fatalf("schema additionalProperties rejected: %v", errs)
	}
	if errs := validateDoc(t, s2, `{"x": 3}`); len(errs) != 1 {
		t.Fatalf("schema additionalProperties not enforced: %v", errs)
	}
}

func TestEnumConstAndNumericForms(t *testing.T) {
	s := compileDoc(t, `{"enum": ["a", "b", 1, 1.0]}`)
	if errs := validateDoc(t, s, `"a"`); len(errs) != 0 {
		t.Fatalf("enum member rejected: %v", errs)
	}
	if errs := validateDoc(t, s, `1`); len(errs) != 0 {
		t.Fatalf("enum 1 rejected (1.0 must equal 1): %v", errs)
	}
	if errs := validateDoc(t, s, `"c"`); len(errs) != 1 {
		t.Fatalf("non-member accepted: %v", errs)
	}
	c := compileDoc(t, `{"const": "0.1"}`)
	if errs := validateDoc(t, c, `"0.1"`); len(errs) != 0 {
		t.Fatalf("const rejected: %v", errs)
	}
	if errs := validateDoc(t, c, `0.1`); len(errs) != 1 {
		t.Fatalf("const type confusion: %v", errs)
	}
	n := compileDoc(t, `{"type": "number", "exclusiveMinimum": 0, "maximum": 1}`)
	if errs := validateDoc(t, n, `0`); len(errs) != 1 {
		t.Fatalf("exclusiveMinimum not enforced: %v", errs)
	}
	if errs := validateDoc(t, n, `1.5`); len(errs) != 1 {
		t.Fatalf("maximum not enforced: %v", errs)
	}
}

func TestArraysAndRefs(t *testing.T) {
	s := compileDoc(t, `{
		"type": "object",
		"properties": {
			"items": {"type": "array", "minItems": 1, "items": {"$ref": "#/$defs/item"}}
		},
		"$defs": {"item": {"type": "object", "required": ["name"], "properties": {"name": {"type": "string"}}}}
	}`)
	if errs := validateDoc(t, s, `{"items": []}`); len(errs) != 1 {
		t.Fatalf("minItems not enforced: %v", errs)
	}
	if errs := validateDoc(t, s, `{"items": [{"name": "x"}]}`); len(errs) != 0 {
		t.Fatalf("valid items rejected: %v", errs)
	}
	if errs := validateDoc(t, s, `{"items": [{}]}`); len(errs) != 1 {
		t.Fatalf("ref item validation failed: %v", errs)
	}
}

func TestCombinatorsAndConditionals(t *testing.T) {
	// RBE deadband pattern: when cadence.mode == report_by_exception then
	// deadband is required.
	s := compileDoc(t, `{
		"type": "object",
		"properties": {
			"cadence": {"type": "object", "properties": {"mode": {"enum": ["periodic", "report_by_exception"]}}, "required": ["mode"]}
		},
		"allOf": [
			{"if": {"properties": {"cadence": {"properties": {"mode": {"const": "report_by_exception"}}, "required": ["mode"]}}, "required": ["cadence"]},
			 "then": {"properties": {"cadence": {"required": ["deadband"]}}}}
		]
	}`)
	if errs := validateDoc(t, s, `{"cadence": {"mode": "report_by_exception", "deadband": 0.5}}`); len(errs) != 0 {
		t.Fatalf("valid RBE rejected: %v", errs)
	}
	if errs := validateDoc(t, s, `{"cadence": {"mode": "report_by_exception"}}`); len(errs) != 1 {
		t.Fatalf("RBE without deadband accepted: %v", errs)
	}
	if errs := validateDoc(t, s, `{"cadence": {"mode": "periodic"}}`); len(errs) != 0 {
		t.Fatalf("periodic wrongly constrained: %v", errs)
	}

	// oneOf with type discrimination.
	one := compileDoc(t, `{"oneOf": [{"type": "string"}, {"type": "number"}]}`)
	if errs := validateDoc(t, one, `"x"`); len(errs) != 0 {
		t.Fatalf("oneOf string rejected: %v", errs)
	}
	if errs := validateDoc(t, one, `true`); len(errs) != 1 {
		t.Fatalf("oneOf accepted non-member: %v", errs)
	}
}

func TestContainsAndFormat(t *testing.T) {
	s := compileDoc(t, `{"type": "array", "contains": {"type": "string", "pattern": "^x"}}`)
	if errs := validateDoc(t, s, `["a", "xb"]`); len(errs) != 0 {
		t.Fatalf("contains rejected valid: %v", errs)
	}
	if errs := validateDoc(t, s, `["a", "b"]`); len(errs) != 1 {
		t.Fatalf("contains not enforced: %v", errs)
	}
	dt := compileDoc(t, `{"type": "string", "format": "date-time"}`)
	if errs := validateDoc(t, dt, `"2026-08-11T05:00:00.123456789Z"`); len(errs) != 0 {
		t.Fatalf("valid date-time rejected: %v", errs)
	}
	if errs := validateDoc(t, dt, `"not a time"`); len(errs) != 1 {
		t.Fatalf("invalid date-time accepted: %v", errs)
	}
}

func TestTypeFailureShortCircuits(t *testing.T) {
	s := compileDoc(t, `{"type": "string", "minLength": 5}`)
	errs := validateDoc(t, s, `123`)
	if len(errs) != 1 {
		t.Fatalf("type failure should short-circuit: %v", errs)
	}
}

func TestPathReporting(t *testing.T) {
	s := compileDoc(t, `{"type": "object", "properties": {"a": {"type": "array", "items": {"type": "object", "required": ["x"]}}}}`)
	errs := validateDoc(t, s, `{"a": [{"x": 1}, {}]}`)
	if len(errs) != 1 || errs[0].Path != "/a/1" {
		t.Fatalf("path wrong: %v", errs)
	}
}
