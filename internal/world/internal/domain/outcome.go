package domain

import (
	"cmp"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type outcomeRecorder struct {
	open   bool
	starts map[string][]model.KPIOutcome
}

func newOutcomeRecorder(enabled bool) *outcomeRecorder {
	if !enabled {
		return nil
	}
	return &outcomeRecorder{starts: map[string][]model.KPIOutcome{}}
}

func (w *World) openOutcome(ids []string, startNS int64) {
	if w.outcome == nil {
		return
	}
	w.outcome.open = true
	for _, id := range ids {
		w.noteOutcomeStart(w.entities[id], startNS)
	}
}

func (w *World) noteOutcomeStart(ent *Entity, atNS int64) {
	if w.outcome == nil || !w.outcome.open {
		return
	}
	var kpis []model.KPIOutcome
	for _, state := range slices.Sorted(slices.Values(ent.kind.KPIs)) {
		kpis = append(kpis, model.KPIOutcome{State: state, Start: w.stateOf(ent, state, atNS)})
	}
	w.outcome.starts[ent.ID] = kpis
}

func (w *World) OutcomeFacts(atNS int64) *model.OutcomeFacts {
	if w.outcome == nil {
		return nil
	}
	facts := &model.OutcomeFacts{HorizonNS: atNS, Entities: []model.EntityOutcome{}, Calls: []model.OutcomeCall{}}
	for _, id := range slices.Sorted(slices.Values(w.entityOrder)) {
		if ent := w.entities[id]; ent.BornNS <= atNS {
			facts.Entities = append(facts.Entities, w.entityOutcome(ent, ent.lastAliveAt(atNS)))
		}
	}
	return facts
}

func (w *World) entityOutcome(ent *Entity, atNS int64) model.EntityOutcome {
	w.advanceEntity(ent, atNS)
	return model.EntityOutcome{EntityID: ent.ID, Kind: ent.kind.Name, Trips: ent.tripOutcomes(atNS),
		Boundaries: ent.boundaryOutcomes(atNS), KPIs: w.kpiOutcomes(ent, atNS)}
}

func (ent *Entity) tripOutcomes(atNS int64) []model.TripOutcome {
	var trips []model.TripOutcome
	for _, state := range tripStates(ent.kind) {
		if span := ent.states[state].tripSpanAt(atNS); span.entries > 0 {
			trips = append(trips, model.TripOutcome{State: state, FirstNS: span.firstNS, Count: span.entries})
		}
	}
	return trips
}

func (ent *Entity) boundaryOutcomes(atNS int64) []model.BoundaryOutcome {
	boundaries := slices.SortedFunc(slices.Values(ent.kind.Boundaries), func(a, b model.Boundary) int { return cmp.Compare(a.Name, b.Name) })
	var violated []model.BoundaryOutcome
	for _, b := range boundaries {
		if span := ent.states[b.State].tripSpanAt(atNS); span.entries > 0 {
			violated = append(violated, model.BoundaryOutcome{Name: b.Name, Severity: b.Severity, Violations: span.entries,
				FirstNS: span.firstNS, DurationS: float64(span.trippedNS) / secondsPerNS})
		}
	}
	return violated
}

func (w *World) kpiOutcomes(ent *Entity, atNS int64) []model.KPIOutcome {
	kpis := slices.Clone(w.outcome.starts[ent.ID])
	for i := range kpis {
		kpis[i].End = w.stateOf(ent, kpis[i].State, atNS)
	}
	return kpis
}

func tripStates(kind *model.Kind) []string {
	var states []string
	for i := range kind.Dynamics {
		if dyn := &kind.Dynamics[i]; dyn.F1 != nil && dyn.F1.Form == model.FormTrip {
			states = append(states, dyn.Target)
		}
	}
	slices.Sort(states)
	return states
}
