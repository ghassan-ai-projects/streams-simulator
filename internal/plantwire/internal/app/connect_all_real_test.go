package app

import (
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/host"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/standin"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func allRealPlan(hostPath, gatewayPath string) string {
	bind := func(instance string) string {
		return `{"instance":"` + instance + `","board":"uno-07","device":"` + instance + `"}`
	}
	return `{"schema_version":"0.1","domain":"arduino-thermal","seed":3,"host_socket":"` + hostPath + `","gateway_socket":"` + gatewayPath +
		`","real":[` + bind("zone-01") + `,` + bind("led-01") + `,` + bind("fan-01") + `]}`
}

func allRealGateway(t *testing.T) *standin.Node {
	t.Helper()
	node, err := standin.New(standin.Config{Session: "bench-1", Board: "uno-07",
		Kinds: []map[string]any{thermalKindFile(t, "thermal_sensor.kind.json"), thermalKindFile(t, "indicator.kind.json"), thermalKindFile(t, "fan.kind.json")},
		Devices: []standin.Device{{Name: "zone-01", Kind: "thermal_sensor", BootID: "61", Initial: map[string]any{"temperature": 24.5}},
			{Name: "led-01", Kind: "indicator", BootID: "61"}, {Name: "fan-01", Kind: "fan", BootID: "61"}}})
	if err != nil {
		t.Fatal(err)
	}
	return node
}

func TestAPlanWithEveryInstanceRealIsAHardwareSessionThatAStrictHostAccepts(t *testing.T) {
	t.Parallel()
	for _, environment := range []string{domain.EnvironmentHardware, domain.EnvironmentMixed} {
		t.Run(environment, func(t *testing.T) {
			t.Parallel()
			hostPath, out := testsupport.SocketPath(t), filepath.Join(t.TempDir(), "c-1")
			gatewayPath := filepath.Join(filepath.Dir(hostPath), "g.sock")
			ready, ended := make(chan struct{}), make(chan hostEnd, 1)
			go func() {
				options := host.Options{Environment: environment, ProtocolMax: 1, Ready: func() { close(ready) }}
				report, err := ServeHost(t.Context(), hostPath, 0, options)
				ended <- hostEnd{report: report, err: err}
			}()
			<-ready
			options := connectOptions(t, allRealPlan(hostPath, gatewayPath), out)
			stop, done := startConnect(t, options, true)
			state := dialGateway(t, gatewayPath, allRealGateway(t)).handshake()
			stop()
			<-done
			end := <-ended
			if !state.Online || end.err != nil || end.report.Refused != "" {
				t.Fatalf("STATE %v, host report %+v, %v; want a %s host to accept the all-real session", state.Record, end.report, end.err, environment)
			}
			sessions := journalOfType(t, out, "session")
			if plan, _ := sessions[0]["plan"].(map[string]any); plan["environment"] != domain.EnvironmentHardware {
				t.Fatalf("journalled plan = %v, want the hardware environment", sessions[0])
			}
			replaysWhole(t, out, options.DomainsDir)
			check, err := CheckConnect(connectOptions(t, allRealPlan(hostPath, gatewayPath), ""))
			if err != nil || check.Environment != domain.EnvironmentHardware {
				t.Fatalf("CheckConnect = %+v, %v, want a hardware session", check, err)
			}
		})
	}
}
