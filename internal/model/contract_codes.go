package model

import "errors"

var contractCodes = map[Contract]string{
	ContractDomainSpec:      "domain_invalid",
	ContractOutputAdapter:   "adapter_invalid",
	ContractScenario:        "scenario_invalid",
	ContractRunArtifact:     "artifact_invalid",
	ContractConsumerVerdict: "artifact_invalid",
	ContractGroundTruth:     "artifact_invalid",
	ContractDeliveryLedger:  "artifact_invalid",
	ContractStateHistory:    "artifact_invalid",
	ContractComponent:       "component_invalid",
	ContractInstrument:      "component_invalid",
	ContractVocabulary:      "component_invalid",
	ContractComposition:     "composition_invalid",
	ContractKindFile:        "kind_failed",
	ContractTwinChoices:     "kind_failed",
	ContractPlantWireBirth:  "kind_failed",
	ContractBoardFile:       "kind_failed",
	ContractSessionPlan:     "plantwire_failed",
	ContractNodeControl:     "plantwire_failed",
	ContractInterruptLog:    "plantwire_failed",
	ContractTwinReport:      "plantwire_failed",
	ContractRunComparison:   "report_failed",
}

func (c Contract) FailureCode() (string, bool) {
	code, known := contractCodes[c]
	return code, known
}

func FailureCodeOf(err error) (string, bool) {
	invalid, ok := AsValidationError(err)
	if !ok {
		return "", false
	}
	return invalid.Contract.FailureCode()
}

func AsValidationError(err error) (*ValidationError, bool) {
	var invalid *ValidationError
	return invalid, errors.As(err, &invalid)
}
