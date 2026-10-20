package app

import (
	"context"
	"errors"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	simdomain "github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/plantwire/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func twinFolderWithTrace(t *testing.T, field []map[string]any) string {
	t.Helper()
	dir := t.TempDir()
	writeFieldTrace(t, dir, field)
	writeHostJournal(t, dir, nil)
	return dir
}

func TestAFolderWithoutAFieldTraceIsRefused(t *testing.T) {
	t.Parallel()
	_, err := TwinReport(t.Context(), twinOptions(t, t.TempDir()))
	if !errors.Is(err, run.ErrFieldTraceMissing) {
		t.Fatalf("TwinReport = %v, want run.ErrFieldTraceMissing", err)
	}
}

func TestAFolderWithoutAJournalIsRefused(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFieldTrace(t, dir, twinBirths())
	if _, err := TwinReport(t.Context(), twinOptions(t, dir)); !errors.Is(err, ErrTwinNoJournal) {
		t.Fatalf("TwinReport = %v, want ErrTwinNoJournal", err)
	}
}

func TestAJournalWithAMalformedCommandIsRefused(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFieldTrace(t, dir, twinBirths())
	testsupport.WriteFileIn(t, dir, "journal.jsonl", []byte(`{"type":"wire","direction":"host","record":{"type":"DCMD"`+"\n"))
	if _, err := TwinReport(t.Context(), twinOptions(t, dir)); !errors.Is(err, domain.ErrTwinJournal) {
		t.Fatalf("TwinReport = %v, want domain.ErrTwinJournal", err)
	}
}

func TestATraceThatAnnouncesNoDeviceIsRefused(t *testing.T) {
	t.Parallel()
	dir := twinFolderWithTrace(t, []map[string]any{{"type": "NSYNC", "through_seq": 1}})
	if _, err := TwinReport(t.Context(), twinOptions(t, dir)); !errors.Is(err, domain.ErrTwinNoDevices) {
		t.Fatalf("TwinReport = %v, want domain.ErrTwinNoDevices", err)
	}
}

func TestATwinThatImplementsNoKindOfTheBoardIsRefused(t *testing.T) {
	t.Parallel()
	pump := map[string]any{"type": "DBIRTH", "device": "pump-1", "kind": "pump", "board": "mega-01", "birth_time_ns": twinStartNS,
		"metrics": []any{map[string]any{"name": "flow", "alias": 1, "value": 1.0, "quality": "good"}}}
	_, err := TwinReport(t.Context(), twinOptions(t, twinFolderWithTrace(t, []map[string]any{pump})))
	if !errors.Is(err, domain.ErrTwinKinds) || !strings.Contains(err.Error(), "board kinds pump") {
		t.Fatalf("TwinReport = %v, want domain.ErrTwinKinds naming the pump kind", err)
	}
}

func TestATraceSpanningMoreThanAWeekIsRefused(t *testing.T) {
	t.Parallel()
	field := append(twinBirths(), twinTick(twinStartNS, twinOpenTemp), twinTick(twinStartNS+8*24*3600*twinTickNS, twinOpenTemp))
	_, err := TwinReport(t.Context(), twinOptions(t, twinFolderWithTrace(t, field)))
	if !errors.Is(err, domain.ErrTwinSpan) {
		t.Fatalf("TwinReport = %v, want domain.ErrTwinSpan", err)
	}
}

