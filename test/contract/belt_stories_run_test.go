package contract

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

var (
	storyDrives      = []string{"D-1", "D-2", "D-3"}
	storyProtections = []string{"relay_ok", "inverter_ok", "overtemp_ok", "free_ok"}
)

// storyOutcome is what a run of a story did: when each watched state first reached its tripped value, the material on
// the belt when D-3 seized, and the plant just before the anchor time.
type storyOutcome struct {
	first        map[string]float64
	seizedMassKg float64
	anchor       map[string]float64
	simulatedS   float64
	facts        *model.OutcomeFacts
}

func (o storyOutcome) at(key string) float64 {
	if at, ok := o.first[key]; ok {
		return at
	}
	return math.Inf(1)
}

func (o storyOutcome) firstOf(entities, states []string) (float64, string) {
	when, what := math.Inf(1), ""
	for _, entity := range entities {
		for _, state := range states {
			if at := o.at(entity + "." + state); at < when {
				when, what = at, entity+"."+state
			}
		}
	}
	return when, what
}

type storyRun struct {
	*beltWorld
	resolutionS float64
	anchorS     float64
	outcome     storyOutcome
}

func number(value any) float64 {
	switch v := value.(type) {
	case int64:
		return float64(v)
	case float64:
		return v
	case json.Number:
		f, _ := v.Float64()
		return f
	}
	panic("not a number")
}

func nanoseconds(value any) int64 {
	if ns, exact := value.(int64); exact {
		return ns
	}
	return int64(number(value))
}

func secondsOf(ns any) float64 { return float64(nanoseconds(ns)-model.DefaultStartTimeNS) / 1e9 }

// runStory executes a story scenario's commands on a noiseless world to its horizon, watching the plant every
// resolutionS seconds, and records its outcome.
func runStory(t *testing.T, story beltStory, row map[string]any, policy string, resolutionS, anchorS float64) storyOutcome {
	t.Helper()
	scenario := story.scenario(row, policy)
	r := &storyRun{beltWorld: newSeededBeltWorld(t, beltPlant(t, story.Plant), scenario.Seed), resolutionS: resolutionS, anchorS: anchorS,
		outcome: storyOutcome{first: map[string]float64{}, anchor: map[string]float64{}}}
	for _, command := range scenario.Commands {
		r.execute(command)
	}
	r.outcome.simulatedS, r.outcome.facts = r.nowS, worldOutcome(r.w)
	return r.outcome
}

func (r *storyRun) execute(command model.ScenarioCommand) {
	args := command.Args
	switch command.Op {
	case model.OpClockAdvance:
		r.watchUntil(secondsOf(args["to_ns"]))
	case model.OpFaultInject:
		params, _ := args["params"].(map[string]any)
		if _, err := r.w.InjectFault(args["entity_id"].(string), args["fault"].(string), nanoseconds(args["onset_ns"]), params); err != nil {
			r.t.Fatal(err)
		}
	case model.OpFaultClear:
		if err := r.w.ClearFault(args["fault_id"].(string), nanoseconds(args["at_ns"])); err != nil {
			r.t.Fatal(err)
		}
	case model.OpEffectorInvoke:
		if _, err := r.w.InvokeEffector(args["effector"].(string), args["entity_id"].(string), args["command_id"].(string),
			args["args"].(map[string]any), nanoseconds(args["at_ns"])); err != nil {
			r.t.Fatal(err)
		}
	}
}

func (r *storyRun) watchUntil(untilS float64) {
	for r.nowS < untilS {
		next := math.Min(untilS, r.nowS+r.resolutionS)
		if r.nowS < r.anchorS && next > r.anchorS {
			next = r.anchorS
		}
		r.advance(next)
		r.observe()
	}
}

func (r *storyRun) observe() {
	for _, drive := range storyDrives {
		for _, state := range storyProtections {
			r.mark(drive, state, 0)
		}
	}
	r.mark("D-3", "run_cmd", 0)
	r.mark("B-1", "stall_ok", 0)
	r.mark("B-1", "stranded", 1)
	if _, seized := r.outcome.first["D-3.free_ok"]; seized && r.outcome.seizedMassKg == 0 {
		r.outcome.seizedMassKg = r.state("B-1", "m")
	}
	if r.nowS == r.anchorS {
		for _, key := range [][2]string{{"D-1", "theta_inv"}, {"B-1", "m"}, {"B-1", "v"}, {"B-1", "feed"}} {
			r.outcome.anchor[key[0]+"."+key[1]] = r.state(key[0], key[1])
		}
	}
}

func (r *storyRun) mark(entity, state string, tripped float64) {
	key := entity + "." + state
	if _, seen := r.outcome.first[key]; !seen && r.state(entity, state) == tripped {
		r.outcome.first[key] = r.nowS
	}
}
