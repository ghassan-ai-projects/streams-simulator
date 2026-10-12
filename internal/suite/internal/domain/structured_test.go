package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestEveryGeneratedScenarioInjectsAFaultIntoAnEntityThatCarriesIt(t *testing.T) {
	t.Parallel()
	testsupport.SkipUnderRace(t, "generates and audits whole suites on one goroutine")
	spec, err := domain.Load(testsupport.StructuredDomain())
	if err != nil {
		t.Fatal(err)
	}
	s, err := Generate(Config{Domain: spec, Profile: "nominal", N: 8, Seed: 5, SampleNS: 120 * 1e9, MaxAttempts: 80})
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Scenarios) == 0 {
		t.Fatalf("no scenarios generated: %s", s.TerminalState)
	}
	for _, sc := range s.Scenarios {
		if kindOfEntity(spec, sc.EntityID).FaultDecl(sc.Fault) == nil {
			t.Fatalf("scenario %s injects %q into %q, whose kind does not declare it", sc.ID, sc.Fault, sc.EntityID)
		}
	}
}
