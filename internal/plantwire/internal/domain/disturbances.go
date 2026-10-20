package domain

import (
	"fmt"
	"math"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const (
	KeywordUnknownFault              = "unknown_fault"
	KeywordDisturbanceOnRealInstance = "disturbance_on_real_instance"
	KeywordDurationWithPulseFault    = "duration_with_pulse_fault"
	KeywordFaultOnCommandedState     = "fault_on_commanded_state"
)

type PlanFault struct {
	ID        string
	Pulse     bool
	Commanded bool
}

type ScheduledFault struct {
	Entity  string
	Fault   string
	OnsetNS int64
	ClearNS int64
	Params  map[string]any
}

func ScheduleDisturbances(disturbances []model.Disturbance, startNS int64) []ScheduledFault {
	scheduled := make([]ScheduledFault, len(disturbances))
	for i, d := range disturbances {
		scheduled[i] = ScheduledFault{Entity: d.Instance, Fault: d.Fault, OnsetNS: startNS + nanoseconds(d.AtS)}
		if d.DurationS != nil {
			scheduled[i].ClearNS = scheduled[i].OnsetNS + nanoseconds(*d.DurationS)
		}
		if d.Severity != nil {
			scheduled[i].Params = map[string]any{"severity": *d.Severity}
		}
	}
	return scheduled
}

func nanoseconds(seconds float64) int64 {
	return int64(math.Round(seconds * 1e9))
}

func (c *planCheck) disturbanceIssues(plan *model.SessionPlan) []model.Issue {
	var issues []model.Issue
	for i, d := range plan.Disturbances {
		issues = append(issues, c.disturbance(fmt.Sprintf("/disturbances/%d", i), d)...)
	}
	return issues
}

func (c *planCheck) disturbance(at string, d model.Disturbance) []model.Issue {
	instance, known := c.known[d.Instance]
	switch {
	case !known:
		return []model.Issue{model.RuleIssue(at+"/instance", "unknown_instance", "instance", fmt.Sprintf("the plant has no instance %q", clip(d.Instance)))}
	case c.instances[d.Instance]:
		return []model.Issue{model.RuleIssue(at+"/instance", KeywordDisturbanceOnRealInstance, "instance", fmt.Sprintf(
			"%s is a real board: a disturbance strikes only a simulated instance", clip(d.Instance)))}
	}
	return faultIssues(at, instance, d)
}

func faultIssues(at string, instance PlanInstance, d model.Disturbance) []model.Issue {
	index := slices.IndexFunc(instance.Faults, func(f PlanFault) bool { return f.ID == d.Fault })
	switch {
	case index < 0:
		return []model.Issue{model.RuleIssue(at+"/fault", KeywordUnknownFault, "fault", fmt.Sprintf(
			"the kind %q of %s declares no fault %q", clip(instance.Kind), clip(d.Instance), clip(d.Fault)))}
	case instance.Faults[index].Pulse && d.DurationS != nil:
		return []model.Issue{model.RuleIssue(at+"/duration_s", KeywordDurationWithPulseFault, "duration_s", fmt.Sprintf(
			"%q is a pulse fault: the domain fixes its duration", clip(d.Fault)))}
	case instance.Faults[index].Commanded:
		return []model.Issue{model.RuleIssue(at+"/fault", KeywordFaultOnCommandedState, "fault", fmt.Sprintf(
			"%q acts on a state a command of %s sets, and the next command would cancel it (D-116): not scriptable yet", clip(d.Fault), clip(d.Instance)))}
	}
	return nil
}
