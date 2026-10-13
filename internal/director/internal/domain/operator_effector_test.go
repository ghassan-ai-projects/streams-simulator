package domain

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestTheOperatorSeesNoEffectorDescriptionOrSchemaAnnotation(t *testing.T) {
	t.Parallel()
	effector := &model.Effector{Name: "maintenance_reset", Description: "recalibrates the probes that were lying", ArgsSchema: map[string]any{
		"type": "object", "title": "hint", "description": "clears the probe_fouling fault",
		"properties": map[string]any{"zone": map[string]any{"type": "string", "description": "the zone whose probe lies", "examples": []any{"z-1"}, "enum": []any{"z-1", "z-2"}}},
		"anyOf":      []any{map[string]any{"required": []any{"zone"}, "$comment": "fault hint"}},
		"required":   []any{"zone"},
	}}
	info := OperatorEffector(effector)
	raw, err := json.Marshal(info)
	if err != nil {
		t.Fatal(err)
	}
	for _, leak := range []string{"lying", "hint", "fouling", "examples", "description", "title", "$comment"} {
		if strings.Contains(string(raw), leak) {
			t.Fatalf("the operator view carries %q: %s", leak, raw)
		}
	}
	for _, kept := range []string{`"enum":["z-1","z-2"]`, `"required":["zone"]`, `"risk_class":"medium"`} {
		if !strings.Contains(string(raw), kept) {
			t.Fatalf("the operator view lost %s: %s", kept, raw)
		}
	}
	if effector.ArgsSchema["description"] == nil {
		t.Fatal("the director's schema must stay intact")
	}
}
