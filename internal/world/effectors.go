package world

// Effector invocation: the closed loop. A consumer actuates the world
// through a declared effector, which applies physical effects with time
// constants and dead times, may fail in declared ways (including both
// no-effect variants, one of which runs a shadow state), may be refused by
// an independent interlock, and is idempotent by command_id.

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// Effector failure modes.
const (
	ModeOK                = "ok"
	ModeSlow              = "slow"
	ModeAckLost           = "ack_lost"
	ModeReject            = "reject"
	ModePartial           = "partial"
	ModeConfirmedNoEffect = "confirmed_no_effect"
	ModeSilentNoEffect    = "silent_no_effect"
)

// InvokeResult is what an effector invocation returns.
type InvokeResult struct {
	Accepted      bool    `json:"accepted"`
	Simulated     bool    `json:"simulated"`
	WorldID       string  `json:"world_id"`
	CommandID     string  `json:"command_id"`
	EffectETANS   int64   `json:"effect_eta_ns,omitempty"`
	Reason        string  `json:"reason,omitempty"`
	Mode          string  `json:"mode,omitempty"`
	AckLatencyMS  float64 `json:"ack_latency_ms,omitempty"`
	EffectApplied bool    `json:"effect_applied,omitempty"`
}

// ErrInterlockRefused is the terminal refusal of an independent safety
// system. Refusal is not a retryable error.
var ErrInterlockRefused = fmt.Errorf("world: interlock refused")

// ErrEffectorRefused is a generic refusal whose detail must never reach the
// operator role.
var ErrEffectorRefused = fmt.Errorf("world: effector refused")

// InvokeEffector actuates an effector. commandID is the idempotency key;
// repeating a call with the same command_id inside the declared window
// returns the original result and applies no second effect.
func (w *World) InvokeEffector(effector, entityID, commandID string, args map[string]any, atNS int64) (*InvokeResult, error) {
	if commandID == "" {
		return nil, fmt.Errorf("world: missing command_id")
	}
	ent := w.entities[entityID]
	if ent == nil || !ent.alive {
		return nil, fmt.Errorf("world: unknown entity %q", entityID)
	}
	eff := w.Spec.Effector(effector)
	if eff == nil {
		return nil, fmt.Errorf("world: unknown effector %q", effector)
	}
	if err := w.validateArgs(eff, args); err != nil {
		return nil, fmt.Errorf("InvokeEffector: %w", err)
	}
	// Idempotency: a repeated command_id inside the window replays the
	// original result.
	if prev, ok := w.idempotent[commandID]; ok {
		if atNS < prev.expiresNS {
			return prev.callResult(), nil
		}
		delete(w.idempotent, commandID)
	}

	// Interlock: an independent safety system may refuse. Refusal is
	// terminal; the runtime must never retry or route around it.
	if eff.Interlock != nil {
		if w.interlockHolds(entityID, eff.Interlock, atNS) {
			if act := eff.Interlock.AutonomousAction; act != nil {
				w.applyAutonomousAction(entityID, act.State, act.Delta, atNS)
			}
			w.recordCall(effector, entityID, commandID, args, atNS, ModeReject, false, true, "interlock_refused", 0, false)
			return nil, ErrInterlockRefused
		}
	}

	// Failure mode and ack latency from the effector's own substream.
	mode := w.pickFailureMode(entityID, eff, atNS)
	latency := w.ackLatency(entityID, eff, mode, atNS)
	effResult := &InvokeResult{
		Simulated:    true,
		WorldID:      w.ID,
		CommandID:    commandID,
		AckLatencyMS: latency,
		Mode:         mode,
	}

	effectApplied := false
	switch mode {
	case ModeOK, ModeSlow, ModeAckLost:
		w.applyEffect(entityID, eff, args, atNS, 1.0, false)
		effectApplied = true
		effResult.Accepted = mode != ModeAckLost
		if !effResult.Accepted {
			effResult.Reason = "ack_lost"
		}
		effResult.EffectETANS = atNS + int64(eff.Effect.DeadTimeS*secondsPerNS)
	case ModeReject:
		effResult.Accepted = false
		effResult.Reason = "effector_refused"
	case ModePartial:
		w.applyEffect(entityID, eff, args, atNS, 0.5, false)
		effectApplied = true
		effResult.Accepted = true
		effResult.EffectETANS = atNS + int64(eff.Effect.DeadTimeS*secondsPerNS)
	case ModeConfirmedNoEffect:
		// The ack lies; no effect anywhere. Confirmation channels keep
		// reporting the truth.
		effResult.Accepted = true
	case ModeSilentNoEffect:
		// The ack lies; the effect applies to a shadow state that the
		// confirmation channels report, so the emitted evidence is
		// internally consistent with success. Only the absent physical
		// outcome betrays the failure.
		w.applyEffect(entityID, eff, args, atNS, 1.0, true)
		effectApplied = true
		effResult.Accepted = true
		effResult.EffectETANS = atNS + int64(eff.Effect.DeadTimeS*secondsPerNS)
	}
	effResult.EffectApplied = effectApplied

	w.recordCall(effector, entityID, commandID, args, atNS, mode, effResult.Accepted, false, effResult.Reason, latency, effectApplied)
	window := eff.IdempotencyWindowS
	if window <= 0 {
		window = 3600
	}
	w.idempotent[commandID] = &idempotentResult{call: w.effectorCalls[len(w.effectorCalls)-1], expiresNS: atNS + int64(window*secondsPerNS)}
	return effResult, nil
}

