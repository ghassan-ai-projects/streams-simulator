package model

import "github.com/ghassan-ai-projects/streams-simulator/internal/schemas"

func ValidateTwinReport(raw []byte) error {
	return validateAgainst(raw, "twin report", ContractTwinReport, schemas.TwinReport())
}
