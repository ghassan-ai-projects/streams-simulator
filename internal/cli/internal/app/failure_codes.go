package app

import (
	"errors"
	"io/fs"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

// classifyFailure applies the precedence of CLI_SURFACE §3: usage, a failed
// check, a contract violation, an unknown id, a missing path, another I/O
// failure, then the command's own code. It uses typed errors only.
func classifyFailure(command string, err error) failureCode {
	if code, ok := knownFailureCode(err); ok {
		return code
	}
	return commandFailed(command)
}

func knownFailureCode(err error) (failureCode, bool) {
	if isUsageFailure(err) {
		return codeUsageInvalid, true
	}
	if code, ok := checkFailureCode(err); ok {
		return code, true
	}
	if invalid, ok := asValidationError(err); ok {
		return contractFailureCode(invalid.Contract)
	}
	return missingOrIOCode(err)
}

func isUsageFailure(err error) bool {
	_, flagUsage := asUsageExit(err)
	_, argumentUsage := asUsageError(err)
	return flagUsage || argumentUsage
}

// checkFailureCode classifies a check that ran and failed: an outcome the
// command reports, or a replay whose recorded inputs no longer match.
func checkFailureCode(err error) (failureCode, bool) {
	if outcome, ok := asOutcomeError(err); ok {
		return outcome.code, true
	}
	if errors.Is(err, run.ErrDigestMismatch) {
		return codeReplayMismatch, true
	}
	return "", false
}

// contractFailureCode maps every contract to its code; a contract without a
// mapping is a programming error a test catches.
func contractFailureCode(contract model.Contract) (failureCode, bool) {
	switch contract {
	case model.ContractDomainSpec:
		return codeDomainInvalid, true
	case model.ContractOutputAdapter:
		return codeAdapterInvalid, true
	case model.ContractRunArtifact, model.ContractConsumerVerdict, model.ContractGroundTruth,
		model.ContractDeliveryLedger, model.ContractStateHistory:
		return codeArtifactInvalid, true
	}
	return "", false
}

func missingOrIOCode(err error) (failureCode, bool) {
	var pathFailure *fs.PathError
	switch {
	case errors.Is(err, domain.ErrUnknownDomain):
		return codeDomainNotFound, true
	case errors.Is(err, adapter.ErrUnknownAdapter):
		return codeAdapterNotFound, true
	case errors.Is(err, fs.ErrNotExist):
		return codeFileNotFound, true
	case errors.As(err, &pathFailure):
		return codeIOFailed, true
	}
	return "", false
}

// exitStatusOf is 2 for a usage failure and 1 for any other.
func exitStatusOf(err error) int {
	if isUsageFailure(err) {
		return 2
	}
	return 1
}
