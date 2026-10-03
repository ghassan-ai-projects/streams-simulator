package truth

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

func recordIdentity(domain, scenario string, seed uint64, entity, fault string) *model.GroundTruthRecord {
	return &model.GroundTruthRecord{
		ScenarioID: scenario,
		Domain:     domain,
		Seed:       seed,
		EntityID:   entity,
		Label:      fault,
	}
}

func attachObservability(rec *model.GroundTruthRecord, fault *model.Fault, res *Result, onsetNS int64) {
	rec.InjectionTimeNS = onsetNS
	rec.FirstObservableTimeNS = res.FirstObservableNS
	rec.UnavoidableTimeNS = res.UnavoidableNS
	rec.Observability = model.ObservabilityInfo{
		DetectorForm:       fault.Observability.Detector.Form,
		Channels:           res.Channels,
		EffectiveSigma:     res.EffectiveSigma,
		FirstObservableSNR: fault.Observability.FirstObservableSNR,
		UnavoidableSNR:     fault.Observability.UnavoidableSNR,
		SolutionMethod:     res.Method,
	}
}

func attachScenarioContext(rec *model.GroundTruthRecord, fault *model.Fault, preDegraded bool, perturbations []string) {
	rec.IsNegativeClass = fault.IsNegativeClass
	rec.ExpectedEpisode = !fault.IsNegativeClass
	rec.TrivialBaselineVerdict = model.TrivialNonTrivial
	rec.PreDegraded = preDegraded
	rec.Perturbations = perturbations
	rec.ExpectedEffector = fault.ExpectedEffector
	if fault.DeadlineS > 0 && rec.FirstObservableTimeNS > 0 {
		rec.DeadlineNS = rec.FirstObservableTimeNS + int64(fault.DeadlineS*1e9)
	}
	attachCounterfactual(rec, fault.Counterfactual)
}

func attachCounterfactual(rec *model.GroundTruthRecord, counterfactual *model.Counterfactual) {
	if counterfactual != nil {
		rec.Counterfactual = &model.Counterfactual{
			IfNoAction:         counterfactual.IfNoAction,
			IfActionByDeadline: counterfactual.IfActionByDeadline,
		}
	}
}
