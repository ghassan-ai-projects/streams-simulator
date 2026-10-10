package app

import (
	"fmt"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

func applyScriptedFaults(r *run.Run, options runOptions) error {
	start := options.startTime
	for _, f := range options.faults.items() {
		entity, fault, offsetS, err := parseTriple(f, "=", "@")
		if err != nil {
			return fmt.Errorf("streamsim: %w", err)
		}
		if _, err := r.InjectFault(entity, fault, start+int64(offsetS*1e9), nil); err != nil {
			return fmt.Errorf("streamsim: %w", err)
		}
	}
	return nil
}

func applyScriptedPerturbations(r *run.Run, options runOptions) error {
	for _, item := range options.perts.items() {
		name, from, until := scriptedPerturbation(item, options.startTime)
		if _, err := r.ApplyPerturb(name, nil, from, until); err != nil {
			return fmt.Errorf("streamsim: %w", err)
		}
	}
	return nil
}

// invokeScriptedEffectors actuates each --effector entry. Command ids are
// cli-<n> by position, so two identical invocations record identical command
// logs.
func invokeScriptedEffectors(r *run.Run, options runOptions) error {
	for n, item := range options.effectors.items() {
		if err := invokeScriptedEffector(r, item, options.startTime, fmt.Sprintf("cli-%d", n)); err != nil {
			return err
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

// effectorSpec is one parsed --effector entry.
type effectorSpec struct {
	effector, entity string
	offsetS          float64
	args             map[string]any
}

// parseEffectorSpec reads effector@entity@offset_s[@json-object]. The
// arguments object is optional: an effector with required arguments needs it.
func parseEffectorSpec(item string) (effectorSpec, error) {
	parts := strings.SplitN(item, "@", 4)
	if len(parts) < 2 {
		return effectorSpec{}, fmt.Errorf("bad triple %q", item)
	}
	spec := effectorSpec{effector: parts[0], entity: parts[1], args: map[string]any{}}
	if len(parts) > 2 {
		spec.offsetS = parseF(parts[2])
	}
	if len(parts) == 4 {
		return spec.withArguments(parts[3])
	}
	return spec, nil
}

func (spec effectorSpec) withArguments(text string) (effectorSpec, error) {
	if err := model.DecodeBytes([]byte(text), &spec.args); err != nil {
		return effectorSpec{}, fmt.Errorf("effector %s: arguments must be a JSON object: %w", spec.effector, err)
	}
	return spec, nil
}

func invokeScriptedEffector(r *run.Run, item string, start int64, command string) error {
	spec, err := parseEffectorSpec(item)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	at := start + int64(spec.offsetS*1e9)
	if _, err := r.InvokeEffector(spec.effector, spec.entity, command, spec.args, at); err != nil {
		return fmt.Errorf("effector %s: %w", spec.effector, err)
	}
	return nil
}
