package domain

import (
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestEveryGeneratedScenarioInjectsAFaultIntoAnEntityOfItsKind(t *testing.T) {
	t.Parallel()
	testsupport.SkipUnderRace(t, "generates and audits whole suites on one goroutine")
	spec, err := domain.Load(testsupport.TwoKindDomain())
	if err != nil {
		t.Fatal(err)
	}
	s, err := Generate(Config{Domain: spec, Profile: "nominal", N: 6, Seed: 5, SampleNS: 120 * 1e9, MaxAttempts: 60})
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Scenarios) == 0 {
		t.Fatalf("no scenarios generated: %s", s.TerminalState)
	}
	for _, sc := range s.Scenarios {
		kind := spec.FaultKind(sc.Fault)
		if kind == nil || !strings.HasPrefix(sc.EntityID, kind.Name+"-") {
			t.Fatalf("scenario %s injects %q into %q, which is not of kind %v", sc.ID, sc.Fault, sc.EntityID, kind)
		}
	}
}
