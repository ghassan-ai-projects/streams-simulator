package domain

import (
	"encoding/json"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type OutcomeScoring struct {
	Status string               `json:"status"`
	Reason string               `json:"reason,omitempty"`
	Result *model.OutcomeReport `json:"result,omitempty"`
}

func DeclaresOutcome(artifact *model.RunArtifact) bool {
	var spec struct {
		GroundTruth struct {
			OutcomeWeights json.RawMessage `json:"outcome_weights"`
		} `json:"ground_truth"`
	}
	return artifact != nil && json.Unmarshal(artifact.DomainSpec, &spec) == nil && len(spec.GroundTruth.OutcomeWeights) > 0
}

func outcomeScoreLines(scoring *OutcomeScoring) []string {
	if scoring == nil {
		return nil
	}
	lines := fieldTable([]field{{"Status", scoring.Status}, {"Reason", scoring.Reason}})
	if result := scoring.Result; result != nil {
		lines = append(lines, fieldTable([]field{{"Valid", yesNo(result.Valid)}, {"Easy", yesNo(result.Easy)}})...)
		lines = append(append(lines, ""), table([]string{"Run", "Cost J", "Trips", "Boundary violations", "Calls"}, outcomeRunRows(result.Runs))...)
		lines = append(append(lines, ""), table([]string{"Baseline", "Delta", "Verdict"}, verdictRows(result.Verdicts))...)
	}
	return lines
}

func outcomeRunRows(runs []model.OutcomeRun) [][]string {
	rows := make([][]string, len(runs))
	for i, run := range runs {
		trips, violations := outcomeCounts(run.Facts)
		rows[i] = []string{code(string(run.Mode)), costText(run.Cost), fmt.Sprint(trips), fmt.Sprint(violations), fmt.Sprint(len(run.Facts.Calls))}
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

func verdictRows(verdicts []model.OutcomeVerdict) [][]string {
	rows := make([][]string, len(verdicts))
	for i, verdict := range verdicts {
		rows[i] = []string{code(string(verdict.Baseline)), fmt.Sprintf("%.4f", verdict.Delta), verdict.Verdict}
	}
	return rows
}
