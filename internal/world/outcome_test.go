package world_test

import (
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func TestOutcomeFactsCoverEveryEntityInIDOrderOnlyWhenRecorded(t *testing.T) {
	t.Parallel()
	if newPondWorld(t, world.Options{}).OutcomeFacts(model.DefaultStartTimeNS) != nil {
		t.Fatal("a world that does not record its outcome returned facts")
	}
	w := newPondWorld(t, world.Options{RecordOutcome: true})
	end := model.DefaultStartTimeNS + 3600e9
	if _, _, err := w.Advance(end); err != nil {
		t.Fatal(err)
	}
	facts := w.OutcomeFacts(end)
	ids := make([]string, len(facts.Entities))
	for i, entity := range facts.Entities {
		ids[i] = entity.EntityID
		if entity.Kind != w.Entity(entity.EntityID).Type {
			t.Fatalf("%s: kind %q, want %q", entity.EntityID, entity.Kind, w.Entity(entity.EntityID).Type)
		}
	}
	if want := slices.Sorted(slices.Values(w.EntityIDs())); !slices.Equal(ids, want) || facts.HorizonNS != end || len(facts.Calls) != 0 {
		t.Fatalf("facts for %v at %d with calls %v, want %v at %d and no calls", ids, facts.HorizonNS, facts.Calls, want, end)
	}
}
