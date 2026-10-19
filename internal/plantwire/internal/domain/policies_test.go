package domain

import (
	"reflect"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func digestOf(char string) string { return "sha256:" + strings.Repeat(char, 64) }

func TestAnAllowListIsSortedAndAllowsOnlyWhatItNames(t *testing.T) {
	t.Parallel()
	list, issues := NewAllowList([]string{digestOf("c"), digestOf("a"), digestOf("b")}, "/policies")
	want := []string{digestOf("a"), digestOf("b"), digestOf("c")}
	if len(issues) != 0 || !reflect.DeepEqual(list.Digests(), want) {
		t.Fatalf("list = %v, issues = %+v, want %v sorted", list.Digests(), issues, want)
	}
	if !list.Allows(digestOf("b")) || list.Allows(digestOf("d")) || list.Allows("") {
		t.Fatal("the list must allow a listed digest and nothing else")
	}
}

func TestAnEmptyAllowListAllowsNothingAndEchoesAnEmptyNonNilList(t *testing.T) {
	t.Parallel()
	for name, digests := range map[string][]string{"nil": nil, "empty": {}} {
		list, issues := NewAllowList(digests, "/policies")
		if len(issues) != 0 || list.Digests() == nil || len(list.Digests()) != 0 || list.Allows(digestOf("b")) || list.Allows("") {
			t.Fatalf("%s: list = %#v, issues = %+v, want an empty list that allows no digest", name, list.Digests(), issues)
		}
	}
}

func TestAnAllowListReportsEachBrokenEntryAtItsPointerWithItsRule(t *testing.T) {
	t.Parallel()
	nine := make([]string, 9)
	for i := range nine {
		nine[i] = digestOf(string("abcdef012"[i]))
	}
	cases := []struct {
		name             string
		digests          []string
		pointer, keyword string
	}{
		{"an upper-case digest", []string{digestOf("a"), strings.ToUpper(digestOf("b"))}, "/policies/1", "policy_malformed"},
		{"a short digest", []string{"sha256:00"}, "/policies/0", "policy_malformed"},
		{"an empty digest", []string{""}, "/policies/0", "policy_malformed"},
		{"a repeat at its second occurrence", []string{digestOf("a"), digestOf("b"), digestOf("a")}, "/policies/2", "policy_listed_twice"},
		{"the reserved stop digest", []string{digestOf("a"), stopPolicy}, "/policies/1", "policy_reserved"},
		{"nine digests", nine, "/policies", "too_many_policies"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			list, issues := NewAllowList(tc.digests, "/policies")
			if len(issues) != 1 || issues[0].Pointer != tc.pointer || issues[0].Keyword != tc.keyword || issues[0].Member != "policies" || issues[0].Message == "" {
				t.Fatalf("issues = %+v, want one %s at %s", issues, tc.keyword, tc.pointer)
			}
			if len(list.Digests()) != 0 || list.Allows(digestOf("a")) {
				t.Fatalf("a list with an issue allows %v, want it to trust nothing", list.Digests())
			}
		})
	}
}

func TestAnAllowListOfEightDistinctDigestsHolds(t *testing.T) {
	t.Parallel()
	eight := make([]string, MaxPolicies)
	for i := range eight {
		eight[i] = digestOf(string("fedcba98"[i]))
	}
	list, issues := NewAllowList(eight, "/policy")
	if len(issues) != 0 || len(list.Digests()) != MaxPolicies || list.Digests()[0] != digestOf("8") {
		t.Fatalf("list = %v, issues = %+v, want eight digests, the smallest first", list.Digests(), issues)
	}
}

func TestThePlansListIsCheckedWithItsOtherRulesAndEchoedSorted(t *testing.T) {
	t.Parallel()
	zone := model.RealBinding{Instance: "ghost", Board: "uno-07", Device: "d"}
	plan := benchPlan(zone)
	plan.Policies = []string{digestOf("b"), digestOf("b")}
	_, issues := CheckPlan(plan, benchInstances())
	if len(issues) != 2 || issues[0].Keyword != "unknown_instance" || issues[1].Pointer != "/policies/1" || issues[1].Keyword != "policy_listed_twice" {
		t.Fatalf("issues = %+v, want the binding issue then the policy issue", issues)
	}
	clean := benchPlan()
	clean.Policies = []string{digestOf("c"), digestOf("a")}
	connect, issues := CheckPlan(clean, benchInstances())
	if len(issues) != 0 || !reflect.DeepEqual(connect.Policies.Digests(), []string{digestOf("a"), digestOf("c")}) || !connect.Policies.Allows(digestOf("c")) {
		t.Fatalf("policies = %v, issues = %+v, want the plan's list sorted", connect.Policies.Digests(), issues)
	}
	none, issues := CheckPlan(benchPlan(), benchInstances())
	if len(issues) != 0 || none.Policies.Digests() == nil || len(none.Policies.Digests()) != 0 {
		t.Fatalf("a plan without policies = %v, %+v, want an empty list", none.Policies.Digests(), issues)
	}
}
