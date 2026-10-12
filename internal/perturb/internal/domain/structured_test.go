package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestOutOfRangeUsesTheRangeNarrowedForTheEntity(t *testing.T) {
	t.Parallel()
	spec, err := domain.Load(testsupport.StructuredDomain())
	if err != nil {
		t.Fatal(err)
	}
	for entity, want := range map[string]float64{"P-slow": 22, "pump-1": 30} {
		l := New("w", 4, spec)
		if _, err := l.Apply(OutOfRange, map[string]any{"rate": 1.0, "magnitude": 10}, 0, 0); err != nil {
			t.Fatal(err)
		}
		event := ev(1, "pump.flow", 5.0)
		event.EntityID = entity
		recs := l.Process(event, 1000000000)
		if got, ok := recs[0].Event.Value.(float64); len(recs) != 1 || !ok || got != want {
			t.Fatalf("%s: value %v, want %v (declared range maximum plus the magnitude)", entity, recs[0].Event.Value, want)
		}
	}
}
