package director

import (
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/app"
	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
)

// WorldRequest is what a caller asks sim.world.create for: domain, seed,
// adapter, sink and time mode. Empty strings select the documented defaults.
type WorldRequest = layer.WorldRequest

// Code is a stable refusal code from the taxonomy in MCP_SURFACE section 7.
type Code = rules.Code

// Nameplate is the static description of a world the operator may read.
type Nameplate = rules.Nameplate

// EffectorInfo is one declared effector, as the nameplate describes it.
type EffectorInfo = rules.EffectorInfo

// Error taxonomy. Operator-facing codes are deliberately coarse so the text
// cannot leak state.
const (
	CodeDomainInvalid       = rules.CodeDomainInvalid
	CodeAdapterInvalid      = rules.CodeAdapterInvalid
	CodeWorldNotFound       = rules.CodeWorldNotFound
	CodeClockBackwards      = rules.CodeClockBackwards
	CodeConsumerNotQuiesced = rules.CodeConsumerNotQuiesced
	CodeProfileRateExceeded = rules.CodeProfileRateExceeded
	CodeTruthSealed         = rules.CodeTruthSealed
	CodeRunUnblinded        = rules.CodeRunUnblinded
	CodeUnknownEffector     = rules.CodeUnknownEffector
	CodeInvalidArgs         = rules.CodeInvalidArgs
	CodeMissingCommandID    = rules.CodeMissingCommandID
	CodeCapabilityDenied    = rules.CodeCapabilityDenied
	CodeEffectorRefused     = rules.CodeEffectorRefused
	CodeInterlockRefused    = rules.CodeInterlockRefused
	CodeNotImplemented      = rules.CodeNotImplemented
)
