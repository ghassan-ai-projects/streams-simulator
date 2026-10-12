package domain

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

var scorecardSections = []string{"instrument", "consumer", "loop", "judgment"}

func consumerLines(consumer *Consumer) []string {
	if consumer == nil {
		return nil
	}
	return fieldTable([]field{
		{"Name", consumer.Name},
		{"Version", consumer.Version},
		{"Arm", consumer.Arm},
		{"Config digest", code(consumer.ConfigDigest)},
	})
}

func scoringLines(scoring *Scoring) []string {
	if scoring == nil {
		return nil
	}
	lines := fieldTable([]field{{"Status", scoring.Status}, {"Reason", scoring.Reason}})
	if len(scoring.Scorecard) == 0 {
		return lines
	}
	var card map[string]any
	if err := model.DecodeBytes(scoring.Scorecard, &card); err != nil {
		return lines
	}
	return append(lines, scorecardLines(card)...)
}

func scorecardLines(card map[string]any) []string {
	var lines []string
	for _, name := range scorecardSections {
		metrics, _ := card[name].(map[string]any)
		if len(metrics) > 0 {
			lines = append(append(lines, "", "### "+name, ""), metricLines(metrics)...)
		}
	}
	return lines
}

func metricLines(metrics map[string]any) []string {
	keys := sortedKeys(metrics)
	rows := make([][]string, len(keys))
	for i, key := range keys {
		rows[i] = []string{key, fmt.Sprint(metrics[key])}
	}
	return table([]string{"Metric", "Value"}, rows)
}
