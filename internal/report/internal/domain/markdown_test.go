package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMarkdownRendersAScoredRunWithEscapedCells(t *testing.T) {
	t.Parallel()
	seed := uint64(3)
	report := &Report{
		SchemaVersion: SchemaVersion,
		Run:           Run{RunID: "r-9", Mode: "evaluate", Complete: true, Ended: true, Seed: &seed},
		Verification:  Verification{Status: FilesMatch, Checks: []Check{}},
		Truth: &Truth{
			Faults: []Fault{{Seq: 0, Action: "inject", EntityID: "a|b", Fault: "line\nbreak"}},
		},
		Scoring: &Scoring{Status: Scored, Scorecard: json.RawMessage(`{"instrument":{"emitted":5,"ledger_complete":true},"loop":{"resolved":false}}`)},
	}
	got := string(Markdown(report))
	for _, want := range []string{
		"# Run report: r-9\n\nmode: evaluate | complete: yes | ended: yes | verification: files_match\n",
		"| Seed | 3 |",
		"| 0 |  | inject | a\\|b | line break |",
		"## Perturbations\n\nNone.",
		"### instrument\n\n| Metric | Value |\n|---|---|\n| emitted | 5 |\n| ledger_complete | true |",
		"### loop\n\n| Metric | Value |\n|---|---|\n| resolved | false |",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("markdown lacks %q:\n%s", want, got)
		}
	}
	if !strings.HasSuffix(got, "\n") || strings.HasSuffix(got, "\n\n") {
		t.Errorf("markdown must end in exactly one newline:\n%q", got[len(got)-20:])
	}
}

func TestMarkdownTitlesARunWithoutAnIDAndOmitsAbsentSections(t *testing.T) {
	t.Parallel()
	got := string(Markdown(&Report{Run: Run{Mode: "explore"}, Verification: Verification{Status: NotChecked}}))
	if !strings.HasPrefix(got, "# Run report: unknown\n") || strings.Contains(got, "## Consumer") || strings.Contains(got, "## Scoring") ||
		!strings.Contains(got, withheld) {
		t.Fatalf("markdown = %q, want an unknown title, no consumer or scoring and the withheld line", got)
	}
}

func TestMarkdownShowsRefusedAndFailedScoringReasons(t *testing.T) {
	t.Parallel()
	report := &Report{Truth: &Truth{EffectorCalls: []EffectorCall{{CommandID: "c", Effector: "e", Outcome: &Outcome{Refusal: "interlock"}}}},
		Scoring: &Scoring{Status: Failed, Reason: "replay does not reproduce"}}
	got := string(Markdown(report))
	if !strings.Contains(got, "refused: interlock") || !strings.Contains(got, "| Reason | replay does not reproduce |") {
		t.Fatalf("markdown lacks the refusal or the scoring reason:\n%s", got)
	}
}
