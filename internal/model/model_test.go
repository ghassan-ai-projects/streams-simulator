package model

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDecodeRejectsTrailingJSON(t *testing.T) {
	var got any
	if err := Decode(strings.NewReader(`{"ok":true} {"trailing":true}`), &got); err == nil {
		t.Fatal("decoder accepted two concatenated JSON documents")
	}
	if err := Decode(strings.NewReader(`{"ok":true} trailing`), &got); err == nil {
		t.Fatal("decoder accepted trailing non-JSON data")
	}
}

func TestDecodePreservesLargeNumbers(t *testing.T) {
	var got map[string]any
	if err := Decode(strings.NewReader(`{"seed":18446744073709551615}`), &got); err != nil {
		t.Fatal(err)
	}
	if got["seed"].(json.Number).String() != "18446744073709551615" {
		t.Fatalf("large number lost precision: %#v", got["seed"])
	}
}
