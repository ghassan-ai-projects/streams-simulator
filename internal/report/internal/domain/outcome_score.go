package domain

import (
	"encoding/json"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type OutcomeScoring struct {
	Status string               `json:"status"`
	Reason string               `json:"reason,omitempty"`
	Unit   string               `json:"unit,omitempty"`
	Labels map[string]string    `json:"labels,omitempty"`
	Result *model.OutcomeReport `json:"result,omitempty"`
}

func ScoredOutcome(result *model.OutcomeReport, artifact *model.RunArtifact) *OutcomeScoring {
	unit, labels := outcomeWords(artifact)
	return &OutcomeScoring{Status: Scored, Unit: unit, Labels: labels, Result: result}
}

func DeclaresOutcome(artifact *model.RunArtifact) bool {
	var spec struct {
		GroundTruth struct {
			OutcomeWeights json.RawMessage `json:"outcome_weights"`
		} `json:"ground_truth"`
	}
	return artifact != nil && json.Unmarshal(artifact.DomainSpec, &spec) == nil && len(spec.GroundTruth.OutcomeWeights) > 0
}

func outcomeRunRows(scoring *OutcomeScoring) [][]string {
	rows := make([][]string, len(scoring.Result.Runs))
	for i, run := range scoring.Result.Runs {
		trips, violations := outcomeCounts(run.Facts)
		rows[i] = []string{code(string(run.Mode)), scoring.policyText(run.Mode), costText(run.Cost), fmt.Sprint(trips), fmt.Sprint(violations),
			fmt.Sprint(len(run.Facts.Calls))}
	}
	return rows
}

func outcomeCounts(facts *model.OutcomeFacts) (trips, violations int) {
	for _, entity := range facts.Entities {
		for _, trip := range entity.Trips {
			trips += trip.Count
		}
		for _, boundary := range entity.Boundaries {
			violations += boundary.Violations
		}
	}
	return trips, violations
}

func costText(cost *model.OutcomeCost) string {
	if cost == nil {
		return "-"
	}
	return fmt.Sprintf("%.4f", cost.J)
}

func verdictRows(scoring *OutcomeScoring) [][]string {
	rows := make([][]string, len(scoring.Result.Verdicts))
	for i, verdict := range scoring.Result.Verdicts {
		rows[i] = []string{code(string(verdict.Baseline)), scoring.policyText(verdict.Baseline), costDelta(verdict.Delta), verdict.Verdict}
	}
	return rows
}

func costDelta(delta float64) string {
	return fmt.Sprintf("%.4f", delta)
}
