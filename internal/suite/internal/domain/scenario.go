package domain

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/audit"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/randutil"
	"github.com/ghassan-ai-projects/streams-simulator/internal/truth"
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
	return g.gradeScenario(scenario, label)
}

// sealScenarioLabel identifies candidates with no observable signal before audit.
func (g *suiteGeneration) sealScenarioLabel(sc *Scenario) (*model.GroundTruthRecord, error) {
	var descriptions []string
	for _, p := range sc.Perturbations {
		descriptions = append(descriptions, p.Name+"@"+fmt.Sprint(p.Params["rate"]))
	}
	label, err := truth.BuildRecord(g.cfg.Domain, g.solver, truth.Injection{
		ScenarioID: sc.ID, Seed: sc.Seed, EntityID: sc.EntityID, FaultID: sc.Fault,
		OnsetNS: sc.OnsetNS, StartNS: sc.StartNS, EntityIDs: g.entities,
		PreDegraded: sc.PreDegraded, Perturbations: descriptions, Setup: sc.Setup,
	})
	if err != nil {
		return nil, fmt.Errorf("suite: build label: %w", err)
	}
	return label, nil
}

// auditScenario evaluates the candidate's delivered, perturbed stream.
func (g *suiteGeneration) auditScenario(sc *Scenario) (*audit.Verdict, error) {
	verdict, err := g.panel.Audit(sc.EntityID, sc.Fault, sc.OnsetNS, sc.StartNS, g.entities, sc.DurationNS, sc.Setup, sc.Perturbations)
	if err != nil {
		return nil, fmt.Errorf("suite: audit: %w", err)
	}
	return verdict, nil
}

func (g *suiteGeneration) drawScenario(seed, idx uint64, sampleFrac float64) *Scenario {
	negative := g.rng.Float64() < sampleFrac
	fault := g.suite.pickFault(g.cfg, g.prof, g.rng, negative)
	entity := g.entities[g.rng.Intn(len(g.entities))]
	onset, preDegraded := g.drawOnset()
	perturbations := drawPerturbations(g.rng, g.perturbCount, g.startNS, g.durationNS)
	scenario := g.scenarioIdentity(seed, idx, entity, fault, onset, preDegraded, perturbations)
	// Changing context can mask a pre-existing fault; omit setup then.
	if !preDegraded {
		scenario.Setup = g.suite.defaultSetup(g.cfg.Domain, entity, g.startNS)
	}
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
		if fault, found := negativeFault(cfg); found {
			return fault
		}
	}
	if len(prof.FaultWeights) > 0 {
		if fault, found := weightedFault(prof, rng); found {
			return fault
		}
	}
	return uniformPositiveFault(cfg, rng)
}

// defaultSetup translates profile data into scenario context calls. The
// simulator never branches on a domain id or effector name here.
func (s *Suite) defaultSetup(spec *domain.Compiled, entity string, startNS int64) []model.SetupCall {
	prof := spec.Profile(s.Profile)
	if prof == nil {
		return nil
	}
	var out []model.SetupCall
	for i, setup := range prof.Setup {
		if setup.Effector == "" || spec.Effector(setup.Effector) == nil {
			continue
		}
		out = append(out, scenarioSetupCall(setup.Effector, setup.CommandID, setup.Args, i, entity, startNS))
	}
	return out
}

func (g *suiteGeneration) gradeScenario(scenario *Scenario, label *model.GroundTruthRecord) (*Scenario, *model.GroundTruthRecord, *audit.Verdict, error) {
	verdict, err := g.auditScenario(scenario)
	if err != nil {
		return nil, nil, nil, err
	}
	finalizeScenario(g.cfg.Domain, scenario, label, verdict)
	return scenario, label, verdict, nil
}
