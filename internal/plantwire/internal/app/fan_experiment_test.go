package app

import (
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/host"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func fanExperimentPlan(t *testing.T, name string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(testsupport.ThermalBoardKindsDir(), "fan-experiment", name+".plan.json"))
	if err != nil {
		t.Fatal(err)
	}
	var plan map[string]any
	if err := json.Unmarshal(raw, &plan); err != nil {
		t.Fatal(err)
	}
	return plan
}

func planOptions(t *testing.T, plan map[string]any, out string) ConnectOptions {
	t.Helper()
	raw, err := json.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	return connectOptions(t, string(raw), out)
}

func originsOf(check ConnectCheck) map[string]string {
	origins := map[string]string{}
	for _, instance := range check.Instances {
		origins[instance.Instance] = instance.Origin + "@" + instance.Board
	}
	return origins
}

func TestTheFanExperimentPlansPassTheirCheckAndDifferOnlyInTheGatewayTheBindingsAndTheDisturbances(t *testing.T) {
	t.Parallel()
	simulate, real := fanExperimentPlan(t, "simulate"), fanExperimentPlan(t, "real")
	simulated, err := CheckConnect(planOptions(t, simulate, ""))
	if err != nil || simulated.Environment != domain.EnvironmentSimulated || len(simulated.Policies) != 1 ||
		!reflect.DeepEqual(originsOf(simulated), map[string]string{"zone-01": "simulated@", "led-01": "simulated@", "fan-01": "simulated@"}) {
		t.Fatalf("simulate.plan.json check = %+v, %v, want every instance simulated and one policy", simulated, err)
	}
	hardware, err := CheckConnect(planOptions(t, real, ""))
	if err != nil || hardware.Environment != domain.EnvironmentHardware || len(hardware.Policies) != 1 ||
		!reflect.DeepEqual(originsOf(hardware), map[string]string{"zone-01": "real@mega-01", "led-01": "real@mega-01", "fan-01": "real@mega-01"}) {
		t.Fatalf("real.plan.json check = %+v, %v, want every instance real on mega-01 and one policy", hardware, err)
	}
	common := func(plan map[string]any) map[string]any {
		kept := maps.Clone(plan)
		delete(kept, "gateway_socket")
		delete(kept, "real")
		delete(kept, "disturbances")
		return kept
	}
	if !reflect.DeepEqual(common(simulate), common(real)) || simulate["pace"] != "wall" {
		t.Fatalf("the plans differ beyond gateway_socket, real and disturbances, or are not paced: %v and %v", simulate, real)
	}
}

func rehearsalScript(at time.Time) []host.Scripted {
	return []host.Scripted{
		{ID: "led-alert", AtNS: at.UnixNano(), Device: "led-01", Command: "show_state", Args: map[string]any{"state": "alert"}},
		{ID: "fan-45", AtNS: at.UnixNano(), Device: "fan-01", Command: "set_cooling", Args: map[string]any{"level_percent": 45},
			Verify: &host.Verify{Metric: "level_percent", Equals: 45, Tolerance: 0.5, WithinNS: int64(10 * time.Second)}},
	}
}

func rehearsalHost(environment string, policy any, script []host.Scripted) host.Options {
	digest, _ := policy.(string)
	return host.Options{Environment: environment, ProtocolMax: 1, Epoch: 7, Policy: digest, Script: script}
}

func commandsSucceeded(t *testing.T, report host.Report, ids ...string) {
	t.Helper()
	for _, id := range ids {
		status := report.Commands[id]
		if status == nil || !reflect.DeepEqual(status.Receipts, []string{"accepted"}) || status.Result != "succeeded" {
			t.Errorf("command %s = %+v, want accepted and succeeded", id, status)
		}
	}
}

func TestTheSimulatedFanExperimentHeatsTheZoneAndShowsTheCommandsInItsReadings(t *testing.T) {
	t.Parallel()
	hostPath, out := testsupport.SocketPath(t), filepath.Join(t.TempDir(), "c-sim")
	plan := fanExperimentPlan(t, "simulate")
	plan["host_socket"] = hostPath
	plan["disturbances"] = []any{map[string]any{"at_s": 1, "instance": "zone-01", "fault": "zone_heat_load", "duration_s": 900}}
	started := startHost(t, hostPath, rehearsalHost("simulated", plan["policies"].([]any)[0], rehearsalScript(time.Now().Add(3*time.Second))))
	options := planOptions(t, plan, out)
	options.SliceNS = int64(500 * time.Millisecond)
	cancel, done := startConnect(t, options, false)
	awaitLive(t, out, func(readings map[string]any) bool {
		return readings["led-01.state"] == "alert" && readings["fan-01.level_percent"] == 45.0
	})
	cancel()
	if ended := <-done; ended.err != nil {
		t.Fatalf("Connect = %+v", ended)
	}
	report := started.stop()
	commandsSucceeded(t, report, "led-alert", "fan-45")
	if verified := report.Commands["fan-45"].Verified; verified == nil || !*verified {
		t.Errorf("fan-45 verification = %v, want level_percent 45 observed", verified)
	}
	if rest, warmest := zoneTemperatures(t, out); warmest <= rest {
		t.Errorf("zone_temperature stayed at or below %v (warmest %v), want the heat load to raise it", rest, warmest)
	}
}

func TestTheRealFanExperimentAdmitsTheStandInBoardAndCarriesTheCommandsToIt(t *testing.T) {
	t.Parallel()
	hostPath, out := testsupport.SocketPath(t), filepath.Join(t.TempDir(), "c-real")
	gatewayPath := filepath.Join(filepath.Dir(hostPath), "g.sock")
	plan := fanExperimentPlan(t, "real")
	plan["host_socket"], plan["gateway_socket"] = hostPath, gatewayPath
	started := startHost(t, hostPath, rehearsalHost("hardware", plan["policies"].([]any)[0], rehearsalScript(time.Unix(0, 1))))
	cancel, done := startConnect(t, planOptions(t, plan, out), true)
	gateway := dialGateway(t, gatewayPath, rehearsalBoard(t))
	if state := gateway.handshake(); !state.Online {
		t.Fatalf("STATE = %v, want the stand-in board admitted against the real plan", state.Record)
	}
	commands := 0
	for _, in := range gateway.sync() {
		if in.Type == "DCMD" {
			commands++
		}
	}
	gateway.sync()
	cancel()
	if ended := <-done; ended.err != nil || commands != 2 {
		t.Fatalf("Connect = %+v with %d commands at the board, want a clean stop after both commands", ended, commands)
	}
	commandsSucceeded(t, started.stop(), "led-alert", "fan-45")
}
