package domain

import (
	"encoding/json"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
)

func TestBoundedGenerationRegression(t *testing.T) {
	t.Parallel()
	spec := loadSpec(t)
	generated, err := Generate(Config{Domain: spec, Profile: "nominal", N: 1, Seed: 91, SampleNS: 300 * 1e9, MaxAttempts: 2})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(generated)
	if err != nil {
		t.Fatal(err)
	}
	const want = "sha256:9247a8f29012c73488ca2e17b0376bdd80cd2b898a4b866f3dedbe7a55905fa1"
	if got := canonical.DigestBytes(raw); got != want {
		t.Fatalf("generation recipe changed: %s", got)
	}
	if generated.Attempts > 2 || len(generated.Scenarios) != len(generated.Labels) {
		t.Fatalf("invalid bounded generation: %+v", generated)
	}
}
