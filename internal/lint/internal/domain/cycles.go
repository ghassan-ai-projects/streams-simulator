package domain

import (
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type source struct {
	kind  *model.Kind
	state string
}

func cyclic(spec *domain.Compiled, kind *model.Kind, state string) bool {
	return reachesCycle(spec, source{kind, state}, map[source]bool{})
}

func reachesCycle(spec *domain.Compiled, from source, seen map[source]bool) bool {
	if seen[from] {
		return false
	}
	seen[from] = true
	if periodic(from.kind.DynamicsOf(from.state)) {
		return true
	}
	return slices.ContainsFunc(drivers(spec, from), func(driver source) bool { return reachesCycle(spec, driver, seen) })
}

func periodic(dyn *model.Dynamics) bool {
	return dyn != nil && dyn.F0 != nil && (len(dyn.F0.Seasonality) > 0 || dyn.F0.Schedule != nil)
}

func drivers(spec *domain.Compiled, from source) []source {
	return slices.Concat(inputDrivers(spec, from), conditionDrivers(from), followDriver(spec, from), flowDrivers(from))
}

func conditionDrivers(from source) []source {
	dyn := from.kind.DynamicsOf(from.state)
	if dyn == nil || dyn.F1 == nil {
		return nil
	}
	var found []source
	for _, state := range conditionStates(dyn.F1) {
		if state != "" && state != from.state {
			found = append(found, source{from.kind, state})
		}
	}
	return found
}

func conditionStates(f1 *model.F1Dyn) []string {
	states := []string{f1.Selector, f1.Watch}
	for _, condition := range []*model.Condition{f1.When, f1.StageUp, f1.StageDown} {
		if condition != nil {
			states = append(states, condition.State, condition.ThresholdState)
		}
	}
	return states
}

func inputDrivers(spec *domain.Compiled, from source) []source {
	dyn := from.kind.DynamicsOf(from.state)
	if dyn == nil || dyn.F1 == nil {
		return nil
	}
	var found []source
	for _, input := range dyn.F1.Inputs {
		if owner := inputOwner(spec, from.kind, input); owner != nil {
			found = append(found, source{owner, input.State})
		}
	}
	return found
}

func inputOwner(spec *domain.Compiled, reader *model.Kind, input model.F1Input) *model.Kind {
	if input.Link == nil {
		return reader
	}
	return providerOf(spec, input.Link.Type)
}

func followDriver(spec *domain.Compiled, from source) []source {
	dyn := from.kind.DynamicsOf(from.state)
	if dyn == nil || dyn.Follow == nil {
		return nil
	}
	var found []source
	for _, followed := range dyn.Follow.Sources {
		if provider := providerOf(spec, followed.Link); provider != nil {
			found = append(found, source{provider, followed.State})
		}
	}
	return found
}

func providerOf(spec *domain.Compiled, linkType string) *model.Kind {
	for _, candidate := range spec.Spec.LinkTypes {
		if candidate.Name == linkType {
			return spec.Kind(candidate.From)
		}
	}
	return nil
}

func flowDrivers(from source) []source {
	var found []source
	for _, flow := range from.kind.Flows {
		if flow.From != from.state && flow.To != from.state {
			continue
		}
		for _, state := range slices.Concat([]string{flow.From, flow.To}, flow.Scale) {
			if state != "" && state != from.state {
				found = append(found, source{from.kind, state})
			}
		}
	}
	return found
}
