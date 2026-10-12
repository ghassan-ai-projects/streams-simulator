package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/ghassan-ai-projects/streams-simulator/internal/cli/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

// cmdScenario exports the scenario that reproduces a recorded run
// (docs/design/SCENARIO_FILE.md §4).
func cmdScenario(s *session, args []string) (any, error) {
	options, err := parseScenarioExport(args, s.stderr)
	if err != nil {
		return nil, err
	}
	return exportScenario(options)
}

// exportOptions are scenario from-run's arguments.
type exportOptions struct {
	artifact, out, domainsDir, adaptersDir string
}

func parseScenarioExport(args []string, stderr io.Writer) (exportOptions, error) {
	var options exportOptions
	fs := newFlagSet("scenario from-run", stderr)
	fs.StringVar(&options.out, "out", "", "scenario file to write (required)")
	fs.StringVar(&options.domainsDir, "domains-dir", "domains", "domain specs directory (the export is checked against its domain)")
	fs.StringVar(&options.adaptersDir, "adapters-dir", "adapters", "adapters directory")
	verb, err := parseCommandVerb(fs, args, "")
	if err != nil {
		return exportOptions{}, err
	}
	options.artifact = fs.Arg(0)
	return options, admitScenarioExport(verb, options)
}

func admitScenarioExport(verb string, options exportOptions) error {
	if verb != "from-run" {
		return usagef("usage: streamsim scenario from-run <run.json> --out <file>")
	}
	if options.artifact == "" || options.out == "" {
		return usagef("scenario from-run requires a run artifact path and --out")
	}
	return nil
}

// exportScenario writes the scenario that reproduces a run (SCENARIO_FILE
// §4): the run's inputs must still be the ones it recorded; effector calls
// the world refused at admission are dropped; the result is checked against
// its contract and then as run --scenario would check it. A run that cannot
// be expressed as a runnable scenario exports nothing.
func exportScenario(options exportOptions) (any, error) {
	art, err := run.LoadArtifact(options.artifact)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	scenario, omitted := run.ScenarioFromArtifact(art)
	cfg, err := exportConfig(art, scenario, options)
	if err != nil {
		return nil, err
	}
	return writeExport(cfg, scenario, omitted, options)
}

// exportConfig is the configuration the export is checked with: the domain
// and adapter the artifact recorded, found in the given directories.
func exportConfig(art *model.RunArtifact, scenario *model.Scenario, options exportOptions) (run.Config, error) {
	spec, adap, err := loadSimulatorInputs(options.domainsDir, options.adaptersDir, scenario.Domain, scenario.Adapter)
	if err != nil {
		return run.Config{}, err
	}
	if err := run.CheckArtifactInputs(art, spec, adap); err != nil {
		return run.Config{}, fmt.Errorf("streamsim: scenario from-run: %w", err)
	}
	return run.ScenarioConfig(scenario, spec, adap, model.SinkInproc, ""), nil
}

func writeExport(cfg run.Config, scenario *model.Scenario, omitted int, options exportOptions) (any, error) {
	pruned, refused, raw, err := checkedExport(cfg, scenario)
	if err != nil {
		return nil, fmt.Errorf("streamsim: scenario from-run: the run cannot be expressed as a runnable scenario: %w", model.WithSource(err, options.artifact))
	}
	if err := files.Write(options.out, raw); err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return map[string]any{"scenario_written": options.out, "commands": len(pruned.Commands), "omitted": omitted, "refused": refused}, nil
}

// checkedExport is the scenario without refused calls, how many were
// dropped, and its file content, once both checks pass.
func checkedExport(cfg run.Config, scenario *model.Scenario) (*model.Scenario, int, []byte, error) {
	pruned, refused, err := run.PruneRefusedCalls(context.Background(), cfg, scenario)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("run: %w", err)
	}
	raw, err := encodeScenario(pruned)
	if err != nil {
		return nil, 0, nil, err
	}
	if err := run.CheckScenario(context.Background(), cfg, pruned); err != nil {
		return nil, 0, nil, fmt.Errorf("run: %w", err)
	}
	return pruned, refused, raw, nil
}

// encodeScenario is the scenario's file content, checked against its own
// contract.
func encodeScenario(scenario *model.Scenario) ([]byte, error) {
	raw, err := json.MarshalIndent(scenario, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode scenario: %w", err)
	}
	if err := model.ValidateScenario(raw); err != nil {
		return nil, fmt.Errorf("exported scenario: %w", err)
	}
	return append(raw, '\n'), nil
}
