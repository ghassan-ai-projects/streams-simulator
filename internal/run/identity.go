package run

import (
	"encoding/json"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func adapterDigest(a *model.Adapter) string {
	raw, _ := json.Marshal(a)
	return canonical.DigestBytes(raw)
}

func worldDigest(r *Run) string {
	entityIDs := make([]any, 0, len(r.World.InitialEntityIDs()))
	for _, id := range r.World.InitialEntityIDs() {
		entityIDs = append(entityIDs, id)
	}
	raw := map[string]any{
		"sim_version":        model.SimVersion,
		"domain":             r.Config.Domain.Digest,
		"seed":               r.Config.Seed,
		"start_time_ns":      r.worldStartTimeNS,
		"entity_ids":         entityIDs,
		"scenario_profile":   r.Config.ScenarioProfile,
		"clock_multiplier":   r.Config.ClockMultiplier,
		"time_mode":          r.Config.TimeMode,
		"noiseless":          r.Config.Noiseless,
		"force_failure_mode": r.Config.ForceFailureMode,
	}
	d, err := canonical.Digest(raw)
	if err != nil {
		// Every value above is one of canonical's explicitly supported types.
		// Keep the failure visible if that invariant changes rather than
		// silently emitting an empty identity.
		return "invalid-world-digest:" + err.Error()
	}
	return d
}
