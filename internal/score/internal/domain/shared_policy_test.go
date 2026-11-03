package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func TestJudgmentUsesEarliestDetectionRegardlessOfOrder(t *testing.T) {
	t.Parallel()
	gt := &model.GroundTruthRecord{EntityID: "entity", Label: "fault", FirstObservableTimeNS: 1000}
	detections := []model.Detection{
		{EntityID: "entity", Label: "wrong", DetectedAt: model.FormatTime(1300)},
		{EntityID: "entity", Label: "fault", DetectedAt: model.FormatTime(1100)},
	}
	want := JudgmentMetrics{Detected: true, LabelCorrect: true, DetectionLatencyNS: 100, DetectionCount: 2}
	for _, order := range [][]model.Detection{detections, {detections[1], detections[0]}} {
		got := judgmentFrom(&model.Verdict{Detections: order}, gt)
		if got != want {
			t.Fatalf("judgment=%+v, want %+v", got, want)
		}
	}
}

func TestSharedConsumerPoliciesRespectPerturbationAndEvidence(t *testing.T) {
	t.Parallel()
	ledger := []model.LedgerRecord{{Seq: 0, Delivered: true, DeliveryID: 1}}
	for _, tc := range []struct {
		name                string
		verdict             *model.Verdict
		perturbations       []string
		conflict, grounding bool
	}{
		{"inactive conflict", &model.Verdict{}, nil, true, true},
		{"active conflict absent", &model.Verdict{}, []string{"id_reuse"}, false, true},
		{"active conflict reported", &model.Verdict{Admission: []model.Admission{{Seq: 0, Outcome: model.AdmissionConflict}}}, []string{"id_reuse"}, true, true},
		{"malformed citation", &model.Verdict{Detections: []model.Detection{{EvidenceRefs: []string{"unknown"}}}}, nil, true, false},
		{"undelivered citation", &model.Verdict{Detections: []model.Detection{{EvidenceRefs: []string{"seq:1"}}}}, nil, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := consumerFrom(tc.verdict, ledger, nil, tc.perturbations)
			if got.IdentityConflict != tc.conflict || got.EvidenceGrounding != tc.grounding {
				t.Fatalf("consumer=%+v", got)
			}
		})
	}
}

func TestLoopRequiresScenarioEntity(t *testing.T) {
	t.Parallel()
	gt := &model.GroundTruthRecord{EntityID: "scenario", ExpectedEffector: "recover"}
	calls := []world.EffectorCall{{Effector: "recover", EntityID: "other", EffectApplied: true}}
	if got := loopFrom(&model.Verdict{}, gt, calls, ""); got.ActionAppropriate {
		t.Fatalf("other entity was credited: %+v", got)
	}
}

func TestAvailableEmptyLedgerDiffersFromUnavailableEvidence(t *testing.T) {
	t.Parallel()
	v := &model.Verdict{}
	if got := consumerFrom(v, nil, nil, nil); got != (ConsumerMetrics{}) {
		t.Fatalf("unavailable ledger: %+v", got)
	}
	got := consumerFrom(v, []model.LedgerRecord{}, nil, nil)
	if !got.DuplicateHandling || !got.EvidenceGrounding || !got.ActionFidelity {
		t.Fatalf("available empty ledger: %+v", got)
	}
}
