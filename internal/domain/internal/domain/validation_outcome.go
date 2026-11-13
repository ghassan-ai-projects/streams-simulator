package domain

import (
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (checks domainChecks) checkOutcomeWeights() error {
	weights := checks.compiled.Spec.GroundTruth.OutcomeWeights
	if weights == nil {
		return nil
	}
	at := checks.within("/ground_truth/outcome_weights")
	for _, check := range []func(*model.OutcomeWeights) error{at.boundaryWeights, at.tripWeights, at.kpiWeights, at.callWeights} {
		if err := check(weights); err != nil {
			return err
		}
	}
	return nil
}

func (checks domainChecks) boundaryWeights(weights *model.OutcomeWeights) error {
	items := weights.Boundaries
	return checks.each("boundaries", len(items), func(at domainChecks, i int) error { return at.boundaryWeight(items[:i], items[i]) })
}

func (checks domainChecks) boundaryWeight(earlier []model.BoundaryWeight, w model.BoundaryWeight) error {
	kind, err := checks.weightedKind(w.Kind)
	if err != nil {
		return err
	}
	if !slices.ContainsFunc(kind.Boundaries, func(b model.Boundary) bool { return b.Name == w.Name }) {
		return checks.bad("/name", "kind %q declares no boundary %q", kind.Name, w.Name)
	}
	if slices.ContainsFunc(earlier, func(e model.BoundaryWeight) bool { return e.Kind == w.Kind && e.Name == w.Name }) {
		return checks.bad("/name", "boundary %q of kind %q is weighted twice", w.Name, kind.Name)
	}
	return nil
}

func (checks domainChecks) tripWeights(weights *model.OutcomeWeights) error {
	items := weights.Trips
	return checks.each("trips", len(items), func(at domainChecks, i int) error { return at.tripWeight(items[:i], items[i]) })
}

func (checks domainChecks) tripWeight(earlier []model.TripWeight, w model.TripWeight) error {
	kind, err := checks.weightedKind(w.Kind)
	if err != nil {
		return err
	}
	if !isTripState(kind, w.State) {
		return checks.bad("/state", "%q is not a trip state of kind %q", w.State, kind.Name)
	}
	if slices.ContainsFunc(earlier, func(e model.TripWeight) bool { return e.Kind == w.Kind && e.State == w.State }) {
		return checks.bad("/state", "trip %q of kind %q is weighted twice", w.State, kind.Name)
	}
	return nil
}

func (checks domainChecks) kpiWeights(weights *model.OutcomeWeights) error {
	items := weights.KPIs
	return checks.each("kpis", len(items), func(at domainChecks, i int) error { return at.kpiWeight(items[:i], items[i]) })
}

func (checks domainChecks) kpiWeight(earlier []model.KPIWeight, w model.KPIWeight) error {
	kind, err := checks.weightedKind(w.Kind)
	if err != nil {
		return err
	}
	if !slices.Contains(kind.KPIs, w.State) {
		return checks.bad("/state", "%q is not a KPI of kind %q", w.State, kind.Name)
	}
	if slices.ContainsFunc(earlier, func(e model.KPIWeight) bool { return e.Kind == w.Kind && e.State == w.State }) {
		return checks.bad("/state", "KPI %q of kind %q is weighted twice", w.State, kind.Name)
	}
	return nil
}

func (checks domainChecks) callWeights(weights *model.OutcomeWeights) error {
	items := weights.Calls
	return checks.each("calls", len(items), func(at domainChecks, i int) error { return at.callWeight(items[:i], items[i]) })
}

func (checks domainChecks) callWeight(earlier []model.CallWeight, w model.CallWeight) error {
	if !checks.compiled.HasEffector(w.Effector) {
		return checks.bad("/effector", "the domain declares no effector %q", w.Effector)
	}
	if slices.ContainsFunc(earlier, func(e model.CallWeight) bool { return e.Effector == w.Effector }) {
		return checks.bad("/effector", "calls of %q are weighted twice", w.Effector)
	}
	return nil
}

func (checks domainChecks) weightedKind(name string) (*model.Kind, error) {
	kind := checks.compiled.Kind(name)
	if kind == nil {
		return nil, checks.bad("/kind", "the domain declares no kind %q", name)
	}
	return kind, nil
}

func isTripState(kind *model.Kind, state string) bool {
	dyn := kind.DynamicsOf(state)
	return dyn != nil && dyn.F1 != nil && dyn.F1.Form == model.FormTrip
}
