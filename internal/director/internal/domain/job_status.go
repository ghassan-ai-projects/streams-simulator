package domain

import (
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type JobState string

const (
	JobRunning   JobState = "running"
	JobPaused    JobState = "paused"
	JobStopping  JobState = "stopping"
	JobCompleted JobState = "completed"
	JobStopped   JobState = "stopped"
	JobFailed    JobState = "failed"
)

func JobActive(state JobState) bool {
	return state == JobRunning || state == JobPaused || state == JobStopping
}

type JobCallOutcome struct {
	AtS         int64  `json:"at_s"`
	EntityID    string `json:"entity_id"`
	Effector    string `json:"effector"`
	Accepted    bool   `json:"accepted"`
	Reason      string `json:"reason,omitempty"`
	EffectETANS int64  `json:"effect_eta_ns,omitempty"`
	Immediate   bool   `json:"immediate,omitempty"`
}

type JobFaultView struct {
	JobFault
	FaultID string `json:"fault_id,omitempty"`
}

type JobStatus struct {
	WorldID        string            `json:"world_id"`
	RunID          string            `json:"run_id"`
	Name           string            `json:"name,omitempty"`
	Purpose        string            `json:"purpose,omitempty"`
	Domain         string            `json:"domain"`
	Seed           int64             `json:"seed"`
	State          JobState          `json:"state"`
	StartNS        int64             `json:"start_ns"`
	ReachedNS      int64             `json:"reached_ns"`
	RequestedEndNS int64             `json:"requested_end_ns"`
	Rate           float64           `json:"rate"`
	Explore        bool              `json:"explore"`
	Faults         []JobFaultView    `json:"faults"`
	Perturbations  []JobPerturbation `json:"perturbations"`
	Calls          []JobCallOutcome  `json:"calls"`
	Failure        string            `json:"failure,omitempty"`
}

const noReasonGiven = "no reason given"

func NewJobCallOutcome(atS int64, entityID, effector string, ack *model.EffectorAcknowledgement) JobCallOutcome {
	outcome := JobCallOutcome{AtS: atS, EntityID: entityID, Effector: effector}
	if ack == nil {
		outcome.Reason = noReasonGiven
		return outcome
	}
	outcome.Accepted, outcome.EffectETANS = ack.Accepted, ack.EffectETANS
	if !ack.Accepted {
		outcome.Reason = ack.Reason
		if outcome.Reason == "" {
			outcome.Reason = noReasonGiven
		}
	}
	return outcome
}

func NewImmediateCallOutcome(atS int64, entityID, effector string, ack *model.EffectorAcknowledgement) JobCallOutcome {
	outcome := NewJobCallOutcome(atS, entityID, effector, ack)
	outcome.Immediate = true
	return outcome
}

func NewJobStatus(worldID, runID string, req JobRequest, startNS int64) JobStatus {
	return JobStatus{WorldID: worldID, RunID: runID, Name: req.Name, Purpose: req.Purpose, Domain: req.Domain, Seed: req.Seed,
		State: JobRunning, StartNS: startNS, ReachedNS: startNS, RequestedEndNS: startNS + req.DurationS*nsPerSecond,
		Rate: req.Rate, Explore: req.Explore, Faults: faultViews(req.Faults), Perturbations: slices.Clone(orEmpty(req.Perturbations)),
		Calls: []JobCallOutcome{}}
}

func faultViews(faults []JobFault) []JobFaultView {
	views := make([]JobFaultView, len(faults))
	for i, fault := range faults {
		views[i] = JobFaultView{JobFault: fault}
	}
	return views
}

func orEmpty(perturbations []JobPerturbation) []JobPerturbation {
	if perturbations == nil {
		return []JobPerturbation{}
	}
	return perturbations
}

func JobStatusView(status JobStatus) JobStatus {
	status.Calls = slices.Clone(status.Calls)
	if !status.Explore && JobActive(status.State) {
		status.Faults, status.Perturbations, status.Failure = []JobFaultView{}, []JobPerturbation{}, ""
		return status
	}
	status.Faults = slices.Clone(status.Faults)
	status.Perturbations = slices.Clone(status.Perturbations)
	return status
}
