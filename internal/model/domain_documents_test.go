package model

import (
	"strings"
	"testing"
)

const digestSample = "sha256:0000000000000000000000000000000000000000000000000000000000000000"

func TestValidateDomainAssumptionsAcceptsTheMinimalDocumentAndNamesTheContractItBreaks(t *testing.T) {
	t.Parallel()
	valid := `{"schema_version":"0.1","domain":{"id":"p","version":"0.1.0","digest":"` + digestSample + `"},"origin":{"kind":"domain"},"kinds":[],"not_modeled":[]}`
	if err := ValidateDomainAssumptions([]byte(valid)); err != nil {
		t.Fatalf("ValidateDomainAssumptions = %v, want the minimal document accepted", err)
	}
	err := ValidateDomainAssumptions([]byte(strings.Replace(valid, `"domain"}`, `"guess"}`, 1)))
	invalid, ok := AsValidationError(err)
	if !ok || invalid.Contract != ContractDomainAssumptions || len(invalid.Issues) == 0 {
		t.Fatalf("ValidateDomainAssumptions = %v, want a violation of %s", err, ContractDomainAssumptions)
	}
	if code, known := invalid.Contract.FailureCode(); !known || code != "domain_failed" {
		t.Fatalf("failure code = %q, %v, want domain_failed", code, known)
	}
	err = ValidateDomainAssumptions([]byte(`{"schema_version":`))
	if invalid, ok := AsValidationError(err); !ok || invalid.Contract != ContractDomainAssumptions || !strings.Contains(invalid.Summary, "domain assumptions not valid JSON") {
		t.Fatalf("ValidateDomainAssumptions = %v, want a violation for text that is not JSON", err)
	}
}

func TestValidateDomainLineageAcceptsTheMinimalDocumentAndNamesTheContractItBreaks(t *testing.T) {
	t.Parallel()
	valid := `{"schema_version":"0.1","plants":[],"device_kinds":[],"unreadable":[]}`
	if err := ValidateDomainLineage([]byte(valid)); err != nil {
		t.Fatalf("ValidateDomainLineage = %v, want the minimal document accepted", err)
	}
	err := ValidateDomainLineage([]byte(strings.Replace(valid, `"plants":[]`, `"plants":[{}]`, 1)))
	invalid, ok := AsValidationError(err)
	if !ok || invalid.Contract != ContractDomainLineage || len(invalid.Issues) == 0 {
		t.Fatalf("ValidateDomainLineage = %v, want a violation of %s", err, ContractDomainLineage)
	}
	if code, known := invalid.Contract.FailureCode(); !known || code != "domain_failed" {
		t.Fatalf("failure code = %q, %v, want domain_failed", code, known)
	}
	err = ValidateDomainLineage([]byte(`{`))
	if invalid, ok := AsValidationError(err); !ok || invalid.Contract != ContractDomainLineage || !strings.Contains(invalid.Summary, "domain lineage not valid JSON") {
		t.Fatalf("ValidateDomainLineage = %v, want a violation for text that is not JSON", err)
	}
}

func TestValidateDomainDiffAcceptsTheMinimalDocumentAndNamesTheContractItBreaks(t *testing.T) {
	t.Parallel()
	side := `{"id":"p","version":"0.1.0","title":"P","digest":"` + digestSample + `","source":{"kind":"file","label":"p.domain.json"}}`
	valid := `{"schema_version":"0.1","a":` + side + `,"b":` + side + `,"identity":{"same_id":true,"same_version":true,"same_digest":true,` +
		`"same_id_and_version_different_digest":false},"kinds":[],"domain_changes":[],"summary":{"changes":0,"kinds_changed":0,"kinds_added":0,"kinds_removed":0},` +
		`"unchanged_summary":{"kinds":0,"readings":0,"dynamics":0,"commands":0,"faults":0,"states":0,"instance_sets":0,"links":0,"link_types":0}}`
	if err := ValidateDomainDiff([]byte(valid)); err != nil {
		t.Fatalf("ValidateDomainDiff = %v, want the minimal document accepted", err)
	}
	err := ValidateDomainDiff([]byte(strings.Replace(valid, `"kind":"file"`, `"kind":"folder"`, 1)))
	invalid, ok := AsValidationError(err)
	if !ok || invalid.Contract != ContractDomainDiff || len(invalid.Issues) == 0 {
		t.Fatalf("ValidateDomainDiff = %v, want a violation of %s", err, ContractDomainDiff)
	}
	if code, known := invalid.Contract.FailureCode(); !known || code != "domain_failed" {
		t.Fatalf("failure code = %q, %v, want domain_failed", code, known)
	}
	err = ValidateDomainDiff([]byte(`{`))
	if invalid, ok := AsValidationError(err); !ok || invalid.Contract != ContractDomainDiff || !strings.Contains(invalid.Summary, "domain diff not valid JSON") {
		t.Fatalf("ValidateDomainDiff = %v, want a violation for text that is not JSON", err)
	}
}
