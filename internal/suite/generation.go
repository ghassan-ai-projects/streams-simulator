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
	if cfg.N <= 0 {
		cfg.N = 100
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 3 * cfg.N
	}
	prof := cfg.Domain.Profile(cfg.Profile)
	if prof == nil {
		return nil, fmt.Errorf("suite: unknown profile %q", cfg.Profile)
	}
	rng := randutil.NewSplitMix64(cfg.Seed ^ randutil.Fnv1a64(cfg.Domain.Spec.ID+"/"+cfg.Profile))
	gt := cfg.Domain.Spec.GroundTruth
	negativeFrac := gt.NegativeClassFraction
	if negativeFrac <= 0 {
		negativeFrac = 0.4
	}
	solver := truth.NewSolver(cfg.Domain, cfg.Seed, cfg.SampleNS, 72*3600*1e9)
	panel := audit.NewPanel(cfg.Domain, cfg.Seed^0x5eed, cfg.SampleNS)

	s := &Suite{
		DomainID: cfg.Domain.Spec.ID,
		Profile:  cfg.Profile,
		Seed:     cfg.Seed,
	}
	perturbCount := map[string]int{}

	startNS := model.DefaultStartTimeNS + 4*3600*1e9
	durationNS := int64(prof.DurationS * 1e9)
	if durationNS <= 0 {
		durationNS = 72 * 3600 * 1e9
	}
	entities := defaultEntities(cfg.Domain, prof)

	return &suiteGeneration{
		cfg: cfg, prof: prof, rng: rng, solver: solver, panel: panel, suite: s,
		perturbCount: perturbCount, negativeFrac: negativeFrac,
		startNS: startNS, durationNS: durationNS, entities: entities,
	}, nil
}

func (g *suiteGeneration) populate() error {
	for len(g.suite.Scenarios) < g.cfg.N && g.suite.Attempts < g.cfg.MaxAttempts {
		g.suite.Attempts++
		// Attempt identity, not admitted-count identity, enters the seed and
		// scenario id. Rejected candidates must never be replayed under the
		// same identity as a later admitted candidate.
		currentAttempt := g.attemptIndex
		g.attemptIndex++
		scenarioSeed := g.cfg.Seed + currentAttempt*0x9e3779b97f4a7c15
		// Adaptive negative sampling: the audit rejects positive scenarios
		// aggressively, so the admission fraction drifts above the declared
		// band; proportional control on the sampling probability holds the
		// band when the domain can sustain it, and the terminal state
		// reports when it cannot (G-06).
		sampleFrac := g.negativeSamplingFraction()
		sc, label, verdict, err := g.buildScenario(scenarioSeed, currentAttempt, sampleFrac)
		if err != nil {
			return fmt.Errorf("streamsim: %w", err)
		}
		if verdict != nil && verdict.Trivial {
			// Kept as a mechanism fixture, excluded from the graded set;
			// the loop keeps going for a replacement.
			g.suite.TrivialExcluded = append(g.suite.TrivialExcluded, TrivialCase{
				ScenarioID: sc.ID, Fault: sc.Fault, Scores: verdict.Scores, Best: verdict.Best,
			})
			continue
		}
		g.admitScenario(sc, label)
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
	for _, p := range sc.Perturbations {
		g.perturbCount[p.Name]++
	}
	if label.IsNegativeClass {
		g.negatives++
	}
	if sc.PreDegraded {
		g.preDegraded++
	}
	if sc.Profile == "correlated_cascade" {
		g.cascadeCount++
	}
	if sc.Profile == "sensor_pathology" {
		g.pathologyCount++
	}
}

func (g *suiteGeneration) reportShortfalls() {
	// Perturbation coverage: every catalog perturbation in >= 5 scenarios.
	var shortfalls []string
	if missing := uncovered(g.perturbCount); len(missing) > 0 {
		shortfalls = append(shortfalls, fmt.Sprintf("perturbation coverage not met at the declared size: %v", missing))
	}
	// Composition checks. All shortfalls are aggregated into one reportable
	// terminal state (G-06): a domain that cannot sustain the declared
	// composition says so, rather than silently degrading.
	negShare := float64(g.negatives) / float64(len(g.suite.Scenarios))
	if len(g.suite.Scenarios) > 0 && (negShare < 0.35 || negShare > 0.45) {
		shortfalls = append(shortfalls, fmt.Sprintf("negative-class fraction %.2f outside [0.35, 0.45]", negShare))
	}
	if len(g.suite.Scenarios) > 0 && g.preDegraded*100/len(g.suite.Scenarios) < 10 {
		shortfalls = append(shortfalls, "pre-degraded fraction below 10%")
	}
	if g.cascadeCount < 3 {
		shortfalls = append(shortfalls, fmt.Sprintf("correlated_cascade scenarios: %d < 3", g.cascadeCount))
	}
	if g.pathologyCount < 10 {
		shortfalls = append(shortfalls, fmt.Sprintf("sensor_pathology scenarios: %d < 10", g.pathologyCount))
	}
	if g.suite.Attempts >= g.cfg.MaxAttempts && len(g.suite.Scenarios) < g.cfg.N {
		shortfalls = append(shortfalls, fmt.Sprintf("audit rejected too many scenarios: %d non-trivial of %d attempts", len(g.suite.Scenarios), g.suite.Attempts))
	}
	g.suite.TerminalState = joinShortfalls(shortfalls)
}
