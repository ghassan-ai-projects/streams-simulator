package domain

import (
	"errors"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestABirthIsAClaimAndOneThatDoesNotHoldIsRefusedWithEveryIssue(t *testing.T) {
	t.Parallel()
	missing := strings.Repeat("0", 64)
	cases := []struct {
		name  string
		edit  func(birth map[string]any)
		issue model.Issue
	}{
		{"a kind whose content changed after it was digested", func(b map[string]any) {
			firstKind(b)["metrics"].([]any)[0].(map[string]any)["unit"] = "gal/h"
		}, model.Issue{Pointer: "/kinds/0/kind_digest", Keyword: "kind_digest_mismatch", Member: "kind_digest"}},
		{"a kind announced twice", func(b map[string]any) {
			b["kinds"] = append(b["kinds"].([]any), firstKind(b))
		}, model.Issue{Pointer: "/kinds/2/kind", Keyword: "kind_duplicate", Member: "kind"}},
		{"a protocol range that is empty", func(b map[string]any) { b["protocol_min"] = 5 },
			model.Issue{Pointer: "/protocol_min", Keyword: "protocol_invalid", Member: "protocol_min"}},
		{"another contract version", func(b map[string]any) { b["contract_version"] = "9.9.9" },
			model.Issue{Pointer: "/contract_version", Keyword: "contract_unsupported", Member: "contract_version"}},
		{"another contract digest", func(b map[string]any) { b["contract_digest"] = "sha256:" + missing },
			model.Issue{Pointer: "/contract_digest", Keyword: "contract_unsupported", Member: "contract_digest"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := parse(birthLine(t, tc.edit))
			invalid := violation(t, err, ErrBirthRefused)
			if len(invalid.Issues) != 1 {
				t.Fatalf("issues = %+v, want exactly one", invalid.Issues)
			}
			got := invalid.Issues[0]
			if got.Pointer != tc.issue.Pointer || got.Keyword != tc.issue.Keyword || got.Member != tc.issue.Member || !got.Located || got.Message == "" {
				t.Fatalf("issue = %+v, want %+v located with a message", got, tc.issue)
			}
			if !strings.Contains(invalid.Summary, "the birth is refused") || strings.Contains(invalid.Summary, "\n") {
				t.Fatalf("summary %q, want one line saying the birth is refused", invalid.Summary)
			}
		})
	}
}

func TestEveryIssueOfARefusedBirthIsReportedTogether(t *testing.T) {
	t.Parallel()
	raw := birthLine(t, func(b map[string]any) {
		b["protocol_min"] = 5
		b["contract_version"] = "9.9.9"
		firstKind(b)["kind"] = "renamed"
		b["kinds"].([]any)[1].(map[string]any)["kind"] = "renamed"
	})
	_, err := parse(raw)
	invalid := violation(t, err, ErrBirthRefused)
	want := "protocol_invalid contract_unsupported kind_digest_mismatch kind_digest_mismatch kind_duplicate"
	if got := strings.Join(keywords(invalid.Issues), " "); got != want {
		t.Fatalf("keywords = %q, want %q", got, want)
	}
}

func TestARecordThatBreaksTheNBIRTHSchemaIsRefusedAtTheMemberItNames(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, pointer, keyword string
		edit                   func(birth map[string]any)
	}{
		{"an environment of neither kind", "/environment", "enum", func(b map[string]any) { b["environment"] = "lab" }},
		{"a board without its environment", "/boards/0/environment", "required", func(b map[string]any) { delete(firstBoard(b), "environment") }},
		{"a board of mixed environment", "/boards/0/environment", "enum", func(b map[string]any) { firstBoard(b)["environment"] = "mixed" }},
		{"a kind without its digest", "/kinds/0/kind_digest", "required", func(b map[string]any) { delete(firstKind(b), "kind_digest") }},
		{"a negative boot sequence", "/bd_seq", "minimum", func(b map[string]any) { b["bd_seq"] = -1 }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := parse(birthLine(t, tc.edit))
			invalid := birthViolationOf(t, err)
			var failure jsonschema.Error
			if !errors.As(err, &failure) || failure.Keyword != tc.keyword {
				t.Fatalf("error = %v, want it caused by a schema error of %s", err, tc.keyword)
			}
			if invalid.Issues[0].Pointer != tc.pointer || invalid.Issues[0].Keyword != tc.keyword || !strings.Contains(invalid.Summary, "NBIRTH fails its schema") {
				t.Fatalf("issues = %+v (%s), want the first at %s for %s", invalid.Issues, invalid.Summary, tc.pointer, tc.keyword)
			}
		})
	}
}
