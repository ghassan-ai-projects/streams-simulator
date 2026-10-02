package cli

import (
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"strings"
)

func applyScriptedFaults(r *run.Run, options runOptions) error {
	start := options.startTime
	if options.faults != "" {
		for _, f := range splitCSV(options.faults) {
			entity, fault, offsetS, err := parseTriple(f, "=", "@")
			if err != nil {
				return fmt.Errorf("streamsim: %w", err)
			}
			if _, err := r.InjectFault(entity, fault, start+int64(offsetS*1e9), nil); err != nil {
				return fmt.Errorf("streamsim: %w", err)
			}
		}
	}
	return nil
}

func applyScriptedPerturbations(r *run.Run, options runOptions) error {
	start := options.startTime
	if options.perts != "" {
		for _, p := range splitCSV(options.perts) {
			parts := strings.SplitN(p, "@", 3)
			name := parts[0]
			fromS := 0.0
			untilS := 0.0
			if len(parts) > 1 {
				fromS = parseF(parts[1])
			}
			if len(parts) > 2 {
				untilS = parseF(parts[2])
			}
			from := start + int64(fromS*1e9)
			until := int64(0)
			if untilS > 0 {
				until = start + int64(untilS*1e9)
			}
			if _, err := r.ApplyPerturb(name, nil, from, until); err != nil {
				return fmt.Errorf("streamsim: %w", err)
			}
		}
	}
	return nil
}

func invokeScriptedEffectors(r *run.Run, options runOptions) error {
	start := options.startTime
	if options.effectors != "" {
		for _, e := range splitCSV(options.effectors) {
			eff, entity, offsetS, err := parseTriple(e, "@", "@")
			if err != nil {
				return fmt.Errorf("streamsim: %w", err)
			}
			cmdID := fmt.Sprintf("cli-%d", timeNanos())
			if _, err := r.InvokeEffector(eff, entity, cmdID, map[string]any{}, start+int64(offsetS*1e9)); err != nil {
				return fmt.Errorf("effector %s: %w", eff, err)
			}
		}
	}
	return nil
}
