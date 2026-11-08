package domain

import (
	"cmp"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func runCost(weights *model.OutcomeWeights, facts *model.OutcomeFacts, unneededCalls bool) *model.OutcomeCost {
	var terms []model.CostTerm
	for _, entity := range facts.Entities {
		terms = append(terms, entityTerms(weights, entity)...)
	}
	if unneededCalls {
		terms = append(terms, callTerms(weights, facts.Calls)...)
	}
	slices.SortFunc(terms, compareTerms)
	cost := &model.OutcomeCost{Terms: []model.CostTerm{}}
	for _, term := range terms {
		cost.J, cost.Terms = cost.J+term.Cost, append(cost.Terms, term)
	}
	return cost
}

func entityTerms(weights *model.OutcomeWeights, entity model.EntityOutcome) []model.CostTerm {
	terms := boundaryTerms(weights, entity)
	terms = append(terms, tripTerms(weights, entity)...)
	return append(terms, kpiTerms(weights, entity)...)
}

func compareTerms(a, b model.CostTerm) int {
	return cmp.Or(cmp.Compare(a.Term, b.Term), cmp.Compare(a.Kind, b.Kind), cmp.Compare(a.Name, b.Name), cmp.Compare(a.EntityID, b.EntityID))
}

func boundaryTerms(weights *model.OutcomeWeights, entity model.EntityOutcome) []model.CostTerm {
	var terms []model.CostTerm
	for _, b := range entity.Boundaries {
		at := slices.IndexFunc(weights.Boundaries, func(w model.BoundaryWeight) bool { return w.Kind == entity.Kind && w.Name == b.Name })
		if at < 0 {
			continue
		}
		w := weights.Boundaries[at]
		terms = appendTerm(terms, model.TermBoundaryViolations, entity, b.Name, float64(b.Violations), w.PerViolation)
		terms = appendTerm(terms, model.TermBoundaryHours, entity, b.Name, b.DurationS/3600, w.PerHour)
	}
	return terms
}

func tripTerms(weights *model.OutcomeWeights, entity model.EntityOutcome) []model.CostTerm {
	var terms []model.CostTerm
	for _, trip := range entity.Trips {
		at := slices.IndexFunc(weights.Trips, func(w model.TripWeight) bool { return w.Kind == entity.Kind && w.State == trip.State })
		if at >= 0 {
			terms = appendTerm(terms, model.TermTrips, entity, trip.State, float64(trip.Count), weights.Trips[at].PerTrip)
		}
	}
	return terms
}

func kpiTerms(weights *model.OutcomeWeights, entity model.EntityOutcome) []model.CostTerm {
	var terms []model.CostTerm
	for _, kpi := range entity.KPIs {
		at := slices.IndexFunc(weights.KPIs, func(w model.KPIWeight) bool { return w.Kind == entity.Kind && w.State == kpi.State })
		if at >= 0 {
			terms = appendTerm(terms, model.TermKPIs, entity, kpi.State, kpi.End-kpi.Start, weights.KPIs[at].PerUnit)
		}
	}
	return terms
}

func callTerms(weights *model.OutcomeWeights, calls []model.OutcomeCall) []model.CostTerm {
	var terms []model.CostTerm
	for _, call := range calls {
		at := slices.IndexFunc(weights.Calls, func(w model.CallWeight) bool { return w.Effector == call.Effector })
		if call.Accepted && at >= 0 {
			terms = addCall(terms, call, weights.Calls[at].PerCall)
		}
	}
	return terms
}

func addCall(terms []model.CostTerm, call model.OutcomeCall, perCall float64) []model.CostTerm {
	at := slices.IndexFunc(terms, func(t model.CostTerm) bool { return t.Name == call.Effector && t.EntityID == call.EntityID })
	if at < 0 {
		return append(terms, model.CostTerm{Term: model.TermCalls, Name: call.Effector, EntityID: call.EntityID, Quantity: 1, Cost: perCall})
	}
	terms[at].Quantity, terms[at].Cost = terms[at].Quantity+1, terms[at].Cost+perCall
	return terms
}

func appendTerm(terms []model.CostTerm, term string, entity model.EntityOutcome, name string, quantity, weight float64) []model.CostTerm {
	if quantity == 0 || weight == 0 {
		return terms
	}
	return append(terms, model.CostTerm{Term: term, Kind: entity.Kind, Name: name, EntityID: entity.EntityID, Quantity: quantity, Cost: quantity * weight})
}
