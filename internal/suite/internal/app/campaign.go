package app

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	rules "github.com/ghassan-ai-projects/streams-simulator/internal/suite/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/suite/internal/files"
)

type CampaignOptions struct {
	DomainsDir, SpecPath, Out string
}

type PublishedCampaign struct {
	Path string
	Runs int
}

func PublishCampaign(options CampaignOptions) (PublishedCampaign, error) {
	spec, err := readCampaignSpec(options.SpecPath)
	if err != nil {
		return PublishedCampaign{}, err
	}
	domains, err := loadCampaignDomains(options.DomainsDir, spec)
	if err != nil {
		return PublishedCampaign{}, err
	}
	campaign, plans, err := rules.BuildCampaign(spec, domains)
	if err != nil {
		return PublishedCampaign{}, fmt.Errorf("streamsim: %w", err)
	}
	return writeCampaign(options.Out, campaign, plans)
}

func readCampaignSpec(path string) (*rules.CampaignSpec, error) {
	raw, err := files.Read(path)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	var spec rules.CampaignSpec
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&spec); err != nil {
		return nil, fmt.Errorf("streamsim: %w", &rules.CampaignSpecError{Message: err.Error()})
	}
	return &spec, nil
}

func loadCampaignDomains(dir string, spec *rules.CampaignSpec) (map[string]*domain.Compiled, error) {
	catalog, err := domain.NewCatalogFromDir(dir)
	if err != nil {
		return nil, fmt.Errorf("streamsim: %w", err)
	}
	domains := map[string]*domain.Compiled{}
	for _, id := range spec.DomainIDs() {
		if domains[id], err = catalog.Describe(id); err != nil {
			return nil, fmt.Errorf("streamsim: %w", err)
		}
	}
	return domains, nil
}
