package domain

import "github.com/ghassan-ai-projects/streams-simulator/internal/model"

type JobState string

const (
	JobRunning   JobState = "running"
	JobStopping  JobState = "stopping"
	JobCompleted JobState = "completed"
	JobStopped   JobState = "stopped"
	JobFailed    JobState = "failed"
)

func JobActive(state JobState) bool {
	return state == JobRunning || state == JobStopping
}

type JobCallOutcome struct {
	AtS         int64  `json:"at_s"`
	EntityID    string `json:"entity_id"`
	Effector    string `json:"effector"`
	Accepted    bool   `json:"accepted"`
	Reason      string `json:"reason,omitempty"`
	EffectETANS int64  `json:"effect_eta_ns,omitempty"`
}

type JobStatus struct {
	WorldID        string           `json:"world_id"`
	RunID          string           `json:"run_id"`
	Domain         string           `json:"domain"`
	Seed           int64            `json:"seed"`
	State          JobState         `json:"state"`
	StartNS        int64            `json:"start_ns"`
	ReachedNS      int64            `json:"reached_ns"`
	RequestedEndNS int64            `json:"requested_end_ns"`
	Rate           float64          `json:"rate"`
	Explore        bool             `json:"explore"`
	Calls          []JobCallOutcome `json:"calls"`
	Failure        string           `json:"failure,omitempty"`
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
