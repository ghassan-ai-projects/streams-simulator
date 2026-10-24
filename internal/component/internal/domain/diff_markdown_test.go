package domain

import (
	"strings"
	"testing"
)

func TestTheMarkdownDiffNamesTheSilentEditTheChangeAndTheInstancesItAffects(t *testing.T) {
	t.Parallel()
	a, b := twinBytes(t, fanZoneChoicesJSON), twinBytes(t, gainChoices("-0.03"))
	text := string(RenderDiffMarkdown(diffOf(t, fileSide(a, "before.domain.json"), fileSide(b, "after.domain.json"))))
	for _, want := range []string{"# Domain diff", "| Source | file before.domain.json | file after.domain.json |", "The model changed and its version did not",
		"## Kind zone (changed)", "Instances: 1 zone-1; now 1 zone-1.", "Relation fan_drives_zone (fan to zone): fan-1>zone-1.",
		"- `gain_changed` zone/dynamics/temp_value/inputs/duty_value@fan_drives_zone: coef changed from -0.02 to -0.03.", "## Unchanged"} {
		if !strings.Contains(text, want) {
			t.Errorf("markdown lacks %q:\n%s", want, text)
		}
	}
}

func TestTheMarkdownDiffOfTwoEqualDomainsListsNoKindAndEscapesTheTable(t *testing.T) {
	t.Parallel()
	raw := twinBytes(t, fanZoneChoicesJSON)
	text := string(RenderDiffMarkdown(diffOf(t, fileSide(raw, "a|b.domain.json"), fileSide(raw, "c.domain.json"))))
	if strings.Contains(text, "## Kind") || strings.Contains(text, "model changed") || !strings.Contains(text, `a\|b.domain.json`) || !strings.Contains(text, "0 change(s)") {
		t.Fatalf("markdown = %s, want no kind section, no silent-edit line and an escaped pipe", text)
	}
}

func TestEveryKindChangeCodeIsOneTheContractNames(t *testing.T) {
	t.Parallel()
	a := twinBytes(t, fanZoneChoicesJSON)
	b := editedDomain(t, a, func(doc map[string]any) {
		addKindCopy(doc, "fan", "fan_b", "fan-2")
		effectorOf(doc, "fan", "set_duty")["safe_stop"] = true
		kindOf(doc, "zone")["flows"] = []any{}
	})
	got := diffOf(t, fileSide(a, "a"), fileSide(b, "b"))
	if got.Summary.KindsAdded != 1 || got.Summary.KindsChanged == 0 || got.Summary.Changes != countChanges(got) {
		t.Fatalf("summary = %+v, want one added kind and the totals of the groups", got.Summary)
	}
}

func countChanges(diff *Diff) int {
	total := len(diff.DomainChanges)
	for _, kind := range diff.Kinds {
		total += len(kind.Changes)
	}
	return total
}
