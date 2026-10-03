package score

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func consumer(r *run.Run, _ *model.GroundTruthRecord) ConsumerMetrics {
	ledger := r.Ledger()
	if ledger == nil {
		ledger = []model.LedgerRecord{}
	}
	return consumerFrom(r.Verdict(), ledger, r.World.EffectorCalls(), r.AppliedPerturbations())
}

func consumerFrom(v *model.Verdict, ledger []model.LedgerRecord, calls []world.EffectorCall, perturbations []string) ConsumerMetrics {
	if ledger == nil {
		return ConsumerMetrics{}
	}
	m := admissionMetrics(v.Admission, ledger, perturbations)
	m.DroppedEventDetection = detectsEachDrop(v.Detections, ledger)
	m.EvidenceGrounding = groundsEvidence(v.Detections, ledger)
	m.ActionFidelity = matchesActions(v.Actions, calls)
	m.InterlockHandling = respectsInterlocks(calls)
	return m
}
