package domain

import (
	"encoding/json"
	"maps"
	"slices"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func mappedMetricIssues(at, name string, kind MappedKind) []model.Issue {
	metric := kind.Metrics[name]
	var issues []model.Issue
	for _, rawWord := range sortedKeys(metric.Table) {
		issues = append(issues, unknownWordIssues(at+"/table/"+pointerEscapes.Replace(rawWord), kind.RawEnums[metric.Raw], rawWord)...)
	}
	return append(issues, scaleIssues(at, name, metric, kind.Commands)...)
}

func scaleIssues(at, name string, metric MappedMetric, commands map[string]MappedCommand) []model.Issue {
	for _, command := range sortedKeys(commands) {
		arg, writes := commands[command].Args[metric.Raw]
		if writes && arg.From == name && scaleOf(arg.Scale) != scaleOf(metric.Scale) {
			message := "the command " + command + " writes " + metric.Raw + " from " + name + " with scale " + scaleText(arg.Scale) +
				", the metric reads it with scale " + scaleText(metric.Scale)
			return []model.Issue{model.RuleIssue(at+"/scale", KeywordBoardScaleMismatch, "scale", message)}
		}
	}
	return nil
}

func scaleOf(scale json.Number) float64 {
	value, err := scale.Float64()
	if scale == "" || err != nil {
		return 1
	}
	return value
}

func scaleText(scale json.Number) string {
	if scale == "" {
		return "1"
	}
	return scale.String()
}

func rawWordIssues(at string, words []string, value any) []model.Issue {
	word, isWord := value.(string)
	if words == nil || !isWord {
		return nil
	}
	return unknownWordIssues(at, words, word)
}

func unknownWordIssues(at string, words []string, word string) []model.Issue {
	if slices.Contains(words, word) {
		return nil
	}
	message := "the raw word " + word + " is not among the raw_enums words [" + strings.Join(words, ", ") + "]"
	return []model.Issue{model.RuleIssue(at, KeywordBoardRawWordUnknown, word, message)}
}

func nameIssues(at, name string) []model.Issue {
	if wireID.MatchString(name) {
		return nil
	}
	return []model.Issue{model.RuleIssue(at, KeywordBoardNameInvalid, name, "a member name is an id: letters, digits and _.:- only, at most 128")}
}

func sortedKeys[V any](members map[string]V) []string {
	return slices.Sorted(maps.Keys(members))
}
