package contract

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func directOutcome(t *testing.T, scenario *model.Scenario) *model.OutcomeFacts {
	t.Helper()
	spec, _, err := run.LoadInputs(filepath.Join(testsupport.RepositoryRoot(), "examples"), testsupport.AdaptersDir(), scenario.Domain, scenario.Adapter)
	if err != nil {
		t.Fatal(err)
	}
	w, err := world.New(spec, scenario.Seed, "", scenario.World.StartTimeNS, world.Options{RecordOutcome: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, command := range scenario.Commands {
		applyDirectly(t, w, command)
	}
	facts := w.OutcomeFacts(w.Clock())
	facts.Calls = []model.OutcomeCall{}
	for _, call := range w.EffectorCalls() {
		facts.Calls = append(facts.Calls, model.OutcomeCall{Effector: call.Effector, EntityID: call.EntityID, AtNS: call.AtNS, Accepted: call.Accepted})
	}
	return facts
}

func applyDirectly(t *testing.T, w *world.World, command model.ScenarioCommand) {
	t.Helper()
	args, err := command.Args, error(nil)
	switch command.Op {
	case model.OpClockAdvance:
		_, _, err = w.Advance(nanoseconds(args["to_ns"]))
	case model.OpFaultInject:
		params, _ := args["params"].(map[string]any)
		_, err = w.InjectFault(args["entity_id"].(string), args["fault"].(string), nanoseconds(args["onset_ns"]), params)
	case model.OpFaultClear:
		err = w.ClearFault(args["fault_id"].(string), nanoseconds(args["at_ns"]))
	case model.OpEffectorInvoke:
		callArgs, _ := args["args"].(map[string]any)
		_, err = w.InvokeEffector(args["effector"].(string), args["entity_id"].(string), args["command_id"].(string), callArgs, nanoseconds(args["at_ns"]))
	default:
		t.Fatalf("the direct run does not apply %s", command.Op)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestTheAsRunOutcomeIsDeterministicAndEqualsTheWorldRunOfTheSameCommands(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("three-hour runs of plant A")
	}
	scenario := storyScenario(t, "s2-a.reference")
	art := runScenarioArtifact(t, scenario)
	first, second := outcomeOf(t, art, model.OutcomeAsRun), outcomeOf(t, art, model.OutcomeAsRun)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("two as_run replays differ:\n%+v\n%+v", *first, *second)
	}
	if direct := directOutcome(t, scenario); !reflect.DeepEqual(first, direct) {
		t.Fatalf("as_run:\n%+v\nthe world run of the same commands:\n%+v", *first, *direct)
	}
	if len(first.Calls) != invocations(scenario) || !first.Calls[0].Accepted || first.HorizonNS != art.WorldConfig.StartTimeNS+10800e9 {
		t.Fatalf("as_run calls %+v at %d, want the reference policy's %d calls and the story's end", first.Calls, first.HorizonNS, invocations(scenario))
	}
}

func invocations(scenario *model.Scenario) int {
	count := 0
	for _, command := range scenario.Commands {
		if command.Op == model.OpEffectorInvoke {
			count++
		}
	}
	return count
}

func TestAPlantWireSessionRunYieldsItsOutcome(t *testing.T) {
	t.Parallel()
	socket, out := testsupport.SocketPath(t), filepath.Join(t.TempDir(), "session")
	hosted, listening := make(chan error, 1), make(chan struct{})
	go func() {
		_, err := plantwire.Host(t.Context(), socket, 0, plantwire.HostOptions{Environment: "simulated", ProtocolMax: 1, Epoch: 1,
			Ready: func() { close(listening) }})
		hosted <- err
	}()
	select {
	case <-listening:
	case err := <-hosted:
		t.Fatalf("the host ended before it listened: %v", err)
	}
	plan := `{"schema_version":"0.1","domain":"belt-conveyor-drives-a","seed":3,"host_socket":"` + socket +
		`","horizon_s":120,"disturbances":[{"at_s":30,"instance":"D-3","fault":"conveyor_drive__vfd_fault"}]}`
	result, err := plantwire.Connect(t.Context(), plantwire.ConnectOptions{Plan: strings.NewReader(plan), Source: "plan.json",
		DomainsDir: filepath.Join(testsupport.RepositoryRoot(), "examples"), AdaptersDir: testsupport.AdaptersDir(), Out: out,
		SliceNS: 10e9, BarrierCap: 30 * time.Second})
	if err != nil || result.Stopped || <-hosted != nil {
		t.Fatalf("session = %+v, %v", result, err)
	}
	art, err := run.LoadArtifact(filepath.Join(out, model.FileRunArtifact))
	if err != nil {
		t.Fatal(err)
	}
	asRun, nothing := outcomeOf(t, art, model.OutcomeAsRun), outcomeOf(t, art, model.OutcomeDoNothing)
	if asRun.HorizonNS != model.DefaultStartTimeNS+120e9 || len(asRun.Entities) != 5 || !reflect.DeepEqual(asRun, nothing) {
		t.Fatalf("as_run %+v, do_nothing %+v: want five entities at the 120 s horizon, alike with no host command", *asRun, *nothing)
	}
}
