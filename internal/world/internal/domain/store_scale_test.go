package domain

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestANegativeScaleStateStopsAFlowInsteadOfReversingIt(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(testsupport.TransferDomain())
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	health := doc["kinds"].([]any)[0].(map[string]any)["state"].([]any)[2].(map[string]any)
	delete(health, "min")
	health["initial"] = -0.5
	if raw, err = json.Marshal(doc); err != nil {
		t.Fatal(err)
	}
	spec, err := domain.Parse(raw, "negative-scale.json")
	if err != nil {
		t.Fatal(err)
	}
	w := newTestWorld(t, spec, 1, model.DefaultStartTimeNS)
	ent := w.entities[w.EntityIDs()[0]]
	if got := w.flowScale(ent, &ent.kind.Flows[0], model.DefaultStartTimeNS); got != 0 {
		t.Fatalf("a flow scaled by pump speed 1 and pipe health -0.5 has scale %v, want 0", got)
	}
}
