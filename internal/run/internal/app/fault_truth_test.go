package app

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

const pondFault = "aerator_failure"

func faultTruthRecords(t *testing.T, dir string) [][]model.FaultInterval {
	t.Helper()
	director, err := decodeJournalBytes(readFile(t, filepath.Join(dir, journal.DirectorFile)))
	if err != nil {
		t.Fatal(err)
	}
	var snapshots [][]model.FaultInterval
	for _, record := range director {
		if record.Type != journal.TypeFaultTruth {
			continue
		}
		var faults []model.FaultInterval
		if err := decodeInto(record.Fields["faults"], &faults); err != nil {
			t.Fatal(err)
		}
		snapshots = append(snapshots, faults)
	}
	return snapshots
}

func newJournaledPond(t *testing.T) (*Run, string) {
	t.Helper()
	cfg := pondConfig(t)
	cfg.JournalDir = t.TempDir()
	r, err := New(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	return r, cfg.JournalDir
}

func wantEnd(t *testing.T, fault model.FaultInterval, cleared int64, by string) {
	t.Helper()
	if fault.ClearedNS == nil || *fault.ClearedNS != cleared || fault.ClearedBy == nil || *fault.ClearedBy != by {
		t.Fatalf("fault = %+v, want cleared at %d by %s", fault, cleared, by)
	}
}

func TestTheDirectorJournalRecordsAFaultIntervalAtEachFaultCommandAndAClearAtItsAtNS(t *testing.T) {
	t.Parallel()
	r, dir := newJournaledPond(t)
	start := model.DefaultStartTimeNS
	id, err := r.InjectFault("site-a/pond-1", pondFault, start+60e9, nil)
	if err != nil {
		t.Fatal(err)
	}
	advanceAndReport(t, r, start+120e9)
	if err := r.ClearFault(id, start+300e9); err != nil {
		t.Fatal(err)
	}
	advanceAndReport(t, r, start+400e9)
	if _, err := r.End(""); err != nil {
		t.Fatal(err)
	}
	snapshots := faultTruthRecords(t, dir)
	if len(snapshots) != 2 {
		t.Fatalf("%d fault.truth records, want one per fault command and none extra at the close", len(snapshots))
	}
	injected, cleared := snapshots[0][0], snapshots[1][0]
	if injected.FaultID != id || injected.EntityID != "site-a/pond-1" || injected.Fault != pondFault ||
		injected.InjectedAtNS != start || injected.OnsetNS != start+60e9 || injected.ClearedNS != nil || injected.ClearedBy != nil {
		t.Fatalf("after the inject: %+v, want an open interval", injected)
	}
	wantEnd(t, cleared, start+300e9, model.ClearedByCommand)
}

func TestAClearScheduledPastTheEndOfTheRunLeavesTheFaultOpenInTheFinalRecord(t *testing.T) {
	t.Parallel()
	r, dir := newJournaledPond(t)
	start := model.DefaultStartTimeNS
	id, err := r.InjectFault("site-a/pond-1", pondFault, start+60e9, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.ClearFault(id, start+900e9); err != nil {
		t.Fatal(err)
	}
	advanceAndReport(t, r, start+400e9)
	if _, err := r.End(""); err != nil {
		t.Fatal(err)
	}
	snapshots := faultTruthRecords(t, dir)
	if len(snapshots) != 3 {
		t.Fatalf("%d fault.truth records, want inject, clear and the final one", len(snapshots))
	}
	wantEnd(t, snapshots[1][0], start+900e9, model.ClearedByCommand)
	if last := snapshots[2][0]; last.ClearedNS != nil || last.ClearedBy != nil {
		t.Fatalf("final = %+v, want the fault still active at the end of the run", last)
	}
}

func TestAPulseFaultRecordsItsOwnEndAndARunWithoutFaultsRecordsNone(t *testing.T) {
	t.Parallel()
	spec, err := domain.Load(filepath.Join(testsupport.RepositoryRoot(), "examples", "factory-utilities.domain.json"))
	if err != nil {
		t.Fatal(err)
	}
	native, err := adapter.Load(testsupport.Adapter("native-jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	dir, start := t.TempDir(), model.DefaultStartTimeNS
	r, err := New(context.Background(), Config{Domain: spec, Adapter: native, Seed: 5, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start, StartTimeSet: true, JournalDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	advanceAndReport(t, r, start+60e9)
	if got := len(faultTruthRecords(t, dir)); got != 0 {
		t.Fatalf("%d fault.truth records before any fault", got)
	}
	calendar := calendarEntity(t, r)
	if _, err := r.InjectFault(calendar, "site_calendar__overtime", start+90e9, nil); err != nil {
		t.Fatal(err)
	}
	advanceAndReport(t, r, start+600e9)
	if _, err := r.End(""); err != nil {
		t.Fatal(err)
	}
	snapshots := faultTruthRecords(t, dir)
	if len(snapshots) != 2 {
		t.Fatalf("%d fault.truth records, want the inject and the final one", len(snapshots))
	}
	wantEnd(t, snapshots[0][0], start+90e9+14400e9, model.ClearedByPulseEnd)
	if last := snapshots[1][0]; last.ClearedNS != nil {
		t.Fatalf("final = %+v, want the four-hour pulse still active ten minutes in", last)
	}
}

func calendarEntity(t *testing.T, r *Run) string {
	t.Helper()
	for _, id := range r.World.EntityIDs() {
		if strings.Contains(id, "CAL-1") {
			return id
		}
	}
	t.Fatal("the factory-utilities example has no CAL-1 entity")
	return ""
}

func decodeInto(from, to any) error {
	raw, err := json.Marshal(from)
	if err != nil {
		return fmt.Errorf("encode: %w", err)
	}
	if err := json.Unmarshal(raw, to); err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	return nil
}
