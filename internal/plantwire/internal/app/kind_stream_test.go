package app

import (
	"errors"
	"io/fs"
	"reflect"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/files"
)

func TestValidateKindStreamSummarisesTheDocumentAsStandardInput(t *testing.T) {
	t.Parallel()
	summary, err := ValidateKindStream(strings.NewReader(pumpKindContent(t)), nil)
	if err != nil || summary.Path != "<stdin>" || summary.Kind != "pump" || summary.Metrics != 2 || summary.Commands != 3 || !strings.HasPrefix(summary.KindDigest, "sha256:") {
		t.Fatalf("ValidateKindStream = %+v, %v, want the pump kind file read from <stdin> with its digest", summary, err)
	}
}

func TestValidateKindStreamHoldsADocumentToTheLimitOfAFile(t *testing.T) {
	t.Parallel()
	padding := strings.Repeat(" ", files.MaxKindFileBytes-len(pumpKindContent(t)))
	if _, err := ValidateKindStream(strings.NewReader(pumpKindContent(t)+padding), nil); err != nil {
		t.Fatalf("a document of exactly %d bytes = %v, want it accepted", files.MaxKindFileBytes, err)
	}
	_, err := ValidateKindStream(strings.NewReader(pumpKindContent(t)+padding+" "), nil)
	if err == nil || !strings.Contains(err.Error(), "<stdin> holds more than 1048576 bytes") {
		t.Fatalf("a document one byte over = %v, want a refusal naming the limit", err)
	}
}

func TestValidateKindStreamDecodesStrictlyAndReportsEveryIssueLikeAFile(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, raw string
		located   bool
		keyword   string
	}{
		{"not JSON", `{`, false, "json"},
		{"two documents", pumpKindContent(t) + `{}`, false, "json"},
		{"a member named twice", `{"kind": "k", "kind": "j", "metrics": [], "commands": []}`, false, "json"},
		{"a schema violation", `{"kind": "k", "metrics": []}`, true, "required"},
		{"a kind_digest", `{"kind": "k", "kind_digest": "sha256:00", "metrics": [], "commands": []}`, true, "kind_digest_forbidden"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := ValidateKindStream(strings.NewReader(tc.raw), nil)
			invalid, ok := model.AsValidationError(err)
			if !ok || invalid.Contract != model.ContractKindFile || invalid.Source != "<stdin>" || len(invalid.Issues) == 0 ||
				invalid.Issues[0].Located != tc.located || invalid.Issues[0].Keyword != tc.keyword {
				t.Fatalf("ValidateKindStream = %v (%+v), want a kind-file violation of <stdin> whose first issue has keyword %s", err, invalid, tc.keyword)
			}
		})
	}
	if _, err := ValidateKindStream(strings.NewReader(`{"kind": "k", "kind": "j"}`), nil); !errors.Is(err, domain.ErrDuplicateMember) {
		t.Fatalf("a member named twice = %v, want ErrDuplicateMember", err)
	}
}

func TestValidateKindStreamAddsTheTwinRulesOnlyWhenGiven(t *testing.T) {
	t.Parallel()
	want := []model.Issue{model.RuleIssue("/metrics/0/resolution", "twin_member_required", "resolution", "needs it"),
		model.RuleIssue("/metrics/1/resolution", "twin_member_required", "resolution", "needs it")}
	rules := func(map[string]any) []model.Issue { return want }
	if _, err := ValidateKindStream(strings.NewReader(pumpKindContent(t)), nil); err != nil {
		t.Fatalf("without rules = %v, want the document accepted", err)
	}
	_, err := ValidateKindStream(strings.NewReader(pumpKindContent(t)), rules)
	invalid, ok := model.AsValidationError(err)
	if !ok || !errors.Is(err, domain.ErrTwinRequirements) || invalid.Source != "<stdin>" || !reflect.DeepEqual(invalid.Issues, want) {
		t.Fatalf("with rules = %v (%+v), want both issues of <stdin>", err, invalid)
	}
}

func TestValidateKindFilesWithRulesNamesTheFileThatFailsThem(t *testing.T) {
	t.Parallel()
	dir := kindsDirectoryWith(t, map[string]string{"pump.kind.json": pumpKindContent(t)})
	rules := func(map[string]any) []model.Issue {
		return []model.Issue{model.RuleIssue("/metrics", "twin_member_required", "metrics", "x")}
	}
	_, err := ValidateKindFilesWith([]string{dir + "/pump.kind.json"}, rules)
	if invalid, ok := model.AsValidationError(err); !ok || invalid.Source != dir+"/pump.kind.json" {
		t.Fatalf("ValidateKindFilesWith = %v, want a violation naming the file", err)
	}
}

func TestWithKindFilesHandsTheKindFilesToTheCallerAndKeepsItsResult(t *testing.T) {
	t.Parallel()
	dir := kindsDirectoryWith(t, map[string]string{"pump.kind.json": pumpKindContent(t)})
	count, err := WithKindFiles(nil, []string{dir}, func(kinds map[string]map[string]any) (int, error) { return len(kinds), nil })
	if err != nil || count != 1 {
		t.Fatalf("WithKindFiles = %d, %v, want the one pump kind file", count, err)
	}
	refusal := errors.New("the caller refuses")
	if _, err := WithKindFiles(nil, []string{dir}, func(map[string]map[string]any) (int, error) { return 0, refusal }); !errors.Is(err, refusal) {
		t.Fatalf("WithKindFiles = %v, want the caller's error unchanged", err)
	}
	if _, err := WithKindFiles(nil, []string{dir + "/absent"}, func(map[string]map[string]any) (int, error) { return 1, nil }); !errors.Is(err, fs.ErrNotExist) {
		t.Fatal("a kinds directory that does not exist must be refused before the caller runs")
	}
}
