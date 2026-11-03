package domain

import (
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

const (
	firstSeenNS = 1_000
	actedAtNS   = 2_000
	deadlineNS  = 5_000
)

func labelOn(entity string) *model.GroundTruthRecord {
	return &model.GroundTruthRecord{ScenarioID: "x/" + entity, Domain: "x", EntityID: entity, Label: "wear", ExpectedEpisode: true,
		ExpectedEffector: "service", FirstObservableTimeNS: firstSeenNS, DeadlineNS: deadlineNS}
}

func callOn(entity, command string) world.EffectorCall {
	return world.EffectorCall{CommandID: command, Effector: "service", EntityID: entity, AtNS: actedAtNS, Mode: "ok", Accepted: true, EffectApplied: true}
}

func detectionOn(entity string) model.Detection {
	return model.Detection{EntityID: entity, DetectedAt: model.FormatTime(firstSeenNS + 500), Label: "wear"}
}

func actionOn(entity, command string) model.Action {
	return model.Action{CommandID: command, Effector: "service", EntityID: entity, IssuedAt: model.FormatTime(actedAtNS), OutcomeBelieved: model.BelievedSucceeded}
}

func twoFaultRun() Evidence {
	return Evidence{RunID: "run-1", Ledger: []model.LedgerRecord{},
		Verdict: &model.Verdict{
			Detections: []model.Detection{detectionOn("pump-1"), detectionOn("pump-3")},
			Actions:    []model.Action{actionOn("pump-1", "c1"), actionOn("pump-3", "c3")},
		},
		Calls: []world.EffectorCall{callOn("pump-1", "c1"), callOn("pump-3", "c3")},
	}
}

func TestEachDisturbanceOfARunIsGradedOnItsOwnEntityAlone(t *testing.T) {
	t.Parallel()
	graded, err := ScoreRun(twoFaultRun(), []*model.GroundTruthRecord{labelOn("pump-1"), labelOn("pump-2")})
	if err != nil || len(graded.Scorecards) != 2 {
		t.Fatalf("ScoreRun = %+v, %v, want two scorecards", graded, err)
	}
	found, missed := graded.Scorecards[0], graded.Scorecards[1]
	if !found.Judgment.Detected || !found.Judgment.LabelCorrect || found.Judgment.DetectionCount != 1 ||
		!found.Loop.ActionAppropriate || found.Loop.EffectCalls != 1 || !found.Loop.DeadlineAdhered {
		t.Fatalf("the detected and treated pump = %+v %+v", found.Judgment, found.Loop)
	}
	if missed.Judgment.Detected || missed.Judgment.DetectionCount != 0 || missed.Loop.ActionAppropriate || missed.Loop.EffectCalls != 0 || missed.Loop.DeadlineAdhered {
		t.Fatalf("the missed pump = %+v %+v, want nothing credited from the other pumps", missed.Judgment, missed.Loop)
	}
}

func TestWhatHappensOnAPumpWithoutADisturbanceIsCountedOnceAtRunLevel(t *testing.T) {
	t.Parallel()
	graded, err := ScoreRun(twoFaultRun(), []*model.GroundTruthRecord{labelOn("pump-1"), labelOn("pump-2")})
	if err != nil {
		t.Fatal(err)
	}
	unexpected := graded.Unexpected
	if len(unexpected.Detections) != 1 || unexpected.Detections[0].EntityID != "pump-3" ||
		len(unexpected.Actions) != 1 || unexpected.Actions[0].EntityID != "pump-3" || unexpected.EffectCalls != 1 {
		t.Fatalf("unexpected = %+v, want the detection, action and call on pump-3", unexpected)
	}
}

func TestARunWithoutADisturbanceHasNoScorecardButKeepsItsUnexpectedFindings(t *testing.T) {
	t.Parallel()
	graded, err := ScoreRun(twoFaultRun(), nil)
	if err != nil || len(graded.Scorecards) != 0 || len(graded.Unexpected.Detections) != 2 || len(graded.Unexpected.Actions) != 2 || graded.Unexpected.EffectCalls != 2 {
		t.Fatalf("ScoreRun = %+v, %v, want no scorecard and all findings unexpected", graded, err)
	}
	if Document(graded) != any(graded) {
		t.Fatal("a run without a scorecard prints the run document")
	}
}

func TestOneDisturbanceScoresExactlyAsItAlwaysDid(t *testing.T) {
	t.Parallel()
	ev, gt := twoFaultRun(), labelOn("pump-1")
	alone, err := Score(ev, gt)
	if err != nil {
		t.Fatal(err)
	}
	graded, err := ScoreRun(ev, []*model.GroundTruthRecord{gt})
	if err != nil || Document(graded) != any(graded.Scorecards[0]) {
		t.Fatalf("ScoreRun = %+v, %v, want the single scorecard as the document", graded, err)
	}
	if *graded.Scorecards[0] != *alone {
		t.Fatalf("scorecard = %+v, want %+v", graded.Scorecards[0], alone)
	}
}

func TestScoreRunNeedsTheVerdict(t *testing.T) {
	t.Parallel()
	if _, err := ScoreRun(Evidence{RunID: "run-1"}, nil); err == nil || !strings.Contains(err.Error(), "no verdict submitted") {
		t.Fatal("a run without a verdict was scored")
	}
}
