package acceptance

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func factoryInputs() []string {
	return []string{"--domains-dir", filepath.Join(testsupport.RepositoryRoot(), "examples"), "--adapters-dir", testsupport.AdaptersDir()}
}

func TestTheFactoryUtilitiesWorldValidatesLintsRunsVerifiesReplaysAndReports(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("runs two hours of the factory through the binary")
	}
	b := newBinary(t)
	domainFile := filepath.Join(testsupport.RepositoryRoot(), "examples", "factory-utilities.domain.json")
	b.run("domain", "validate", domainFile).mustSucceed(t)
	if lint := b.run("domain", "lint", domainFile, "--hours", "24").mustSucceed(t).document(t); lint["errors"] != float64(0) || lint["warnings"] != float64(0) {
		t.Fatalf("lint = %v, want a clean nominal day", lint)
	}
	dir := t.TempDir()
	args := append(append([]string{"run"}, factoryInputs()...), "--domain", "factory-utilities", "--seed", "3", "--duration", "7200",
		"--sink", "file", "--out", dir, "--fault", "CT-1=cooling_tower__fan_failure@1800", "--fault", "R-1=air_receiver__drain_stuck_open@3600",
		"--perturb", "drop@600@1200")
	if got := b.run(args...).mustSucceed(t).document(t); got["reproducible"] != true {
		t.Fatalf("run = %v", got)
	}
	for _, command := range []string{"verify", "replay"} {
		got := b.run(append(append([]string{command}, factoryInputs()...), filepath.Join(dir, "run.json"))...).mustSucceed(t).document(t)
		if got["matches"] != true || got["version_match"] != true {
			t.Fatalf("%s = %v", command, got)
		}
	}
	checkFactoryReport(t, b, dir)
}

func checkFactoryReport(t *testing.T, b *binary, dir string) {
	t.Helper()
	out := filepath.Join(t.TempDir(), "report.md")
	b.run("report", dir, "--format", "markdown", "--out", out).mustSucceed(t)
	if info, err := os.Stat(out); err != nil || info.Size() == 0 {
		t.Fatalf("markdown report: %v", err)
	}
	report := b.run("report", dir).mustSucceed(t).document(t)
	truth, _ := report["truth"].(map[string]any)
	faults, _ := truth["faults"].([]any)
	if report["truth_unavailable"] != nil || len(faults) != 2 {
		t.Fatalf("report truth = %v (%v), want the two injected faults of the run it summarizes", truth, report["truth_unavailable"])
	}
	for i, want := range []string{"cooling_tower__fan_failure", "air_receiver__drain_stuck_open"} {
		if got := faults[i].(map[string]any)["fault"]; got != want {
			t.Errorf("truth fault %d = %v, want %s", i, got, want)
		}
	}
	checkFactoryKPIs(t, truth)
}

func checkFactoryKPIs(t *testing.T, truth map[string]any) {
	t.Helper()
	kpis, _ := truth["kpis"].(map[string]any)
	entities, _ := kpis["entities"].([]any)
	found := map[string]bool{}
	for _, raw := range entities {
		entity := raw.(map[string]any)
		for _, key := range []string{"kpis", "alarms"} {
			items, _ := entity[key].([]any)
			for _, item := range items {
				found[entity["entity_id"].(string)+"."+item.(map[string]any)["state"].(string)] = true
			}
		}
	}
	for _, want := range []string{"CH-1.cop", "CT-1.approach", "AC-1.specific_power", "AC-1.temp_alarm", "TX-1.temp_alarm", "MM-1.loading"} {
		if !found[want] {
			t.Errorf("KPI truth has no %s: %v", want, kpis)
		}
	}
	if system, _ := kpis["alarm_system"].(map[string]any); system == nil {
		t.Errorf("KPI truth has no alarm-system KPIs: %v", kpis)
	}
}
