package domain

import (
	"fmt"

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

// invocation is one effector actuation request: what was asked, of which
// entity, under which idempotency key, at what world time.
type invocation struct {
	effector  string
	entityID  string
	commandID string
	args      map[string]any
	atNS      int64
}

// callOutcome is what became of an invocation, as recorded in the call log.
type callOutcome struct {
	mode          string
	accepted      bool
	interlock     bool
	reason        string
	latencyMS     float64
	effectApplied bool
}

// InvokeEffector actuates an effector. commandID is the idempotency key;
// repeating a call with the same command_id inside the declared window
// returns the original result and applies no second effect.
func (w *World) InvokeEffector(effector, entityID, commandID string, args map[string]any, atNS int64) (*InvokeResult, error) {
	inv := invocation{effector: effector, entityID: entityID, commandID: commandID, args: args, atNS: atNS}
	eff, err := w.admitInvocation(inv)
	if err != nil {
		return nil, err
	}
	if result, known := w.replayInvocation(inv.commandID, inv.atNS); known {
		return result, nil
	}
	if w.refuseInterlock(eff, inv) {
		return nil, ErrInterlockRefused
	}
	return w.executeInvocation(eff, inv), nil
}

func (w *World) admitInvocation(inv invocation) (*model.Effector, error) {
	eff, err := w.invocationIdentity(inv.effector, inv.entityID, inv.commandID)
	if err != nil {
		return nil, err
	}
	if err := w.validateArgs(eff, inv.args); err != nil {
		return nil, fmt.Errorf("InvokeEffector: %w", err)
	}
	return eff, nil
}

func (w *World) invocationIdentity(effector, entityID, commandID string) (*model.Effector, error) {
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
	return eff, nil
}

// A command inside its idempotency window replays without another effect.
func (w *World) replayInvocation(commandID string, atNS int64) (*InvokeResult, bool) {
	if prev, known := w.idempotent[commandID]; known {
		if atNS < prev.expiresNS {
			return prev.callResult(), true
		}
		delete(w.idempotent, commandID)
	}
	return nil, false
}

// Safety refusal is terminal and may execute its independent autonomous action.
func (w *World) refuseInterlock(eff *model.Effector, inv invocation) bool {
	if eff.Interlock == nil || !w.interlockHolds(inv.entityID, eff.Interlock, inv.atNS) {
		return false
	}
	if action := eff.Interlock.AutonomousAction; action != nil {
		w.applyAutonomousAction(inv.entityID, action.State, action.Delta, inv.atNS)
	}
	w.recordCall(inv, callOutcome{mode: ModeReject, interlock: true, reason: "interlock_refused"})
	return true
}

func (w *World) executeInvocation(eff *model.Effector, inv invocation) *InvokeResult {
	mode := w.pickFailureMode(inv.entityID, eff)
	latency := w.ackLatency(inv.entityID, eff, mode)
	result := w.executeEffectorMode(eff, inv, mode, latency)
	w.recordCall(inv, callOutcome{
		mode: mode, accepted: result.Accepted, reason: result.Reason,
		latencyMS: latency, effectApplied: result.EffectApplied,
	})
	w.cacheInvocation(eff, inv.commandID, inv.atNS)
	return result
}

func (w *World) cacheInvocation(eff *model.Effector, commandID string, atNS int64) {
	window := eff.IdempotencyWindowS
	if window <= 0 {
		window = 3600
	}
	w.idempotent[commandID] = &idempotentResult{call: w.effectorCalls[len(w.effectorCalls)-1], expiresNS: atNS + int64(window*secondsPerNS)}
}

func (w *World) executeEffectorMode(eff *model.Effector, inv invocation, mode string, latency float64) *InvokeResult {
	result := &InvokeResult{
		Simulated:    true,
		WorldID:      w.ID,
		CommandID:    inv.commandID,
		AckLatencyMS: latency,
		Mode:         mode,
	}
	w.applyEffectorOutcome(result, eff, inv)
	return result
}

func (w *World) applyEffectorOutcome(result *InvokeResult, eff *model.Effector, inv invocation) {
	switch result.Mode {
	case ModeOK, ModeSlow, ModeAckLost:
		w.applyAcknowledgedEffect(result, eff, inv)
	case ModeReject:
		result.Reason = "effector_refused"
	case ModePartial, ModeSilentNoEffect:
		w.applyUnconfirmedEffect(result, eff, inv)
	case ModeConfirmedNoEffect:
		result.Accepted = true // The ack lies; confirmation channels report truth.
	}
}

func (w *World) applyAcknowledgedEffect(result *InvokeResult, eff *model.Effector, inv invocation) {
	w.applyEffect(inv.entityID, eff, inv.args, inv.atNS, 1.0, false)
	result.EffectApplied = true
	result.Accepted = result.Mode != ModeAckLost
	if !result.Accepted {
		result.Reason = "ack_lost"
	}
	result.EffectETANS = inv.atNS + int64(eff.Effect.DeadTimeS*secondsPerNS)
}

// Silent-no-effect applies to shadow state only; partial applies half physically.
func (w *World) applyUnconfirmedEffect(result *InvokeResult, eff *model.Effector, inv invocation) {
	strength, shadow := 0.5, false
	if result.Mode == ModeSilentNoEffect {
		strength, shadow = 1.0, true
	}
	w.applyEffect(inv.entityID, eff, inv.args, inv.atNS, strength, shadow)
	result.EffectApplied = true
	result.Accepted = true
	result.EffectETANS = inv.atNS + int64(eff.Effect.DeadTimeS*secondsPerNS)
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
