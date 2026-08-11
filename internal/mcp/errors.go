// Package mcp implements one MCP server with two roles: director (build,
// drive, perturb and audit worlds; holds the truth) and operator (act on a
// world as a consumer of it; sees no truth, no faults, no hidden state).
// Every tool is domain-agnostic: the binary contains no effector name, no
// consumer name, no consumer schema.
package mcp

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
