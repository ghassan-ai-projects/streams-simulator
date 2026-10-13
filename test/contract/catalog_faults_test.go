package contract

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/audit"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/truth"
)

const (
	catalogFaultOnsetNS = 3600e9
	catalogHorizonNS    = 24 * 3600e9
	catalogAuditNS      = 8 * 3600e9
	catalogAuditOnsetNS = 2 * 3600e9
	catalogSampleStepNS = 60e9
)

func checkCatalogFaultsObservable(t *testing.T, p catalogPlant) {
	solver, err := truth.NewSolver(p.spec, catalogSeed, catalogSampleStepNS, catalogHorizonNS)
	if err != nil {
		t.Fatal(err)
	}
	for _, fault := range p.kind().Faults {
		if fault.IsNegativeClass {
			continue
		}
		record, err := truth.BuildRecord(p.spec, solver, truth.Injection{ScenarioID: "catalog/" + fault.ID, Seed: catalogSeed,
			EntityID: p.entities[0], FaultID: fault.ID, OnsetNS: model.DefaultStartTimeNS + catalogFaultOnsetNS,
			StartNS: model.DefaultStartTimeNS, EntityIDs: p.entities})
		if err != nil {
			t.Fatalf("%s: %v", fault.ID, err)
		}
		if record.FirstObservableTimeNS == 0 {
			t.Errorf("%s is not observable within 24 h (detector %s on %v, first SNR %v)", fault.ID,
				record.Observability.DetectorForm, record.Observability.Channels, record.Observability.FirstObservableSNR)
		}
	}
}

func checkCatalogAudit(t *testing.T, p catalogPlant) {
	panel, err := audit.NewPanel(p.spec, catalogSeed^0x5eed, catalogSampleStepNS)
	if err != nil {
		t.Fatal(err)
	}
	own, strict := auditedFaults(p)
	nonTrivial := 0
	for _, fault := range own {
		start := model.DefaultStartTimeNS
		verdict, err := panel.Audit(p.entities[0], fault, start+catalogAuditOnsetNS, start, p.entities, catalogAuditNS, nil, nil)
		if err != nil {
			t.Fatalf("%s: %v", fault, err)
		}
		t.Logf("%s: trivial=%v best=%s %.3f", fault, verdict.Trivial, verdict.Best, verdict.BestScore)
		if !verdict.Trivial {
			nonTrivial++
		}
	}
	if strict && nonTrivial == 0 {
		t.Errorf("every audited fault of the equipment (%v) is trivial for a one-line baseline; at least one must be hard to detect", own)
	}
}

// auditedFaults lists the equipment's own faults that are not negative class; a part with none is audited, without the
// requirement to be hard, on its first other fault.
func auditedFaults(p catalogPlant) (faults []string, strict bool) {
	own := equipmentFaultIDs(p)
	var others []string
	for _, fault := range p.kind().Faults {
		switch {
		case fault.IsNegativeClass:
		case own[fault.ID]:
			faults = append(faults, fault.ID)
		default:
			others = append(others, fault.ID)
		}
	}
	if len(faults) == 0 && len(others) > 0 {
		return others[:1], false
	}
	return faults, true
}

func equipmentFaultIDs(p catalogPlant) map[string]bool {
	own := map[string]bool{}
	kind, _ := p.doc["kind"].(map[string]any)
	faults, _ := kind["faults"].([]any)
	prefix := p.kind().Name + "__"
	for _, raw := range faults {
		own[prefix+raw.(map[string]any)["id"].(string)] = true
	}
	return own
}
