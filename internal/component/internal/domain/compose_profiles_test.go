package domain

import "testing"

func TestComposeRefusesAProfileDeclaredNotApplicableWithoutARealReason(t *testing.T) {
	t.Parallel()
	raw := mutate(t, withEquipment(t, kilnEntry(nil)), func(doc map[string]any) {
		member(t, doc, "profiles", 1).(map[string]any)["not_applicable"] = "too short"
	})
	_, err := compose(t, baseLibrary(t), raw)
	wantIssue(t, err, "composition-v0.1", "/profiles/1/not_applicable", "minLength")
}
