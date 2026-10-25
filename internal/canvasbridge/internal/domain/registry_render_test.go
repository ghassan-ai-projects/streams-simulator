package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestAnActionResultThatCannotBeEncodedNamesTheAction(t *testing.T) {
	t.Parallel()
	var marshal *json.MarshalerError
	_, err := Action{Name: "kind_check"}.encode(map[string]json.RawMessage{"broken": json.RawMessage("{")})
	if !errors.As(err, &marshal) || !strings.Contains(err.Error(), "canvasbridge: kind_check result") {
		t.Fatalf("encode = %v, want a marshal error naming the action", err)
	}
}

func TestANumberArgumentIsRenderedAsItsShortestLiteral(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		value any
		want  string
		fails bool
	}{
		{"a whole float", float64(3), "3", false},
		{"a fraction", 0.25, "0.25", false},
		{"a large float without an exponent", 1e21, "1000000000000000000000", false},
		{"a JSON number", json.Number("12.50"), "12.50", false},
		{"a malformed JSON number", json.Number("twelve"), "", true},
		{"a string", "3", "", true},
		{"a bool", true, "", true},
		{"nothing", nil, "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			part, err := renderNumber(tc.value)
			if tc.fails {
				if err == nil || !strings.HasPrefix(err.Error(), "want a number") {
					t.Fatalf("renderNumber(%v) = %v, want a 'want a number' error", tc.value, err)
				}
				return
			}
			if err != nil || part.Literal != tc.want || part.Path != "" {
				t.Fatalf("renderNumber(%v) = %+v, %v, want literal %s", tc.value, part, err, tc.want)
			}
		})
	}
}

func TestAnArgumentOfAnUnknownShapeIsNotRendered(t *testing.T) {
	t.Parallel()
	for _, shape := range []ArgShape{"", "stems", "switch", "mystery"} {
		t.Run(string(shape), func(t *testing.T) {
			t.Parallel()
			part, err := CLIArg{Key: "k", Shape: shape}.render("value")
			if err == nil || !strings.Contains(err.Error(), `unknown shape "`+string(shape)+`"`) || part != (ArgvPart{}) {
				t.Fatalf("render with shape %q = %+v, %v, want an unknown shape error", shape, part, err)
			}
		})
	}
}

func TestANumberShapedArgumentRendersThroughRender(t *testing.T) {
	t.Parallel()
	part, err := CLIArg{Key: "min", Flag: "--min", Shape: ShapeNum}.render(json.Number("4"))
	if err != nil || part.Literal != "4" {
		t.Fatalf("render = %+v, %v, want the literal 4", part, err)
	}
}
