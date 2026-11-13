package domain

import (
	"maps"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func kpiUnits(kpis *model.KPITruth, artifact *model.RunArtifact) map[string]string {
	spec, ok := embeddedDomain(artifact)
	if kpis == nil || !ok {
		return nil
	}
	units := map[string]string{}
	for _, kind := range spec.Kinds {
		for _, state := range kpiStates(kind) {
			noteUnit(units, state)
		}
	}
	return declaredUnits(units)
}

func declaredUnits(units map[string]string) map[string]string {
	maps.DeleteFunc(units, func(_, unit string) bool { return unit == "" })
	if len(units) == 0 {
		return nil
	}
	return units
}

// A state two kinds give different units keeps none.
func noteUnit(units map[string]string, state model.State) {
	if known, seen := units[state.Name]; seen && known != state.Unit {
		units[state.Name] = ""
		return
	}
	units[state.Name] = state.Unit
}

func kpiStates(kind model.Kind) []model.State {
	var states []model.State
	for _, state := range kind.State {
		if slices.Contains(kind.KPIs, state.Name) {
			states = append(states, state)
		}
	}
	return states
}
