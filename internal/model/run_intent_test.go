package model

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/schemas"
)

func TestTheTerminationReasonsAreTheEnumOfBothContracts(t *testing.T) {
	t.Parallel()
	for name, schema := range map[string][]byte{"run artifact": schemas.RunArtifact(), "run report": schemas.RunReport()} {
		var doc map[string]any
		if err := json.Unmarshal(schema, &doc); err != nil {
			t.Fatal(err)
		}
		reasons := enumAt(t, doc, name)
		if !slices.Equal(reasons, TerminationReasons()) {
			t.Errorf("%s enum = %v, want %v", name, reasons, TerminationReasons())
		}
	}
}

func enumAt(t *testing.T, schema map[string]any, name string) []string {
	t.Helper()
	node := schema["properties"].(map[string]any)
	if name == "run report" {
		node = node["run"].(map[string]any)["properties"].(map[string]any)
	}
	reason := node["termination"].(map[string]any)["properties"].(map[string]any)["reason"].(map[string]any)
	var out []string
	for _, value := range reason["enum"].([]any) {
		out = append(out, value.(string))
	}
	return out
}

func TestAnUnknownTerminationReasonIsNotOne(t *testing.T) {
	t.Parallel()
	if !IsTerminationReason(TerminationCompleted) || !IsTerminationReason(TerminationStoppedByOperator) || IsTerminationReason("") || IsTerminationReason("walked_away") {
		t.Fatal("IsTerminationReason disagrees with the enum")
	}
}
