package domain

import (
	"maps"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func (b *entryBuild) expectActions() error {
	faults := b.allFaults()
	for _, id := range slices.Sorted(maps.Keys(b.entry.Expected)) {
		fault := faultNamed(faults, b.name+"__"+id)
		if fault == nil {
			return b.refuse(relative("expected_actions", id), "%s declares no fault %q", b.equipment.ID, id)
		}
		fault["expected_actions"] = qualifiedActions(b.entry.Expected[id])
	}
	return nil
}

func faultNamed(faults []any, id string) map[string]any {
	for _, fault := range faults {
		if document := mapOf(fault); document["id"] == id {
			return document
		}
	}
	return nil
}

func qualifiedActions(actions []model.FaultAction) []any {
	qualified := make([]any, len(actions))
	for i, action := range actions {
		qualified[i] = map[string]any{"kind": action.Kind, "effector": action.Kind + "__" + action.Effector}
	}
	return qualified
}
