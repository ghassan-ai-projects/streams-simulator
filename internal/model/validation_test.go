package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
)

func TestIssueMarshalsItsPointerOnlyWhenLocated(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		issue Issue
		want  string
	}{
		{"located", LocatedIssue("/channels/0", "bad"), `{"pointer":"/channels/0","message":"bad"}`},
		{"whole document", LocatedIssue("", "whole document"), `{"pointer":"","message":"whole document"}`},
		{"unlocated", UnlocatedIssue("cross-reference"), `{"message":"cross-reference"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := json.Marshal(tc.issue)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tc.want {
				t.Fatalf("Marshal(%+v) = %s, want %s", tc.issue, got, tc.want)
			}
		})
	}
}

func TestSchemaIssuesKeepEveryViolationAndItsPath(t *testing.T) {
	t.Parallel()
	got := SchemaIssues([]jsonschema.Error{{Path: "", Msg: "missing"}, {Path: "/id", Msg: "pattern"}})
	want := []Issue{LocatedIssue("", "missing"), LocatedIssue("/id", "pattern")}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("SchemaIssues = %+v, want %+v", got, want)
	}
}

func TestValidatorsReturnAValidationErrorNamingTheContract(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		validate func([]byte) error
		raw      string
		contract Contract
		located  bool
	}{
		{"verdict not JSON", ValidateVerdict, `not json`, ContractConsumerVerdict, false},
		{"verdict wrong shape", ValidateVerdict, `{"unexpected":true}`, ContractConsumerVerdict, true},
		{"artifact wrong shape", ValidateRunArtifact, `[]`, ContractRunArtifact, true},
		{"ground truth wrong shape", ValidateGroundTruth, `{}`, ContractGroundTruth, true},
		{"scenario wrong shape", ValidateScenario, `{"schema_version":"0.1"}`, ContractScenario, true},
		{"scenario not JSON", ValidateScenario, `{`, ContractScenario, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var invalid *ValidationError
			if err := tc.validate([]byte(tc.raw)); !errors.As(err, &invalid) {
				t.Fatalf("err = %v, want a *ValidationError", err)
			}
			if invalid.Contract != tc.contract || len(invalid.Issues) == 0 || invalid.Issues[0].Located != tc.located {
				t.Fatalf("%+v, want contract %s with located=%v issues", invalid, tc.contract, tc.located)
			}
		})
	}
}

func TestNotJSONErrorKeepsItsCauseAndSummary(t *testing.T) {
	t.Parallel()
	cause := errors.New("unexpected end of input")
	err := NotJSONError(ContractGroundTruth, "label.json", "decode label.json", cause)
	if !errors.Is(err, cause) || err.Error() != "decode label.json: unexpected end of input" || err.Source != "label.json" {
		t.Fatalf("NotJSONError = %+v (%q)", err, err.Error())
	}
}

func TestWithSourceNamesOnlyAnUnnamedValidationError(t *testing.T) {
	t.Parallel()
	unnamed := fmt.Errorf("wrapped: %w", &ValidationError{Contract: ContractRunArtifact, Summary: "s"})
	var invalid *ValidationError
	if !errors.As(WithSource(unnamed, "run.json"), &invalid) || invalid.Source != "run.json" {
		t.Fatalf("unnamed error not given its source: %+v", invalid)
	}
	named := &ValidationError{Source: "first.json"}
	var renamed *ValidationError
	if !errors.As(WithSource(named, "second.json"), &renamed) || renamed.Source != "first.json" {
		t.Fatalf("a named source was overwritten: %q", named.Source)
	}
	plain := errors.New("plain")
	if got := WithSource(plain, "x"); !errors.Is(got, plain) || got.Error() != "plain" {
		t.Fatal("a non-validation error must be returned unchanged")
	}
	if WithSource(nil, "x") != nil {
		t.Fatal("nil must stay nil")
	}
}

func TestEveryContractMapsToAStableFailureCode(t *testing.T) {
	t.Parallel()
	cases := map[Contract]string{
		ContractDomainSpec: "domain_invalid", ContractOutputAdapter: "adapter_invalid", ContractScenario: "scenario_invalid",
		ContractRunArtifact: "artifact_invalid", ContractConsumerVerdict: "artifact_invalid", ContractGroundTruth: "artifact_invalid",
		ContractDeliveryLedger: "artifact_invalid", ContractStateHistory: "artifact_invalid",
		ContractComponent: "component_invalid", ContractInstrument: "component_invalid", ContractVocabulary: "component_invalid",
		ContractComposition: "composition_invalid", ContractKindFile: "kind_failed", ContractTwinChoices: "kind_failed",
	}
	for contract, want := range cases {
		if got, ok := contract.FailureCode(); !ok || got != want {
			t.Errorf("%s FailureCode = %q, %v, want %q", contract, got, ok, want)
		}
	}
	if code, ok := Contract("unknown").FailureCode(); ok {
		t.Errorf("an unknown contract has code %q", code)
	}
}

func TestFailureCodeOfReadsTheContractOfAWrappedValidationError(t *testing.T) {
	t.Parallel()
	invalid := &ValidationError{Contract: ContractComposition, Summary: "bad"}
	if code, ok := FailureCodeOf(fmt.Errorf("compose: %w", invalid)); !ok || code != "composition_invalid" {
		t.Fatalf("FailureCodeOf = %q, %v, want composition_invalid", code, ok)
	}
	if code, ok := FailureCodeOf(errors.New("plain")); ok {
		t.Fatalf("a plain error has code %q", code)
	}
}
