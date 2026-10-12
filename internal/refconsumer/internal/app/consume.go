// Package app is the use-case layer of the refconsumer module: it runs the
// reference consumer over a trace file, optionally closing the loop over the
// operator endpoint of a director process, and writes the verdict it reports.
package app

import (
	"encoding/json"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/refconsumer/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/refconsumer/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/refconsumer/internal/mcpclient"
)

// Options say which trace to consume, where the verdict goes, how the detector
// is tuned, and, when MCPEndpoint is set, the operator endpoint, capability
// token and run the consumer actuates and reports against.
type Options struct {
	Trace, Out, Effector, MCPEndpoint, Token, RunID string
	Threshold                                       float64
	Window                                          int
}

// Summary is what a consumed trace left behind: the verdict file and how many
// detections it reports.
type Summary struct {
	Out        string
	Detections int
}

// Consume reads the trace, detects episodes in it and writes the verdict. With
// an operator endpoint the consumer connects to it, actuates the effector on
// detection and reports through it.
func Consume(options Options) (Summary, error) {
	raw, err := files.ReadTrace(options.Trace)
	if err != nil {
		return Summary{}, fmt.Errorf("streamsim: %w", err)
	}
	if options.MCPEndpoint == "" {
		return detect(raw, options, &domain.Nameplate{WorldID: "cli"}, nil, &fileSink{}, "cli")
	}
	np, operator, err := connect(options.MCPEndpoint, options.Token, options.RunID)
	if err != nil {
		return Summary{}, fmt.Errorf("streamsim: %w", err)
	}
	defer func() { _ = operator.Close() }()
	return detect(raw, options, np, operator, operator, options.RunID)
}

func connect(endpoint, token, runID string) (*domain.Nameplate, *mcpclient.MCPOperator, error) {
	operator, err := mcpclient.NewMCPOperator(endpoint, token, runID)
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

func detect(raw []byte, options Options, np *domain.Nameplate, invoker domain.EffectorInvoker, sink domain.VerdictSink, id string) (Summary, error) {
	cfg := domain.Config{Threshold: options.Threshold, Window: options.Window, MinConsecutive: 3,
		AbsenceFactor: 3, OnDetectionEffector: options.Effector}
	r := domain.New(cfg, np, invoker, sink, id)
	v, err := r.Process(raw, 0)
	if err != nil {
		return Summary{}, fmt.Errorf("streamsim: %w", err)
	}
	if err := writeVerdict(v, options.Out); err != nil {
		return Summary{}, err
	}
	return Summary{Out: options.Out, Detections: len(v.Detections)}, nil
}

func writeVerdict(v *model.Verdict, out string) error {
	rawV, _ := json.MarshalIndent(v, "", "  ")
	if err := files.WriteVerdict(out, rawV); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	return nil
}

// fileSink accepts the verdict without reporting it anywhere: with no operator
// endpoint the verdict file is the whole report.
type fileSink struct{}

func (f *fileSink) SubmitVerdict(v *model.Verdict) error { return nil }
