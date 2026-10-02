package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/refconsumer"
	"os"
)

func cmdRefconsumer(args []string) error {
	fs := flag.NewFlagSet("refconsumer", flag.ExitOnError)
	trace := fs.String("trace", "", "native-format trace file (JSONL)")
	out := fs.String("out", "verdict.json", "verdict output path")
	threshold := fs.Float64("threshold", 4, "z-score threshold")
	window := fs.Int("window", 30, "detector window")
	effector := fs.String("effector", "", "effector to actuate on detection (from the nameplate)")
	mcpEndpoint := fs.String("mcp", "", "operator endpoint URL (closes the loop over MCP)")
	token := fs.String("token", "", "capability token from sim.world.create")
	runID := fs.String("run", "", "run id to report against")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	if *trace == "" {
		return fmt.Errorf("refconsumer requires --trace")
	}
	if *effector != "" && *mcpEndpoint == "" {
		return fmt.Errorf("refconsumer: --effector requires --mcp (actuation needs the operator surface)")
	}
	// #nosec G703 -- a CLI flag naming a trace file is user intent, not attacker input.
	raw, err := os.ReadFile(*trace)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	cfg := refconsumer.Config{
		Threshold: *threshold, Window: *window, MinConsecutive: 3,
		AbsenceFactor: 3, OnDetectionEffector: *effector,
	}
	var (
		np       *refconsumer.Nameplate
		invoker  refconsumer.EffectorInvoker
		sink     refconsumer.VerdictSink
		id       = "cli"
		operator *refconsumer.MCPOperator
	)
	if *mcpEndpoint != "" {
		operator, err = refconsumer.NewMCPOperator(*mcpEndpoint, *token, *runID)
		if err != nil {
			return fmt.Errorf("streamsim: %w", err)
		}
		defer func() { _ = operator.Close() }()
		np, err = operator.Nameplate()
		if err != nil {
			return fmt.Errorf("streamsim: %w", err)
		}
		invoker = operator
		sink = operator
		id = *runID
	} else {
		np = &refconsumer.Nameplate{WorldID: "cli"}
		sink = &fileSink{}
	}
	r := refconsumer.New(cfg, np, invoker, sink, id)
	v, err := r.Process(raw, 0)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	rawV, _ := json.MarshalIndent(v, "", "  ")
	// #nosec G703 -- the CLI output path is user intent.
	if err := os.WriteFile(*out, rawV, 0o600); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	return printJSON(map[string]any{"verdict_written": *out, "detections": len(v.Detections)})
}

type fileSink struct{}

func (f *fileSink) SubmitVerdict(v *model.Verdict) error { return nil }
