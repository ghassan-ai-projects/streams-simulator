// Package domain holds the director module's vocabulary: the refusal
// taxonomy (stable codes and the typed failure that carries them) and the
// operator view, the only surface a consumer is given. It is pure: it reads
// no clock and touches no file or socket.
package domain

import "fmt"

// Code is a stable error code from the taxonomy in MCP_SURFACE §7.
type Code string

// Error taxonomy. Operator-facing codes are deliberately coarse so the
// text cannot leak state.
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
)

// ToolError is a typed tool failure carrying a stable code. The message
// never carries fault-specific detail on operator-facing paths.
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

// ToolErrorf builds a coded tool failure for the protocol layer's own checks.
func ToolErrorf(code Code, format string, args ...any) *ToolError {
	return errTool(code, format, args...)
}
