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
func (g *suiteGeneration) buildScenario(seed, idx uint64, sampleFrac float64) (*Scenario, *model.GroundTruthRecord, *audit.Verdict, error) {
	scenario := g.drawScenario(seed, idx, sampleFrac)
	label, err := g.sealScenarioLabel(scenario)
	if err != nil {
		return nil, nil, nil, err
	}
	if label.FirstObservableTimeNS == 0 {
		return scenario, label, &audit.Verdict{Trivial: true}, nil
	}

	verdict, err := g.auditScenario(scenario)
	if err != nil {
		return nil, nil, nil, err
	}

	label.TrivialBaselineVerdict = verdictTrivialString(verdict)
	if verdict != nil {
		label.TrivialBaselineDetail = verdict.Scores
	}
	// The executable command log.
	scenario.CommandLog = buildCommandLog(g.cfg.Domain, scenario)
	return scenario, label, verdict, nil
}

// sealScenarioLabel identifies candidates with no observable signal before audit.
func (g *suiteGeneration) sealScenarioLabel(sc *Scenario) (*model.GroundTruthRecord, error) {
	var descriptions []string
	for _, p := range sc.Perturbations {
		descriptions = append(descriptions, p.Name+"@"+fmt.Sprint(p.Params["rate"]))
	}
	label, err := truth.BuildRecord(g.cfg.Domain, g.solver, sc.ID, sc.Seed, sc.EntityID, sc.Fault,
		sc.OnsetNS, sc.StartNS, g.entities, sc.PreDegraded, descriptions, sc.Setup)
	if err != nil {
		return nil, fmt.Errorf("suite: build label: %w", err)
	}
	return label, nil
}

// auditScenario evaluates the candidate's delivered, perturbed stream.
func (g *suiteGeneration) auditScenario(sc *Scenario) (*audit.Verdict, error) {
	var perturbations []audit.Perturbation
	for _, p := range sc.Perturbations {
		perturbations = append(perturbations, audit.Perturbation{Name: p.Name, Params: p.Params, FromNS: p.FromNS, UntilNS: p.UntilNS})
	}
	verdict, err := g.panel.Audit(sc.EntityID, sc.Fault, sc.OnsetNS, sc.StartNS, g.entities, sc.DurationNS, sc.Setup, perturbations)
	if err != nil {
		return nil, fmt.Errorf("suite: audit: %w", err)
	}
	return verdict, nil
}

func (g *suiteGeneration) drawScenario(seed, idx uint64, sampleFrac float64) *Scenario {
	gt := g.cfg.Domain.Spec.GroundTruth
	isNegative := g.rng.Float64() < sampleFrac
	faultID := g.suite.pickFault(g.cfg, g.prof, g.rng, isNegative)

	entity := g.entities[g.rng.Intn(len(g.entities))]
	profileName := g.cfg.Profile
	preDeg := false
	var onsetNS int64
	// Randomized onset: uniform across the first half of the trace, so the
	// fault has room to develop and run-to-failure bias is countered. 10-15%
	// of scenarios begin with the asset already degraded (fault before t0).
	if g.rng.Float64() < gt.PreDegradedFraction {
		preDeg = true
		onsetNS = g.startNS - int64(g.rng.Float64()*float64(g.durationNS/4))
	} else {
		onsetNS = g.startNS + int64(g.rng.Float64()*float64(g.durationNS/2))
	}
	if onsetNS < 0 {
		onsetNS = 0
	}

	// Perturbations: sample 0-2, with forced coverage for thin ones.
	scPerts := drawPerturbations(g.rng, g.perturbCount, g.startNS, g.durationNS)
	scenario := &Scenario{
		ID:            fmt.Sprintf("%s/%04d", g.cfg.Domain.Spec.ID, idx),
		Seed:          seed,
		Profile:       profileName,
		EntityID:      entity,
		Fault:         faultID,
		StartNS:       g.startNS,
		OnsetNS:       onsetNS,
		DurationNS:    g.durationNS,
		PreDegraded:   preDeg,
		Perturbations: scPerts,
	}

	// Profile setup supplies the initial context. Pre-degraded starts omit
	// setup because changing the context could mask the earlier fault.
	var setup []truth.SetupCall
	if !preDeg {
		setup = g.suite.defaultSetup(g.cfg.Domain, entity, g.startNS)
	}
	scenario.Setup = setup
	return scenario
}

func drawPerturbations(rng *randutil.SplitMix64, perturbCount map[string]int, startNS, durationNS int64) []Perturbation {
	scPerts := []Perturbation{}
	for _, name := range pickPerturbations(rng, perturbCount) {
		params := samplePerturbParams(rng, name)
		from := startNS + int64(rng.Float64()*float64(durationNS/2))
		until := from + int64((0.25+rng.Float64()*0.5)*float64(durationNS/2))
		scPerts = append(scPerts, Perturbation{Name: name, Params: params, FromNS: from, UntilNS: until})
	}

	return scPerts
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
