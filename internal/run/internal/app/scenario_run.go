package app

import (
	"context"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type ScenarioRunOptions struct {
	DomainsDir, AdaptersDir, SinkName, SinkTarget, OutDir string
	Explore                                               bool
}

type RunSummary struct {
	RunID, TraceDigest string
	Reproducible       bool
	Emitted            int64
	Ledger             int
}

type ScenarioCheckError struct {
	Err error
}

func (e *ScenarioCheckError) Error() string { return e.Err.Error() }

func (e *ScenarioCheckError) Unwrap() error { return e.Err }

func RunScenario(ctx context.Context, scenario *model.Scenario, options ScenarioRunOptions) (RunSummary, error) {
	spec, adap, err := LoadInputs(options.DomainsDir, options.AdaptersDir, scenario.Domain, scenario.Adapter)
	if err != nil {
		return RunSummary{}, fmt.Errorf("%w", err)
	}
	cfg := ScenarioConfig(scenario, spec, adap, options.SinkName, options.SinkTarget)
	if err := CheckScenario(ctx, cfg, scenario); err != nil {
		return RunSummary{}, &ScenarioCheckError{Err: err}
	}
	cfg.JournalDir, cfg.Explore = options.OutDir, options.Explore
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
