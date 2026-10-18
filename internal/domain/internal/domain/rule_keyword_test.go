package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestARuleOfTheDomainIsReportedAsALocatedIssueWithTheKeywordDomainRule(t *testing.T) {
	t.Parallel()
	invalid := invalidDomain("pond.domain.json", "/kinds/0/channels/1", "channel observes a state that is not declared")
	want := model.RuleIssue("/kinds/0/channels/1", "domain_rule", "", "channel observes a state that is not declared")
	if len(invalid.Issues) != 1 || invalid.Issues[0] != want || invalid.Source != "pond.domain.json" {
		t.Fatalf("issues = %+v from %s, want %+v", invalid.Issues, invalid.Source, want)
	}
}
