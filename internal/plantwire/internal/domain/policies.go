package domain

import (
	"fmt"
	"regexp"
	"slices"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const MaxPolicies = 8

var policyDigest = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

type AllowList struct {
	digests []string
}

func NewAllowList(digests []string, pointer string) (AllowList, []model.Issue) {
	if len(digests) > MaxPolicies {
		return AllowList{}, []model.Issue{model.RuleIssue(pointer, "too_many_policies", "policies",
			fmt.Sprintf("at most %d policy digests may be listed", MaxPolicies))}
	}
	if issues := listIssues(digests, pointer); len(issues) > 0 {
		return AllowList{}, issues
	}
	return AllowList{digests: slices.Sorted(slices.Values(digests))}, nil
}

func listIssues(digests []string, pointer string) []model.Issue {
	var issues []model.Issue
	seen := map[string]bool{}
	for i, digest := range digests {
		if issue, broken := policyIssue(fmt.Sprintf("%s/%d", pointer, i), digest, seen); broken {
			issues = append(issues, issue)
		}
		seen[digest] = true
	}
	return issues
}

func policyIssue(at, digest string, seen map[string]bool) (model.Issue, bool) {
	switch {
	case !policyDigest.MatchString(digest):
		return model.RuleIssue(at, "policy_malformed", "policies", "a policy digest is sha256: and 64 lower-case hexadecimal digits"), true
	case digest == stopPolicy:
		return model.RuleIssue(at, "policy_reserved", "policies", "this digest is reserved for the node's own safe stops"), true
	case seen[digest]:
		return model.RuleIssue(at, "policy_listed_twice", "policies", "this policy digest is already listed"), true
	}
	return model.Issue{}, false
}

func (l AllowList) Allows(digest string) bool {
	return slices.Contains(l.digests, digest)
}

func (l AllowList) Digests() []string {
	return append([]string{}, l.digests...)
}
