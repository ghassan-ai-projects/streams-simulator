package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

const (
	benchStartNS = int64(1767225600000000000)
	benchTickNS  = int64(1_000_000_000)
)

func benchBirth(device, kind string, metrics ...string) map[string]any {
	list := []any{}
	for i, name := range metrics {
		list = append(list, map[string]any{"name": name, "alias": i + 1, "value": 34.0, "quality": "good"})
	}
	return map[string]any{"type": "DBIRTH", "device": device, "kind": kind, "board": "mega-01", "birth_time_ns": benchStartNS, "metrics": list}
}

func benchTick(at int64, temperature float64) map[string]any {
	return map[string]any{"type": "DDATA", "device": "zone-01", "historical": false, "metrics": []any{
		map[string]any{"alias": 1, "value": temperature, "quality": "good", "event_time_ns": at}}}
}

func benchLease(t *testing.T) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(testsupport.RepositoryRoot(), "contracts", "plant-wire", "fixtures", "valid", "DCMD.json"))
	if err != nil {
		t.Fatal(err)
	}
	var record map[string]any
	if err := json.Unmarshal(raw, &record); err != nil {
		t.Fatal(err)
	}
	record["command_id"], record["device"], record["command"] = "cmd-1", "fan-01", "set_cooling"
	record["args"] = map[string]any{"level_percent": 30}
	return record
}

// benchRunFolder records the zone every 5 s for 100 s with one accepted fan lease after 55 s;
// from 80 s on the zone reads the hot value.
func benchRunFolder(t *testing.T, hot float64) string {
	t.Helper()
	dir := t.TempDir()
	trace, err := run.NewFieldTrace(dir)
	if err != nil {
		t.Fatal(err)
	}
	records := []map[string]any{benchBirth("zone-01", "thermal_sensor", "temperature", "humidity"), benchBirth("fan-01", "fan", "level_percent")}
	for at := benchStartNS + 5*benchTickNS; at <= benchStartNS+100*benchTickNS; at += 5 * benchTickNS {
		value := 34.0
		if at >= benchStartNS+80*benchTickNS {
			value = hot
		}
		records = append(records, benchTick(at, value))
		if at == benchStartNS+55*benchTickNS {
			records = append(records, map[string]any{"type": "DRECEIPT", "command_id": "cmd-1", "status": "accepted"},
				map[string]any{"type": "DRESULT", "command_id": "cmd-1", "reported_outcome": "succeeded"})
		}
	}
	for _, record := range records {
		if err := trace.Add(record); err != nil {
			t.Fatal(err)
		}
	}
	if err := trace.Close(); err != nil {
		t.Fatal(err)
	}
	wire, err := run.NewWireTrace(dir)
	if err != nil {
		t.Fatal(err)
	}
	wire.Wire("host", benchLease(t))
	if err := wire.Close(); err != nil {
		t.Fatal(err)
	}
	return dir
}

func twinReportArgs(t *testing.T, dir string, extra ...string) []string {
	t.Helper()
	raw, err := os.ReadFile(testsupport.ThermalBoardDomain())
	if err != nil {
		t.Fatal(err)
	}
	domains := t.TempDir()
	testsupport.WriteFileIn(t, domains, "arduino-thermal.domain.json", raw)
	args := []string{"plantwire", "twin-report", dir, "--domain", "arduino-thermal", "--domains-dir", domains, "--kinds-dir", testsupport.ThermalBoardKindsDir()}
	return append(args, extra...)
}

func TestPlantWireTwinReportUsageIsChecked(t *testing.T) {
	t.Parallel()
	requireUsageRefusals(t, [][]string{
		{"plantwire", "twin-report"},
		{"plantwire", "twin-report", "run-1", "--kinds-dir", "kinds"},
		{"plantwire", "twin-report", "run-1", "--domain", "arduino-thermal"},
		{"plantwire", "twin-report", "run-1", "run-2", "--domain", "arduino-thermal", "--kinds-dir", "kinds"},
		{"plantwire", "twin-report", "run-1", "--domain", "arduino-thermal", "--kinds-dir", "kinds", "--adapters-dir", "adapters"},
		{"plantwire", "twin-report", "run-1", "--domain", "arduino-thermal", "--kinds-dir", "kinds", "--json"},
	})
}

