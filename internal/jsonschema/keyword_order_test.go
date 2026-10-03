package jsonschema

import (
	"reflect"
	"testing"
)

func TestValidationPreservesKeywordOrderAndTypeShortCircuit(t *testing.T) {
	t.Parallel()
	schema, err := Compile(map[string]any{
		"type": "string", "enum": []any{"allowed"}, "pattern": "^x", "minLength": float64(5),
		"allOf": []any{map[string]any{"const": "expected"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []Error{{Msg: "value not in enum"}, {Msg: "string does not match pattern ^x"}, {Msg: "string shorter than minLength 5"}, {Msg: "value does not match const"}}
	if got := schema.Validate("a"); !reflect.DeepEqual(got, want) {
		t.Fatalf("errors=%+v, want %+v", got, want)
	}
	if got := schema.Validate(float64(1)); len(got) != 1 || got[0].Msg != "expected type string, got number" {
		t.Fatalf("type must short-circuit: %+v", got)
	}
}
