package model

import "github.com/ghassan-ai-projects/streams-simulator/internal/schemas"

func ValidateRunComparison(raw []byte) error {
	return validateAgainst(raw, "run comparison", ContractRunComparison, schemas.RunComparison())
}
