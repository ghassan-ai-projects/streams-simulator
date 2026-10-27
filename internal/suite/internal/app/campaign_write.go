package app

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	rules "github.com/ghassan-ai-projects/streams-simulator/internal/suite/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/suite/internal/files"
)

func writeCampaign(out string, campaign *rules.Campaign, plans []model.SessionPlan) (PublishedCampaign, error) {
	if err := checkCampaign(campaign); err != nil {
		return PublishedCampaign{}, err
	}
	for i := range plans {
		if err := writeJSON(out, rules.PlanFileName(i+1), plans[i]); err != nil {
			return PublishedCampaign{}, err
		}
	}
	if err := writeJSON(out, rules.CampaignFileName, campaign); err != nil {
		return PublishedCampaign{}, err
	}
	return PublishedCampaign{Path: filepath.Join(out, rules.CampaignFileName), Runs: len(campaign.Runs)}, nil
}

func writeJSON(dir, name string, document any) error {
	raw, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	if _, err := files.Write(dir, name, append(raw, '\n')); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	return nil
}

func checkCampaign(campaign *rules.Campaign) error {
	raw, err := json.Marshal(campaign)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	if err := model.ValidateCampaign(raw); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	return nil
}
