package app

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/process"
)

const divergingReport = `{"schema_version":"0.1","verdict":"diverges","summary":{"divergences":1},"metrics":[],"divergences":[{"name":"zone-01.temperature#1"}]}`

func TestRunTwinReportRunsTheBinaryOnTheRunWithTheWorkspaceDomainsAndKinds(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.command = process.CommandResult{Stdout: []byte(divergingReport)}
	args := map[string]any{"run": "w-1", "domain": "arduino-thermal", "tolerance": "zone-01.temperature=2,thermal_sensor.humidity=5", "span_fraction": 0.1}
	if _, err := h.bridge.RunAction(context.Background(), "run.twin_report", args); err != nil {
		t.Fatal(err)
	}
	want := []string{"plantwire", "twin-report", filepath.Join(h.dir, "runs", "w-1"), "--domain", "arduino-thermal",
		"--tolerance", "zone-01.temperature=2,thermal_sensor.humidity=5", "--span-fraction", "0.1",
		"--domains-dir", h.domainsDir(), "--kinds-dir", filepath.Join(h.dir, "kinds")}
	if !slices.Equal(h.argv[0], want) || h.stdin[0] != nil {
		t.Fatalf("argv %v stdin %q, want %v and no input", h.argv[0], h.stdin[0], want)
	}
	if slices.Contains(h.argv[0], "--adapters-dir") {
		t.Fatalf("argv %v passes the adapters directory, which a twin report does not read", h.argv[0])
	}
}

func TestRunTwinReportPassesTheKindsFolderEvenBeforeItExists(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	if h.bridge.edges.Workspace.Exists("kinds") {
		t.Fatal("the harness workspace already has a kinds folder")
	}
	if _, err := h.bridge.RunAction(context.Background(), "run.twin_report", map[string]any{"run": "w-1", "domain": "arduino-thermal"}); err != nil {
		t.Fatal(err)
	}
	want := []string{"plantwire", "twin-report", filepath.Join(h.dir, "runs", "w-1"), "--domain", "arduino-thermal",
		"--domains-dir", h.domainsDir(), "--kinds-dir", filepath.Join(h.dir, "kinds")}
	if !slices.Equal(h.argv[0], want) {
		t.Fatalf("argv %v, want %v", h.argv[0], want)
	}
}

func TestRunTwinReportReturnsTheDocumentUnchangedWhateverItsVerdict(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.command = process.CommandResult{Stdout: []byte(divergingReport)}
	result, err := h.bridge.RunAction(context.Background(), "run.twin_report", map[string]any{"run": "w-1", "domain": "arduino-thermal"})
	if err != nil {
		t.Fatalf("a diverging report is a result, not an error: %v", err)
	}
	var got, want any
	if err := json.Unmarshal(result, &got); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(divergingReport), &want); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("result %s, want the binary's document %s", result, divergingReport)
	}
}

func TestRunTwinReportRelaysTheBinarysFailure(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	h.command = process.CommandResult{ExitCode: 1, Stdout: []byte(`{"error":{"code":"plantwire_failed","message":"no field trace in the run folder"}}`)}
	var failure *domain.Failure
	_, err := h.bridge.RunAction(context.Background(), "run.twin_report", map[string]any{"run": "w-1", "domain": "arduino-thermal"})
	if !errors.As(err, &failure) || failure.Code != "plantwire_failed" {
		t.Fatalf("err = %v, want the binary's plantwire_failed relayed", err)
	}
	h.command = process.CommandResult{ExitCode: 1, Stdout: []byte("not json"), Stderr: []byte("boom")}
	if _, err := h.bridge.RunAction(context.Background(), "run.twin_report", map[string]any{"run": "w-1", "domain": "arduino-thermal"}); !errors.As(err, &failure) ||
		failure.Code != "command_failed" {
		t.Fatalf("err = %v, want command_failed for a command without a document", err)
	}
}

func TestRunTwinReportRefusesArgumentsTheActionDoesNotTake(t *testing.T) {
	t.Parallel()
	h := newHarness(t)
	tests := map[string]map[string]any{
		"a missing run":     {"domain": "arduino-thermal"},
		"a missing domain":  {"run": "w-1"},
		"a run that climbs": {"run": "..", "domain": "arduino-thermal"},
		"a flag as domain":  {"run": "w-1", "domain": "--kinds-dir=/"},
		"an unknown key":    {"run": "w-1", "domain": "arduino-thermal", "kinds_dir": "/"},
		"text as the span":  {"run": "w-1", "domain": "arduino-thermal", "span_fraction": "0.1; rm"},
	}
	for name, args := range tests {
		if _, err := h.bridge.RunAction(context.Background(), "run.twin_report", args); !errors.Is(err, domain.ErrBadArguments) {
			t.Errorf("%s: err = %v, want ErrBadArguments", name, err)
		}
	}
	if len(h.argv) != 0 {
		t.Fatalf("argv %v, want no command to have run", h.argv)
	}
}
