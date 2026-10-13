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
	CodeDomainInvalid    = rules.CodeDomainInvalid
	CodeInvalidArgs      = rules.CodeInvalidArgs
	CodeCapabilityDenied = rules.CodeCapabilityDenied
	CodeNotImplemented   = rules.CodeNotImplemented
)
