package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func runDigest(t *testing.T, graded *RunScore) string {
	t.Helper()
	raw, err := json.Marshal(graded)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func mixedVerdictRun() Evidence {
	ev := limpFeedRun()
	ev.Verdict.Detections = append(ev.Verdict.Detections,
		model.Detection{EntityID: "D-3", DetectedAt: model.FormatTime(firstSeenNS - 100), Label: "vfd_fault"},
		model.Detection{EntityID: "D-1", DetectedAt: model.FormatTime(firstSeenNS + 200), Label: "bearing"},
		model.Detection{EntityID: "D-1", DetectedAt: model.FormatTime(firstSeenNS + 900), Label: "wear"})
	ev.Verdict.Actions = append(ev.Verdict.Actions, actionOn("D-1", "c1"))
	ev.Calls = append(ev.Calls, callOn("D-1", "c1"), appliedCall("stop_drive", "D-3", "c2"))
	return ev
}

// The digests were taken before a label was keyed by entity and fault (M4 row 2): with one label per entity a run must
// score bit for bit as it did.
func TestOneLabelPerEntityScoresBitForBitAsBeforeLabelsWereKeyedByFault(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		ev     Evidence
		labels []*model.GroundTruthRecord
		want   string
	}{
		{"two pumps", twoFaultRun(), []*model.GroundTruthRecord{labelOn("pump-1"), labelOn("pump-2")},
			"a372654aaba856df4ffb8a926959c99e1fefff8da8e47cfd0f1956dec8d8d9b6"},
		{"a drive trip and a worn drive", mixedVerdictRun(), []*model.GroundTruthRecord{tripLabel(), labelOn("D-1")},
			"34e0a4760988defcba0bbba350ce2d2d2a6902daa5437968ab135cbb59b4719e"},
		{"one label", mixedVerdictRun(), []*model.GroundTruthRecord{tripLabel()},
			"633077a9280509619ee3b9af334ab32e0da407246fbddcdb3823bee44b7270fc"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			graded, err := ScoreRun(tc.ev, tc.labels)
			if err != nil {
				t.Fatal(err)
			}
			if got := runDigest(t, graded); got != tc.want {
				t.Errorf("%s: run score digest %s, want %s", tc.name, got, tc.want)
			}
		})
	}
}

const (
	frozenFault = "conveyor_drive__vibration__frozen"
	bearingLoss = "conveyor_drive__lube_loss"
	frozenAtNS  = 1_000e9
	bearingNS   = 3_000e9
)

// frozenBearingRun is S5-A in miniature: D-3's vibration reading freezes at 1000 s, its bearing loses lubrication and is
// observable from 3000 s; the consumer reports the frozen channel at 2000 s and the bearing at 5000 s, then cuts the feed
// and stops D-3.
func frozenBearingRun() (Evidence, []*model.GroundTruthRecord) {
	frozen := &model.GroundTruthRecord{ScenarioID: "s5-a/1", Domain: "belt", EntityID: "D-3", Label: frozenFault, ExpectedEpisode: true,
		InjectionTimeNS: frozenAtNS, FirstObservableTimeNS: frozenAtNS}
	bearing := &model.GroundTruthRecord{ScenarioID: "s5-a/1", Domain: "belt", EntityID: "D-3", Label: bearingLoss, ExpectedEpisode: true,
		InjectionTimeNS: 2_500e9, FirstObservableTimeNS: bearingNS, DeadlineNS: 9_000e9,
		ExpectedActions: []model.ExpectedAction{{EntityID: "F-1", Effector: "set_feed"}, {EntityID: "D-3", Effector: "stop_drive"}}}
	feed, stop := appliedCall("set_feed", "F-1", "c1"), appliedCall("stop_drive", "D-3", "c2")
	feed.AtNS, stop.AtNS = 6_000e9, 6_100e9
	ev := Evidence{RunID: "run-s5", Ledger: []model.LedgerRecord{}, Verdict: &model.Verdict{
		Detections: []model.Detection{{EntityID: "D-3", DetectedAt: model.FormatTime(2_000e9), Label: frozenFault},
			{EntityID: "D-3", DetectedAt: model.FormatTime(5_000e9), Label: bearingLoss}},
		Actions: []model.Action{decidedAction("set_feed", "F-1", "c1"), decidedAction("stop_drive", "D-3", "c2")}},
		Calls: []world.EffectorCall{feed, stop}}
	return ev, []*model.GroundTruthRecord{frozen, bearing}
}

