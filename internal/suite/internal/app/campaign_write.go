package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	rules "github.com/ghassan-ai-projects/streams-simulator/internal/suite/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/suite/internal/files"
)

var ErrFolderTaken = errors.New("the output folder already holds a campaign; generate into a new folder")

func writeCampaign(out string, campaign *rules.Campaign, plans []model.SessionPlan) (PublishedCampaign, error) {
	if err := checkCampaign(campaign); err != nil {
		return PublishedCampaign{}, err
	}
	encoded, err := encodePlans(plans)
	if err != nil {
		return PublishedCampaign{}, err
	}
	if err := writeFiles(out, campaign, encoded); err != nil {
		return PublishedCampaign{}, err
	}
	return PublishedCampaign{Path: filepath.Join(out, rules.CampaignFileName), Runs: len(campaign.Runs)}, nil
}

func writeFiles(out string, campaign *rules.Campaign, encoded [][]byte) error {
	if err := refuseTakenFolder(out); err != nil {
		return err
	}
	for i, raw := range encoded {
		if err := writeBytes(out, rules.PlanFileName(i+1), raw); err != nil {
			return err
		}
	}
	return writeJSON(out, rules.CampaignFileName, campaign)
}

func refuseTakenFolder(out string) error {
	names, err := files.Names(out)
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	if slices.ContainsFunc(names, rules.IsCampaignFile) {
		return fmt.Errorf("streamsim: %w: %s", ErrFolderTaken, out)
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

func encodePlans(plans []model.SessionPlan) ([][]byte, error) {
	encoded := make([][]byte, len(plans))
	for i := range plans {
		raw, err := json.MarshalIndent(plans[i], "", "  ")
		if err != nil {
			return nil, fmt.Errorf("streamsim: %w", err)
		}
		if _, err := model.DecodeSessionPlan(raw); err != nil {
			return nil, fmt.Errorf("streamsim: %w", &rules.CampaignSpecError{
				Message: fmt.Sprintf("run %d: host_socket or policies break session-plan-v0.1: %v", i+1, err)})
		}
		encoded[i] = append(raw, '\n')
	}
	return encoded, nil
}

func writeJSON(dir, name string, document any) error {
	raw, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	return writeBytes(dir, name, append(raw, '\n'))
}

func writeBytes(dir, name string, raw []byte) error {
	if _, err := files.Write(dir, name, raw); err != nil {
		return fmt.Errorf("streamsim: %w", err)
	}
	return nil
}
