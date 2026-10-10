package domain

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

func consumer(ev Evidence, _ *model.GroundTruthRecord) ConsumerMetrics {
	ledger := ev.Ledger
	if ledger == nil {
		ledger = []model.LedgerRecord{}
	}
	return consumerFrom(ev.Verdict, ledger, ev.Calls, ev.Perturbations)
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
