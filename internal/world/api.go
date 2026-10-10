package world

import (
	"errors"

	layer "github.com/ghassan-ai-projects/streams-simulator/internal/world/internal/domain"
)

// Options tune world construction. Everything here is part of the
// determinism tuple when it changes output.
type Options = layer.Options

// InvokeResult is what an effector invocation returns.
type InvokeResult = layer.InvokeResult

// EffectorCall is the director-side record of one effector invocation: the
// authority when scoring actions, not the consumer's claim.
type EffectorCall = layer.EffectorCall

// FaultInfo describes one active fault instance.
type FaultInfo = layer.FaultInfo

// Entity is one simulated producer, as hosts may read it.
type Entity = layer.Entity

// Effector failure modes: the closed vocabulary of InvokeResult.Mode.
const (
	ModeOK                = layer.ModeOK
	ModeSlow              = layer.ModeSlow
	ModeAckLost           = layer.ModeAckLost
	ModeReject            = layer.ModeReject
	ModePartial           = layer.ModePartial
	ModeConfirmedNoEffect = layer.ModeConfirmedNoEffect
	ModeSilentNoEffect    = layer.ModeSilentNoEffect
)

// ErrInterlockRefused is the terminal refusal of an independent safety
// system. Refusal is not a retryable error.
var ErrInterlockRefused = layer.ErrInterlockRefused

// ErrClockBackwards is returned by Advance when the target time is before the
// current world time.
var ErrClockBackwards = layer.ErrClockBackwards

// ErrNoSpec is returned by New when no domain spec is given.
var ErrNoSpec = errors.New("world: a domain spec is required")
