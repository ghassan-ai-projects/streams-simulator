package domain

import (
	"errors"
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

var ErrTwinRequirements = errors.New("plantwire: the kind file does not meet the requirements of a twin")

type TwinRules func(entry map[string]any) []model.Issue

func CheckTwinRules(kinds KindFiles, summaries []KindSummary, rules TwinRules) error {
	if rules == nil {
		return nil
	}
	for _, summary := range summaries {
		if issues := rules(kinds[summary.Kind]); len(issues) > 0 {
			return twinViolation(summary.Path, issues)
		}
	}
	return nil
}

func twinViolation(path string, issues []model.Issue) error {
	first := issues[0]
	summary := fmt.Sprintf("plantwire: kind file %s fails the requirements of a twin: %s: %s", path, first.Pointer, first.Message)
	return &model.ValidationError{Contract: model.ContractKindFile, Source: path, Summary: summary, Issues: issues, Cause: ErrTwinRequirements}
}
