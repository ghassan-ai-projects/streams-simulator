package domain

import (
	"encoding/json"
	"os"
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestAProfileCountSizesTheFirstTemplatedSet(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(testsupport.StructuredDomain())
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	sets := doc["instances"].([]any)
	doc["instances"] = append([]any{sets[4]}, append(sets[:4:4], sets[5:]...)...)
	reordered, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	spec, err := domain.Parse(reordered, "reordered.json")
	if err != nil {
		t.Fatal(err)
	}
	ids := PopulationIDs(spec, 5)
	if !slices.Contains(ids, "air_header") || !slices.Contains(ids, "tank-5") || slices.Contains(ids, "tank-6") {
		t.Fatalf("population = %v, want the header and five templated tanks", ids)
	}
}
