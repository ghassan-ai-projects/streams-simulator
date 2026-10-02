package domain

import (
	"strings"
	"testing"
)

func TestCrossReferenceValidationChecksChannelsBeforeFaults(t *testing.T) {
	spec, err := Load("../../domains/rotating-machinery.domain.json")
	if err != nil {
		t.Fatal(err)
	}
	spec.Spec.Channels[0].Observes = "undeclared"
	spec.Spec.Faults[0].Affects[0].State = "also_undeclared"
	err = crossCheck(spec, "ordering-input")
	if err == nil || !strings.Contains(err.Error(), `ordering-input: channel`) || !strings.Contains(err.Error(), `observes undeclared state "undeclared"`) {
		t.Fatalf("validation priority changed: %v", err)
	}
}
