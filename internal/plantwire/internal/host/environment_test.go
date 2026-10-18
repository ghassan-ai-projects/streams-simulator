package host

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
)

func environmentBirth(t *testing.T, node string, boards ...string) domain.Inbound {
	t.Helper()
	listed := []any{}
	for i, environment := range boards {
		listed = append(listed, map[string]any{"board": fmt.Sprintf("b-%d", i+1), "environment": environment, "members": []any{fmt.Sprintf("P%d", i+1)}})
	}
	return birth(t, node, map[string]any{"boards": listed})
}

func TestTheHandshakeChecksTheNodesEnvironmentAgainstItsBoardsAndTheHostsOwnSetting(t *testing.T) {
	t.Parallel()
	const sim, hw, mixed = domain.EnvironmentSimulated, domain.EnvironmentHardware, domain.EnvironmentMixed
	cases := []struct {
		name   string
		host   string
		node   string
		boards []string
		reason string
	}{
		{"a simulated host takes a simulated node", sim, sim, []string{sim, sim}, ""},
		{"a hardware host takes a hardware node", hw, hw, []string{hw}, ""},
		{"a mixed host takes a simulated node", mixed, sim, []string{sim}, ""},
		{"a mixed host takes a hardware node", mixed, hw, []string{hw}, ""},
		{"a mixed host takes a mixed node", mixed, mixed, []string{sim, hw}, ""},
		{"a simulated host refuses a mixed node", sim, mixed, []string{sim, hw}, "the host accepts simulated nodes only, the node announces mixed"},
		{"a hardware host refuses a mixed node", hw, mixed, []string{sim, hw}, "the host accepts hardware nodes only, the node announces mixed"},
		{"a simulated host refuses a hardware node", sim, hw, []string{hw}, "the host accepts simulated nodes only, the node announces hardware"},
		{"a hardware node with a simulated board", mixed, hw, []string{hw, sim}, `node environment hardware needs every board hardware, board "b-2" is "simulated"`},
		{"a simulated node with a hardware board", mixed, sim, []string{hw}, `node environment simulated needs every board simulated, board "b-1" is "hardware"`},
		{"a mixed node with only simulated boards", mixed, mixed, []string{sim, sim}, "none of its 2 boards is hardware"},
		{"a mixed node with only hardware boards", mixed, mixed, []string{hw}, "none of its 1 boards is simulated"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			h := New(Options{Environment: tc.host, ProtocolMax: 1, Epoch: 1})
			replies := h.Receive(environmentBirth(t, tc.node, tc.boards...))
			if len(replies) != 1 || replies[0]["type"] != "STATE" || domain.Check(replies[0]) != nil {
				t.Fatalf("replies %v, want one valid STATE", replies)
			}
			assertEnvironmentAnswer(t, h, replies[0], tc.reason)
		})
	}
}

func assertEnvironmentAnswer(t *testing.T, h *Host, state map[string]any, reason string) {
	t.Helper()
	report := h.Report()
	if reason == "" {
		if state["online"] != true || state["refusal"] != nil || report.Refused != "" || report.Reason != "" || !h.Accepted() {
			t.Fatalf("STATE %v with report %+v, want the session accepted", state, report)
		}
		return
	}
	if state["online"] != false || state["refusal"] != "environment_mismatch" || report.Refused != "environment_mismatch" || !strings.Contains(report.Reason, reason) || h.Accepted() {
		t.Fatalf("STATE %v with report %+v, want an environment_mismatch refusal saying %q", state, report, reason)
	}
}

func TestAnEnvironmentRefusalDoesNotHideAnEarlierProtocolRefusalNorSwallowALaterOne(t *testing.T) {
	t.Parallel()
	h := New(Options{Environment: domain.EnvironmentSimulated, ProtocolMax: 1})
	h.Receive(environmentBirth(t, domain.EnvironmentSimulated, domain.EnvironmentSimulated))
	if report := h.Report(); report.Refused != "" || report.Reason != "" {
		t.Fatalf("report %+v, want a consistent simulated node accepted with no reason", report)
	}
	protocol := New(Options{Environment: domain.EnvironmentSimulated, ProtocolMax: 1})
	protocol.Receive(birth(t, "hardware", map[string]any{"protocol_min": 2, "protocol_max": 1}))
	if report := protocol.Report(); report.Refused != "protocol_invalid" || report.Reason != "" {
		t.Fatalf("report %+v, want protocol_invalid first, with no environment reason", report)
	}
}
