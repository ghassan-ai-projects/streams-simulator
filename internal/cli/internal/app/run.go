package app

import (
	"context"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/cli/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

// cmdRun executes one scenario: the file named by --scenario, or the one the
// flags describe. Both take the same path (docs/design/SCENARIO_FILE.md §3).
func cmdRun(s *session, args []string) (any, error) {
	options, err := parseRunOptions(args, s.stderr)
	if err != nil {
		return nil, err
	}
	source, err := runScenario(options)
	if err != nil {
		return nil, err
	}
	return executeScenarioRun(source, options)
}

// scenarioSource is the scenario a run executes and, for one the flags
// describe, the flag behind each command.
type scenarioSource struct {
	scenario *model.Scenario
	origins  []string
}

func runScenario(options runOptions) (scenarioSource, error) {
	if options.scenarioPath != "" {
		scenario, err := loadScenarioFile(options.scenarioPath)
		return scenarioSource{scenario: scenario}, err
	}
	scenario, origins, err := scenarioFromFlags(options)
	return scenarioSource{scenario: scenario, origins: origins}, err
}

// loadScenarioFile reads a scenario and checks it against scenario-v0.1.
func loadScenarioFile(path string) (*model.Scenario, error) {
	raw, err := files.Read(path)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	scenario, err := model.DecodeScenario(raw)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", model.WithSource(err, path))
	}
	return scenario, nil
}

func executeScenarioRun(source scenarioSource, options runOptions) (any, error) {
	cfg, err := checkedScenarioConfig(source, options)
	if err != nil {
		return nil, err
	}
	r, err := run.New(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	if err := r.ExecuteScenario(context.Background(), source.scenario); err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return publishRun(r, options.outDir)
}

// checkedScenarioConfig resolves the scenario's domain and adapter and checks
// the scenario against them before any sink opens.
func checkedScenarioConfig(source scenarioSource, options runOptions) (run.Config, error) {
	scenario := source.scenario
	spec, adap, err := loadSimulatorInputs(options.domainsDir, options.adaptersDir, scenario.Domain, scenario.Adapter)
	if err != nil {
		return run.Config{}, err
	}
	cfg := run.ScenarioConfig(scenario, spec, adap, options.sinkName, options.sinkTarget)
	if err := run.CheckScenario(context.Background(), cfg, scenario); err != nil {
		return run.Config{}, scenarioCheckFailure(options, source, err)
	}
	return cfg, nil
}

// scenarioCheckFailure is a scenario file's violation, or, for a scenario the
// flags describe, a usage failure whose issues name the flags.
func scenarioCheckFailure(options runOptions, source scenarioSource, err error) error {
	if options.scenarioPath != "" {
		return fmt.Errorf("streamsim: %w", model.WithSource(err, options.scenarioPath))
	}
	return flagScenarioFailure(err, source.origins)
}

func publishRun(r *run.Run, out string) (any, error) {
	art, err := r.End(out)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return map[string]any{"run_id": art.RunID, "trace_digest": art.ExpectedTraceDigest,
		"reproducible": art.Reproducible, "emitted": art.Counts.Emitted, "ledger": len(r.Ledger())}, nil
}
