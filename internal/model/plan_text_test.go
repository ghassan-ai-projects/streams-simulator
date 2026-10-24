package model

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/schemas"
)

func TestAuthoredPlanTextIsPlainAndBounded(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		planName, purpose string
		want              string
	}{
		"both empty":              {"", "", ""},
		"plain text":              {"Cooling loss, staged response", "Observe the chiller with the tower sensor out.", ""},
		"name at the limit":       {strings.Repeat("é", MaxPlanNameChars), "", ""},
		"purpose at the limit":    {"", strings.Repeat("é", MaxPlanPurposeChars), ""},
		"name over the limit":     {strings.Repeat("a", MaxPlanNameChars+1), "", "name"},
		"purpose over the limit":  {"", strings.Repeat("a", MaxPlanPurposeChars+1), "purpose"},
		"newline in the name":     {"two\nlines", "", "name"},
		"tab in the purpose":      {"", "a\tb", "purpose"},
		"escape in the purpose":   {"", "a\x1b[31mb", "purpose"},
		"delete in the name":      {"a\x7fb", "", "name"},
		"invalid utf-8 in a name": {"a\xffb", "", "name"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			err := CheckPlanText(tc.planName, tc.purpose)
			if tc.want == "" && err != nil {
				t.Fatalf("CheckPlanText = %v, want it accepted", err)
			}
			if tc.want != "" && (err == nil || !strings.Contains(err.Error(), tc.want)) {
				t.Fatalf("CheckPlanText = %v, want a refusal naming %q", err, tc.want)
			}
		})
	}
}

func TestTheRunPlanTextLimitsAreTheLimitsOfBothContracts(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		schema []byte
		path   []string
	}{
		"run artifact": {schemas.RunArtifact(), []string{"properties", "plan", "properties"}},
		"run report":   {schemas.RunReport(), []string{"properties", "run", "properties"}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var doc map[string]any
			if err := json.Unmarshal(tc.schema, &doc); err != nil {
				t.Fatal(err)
			}
			node := doc
			for _, key := range tc.path {
				node = node[key].(map[string]any)
			}
			if got := node["name"].(map[string]any)["maxLength"]; got != float64(MaxPlanNameChars) {
				t.Errorf("name maxLength = %v, want %d", got, MaxPlanNameChars)
			}
			if got := node["purpose"].(map[string]any)["maxLength"]; got != float64(MaxPlanPurposeChars) {
				t.Errorf("purpose maxLength = %v, want %d", got, MaxPlanPurposeChars)
			}
		})
	}
}
