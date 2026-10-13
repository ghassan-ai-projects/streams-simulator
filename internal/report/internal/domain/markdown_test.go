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
		"### instrument\n\n| Metric | Value |\n|---|---|\n| emitted | 5 |\n| ledger\\_complete | true |",
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

func TestMarkdownEscapesEveryMarkdownCharacterOfFreeText(t *testing.T) {
	t.Parallel()
	hostile := "a|b*c_d`e[f]g<h>i&j\\k\nl"
	report := &Report{
		Run:      Run{RunID: hostile, Mode: "evaluate", IncompleteReason: hostile},
		Consumer: &Consumer{Name: hostile, Version: "1", Arm: "[x](http://example.test)"},
		Truth:    &Truth{EffectorCalls: []EffectorCall{{CommandID: "c", Effector: "e", Outcome: &Outcome{Refusal: hostile}}}},
	}
	got := string(Markdown(report))
	want := `a\|b\*c\_d\` + "`" + `e\[f\]g\<h\>i\&j\\k l`
	for _, line := range []string{"# Run report: " + want, "| Incomplete reason | " + want + " |", "| Name | " + want + " |",
		"| Arm | \\[x\\](http://example.test) |", "refused: " + want} {
		if !strings.Contains(got, line) {
			t.Errorf("markdown lacks %q:\n%s", line, got)
		}
	}
}

func TestMarkdownKeepsACodeSpanLiteralAndOutlivesBackticksInIt(t *testing.T) {
	t.Parallel()
	got := string(Markdown(&Report{Truth: &Truth{Faults: []Fault{{Action: "inject", Params: map[string]any{"a_b": "x`y|z"}}}}}))
	if !strings.Contains(got, "``{\"a_b\":\"x`y\\|z\"}``") {
		t.Fatalf("a code span must keep its text literal, fenced past the backtick it holds:\n%s", got)
	}
}
