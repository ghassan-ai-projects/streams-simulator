package domain

import (
	"fmt"
	"sort"

	"github.com/ghassan-ai-projects/streams-simulator/internal/audit"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/randutil"
)

func finalizeScenario(spec *domain.Compiled, sc *Scenario, label *model.GroundTruthRecord, verdict *audit.Verdict) {
	label.TrivialBaselineVerdict = verdictTrivialString(verdict)
	if verdict != nil {
		label.TrivialBaselineDetail = verdict.Scores
	}
	sc.CommandLog = buildCommandLog(spec, sc)
}

func (g *suiteGeneration) drawOnset() (int64, bool) {

	degraded := g.rng.Float64() < g.cfg.Domain.Spec.GroundTruth.PreDegradedFraction
	var onset int64
	if degraded {
		onset = g.startNS - int64(g.rng.Float64()*float64(g.durationNS/4))
	} else {
		onset = g.startNS + int64(g.rng.Float64()*float64(g.durationNS/2))
	}
	if onset < 0 {
		onset = 0
	}
	return onset, degraded
}

func (g *suiteGeneration) scenarioIdentity(seed, idx uint64, entity, fault string, onset int64, degraded bool, perturbations []Perturbation) *Scenario {
	return &Scenario{ID: fmt.Sprintf("%s/%04d", g.cfg.Domain.Spec.ID, idx), Seed: seed, Profile: g.cfg.Profile,
		EntityID: entity, Fault: fault, StartNS: g.startNS, OnsetNS: onset, DurationNS: g.durationNS,
		PreDegraded: degraded, Perturbations: perturbations}
}

func weightedFault(prof *model.Profile, rng *randutil.SplitMix64) (string, bool) {
	ids, total := orderedFaultWeights(prof)
	if total > 0 {
		draw := rng.Float64() * total
		for _, id := range ids {
			weight := prof.FaultWeights[id]
			if draw < weight {
				return id, true
			}
			draw -= weight
		}
	}
	return "", false
}

func orderedFaultWeights(prof *model.Profile) ([]string, float64) {
	ids := make([]string, 0, len(prof.FaultWeights))
	// determinism-safe: sort before accumulating or selecting weights.
	for id := range prof.FaultWeights {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	total := 0.0
	for _, id := range ids {
		total += prof.FaultWeights[id]
	}
	return ids, total
}

func uniformPositiveFault(cfg Config, rng *randutil.SplitMix64) string {
	var positives []string
	for _, fault := range declaredFaults(cfg.Domain) {
		if !fault.IsNegativeClass {
			positives = append(positives, fault.ID)
		}
	}
	if len(positives) == 0 {
		return ""
	}
	sort.Strings(positives)
	return positives[rng.Intn(len(positives))]
}

func scenarioSetupCall(effector, command string, args map[string]any, index int, entity string, at int64) model.SetupCall {
	resolved := substituteEntity(args, entity)
	if command == "" {
		command = fmt.Sprintf("setup-%d", index)
	}
	return model.SetupCall{Effector: effector, EntityID: entity, CommandID: command, Args: resolved, AtNS: at}
}
