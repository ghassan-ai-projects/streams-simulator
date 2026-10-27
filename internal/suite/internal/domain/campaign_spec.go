package domain

import (
	"fmt"
	"maps"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
)

const (
	CampaignSpecVersion = "0.1"
	maxCampaignRuns     = 4096
	maxRunSeconds       = 2592000
	maxPhaseName        = 64
	maxFaultWeight      = 1e6
)

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
	for _, check := range []func() error{s.checkHeader, s.checkTiming, s.checkSize} {
		if err := check(); err != nil {
			return err
		}
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
	case len(s.Phases) == 0:
		return specErrorf("at least one phase is required")
	}
	return nil
}

func (s *CampaignSpec) checkTiming() error {
	switch {
	case s.DurationS <= 0 || s.DurationS > maxRunSeconds:
		return specErrorf("duration_s must be above 0 and at most %d", maxRunSeconds)
	case s.NegativeDurationS <= 0 || s.FaultWindowS <= 0:
		return specErrorf("negative_duration_s and fault_start_window_s must be positive")
	case s.FaultWindowS+s.NegativeDurationS > s.DurationS:
		return specErrorf("fault_start_window_s plus negative_duration_s must fit inside duration_s")
	}
	return nil
}

func (s *CampaignSpec) checkSize() error {
	total := 0
	for _, phase := range s.Phases {
		if len(phase.Name) > maxPhaseName || phase.Runs > maxCampaignRuns {
			return specErrorf("phase %q: a name is at most %d characters and a phase at most %d runs", phase.Name, maxPhaseName, maxCampaignRuns)
		}
		total += max(phase.Runs, 0)
	}
	if total > maxCampaignRuns {
		return specErrorf("a campaign has at most %d runs, got %d", maxCampaignRuns, total)
	}
	return nil
}
