package domain

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const (
	KeywordHorizonWithGateway      = "horizon_with_gateway"
	KeywordDisturbanceAfterHorizon = "disturbance_after_horizon"
	KeywordClearAfterHorizon       = "clear_after_horizon"
)

const openEndedNS int64 = 100 * 365 * 24 * 3600 * 1_000_000_000

func SessionEndNS(plan *model.SessionPlan, startNS int64) int64 {
	if plan.HorizonS == nil {
		return startNS + openEndedNS
	}
	return startNS + nanoseconds(*plan.HorizonS)
}

func horizonIssues(plan *model.SessionPlan) []model.Issue {
	if plan.HorizonS == nil {
		return nil
	}
	if plan.GatewaySocket != "" {
		return []model.Issue{model.RuleIssue("/horizon_s", KeywordHorizonWithGateway, "horizon_s",
			"a session with a gateway runs until it is stopped: horizon_s is only for a session of simulated boards")}
	}
	var issues []model.Issue
	for i, d := range plan.Disturbances {
		issues = append(issues, disturbanceHorizonIssues(fmt.Sprintf("/disturbances/%d", i), d, *plan.HorizonS)...)
	}
	return issues
}

func disturbanceHorizonIssues(at string, d model.Disturbance, horizonS float64) []model.Issue {
	switch {
	case d.AtS >= horizonS:
		return []model.Issue{model.RuleIssue(at+"/at_s", KeywordDisturbanceAfterHorizon, "at_s", fmt.Sprintf(
			"the disturbance starts at %g s, not before the session's horizon of %g s", d.AtS, horizonS))}
	case d.DurationS != nil && d.AtS+*d.DurationS > horizonS:
		return []model.Issue{model.RuleIssue(at+"/duration_s", KeywordClearAfterHorizon, "duration_s", fmt.Sprintf(
			"the disturbance clears at %g s, after the session's horizon of %g s", d.AtS+*d.DurationS, horizonS))}
	}
	return nil
}
