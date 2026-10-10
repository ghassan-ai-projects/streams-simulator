package world

import (
	"fmt"
	"math"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/randutil"
)

func (w *World) validateArgs(eff *model.Effector, args map[string]any) error {
	if len(eff.ArgsSchema) == 0 {
		return nil
	}
	schema, err := w.effectorArgumentSchema(eff)
	if err != nil {
		return err
	}
	if errors := schema.Validate(args); len(errors) > 0 {
		return fmt.Errorf("world: invalid args for %q: %s", eff.Name, errors[0].Msg)
	}
	return nil
}

// pickFailureMode samples from the declared failure-mode distribution.
func (w *World) pickFailureMode(entityID string, eff *model.Effector) string {
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
func (w *World) SetFailureMode(mode string) {
	w.forceFailureMode = mode
}

// ackLatency samples the ack latency; slow mode is 10x (bounded).
func (w *World) ackLatency(entityID string, eff *model.Effector, mode string) float64 {
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

func sampleFailureMode(eff *model.Effector, rng *randutil.SplitMix64) string {
	total := failureModeWeight(eff)
	if total <= 0 {
		return ModeOK
	}
	draw := rng.Float64() * total
	for _, mode := range eff.Ack.FailureModes {
		if draw < mode.Probability {
			return mode.Mode
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
