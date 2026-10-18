package domain

import (
	"errors"
	"reflect"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const (
	validKind = `{"kind": "valve", "metrics": [], "commands": []}`
	otherKind = `{"kind": "pump", "metrics": [], "commands": []}`
)

func parsedKinds(t *testing.T, sources ...KindSource) (KindFiles, []KindSummary) {
	t.Helper()
	files, summaries, err := ParseKindFiles(sources)
	if err != nil {
		t.Fatal(err)
	}
	return files, summaries
}

func TestNoTwinRulesChecksNothing(t *testing.T) {
	t.Parallel()
	files, summaries := parsedKinds(t, KindSource{Path: "valve.kind.json", Raw: []byte(validKind)})
	if err := CheckTwinRules(files, summaries, nil); err != nil {
		t.Fatalf("CheckTwinRules without rules = %v, want none", err)
	}
}

func TestTwinRulesRefuseTheFirstKindFileThatBreaksThemWithEveryIssueOfIt(t *testing.T) {
	t.Parallel()
	files, summaries := parsedKinds(t, KindSource{Path: "valve.kind.json", Raw: []byte(validKind)}, KindSource{Path: "pump.kind.json", Raw: []byte(otherKind)})
	issues := []model.Issue{
		model.RuleIssue("/metrics", "twin_member_required", "metrics", "kind has no metric"),
		model.RuleIssue("/kind", "twin_name_pattern", "kind", "bad name"),
	}
	var asked []string
	rules := func(entry map[string]any) []model.Issue {
		asked = append(asked, entry["kind"].(string))
		return issues
	}
	err := CheckTwinRules(files, summaries, rules)
	invalid, ok := model.AsValidationError(err)
	if !ok || !errors.Is(err, ErrTwinRequirements) || invalid.Contract != model.ContractKindFile || invalid.Source != "valve.kind.json" || !reflect.DeepEqual(invalid.Issues, issues) {
		t.Fatalf("CheckTwinRules = %v (%+v), want a kind-file violation of valve.kind.json carrying both issues", err, invalid)
	}
	if !reflect.DeepEqual(asked, []string{"valve"}) {
		t.Fatalf("rules asked about %v, want only the first kind file: the check stops at the first file that fails", asked)
	}
	const summary = "plantwire: kind file valve.kind.json fails the requirements of a twin: /metrics: kind has no metric"
	if invalid.Summary != summary {
		t.Fatalf("summary = %q, want %q", invalid.Summary, summary)
	}
}

func TestTwinRulesThatFindNothingAcceptEveryKindFile(t *testing.T) {
	t.Parallel()
	files, summaries := parsedKinds(t, KindSource{Path: "valve.kind.json", Raw: []byte(validKind)}, KindSource{Path: "pump.kind.json", Raw: []byte(otherKind)})
	var asked int
	rules := func(map[string]any) []model.Issue {
		asked++
		return nil
	}
	if err := CheckTwinRules(files, summaries, rules); err != nil || asked != 2 {
		t.Fatalf("CheckTwinRules = %v after asking %d times, want it to accept both kind files", err, asked)
	}
}