func driftedThermalKinds(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	names, err := os.ReadDir(testsupport.ThermalBoardKindsDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range names {
		if !strings.HasSuffix(entry.Name(), ".kind.json") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(testsupport.ThermalBoardKindsDir(), entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		content := string(raw)
		if entry.Name() == "thermal_sensor.kind.json" {
			content = strings.Replace(content, `"Cel"`, `"degC"`, 1)
		}
		testsupport.WriteFileIn(t, dir, entry.Name(), []byte(content))
	}
	return dir
}

func TestAKindFileThePlantDriftedFromIsRefused(t *testing.T) {
	t.Parallel()
	options := twinOptions(t, agreeingFolder(t))
	options.KindsDirs = []string{driftedThermalKinds(t)}
	_, err := TwinReport(t.Context(), options)
	if _, validation := model.AsValidationError(err); !errors.Is(err, domain.ErrKindDrift) || !validation {
		t.Fatalf("TwinReport = %v, want a validation error caused by domain.ErrKindDrift", err)
	}
}

func TestAMissingKindsDirectoryIsRefused(t *testing.T) {
	t.Parallel()
	options := twinOptions(t, agreeingFolder(t))
	options.KindsDirs = []string{filepath.Join(t.TempDir(), "absent")}
	if _, err := TwinReport(t.Context(), options); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("TwinReport = %v, want fs.ErrNotExist for the absent kinds directory", err)
	}
}

func TestAnUnknownDomainIsRefused(t *testing.T) {
	t.Parallel()
	options := twinOptions(t, agreeingFolder(t))
	options.Domain = "no-such-twin"
	if _, err := TwinReport(t.Context(), options); !errors.Is(err, simdomain.ErrUnknownDomain) {
		t.Fatalf("TwinReport = %v, want domain.ErrUnknownDomain", err)
	}
}

func TestTolerancesThatCannotBeKeptAreRefused(t *testing.T) {
	t.Parallel()
	tests := map[string]func(*TwinReportOptions){
		"a key that names no metric of the board": func(o *TwinReportOptions) { o.Tolerances = []string{"zone-09.temperature=1"} },
		"a negative value":                        func(o *TwinReportOptions) { o.Tolerances = []string{"zone-01.temperature=-1"} },
		"a declaration without a value":           func(o *TwinReportOptions) { o.Tolerances = []string{"zone-01.temperature"} },
		"a span fraction above one":               func(o *TwinReportOptions) { o.SpanFraction = 2 },
	}
	for name, change := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			options := twinOptions(t, agreeingFolder(t))
			change(&options)
			if _, err := TwinReport(t.Context(), options); !errors.Is(err, domain.ErrTwinTolerance) {
				t.Fatalf("TwinReport = %v, want domain.ErrTwinTolerance", err)
			}
		})
	}
}

func TestDeclaredTolerancesAndASpanFractionReachTheReport(t *testing.T) {
	t.Parallel()
	options := twinOptions(t, agreeingFolder(t))
	options.Tolerances = []string{"zone-01.temperature=4,thermal_sensor.humidity=9"}
	options.SpanFraction = 0.1
	report, err := TwinReport(t.Context(), options)
	if err != nil {
		t.Fatalf("TwinReport: %v", err)
	}
	if tol := twinMetric(t, report, "zone-01", "temperature").Tolerance; tol.Source != domain.ToleranceDeclared || *tol.Value != 4 {
		t.Fatalf("temperature tolerance = %+v, want the declared 4", tol)
	}
	if tol := twinMetric(t, report, "zone-01", "humidity").Tolerance; tol.Source != domain.ToleranceDeclared || *tol.Value != 9 {
		t.Fatalf("humidity tolerance = %+v, want the kind's declared 9", tol)
	}
	if tol := twinMetric(t, report, "fan-01", "level_percent").Tolerance; tol.Source != domain.ToleranceDefaultSpan || math.Abs(*tol.Value-6) > 1e-9 || report.ToleranceRule.SpanFraction != 0.1 {
		t.Fatalf("duty tolerance = %+v rule %+v, want 0.1 of the 60 span", tol, report.ToleranceRule)
	}
}

