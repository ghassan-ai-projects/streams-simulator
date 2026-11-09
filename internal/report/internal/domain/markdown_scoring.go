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
	return append(lines, gradedLines(card)...)
}

func gradedLines(document map[string]any) []string {
	cards, several := document["scorecards"].([]any)
	if !several {
		return scorecardLines(document, "### ")
	}
	var lines []string
	for i, item := range cards {
		card, _ := item.(map[string]any)
		lines = append(append(lines, "", fmt.Sprintf("### Scorecard %d", i+1), ""), labelLines(card)...)
		lines = append(lines, scorecardLines(card, "#### ")...)
	}
	return append(lines, unexpectedLines(document)...)
}

func labelLines(card map[string]any) []string {
	label, _ := card["ground_truth"].(map[string]any)
	entity, _ := label["entity_id"].(string)
	fault, _ := label["label"].(string)
	return fieldTable([]field{{"Entity", code(entity)}, {"Label", code(fault)}})
}

func unexpectedLines(document map[string]any) []string {
	unexpected, _ := document["unexpected"].(map[string]any)
	detections, _ := unexpected["detections"].([]any)
	actions, _ := unexpected["actions"].([]any)
	counts := map[string]any{"detections": len(detections), "actions": len(actions), "effect_calls": unexpected["effect_calls"]}
	return append([]string{"", "### unexpected", ""}, metricLines(counts)...)
}

func scorecardLines(card map[string]any, heading string) []string {
	var lines []string
	for _, name := range scorecardSections {
		metrics, _ := card[name].(map[string]any)
		if len(metrics) > 0 {
			lines = append(append(lines, "", heading+name, ""), metricLines(metrics)...)
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
