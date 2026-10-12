package app

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// churnScenario drives the k8s cluster through births, deaths, a fault, a
// perturbation and its clear, a retirement and effector calls (one refused at
// admission), over three hours.
func churnScenario(start int64) *model.Scenario {
	evict := func(entity, id string, at int64) model.ScenarioCommand {
		return model.ScenarioCommand{Op: model.OpEffectorInvoke, Args: map[string]any{"effector": "evict_pod",
			"entity_id": entity, "command_id": id, "args": map[string]any{"cluster_id": "cluster-a", "pod_id": entity}, "at_ns": at}}
	}
	return &model.Scenario{ID: "churn", World: model.ScenarioWorld{StartTimeNS: start}, Commands: []model.ScenarioCommand{
		advanceTo(start + 3600e9),
		{Op: model.OpFaultInject, Args: map[string]any{"entity_id": churnBorn, "fault": "crash_loop", "onset_ns": start + 3600e9, "params": nil}},
		{Op: model.OpPerturbApply, Args: map[string]any{"perturbation": "drop", "params": nil, "from_ns": start + 3600e9, "until_ns": int64(0)}},
		{Op: model.OpEntityRetire, Args: map[string]any{"entity_id": "cluster-a/pod-3", "reason": "drained", "at_ns": start + 3600e9}},
		evict("cluster-a/pod-3", "c-gone", start+3600e9),
		evict(churnBorn, "c-1", start+3600e9),
		advanceTo(start + 7200e9),
		{Op: model.OpPerturbClear, Args: map[string]any{"perturb_id": "drop-0"}},
		advanceTo(start + 3*3600e9)}}
}

// The dry run's claim to refuse exactly what the run refuses rests on the
// silent world being the emitting world without its emissions: same
// entities born and retired, same faults, same effector calls, same log.
func TestTheSilentWorldIsTheEmittingWorldWithoutItsEmissions(t *testing.T) {
	t.Parallel()
	cfg := churnConfig(t)
	scenario := churnScenario(cfg.StartTimeNS)
	emitting, err := New(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	silent, err := newShadow(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, step := range scenario.Commands {
		command := model.Command{Op: step.Op, Args: step.Args}
		got, want := executeCommand(context.Background(), silent, &command, keepEveryRefusal), executeCommand(context.Background(), emitting, &command, keepEveryRefusal)
		if (got == nil) != (want == nil) {
			t.Fatalf("%s: silent %v, emitting %v", step.Op, got, want)
		}
	}
	if len(emitting.World.EffectorCalls()) == 0 || len(emitting.World.ListFaults()) == 0 {
		t.Fatal("the scenario applied no effector call or fault, so the comparison proves nothing")
	}
	if emitting.World.EmittedCount() == 0 || silent.World.EmittedCount() != 0 {
		t.Fatalf("emitted: emitting %d, silent %d", emitting.World.EmittedCount(), silent.World.EmittedCount())
	}
	same := map[string][2]any{
		"entities":       {silent.World.EntityIDs(), emitting.World.EntityIDs()},
		"faults":         {silent.World.ListFaults(), emitting.World.ListFaults()},
		"effector calls": {silent.World.EffectorCalls(), emitting.World.EffectorCalls()},
		"command log":    {silent.commandLog, emitting.commandLog},
	}
	for name, pair := range same {
		if !reflect.DeepEqual(pair[0], pair[1]) {
			t.Errorf("%s differ:\n silent   %+v\n emitting %+v", name, pair[0], pair[1])
		}
	}
}

// A recorded consumer call the world refused at admission (here, to a pod
// retired before it) is dropped from the export, which still reproduces the
// run.
func TestARefusedCallIsDroppedAndTheRunStillReproduces(t *testing.T) {
	t.Parallel()
	cfg := churnConfig(t)
	original := recordedArtifact(t, cfg, churnScenario(cfg.StartTimeNS))
	exported, _ := ScenarioFromArtifact(original)
	pruned, refused, err := PruneRefusedCalls(context.Background(), cfg, exported)
	if err != nil || refused != 1 || len(pruned.Commands) != len(exported.Commands)-1 {
		t.Fatalf("pruned %d of %d commands, refused %d, err %v; want the one call to the retired pod dropped", len(pruned.Commands), len(exported.Commands), refused, err)
	}
	if err := CheckScenario(context.Background(), cfg, pruned); err != nil {
		t.Fatalf("the pruned export fails its check: %v", err)
	}
	reproduced := executedArtifact(t, cfg, pruned)
	if reproduced.ExpectedTraceDigest != original.ExpectedTraceDigest || reproduced.WorldDigest != original.WorldDigest {
		t.Fatalf("trace %s / world %s, want %s / %s", reproduced.ExpectedTraceDigest, reproduced.WorldDigest, original.ExpectedTraceDigest, original.WorldDigest)
	}
}

// recordedArtifact runs commands as a director session does: a call the world
// refuses is recorded and the session goes on.
func recordedArtifact(t *testing.T, cfg Config, scenario *model.Scenario) *model.RunArtifact {
	t.Helper()
	r, err := New(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, step := range scenario.Commands {
		command := model.Command{Op: step.Op, Args: step.Args}
		if err := executeCommand(context.Background(), r, &command, keepEveryRefusal); err != nil {
			t.Fatal(err)
		}
	}
	art, err := r.End("")
	if err != nil {
		t.Fatal(err)
	}
	return art
}

// An artifact whose log advances beyond the run span does not replay: the
// rule holds on every surface.
func TestAnArtifactBeyondTheRunSpanDoesNotReplay(t *testing.T) {
	t.Parallel()
	cfg := pondConfig(t)
	art := executedArtifact(t, cfg, &model.Scenario{Commands: []model.ScenarioCommand{advanceTo(cfg.StartTimeNS + 600e9)}})
	art.CommandLog[len(art.CommandLog)-1].Args["to_ns"] = cfg.StartTimeNS + MaxRunSpanNS + 1
	if _, err := ReplayArtifact(context.Background(), art, cfg.Domain, cfg.Adapter, ""); !errors.Is(err, ErrRunSpanExceeded) {
		t.Fatalf("ReplayArtifact = %v, want ErrRunSpanExceeded", err)
	}
}

func TestACanceledCheckStops(t *testing.T) {
	t.Parallel()
	cfg := churnConfig(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := CheckScenario(ctx, cfg, churnScenario(cfg.StartTimeNS)); !errors.Is(err, context.Canceled) {
		t.Fatalf("CheckScenario = %v, want context.Canceled", err)
	}
}
