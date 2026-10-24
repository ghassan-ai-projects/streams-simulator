package model

import "github.com/ghassan-ai-projects/streams-simulator/internal/schemas"

func ValidateDomainAssumptions(raw []byte) error {
	return validateAgainst(raw, "domain assumptions", ContractDomainAssumptions, schemas.DomainAssumptions())
}

func ValidateDomainLineage(raw []byte) error {
	return validateAgainst(raw, "domain lineage", ContractDomainLineage, schemas.DomainLineage())
}

func ValidateDomainDiff(raw []byte) error {
	return validateAgainst(raw, "domain diff", ContractDomainDiff, schemas.DomainDiff())
}
