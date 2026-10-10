package app

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func worldDigest(r *Run) string {
	digest, err := canonical.Digest(worldIdentity(r, initialEntityValues(r)))
	if err != nil {
		// Values are canonical-supported types; surface violations rather than empty identity.
		return "invalid-world-digest:" + err.Error()
	}
	return digest
}

func initialEntityValues(r *Run) []any {
	entities := make([]any, 0, len(r.World.InitialEntityIDs()))
	for _, id := range r.World.InitialEntityIDs() {
		entities = append(entities, id)
	}
	return entities
}

func worldIdentity(r *Run, entities []any) map[string]any {
	return map[string]any{
		"sim_version":        model.SimVersion,
		"domain":             r.Config.Domain.Digest,
		"seed":               r.Config.Seed,
		"start_time_ns":      r.worldStartTimeNS,
		"entity_ids":         entities,
		"scenario_profile":   r.Config.ScenarioProfile,
		"clock_multiplier":   r.Config.ClockMultiplier,
		"time_mode":          r.Config.TimeMode,
		"noiseless":          r.Config.Noiseless,
		"force_failure_mode": string(r.Config.ForceFailureMode),
	}
}
