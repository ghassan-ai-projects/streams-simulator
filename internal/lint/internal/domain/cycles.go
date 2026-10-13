package domain

import (
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func cyclic(kind *model.Kind, state string, seen map[string]bool) bool {
	dyn := kind.DynamicsOf(state)
	if dyn == nil || seen[state] {
		return false
	}
	seen[state] = true
	if dyn.F0 != nil && (len(dyn.F0.Seasonality) > 0 || dyn.F0.Schedule != nil) {
		return true
	}
	return dyn.F1 != nil && slices.ContainsFunc(dyn.F1.Inputs, func(in model.F1Input) bool { return cyclic(kind, in.State, seen) })
}
