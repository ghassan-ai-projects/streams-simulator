package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const (
	kindDigestMember       = "kind_digest"
	KeywordDigestForbidden = "kind_digest_forbidden"
	KeywordRangeEmpty      = "kind_range_empty"
)

var (
	ErrKindDigestInFile = errors.New("plantwire: a kind file carries no kind_digest")
	ErrEmptyRange       = errors.New("plantwire: a range needs min below max")
)

var rangeMembers = []string{"eu_range", "instrument_range"}

func digestInKindFile() error {
	message := "a kind file must not carry kind_digest: it is computed from the file, so remove the member"
	return &model.ValidationError{Contract: model.ContractKindFile, Summary: "plantwire: kind file fails its schema: /kind_digest: " + message,
		Issues: []model.Issue{model.RuleIssue("/"+kindDigestMember, KeywordDigestForbidden, kindDigestMember, message)}, Cause: ErrKindDigestInFile}
}

func checkKindContent(entry map[string]any) error {
	issues := rangeIssues(entry)
	if len(issues) == 0 {
		return checkCanonical(entry)
	}
	issues = sortedByPointer(issues)
	return &model.ValidationError{Contract: model.ContractKindFile, Summary: "plantwire: kind file fails its schema: " + issues[0].Pointer + ": " + issues[0].Message,
		Issues: issues, Cause: ErrEmptyRange}
}

func rangeIssues(entry map[string]any) []model.Issue {
	var issues []model.Issue
	metrics, _ := entry["metrics"].([]any)
	for i, metric := range metrics {
		fields, _ := metric.(map[string]any)
		issues = append(issues, metricRangeIssues(fmt.Sprintf("/metrics/%d", i), fields)...)
	}
	return issues
}

func metricRangeIssues(pointer string, metric map[string]any) []model.Issue {
	var issues []model.Issue
	for _, member := range rangeMembers {
		if issue, empty := emptyRange(metric[member], pointer+"/"+member); empty {
			issues = append(issues, issue)
		}
	}
	return issues
}

func emptyRange(value any, pointer string) (model.Issue, bool) {
	bounds, _ := value.(map[string]any)
	low, lowKnown := numberOf(bounds["min"])
	high, highKnown := numberOf(bounds["max"])
	if !lowKnown || !highKnown || low < high {
		return model.Issue{}, false
	}
	member := pointer[strings.LastIndex(pointer, "/")+1:]
	return model.RuleIssue(pointer, KeywordRangeEmpty, member, fmt.Sprintf("min %v is not below max %v", low, high)), true
}

func numberOf(value any) (float64, bool) {
	switch number := value.(type) {
	case json.Number:
		parsed, err := number.Float64()
		return finite(parsed, err == nil)
	case float64:
		return finite(number, true)
	}
	return 0, false
}

func finite(number float64, parsed bool) (float64, bool) {
	if !parsed || math.IsInf(number, 0) || math.IsNaN(number) {
		return 0, false
	}
	return number, true
}
