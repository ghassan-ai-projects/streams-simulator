// Package suite generates graded scenario suites: declared negative-class
// fraction, randomized onset (including pre-degraded starts), perturbation
// coverage, and a generate-audit-regenerate loop that keeps only
// non-trivial scenarios. A domain that cannot produce non-trivial scenarios
// at the declared prevalence reports a terminal state rather than an empty
// directory (G-06).
package suite

import (
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
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
	generation, err := newSuiteGeneration(cfg)
	if err != nil {
		return nil, err
	}
	if err := generation.populate(); err != nil {
		return nil, err
	}
	generation.reportShortfalls()
	return generation.suite, nil
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