func TestMissingOptionsAreRefused(t *testing.T) {
	t.Parallel()
	whole := twinOptions(t, agreeingFolder(t))
	tests := map[string]func(*TwinReportOptions){
		"no run folder":         func(o *TwinReportOptions) { o.Run = "" },
		"no domains directory":  func(o *TwinReportOptions) { o.DomainsDir = "" },
		"no domain":             func(o *TwinReportOptions) { o.Domain = "" },
		"no kinds directory":    func(o *TwinReportOptions) { o.KindsDirs = nil },
		"every option missing":  func(o *TwinReportOptions) { *o = TwinReportOptions{} },
		"only a run is present": func(o *TwinReportOptions) { *o = TwinReportOptions{Run: o.Run} },
	}
	for name, change := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			options := whole
			change(&options)
			if _, err := TwinReport(t.Context(), options); !errors.Is(err, ErrBadOptions) {
				t.Fatalf("TwinReport = %v, want ErrBadOptions", err)
			}
		})
	}
}

func TestACancelledContextStopsTheReport(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := TwinReport(ctx, twinOptions(t, agreeingFolder(t))); !errors.Is(err, context.Canceled) {
		t.Fatalf("TwinReport = %v, want context.Canceled", err)
	}
}

func TestAFinishedRunWhoseEvidenceMatchesItsArtifactIsReported(t *testing.T) {
	t.Parallel()
	dir := twinArtifactFolder(t, twinBirths(), true)
	report, err := TwinReport(t.Context(), twinOptions(t, dir))
	if err != nil {
		t.Fatalf("TwinReport: %v", err)
	}
	art, err := run.LoadArtifact(filepath.Join(dir, model.FileRunArtifact))
	if err != nil || report.Run.FieldTraceDigest != art.FieldTraceDigest || report.Run.JournalDigest != art.JournalFileDigest {
		t.Fatalf("run evidence %+v (%v), artifact digests %s and %s, want the report to name the digests the run recorded",
			report.Run, err, art.FieldTraceDigest, art.JournalFileDigest)
	}
	if report.Verdict != domain.VerdictCannotJudge || report.Reason != domain.ReasonNoCommands {
		t.Fatalf("verdict %s reason %q, want cannot_judge: the run's journal holds no host command", report.Verdict, report.Reason)
	}
}

func TestAFieldTraceThatDiffersFromItsArtifactIsRefused(t *testing.T) {
	t.Parallel()
	dir := twinArtifactFolder(t, twinBirths(), true)
	if err := os.WriteFile(filepath.Join(dir, model.FileFieldTrace), []byte("tampered\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := TwinReport(t.Context(), twinOptions(t, dir)); !errors.Is(err, run.ErrFieldTraceDigest) {
		t.Fatalf("TwinReport = %v, want run.ErrFieldTraceDigest", err)
	}
}

func TestAFieldTraceTheArtifactDoesNotRecordIsRefused(t *testing.T) {
	t.Parallel()
	dir := twinArtifactFolder(t, twinBirths(), false)
	if _, err := TwinReport(t.Context(), twinOptions(t, dir)); !errors.Is(err, run.ErrFieldTraceUncovered) {
		t.Fatalf("TwinReport = %v, want run.ErrFieldTraceUncovered", err)
	}
}

func TestAJournalThatDiffersFromItsArtifactIsRefused(t *testing.T) {
	t.Parallel()
	dir := twinArtifactFolder(t, twinBirths(), true)
	journal, err := os.OpenFile(filepath.Join(dir, model.FileObserverJournal), os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := journal.WriteString("\n"); err != nil {
		t.Fatal(err)
	}
	if err := journal.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := TwinReport(t.Context(), twinOptions(t, dir)); !errors.Is(err, ErrTwinJournalDigest) {
		t.Fatalf("TwinReport = %v, want ErrTwinJournalDigest", err)
	}
}

func TestAnInterruptLogTheArtifactDoesNotRecordIsRefused(t *testing.T) {
	t.Parallel()
	dir := twinArtifactFolder(t, twinBirths(), true)
	writeInterrupts(t, dir, heldFanLog(t, "c-1"))
	if _, err := TwinReport(t.Context(), twinOptions(t, dir)); !errors.Is(err, run.ErrInterruptLogUncovered) {
		t.Fatalf("TwinReport = %v, want run.ErrInterruptLogUncovered", err)
	}
}
