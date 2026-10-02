package world

import (
	"fmt"
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
