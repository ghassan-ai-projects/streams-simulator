package director

import (
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/app"
	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
)

type WorldRequest = layer.WorldRequest

type Code = rules.Code

type Nameplate = rules.Nameplate

type EffectorInfo = rules.EffectorInfo

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
