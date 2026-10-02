package world

import (
	"fmt"
	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"math"
)

func (w *World) validateArgs(eff *model.Effector, args map[string]any) error {
	if len(eff.ArgsSchema) == 0 {
		return nil
	}
	sch := w.argSchemas[eff.Name]
	if sch == nil {
		var err error
		sch, err = jsonschema.Compile(eff.ArgsSchema)
		if err != nil {
			return fmt.Errorf("world: effector %q args schema: %w", eff.Name, err)
		}
		w.argSchemas[eff.Name] = sch
	}
	if errs := sch.Validate(args); len(errs) > 0 {
		return fmt.Errorf("world: invalid args for %q: %s", eff.Name, errs[0].Msg)
	}
	return nil
}

// pickFailureMode samples from the declared failure-mode distribution.
func (w *World) pickFailureMode(entityID string, eff *model.Effector, atNS int64) string {
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
	total := 0.0
	for _, fm := range eff.Ack.FailureModes {
		total += fm.Probability
	}
	if total <= 0 {
		return ModeOK
	}
	r := rng.Float64() * total
	for _, fm := range eff.Ack.FailureModes {
		if r < fm.Probability {
			return fm.Mode
		}
		r -= fm.Probability
	}
	return ModeOK
}

// SetFailureMode overrides the failure-mode selection for subsequent
// invocations ("" restores the declared distribution). Test-only knob; the
// distribution is fixed per run otherwise.
func (w *World) SetFailureMode(mode string) {
	w.forceFailureMode = mode
}

// ackLatency samples the ack latency; slow mode is 10x (bounded).
func (w *World) ackLatency(entityID string, eff *model.Effector, mode string, atNS int64) float64 {
	rng := w.substream(entityID + "/" + eff.Name + "/delay")
	mean := eff.Ack.LatencyMS.Mean
	if mean <= 0 {
		mean = 100
	}
	sigma := eff.Ack.LatencyMS.Sigma
	l := mean + sigma*rng.Norm()
	if l < 1 {
		l = 1
	}
	if mode == ModeSlow {
		l *= 10
	}
	return math.Round(l)
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