func detectedAtNS(card *Scorecard) int64 {
	return card.GroundTruth.FirstObservableTimeNS + card.Judgment.DetectionLatencyNS
}

func TestTwoLabelsOnOneEntityAreEachJudgedOnTheDetectionsAndCallsThatNameTheirFault(t *testing.T) {
	t.Parallel()
	ev, labels := frozenBearingRun()
	graded, err := ScoreRun(ev, labels)
	if err != nil {
		t.Fatal(err)
	}
	stuck, bearing := graded.Scorecards[0], graded.Scorecards[1]
	if j := stuck.Judgment; !j.Detected || !j.LabelCorrect || j.Suspicious || j.DetectionCount != 1 || j.DetectionLatencyNS != 1_000e9 {
		t.Fatalf("frozen channel judgment %+v, want its own report alone, 1000 s after the freeze", j)
	}
	if j := bearing.Judgment; !j.Detected || !j.LabelCorrect || j.Suspicious || j.DetectionCount != 1 || j.DetectionLatencyNS != 2_000e9 {
		t.Fatalf("bearing judgment %+v, want its own report alone (the earlier frozen-channel report is not a premature bearing call)", j)
	}
	if stuck.Loop.EffectCalls != 0 || bearing.Loop.EffectCalls != 2 || !bearing.Loop.ActionAppropriate || !bearing.Loop.DeadlineAdhered {
		t.Fatalf("loops %+v and %+v, want the feed cut and the stop of D-3 with the bearing, which expects them", stuck.Loop, bearing.Loop)
	}
	if detectedAtNS(stuck) >= detectedAtNS(bearing) {
		t.Fatalf("the scorecards place the frozen channel at %d ns and the bearing at %d ns, want the channel first", detectedAtNS(stuck), detectedAtNS(bearing))
	}
	if len(graded.Unexpected.Detections) != 0 || len(graded.Unexpected.Actions) != 0 || graded.Unexpected.EffectCalls != 0 {
		t.Fatalf("unexpected = %+v, want nothing", graded.Unexpected)
	}
}

func TestWhatNamesNoneOfAnEntitysFaultsIsJudgedWithEachOfItsLabels(t *testing.T) {
	t.Parallel()
	ev, labels := frozenBearingRun()
	ev.Verdict.Detections = append(ev.Verdict.Detections, model.Detection{EntityID: "D-3", DetectedAt: model.FormatTime(4_000e9), Label: "overheating"})
	reset := appliedCall("reset_trip", "D-3", "c3")
	ev.Verdict.Actions, ev.Calls = append(ev.Verdict.Actions, decidedAction("reset_trip", "D-3", "c3")), append(ev.Calls, reset)
	graded, err := ScoreRun(ev, labels)
	if err != nil {
		t.Fatal(err)
	}
	stuck, bearing := graded.Scorecards[0], graded.Scorecards[1]
	if stuck.Judgment.DetectionCount != 2 || bearing.Judgment.DetectionCount != 2 || stuck.Loop.EffectCalls != 1 || bearing.Loop.EffectCalls != 3 {
		t.Fatalf("judgments %+v %+v, loops %+v %+v: want the unnamed report and the reset with both labels", stuck.Judgment, bearing.Judgment, stuck.Loop, bearing.Loop)
	}
	if !stuck.Judgment.LabelCorrect || bearing.Judgment.LabelCorrect || bearing.Judgment.DetectionLatencyNS != 1_000e9 {
		t.Fatalf("judgments %+v %+v, want the frozen channel's own report first and the bearing's first report the unnamed one", stuck.Judgment, bearing.Judgment)
	}
}
