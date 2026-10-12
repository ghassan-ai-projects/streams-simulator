package domain

import (
	"errors"
	"fmt"
	"math"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/randutil"
)

func (w *World) validateArgs(eff *model.Effector, args map[string]any) error {
	if err := w.argsViolation(eff, args); err != nil {
		return fmt.Errorf("world: invalid args for %q: %w", eff.Name, err)
	}
	return nil
}

// CheckEffectorArgs reports whether effector is declared by the domain and
// args satisfy its args_schema, without invoking it. The error states the
// violation.
func (w *World) CheckEffectorArgs(effector string, args map[string]any) error {
	eff := w.Spec.Effector(effector)
	if eff == nil {
		return fmt.Errorf("the domain declares no effector %q", effector)
	}
	return w.argsViolation(eff, args)
}

// argsViolation is the first way args break the effector's args_schema.
func (w *World) argsViolation(eff *model.Effector, args map[string]any) error {
	if len(eff.ArgsSchema) == 0 {
		return nil
	}
	schema, err := w.effectorArgumentSchema(eff)
	if err != nil {
		return err
	}
	if violations := schema.Validate(args); len(violations) > 0 {
		return errors.New(violations[0].Msg)
	}
	return nil
}

// pickFailureMode samples from the declared failure-mode distribution.
func (w *World) pickFailureMode(entityID string, eff *model.Effector) FailureMode {
	if w.forceEffectorOK {
		return ModeOK
	}
	if w.forceFailureMode != "" {
		return w.forceFailureMode
	}
	if len(eff.Ack.FailureModes) == 0 {
		return ModeOK
	}
	rng := w.substream(entityID + "/" + eff.Name + "/fault_shape")
	return sampleFailureMode(eff, rng)
}

// SetFailureMode overrides the failure-mode selection for subsequent
// invocations ("" restores the declared distribution). Test-only knob; the
// distribution is fixed per run otherwise.
func (w *World) SetFailureMode(mode FailureMode) {
	w.forceFailureMode = mode
}

// ackLatency samples the ack latency; slow mode is 10x (bounded).
func (w *World) ackLatency(entityID string, eff *model.Effector, mode FailureMode) float64 {
	rng := w.substream(entityID + "/" + eff.Name + "/delay")
	mean := acknowledgementMean(eff)
	latency := mean + eff.Ack.LatencyMS.Sigma*rng.Norm()
	if latency < 1 {
		latency = 1
	}
	if mode == ModeSlow {
		latency *= 10
	}
	return math.Round(latency)
}

// interlockHolds evaluates the interlock predicate over hidden state.
func (w *World) interlockHolds(entityID string, il *model.Interlock, atNS int64) bool {
	v := w.stateAt(entityID, il.State, atNS)
	switch il.Operator {
	case "lt":
		return v < il.Threshold
	case "lte":
		return v <= il.Threshold
	case "gt":
		return v > il.Threshold
	case "gte":
		return v >= il.Threshold
	}
	return false
}

func (w *World) effectorArgumentSchema(eff *model.Effector) (*jsonschema.Schema, error) {
	schema := w.argSchemas[eff.Name]
	if schema != nil {
		return schema, nil
	}
	schema, err := jsonschema.Compile(eff.ArgsSchema)
	if err != nil {
		return nil, fmt.Errorf("world: effector %q args schema: %w", eff.Name, err)
	}
	w.argSchemas[eff.Name] = schema
	return schema, nil
}

func sampleFailureMode(eff *model.Effector, rng *randutil.SplitMix64) FailureMode {
	total := failureModeWeight(eff)
	if total <= 0 {
		return ModeOK
	}
	draw := rng.Float64() * total
	for _, mode := range eff.Ack.FailureModes {
		if draw < mode.Probability {
			return FailureMode(mode.Mode)
		}
		draw -= mode.Probability
	}
	return ModeOK
}

func acknowledgementMean(eff *model.Effector) float64 {
	mean := eff.Ack.LatencyMS.Mean
	if mean <= 0 {
		mean = 100
	}
	return mean
}

func failureModeWeight(eff *model.Effector) float64 {
	total := 0.0
	for _, mode := range eff.Ack.FailureModes {
		total += mode.Probability
	}
	return total
}
