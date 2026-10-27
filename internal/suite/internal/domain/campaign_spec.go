package domain

import (
	"fmt"
	"maps"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
)

const CampaignSpecVersion = "0.1"

type CampaignSpec struct {
	SchemaVersion     string          `json:"schema_version"`
	Seed              uint64          `json:"seed"`
	HostSocket        string          `json:"host_socket"`
	Policies          []string        `json:"policies,omitempty"`
	DurationS         float64         `json:"duration_s"`
	FaultWindowS      float64         `json:"fault_start_window_s"`
	NegativeDurationS float64         `json:"negative_duration_s"`
	Phases            []CampaignPhase `json:"phases"`
}

type CampaignPhase struct {
	Name          string             `json:"name"`
	Domain        string             `json:"domain"`
	Runs          int                `json:"runs"`
	NegativeShare float64            `json:"negative_share"`
	Faults        map[string]float64 `json:"faults"`
}

type CampaignSpecError struct{ Message string }

func (e *CampaignSpecError) Error() string { return "campaign spec: " + e.Message }

func specErrorf(format string, args ...any) error {
	return &CampaignSpecError{Message: fmt.Sprintf(format, args...)}
}

func (s *CampaignSpec) DomainIDs() []string {
	ids := map[string]bool{}
	for _, phase := range s.Phases {
		ids[phase.Domain] = true
	}
	return slices.Sorted(maps.Keys(ids))
}

func (s *CampaignSpec) Check(domains map[string]*domain.Compiled) error {
	if err := s.checkHeader(); err != nil {
		return err
	}
	seen := map[string]bool{}
	for i := range s.Phases {
		if err := s.Phases[i].checkAmong(seen, domains); err != nil {
			return err
		}
	}
	return nil
}

func (p *CampaignPhase) checkAmong(seen map[string]bool, domains map[string]*domain.Compiled) error {
	if seen[p.Name] {
		return specErrorf("phase %q is listed twice", p.Name)
	}
	seen[p.Name] = true
	return p.check(domains[p.Domain])
}

func (s *CampaignSpec) checkHeader() error {
	switch {
	case s.SchemaVersion != CampaignSpecVersion:
		return specErrorf("schema_version must be %q, got %q", CampaignSpecVersion, s.SchemaVersion)
	case s.HostSocket == "":
		return specErrorf("host_socket is required")
	case s.DurationS <= 0 || s.NegativeDurationS <= 0:
		return specErrorf("duration_s and negative_duration_s must be positive")
	case s.FaultWindowS <= 0 || s.FaultWindowS >= s.DurationS:
		return specErrorf("fault_start_window_s must be positive and shorter than duration_s")
	case len(s.Phases) == 0:
		return specErrorf("at least one phase is required")
	}
	return nil
}
