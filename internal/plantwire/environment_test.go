package plantwire

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestAMixedHostAcceptsASimulatedNodeAndTheJournalHoldsItsBoardsAsSimulated(t *testing.T) {
	t.Parallel()
	options := referenceOptions()
	options.Environment = "mixed"
	result := pair(t, options, 0, 5*time.Second, nil)
	if result.node != nil || result.host != nil || result.report.Refused != "" || result.report.Reason != "" {
		t.Fatalf("node %v, host %v, report %+v, want a mixed host to accept a simulated node", result.node, result.host, result.report)
	}
	var birth map[string]any
	for _, wire := range journalRecords(t, result.run, "wire") {
		if record, _ := wire["record"].(map[string]any); wire["direction"] == "node" && record["type"] == "NBIRTH" && birth == nil {
			birth = record
		}
	}
	boards, _ := birth["boards"].([]any)
	if birth["contract_version"] != "1.3.0" || birth["environment"] != "simulated" || len(boards) == 0 {
		t.Fatalf("journalled NBIRTH %v, want contract 1.3.0, a simulated node and its boards", birth)
	}
	for _, board := range boards {
		if got := board.(map[string]any)["environment"]; got != "simulated" {
			t.Errorf("journalled board %v is %v, want simulated", board.(map[string]any)["board"], got)
		}
	}
}

func TestAHostThatRefusesAnEnvironmentSaysWhichRule(t *testing.T) {
	t.Parallel()
	wrong := referenceOptions()
	wrong.Environment = "hardware"
	result := pair(t, wrong, 0, 5*time.Second, nil)
	reason := "the host accepts hardware nodes only, the node announces simulated"
	if !errors.Is(result.node, ErrRefused) || result.report.Refused != "environment_mismatch" || !strings.Contains(result.report.Reason, reason) {
		t.Fatalf("node %v, report %+v, want an environment_mismatch refusal saying %q", result.node, result.report, reason)
	}
}
