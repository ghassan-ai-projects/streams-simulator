package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

const decidedNS = 1_500

func tripLabel() *model.GroundTruthRecord {
	return &model.GroundTruthRecord{ScenarioID: "belt/1", Domain: "belt", EntityID: "D-3", Label: "vfd_fault", ExpectedEpisode: true,
		FirstObservableTimeNS: firstSeenNS, DeadlineNS: deadlineNS,
		ExpectedActions: []model.ExpectedAction{{EntityID: "F-1", Effector: "set_feed"}, {EntityID: "D-3", Effector: "stop_drive"}}}
}

func appliedCall(effector, entity, command string) world.EffectorCall {
	return world.EffectorCall{CommandID: command, Effector: effector, EntityID: entity, AtNS: actedAtNS, Mode: "ok", Accepted: true, EffectApplied: true}
}

func decidedAction(effector, entity, command string) model.Action {
	return model.Action{CommandID: command, Effector: effector, EntityID: entity, IssuedAt: model.FormatTime(decidedNS), OutcomeBelieved: model.BelievedSucceeded}
}

func limpFeedRun() Evidence {
	return Evidence{RunID: "run-1", Ledger: []model.LedgerRecord{},
		Verdict: &model.Verdict{
			Detections: []model.Detection{{EntityID: "D-3", DetectedAt: model.FormatTime(firstSeenNS + 500), Label: "vfd_fault"}, {EntityID: "F-1", DetectedAt: model.FormatTime(firstSeenNS + 500)}},
			Actions:    []model.Action{decidedAction("set_feed", "F-1", "sha256:aa"), decidedAction("stop_drive", "D-2", "sha256:bb")},
		},
		Calls: []world.EffectorCall{appliedCall("set_feed", "F-1", "sha256:aa"), appliedCall("stop_drive", "D-2", "sha256:bb")},
	}
}

func TestAnActionDecidedBeforeTheWorldAppliedItIsFaithfulByItsCommandID(t *testing.T) {
	t.Parallel()
	graded, err := ScoreRun(limpFeedRun(), []*model.GroundTruthRecord{tripLabel()})
	if err != nil {
		t.Fatal(err)
	}
	if !graded.Scorecards[0].Consumer.ActionFidelity {
		t.Fatal("two actions issued at 1500 ns that name the command ids of the calls applied at 2000 ns are not faithful")
	}
	ev := limpFeedRun()
	ev.Verdict.Actions[0].CommandID = "sha256:cc"
	if consumer(ev, tripLabel()).ActionFidelity {
		t.Fatal("an action whose command id names no call is faithful")
	}
}

func TestACallOnAnInstanceTheFaultExpectsIsTheRightActionAndNotUnexpected(t *testing.T) {
	t.Parallel()
	graded, err := ScoreRun(limpFeedRun(), []*model.GroundTruthRecord{tripLabel()})
	if err != nil {
		t.Fatal(err)
	}
	loop := graded.Scorecards[0].Loop
	if !loop.ActionAppropriate || !loop.DeadlineAdhered || loop.EffectCalls != 2 {
		t.Fatalf("loop = %+v, want the feed call on F-1 appropriate and before the deadline, both calls counted", loop)
	}
	unexpected := graded.Unexpected
	if len(unexpected.Actions) != 1 || unexpected.Actions[0].EntityID != "D-2" || unexpected.EffectCalls != 1 {
		t.Fatalf("unexpected = %+v, want only the stop of D-2", unexpected)
	}
	if len(unexpected.Detections) != 1 || unexpected.Detections[0].EntityID != "F-1" {
		t.Fatalf("unexpected detections = %+v, want the one on F-1, which carries no label", unexpected.Detections)
	}
}

func TestAnExpectedActionThatHadNoEffectIsNotTheRightAction(t *testing.T) {
	t.Parallel()
	ev := limpFeedRun()
	ev.Calls[0].EffectApplied = false
	if loop(ev, tripLabel()).ActionAppropriate {
		t.Fatal("a feed call without effect was credited")
	}
	ev.Calls[0].EffectApplied, ev.Calls[0].EntityID = true, "F-2"
	if loop(ev, tripLabel()).ActionAppropriate {
		t.Fatal("a feed call on an instance the label does not name was credited")
	}
}

func TestAmongSeveralLabelsEachLoopSeesTheCallsOnItsEntityAndItsExpectedInstances(t *testing.T) {
	t.Parallel()
	other := labelOn("D-1")
	graded, err := ScoreRun(limpFeedRun(), []*model.GroundTruthRecord{tripLabel(), other})
	if err != nil {
		t.Fatal(err)
	}
	trip, wear := graded.Scorecards[0], graded.Scorecards[1]
	if !trip.Loop.ActionAppropriate || trip.Loop.EffectCalls != 1 || !trip.Judgment.Detected || trip.Judgment.DetectionCount != 1 {
		t.Fatalf("D-3 = %+v %+v, want the F-1 call alone in its loop and its own detection alone in its judgment", trip.Loop, trip.Judgment)
	}
	if wear.Loop.EffectCalls != 0 || wear.Loop.ActionAppropriate || wear.Judgment.DetectionCount != 0 {
		t.Fatalf("D-1 = %+v %+v, want nothing from the other instances", wear.Loop, wear.Judgment)
	}
	if graded.Unexpected.EffectCalls != 1 || len(graded.Unexpected.Detections) != 1 {
		t.Fatalf("unexpected = %+v, want the D-2 stop and the F-1 detection", graded.Unexpected)
	}
}
