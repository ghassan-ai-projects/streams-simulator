package protocol

import (
	"encoding/json"
	"testing"
)

func TestArgumentReadersAcceptTheNumberFormsAJSONClientSends(t *testing.T) {
	t.Parallel()
	args := map[string]any{
		"float": float64(7), "int": int64(8), "number": json.Number("9"), "text": "10", "word": "x", "fraction": 1.5,
	}
	for key, want := range map[string]int64{"float": 7, "int": 8, "number": 9, "text": 10} {
		if got, ok := intArg(args, key); !ok || got != want {
			t.Errorf("intArg(%s) = %d, %v; want %d", key, got, ok, want)
		}
	}
	if _, ok := intArg(args, "word"); ok {
		t.Error("a non-numeric string is not an integer")
	}
	if got := numberArg(args, "absent", 42); got != 42 {
		t.Errorf("Num default = %d", got)
	}
	if got := stringArg(args, "text"); got != "10" {
		t.Errorf("Str = %q", got)
	}
}
