package model

import (
	"encoding/json"
	"strings"
	"testing"
)

const digestA = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func minimalTwinReport() map[string]any {
	return map[string]any{
		"schema_version": "0.1", "verdict": "agrees",
		"summary": map[string]any{"devices": 1, "metrics": 1, "compared": 2, "within": 2, "diverged": 0, "not_compared": 0,
			"unattributed_readings": 0, "metrics_cannot_judge": 0, "divergences": 0, "commands_applied": 1, "commands_not_applied": 0},
		"run":            map[string]any{"field_trace_digest": digestA, "journal_digest": digestA, "evidence_bound": true},
		"twin":           map[string]any{"domain": "arduino-thermal", "domain_digest": digestA, "kinds": []any{map[string]any{"kind": "fan", "kind_digest": digestA}}, "start_ns": 100},
		"tolerance_rule": map[string]any{"default": "span_fraction", "span_fraction": 0.05, "description": "a fraction of the span"},
		"commands":       []any{map[string]any{"command_id": "c-1", "device": "fan-01", "command": "set_pwm_lease", "status": "applied", "at_ns": 150}},
		"metrics": []any{map[string]any{"device": "zone-01", "kind": "thermal_sensor", "metric": "temperature", "unit": "Cel",
			"tolerance": map[string]any{"value": 2.5, "source": "default_span"}, "verdict": "agrees",
			"summary":              map[string]any{"compared": 2, "within": 2, "diverged": 0, "not_compared": 1},
			"not_compared_reasons": []any{map[string]any{"reason": "before_first_command", "count": 1}},
			"series":               map[string]any{"points": []any{map[string]any{"at_ns": 200, "board": 34.0, "twin": 33.9}}, "total": 2, "truncated": false}}},
		"divergences": []any{},
	}
}

func marshalReport(t *testing.T, report map[string]any) []byte {
	t.Helper()
	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func metricOf(report map[string]any) map[string]any {
	return report["metrics"].([]any)[0].(map[string]any)
}

func TestAMinimalTwinReportIsAccepted(t *testing.T) {
	t.Parallel()
	if err := ValidateTwinReport(marshalReport(t, minimalTwinReport())); err != nil {
		t.Fatalf("ValidateTwinReport = %v, want the minimal report accepted", err)
	}
}

func TestATwinReportWithADivergenceANullToleranceAndAReasonIsAccepted(t *testing.T) {
	t.Parallel()
	report := minimalTwinReport()
	report["verdict"], report["reason"] = "cannot_judge", "no_readings_compared"
	metricOf(report)["tolerance"] = map[string]any{"value": nil, "source": "none"}
	report["divergences"] = []any{map[string]any{"name": "zone-01.temperature#1", "device": "zone-01", "metric": "temperature", "first_ns": 1, "last_ns": 2,
		"readings": 2, "tolerance": 2.5, "worst_error": 6, "worst_at_ns": 2, "board_value": 40, "twin_value": 34, "commands": []any{"c-1"}}}
	report["run"].(map[string]any)["interrupt_log_digest"] = digestA
	if err := ValidateTwinReport(marshalReport(t, report)); err != nil {
		t.Fatalf("ValidateTwinReport = %v, want a null tolerance, a divergence and the optional members accepted", err)
	}
}

func TestATwinReportThatBreaksItsContractIsRefusedAsAViolationOfIt(t *testing.T) {
	t.Parallel()
	tests := map[string]func(map[string]any){
		"a member nothing defines":             func(r map[string]any) { r["verbose"] = true },
		"a verdict outside the enum":           func(r map[string]any) { r["verdict"] = "maybe" },
		"a digest that is no digest":           func(r map[string]any) { r["run"].(map[string]any)["journal_digest"] = "sha256:abc" },
		"a negative declared tolerance":        func(r map[string]any) { metricOf(r)["tolerance"] = map[string]any{"value": -1, "source": "declared"} },
		"a source outside the enum":            func(r map[string]any) { metricOf(r)["tolerance"] = map[string]any{"value": 1, "source": "guessed"} },
		"a reason outside the enum":            func(r map[string]any) { r["reason"] = "tired" },
		"a missing required member":            func(r map[string]any) { delete(r, "divergences") },
		"a schema version of another contract": func(r map[string]any) { r["schema_version"] = "0.2" },
		"a command status outside the enum":    func(r map[string]any) { r["commands"].([]any)[0].(map[string]any)["status"] = "maybe" },
		"a not-compared reason outside the enum": func(r map[string]any) {
			metricOf(r)["not_compared_reasons"] = []any{map[string]any{"reason": "bored", "count": 1}}
		},
		"a series over two thousand points": func(r map[string]any) {
			points := make([]any, 2001)
			for i := range points {
				points[i] = map[string]any{"at_ns": i, "board": 1.0, "twin": 1.0}
			}
			metricOf(r)["series"] = map[string]any{"points": points, "total": 2001, "truncated": false}
		},
	}
	for name, change := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			report := minimalTwinReport()
			change(report)
			err := ValidateTwinReport(marshalReport(t, report))
			invalid, ok := AsValidationError(err)
			if !ok || invalid.Contract != ContractTwinReport || len(invalid.Issues) == 0 {
				t.Fatalf("ValidateTwinReport = %v, want a violation of %s", err, ContractTwinReport)
			}
			if code, known := invalid.Contract.FailureCode(); !known || code != "plantwire_failed" {
				t.Fatalf("failure code = %q, %v, want plantwire_failed", code, known)
			}
		})
	}
}

func TestATwinReportThatIsNotJSONIsRefusedAsAViolationOfTheContract(t *testing.T) {
	t.Parallel()
	err := ValidateTwinReport([]byte(`{"schema_version":`))
	invalid, ok := AsValidationError(err)
	if !ok || invalid.Contract != ContractTwinReport || !strings.Contains(invalid.Summary, "twin report not valid JSON") {
		t.Fatalf("ValidateTwinReport = %v, want a violation of %s for text that is not JSON", err, ContractTwinReport)
	}
}
