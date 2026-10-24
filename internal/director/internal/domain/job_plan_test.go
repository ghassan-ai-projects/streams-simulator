package domain

import (
	"encoding/json"
	"reflect"
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const planStart = int64(1_000_000_000_000)

func ops(commands []model.ScenarioCommand) []string {
	out := make([]string, len(commands))
	for i, command := range commands {
		out[i] = command.Op
		switch command.Op {
		case model.OpClockAdvance:
			out[i] += " to+" + itoa((command.Args["to_ns"].(int64)-planStart)/1e9)
		case model.OpFaultInject:
			out[i] += " " + command.Args["entity_id"].(string) + " " + command.Args["fault"].(string) + " @+" + itoa((command.Args["onset_ns"].(int64)-planStart)/1e9)
		case model.OpEffectorInvoke:
			out[i] += " " + command.Args["command_id"].(string) + " " + command.Args["effector"].(string) + " @+" + itoa((command.Args["at_ns"].(int64)-planStart)/1e9)
		}
	}
	return out
}

func itoa(n int64) string {
	raw, _ := json.Marshal(n)
	return string(raw)
}

func TestAPlanBecomesChunkedAdvancesWithActionsAtEachBoundary(t *testing.T) {
	t.Parallel()
	request := JobRequest{Domain: "aquaculture-pond", Seed: 3, DurationS: 100,
		Faults: []JobFault{{EntityID: "p-1", Fault: "f-late", AtS: 40}, {EntityID: "p-1", Fault: "f-start", AtS: 0}, {EntityID: "p-2", Fault: "f-40", AtS: 40}},
		Calls: []JobCall{{EntityID: "p-1", Effector: "e-40", Args: map[string]any{"x": 1}, AtS: 40}, {EntityID: "p-1", Effector: "e-0", AtS: 0},
			{EntityID: "p-2", Effector: "e-end", AtS: 100}}}
	scenario := PlanScenario(request, "native-jsonl", planStart)
	want := []string{
		"fault.inject p-1 f-start @+0", "effector.invoke job-2 e-0 @+0",
		"clock.advance to+5", "clock.advance to+10", "clock.advance to+15", "clock.advance to+20", "clock.advance to+25",
		"clock.advance to+30", "clock.advance to+35", "clock.advance to+40",
		"fault.inject p-1 f-late @+40", "fault.inject p-2 f-40 @+40", "effector.invoke job-1 e-40 @+40",
		"clock.advance to+45", "clock.advance to+50", "clock.advance to+55", "clock.advance to+60", "clock.advance to+65",
		"clock.advance to+70", "clock.advance to+75", "clock.advance to+80", "clock.advance to+85", "clock.advance to+90",
		"clock.advance to+95", "clock.advance to+100", "effector.invoke job-3 e-end @+100",
	}
	if got := ops(scenario.Commands); !reflect.DeepEqual(got, want) {
		t.Fatalf("commands:\n got %v\nwant %v", got, want)
	}
}

func TestTheChunkCountFollowsThePaceAndNeverExceedsTheDuration(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		rate     float64
		duration int64
		advances int
	}{
		"as fast as possible: 20 chunks":   {0, 3600, 20},
		"paced: 30 chunks":                 {60, 3600, 30},
		"a short run is one step a second": {0, 7, 7},
		"one second":                       {60, 1, 1},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			scenario := PlanScenario(JobRequest{Domain: "d", DurationS: tc.duration, Rate: tc.rate}, "native-jsonl", planStart)
			if got := len(scenario.Commands); got != tc.advances {
				t.Fatalf("%d commands, want %d advances", got, tc.advances)
			}
			last := scenario.Commands[len(scenario.Commands)-1]
			if last.Op != model.OpClockAdvance || last.Args["to_ns"] != planStart+tc.duration*1e9 {
				t.Fatalf("last command = %+v, want an advance to the requested end", last)
			}
		})
	}
}

func TestThePlannedScenarioIsDeterministicAndSchemaValid(t *testing.T) {
	t.Parallel()
	request := validJob()
	first, second := PlanScenario(request, "native-jsonl", planStart), PlanScenario(request, "native-jsonl", planStart)
	if !reflect.DeepEqual(first, second) {
		t.Fatal("the same plan built two different scenarios")
	}
	raw, err := json.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	if err := model.ValidateScenario(raw); err != nil {
		t.Fatalf("the scenario does not satisfy scenario-v0.1: %v", err)
	}
	if first.Domain != "aquaculture-pond" || first.Seed != 7 || first.Adapter != "native-jsonl" || first.World.StartTimeNS != planStart {
		t.Fatalf("scenario header = %+v", first)
	}
}

func TestPlannedArgumentsAreExactlyTheCommandLogShape(t *testing.T) {
	t.Parallel()
	scenario := PlanScenario(validJob(), "native-jsonl", planStart)
	var fault, call *model.ScenarioCommand
	for i := range scenario.Commands {
		switch scenario.Commands[i].Op {
		case model.OpFaultInject:
			fault = &scenario.Commands[i]
		case model.OpEffectorInvoke:
			call = &scenario.Commands[i]
		}
	}
	wantFault := map[string]any{"entity_id": "pond-1", "fault": "aerator_failure", "onset_ns": planStart + 600e9, "params": map[string]any{}}
	wantCall := map[string]any{"effector": "start_aerator", "entity_id": "pond-1", "command_id": "job-1", "args": map[string]any{"pond_id": "pond-1"}, "at_ns": planStart + 1200e9}
	if !reflect.DeepEqual(fault.Args, wantFault) || !reflect.DeepEqual(call.Args, wantCall) {
		t.Fatalf("fault args %v, call args %v", fault.Args, call.Args)
	}
}

