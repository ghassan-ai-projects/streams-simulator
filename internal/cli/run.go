package cli

import (
	"context"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

func cmdRun(args []string) error {
	options, err := parseRunOptions(args)
	if err != nil {
		return err
	}
	cfg, err := loadRunConfig(options)
	if err != nil {
		return err
	}
	r, err := run.New(context.Background(), cfg)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	for _, apply := range []func(*run.Run, runOptions) error{applyScriptedFaults, applyScriptedPerturbations, invokeScriptedEffectors} {
		if err := apply(r, options); err != nil {
			return err
		}
	}
	if _, err := r.Advance(context.Background(), options.startTime+int64(options.durationS*1e9), false); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	art, err := r.End(options.outDir)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	return printJSON(map[string]any{
		"run_id": art.RunID, "trace_digest": art.ExpectedTraceDigest,
		"reproducible": art.Reproducible, "emitted": art.Counts.Emitted,
		"ledger": len(r.Ledger()),
	})
}
