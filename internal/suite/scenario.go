package suite

import (
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/audit"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/randutil"
	"github.com/ghassan-ai-projects/streams-simulator/internal/truth"
	"sort"
)

// buildScenario constructs one candidate scenario and its audit verdict.
func (s *Suite) buildScenario(cfg Config, rng *randutil.SplitMix64, solver *truth.Solver, panel *audit.Panel,
	prof *model.Profile, entities []string, startNS, durationNS int64, seed, idx uint64,
	perturbCount map[string]int, sampleFrac float64) (*Scenario, *model.GroundTruthRecord, *audit.Verdict, error) {

	gt := cfg.Domain.Spec.GroundTruth
	isNegative := rng.Float64() < sampleFrac
	faultID := s.pickFault(cfg, prof, rng, isNegative)

	entity := entities[rng.Intn(len(entities))]
	profileName := cfg.Profile
	preDeg := false
	var onsetNS int64
	// Randomized onset: uniform across the first half of the trace, so the
	// fault has room to develop and run-to-failure bias is countered. 10-15%
	// of scenarios begin with the asset already degraded (fault before t0).
	if rng.Float64() < gt.PreDegradedFraction {
		preDeg = true
		onsetNS = startNS - int64(rng.Float64()*float64(durationNS/4))
	} else {
		onsetNS = startNS + int64(rng.Float64()*float64(durationNS/2))
	}
	if onsetNS < 0 {
		onsetNS = 0
	}

	// Perturbations: sample 0-2, with forced coverage for thin ones.
	scPerts := []Perturbation{}
	for _, name := range pickPerturbations(rng, perturbCount) {
		params := samplePerturbParams(rng, name)
		from := startNS + int64(rng.Float64()*float64(durationNS/2))
		until := from + int64((0.25+rng.Float64()*0.5)*float64(durationNS/2))
		scPerts = append(scPerts, Perturbation{Name: name, Params: params, FromNS: from, UntilNS: until})
	}

	scenario := &Scenario{
		ID:            fmt.Sprintf("%s/%04d", cfg.Domain.Spec.ID, idx),
		Seed:          seed,
		Profile:       profileName,
		EntityID:      entity,
		Fault:         faultID,
		StartNS:       startNS,
		OnsetNS:       onsetNS,
		DurationNS:    durationNS,
		PreDegraded:   preDeg,
		Perturbations: scPerts,
	}

	// Setup: aquaculture scenarios start the aerator for the night so
	// aerator_failure is meaningful — except pre-degraded starts, where the
	// fault predates t0 and starting the aerator would mask it.
	var setup []truth.SetupCall
	if !preDeg {
		setup = s.defaultSetup(cfg.Domain, entity, startNS)
	}
	scenario.Setup = setup

	// Seal the label first: unobservable scenarios (no signal crosses the
	// noise floor within the horizon) cannot be graded and are regenerated.
	var pertStrs []string
	for _, p := range scPerts {
		pertStrs = append(pertStrs, p.Name+"@"+fmt.Sprint(p.Params["rate"]))
	}
	label, err := truth.BuildRecord(cfg.Domain, solver, scenario.ID, seed, entity, faultID,
		onsetNS, startNS, entities, preDeg, pertStrs, setup)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("suite: build label: %w", err)
	}
	if label.FirstObservableTimeNS == 0 {
		return scenario, label, &audit.Verdict{Trivial: true}, nil
	}

	// Audit the candidate: trivial scenarios are excluded from the graded
	// set and the loop regenerates. The audit consumes the delivered stream
	// with the scenario's declared perturbations applied.
	var auditPerts []audit.Perturbation
	for _, p := range scPerts {
		auditPerts = append(auditPerts, audit.Perturbation{Name: p.Name, Params: p.Params, FromNS: p.FromNS, UntilNS: p.UntilNS})
	}
	verdict, err := panel.Audit(entity, faultID, onsetNS, startNS, entities, durationNS, setup, auditPerts)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("suite: audit: %w", err)
	}

	label.TrivialBaselineVerdict = verdictTrivialString(verdict)
	if verdict != nil {
		label.TrivialBaselineDetail = verdict.Scores
	}
	// The executable command log.
	scenario.CommandLog = buildCommandLog(cfg.Domain, scenario)
	return scenario, label, verdict, nil
}

func verdictTrivialString(v *audit.Verdict) string {
	if v == nil {
		return model.TrivialNonTrivial
	}
	if v.Trivial {
		return model.TrivialTrivial
	}
	return model.TrivialNonTrivial
}

// pickFault samples a fault id per the profile weights and the negative
// fraction.
func (s *Suite) pickFault(cfg Config, prof *model.Profile, rng *randutil.SplitMix64, isNegative bool) string {
	if isNegative {
		for i := range cfg.Domain.Spec.Faults {
			if cfg.Domain.Spec.Faults[i].IsNegativeClass {
				return cfg.Domain.Spec.Faults[i].ID
			}
		}
	}
	if len(prof.FaultWeights) > 0 {
		total := 0.0
		ids := make([]string, 0, len(prof.FaultWeights))
		for id := range prof.FaultWeights {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			w := prof.FaultWeights[id]
			total += w
		}
		if total > 0 {
			r := rng.Float64() * total
			for _, id := range ids {
				w := prof.FaultWeights[id]
				if r < w {
					return id
				}
				r -= w
			}
		}
	}
	// Uniform over positive faults.
	var positives []string
	for i := range cfg.Domain.Spec.Faults {
		if !cfg.Domain.Spec.Faults[i].IsNegativeClass {
			positives = append(positives, cfg.Domain.Spec.Faults[i].ID)
		}
	}
	if len(positives) == 0 {
		return ""
	}
	sort.Strings(positives)
	return positives[rng.Intn(len(positives))]
}

// defaultSetup translates profile data into scenario context calls. The
// simulator never branches on a domain id or effector name here.
func (s *Suite) defaultSetup(spec *domain.Compiled, entity string, startNS int64) []truth.SetupCall {
	prof := spec.Profile(s.Profile)
	if prof == nil {
		return nil
	}
	var out []truth.SetupCall
	for i, setup := range prof.Setup {
		if setup.Effector == "" || spec.Effector(setup.Effector) == nil {
			continue
		}
		args := substituteEntity(setup.Args, entity)
		commandID := setup.CommandID
		if commandID == "" {
			commandID = fmt.Sprintf("setup-%d", i)
		}
		out = append(out, truth.SetupCall{Effector: setup.Effector, EntityID: entity, CommandID: commandID, Args: args, AtNS: startNS})
	}
	return out
}
