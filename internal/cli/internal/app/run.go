package app

import (
	"context"
	"errors"
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
	summary, err := run.RunScenario(context.Background(), source.scenario, run.ScenarioRunOptions{DomainsDir: options.domainsDir,
		AdaptersDir: options.adaptersDir, SinkName: options.sinkName, SinkTarget: options.sinkTarget, OutDir: options.outDir, Explore: options.explore})
	var check *run.ScenarioCheckError
	if errors.As(err, &check) {
		return nil, scenarioCheckFailure(options, source, check.Err)
	}
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return map[string]any{"run_id": summary.RunID, "trace_digest": summary.TraceDigest,
		"reproducible": summary.Reproducible, "emitted": summary.Emitted, "ledger": summary.Ledger}, nil
}

// scenarioCheckFailure is a scenario file's violation, or, for a scenario the
// flags describe, a usage failure whose issues name the flags.
func scenarioCheckFailure(options runOptions, source scenarioSource, err error) error {
	if options.scenarioPath != "" {
		return fmt.Errorf("streamsim: %w", model.WithSource(err, options.scenarioPath))
	}
	return flagScenarioFailure(err, source.origins)
}
