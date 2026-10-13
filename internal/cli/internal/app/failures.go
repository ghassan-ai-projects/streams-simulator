package app

import (
	"errors"
	"fmt"
)

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
	codeComponentInvalid    failureCode = "component_invalid"
	codeCompositionInvalid  failureCode = "composition_invalid"
	codeReplayMismatch      failureCode = "replay_mismatch"
	codeAdapterVerifyFailed failureCode = "adapter_verify_failed"
	codeIOFailed            failureCode = "io_failed"
)

func commandFailed(command string) failureCode {
	return failureCode(command + "_failed")
}

type usageError struct {
	message string
	cause   error
}

func (e *usageError) Error() string { return e.message }

func (e *usageError) Unwrap() error { return e.cause }

func usagef(format string, args ...any) error {
	return &usageError{message: fmt.Sprintf(format, args...)}
}

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
