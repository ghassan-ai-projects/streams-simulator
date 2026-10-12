package app

import (
	"context"
	"fmt"
	"io"

	"github.com/ghassan-ai-projects/streams-simulator/internal/cli/internal/files"
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
	exported, err := run.ExportScenarioFile(context.Background(), options.artifact,
		run.Directories{Domains: options.domainsDir, Adapters: options.adaptersDir})
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	if err := files.Write(options.out, exported.Raw); err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return map[string]any{"scenario_written": options.out, "commands": len(exported.Scenario.Commands),
		"omitted": exported.Omitted, "refused": exported.Refused}, nil
}
