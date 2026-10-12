package app

import (
	"errors"
	"fmt"
)

// failureCode is the stable classification of a failed command
// (docs/design/CLI_SURFACE.md §3). The values are the `error.code` enum of
// cli-output-v0.1; a test holds the two in step.
type failureCode string

const (
	codeUsageInvalid        failureCode = "usage_invalid"
	codeFileNotFound        failureCode = "file_not_found"
	codeDomainNotFound      failureCode = "domain_not_found"
	codeAdapterNotFound     failureCode = "adapter_not_found"
	codeDomainInvalid       failureCode = "domain_invalid"
	codeAdapterInvalid      failureCode = "adapter_invalid"
	codeArtifactInvalid     failureCode = "artifact_invalid"
	codeScenarioInvalid     failureCode = "scenario_invalid"
	codeReplayMismatch      failureCode = "replay_mismatch"
	codeAdapterVerifyFailed failureCode = "adapter_verify_failed"
	codeIOFailed            failureCode = "io_failed"
)

// commandFailed is the code of a failure no other class describes.
func commandFailed(command string) failureCode {
	return failureCode(command + "_failed")
}

// usageError is an invocation the command cannot run as written: a missing
// or unknown verb, argument or flag value. It exits 2 with usage_invalid.
// cause, when set, is the violation behind it, whose issues the failure
// document carries.
type usageError struct {
	message string
	cause   error
}

func (e *usageError) Error() string { return e.message }

// Unwrap returns the violation behind the usage failure, if any.
func (e *usageError) Unwrap() error { return e.cause }

// usagef builds a usageError.
func usagef(format string, args ...any) error {
	return &usageError{message: fmt.Sprintf(format, args...)}
}

// outcomeError is a command that ran but whose check failed. It carries the
// command's result document so a tool can see what diverged.
type outcomeError struct {
	code    failureCode
	message string
	result  map[string]any
}

func (e *outcomeError) Error() string { return e.message }

func asUsageError(err error) (*usageError, bool) {
	var usage *usageError
	return usage, errors.As(err, &usage)
}

func asOutcomeError(err error) (*outcomeError, bool) {
	var outcome *outcomeError
	return outcome, errors.As(err, &outcome)
}