func recoveryJob() JobRequest {
	return JobRequest{Domain: "aquaculture-pond", Seed: 3, DurationS: 100, Rate: 0,
		Faults:        []JobFault{{EntityID: "p-1", Fault: "late", AtS: 60}, {EntityID: "p-1", Fault: "early", AtS: 10, UntilS: at(35)}, {EntityID: "p-2", Fault: "open", AtS: 10}},
		Perturbations: []JobPerturbation{{Kind: "drop", Params: map[string]any{"rate": 0.3}, AtS: 20, UntilS: at(70)}, {Kind: "malformed", AtS: 100}}}
}

func TestAFaultsPlannedIdIsItsRankInInjectionOrder(t *testing.T) {
	t.Parallel()
	want := []string{"f-2", "f-0", "f-1"}
	if got := FaultIDs(recoveryJob().Faults); !reflect.DeepEqual(got, want) {
		t.Fatalf("ids = %v, want %v: ranked by time, then by position in the request", got, want)
	}
}

func TestARecoveredFaultIsClearedByIdAtItsEndAndADeliveryConditionIsAWindow(t *testing.T) {
	t.Parallel()
	scenario := PlanScenario(recoveryJob(), "native-jsonl", planStart)
	var inject, clear, apply []model.ScenarioCommand
	for _, command := range scenario.Commands {
		switch command.Op {
		case model.OpFaultInject:
			inject = append(inject, command)
		case model.OpFaultClear:
			clear = append(clear, command)
		case model.OpPerturbApply:
			apply = append(apply, command)
		}
	}
	if len(inject) != 3 || len(clear) != 1 || len(apply) != 2 {
		t.Fatalf("%d injections, %d clears, %d applications, want 3, 1, 2", len(inject), len(clear), len(apply))
	}
	if want := map[string]any{"fault_id": "f-0", "at_ns": planStart + 35e9}; !reflect.DeepEqual(clear[0].Args, want) {
		t.Fatalf("clear = %v, want %v", clear[0].Args, want)
	}
	wantDrop := map[string]any{"perturbation": "drop", "params": map[string]any{"rate": 0.3}, "from_ns": planStart + 20e9, "until_ns": planStart + 70e9}
	wantOpen := map[string]any{"perturbation": "malformed", "params": map[string]any{}, "from_ns": planStart + 100e9, "until_ns": int64(0)}
	if !reflect.DeepEqual(apply[0].Args, wantDrop) || !reflect.DeepEqual(apply[1].Args, wantOpen) {
		t.Fatalf("applications = %v, %v", apply[0].Args, apply[1].Args)
	}
}

func TestTheClockStopsAtTheMomentOfARecoveryAndOfTheEndOfAWindow(t *testing.T) {
	t.Parallel()
	got := ops(PlanScenario(recoveryJob(), "native-jsonl", planStart).Commands)
	for _, want := range []string{"clock.advance to+35", "clock.advance to+70"} {
		if !slices.Contains(got, want) {
			t.Errorf("no %q in %v", want, got)
		}
	}
	at35 := slices.Index(got, "clock.advance to+35")
	if next := got[at35+1]; next != "fault.clear" {
		t.Fatalf("after reaching 35 s the plan runs %q, want the clear", next)
	}
}

func TestAPlanWithRecoveriesAndConditionsStillSatisfiesTheScenarioContract(t *testing.T) {
	t.Parallel()
	raw, err := json.Marshal(PlanScenario(recoveryJob(), "native-jsonl", planStart))
	if err != nil {
		t.Fatal(err)
	}
	if err := model.ValidateScenario(raw); err != nil {
		t.Fatalf("the scenario does not satisfy scenario-v0.1: %v", err)
	}
}

func TestAtTheSameMomentAPlanInjectsThenClearsThenPerturbsThenCalls(t *testing.T) {
	t.Parallel()
	request := JobRequest{Domain: "d", DurationS: 40,
		Faults:        []JobFault{{EntityID: "p", Fault: "a", AtS: 10, UntilS: at(20)}, {EntityID: "p", Fault: "b", AtS: 20}},
		Perturbations: []JobPerturbation{{Kind: "drop", AtS: 20}},
		Calls:         []JobCall{{EntityID: "p", Effector: "e", AtS: 20}}}
	got := ops(PlanScenario(request, "native-jsonl", planStart).Commands)
	start := slices.Index(got, "clock.advance to+20")
	want := []string{"fault.inject p b @+20", "fault.clear", "perturb.apply", "effector.invoke job-1 e @+20"}
	if !reflect.DeepEqual(got[start+1:start+5], want) {
		t.Fatalf("at 20 s: %v, want %v", got[start+1:start+5], want)
	}
}
