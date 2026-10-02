// Package suite generates graded scenario suites: declared negative-class
// fraction, randomized onset (including pre-degraded starts), perturbation
// coverage, and a generate-audit-regenerate loop that keeps only
// non-trivial scenarios. A domain that cannot produce non-trivial scenarios
// at the declared prevalence reports a terminal state rather than an empty
// directory (G-06).
package suite

import (
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/audit"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/randutil"
	"github.com/ghassan-ai-projects/streams-simulator/internal/truth"
)

// Perturbation is one delivery perturbation in a scenario.
type Perturbation struct {
	Name    string         `json:"name"`
	Params  map[string]any `json:"params,omitempty"`
	FromNS  int64          `json:"from_ns,omitempty"`
	UntilNS int64          `json:"until_ns,omitempty"`
}

// Scenario is one executable scenario recipe.
type Scenario struct {
	ID            string            `json:"id"`
	Seed          uint64            `json:"seed"`
	Profile       string            `json:"profile"`
	EntityID      string            `json:"entity_id"`
	Fault         string            `json:"fault"`
	StartNS       int64             `json:"start_ns"`
	OnsetNS       int64             `json:"onset_ns"`
	DurationNS    int64             `json:"duration_ns"`
	PreDegraded   bool              `json:"pre_degraded,omitempty"`
	Perturbations []Perturbation    `json:"perturbations,omitempty"`
	Setup         []truth.SetupCall `json:"setup,omitempty"`
	CommandLog    []model.Command   `json:"command_log"`
}

// TrivialCase is a scenario the audit rejected, kept as a mechanism
// regression fixture.
type TrivialCase struct {
	ScenarioID string             `json:"scenario_id"`
	Fault      string             `json:"fault"`
	Scores     map[string]float64 `json:"scores"`
	Best       string             `json:"best"`
}

// Suite is a generated graded suite.
type Suite struct {
	DomainID        string                    `json:"domain"`
	Profile         string                    `json:"profile"`
	Seed            uint64                    `json:"seed"`
	Scenarios       []Scenario                `json:"scenarios"`
	Labels          []model.GroundTruthRecord `json:"labels"`
	TrivialExcluded []TrivialCase             `json:"trivial_excluded,omitempty"`
	Attempts        int                       `json:"attempts"`
	TerminalState   string                    `json:"terminal_state,omitempty"`
}

// Config pins suite generation.
type Config struct {
	Domain      *domain.Compiled
	Profile     string
	N           int
	Seed        uint64
	SampleNS    int64
	MaxAttempts int
}

// Generate builds a suite with the generate-audit-regenerate loop.
func Generate(cfg Config) (*Suite, error) {
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
	negatives := 0
	preDegraded := 0
	cascadeCount := 0
	pathologyCount := 0
	attemptIndex := uint64(0)

	startNS := model.DefaultStartTimeNS + 4*3600*1e9
	durationNS := int64(prof.DurationS * 1e9)
	if durationNS <= 0 {
		durationNS = 72 * 3600 * 1e9
	}
	entities := defaultEntities(cfg.Domain, prof)

	for len(s.Scenarios) < cfg.N && s.Attempts < cfg.MaxAttempts {
		s.Attempts++
		// Attempt identity, not admitted-count identity, enters the seed and
		// scenario id. Rejected candidates must never be replayed under the
		// same identity as a later admitted candidate.
		currentAttempt := attemptIndex
		attemptIndex++
		scenarioSeed := cfg.Seed + currentAttempt*0x9e3779b97f4a7c15
		// Adaptive negative sampling: the audit rejects positive scenarios
		// aggressively, so the admission fraction drifts above the declared
		// band; proportional control on the sampling probability holds the
		// band when the domain can sustain it, and the terminal state
		// reports when it cannot (G-06).
		admitted := 0.0
		if len(s.Scenarios) > 0 {
			admitted = float64(negatives) / float64(len(s.Scenarios))
		}
		sampleFrac := negativeFrac + (negativeFrac-admitted)*2.0
		if sampleFrac < 0.1 {
			sampleFrac = 0.1
		}
		if sampleFrac > 0.7 {
			sampleFrac = 0.7
		}
		sc, label, verdict, err := s.buildScenario(cfg, rng, solver, panel, prof, entities,
			startNS, durationNS, scenarioSeed, currentAttempt, perturbCount, sampleFrac)
		if err != nil {
			return nil, fmt.Errorf("streamsim: %w", err)
		}
		if verdict != nil && verdict.Trivial {
			// Kept as a mechanism fixture, excluded from the graded set;
			// the loop keeps going for a replacement.
			s.TrivialExcluded = append(s.TrivialExcluded, TrivialCase{
				ScenarioID: sc.ID, Fault: sc.Fault, Scores: verdict.Scores, Best: verdict.Best,
			})
			continue
		}
		s.Scenarios = append(s.Scenarios, *sc)
		s.Labels = append(s.Labels, *label)
		for _, p := range sc.Perturbations {
			perturbCount[p.Name]++
		}
		if label.IsNegativeClass {
			negatives++
		}
		if sc.PreDegraded {
			preDegraded++
		}
		if sc.Profile == "correlated_cascade" {
			cascadeCount++
		}
		if sc.Profile == "sensor_pathology" {
			pathologyCount++
		}
	}

	// Perturbation coverage: every catalog perturbation in >= 5 scenarios.
	var shortfalls []string
	if missing := uncovered(perturbCount); len(missing) > 0 {
		shortfalls = append(shortfalls, fmt.Sprintf("perturbation coverage not met at the declared size: %v", missing))
	}
	// Composition checks. All shortfalls are aggregated into one reportable
	// terminal state (G-06): a domain that cannot sustain the declared
	// composition says so, rather than silently degrading.
	negShare := float64(negatives) / float64(len(s.Scenarios))
	if len(s.Scenarios) > 0 && (negShare < 0.35 || negShare > 0.45) {
		shortfalls = append(shortfalls, fmt.Sprintf("negative-class fraction %.2f outside [0.35, 0.45]", negShare))
	}
	if len(s.Scenarios) > 0 && preDegraded*100/len(s.Scenarios) < 10 {
		shortfalls = append(shortfalls, "pre-degraded fraction below 10%")
	}
	if cascadeCount < 3 {
		shortfalls = append(shortfalls, fmt.Sprintf("correlated_cascade scenarios: %d < 3", cascadeCount))
	}
	if pathologyCount < 10 {
		shortfalls = append(shortfalls, fmt.Sprintf("sensor_pathology scenarios: %d < 10", pathologyCount))
	}
	if s.Attempts >= cfg.MaxAttempts && len(s.Scenarios) < cfg.N {
		shortfalls = append(shortfalls, fmt.Sprintf("audit rejected too many scenarios: %d non-trivial of %d attempts", len(s.Scenarios), s.Attempts))
	}
	s.TerminalState = joinShortfalls(shortfalls)
	return s, nil
}

func joinShortfalls(shortfalls []string) string {
	out := ""
	for i, sf := range shortfalls {
		if i > 0 {
			out += "; "
		}
		out += sf
	}
	return out
}
