package testsupport

import (
	"reflect"
	"testing"
)

func TestArgsForSchemaBindsRequiredArguments(t *testing.T) {
	t.Parallel()
	schema := map[string]any{
		"properties": map[string]any{
			"pond_id": map[string]any{"type": "string"},
			"mode":    map[string]any{"type": "string"},
			"rate":    map[string]any{"type": "number"},
			"on":      map[string]any{"type": "boolean"},
			"other":   map[string]any{"type": "object"},
		},
		"required": []any{"pond_id", "mode", "rate", "on", "other", "", "missing"},
	}
	want := map[string]any{"pond_id": "e1", "mode": "x", "rate": 1, "on": true, "other": "", "missing": ""}
	if got := ArgsForSchema(schema, "e1"); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got := ArgsForSchema(nil, "e1"); len(got) != 0 {
		t.Fatalf("nil schema produced %v", got)
	}
}
