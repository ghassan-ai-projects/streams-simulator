// Package domain holds the director module's vocabulary: the refusal
// taxonomy (stable codes and the typed failure that carries them) and the
// operator view, the only surface a consumer is given. It is pure: it reads
// no clock and touches no file or socket.
package domain

import "fmt"

type Code string

const (
	CodeDomainInvalid       Code = "domain_invalid"
	CodeAdapterInvalid      Code = "adapter_invalid"
	CodeWorldNotFound       Code = "world_not_found"
	CodeClockBackwards      Code = "clock_backwards"
	CodeConsumerNotQuiesced Code = "consumer_not_quiesced"
	CodeProfileRateExceeded Code = "profile_rate_exceeded"
	CodeTruthSealed         Code = "truth_sealed"
	CodeRunUnblinded        Code = "run_unblinded"
	CodeUnknownEffector     Code = "unknown_effector"
	CodeInvalidArgs         Code = "invalid_args"
	CodeMissingCommandID    Code = "missing_command_id"
	CodeCapabilityDenied    Code = "capability_denied"
	CodeEffectorRefused     Code = "effector_refused"
	CodeInterlockRefused    Code = "interlock_refused"
	CodeNotImplemented      Code = "not_implemented"
	CodeComponentInvalid    Code = "component_invalid"
	CodeCompositionInvalid  Code = "composition_invalid"
	CodeNoLibrary           Code = "no_component_library"
	CodeJobActive           Code = "job_active"
	CodeJobNotRunning       Code = "job_not_running"
	CodeJobNotPaused        Code = "job_not_paused"
	CodeEvaluateBlind       Code = "evaluate_blind"
)

type ToolError struct {
	Code Code
	Msg  string
}

func (e *ToolError) Error() string {
	return string(e.Code) + ": " + e.Msg
}

func errTool(code Code, format string, args ...any) *ToolError {
	return &ToolError{Code: code, Msg: fmt.Sprintf(format, args...)}
}

func ToolErrorf(code Code, format string, args ...any) *ToolError {
	return errTool(code, format, args...)
}
