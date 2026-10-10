package app

import (
	"fmt"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
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
	if options.perts != "" {
		for _, item := range splitCSV(options.perts) {
			name, from, until := scriptedPerturbation(item, options.startTime)
			if _, err := r.ApplyPerturb(name, nil, from, until); err != nil {
				return fmt.Errorf("streamsim: %w", err)
			}
		}
	}
	return nil
}

func invokeScriptedEffectors(r *run.Run, options runOptions, nanos func() int64) error {
	if options.effectors != "" {
		for _, item := range splitCSV(options.effectors) {
			if err := invokeScriptedEffector(r, item, options.startTime, nanos); err != nil {
				return err
			}
		}
	}
	return nil
}

func scriptedPerturbation(item string, start int64) (string, int64, int64) {
	parts := strings.SplitN(item, "@", 3)
	fromS, untilS := 0.0, 0.0
	if len(parts) > 1 {
		fromS = parseF(parts[1])
	}
	if len(parts) > 2 {
		untilS = parseF(parts[2])
	}
	until := int64(0)
	if untilS > 0 {
		until = start + int64(untilS*1e9)
	}
	return parts[0], start + int64(fromS*1e9), until
}

func invokeScriptedEffector(r *run.Run, item string, start int64, nanos func() int64) error {
	effector, entity, offset, err := parseTriple(item, "@", "@")
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	command := fmt.Sprintf("cli-%d", nanos())
	if _, err := r.InvokeEffector(effector, entity, command, map[string]any{}, start+int64(offset*1e9)); err != nil {
		return fmt.Errorf("effector %s: %w", effector, err)
	}
	return nil
}
