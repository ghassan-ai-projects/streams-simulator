package world

import (
	"errors"

	layer "github.com/ghassan-ai-projects/streams-simulator/internal/world/internal/domain"
)

type Options = layer.Options

type InvokeResult = layer.InvokeResult

type EffectorCall = layer.EffectorCall

type FaultInfo = layer.FaultInfo

type Entity = layer.Entity

type FailureMode = layer.FailureMode

const (
	ModeOK                = layer.ModeOK
	ModeSlow              = layer.ModeSlow
	ModeAckLost           = layer.ModeAckLost
	ModeReject            = layer.ModeReject
	ModePartial           = layer.ModePartial
	ModeConfirmedNoEffect = layer.ModeConfirmedNoEffect
	ModeSilentNoEffect    = layer.ModeSilentNoEffect
)

var ErrInterlockRefused = layer.ErrInterlockRefused

var ErrCommandIDReused = layer.ErrCommandIDReused

var ErrInvalidInvocation = layer.ErrInvalidInvocation

const ReasonPreconditionRefused = layer.ReasonPreconditionRefused

const MaxStartNS = layer.MaxStartNS

var ErrClockBackwards = layer.ErrClockBackwards

var ErrFaultOnsetInPast = layer.ErrFaultOnsetInPast

var ErrFaultClearInPast = layer.ErrFaultClearInPast

var ErrNoSpec = errors.New("world: a domain spec is required")
