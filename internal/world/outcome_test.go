package world_test

import (
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
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

func TestTheFacadeJudgesConditionsAndBoundariesAsTheOutcomeRecords(t *testing.T) {
	t.Parallel()
	spec, err := domain.Load(testsupport.BeltConveyorDomain("b"))
	if err != nil {
		t.Fatal(err)
	}
	w, err := world.New(spec, 1, "w-belt", model.DefaultStartTimeNS, world.Options{Noiseless: true, EmitDisabled: true, RecordOutcome: true})
	if err != nil {
		t.Fatal(err)
	}
	at := func(s int64) int64 { return model.DefaultStartTimeNS + s*1e9 }
	if _, err := w.InjectFault("D-3", "conveyor_drive__vfd_fault", at(30), nil); err != nil {
		t.Fatal(err)
	}
	unavailable := model.Condition{State: "available", Operator: "lt", Threshold: 0.5}
	if w.ConditionHolds("D-3", unavailable, at(29)) || !w.ConditionHolds("D-3", unavailable, at(31)) || w.ConditionHolds("D-1", unavailable, at(31)) {
		t.Fatal("D-3 should become unavailable at its fault and D-1 stay available")
	}
	if _, _, err := w.Advance(at(600)); err != nil {
		t.Fatal(err)
	}
	var belt model.EntityOutcome
	for _, entity := range w.OutcomeFacts(at(600)).Entities {
		if entity.EntityID == "B-1" {
			belt = entity
		}
	}
	if len(belt.Boundaries) != 1 || !slices.Contains(w.ViolatedBoundaries("B-1", at(600)), belt.Boundaries[0].Name) {
		t.Fatalf("recorded boundaries %+v, violated now %v: want the stranded belt in both", belt.Boundaries, w.ViolatedBoundaries("B-1", at(600)))
	}
}
