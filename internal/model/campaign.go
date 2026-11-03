package model

import "github.com/ghassan-ai-projects/streams-simulator/internal/schemas"

func ValidateCampaign(raw []byte) error {
	return validateAgainst(raw, "campaign", ContractCampaign, schemas.Campaign())
}
