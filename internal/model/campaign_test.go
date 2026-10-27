package model

import (
	"strings"
	"testing"
)

func TestValidateCampaignAcceptsARunAndRefusesAnythingThatReadsAsTruth(t *testing.T) {
	t.Parallel()
	const run = `{"index":1,"phase":"learn","domain":"pump-station","seed":7,"duration_s":172800,"session_plan":"run-0001.session-plan.json"}`
	valid := `{"schema_version":"0.1","domain":"pump-station","seed":1,"runs":[` + run + `]}`
	if err := ValidateCampaign([]byte(valid)); err != nil {
		t.Fatalf("ValidateCampaign = %v, want the minimal campaign accepted", err)
	}
	for name, broken := range map[string]string{
		"expected action": strings.Replace(valid, `"phase":"learn"`, `"phase":"learn","expected_effector":"schedule_maintenance"`, 1),
		"deadline":        strings.Replace(valid, `"phase":"learn"`, `"phase":"learn","deadline_s":43200`, 1),
		"injected fault":  strings.Replace(valid, `"phase":"learn"`, `"phase":"learn","disturbances":[{"at_s":60,"instance":"site-a/pump-1","fault":"bearing_wear"}]`, 1),
		"no runs":         strings.Replace(valid, `[`+run+`]`, `[]`, 1),
	} {
		invalid, ok := AsValidationError(ValidateCampaign([]byte(broken)))
		if !ok || invalid.Contract != ContractCampaign || len(invalid.Issues) == 0 {
			t.Fatalf("%s: want a violation of %s, got %v", name, ContractCampaign, invalid)
		}
	}
	if code, known := ContractCampaign.FailureCode(); !known || code != "suite_failed" {
		t.Fatalf("failure code = %q, %v, want suite_failed", code, known)
	}
}
