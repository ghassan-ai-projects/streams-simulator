package domain

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func structuredDocument(t *testing.T, change func(doc map[string]any)) []byte {
	t.Helper()
	return documentOf(t, testsupport.StructuredDomain(), change)
}

func documentOf(t *testing.T, path string, change func(doc map[string]any)) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	change(doc)
	out, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func at(doc any, keys ...any) any {
	for _, key := range keys {
		switch typed := key.(type) {
		case string:
			doc = doc.(map[string]any)[typed]
		case int:
			doc = doc.([]any)[typed]
		}
	}
	return doc
}

func put(doc any, value any, keys ...any) {
	parent := at(doc, keys[:len(keys)-1]...)
	switch last := keys[len(keys)-1].(type) {
	case string:
		parent.(map[string]any)[last] = value
	case int:
		parent.([]any)[last] = value
	}
}

func structuredViolation(t *testing.T, change func(doc map[string]any), pointer string) {
	t.Helper()
	_, err := Parse(structuredDocument(t, change), "structured.json")
	var invalid *model.ValidationError
	if !errors.As(err, &invalid) || len(invalid.Issues) != 1 || invalid.Issues[0].Pointer != pointer {
		t.Fatalf("err = %v, want one issue at %s", err, pointer)
	}
}

func loadStructured(t *testing.T) *Compiled {
	t.Helper()
	c, err := Load(testsupport.StructuredDomain())
	if err != nil {
		t.Fatal(err)
	}
	return c
}

type violationCase struct {
	name    string
	change  func(doc map[string]any)
	pointer string
}

func runViolations(t *testing.T, cases []violationCase) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			structuredViolation(t, tc.change, tc.pointer)
		})
	}
}
