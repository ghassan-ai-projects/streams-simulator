package run

// Slice C (G4): the delivery ledger is append-only and durable at command
// boundaries. A crash leaves the recoverable delivered prefix — the trace
// file is consistent with the recovered ledger — and conservation holds
// beyond 10,000 records with drops and duplicates in play.

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// loadLedgerFile reads rows from a JSONL ledger file.
func loadLedgerFile(t *testing.T, path string) []model.LedgerRecord {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var rows []model.LedgerRecord
	for _, line := range bytes.Split(raw, []byte("\n")) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var rec model.LedgerRecord
		if err := json.Unmarshal(line, &rec); err != nil {
			t.Fatalf("unmarshal ledger row: %v", err)
		}
		rows = append(rows, rec)
	}
	return rows
}

// TestDurableLedgerSurvivesCrash: rows reach the ledger file at each command
// boundary, so an abandoned run leaves a recoverable delivered prefix with
// no artifact, and the recovered trace is consistent with it. The run can
// resume; End fsyncs and produces the artifact.
func TestDurableLedgerSurvivesCrash(t *testing.T) {
	spec, a := testBase(t)
	dir := t.TempDir()
	start := model.DefaultStartTimeNS + 4*3600*1e9
	r, err := New(context.Background(), Config{
		Domain: spec, Adapter: a, Seed: 3, SinkName: model.SinkFile,
		SinkTarget: filepath.Join(dir, "trace.jsonl"),
		TimeMode:   model.TimeStepped, StartTimeNS: start, StartTimeSet: true,
		LedgerPath: filepath.Join(dir, "ledger.jsonl"),
	})
	if err != nil {
		t.Fatal(err)
	}
	ledgerPath := filepath.Join(dir, "ledger.jsonl")
	tracePath := filepath.Join(dir, "trace.jsonl")

	// Boundary 1: the advance completes; rows and trace bytes are on file
	// descriptors.
	if _, err := r.Advance(context.Background(), start+1*3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	rows := loadLedgerFile(t, ledgerPath)
	if len(rows) == 0 {
		t.Fatal("no ledger rows after the first boundary")
	}
	if len(rows) != len(r.Ledger()) {
		t.Fatalf("recovered rows %d != in-memory ledger %d", len(rows), len(r.Ledger()))
	}

	// "Crash": the process dies without End. No artifact exists; the ledger
	// and trace files are the recoverable delivered prefix.
	if _, err := os.Stat(filepath.Join(dir, "run.json")); !os.IsNotExist(err) {
		t.Fatal("a crashed run must leave no artifact")
	}
	delivered := 0
	for _, row := range rows {
		if row.Delivered {
			delivered++
		}
	}
	traceBytes, err := os.ReadFile(tracePath)
	if err != nil {
		t.Fatal(err)
	}
	if lines := bytes.Count(traceBytes, []byte("\n")); lines != delivered {
		t.Fatalf("recovered trace %d lines != recovered ledger %d delivered rows", lines, delivered)
	}

	// The run resumes: later rows append to the same file.
	if _, err := r.Advance(context.Background(), start+2*3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	if rows = loadLedgerFile(t, ledgerPath); len(rows) != len(r.Ledger()) {
		t.Fatalf("appended rows %d != in-memory ledger %d", len(rows), len(r.Ledger()))
	}

	// End fsyncs the durable ledger and writes the artifact. The resumed
	// run ends normally, so the artifact is complete.
	if art, err := r.End(dir); err != nil {
		t.Fatal(err)
	} else if art.Incomplete {
		t.Fatal("a resumed run that ended cleanly must be complete")
	}
	if rows = loadLedgerFile(t, ledgerPath); len(rows) != len(r.Ledger()) {
		t.Fatalf("final rows %d != in-memory ledger %d", len(rows), len(r.Ledger()))
	}
	if _, err := os.Stat(filepath.Join(dir, "run.json")); err != nil {
		t.Fatalf("artifact missing after End: %v", err)
	}
}

// TestConservationAtScaleBeyondTenThousand: with drops and duplicates in
// play, every emitted seq has exactly one primary row, duplicated rows are
// surplus, delivery ids are unique and non-zero, and the identity holds
// beyond 10,000 records — the G4 evidence the in-memory ledger never had.
func TestConservationAtScaleBeyondTenThousand(t *testing.T) {
	spec, a := testBase(t)
	// A 1ms heartbeat produces >10,000 deliveries in a short stepped run
	// (the same recipe as TestHistoryIsNotSilentlyCapped).
	found := false
	for i := range spec.Spec.Channels {
		if spec.Spec.Channels[i].Name == "pond.heartbeat" {
			spec.Spec.Channels[i].Cadence.PeriodS = 0.001
			found = true
		}
	}
	if !found {
		t.Fatal("heartbeat channel missing from fixture")
	}
	start := model.DefaultStartTimeNS
	r, err := New(context.Background(), Config{
		Domain: spec, Adapter: a, Seed: 9, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start, StartTimeSet: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.ApplyPerturb("drop", map[string]any{"rate": 0.2}, 0, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := r.ApplyPerturb("duplicate_burst", map[string]any{"rate": 0.1}, 0, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(context.Background(), start+11*1e9, false); err != nil {
		t.Fatal(err)
	}
	emitted := r.World.EmittedCount()
	ledger := r.Ledger()
	if len(ledger) <= 10000 {
		t.Fatalf("need >10,000 delivery instances for the G4 evidence, got %d", len(ledger))
	}
	if int64(len(ledger)) < emitted {
		t.Fatalf("ledger %d rows cannot be less than %d emitted", len(ledger), emitted)
	}
	seenIDs := map[uint64]bool{}
	primary := map[int64]int{}
	duplicated := int64(0)
	for _, row := range ledger {
		if row.DeliveryID == 0 || seenIDs[row.DeliveryID] {
			t.Fatalf("delivery ids must be unique and non-zero (got %d)", row.DeliveryID)
		}
		seenIDs[row.DeliveryID] = true
		if row.DeliveryReason == model.DeliveryDuplicated {
			duplicated++
			continue
		}
		primary[row.Seq]++
	}
	if int64(len(primary)) != emitted {
		t.Fatalf("emitted %d events but %d distinct primary seqs", emitted, len(primary))
	}
	for seq := int64(0); seq < emitted; seq++ {
		if n := primary[seq]; n != 1 {
			t.Fatalf("seq %d has %d primary rows; every event needs exactly one", seq, n)
		}
	}
	if duplicated != int64(len(ledger))-emitted {
		t.Fatalf("duplicated rows are surplus: got %d surplus, expected %d duplicated", int64(len(ledger))-emitted, duplicated)
	}
}
