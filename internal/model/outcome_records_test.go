package model

import (
	"encoding/json"
	"slices"
	"testing"
)

func TestAnEffectorsEntityArgumentsAreItsRequiredStringIDs(t *testing.T) {
	t.Parallel()
	schema := map[string]any{"required": []any{"percent", "feeder_id", "line_id", "count_id"}, "properties": map[string]any{
		"percent": map[string]any{"type": "number"}, "feeder_id": map[string]any{"type": "string"},
		"line_id": map[string]any{"type": "string"}, "count_id": map[string]any{"type": "integer"}, "belt_id": map[string]any{"type": "string"}}}
	got := (&Effector{ArgsSchema: schema}).EntityArguments()
	if !slices.Equal(got, []string{"feeder_id", "line_id"}) {
		t.Fatalf("entity arguments = %v, want the required string arguments named *_id", got)
	}
}

func TestAGroundTruthWithoutOutcomeDataKeepsItsShape(t *testing.T) {
	t.Parallel()
	raw, err := json.Marshal(GroundTruth{NegativeClassFraction: 0.4})
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != `{"negative_class_fraction":0.4}` {
		t.Fatalf("ground truth = %s, want outcome_weights and baseline_policies left out", raw)
	}
}

func TestAPolicyCallTargetsMatchedOnlyByThatWord(t *testing.T) {
	t.Parallel()
	for target, want := range map[string]bool{PolicyTargetMatched: true, "F-1": false, "Matched": false} {
		if got := (PolicyCall{Target: target}).TargetsMatched(); got != want {
			t.Errorf("target %q: matched = %v, want %v", target, got, want)
		}
	}
}

func TestOutcomeFactsLeaveOutTheTermsAnEntityDidNotHave(t *testing.T) {
	t.Parallel()
	facts := OutcomeFacts{HorizonNS: 3600e9, Entities: []EntityOutcome{{EntityID: "F-1", Kind: "bulk_feeder"}},
		Calls: []OutcomeCall{{Effector: "bulk_feeder__set_feed", EntityID: "F-1", AtNS: 60e9, Accepted: true}}}
	raw, err := json.Marshal(facts)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"horizon_ns":3600000000000,"entities":[{"entity_id":"F-1","kind":"bulk_feeder"}],` +
		`"calls":[{"effector":"bulk_feeder__set_feed","entity_id":"F-1","at_ns":60000000000,"accepted":true}]}`
	if string(raw) != want {
		t.Fatalf("facts = %s, want %s", raw, want)
	}
}
