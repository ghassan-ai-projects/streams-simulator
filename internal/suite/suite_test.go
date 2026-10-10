package suite_test

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/suite"
)

func shipped(t *testing.T, id string) *domain.Compiled {
	t.Helper()
	spec, err := domain.Load(filepath.Join("..", "..", "domains", id+".domain.json"))
	if err != nil {
		t.Fatal(err)
	}
	return spec
}

func TestGenerateRefusesAnUnknownProfile(t *testing.T) {
	t.Parallel()
	_, err := suite.Generate(suite.Config{Domain: shipped(t, "rotating-machinery"), Profile: "no_such_profile", N: 1, Seed: 1})
	if err == nil || !strings.Contains(err.Error(), `unknown profile "no_such_profile"`) {
		t.Fatalf("err = %v", err)
	}
}

func TestGenerateIsDeterministicAndLabelsEveryScenario(t *testing.T) {
	t.Parallel()
	cfg := suite.Config{Domain: shipped(t, "rotating-machinery"), Profile: "nominal", N: 2, Seed: 5}
	first, err := suite.Generate(cfg)
	if err != nil {
		t.Fatal(err)
	}
	second, err := suite.Generate(cfg)
	if err != nil {
		t.Fatal(err)
	}
	a, errA := json.Marshal(first)
	b, errB := json.Marshal(second)
	if errA != nil || errB != nil || string(a) != string(b) {
		t.Fatalf("the same configuration produced different suites (%v, %v)", errA, errB)
	}
	if first.DomainID != "rotating-machinery" || first.Profile != "nominal" || first.Seed != 5 {
		t.Fatalf("identity = %q %q %d", first.DomainID, first.Profile, first.Seed)
	}
	if len(first.Labels) != len(first.Scenarios) {
		t.Fatalf("%d labels for %d scenarios", len(first.Labels), len(first.Scenarios))
	}
}
