package app

import (
	"errors"
	"io/fs"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

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

func checkFailureCode(err error) (failureCode, bool) {
	if outcome, ok := asOutcomeError(err); ok {
		return outcome.code, true
	}
	if errors.Is(err, run.ErrDigestMismatch) {
		return codeReplayMismatch, true
	}
	return "", false
}

var contractFailureCodes = map[model.Contract]failureCode{
	model.ContractDomainSpec:      codeDomainInvalid,
	model.ContractOutputAdapter:   codeAdapterInvalid,
	model.ContractScenario:        codeScenarioInvalid,
	model.ContractRunArtifact:     codeArtifactInvalid,
	model.ContractConsumerVerdict: codeArtifactInvalid,
	model.ContractGroundTruth:     codeArtifactInvalid,
	model.ContractDeliveryLedger:  codeArtifactInvalid,
	model.ContractStateHistory:    codeArtifactInvalid,
	model.ContractComponent:       codeComponentInvalid,
	model.ContractInstrument:      codeComponentInvalid,
	model.ContractVocabulary:      codeComponentInvalid,
	model.ContractComposition:     codeCompositionInvalid,
}

func contractFailureCode(contract model.Contract) (failureCode, bool) {
	code, known := contractFailureCodes[contract]
	return code, known
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

func exitStatusOf(err error) int {
	if isUsageFailure(err) {
		return 2
	}
	return 1
}
