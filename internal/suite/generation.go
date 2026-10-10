package suite

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/audit"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/randutil"
	"github.com/ghassan-ai-projects/streams-simulator/internal/truth"
)

type suiteGeneration struct {
	cfg                 Config
	prof                *model.Profile
	rng                 *randutil.SplitMix64
	solver              *truth.Solver
	panel               *audit.Panel
	suite               *Suite
	perturbCount        map[string]int
	negatives           int
	preDegraded         int
	cascadeCount        int
	pathologyCount      int
	attemptIndex        uint64
	negativeFrac        float64
	startNS, durationNS int64
	entities            []string
}

func newSuiteGeneration(cfg Config) (*suiteGeneration, error) {
	cfg = generationDefaults(cfg)
	prof := cfg.Domain.Profile(cfg.Profile)
	if prof == nil {
		return nil, fmt.Errorf("suite: unknown profile %q", cfg.Profile)
	}
	g, err := prepareGeneration(cfg, prof)
	if err != nil {
		return nil, err
	}
	g.startNS = model.DefaultStartTimeNS + 4*3600*1e9
	g.durationNS = profileDuration(prof)
	g.entities = defaultEntities(cfg.Domain, prof)
	return g, nil
}

func (g *suiteGeneration) populate() error {
	for len(g.suite.Scenarios) < g.cfg.N && g.suite.Attempts < g.cfg.MaxAttempts {
		if err := g.attemptScenario(); err != nil {
			return err
		}
	}
	return nil
}

func (g *suiteGeneration) negativeSamplingFraction() float64 {
	admitted := 0.0
	if len(g.suite.Scenarios) > 0 {
		admitted = float64(g.negatives) / float64(len(g.suite.Scenarios))
	}
	sampleFrac := g.negativeFrac + (g.negativeFrac-admitted)*2.0
	if sampleFrac < 0.1 {
		sampleFrac = 0.1
	}
	if sampleFrac > 0.7 {
		sampleFrac = 0.7
	}
	return sampleFrac
}

func (g *suiteGeneration) admitScenario(sc *Scenario, label *model.GroundTruthRecord) {
	g.suite.Scenarios = append(g.suite.Scenarios, *sc)
	g.suite.Labels = append(g.suite.Labels, *label)
	for _, perturbation := range sc.Perturbations {
		g.perturbCount[perturbation.Name]++
	}
	g.countComposition(sc, label)
}

func (g *suiteGeneration) reportShortfalls() {
	// Aggregate composition and coverage failures into one terminal state.
	var shortfalls []string
	if missing := uncovered(g.perturbCount); len(missing) > 0 {
		shortfalls = append(shortfalls, fmt.Sprintf("perturbation coverage not met at the declared size: %v", missing))
	}
	shortfalls = g.compositionShortfalls(shortfalls)
	shortfalls = g.profileShortfalls(shortfalls)
	if g.suite.Attempts >= g.cfg.MaxAttempts && len(g.suite.Scenarios) < g.cfg.N {
		shortfalls = append(shortfalls, fmt.Sprintf("audit rejected too many scenarios: %d non-trivial of %d attempts", len(g.suite.Scenarios), g.suite.Attempts))
	}
	g.suite.TerminalState = joinShortfalls(shortfalls)
}
