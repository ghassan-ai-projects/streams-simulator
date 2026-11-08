package domain

import (
	"reflect"
	"slices"
	"testing"
)

func togglingWorld(t *testing.T, record bool) *World {
	t.Helper()
	w := outcomeWorld(t, false, record)
	for _, call := range []struct {
		effector string
		seconds  float64
	}{{"cut", 100}, {"restore", 200}, {"cut", 400}, {"restore", 450}} {
		invoke(t, w, call.effector, call.seconds)
	}
	return w
}

func TestViolatedBoundariesAndTheRecorderAgreeAtEverySecond(t *testing.T) {
	t.Parallel()
	w := togglingWorld(t, true)
	inside := func(s float64) bool { return (s >= 111 && s < 201) || (s >= 411 && s < 451) }
	previous := 0.0
	for s := 0.0; s <= 600; s++ {
		listed := slices.Contains(w.ViolatedBoundaries("e-1", at(s)), "dry_run")
		duration := boundaryOf(entityFacts(t, w, "e-1", s), "dry_run").DurationS
		if listed != inside(s) || (s > 0 && (duration > previous) != inside(s-1)) {
			t.Fatalf("at %v s: listed %v, recorded duration %v after %v; want both to say violated = %v", s, listed, duration, previous, inside(s))
		}
		previous = duration
	}
}

func TestReadingAheadOfTheClockRecordsNothing(t *testing.T) {
	t.Parallel()
	w := outcomeWorld(t, false, true)
	if got := w.ViolatedBoundaries("e-1", at(5000)); !slices.Contains(got, "oil_hot") {
		t.Fatalf("violated at 5000 s = %v, want oil_hot", got)
	}
	invoke(t, w, "cut", 100)
	if e1 := entityFacts(t, w, "e-1", 900); len(e1.Boundaries) != 1 || e1.Boundaries[0].Name != "dry_run" || tripOf(e1, "hot").Count != 0 {
		t.Fatalf("facts at 900 s = %+v, want only dry_run: the look ahead to 5000 s left no trace", e1)
	}
}

func TestRecordingChangesNoValueAndIsOffByDefault(t *testing.T) {
	t.Parallel()
	recorded, plain := togglingWorld(t, true), togglingWorld(t, false)
	if plain.OutcomeFacts(at(600)) != nil {
		t.Fatal("a world without RecordOutcome returned outcome facts")
	}
	sameValues(t, recorded, plain, 50, 111, 300, 451, 600)
	facts := recorded.OutcomeFacts(at(600))
	sameValues(t, recorded, plain, 600, 1100, 1200)
	if !reflect.DeepEqual(recorded.EffectorCalls(), plain.EffectorCalls()) {
		t.Fatal("recording changed the effector calls")
	}
	ids := []string{facts.Entities[0].EntityID, facts.Entities[1].EntityID, facts.Entities[2].EntityID}
	if !slices.Equal(ids, []string{"e-1", "e-2", "e-3"}) || facts.HorizonNS != at(600) || facts.Calls == nil || len(facts.Calls) != 0 {
		t.Fatalf("facts = %+v, want three entities by id, the query time and no calls", facts)
	}
	if e2 := facts.Entities[1]; e2.Kind != "test" || len(e2.Trips) != 0 || len(e2.Boundaries) != 0 || len(e2.KPIs) != 2 {
		t.Fatalf("an untouched entity = %+v, want its KPIs only", e2)
	}
}

func sameValues(t *testing.T, recorded, plain *World, seconds ...float64) {
	t.Helper()
	for _, s := range seconds {
		for _, state := range []string{"low_flow", "hot", "flow", "temp"} {
			if got, want := recorded.StateValue("e-1", state, at(s)), plain.StateValue("e-1", state, at(s)); got != want {
				t.Fatalf("%s at %v s: %v recorded, %v not", state, s, got, want)
			}
		}
	}
}