// callResult reconstructs the original invoke result for an idempotent
// replay.
func (r *idempotentResult) callResult() *InvokeResult {
	c := r.call
	return &InvokeResult{
		Accepted:      c.Accepted,
		Simulated:     true,
		WorldID:       c.WorldID,
		CommandID:     c.CommandID,
		Reason:        c.Reason,
		Mode:          c.Mode,
		AckLatencyMS:  c.AckLatencyMS,
		EffectApplied: c.EffectApplied,
	}
}

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

// applyAutonomousAction applies the interlock's autonomous state delta.
func (w *World) applyAutonomousAction(entityID, state string, delta float64, atNS int64) {
	k := &kick{entity: entityID, state: state, delta: delta, startNS: atNS, tauNS: 0}
	key := driverKey{entity: entityID, state: state}
	w.kicks[key] = append(w.kicks[key], k)
	w.schedule(kindEffectStart, entityID, state, atNS, k)
}

// applyEffect schedules the effect's state deltas as kicks, in the real
// world or in the shadow world (silent_no_effect).
func (w *World) applyEffect(entityID string, eff *model.Effector, args map[string]any, atNS int64, scale float64, shadow bool) {
	start := atNS + int64(eff.Effect.DeadTimeS*secondsPerNS)
	tau := eff.Effect.TimeConstantS * secondsPerNS
	for _, d := range eff.Effect.StateDeltas {
		delta := d.Delta
		if d.FromArg != "" {
			if v, ok := argFloat(args, d.FromArg); ok {
				delta = v
			}
		}
		delta *= scale
		k := &kick{entity: entityID, state: d.State, delta: delta, startNS: start, tauNS: tau}
		key := driverKey{entity: entityID, state: d.State}
		if shadow {
			w.shadow[key] = append(w.shadow[key], k)
		} else {
			w.kicks[key] = append(w.kicks[key], k)
		}
		w.schedule(kindEffectStart, entityID, d.State, start, k)
	}
}

func argFloat(args map[string]any, name string) (float64, bool) {
	if args == nil {
		return 0, false
	}
	v, ok := args[name]
	if !ok {
		return 0, false
	}
	switch x := v.(type) {
	case float64:
		return x, true
	case json.Number:
		f, err := x.Float64()
		return f, err == nil
	case int64:
		return float64(x), true
	case int:
		return float64(x), true
	}
	return 0, false
}

func (w *World) recordCall(effector, entityID, commandID string, args map[string]any, atNS int64, mode string, accepted, interlock bool, reason string, latency float64, effectApplied bool) {
	call := &EffectorCall{
		CommandID:        commandID,
		Effector:         effector,
		EntityID:         entityID,
		WorldID:          w.ID,
		Args:             args,
		AtNS:             atNS,
		Mode:             mode,
		Accepted:         accepted,
		InterlockRefused: interlock,
		Reason:           reason,
		AckLatencyMS:     latency,
		EffectApplied:    effectApplied,
	}
	call.ResultDigest = resultDigest(call)
	w.effectorCalls = append(w.effectorCalls, *call)
}

func resultDigest(c *EffectorCall) string {
	// Best-effort short digest of the outcome; digests for the run artifact
	// are computed by the run layer from canonical JSON.
	b, _ := json.Marshal(map[string]any{
		"mode": c.Mode, "accepted": c.Accepted, "reason": c.Reason,
		"effect_applied": c.EffectApplied,
	})
	return string(b)
}

// EffectorCalls returns the effector call log (the authority when scoring
// actions).
func (w *World) EffectorCalls() []EffectorCall {
	out := make([]EffectorCall, len(w.effectorCalls))
	copy(out, w.effectorCalls)
	return out
}

// HiddenStateSnapshot returns the true hidden state values of every live
// entity at time t (director-only; feeds world_state_history).
func (w *World) HiddenStateSnapshot(t int64) map[string]map[string]float64 {
	out := map[string]map[string]float64{}
	for _, id := range w.entityOrder {
		ent := w.entities[id]
		if ent == nil || !ent.alive {
			continue
		}
		row := map[string]float64{}
		for _, name := range w.sortedStateNames() {
			row[name] = w.stateAt(id, name, t)
		}
		out[id] = row
	}
	return out
}

// sortedStateNames returns the declared state names in a stable order.
func (w *World) sortedStateNames() []string {
	names := w.Spec.StateNames()
	sort.Strings(names)
	return names
}

// PendingKicks is the number of scheduled-but-unapplied effect kicks.
func (w *World) PendingKicks() int {
	n := 0
	for _, key := range w.sortedKickStates() {
		for _, k := range w.kicks[key] {
			if !k.applied {
				n++
			}
		}
	}
	return n
}

// sortedKickStates returns the states with pending kicks in a stable order.
func (w *World) sortedKickStates() []driverKey {
	var out []driverKey
	// determinism-safe: collected here, sorted below before any output.
	for key := range w.kicks {
		out = append(out, key)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].entity != out[j].entity {
			return out[i].entity < out[j].entity
		}
		return out[i].state < out[j].state
	})
	return out
}
