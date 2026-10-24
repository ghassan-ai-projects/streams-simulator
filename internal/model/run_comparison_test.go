package model

import (
	"strings"
	"testing"
)

func TestValidateRunComparisonAcceptsTheMinimalDocumentAndNamesTheContractItBreaks(t *testing.T) {
	t.Parallel()
	valid := `{"schema_version":"0.1","a":` + minimalComparedRun + `,"b":` + minimalComparedRun + `,"plan":{"faults":` + emptyDiff + `,"perturbations":` + emptyDiff +
		`,"environment":` + emptyDiff + `,"effector_calls":` + emptyDiff + `},"comparability_limits":[],"results":{"kpis":[],"alarms":[],` +
		`"alarm_system":{"a":null,"b":null,"scope":{"kind":"whole_run","a_simulated_length_ns":null,"b_simulated_length_ns":null}}},` +
		`"windowed_results":{"status":"not_available","reason":"x"}}`
	if err := ValidateRunComparison([]byte(valid)); err != nil {
		t.Fatalf("ValidateRunComparison = %v, want the minimal document accepted", err)
	}
	err := ValidateRunComparison([]byte(strings.Replace(valid, `"not_available"`, `"computed"`, 1)))
	invalid, ok := AsValidationError(err)
	if !ok || invalid.Contract != ContractRunComparison || len(invalid.Issues) == 0 {
		t.Fatalf("ValidateRunComparison = %v, want a violation of %s", err, ContractRunComparison)
	}
	if code, known := invalid.Contract.FailureCode(); !known || code != "report_failed" {
		t.Fatalf("failure code = %q, %v, want report_failed", code, known)
	}
	err = ValidateRunComparison([]byte(`{"schema_version":`))
	if invalid, ok := AsValidationError(err); !ok || invalid.Contract != ContractRunComparison || !strings.Contains(invalid.Summary, "run comparison not valid JSON") {
		t.Fatalf("ValidateRunComparison = %v, want a violation for text that is not JSON", err)
	}
}

const (
	minimalComparedRun = `{"mode":"explore","start_time_ns":0,"end_time_ns":0,"simulated_length_ns":0,"complete":true,"verification":"files_match","kpi_truth":false}`
	emptyDiff          = `{"identical":[],"different":[],"only_in_a":[],"only_in_b":[]}`
)
