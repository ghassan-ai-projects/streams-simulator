package suite

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
	var cmds []model.Command
	seq := 0
	appendCmd := func(op string, args map[string]any, atNS int64) {
		cmds = append(cmds, model.Command{Seq: int64(seq), AtNS: atNS, Op: op, Args: args})
		seq++
	}
	appendCmd(model.OpWorldCreate, map[string]any{
		"domain": spec.Spec.ID, "seed": sc.Seed, "start_time_ns": sc.StartNS,
		"time_mode": model.TimeStepped,
	}, sc.StartNS)
	for _, call := range sc.Setup {
		appendCmd(model.OpEffectorInvoke, map[string]any{
			"effector": call.Effector, "entity_id": call.EntityID,
			"command_id": call.CommandID, "args": call.Args, "at_ns": call.AtNS,
		}, call.AtNS)
	}
	appendCmd(model.OpFaultInject, map[string]any{
		"entity_id": sc.EntityID, "fault": sc.Fault, "onset_ns": sc.OnsetNS,
	}, sc.OnsetNS)
	for _, p := range sc.Perturbations {
		appendCmd(model.OpPerturbApply, map[string]any{
			"perturbation": p.Name, "params": p.Params,
			"from_ns": p.FromNS, "until_ns": p.UntilNS,
		}, p.FromNS)
	}
	appendCmd(model.OpClockAdvance, map[string]any{
		"to_ns": sc.StartNS + sc.DurationNS, "await_consumer": false,
	}, sc.StartNS+sc.DurationNS)
	return cmds
}

func pickPerturbations(rng *randutil.SplitMix64, counts map[string]int) []string {
	// Force coverage: perturbations below 5 scenarios ride along.
	var forced []string
	for _, name := range perturb.Names {
		if counts[name] < 5 {
			forced = append(forced, name)
		}
	}
	if len(forced) > 0 {
		return forced[:min(2, len(forced))]
	}
	var out []string
	n := rng.Intn(3) // 0..2
	names := perturb.Names
	for i := 0; i < n; i++ {
		out = append(out, names[rng.Intn(len(names))])
	}
	return out
}

func samplePerturbParams(rng *randutil.SplitMix64, name string) map[string]any {
	p := map[string]any{}
	switch name {
	case "drop", "duplicate_burst", "id_reuse", "out_of_enum", "out_of_range":
		p["rate"] = round3(0.005 + rng.Float64()*0.04)
	case "delay_tail":
		p["mean_s"] = round1(30 + rng.Float64()*300)
		p["sigma_s"] = round1(60 + rng.Float64()*300)
	case "storm":
		p["multiplier"] = 2 + rng.Intn(4)
	case "clock_skew":
		p["offset_s"] = round1(60 + rng.Float64()*300)
	case "oversize":
		p["bytes"] = 4096
	}
	return p
}

func round3(f float64) float64 { return float64(int(f*1000+0.5)) / 1000 }

func round1(f float64) float64 { return float64(int(f*10+0.5)) / 10 }

func uncovered(counts map[string]int) []string {
	var out []string
	for _, name := range perturb.Names {
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
