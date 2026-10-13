package jsonvalue

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEveryDecodedNumberFormReadsTheSame(t *testing.T) {
	t.Parallel()
	for _, value := range []any{2.5, json.Number("2.5")} {
		if got, ok := Number(value); !ok || got != 2.5 {
			t.Fatalf("Number(%#v) = %v, %v", value, got, ok)
		}
	}
	for _, value := range []any{3, int64(3), uint64(3), json.Number("3")} {
		if got, ok := Number(value); !ok || got != 3 {
			t.Fatalf("Number(%#v) = %v, %v", value, got, ok)
		}
	}
	for _, value := range []any{"3", json.Number("x"), nil, true} {
		if _, ok := Number(value); ok {
			t.Fatalf("Number(%#v) read a number", value)
		}
	}
}

func TestAStrictDocumentRefusesAnUndeclaredField(t *testing.T) {
	t.Parallel()
	var target struct {
		A int `json:"a"`
	}
	if err := DecodeStrict([]byte(`{"a":1}`), &target); err != nil || target.A != 1 {
		t.Fatalf("decode = %v, %+v", err, target)
	}
	if err := DecodeStrict([]byte(`{"a":1,"b":2}`), &target); err == nil || !strings.Contains(err.Error(), `unknown field "b"`) {
		t.Fatalf("an undeclared field = %v", err)
	}
}
