package domain

import (
	"fmt"
	"math"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type JobFault struct {
	EntityID string `json:"entity_id"`
	Fault    string `json:"fault"`
	AtS      int64  `json:"at_s"`
	UntilS   *int64 `json:"until_s,omitempty"`
}

type JobPerturbation struct {
	Kind   string         `json:"kind"`
	Params map[string]any `json:"params,omitempty"`
	AtS    int64          `json:"at_s"`
	UntilS *int64         `json:"until_s,omitempty"`
}

type JobCall struct {
	EntityID string         `json:"entity_id"`
	Effector string         `json:"effector"`
	Args     map[string]any `json:"args,omitempty"`
	AtS      int64          `json:"at_s"`
}

type JobRequest struct {
	Name          string            `json:"name,omitempty"`
	Purpose       string            `json:"purpose,omitempty"`
	Domain        string            `json:"domain"`
	Seed          int64             `json:"seed"`
	Explore       bool              `json:"explore"`
	DurationS     int64             `json:"duration_s"`
	Rate          float64           `json:"rate"`
	Faults        []JobFault        `json:"faults"`
	Perturbations []JobPerturbation `json:"perturbations,omitempty"`
	Calls         []JobCall         `json:"calls"`
}

const (
	maxPlanSeconds = int64(1) << 32
	rateProblem    = "rate must be a finite number, 0 or more"
)

func CheckJob(req JobRequest) error {
	for _, problem := range []string{req.checkHeader(), req.checkFaults(), req.checkPerturbations(), req.checkCalls()} {
		if problem != "" {
			return ToolErrorf(CodeInvalidArgs, "job: %s", problem)
		}
	}
	return nil
}

func (req JobRequest) checkHeader() string {
	switch {
	case req.Domain == "":
		return "domain is required"
	case req.Seed < 0:
		return "seed must be 0 or more"
	case req.DurationS <= 0 || req.DurationS > maxPlanSeconds:
		return "duration_s must be a positive number of seconds"
	case !validRate(req.Rate):
		return rateProblem
	}
	return planTextProblem(req)
}

func planTextProblem(req JobRequest) string {
	if err := model.CheckPlanText(req.Name, req.Purpose); err != nil {
		return err.Error()
	}
	return ""
}

func (req JobRequest) checkFaults() string {
	for i, fault := range req.Faults {
		problem := req.checkAction(fault.EntityID, "fault", fault.Fault, fault.AtS)
		if problem == "" {
			problem = req.checkEnd(fault.AtS, fault.UntilS)
		}
		if problem != "" {
			return fmt.Sprintf("faults[%d].%s", i, problem)
		}
	}
	return ""
}

func (req JobRequest) checkPerturbations() string {
	for i, perturbation := range req.Perturbations {
		if problem := req.checkPerturbation(perturbation); problem != "" {
			return fmt.Sprintf("perturbations[%d].%s", i, problem)
		}
	}
	return ""
}

func (req JobRequest) checkPerturbation(perturbation JobPerturbation) string {
	if perturbation.Kind == "" {
		return "kind is required"
	}
	if perturbation.AtS < 0 || perturbation.AtS > req.DurationS {
		return "at_s must be within the run, 0 to duration_s"
	}
	return req.checkEnd(perturbation.AtS, perturbation.UntilS)
}

func (req JobRequest) checkEnd(atS int64, untilS *int64) string {
	if untilS != nil && (*untilS <= atS || *untilS > req.DurationS) {
		return "until_s must be after at_s and within the run, up to duration_s"
	}
	return ""
}

func (req JobRequest) checkCalls() string {
	for i, call := range req.Calls {
		if problem := req.checkAction(call.EntityID, "effector", call.Effector, call.AtS); problem != "" {
			return fmt.Sprintf("calls[%d].%s", i, problem)
		}
	}
	return ""
}

func (req JobRequest) checkAction(entityID, nameField, name string, atS int64) string {
	switch {
	case entityID == "":
		return "entity_id is required"
	case name == "":
		return nameField + " is required"
	case atS < 0 || atS > req.DurationS:
		return "at_s must be within the run, 0 to duration_s"
	}
	return ""
}

func validRate(rate float64) bool {
	return rate >= 0 && !math.IsNaN(rate) && !math.IsInf(rate, 0)
}
