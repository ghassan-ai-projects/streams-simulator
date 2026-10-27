package domain

import (
	"fmt"
	"math"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/randutil"
)

const (
	CampaignSchemaVersion = "0.1"
	CampaignFileName      = "campaign.json"
	maxSafeSeed           = 1<<53 - 1
)

type Campaign struct {
	SchemaVersion string        `json:"schema_version"`
	Domain        string        `json:"domain"`
	Seed          uint64        `json:"seed"`
	Runs          []CampaignRun `json:"runs"`
}

type CampaignRun struct {
	Index       int     `json:"index"`
	Phase       string  `json:"phase"`
	Domain      string  `json:"domain"`
	Seed        uint64  `json:"seed"`
	DurationS   float64 `json:"duration_s"`
	SessionPlan string  `json:"session_plan"`
}

func PlanFileName(index int) string { return fmt.Sprintf("run-%04d.session-plan.json", index) }

func IsCampaignFile(name string) bool {
	return name == CampaignFileName || (strings.HasPrefix(name, "run-") && strings.HasSuffix(name, ".session-plan.json"))
}

type campaignBuild struct {
	spec     *CampaignSpec
	domains  map[string]*domain.Compiled
	campaign *Campaign
	plans    []model.SessionPlan
}

func BuildCampaign(spec *CampaignSpec, domains map[string]*domain.Compiled) (*Campaign, []model.SessionPlan, error) {
	if err := spec.Check(domains); err != nil {
		return nil, nil, err
	}
	b := &campaignBuild{spec: spec, domains: domains,
		campaign: &Campaign{SchemaVersion: CampaignSchemaVersion, Domain: spec.Phases[0].Domain, Seed: spec.Seed}}
	for i := range spec.Phases {
		b.addPhase(&spec.Phases[i])
	}
	return b.campaign, b.plans, nil
}

func (b *campaignBuild) addPhase(phase *CampaignPhase) {
	compiled := b.domains[phase.Domain]
	negative, _ := negativeFaultOf(compiled)
	rng := randutil.Substream(b.spec.Seed, "campaign/phase/"+phase.Name)
	for _, fault := range phase.schedule(negative, rng) {
		b.addRun(phase, fault, fault == negative)
	}
}

func (b *campaignBuild) addRun(phase *CampaignPhase, fault string, transient bool) {
	index := len(b.campaign.Runs) + 1
	rng := randutil.Substream(b.spec.Seed, fmt.Sprintf("campaign/run/%d", index))
	seed := rng.Next() & maxSafeSeed
	disturbance := b.disturbance(phase.Domain, fault, transient, rng)
	b.campaign.Runs = append(b.campaign.Runs, CampaignRun{Index: index, Phase: phase.Name, Domain: phase.Domain, Seed: seed,
		DurationS: b.spec.DurationS, SessionPlan: PlanFileName(index)})
	b.plans = append(b.plans, b.sessionPlan(phase.Domain, seed, disturbance))
}

func (b *campaignBuild) disturbance(domainID, fault string, transient bool, rng *randutil.SplitMix64) model.Disturbance {
	compiled := b.domains[domainID]
	entities := faultEntities(compiled, fault)
	disturbance := model.Disturbance{Instance: entities[rng.Intn(len(entities))], Fault: fault,
		AtS: math.Floor(rng.Float64() * b.spec.FaultWindowS)}
	if transient && !isPulse(compiled, disturbance.Instance, fault) {
		duration := b.spec.NegativeDurationS
		disturbance.DurationS = &duration
	}
	return disturbance
}

func (b *campaignBuild) sessionPlan(domainID string, seed uint64, disturbance model.Disturbance) model.SessionPlan {
	return model.SessionPlan{SchemaVersion: model.SessionPlanSchemaVersion, Domain: domainID, Seed: &seed,
		HostSocket: b.spec.HostSocket, Policies: b.spec.Policies, Disturbances: []model.Disturbance{disturbance}}
}
