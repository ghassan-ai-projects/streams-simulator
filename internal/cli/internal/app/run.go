package app

import (
	"context"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

func cmdRun(s *session, args []string) (any, error) {
	options, err := parseRunOptions(args, s.stderr)
	if err != nil {
		return nil, err
	}
	cfg, err := loadRunConfig(options)
	if err != nil {
		return nil, err
	}
	return executeRun(cfg, options, func() int64 { return s.now().UnixNano() })
}

func executeRun(cfg run.Config, options runOptions, nanos func() int64) (any, error) {
	r, err := run.New(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	if err := applyRunScript(r, options, nanos); err != nil {
		return nil, err
	}
	if _, err := r.Advance(context.Background(), options.startTime+int64(options.durationS*1e9), false); err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return publishRun(r, options.outDir)
}

func applyRunScript(r *run.Run, options runOptions, nanos func() int64) error {
	if err := applyScriptedFaults(r, options); err != nil {
		return err
	}
	if err := applyScriptedPerturbations(r, options); err != nil {
		return err
	}
	return invokeScriptedEffectors(r, options, nanos)
}

func publishRun(r *run.Run, out string) (any, error) {
	art, err := r.End(out)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	return map[string]any{"run_id": art.RunID, "trace_digest": art.ExpectedTraceDigest,
		"reproducible": art.Reproducible, "emitted": art.Counts.Emitted, "ledger": len(r.Ledger())}, nil
}
