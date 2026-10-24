package director

import (
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/app"
	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
)

type WorldRequest = layer.WorldRequest

type Code = rules.Code

type JobRequest = rules.JobRequest

type JobFault = rules.JobFault

type JobCall = rules.JobCall

type JobPerturbation = rules.JobPerturbation

type JobFaultView = rules.JobFaultView

type JobStatus = rules.JobStatus

type JobCallOutcome = rules.JobCallOutcome

type JobState = rules.JobState

type JobImmediate = rules.JobImmediate

type Nameplate = rules.Nameplate

type EffectorInfo = rules.EffectorInfo

const (
	CodeDomainInvalid    = rules.CodeDomainInvalid
	CodeInvalidArgs      = rules.CodeInvalidArgs
	CodeCapabilityDenied = rules.CodeCapabilityDenied
	CodeNotImplemented   = rules.CodeNotImplemented
	CodeJobActive        = rules.CodeJobActive
	CodeJobNotRunning    = rules.CodeJobNotRunning
	CodeJobNotPaused     = rules.CodeJobNotPaused
	CodeEvaluateBlind    = rules.CodeEvaluateBlind

	JobRunning   = rules.JobRunning
	JobPaused    = rules.JobPaused
	JobStopping  = rules.JobStopping
	JobCompleted = rules.JobCompleted
	JobStopped   = rules.JobStopped
	JobFailed    = rules.JobFailed
)
