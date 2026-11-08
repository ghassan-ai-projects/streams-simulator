package contract

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func storyScenario(t *testing.T, name string) *model.Scenario {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(testsupport.BeltConveyorStoriesDir(), "scenarios", name+".scenario.json"))
	if err != nil {
		t.Fatal(err)
	}
	scenario, err := model.DecodeScenario(raw)
	if err != nil {
		t.Fatal(err)
	}
	return scenario
}

func runScenarioArtifact(t *testing.T, scenario *model.Scenario) *model.RunArtifact {
	t.Helper()
	spec, native, err := run.LoadInputs(filepath.Join(testsupport.RepositoryRoot(), "examples"), testsupport.AdaptersDir(), scenario.Domain, scenario.Adapter)
	if err != nil {
		t.Fatal(err)
	}
	r, err := run.New(t.Context(), run.Config{Domain: spec, Adapter: native, Seed: scenario.Seed, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: scenario.World.StartTimeNS, StartTimeSet: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := r.ExecuteScenario(t.Context(), scenario); err != nil {
		t.Fatal(err)
	}
	art, err := r.End("")
	if err != nil {
		t.Fatal(err)
	}
	return art
}

func outcomeOf(t *testing.T, art *model.RunArtifact, mode model.OutcomeMode) *model.OutcomeFacts {
	t.Helper()
	facts, err := run.ReplayOutcome(t.Context(), art, nil, nil, mode)
	if err != nil {
		t.Fatal(err)
	}
	return facts
}

func withFeedCalls(scenario *model.Scenario, atNS ...int64) *model.Scenario {
	written := *scenario
	written.Commands = slices.Clone(scenario.Commands)
	for i, at := range atNS {
		index := slices.IndexFunc(written.Commands, func(c model.ScenarioCommand) bool {
			return c.Op == model.OpClockAdvance && nanoseconds(c.Args["to_ns"]) > at
		})
		call := model.ScenarioCommand{Op: model.OpEffectorInvoke, Args: map[string]any{"effector": "bulk_feeder__set_feed", "entity_id": "F-1",
			"command_id": "hand-" + string(rune('a'+i)), "args": map[string]any{"feeder_id": "F-1", "percent": 67.0}, "at_ns": at}}
		written.Commands = slices.Insert(written.Commands, index, model.ScenarioCommand{Op: model.OpClockAdvance, Args: map[string]any{"to_ns": at}}, call)
	}
	return &written
}

func nextSampleNS(atNS int64) int64 {
	elapsed := atNS - model.DefaultStartTimeNS
	return model.DefaultStartTimeNS + (elapsed+int64(1e9)-1)/int64(1e9)*int64(1e9)
}

// B4 samples D-3 at its onset before the fault is injected there, so it sees the fault one second later.
func TestAPolicyRunEqualsTheHandWrittenLogOfItsCalls(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("four three-hour runs of plant B")
	}
	doNothing := storyScenario(t, "s2-b.do-nothing")
	policy := outcomeOf(t, runScenarioArtifact(t, doNothing), model.PolicyOutcome("b4"))
	first := nanoseconds(doNothing.Commands[1].Args["onset_ns"]) + int64(1e9)
	once := outcomeOf(t, runScenarioArtifact(t, withFeedCalls(doNothing, first)), model.OutcomeAsRun)
	second := nextSampleNS(recordedTrip(entityOutcome(t, once, "D-1"), "inverter_ok").FirstNS)
	if other := nextSampleNS(recordedTrip(entityOutcome(t, once, "D-2"), "inverter_ok").FirstNS); other != second {
		t.Fatalf("the survivors' inverters trip at different samples (%d, %d); the hand-written log assumes one", second, other)
	}
	hand := outcomeOf(t, runScenarioArtifact(t, withFeedCalls(doNothing, first, second, second)), model.OutcomeAsRun)
	t.Logf("B4 fires at %.0f s (D-3) and %.0f s (D-1, D-2); stranded from %.0f s", secondsOf(first), secondsOf(second),
		secondsOf(entityOutcome(t, policy, "B-1").Boundaries[0].FirstNS))
	if !reflect.DeepEqual(policy, hand) {
		t.Fatalf("policy b4:\n%+v\nhand-written log:\n%+v", *policy, *hand)
	}
}

func TestDoingNothingOnAReferenceRunIsTheStorysDoNothingRun(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("three-hour runs of both plants")
	}
	for _, story := range []string{"s2-a", "s2-b"} {
		t.Run(story, func(t *testing.T) {
			t.Parallel()
			removed := outcomeOf(t, runScenarioArtifact(t, storyScenario(t, story+".reference")), model.OutcomeDoNothing)
			nothing := outcomeOf(t, runScenarioArtifact(t, storyScenario(t, story+".do-nothing")), model.OutcomeAsRun)
			if !reflect.DeepEqual(removed, nothing) || len(removed.Calls) != 0 {
				t.Fatalf("do_nothing of the reference run:\n%+v\nthe do-nothing run:\n%+v", *removed, *nothing)
			}
		})
	}
}
