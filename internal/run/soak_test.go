package run

// Slice G: the soak target and the emission-path benchmarks. The soak test
// is gated behind SOAK=1 (make soak): a deterministic run of >1,000,000
// delivered records asserting conservation and replay identity, bounded to
// 15 minutes. Benchmarks profile the hot paths for the perf target.

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// TestSoakConservationAtScale drives the full pipeline past a million
// delivered records and asserts the conservation identity plus replay
// determinism under load. Gated behind SOAK=1 so unit runs stay fast.
func TestSoakConservationAtScale(t *testing.T) {
	if os.Getenv("SOAK") != "1" {
		t.Skip("soak run requires SOAK=1 (make soak)")
	}
	spec, a := testBase(t)
	for i := range spec.Spec.Channels {
		if spec.Spec.Channels[i].Name == "pond.heartbeat" {
			spec.Spec.Channels[i].Cadence.PeriodS = 0.001
		}
	}
	start := model.DefaultStartTimeNS
	r, err := New(context.Background(), Config{
		Domain: spec, Adapter: a, Seed: 3, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start, StartTimeSet: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.ApplyPerturb("drop", map[string]any{"rate": 0.1}, 0, 0); err != nil {
		t.Fatal(err)
	}
	// 150 seconds at 1ms heartbeat across 8 ponds > 1,000,000 records
	// (8 x 150 x 1000 = 1.2M), bounded so the soak stays a soak.
	if _, err := r.Advance(context.Background(), start+150*1e9, false); err != nil {
		t.Fatal(err)
	}
	emitted := r.World.EmittedCount()
	ledger := r.Ledger()
	if emitted < 1_000_000 {
		t.Fatalf("soak must exceed 1,000,000 delivered records, got %d", emitted)
	}
	// Conservation: exactly one primary row per seq, duplicates as surplus.
	seenIDs := map[uint64]bool{}
	primary := map[int64]int{}
	for _, row := range ledger {
		if row.DeliveryID == 0 || seenIDs[row.DeliveryID] {
			t.Fatalf("delivery id broken at %d", row.DeliveryID)
		}
		seenIDs[row.DeliveryID] = true
		if row.DeliveryReason == model.DeliveryDuplicated {
			continue
		}
		primary[row.Seq]++
	}
	if int64(len(primary)) != emitted {
		t.Fatalf("conservation broke at scale: %d distinct seqs vs %d emitted", len(primary), emitted)
	}
	for seq := int64(0); seq < emitted; seq++ {
		if primary[seq] != 1 {
			t.Fatalf("seq %d has %d primary rows", seq, primary[seq])
		}
	}
	// Replay identity under load: the artifact reproduces byte-for-byte.
	dir := t.TempDir()
	if _, err := r.End(dir); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadArtifact(dir + "/run.json")
	if err != nil {
		t.Fatal(err)
	}
	res, err := ReplayArtifact(context.Background(), loaded, spec, a, "")
	if err != nil {
		t.Fatal(err)
	}
	if !res.Matches {
		t.Fatalf("replay diverged at scale: %s vs %s", res.GotDigest, res.WantDigest)
	}
}

// BenchmarkEmissionPath profiles the world emission pipeline.
func BenchmarkEmissionPath(b *testing.B) {
	spec, err := domain.Load(aquaculturePath)
	if err != nil {
		b.Fatal(err)
	}
	a, err := adapter.Load(nativeAdapter)
	if err != nil {
		b.Fatal(err)
	}
	start := model.DefaultStartTimeNS
	r, err := New(context.Background(), Config{
		Domain: spec, Adapter: a, Seed: 9, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start, StartTimeSet: true,
	})
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := r.Advance(context.Background(), start+int64(i+1)*int64(time.Minute), false); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkLedgerAppend profiles the durable ledger write path.
func BenchmarkLedgerAppend(b *testing.B) {
	spec, err := domain.Load(aquaculturePath)
	if err != nil {
		b.Fatal(err)
	}
	a, err := adapter.Load(nativeAdapter)
	if err != nil {
		b.Fatal(err)
	}
	dir := b.TempDir()
	start := model.DefaultStartTimeNS
	r, err := New(context.Background(), Config{
		Domain: spec, Adapter: a, Seed: 9, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start, StartTimeSet: true,
		LedgerPath: dir + "/ledger.jsonl",
	})
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := r.Advance(context.Background(), start+int64(i+1)*int64(time.Minute), false); err != nil {
			b.Fatal(err)
		}
	}
}
