package domain

import (
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestCrossReferenceValidationChecksChannelsBeforeFaults(t *testing.T) {
	t.Parallel()
	spec, err := Load(testsupport.Domain("rotating-machinery"))
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
