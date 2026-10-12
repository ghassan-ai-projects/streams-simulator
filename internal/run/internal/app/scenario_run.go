package app

import (
	"context"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// ScenarioRunOptions say where a scenario's domain and adapter are looked up,
// which sink delivers the run and where its evidence is published (empty
// publishes nothing).
type ScenarioRunOptions struct {
	DomainsDir, AdaptersDir, SinkName, SinkTarget, OutDir string
}

// RunSummary is what a finished scenario run left behind.
type RunSummary struct {
	RunID, TraceDigest string
	Reproducible       bool
	Emitted            int64
	Ledger             int
}

// ScenarioCheckError is the violation the static check or dry run found in a
// scenario, before any sink opened. Err is the violation itself.
type ScenarioCheckError struct {
	Err error
}

// Error is the violation's message.
func (e *ScenarioCheckError) Error() string { return e.Err.Error() }

// Unwrap returns the violation.
func (e *ScenarioCheckError) Unwrap() error { return e.Err }

// RunScenario resolves the scenario's domain and adapter, checks the scenario
// against them before any sink opens (a violation is a *ScenarioCheckError),
// then executes its commands in order and ends the run.
func RunScenario(ctx context.Context, scenario *model.Scenario, options ScenarioRunOptions) (RunSummary, error) {
	spec, adap, err := LoadInputs(options.DomainsDir, options.AdaptersDir, scenario.Domain, scenario.Adapter)
	if err != nil {
		return RunSummary{}, fmt.Errorf("%w", err)
	}
	cfg := ScenarioConfig(scenario, spec, adap, options.SinkName, options.SinkTarget)
	if err := CheckScenario(ctx, cfg, scenario); err != nil {
		return RunSummary{}, &ScenarioCheckError{Err: err}
	}
	return execute(ctx, cfg, scenario, options.OutDir)
}

func execute(ctx context.Context, cfg Config, scenario *model.Scenario, outDir string) (RunSummary, error) {
	r, err := New(ctx, cfg)
	if err != nil {
		return RunSummary{}, fmt.Errorf("streamsim: %w", err)
	}
	if err := r.ExecuteScenario(ctx, scenario); err != nil {
		return RunSummary{}, fmt.Errorf("streamsim: %w", err)
	}
	art, err := r.End(outDir)
	if err != nil {
		return RunSummary{}, fmt.Errorf("streamsim: %w", err)
	}
	return RunSummary{RunID: art.RunID, TraceDigest: art.ExpectedTraceDigest, Reproducible: art.Reproducible,
		Emitted: art.Counts.Emitted, Ledger: len(r.Ledger())}, nil
}
