package domain

import (
	"math"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// An effect lands 1 ms after its call, a trip samples whole seconds and a step's value shows from the step's end.
func outcomeSpec(latch bool) func(*model.DomainSpec) {
	return func(s *model.DomainSpec) {
		kind := &s.Kinds[0]
		kind.State = append(kind.State, model.State{Name: "temp", Initial: 0}, model.State{Name: "flow", Initial: 100},
			model.State{Name: "resets", Initial: 0}, model.State{Name: "low_flow", Initial: 0}, model.State{Name: "hot", Initial: 0})
		kind.Dynamics = append(kind.Dynamics,
			model.Dynamics{Target: "temp", Tier: "F0", F0: &model.F0Dyn{Baseline: 50, TrendPerHour: 36}},
			model.Dynamics{Target: "hot", Tier: "F1", DTMs: 1000, F1: &model.F1Dyn{Form: model.FormTrip, When: &model.Condition{State: "temp", Operator: "gt", Threshold: 60}, ForS: 30, TrippedValue: 1}},
			model.Dynamics{Target: "low_flow", Tier: "F1", DTMs: 1000, F1: &model.F1Dyn{Form: model.FormTrip, When: &model.Condition{State: "flow", Operator: "lt", Threshold: 50}, ForS: 10, TrippedValue: 1, Latch: latch, ResetOn: "resets"}})
		kind.Boundaries = []model.Boundary{{Name: "oil_hot", State: "hot", Severity: "critical"}, {Name: "dry_run", State: "low_flow", Severity: "major"}}
		kind.KPIs = []string{"temp", "flow"}
		kind.Effectors = []model.Effector{protectionEffector("cut", "flow", -80), protectionEffector("restore", "flow", 80), protectionEffector("reset", "resets", 1)}
		s.Instances[0].Count.Default = 3
	}
}

func outcomeWorld(t *testing.T, latch, record bool) *World {
	t.Helper()
	w, err := New(testSpec(t, outcomeSpec(latch)), 42, "w-outcome", model.DefaultStartTimeNS, Options{EmitDisabled: true, RecordOutcome: record})
	if err != nil {
		t.Fatal(err)
	}
	return w
}

func entityFacts(t *testing.T, w *World, id string, seconds float64) model.EntityOutcome {
	t.Helper()
	facts := w.OutcomeFacts(at(seconds))
	for _, entity := range facts.Entities {
		if entity.EntityID == id {
			return entity
		}
	}
	t.Fatalf("no facts for %s in %+v", id, facts)
	return model.EntityOutcome{}
}

func tripOf(entity model.EntityOutcome, state string) model.TripOutcome {
	for _, trip := range entity.Trips {
		if trip.State == state {
			return trip
		}
	}
	return model.TripOutcome{}
}

func boundaryOf(entity model.EntityOutcome, name string) model.BoundaryOutcome {
	for _, boundary := range entity.Boundaries {
		if boundary.Name == name {
			return boundary
		}
	}
	return model.BoundaryOutcome{}
}

func TestTheFirstTripTimeIsTheClosedFormCrossingPlusTheHold(t *testing.T) {
	t.Parallel()
	w := outcomeWorld(t, false, true)
	crossing := (60.0 - 50) / (36.0 / 3600)
	visible := math.Floor(crossing) + 1 + 30
	e1 := entityFacts(t, w, "e-1", 3600)
	if trip := tripOf(e1, "hot"); trip.Count != 1 || trip.FirstNS != at(visible) {
		t.Fatalf("hot trip %+v, want one entry at %v s (analytic crossing %v s + 30 s hold)", trip, visible, crossing)
	}
	if b := boundaryOf(e1, "oil_hot"); b.Violations != 1 || b.FirstNS != at(visible) || b.DurationS != 3600-visible || b.Severity != "critical" {
		t.Fatalf("oil_hot %+v, want one critical violation from %v s lasting to the query", b, visible)
	}
	if before := entityFacts(t, w, "e-1", visible-1); len(before.Trips) != 0 || len(before.Boundaries) != 0 {
		t.Fatalf("facts a second before the trip = %+v, want no trip and no violation", before)
	}
}

func TestAnUnlatchedTripThatTogglesCountsEachEntryAndItsExactDuration(t *testing.T) {
	t.Parallel()
	w := outcomeWorld(t, false, true)
	for _, call := range []struct {
		effector string
		seconds  float64
	}{{"cut", 100}, {"restore", 200}, {"cut", 400}, {"restore", 450}} {
		invoke(t, w, call.effector, call.seconds)
	}
	entries, exits := []float64{100 + 1 + 10, 400 + 1 + 10}, []float64{200 + 1, 450 + 1}
	for _, tc := range []struct {
		query    float64
		count    int
		duration float64
	}{{150, 1, 150 - entries[0]}, {300, 1, exits[0] - entries[0]}, {430, 2, exits[0] - entries[0] + 430 - entries[1]}, {1000, 2, exits[0] - entries[0] + exits[1] - entries[1]}} {
		e1 := entityFacts(t, w, "e-1", tc.query)
		trip, dry := tripOf(e1, "low_flow"), boundaryOf(e1, "dry_run")
		if trip.Count != tc.count || trip.FirstNS != at(entries[0]) || dry.Violations != tc.count || dry.DurationS != tc.duration {
			t.Fatalf("at %v s: trip %+v, dry_run %+v; want %d entries from %v s lasting %v s", tc.query, trip, dry, tc.count, entries[0], tc.duration)
		}
	}
}

func TestALatchedTripCountsOnceUntilItsResetAndAgainAfter(t *testing.T) {
	t.Parallel()
	w := outcomeWorld(t, true, true)
	invoke(t, w, "cut", 100)
	invoke(t, w, "restore", 200)
	invoke(t, w, "cut", 250)
	invoke(t, w, "restore", 260)
	if trip := tripOf(entityFacts(t, w, "e-1", 290), "low_flow"); trip.Count != 1 {
		t.Fatalf("latched trip before its reset %+v, want one entry however often the condition toggles", trip)
	}
	invoke(t, w, "reset", 300)
	if w.StateValue("e-1", "low_flow", at(301)) != 1 || w.StateValue("e-1", "low_flow", at(302)) != 0 {
		t.Fatal("the latch did not clear at 302 s, which the oracle assumes")
	}
	invoke(t, w, "cut", 500)
	e1 := entityFacts(t, w, "e-1", 600)
	latched, again := 302-111.0, 600-511.0
	if trip, dry := tripOf(e1, "low_flow"), boundaryOf(e1, "dry_run"); trip.Count != 2 || trip.FirstNS != at(111) || dry.DurationS != latched+again {
		t.Fatalf("trip %+v, dry_run %+v; want two entries (111 s and 511 s) lasting %v s", trip, dry, latched+again)
	}
}

func TestKPIsStartAtTheWorldStartAndEndAtTheQuery(t *testing.T) {
	t.Parallel()
	w := outcomeWorld(t, false, true)
	invoke(t, w, "cut", 100)
	e1 := entityFacts(t, w, "e-1", 900)
	if len(e1.KPIs) != 2 || e1.KPIs[0].State != "flow" || e1.KPIs[1].State != "temp" {
		t.Fatalf("kpis %+v, want flow and temp in name order", e1.KPIs)
	}
	for _, kpi := range e1.KPIs {
		if start, end := w.StateValue("e-1", kpi.State, at(0)), w.StateValue("e-1", kpi.State, at(900)); kpi.Start != start || kpi.End != end {
			t.Fatalf("%s: %+v, want start %v and end %v", kpi.State, kpi, start, end)
		}
	}
	if e1.KPIs[0].Start != 100 || e1.KPIs[0].End != 20 || e1.KPIs[1].Start != 50 || e1.KPIs[1].End != 50+36*900.0/3600 {
		t.Fatalf("kpis %+v, want flow 100 -> 20 and temp 50 -> 59", e1.KPIs)
	}
}
