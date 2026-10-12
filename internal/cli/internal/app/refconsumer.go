package app

import (
	"flag"
	"fmt"
	"io"

	"github.com/ghassan-ai-projects/streams-simulator/internal/refconsumer"
)

func cmdRefconsumer(s *session, args []string) (any, error) {
	options, err := parseRefconsumerOptions(args, s.stderr)
	if err != nil {
		return nil, err
	}
	summary, err := refconsumer.Consume(refconsumer.ConsumeOptions{Trace: options.trace, Out: options.out,
		Effector: options.effector, MCPEndpoint: options.mcpEndpoint, Token: options.token, RunID: options.runID,
		Threshold: options.threshold, Window: options.window})
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return map[string]any{"verdict_written": summary.Out, "detections": summary.Detections}, nil
}

type refconsumerOptions struct {
	trace, out, effector, mcpEndpoint, token, runID string
	threshold                                       float64
	window                                          int
}

func parseRefconsumerOptions(args []string, stderr io.Writer) (refconsumerOptions, error) {
	var options refconsumerOptions
	fs := newFlagSet("refconsumer", stderr)
	registerRefconsumerFlags(fs, &options)
	if err := parseOnlyFlags(fs, args); err != nil {
		return refconsumerOptions{}, err
	}
	return admitRefconsumerOptions(options)
}

func admitRefconsumerOptions(options refconsumerOptions) (refconsumerOptions, error) {
	if options.trace == "" {
		return refconsumerOptions{}, usagef("refconsumer requires --trace")
	}
	if options.window < 1 || options.threshold <= 0 {
		return refconsumerOptions{}, usagef("refconsumer: --window must be at least 1 and --threshold positive, got %d and %v", options.window, options.threshold)
	}
	if options.effector != "" && options.mcpEndpoint == "" {
		return refconsumerOptions{}, usagef("refconsumer: --effector requires --mcp (actuation needs the operator surface)")
	}
	return options, nil
}

func registerRefconsumerFlags(fs *flag.FlagSet, options *refconsumerOptions) {
	fs.StringVar(&options.trace, "trace", "", "native-format trace file (JSONL)")
	fs.StringVar(&options.out, "out", "verdict.json", "verdict output path")
	fs.Float64Var(&options.threshold, "threshold", 4, "z-score threshold")
	fs.IntVar(&options.window, "window", 30, "detector window")
	fs.StringVar(&options.effector, "effector", "", "effector to actuate on detection (from the nameplate)")
	fs.StringVar(&options.mcpEndpoint, "mcp", "", "operator endpoint URL (closes the loop over MCP)")
	fs.StringVar(&options.token, "token", "", "capability token from sim.world.create")
	fs.StringVar(&options.runID, "run", "", "run id to report against")
}
