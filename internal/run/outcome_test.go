package run_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func at(seconds int64) int64 { return model.DefaultStartTimeNS + seconds*1e9 }

func conveyorArtifact(t *testing.T) *model.RunArtifact {
	t.Helper()
	spec, err := domain.Load(testsupport.BeltConveyorDomain("a"))
	if err != nil {
		t.Fatal(err)
	}
	native, err := adapter.Load(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	r, err := run.New(t.Context(), run.Config{Domain: spec, Adapter: native, Seed: 2, SinkName: model.SinkInproc, TimeMode: model.TimeStepped,
		StartTimeNS: model.DefaultStartTimeNS, StartTimeSet: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(t.Context(), at(60), false); err != nil {
		t.Fatal(err)
	}
	if _, err := r.InjectFault("D-3", "conveyor_drive__vfd_fault", at(60), nil); err != nil {
		t.Fatal(err)
	}
	if _, err := r.InvokeEffector("conveyor_drive__stop_drive", "D-1", "c-1", map[string]any{"run_command": 0.0}, at(60)); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(t.Context(), at(300), false); err != nil {
		t.Fatal(err)
	}
	art, err := r.End("")
	if err != nil {
		t.Fatal(err)
	}
	return art
}

func TestEachOutcomeModeReplaysTheRunToItsEndWithItsOwnCalls(t *testing.T) {
	t.Parallel()
	art := conveyorArtifact(t)
	cases := []struct {
		mode  model.OutcomeMode
		calls []model.OutcomeCall
	}{
		{model.OutcomeAsRun, []model.OutcomeCall{{Effector: "conveyor_drive__stop_drive", EntityID: "D-1", AtNS: at(60), Accepted: false}}},
		{model.OutcomeDoNothing, []model.OutcomeCall{}},
		{model.PolicyOutcome("b4"), []model.OutcomeCall{{Effector: "bulk_feeder__set_feed", EntityID: "F-1", AtNS: at(61), Accepted: true}}},
	}
	for _, tc := range cases {
		t.Run(string(tc.mode), func(t *testing.T) {
			t.Parallel()
			facts, err := run.ReplayOutcome(t.Context(), art, nil, nil, tc.mode)
			if err != nil {
				t.Fatal(err)
			}
			if facts.HorizonNS != at(300) || len(facts.Entities) != 5 || !slices.Equal(facts.Calls, tc.calls) {
				t.Fatalf("facts at %d with calls %+v, want the run's end and %+v", facts.HorizonNS, facts.Calls, tc.calls)
			}
		})
	}
}

func TestAnUnknownOutcomeModeOrPolicyIsATypedError(t *testing.T) {
	t.Parallel()
	art := conveyorArtifact(t)
	if _, err := run.ReplayOutcome(t.Context(), art, nil, nil, model.PolicyOutcome("b9")); !errors.Is(err, run.ErrUnknownPolicy) {
		t.Fatalf("err = %v, want ErrUnknownPolicy", err)
	}
	if _, err := run.ReplayOutcome(t.Context(), art, nil, nil, "everything"); !errors.Is(err, run.ErrUnknownOutcomeMode) {
		t.Fatalf("err = %v, want ErrUnknownOutcomeMode", err)
	}
}
