package app

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

const heatLoad = `,"disturbances":[{"at_s":120,"instance":"zone-01","fault":"zone_heat_load","duration_s":600,"severity":0.5}]}`

func commandLog(t *testing.T, out string) []model.Command {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(out, model.FileRunArtifact))
	if err != nil {
		t.Fatal(err)
	}
	var artifact struct {
		CommandLog []model.Command `json:"command_log"`
	}
	if err := json.Unmarshal(raw, &artifact); err != nil {
		t.Fatal(err)
	}
	return artifact.CommandLog
}

func fileHolds(t *testing.T, out, name, text string) bool {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(out, name))
	if err != nil {
		t.Fatal(err)
	}
	return strings.Contains(string(raw), text)
}

func TestAPlannedDisturbanceIsInjectedAtSessionStartRecordedAndSealedAndReplays(t *testing.T) {
	t.Parallel()
	hostPath, out := testsupport.SocketPath(t), filepath.Join(t.TempDir(), "c-1")
	host := serveWireHost(t, hostPath, false)
	options := connectOptions(t, strings.TrimSuffix(planText(hostPath, ""), "}")+heatLoad, out)
	cancel, done := startConnect(t, options, false)
	host.waitFor(func(r []map[string]any) bool { return len(syncTimes(r)) >= 15 })
	cancel()
	if ended := <-done; ended.err != nil {
		t.Fatal(ended.err)
	}
	log := commandLog(t, out)
	if len(log) < 2 || log[0].Op != model.OpFaultInject || log[1].Op != model.OpFaultClear || log[0].Args["entity_id"] != "zone-01" ||
		log[0].Args["onset_ns"] != float64(model.DefaultStartTimeNS+120e9) || log[1].Args["at_ns"] != float64(model.DefaultStartTimeNS+720e9) {
		t.Fatalf("command log starts %+v, want the heat load injected at start + 120 s and cleared 600 s later", log[:min(len(log), 2)])
	}
	for _, name := range []string{model.FileObserverJournal, model.FileTrace} {
		if fileHolds(t, out, name, "zone_heat_load") {
			t.Errorf("%s names the disturbance: the schedule is sealed truth", name)
		}
	}
	if !fileHolds(t, out, model.FileDirectorJournal, "zone_heat_load") {
		t.Error("the sealed journal does not hold the disturbance")
	}
	for _, record := range host.snapshot() {
		if raw, _ := json.Marshal(record); strings.Contains(string(raw), "zone_heat_load") || strings.Contains(string(raw), "disturbance") {
			t.Fatalf("the host received a %v record that names the disturbance", record["type"])
		}
	}
	replaysWhole(t, out, options.DomainsDir)
}

func TestTheCheckOfAPlanWithDisturbancesNeverEchoesThemAndRefusesABrokenOne(t *testing.T) {
	t.Parallel()
	check, err := CheckConnect(connectOptions(t, strings.TrimSuffix(planText("/tmp/h.sock", ""), "}")+heatLoad, ""))
	if raw, _ := json.Marshal(check); err != nil || strings.Contains(string(raw), "zone_heat_load") || strings.Contains(string(raw), "disturbance") {
		t.Fatalf("check %s, %v, want a passing check that does not echo the schedule", raw, err)
	}
	broken := strings.Replace(heatLoad, "zone_heat_load", "fan_stall", 1)
	_, err = CheckConnect(connectOptions(t, strings.TrimSuffix(planText("/tmp/h.sock", ""), "}")+broken, ""))
	invalid, ok := model.AsValidationError(err)
	if !errors.Is(err, ErrPlanRefused) || !ok || invalid.Issues[0].Pointer != "/disturbances/0/fault" || invalid.Issues[0].Keyword != "unknown_fault" {
		t.Fatalf("a fan fault on the zone: %v, want unknown_fault at /disturbances/0/fault", err)
	}
}

func TestADisturbanceStartsWithinTheLongestRun(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(filepath.Join(testsupport.RepositoryRoot(), "contracts", "session-plan-v0.1.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	var schema struct {
		Defs struct {
			Disturbance struct {
				Properties map[string]struct {
					Maximum float64 `json:"maximum"`
				} `json:"properties"`
			} `json:"disturbance"`
		} `json:"$defs"`
	}
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatal(err)
	}
	for _, member := range []string{"at_s", "duration_s"} {
		if got := schema.Defs.Disturbance.Properties[member].Maximum; got*1e9 != float64(run.MaxRunSpanNS) {
			t.Errorf("%s maximum %v s, the longest run is %d s", member, got, run.MaxRunSpanNS/1e9)
		}
	}
}