func TestPlantWireNamesTwinReportAmongItsSubcommands(t *testing.T) {
	t.Parallel()
	got := invoke(t, "plantwire", "no-such-subcommand")
	body := got.failure(t)
	if message, _ := body["message"].(string); got.code != 2 || !strings.Contains(message, "twin-report") {
		t.Fatalf("exit %d with %v, want the usage refusal to list twin-report", got.code, body)
	}
}

func TestATwinReportOfADivergingBenchPrintsOneDocumentOfItsContract(t *testing.T) {
	t.Parallel()
	got := invoke(t, twinReportArgs(t, benchRunFolder(t, 41))...).mustSucceed(t)
	doc := got.document(t, "plantwire_twin_report")
	raw, err := os.ReadFile(testsupport.Contract("twin-report-v0.1"))
	if err != nil {
		t.Fatal(err)
	}
	schema, err := jsonschema.CompileJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	var parsed any
	if err := model.DecodeBytes([]byte(got.stdout), &parsed); err != nil {
		t.Fatal(err)
	}
	if errs := schema.Validate(parsed); len(errs) > 0 {
		t.Fatalf("stdout does not match twin-report-v0.1:\n%s", jsonschema.FormatErrors(errs))
	}
	summary, _ := doc["summary"].(map[string]any)
	if doc["verdict"] != "diverges" || summary["divergences"] != float64(1) || summary["commands_applied"] != float64(1) {
		t.Fatalf("verdict %v summary %v, want diverges with one divergence and the lease applied", doc["verdict"], summary)
	}
}

func TestATwinReportOfABenchThatFollowsItsTwinAgrees(t *testing.T) {
	t.Parallel()
	doc := invoke(t, twinReportArgs(t, benchRunFolder(t, 34))...).mustSucceed(t).document(t, "plantwire_twin_report")
	if doc["verdict"] != "agrees" {
		t.Fatalf("verdict %v, want agrees", doc["verdict"])
	}
}

func TestDeclaredTolerancesMayBeRepeatedOrCommaSeparated(t *testing.T) {
	t.Parallel()
	dir := benchRunFolder(t, 41)
	args := twinReportArgs(t, dir, "--tolerance", "zone-01.temperature=20", "--tolerance", "thermal_sensor.humidity=9,fan-01.level_percent=50")
	doc := invoke(t, args...).mustSucceed(t).document(t, "plantwire_twin_report")
	metrics, _ := doc["metrics"].([]any)
	tolerances := map[string]float64{}
	for _, entry := range metrics {
		metric, _ := entry.(map[string]any)
		tolerance, _ := metric["tolerance"].(map[string]any)
		value, _ := tolerance["value"].(float64)
		tolerances[metric["device"].(string)+"."+metric["metric"].(string)] = value
	}
	want := map[string]float64{"zone-01.temperature": 20, "zone-01.humidity": 9, "fan-01.level_percent": 50}
	for key, value := range want {
		if tolerances[key] != value {
			t.Errorf("tolerance of %s = %v, want the declared %v (all: %v)", key, tolerances[key], value, tolerances)
		}
	}
	if doc["verdict"] != "agrees" {
		t.Fatalf("verdict %v, want agrees: the 7 degC gap is inside the declared 20", doc["verdict"])
	}
}

func TestATwinReportThatCannotBeMadeFailsWithItsOwnCode(t *testing.T) {
	t.Parallel()
	dir := benchRunFolder(t, 41)
	tests := []struct {
		name    string
		args    []string
		code    string
		message string
	}{
		{"a tolerance that names nothing", twinReportArgs(t, dir, "--tolerance", "zone-09.temperature=1"), "plantwire_failed", "zone-09.temperature"},
		{"a folder with no field trace", twinReportArgs(t, t.TempDir()), "plantwire_failed", "no field trace"},
		{"a domain the directory lacks", append(twinReportArgs(t, dir), "--domain", "no-such-twin"), "domain_not_found", "no-such-twin"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := invoke(t, tt.args...)
			body := got.failure(t)
			message, _ := body["message"].(string)
			if got.code != 1 || body["code"] != tt.code || !strings.Contains(message, tt.message) {
				t.Fatalf("exit %d with %v, want %s mentioning %q", got.code, body, tt.code, tt.message)
			}
		})
	}
}
