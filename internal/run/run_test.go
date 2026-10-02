package run

import (
	"context"
	"errors"
	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"strings"
	"testing"
	"time"
)

type failingSink struct{}

func (failingSink) Write([]byte) error { return errors.New("injected sink failure") }

func (failingSink) Close() ([]byte, error) { return nil, nil }

const (
	aquaculturePath = "../../docs/examples/aquaculture-pond.domain.json"
	nativeAdapter   = "../../adapters/native-jsonl.adapter.json"
	agenticAdapter  = "../../adapters/agentic-stream.adapter.json"
)

// testBase loads the aquaculture-pond domain and native-jsonl adapter.
func testBase(t *testing.T) (*domain.Compiled, *model.Adapter) {
	t.Helper()
	spec, err := domain.Load(aquaculturePath)
	if err != nil {
		t.Fatal(err)
	}
	a, err := adapter.Load(nativeAdapter)
	if err != nil {
		t.Fatal(err)
	}
	return spec, a
}

// TestRunByteReproducible is S1 gate 4: three runs byte-identical, and the
// artifact replays to a matching digest.
func TestRunByteReproducible(t *testing.T) {
	spec, a := testBase(t)
	start := model.DefaultStartTimeNS + 4*3600*1e9 // 04:00, pre-dawn
	cfg := func() Config {
		return Config{
			Domain: spec, Adapter: a, Seed: 42, SinkName: model.SinkInproc,
			TimeMode: model.TimeStepped, StartTimeNS: start,
		}
	}
	var digests []string
	var traces [][]byte
	for i := 0; i < 3; i++ {
		r, err := New(context.Background(), cfg())
		if err != nil {
			t.Fatal(err)
		}
		if _, err := r.Advance(context.Background(), start+6*3600*1e9, false); err != nil {
			t.Fatal(err)
		}
		art, err := r.End("")
		if err != nil {
			t.Fatal(err)
		}
		digests = append(digests, art.ExpectedTraceDigest)
		traces = append(traces, r.trace)
	}
	for i := 1; i < 3; i++ {
		if digests[i] != digests[0] {
			t.Fatalf("digest %d diverged: %s != %s", i, digests[i], digests[0])
		}
		if string(traces[i]) != string(traces[0]) {
			t.Fatalf("trace %d diverged byte-for-byte", i)
		}
	}

	// verify reproduces from the artifact.
	art := buildArtifact(t, cfg())
	res, err := ReplayArtifact(context.Background(), art, spec, a, "")
	if err != nil {
		t.Fatal(err)
	}
	if !res.Matches {
		t.Fatalf("replay mismatch: got %s want %s (first divergence %d)", res.GotDigest, res.WantDigest, res.FirstDivergence)
	}
}

func TestRunRejectsNonMonotonicNativeObservedTime(t *testing.T) {
	spec, a := testBase(t)
	start := model.DefaultStartTimeNS
	r, err := New(context.Background(), Config{
		Domain: spec, Adapter: a, Seed: 1, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start,
	})
	if err != nil {
		t.Fatal(err)
	}
	base := model.SimEvent{
		Seq: 0, WorldID: r.World.ID, EntityType: "fish", EntityID: r.World.InitialEntityIDs()[0],
		Channel: "temperature", EventTime: model.FormatTime(start),
	}
	first := base
	first.ObservedTime = model.FormatTime(start + 2)
	second := base
	second.Seq = 1
	second.ObservedTime = model.FormatTime(start + 1)
	r.onEmit(first)
	r.onEmit(second)
	_, err = r.End("")
	if err == nil || !strings.Contains(err.Error(), "strict observed-time guard") {
		t.Fatalf("run accepted non-monotonic native observed_time: %v", err)
	}
}

func TestReplayUsesEmbeddedDomainAndAdapter(t *testing.T) {
	spec, a := testBase(t)
	art := buildArtifact(t, Config{
		Domain: spec, Adapter: a, Seed: 123, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: model.DefaultStartTimeNS,
	})
	if len(art.DomainSpec) == 0 || len(art.AdapterSpec) == 0 {
		t.Fatal("run artifact did not embed its validated source documents")
	}
	res, err := ReplayArtifact(context.Background(), art, nil, nil, "")
	if err != nil {
		t.Fatalf("embedded replay failed: %v", err)
	}
	if !res.Matches {
		t.Fatalf("embedded replay mismatch: got=%s want=%s", res.GotDigest, res.WantDigest)
	}
}

func TestWorldDigestAndCommandTimesAreLossless(t *testing.T) {
	spec, a := testBase(t)
	start := int64(1<<60) + 123
	r, err := New(context.Background(), Config{
		Domain: spec, Adapter: a, Seed: ^uint64(0), SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start,
	})
	if err != nil {
		t.Fatal(err)
	}
	to := start + 7
	if _, err := r.Advance(context.Background(), to, false); err != nil {
		t.Fatal(err)
	}
	if got, ok := r.commandLog[0].Args["to_ns"].(int64); !ok || got != to {
		t.Fatalf("command time lost precision: type/value %T/%v", r.commandLog[0].Args["to_ns"], r.commandLog[0].Args["to_ns"])
	}
	if got := r.Digest(); got == "" || strings.HasPrefix(got, "invalid-world-digest:") {
		t.Fatalf("world digest is not usable: %q", got)
	}

	other, err := New(context.Background(), Config{
		Domain: spec, Adapter: a, Seed: ^uint64(0) - 1, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start,
	})
	if err != nil {
		t.Fatal(err)
	}
	if r.Digest() == other.Digest() {
		t.Fatal("adjacent uint64 seeds must produce distinct world digests")
	}
}

func TestQuiescenceWaitIsRaceFreeAndWakesOnReport(t *testing.T) {
	spec, a := testBase(t)
	start := model.DefaultStartTimeNS + 4*3600*1e9
	r, err := New(context.Background(), Config{
		Domain: spec, Adapter: a, Seed: 77, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start,
	})
	if err != nil {
		t.Fatal(err)
	}
	to := start + int64(time.Hour)
	// Park the waiter deterministically: the hook fires only when the wait
	// is about to block, so the report lands while the waiter is parked —
	// no wall-clock sleeps.
	parked := make(chan struct{})
	r.SetQuiesceParkedHook(func() { parked <- struct{}{} })
	done := make(chan error, 1)
	go func() {
		_, err := r.Advance(context.Background(), to, true)
		done <- err
	}()
	<-parked
	r.ReportQuiesced(to)
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("quiescence wait did not wake")
	}
}

func TestRunViewsAreDefensiveCopies(t *testing.T) {
	r := &Run{
		trace:   []byte("trace"),
		ledger:  []model.LedgerRecord{{DeliveryID: 1, Seq: 2}},
		history: []stateSnapshot{{States: map[string]float64{"x": 1}}},
		verdict: &model.Verdict{RunID: "r", Counters: map[string]int64{"records_seen": 1}},
	}
	trace := r.Trace()
	trace[0] = 'X'
	ledger := r.Ledger()
	ledger[0].Seq = 99
	history := r.History()
	history[0].States["x"] = 99
	verdict := r.Verdict()
	verdict.Counters["records_seen"] = 99
	if string(r.trace) != "trace" || r.ledger[0].Seq != 2 || r.history[0].States["x"] != 1 || r.verdict.Counters["records_seen"] != 1 {
		t.Fatal("run exposed mutable internal state")
	}
}
