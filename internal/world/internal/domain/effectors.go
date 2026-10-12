package domain

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type FailureMode string

const (
	ModeOK                FailureMode = "ok"
	ModeSlow              FailureMode = "slow"
	ModeAckLost           FailureMode = "ack_lost"
	ModeReject            FailureMode = "reject"
	ModePartial           FailureMode = "partial"
	ModeConfirmedNoEffect FailureMode = "confirmed_no_effect"
	ModeSilentNoEffect    FailureMode = "silent_no_effect"
)

const ReasonAcknowledgementTimeout = "acknowledgement_timeout"

func Acknowledge(r *InvokeResult) *model.EffectorAcknowledgement {
	ack := &model.EffectorAcknowledgement{Accepted: r.Accepted, Simulated: r.Simulated, WorldID: r.WorldID,
		CommandID: r.CommandID, EffectETANS: r.EffectETANS, Reason: r.Reason}
	if r.Mode == ModeAckLost {
		ack.EffectETANS, ack.Reason = 0, ReasonAcknowledgementTimeout
	}
	return ack
}

func KnownFailureMode(mode FailureMode) bool {
	switch mode {
	case ModeOK, ModeSlow, ModeAckLost, ModeReject, ModePartial, ModeConfirmedNoEffect, ModeSilentNoEffect:
		return true
	}
	return false
}

type InvokeResult struct {
	Accepted      bool        `json:"accepted"`
	Simulated     bool        `json:"simulated"`
	WorldID       string      `json:"world_id"`
	CommandID     string      `json:"command_id"`
	EffectETANS   int64       `json:"effect_eta_ns,omitempty"`
	Reason        string      `json:"reason,omitempty"`
	Mode          FailureMode `json:"mode,omitempty"`
	AckLatencyMS  float64     `json:"ack_latency_ms,omitempty"`
	EffectApplied bool        `json:"effect_applied,omitempty"`
	Cause         string      `json:"-"`
}

var ErrInterlockRefused = fmt.Errorf("world: interlock refused")

var ErrCommandIDReused = errors.New("world: command_id already used for a different request")

type invocation struct {
	effector  string
	entityID  string
	commandID string
	args      map[string]any
	atNS      int64
}

type callOutcome struct {
	mode          FailureMode
	accepted      bool
	interlock     bool
	reason        string
	latencyMS     float64
	effectApplied bool
}

func (w *World) InvokeEffector(effector, entityID, commandID string, args map[string]any, atNS int64) (*InvokeResult, error) {
	inv := invocation{effector: effector, entityID: entityID, commandID: commandID, args: args, atNS: atNS}
	eff, err := w.admitInvocation(inv)
	if err != nil {
		return nil, invalidInvocation{err: err}
	}
	if result, known, err := w.replayInvocation(inv); err != nil || known {
		return result, err
	}
	if refusal, err := w.refuseInvocation(eff, inv); refusal != nil || err != nil {
		return refusal, err
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
	eff := ent.kind.EffectorDecl(effector)
	if eff == nil {
		return nil, fmt.Errorf("world: unknown effector %q", effector)
	}
	return eff, nil
}

func (w *World) replayInvocation(inv invocation) (*InvokeResult, bool, error) {
	prev, known := w.idempotent[inv.commandID]
	if !known {
		return nil, false, nil
	}
	if inv.atNS >= prev.expiresNS {
		delete(w.idempotent, inv.commandID)
		return nil, false, nil
	}
	if prev.request != requestKey(inv) {
		return nil, false, fmt.Errorf("%w: %q", ErrCommandIDReused, inv.commandID)
	}
	replayed := prev.result
	return &replayed, true, nil
}

func requestKey(inv invocation) string {
	args := inv.args
	if args == nil {
		args = map[string]any{}
	}
	raw, err := json.Marshal([]any{inv.effector, inv.entityID, numbersByValue(args)})
	if err != nil {

		return fmt.Sprintf("%q|%q|%v", inv.effector, inv.entityID, args)
	}
	return string(raw)
}

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
	w.cacheInvocation(eff, inv, result)
	return result
}

func (w *World) cacheInvocation(eff *model.Effector, inv invocation, result *InvokeResult) {
	window := eff.IdempotencyWindowS
	if window <= 0 {
		window = 3600
	}
	w.idempotent[inv.commandID] = &idempotentResult{
		request: requestKey(inv), result: *result, expiresNS: inv.atNS + int64(window*secondsPerNS),
	}
}

func (w *World) executeEffectorMode(eff *model.Effector, inv invocation, mode FailureMode, latency float64) *InvokeResult {
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

		result.Accepted = true
		result.EffectETANS = inv.atNS + int64(eff.Effect.DeadTimeS*secondsPerNS)
	}
}

func (w *World) applyAcknowledgedEffect(result *InvokeResult, eff *model.Effector, inv invocation) {
	w.applyEffect(inv.entityID, eff, inv.args, inv.atNS, 1.0, false)
	w.followUpCall(eff, inv)
	result.EffectApplied = true
	result.Accepted = result.Mode != ModeAckLost
	if !result.Accepted {
		result.Reason = "ack_lost"
	}
	result.EffectETANS = inv.atNS + int64(eff.Effect.DeadTimeS*secondsPerNS)
}

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
