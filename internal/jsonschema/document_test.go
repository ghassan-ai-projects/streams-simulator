package jsonschema

import (
	"fmt"
	"strings"
	"testing"
)

func TestCompileJSONCompilesAWorkingValidator(t *testing.T) {
	t.Parallel()
	schema, err := CompileJSON([]byte(`{"type":"object","required":["n"],"properties":{"n":{"type":"integer"},"s":{"type":"string","minLength":3}}}`))
	if err != nil {
		t.Fatal(err)
	}
	for name, tc := range map[string]struct {
		doc     map[string]any
		invalid bool
	}{
		"valid":            {doc: map[string]any{"n": float64(1)}},
		"missing required": {doc: map[string]any{}, invalid: true},
		"short string":     {doc: map[string]any{"n": float64(1), "s": "ab"}, invalid: true},
		"long enough":      {doc: map[string]any{"n": float64(1), "s": "abc"}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if errs := schema.Validate(tc.doc); (len(errs) > 0) != tc.invalid {
				t.Fatalf("errors = %v, invalid want %v", errs, tc.invalid)
			}
		})
	}
}

func TestCompileJSONRejectsBadSchemaDocuments(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct{ raw, want string }{
		"not json":   {`{"type":`, "not valid JSON"},
		"empty":      {``, "not valid JSON"},
		"not object": {`[1]`, "must be an object"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := CompileJSON([]byte(tc.raw)); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want %q", err, tc.want)
			}
		})
	}
}

// Trailing data after the first value is ignored today (the embedded schemas
// never have any); pinned so a change is deliberate.
func TestCompileJSONIgnoresTrailingData(t *testing.T) {
	t.Parallel()
	if _, err := CompileJSON([]byte(`{"type":"object"} trailing`)); err != nil {
		t.Fatalf("err = %v", err)
	}
}

func TestFormatErrorsBoundsTheReport(t *testing.T) {
	t.Parallel()
	errs := make([]Error, 12)
	for i := range errs {
		errs[i] = Error{Path: fmt.Sprintf("/p%d", i), Msg: "bad"}
	}
	got := FormatErrors(errs)
	lines := strings.Split(got, "\n")
	if len(lines) != 11 || lines[10] != "  ... and 2 more" {
		t.Fatalf("report = %q", got)
	}
	if FormatErrors(nil) != "" {
		t.Fatal("no errors must render empty")
	}
	if one := FormatErrors(errs[:1]); strings.HasSuffix(one, "\n") || !strings.HasPrefix(one, "  ") {
		t.Fatalf("single error = %q", one)
	}
}
