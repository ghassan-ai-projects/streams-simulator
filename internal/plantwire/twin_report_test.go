package plantwire

import (
	"errors"
	"os"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

const twinBirthNS = 1767225600000000000

func twinFolder(t *testing.T, trace []map[string]any, withJournal bool) string {
	t.Helper()
	dir := t.TempDir()
	field, err := run.NewFieldTrace(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range trace {
		if err := field.Add(record); err != nil {
			t.Fatal(err)
		}
	}
	if err := field.Close(); err != nil {
		t.Fatal(err)
	}
	if withJournal {
		wire, err := run.NewWireTrace(dir)
		if err != nil {
			t.Fatal(err)
		}
		if err := wire.Close(); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func twinBoard(device, kind string) []map[string]any {
	return []map[string]any{{"type": "DBIRTH", "device": device, "kind": kind, "board": "mega-01", "birth_time_ns": twinBirthNS,
		"metrics": []any{map[string]any{"name": "temperature", "alias": 1, "value": 34.0, "quality": "good"}}}}
}

func twinOptions(t *testing.T, dir string) TwinReportOptions {
	t.Helper()
	raw, err := os.ReadFile(testsupport.ThermalBoardDomain())
	if err != nil {
		t.Fatal(err)
	}
	domains := t.TempDir()
	testsupport.WriteFileIn(t, domains, "arduino-thermal.domain.json", raw)
	return TwinReportOptions{Run: dir, DomainsDir: domains, Domain: "arduino-thermal", KindsDirs: []string{testsupport.ThermalBoardKindsDir()}}
}

func TestTheFacadeReportsARunFolderWithNoCommandAsUnjudgeable(t *testing.T) {
	t.Parallel()
	report, err := TwinReport(t.Context(), twinOptions(t, twinFolder(t, twinBoard("zone-01", "thermal_sensor"), true)))
	if err != nil {
		t.Fatalf("TwinReport: %v", err)
	}
	if report.SchemaVersion != "0.1" || report.Verdict != "cannot_judge" || report.Reason != "no_commands_applied" || report.Twin.Domain != "arduino-thermal" {
		t.Fatalf("report = %s %s %s %s, want an arduino-thermal report that cannot judge for want of a command",
			report.SchemaVersion, report.Verdict, report.Reason, report.Twin.Domain)
	}
}

func TestTheFacadeReExportsTheReportsErrors(t *testing.T) {
	t.Parallel()
	declared := twinOptions(t, twinFolder(t, twinBoard("zone-01", "thermal_sensor"), true))
	declared.Tolerances = []string{"zone-09.temperature=1"}
	tests := []struct {
		name    string
		options TwinReportOptions
		want    error
	}{
		{"no field trace", twinOptions(t, t.TempDir()), run.ErrFieldTraceMissing},
		{"no journal", twinOptions(t, twinFolder(t, twinBoard("zone-01", "thermal_sensor"), false)), ErrTwinNoJournal},
		{"no device", twinOptions(t, twinFolder(t, []map[string]any{{"type": "NSYNC", "through_seq": 1}}, true)), ErrTwinNoDevices},
		{"a kind the twin does not implement", twinOptions(t, twinFolder(t, twinBoard("pump-1", "pump"), true)), ErrTwinKinds},
		{"a tolerance that names nothing", declared, ErrTwinTolerance},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if _, err := TwinReport(t.Context(), tt.options); !errors.Is(err, tt.want) {
				t.Fatalf("TwinReport = %v, want %v", err, tt.want)
			}
		})
	}
}
