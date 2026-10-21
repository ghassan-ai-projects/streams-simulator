package domain

import (
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const (
	KeywordBoardReadbackMismatch = "board_readback_mismatch"
	KeywordBoardTableIncomplete  = "board_table_incomplete"
	KeywordBoardRawShared        = "board_raw_shared"
)

func readBackIssues(at string, kind MappedKind) []model.Issue {
	var issues []model.Issue
	readers := map[string]string{}
	for _, name := range sortedKeys(kind.Metrics) {
		pointer := at + "/metrics/" + pointerEscapes.Replace(name)
		metric := kind.Metrics[name]
		issues = append(issues, sharedRawIssues(pointer, name, metric.Raw, readers)...)
		issues = append(issues, incompleteTableIssues(pointer, metric, kind.RawEnums[metric.Raw])...)
	}
	for _, name := range sortedKeys(kind.Commands) {
		issues = append(issues, roundTripIssues(at+"/commands/"+pointerEscapes.Replace(name), kind.Commands[name], kind.Metrics)...)
	}
	return issues
}

func sharedRawIssues(at, name, raw string, readers map[string]string) []model.Issue {
	first, shared := readers[raw]
	if !shared {
		readers[raw] = name
		return nil
	}
	message := "the metrics " + first + " and " + name + " both read the raw value " + raw + ": one metric per raw reading"
	return []model.Issue{model.RuleIssue(at+"/raw", KeywordBoardRawShared, raw, message)}
}

func incompleteTableIssues(at string, metric MappedMetric, words []string) []model.Issue {
	missing := missingRows(metric.Table, words)
	if len(metric.Table) == 0 || len(missing) == 0 {
		return nil
	}
	message := "every word of the raw value " + metric.Raw + " reads back as a kind word; no row for " + strings.Join(missing, ", ")
	return []model.Issue{model.RuleIssue(at+"/table", KeywordBoardTableIncomplete, missing[0], message)}
}

func missingRows(table map[string]string, words []string) []string {
	var missing []string
	for _, word := range words {
		if _, held := table[word]; !held {
			missing = append(missing, word)
		}
	}
	return missing
}

func roundTripIssues(at string, command MappedCommand, metrics map[string]MappedMetric) []model.Issue {
	metric, reads := metrics[command.By]
	if command.By == "" || !reads || len(metric.Table) == 0 {
		return nil
	}
	var issues []model.Issue
	for _, word := range sortedKeys(command.Table) {
		written, writes := command.Table[word][metric.Raw].(string)
		if read := metric.Table[written]; writes && read != word {
			pointer := at + "/table/" + pointerEscapes.Replace(word) + "/" + pointerEscapes.Replace(metric.Raw)
			issues = append(issues, readbackMismatch(pointer, word, written, read, metric.Raw))
		}
	}
	return issues
}

func readbackMismatch(at, word, written, read, raw string) model.Issue {
	message := "the row " + word + " writes " + raw + " " + written + ", which the metric table reads back as " + quoted(read)
	return model.RuleIssue(at, KeywordBoardReadbackMismatch, word, message)
}

func quoted(word string) string {
	if word == "" {
		return "nothing"
	}
	return word
}
