package app

import (
	"strings"
	"testing"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
)

func TestInjectingAFaultWhoseOnsetIsBeforeTheClockIsRefusedAsInvalid(t *testing.T) {
	t.Parallel()
	d := newTestDirector(t)
	worldID := createWorld(t, d)
	w := d.World(worldID)
	entity, fault := w.Run.World.EntityIDs()[0], w.Run.Domain().Spec.Kinds[0].Faults[0].ID
	if _, err := d.AdvanceBy(t.Context(), worldID, 600e9, false); err != nil {
		t.Fatal(err)
	}
	_, err := d.InjectFault(worldID, entity, fault, w.Run.World.Clock()-60e9, nil)
	if toolCode(err) != rules.CodeDomainInvalid || !strings.Contains(err.Error(), "fault onset is before the clock") {
		t.Fatalf("err = %v, want domain_invalid naming the onset", err)
	}
	if len(w.Run.World.ListFaults()) != 0 {
		t.Fatal("the refused fault was registered")
	}
}
