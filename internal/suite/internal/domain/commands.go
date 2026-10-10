package domain

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/perturb"
	"github.com/ghassan-ai-projects/streams-simulator/internal/randutil"
)

func substituteEntity(in map[string]any, entity string) map[string]any {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]any, len(in))
	// determinism-safe: rewrites values into a new map.
	for k, v := range in {
		if s, ok := v.(string); ok && s == "{entity_id}" {
			out[k] = entity
		} else {
			out[k] = v
		}
	}
	return out
}

// buildCommandLog renders the scenario as replayable commands.
func buildCommandLog(spec *domain.Compiled, sc *Scenario) []model.Command {
	var log scenarioCommandLog
	log.append(model.OpWorldCreate, map[string]any{"domain": spec.Spec.ID, "seed": sc.Seed, "start_time_ns": sc.StartNS, "time_mode": model.TimeStepped}, sc.StartNS)
	log.appendSetup(sc.Setup)
	log.append(model.OpFaultInject, map[string]any{"entity_id": sc.EntityID, "fault": sc.Fault, "onset_ns": sc.OnsetNS}, sc.OnsetNS)
	log.appendPerturbations(sc.Perturbations)
	end := sc.StartNS + sc.DurationNS
	log.append(model.OpClockAdvance, map[string]any{"to_ns": end, "await_consumer": false}, end)
	return log.commands
}

func pickPerturbations(rng *randutil.SplitMix64, counts map[string]int) []string {
	forced := uncovered(counts)
	if len(forced) > 0 {
		return forced[:min(2, len(forced))]
	}
	return drawCatalogPerturbations(rng)
}

func samplePerturbParams(rng *randutil.SplitMix64, name string) map[string]any {
	p := map[string]any{}
	if name == "delay_tail" {
		p["mean_s"] = round1(30 + rng.Float64()*300)
		p["sigma_s"] = round1(60 + rng.Float64()*300)
		return p
	}
	populatePerturbParams(p, rng, name)
	return p
}

func round3(f float64) float64 { return float64(int(f*1000+0.5)) / 1000 }

func round1(f float64) float64 { return float64(int(f*10+0.5)) / 10 }

func uncovered(counts map[string]int) []string {
	var out []string
	for _, name := range perturb.Names() {
		if counts[name] < 5 {
			out = append(out, name)
		}
	}
	return out
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func drawCatalogPerturbations(rng *randutil.SplitMix64) []string {
	var out []string
	names := perturb.Names()
	n := rng.Intn(3)
	for i := 0; i < n; i++ {
		out = append(out, names[rng.Intn(len(names))])
	}
	return out
}

func populatePerturbParams(p map[string]any, rng *randutil.SplitMix64, name string) {
	switch name {
	case "drop", "duplicate_burst", "id_reuse", "out_of_enum", "out_of_range":
		p["rate"] = round3(0.005 + rng.Float64()*0.04)
	case "storm":
		p["multiplier"] = 2 + rng.Intn(4)
	case "clock_skew":
		p["offset_s"] = round1(60 + rng.Float64()*300)
	case "oversize":
		p["bytes"] = 4096
	}
}
