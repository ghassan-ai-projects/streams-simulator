package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/refconsumer"
)

func cmdRefconsumer(args []string) error {
	options, err := parseRefconsumerOptions(args)
	if err != nil {
		return err
	}
	// #nosec G703 -- a CLI flag naming a trace file is user intent.
	raw, err := os.ReadFile(options.trace)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	return consumeReferenceTrace(raw, options)
}

func connectReferenceConsumer(endpoint, token, runID string) (*refconsumer.Nameplate, *refconsumer.MCPOperator, error) {
	operator, err := refconsumer.NewMCPOperator(endpoint, token, runID)
	if err != nil {
		return nil, nil, fmt.Errorf("%w", err)
	}
	nameplate, err := operator.Nameplate()
	if err != nil {
		_ = operator.Close()
		return nil, nil, fmt.Errorf("%w", err)
	}
	return nameplate, operator, nil
}

func writeConsumerVerdict(v *model.Verdict, out string) error {
	rawV, _ := json.MarshalIndent(v, "", "  ")
	// #nosec G703 -- the CLI output path is user intent.
	if err := os.WriteFile(out, rawV, 0o600); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	return nil
}

type fileSink struct{}

func (f *fileSink) SubmitVerdict(v *model.Verdict) error { return nil }

type refconsumerOptions struct {
	trace, out, effector, mcpEndpoint, token, runID string
	threshold                                       float64
	window                                          int
}

func parseRefconsumerOptions(args []string) (refconsumerOptions, error) {
	var options refconsumerOptions
	fs := flag.NewFlagSet("refconsumer", flag.ExitOnError)
	registerRefconsumerFlags(fs, &options)
	if err := fs.Parse(args); err != nil {
		return refconsumerOptions{}, fmt.Errorf("streamsim: %w", err)
	}
	if options.trace == "" {
		return refconsumerOptions{}, fmt.Errorf("refconsumer requires --trace")
	}
	if options.effector != "" && options.mcpEndpoint == "" {
		return refconsumerOptions{}, fmt.Errorf("refconsumer: --effector requires --mcp (actuation needs the operator surface)")
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

func consumeReferenceTrace(raw []byte, options refconsumerOptions) error {
	if options.mcpEndpoint == "" {
		return runReferenceConsumer(raw, options, &refconsumer.Nameplate{WorldID: "cli"}, nil, &fileSink{}, "cli")
	}
	np, operator, err := connectReferenceConsumer(options.mcpEndpoint, options.token, options.runID)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	defer func() { _ = operator.Close() }()
	return runReferenceConsumer(raw, options, np, operator, operator, options.runID)
}

func runReferenceConsumer(raw []byte, options refconsumerOptions, np *refconsumer.Nameplate, invoker refconsumer.EffectorInvoker, sink refconsumer.VerdictSink, id string) error {
	cfg := refconsumer.Config{Threshold: options.threshold, Window: options.window, MinConsecutive: 3,
		AbsenceFactor: 3, OnDetectionEffector: options.effector}
	r := refconsumer.New(cfg, np, invoker, sink, id)
	v, err := r.Process(raw, 0)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	if err := writeConsumerVerdict(v, options.out); err != nil {
		return err
	}
	return printJSON(map[string]any{"verdict_written": options.out, "detections": len(v.Detections)})
}
