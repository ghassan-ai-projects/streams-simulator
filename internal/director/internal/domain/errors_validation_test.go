package domain

import (
	"errors"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestValidationToolErrorAddsThePointersTheSummaryLeavesOut(t *testing.T) {
	t.Parallel()
	invalid := &model.ValidationError{Contract: model.ContractDomainSpec, Summary: "domain: x: broken",
		Issues: []model.Issue{model.LocatedIssue("/kinds/0", "broken"), model.UnlocatedIssue("whole")}}
	got := ValidationToolError(invalid, CodeComponentInvalid)
	if got.Code != CodeDomainInvalid || got.Msg != "domain: x: broken\n  /kinds/0: broken" {
		t.Fatalf("tool error = %q %q, want domain_invalid with the located pointer appended", got.Code, got.Msg)
	}
	if plain := ValidationToolError(errors.New("plain"), CodeComponentInvalid); plain.Code != CodeComponentInvalid {
		t.Fatalf("a plain error = %q, want the fallback", plain.Code)
	}
}
