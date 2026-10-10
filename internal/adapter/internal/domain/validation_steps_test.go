package domain

import (
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestNestedProjectionValidationRetainsSourceContext(t *testing.T) {
	t.Parallel()
	a := &model.Adapter{Record: model.RecordTemplate{Fields: []model.Field{{Name: "nested", From: model.ValueExpr{Op: "object", Fields: []model.Field{
		{Name: "id", From: model.ValueExpr{Op: "source", Source: "seq"}},
		{Name: "bad", From: model.ValueExpr{Op: "concat", Parts: []model.ValueExpr{{Op: "source", Source: "not_native"}}}},
	}}}}}}
	err := crossCheck(a, "nested-input")
	if err == nil || !strings.Contains(err.Error(), `nested-input: unknown source "not_native"`) {
		t.Fatalf("source context lost: %v", err)
	}
	a.Record.Fields[0].From.Fields = a.Record.Fields[0].From.Fields[:1]
	if err := crossCheck(a, "nested-input"); err != nil {
		t.Fatalf("nested identity rejected: %v", err)
	}
}
