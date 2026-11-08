package domain

import (
	"errors"
	"reflect"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const playStart int64 = 1_000_000_000_000

func sec(s float64) int64 { return playStart + int64(s*1e9) }

func playSpec(forS float64, then ...model.PolicyCall) *model.DomainSpec {
	feed := model.Effector{Name: "set_feed", ArgsSchema: map[string]any{"required": []any{"feeder_id", "percent"},
		"properties": map[string]any{"feeder_id": map[string]any{"type": "string"}, "percent": map[string]any{"type": "number"}}}}
	stop := model.Effector{Name: "stop_drive", ArgsSchema: map[string]any{"required": []any{"drive_id"},
		"properties": map[string]any{"drive_id": map[string]any{"type": "string"}}}}
	return &model.DomainSpec{Kinds: []model.Kind{{Name: "feeder", Effectors: []model.Effector{feed}}, {Name: "drive", Effectors: []model.Effector{stop}}},
		GroundTruth: model.GroundTruth{BaselinePolicies: []model.BaselinePolicy{{Name: "b", ForS: forS, Then: then,
			When: model.PolicyCondition{Kind: "drive", State: "available", Operator: "lt", Threshold: 0.5}}}}}
}

func newPlay(t *testing.T, spec *model.DomainSpec) *PolicyPlay {
	t.Helper()
	play, err := NewPolicyPlay(spec, "b", playStart)
	if err != nil {
		t.Fatal(err)
	}
	return play
}

func observeSeconds(play *PolicyPlay, through float64, holding func(s float64) map[string]bool) []PolicyCall {
	var calls []PolicyCall
	for stop, ok := play.NextStop(sec(through)); ok; stop, ok = play.NextStop(sec(through)) {
		if play.SampleDue(stop) {
			play.Observe(stop, holding(float64(stop-playStart)/1e9))
		}
		calls = append(calls, play.Due(stop)...)
	}
	return calls
}

func TestAPolicyFiresOnEachInstancesRisingEdgeHeldForItsHoldTime(t *testing.T) {
	t.Parallel()
	play := newPlay(t, playSpec(3, model.PolicyCall{Effector: "stop_drive", Target: model.PolicyTargetMatched}))
	down := func(s float64) map[string]bool {
		return map[string]bool{"D-1": s >= 10 && s < 12 || s >= 20 && s < 40, "D-2": s >= 30}
	}
	calls := observeSeconds(play, 60, down)
	want := []PolicyCall{
		{Effector: "stop_drive", EntityID: "D-1", CommandID: "policy-b-1", Args: map[string]any{"drive_id": "D-1"}, AtNS: sec(23)},
		{Effector: "stop_drive", EntityID: "D-2", CommandID: "policy-b-2", Args: map[string]any{"drive_id": "D-2"}, AtNS: sec(33)},
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls %+v, want D-1 at 23 s (its 2 s dip is shorter than the hold) and D-2 at 33 s, once each", calls)
	}
}

func TestAFiringIssuesItsCallsInOrderAtTheirDelaysToTheirTargets(t *testing.T) {
	t.Parallel()
	play := newPlay(t, playSpec(0,
		model.PolicyCall{Effector: "set_feed", Target: "F-1", Args: map[string]any{"percent": 67.0}, DelayS: 2},
		model.PolicyCall{Effector: "stop_drive", Target: model.PolicyTargetMatched}))
	calls := observeSeconds(play, 10, func(s float64) map[string]bool { return map[string]bool{"D-3": s >= 5} })
	want := []PolicyCall{
		{Effector: "stop_drive", EntityID: "D-3", CommandID: "policy-b-2", Args: map[string]any{"drive_id": "D-3"}, AtNS: sec(5)},
		{Effector: "set_feed", EntityID: "F-1", CommandID: "policy-b-1", Args: map[string]any{"feeder_id": "F-1", "percent": 67.0}, AtNS: sec(7)},
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls %+v, want the matched stop at the firing and the feed to F-1 2 s later", calls)
	}
}

func TestAConditionTrueAtTheFirstSampleFiresAndTheGridIsOneSecond(t *testing.T) {
	t.Parallel()
	play := newPlay(t, playSpec(0, model.PolicyCall{Effector: "stop_drive", Target: model.PolicyTargetMatched}))
	if stop, ok := play.NextStop(sec(0)); !ok || stop != sec(0) {
		t.Fatalf("first stop %d, %v; want a sample at the world start", stop, ok)
	}
	play.Observe(sec(0), map[string]bool{"D-1": true})
	if due := play.Due(sec(0)); len(due) != 1 {
		t.Fatalf("due %+v, want a call for a condition already present at the start", due)
	}
	if stop, ok := play.NextStop(sec(0.5)); ok || stop != sec(1) {
		t.Fatalf("next stop %d (%v), want the next sample one second later", stop, ok)
	}
}

func TestAnUnknownPolicyIsATypedError(t *testing.T) {
	t.Parallel()
	if _, err := NewPolicyPlay(playSpec(0), "b9", playStart); !errors.Is(err, ErrUnknownPolicy) {
		t.Fatalf("err = %v, want ErrUnknownPolicy", err)
	}
}

func TestOutcomeModesKeepOrDropTheEffectorCalls(t *testing.T) {
	t.Parallel()
	log := []model.Command{{Seq: 0, Op: model.OpClockAdvance}, {Seq: 1, Op: model.OpEffectorInvoke}, {Seq: 2, Op: model.OpFaultInject}}
	for mode, want := range map[model.OutcomeMode][]int64{model.OutcomeAsRun: {0, 1, 2}, model.OutcomeDoNothing: {0, 2}, model.PolicyOutcome("b4"): {0, 2}} {
		kept, err := OutcomeCommands(log, mode)
		var seqs []int64
		for _, command := range kept {
			seqs = append(seqs, command.Seq)
		}
		if err != nil || !reflect.DeepEqual(seqs, want) {
			t.Errorf("%s: kept %v (%v), want %v", mode, seqs, err, want)
		}
	}
	if _, err := OutcomeCommands(log, "replay"); !errors.Is(err, ErrUnknownOutcomeMode) {
		t.Fatalf("err = %v, want ErrUnknownOutcomeMode", err)
	}
}
