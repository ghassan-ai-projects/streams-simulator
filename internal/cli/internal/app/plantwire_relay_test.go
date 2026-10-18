package app

import (
	"encoding/json"
	"net"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestPlantWireRelayUsageIsChecked(t *testing.T) {
	t.Parallel()
	requireUsageRefusals(t, [][]string{{"plantwire", "relay"}, {"plantwire", "relay", "--gateway-socket", "/tmp/g.sock"},
		{"plantwire", "relay", "--host-socket", "/tmp/h.sock"}, {"plantwire", "relay", "--gateway-socket", "/tmp/g.sock", "--host-socket", "/tmp/h.sock", "extra"},
		{"plantwire", "relay", "--gateway-socket", "/tmp/g.sock", "--host-socket", "/tmp/h.sock", "--domain", "pump-board"}})
}

func requireUsageRefusals(t *testing.T, cases [][]string) {
	t.Helper()
	for _, args := range cases {
		got := invoke(t, args...)
		if body := got.failure(t); got.code != 2 || body["code"] != "usage_invalid" {
			t.Errorf("%v exited %d with %v, want a usage refusal", args, got.code, body)
		}
	}
}

func TestOnlyTheRelayHasAGatewayOptionSoNoReplayOrRunCanReachHardware(t *testing.T) {
	t.Parallel()
	for _, args := range [][]string{{"run", "--domain", "pump-board", "--gateway-socket", "/tmp/g.sock"}, {"replay", "--gateway-socket", "/tmp/g.sock", "artifact"},
		{"verify", "--gateway-socket", "/tmp/g.sock", "artifact"}, {"scenario", "from-run", "--gateway-socket", "/tmp/g.sock", "artifact"},
		{"plantwire", "node", "--host-socket", "/tmp/h.sock", "--domain", "pump-board", "--out", "/tmp/o", "--gateway-socket", "/tmp/g.sock"},
		{"plantwire", "host", "--socket", "/tmp/h.sock", "--gateway-socket", "/tmp/g.sock"}} {
		got := invoke(t, args...)
		body := got.failure(t)
		message, _ := body["message"].(string)
		if got.code != 2 || body["code"] != "usage_invalid" || !strings.Contains(message+got.stderr, "gateway-socket") {
			t.Errorf("%v exited %d with %v, want a usage refusal of the undefined --gateway-socket", args, got.code, body)
		}
	}
}

func TestAPlantWireRelayOnAGatewaySocketInUseFailsBeforeItServesAnything(t *testing.T) {
	t.Parallel()
	gatewayPath := testsupport.SocketPath(t)
	hostPath := filepath.Join(filepath.Dir(gatewayPath), "h.sock")
	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "unix", gatewayPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	got := invoke(t, "plantwire", "relay", "--gateway-socket", gatewayPath, "--host-socket", hostPath)
	body := got.failure(t)
	if message, _ := body["message"].(string); got.code != 1 || body["code"] != "plantwire_failed" || !strings.Contains(message, "in use") {
		t.Fatalf("exit %d with %v, want plantwire_failed saying the socket path is in use", got.code, body)
	}
}

func TestARelayReportMatchesTheCliOutputContract(t *testing.T) {
	t.Parallel()
	for name, stats := range map[string]plantwire.RelayStats{"a quiet relay": {},
		"a busy relay": {GatewaySessions: 2, HostSessions: 3, Up: 40, Down: 12, Dropped: 1, Refused: "environment_mismatch", Reason: "the node announces simulated", HostRefused: "inventory_mismatch"}} {
		raw, err := json.Marshal(stats)
		if err != nil {
			t.Fatal(err)
		}
		var doc any
		if err := model.DecodeBytes(raw, &doc); err != nil {
			t.Fatal(err)
		}
		for _, def := range []string{"plantwire_relay", ""} {
			if errs := cliOutputSchema(t, def).Validate(doc); len(errs) > 0 {
				t.Errorf("%s does not match cli-output %q:\n%s\n%s", name, def, jsonschema.FormatErrors(errs), raw)
			}
		}
	}
}
