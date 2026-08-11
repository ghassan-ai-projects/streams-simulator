package run

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

type failingSink struct{}

func (failingSink) Write([]byte) error     { return errors.New("injected sink failure") }
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
		if _, err := r.Advance(start+6*3600*1e9, false); err != nil {
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
	if _, err := r.Advance(to, false); err != nil {
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
	done := make(chan error, 1)
	go func() {
		_, err := r.Advance(to, true)
		done <- err
	}()
	// Report from the consumer side while Advance is waiting. The wait must
	// wake immediately without polling a shared unsynchronized field.
	time.Sleep(10 * time.Millisecond)
	r.ReportQuiesced(to)
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("quiescence wait did not wake")
	}
}

// buildArtifact runs the config and returns its artifact.
func buildArtifact(t *testing.T, cfg Config) *model.RunArtifact {
	t.Helper()
	r, err := New(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(cfg.StartTimeNS+6*3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	art, err := r.End("")
	if err != nil {
		t.Fatal(err)
	}
	return art
}

// TestRunSinkEquivalence: inproc and file produce byte-identical traces.
func TestRunSinkEquivalence(t *testing.T) {
	spec, a := testBase(t)
	start := model.DefaultStartTimeNS + 4*3600*1e9
	run := func(sink, target string) []byte {
		r, err := New(context.Background(), Config{
			Domain: spec, Adapter: a, Seed: 7, SinkName: sink,
			SinkTarget: target, TimeMode: model.TimeStepped, StartTimeNS: start,
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := r.Advance(start+2*3600*1e9, false); err != nil {
			t.Fatal(err)
		}
		art, err := r.End("")
		if err != nil {
			t.Fatal(err)
		}
		return []byte(art.ExpectedTraceDigest + "|" + string(r.trace))
	}
	inproc := run(model.SinkInproc, "")
	file := run(model.SinkFile, filepath.Join(t.TempDir(), "trace.jsonl"))
	if string(inproc) != string(file) {
		t.Fatalf("sink equivalence broken: inproc and file diverged")
	}
}

func TestStreamingRunIncludesAdapterPreambleAndPostamble(t *testing.T) {
	spec, err := domain.Load(aquaculturePath)
	if err != nil {
		t.Fatal(err)
	}
	a, err := adapter.Load(agenticAdapter)
	if err != nil {
		t.Fatal(err)
	}
	start := model.DefaultStartTimeNS + 4*3600*1e9
	r, err := New(context.Background(), Config{
		Domain: spec, Adapter: a, Seed: 11, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(start+3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	if _, err := r.End(""); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(r.Trace())), "\n")
	if len(lines) < 3 {
		t.Fatalf("expected preamble, event, and postamble; got %d lines", len(lines))
	}
	if !strings.Contains(lines[0], `"record_type":"runtime_config"`) {
		t.Fatalf("missing runtime preamble: %s", lines[0])
	}
	if !strings.Contains(lines[len(lines)-1], `"record_type":"trace_end"`) {
		t.Fatalf("missing trace postamble: %s", lines[len(lines)-1])
	}
}

func TestSinkFailureMarksRunIncompleteWithoutPanic(t *testing.T) {
	spec, a := testBase(t)
	start := model.DefaultStartTimeNS + 4*3600*1e9
	r, err := New(context.Background(), Config{
		Domain: spec, Adapter: a, Seed: 12, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start,
	})
	if err != nil {
		t.Fatal(err)
	}
	r.Sink = failingSink{}
	if _, err := r.Advance(start+3600*1e9, false); err == nil {
		t.Fatal("sink failure must be returned from Advance")
	}
	art, endErr := r.End("")
	if endErr == nil {
		t.Fatal("incomplete run must retain its failure")
	}
	if art == nil || !art.Incomplete || art.Error == "" {
		t.Fatalf("incomplete artifact missing failure state: %+v", art)
	}
}

// TestRunLedgerDistinguishesDrop: a dropped event is in the ledger as
// undelivered and absent from the trace.
func TestRunLedgerDistinguishesDrop(t *testing.T) {
	spec, a := testBase(t)
	start := model.DefaultStartTimeNS + 4*3600*1e9
	r, err := New(context.Background(), Config{
		Domain: spec, Adapter: a, Seed: 3, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.ApplyPerturb("drop", map[string]any{"rate": 1.0}, 0, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(start+1*3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	art, err := r.End("")
	if err != nil {
		t.Fatal(err)
	}
	if art.Counts.Emitted == 0 {
		t.Fatal("no events emitted")
	}
	ledger := r.Ledger()
	dropped := 0
	for _, l := range ledger {
		if l.Delivered {
			t.Fatalf("rate=1 drop must deliver nothing, but seq %d was delivered", l.Seq)
		}
		if l.DeliveryReason != model.DeliveryDroppedByPerturb {
			t.Fatalf("wrong drop reason: %s", l.DeliveryReason)
		}
		dropped++
	}
	if dropped != len(ledger) {
		t.Fatalf("ledger/drop mismatch")
	}
	if len(strings.TrimSpace(string(r.trace))) != 0 {
		t.Fatalf("trace must be empty under rate=1 drop, got %q", r.trace)
	}
}

func TestLedgerDeliveryIDsAreUniqueAcrossDuplicates(t *testing.T) {
	spec, a := testBase(t)
	start := model.DefaultStartTimeNS + 4*3600*1e9
	r, err := New(context.Background(), Config{
		Domain: spec, Adapter: a, Seed: 4, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.ApplyPerturb("duplicate_burst", map[string]any{"rate": 1.0}, 0, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(start+3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	if _, err := r.End(""); err != nil {
		t.Fatal(err)
	}
	seen := map[uint64]bool{}
	for _, row := range r.Ledger() {
		if row.DeliveryID == 0 || seen[row.DeliveryID] {
			t.Fatalf("ledger delivery id is missing or duplicated: %+v", row)
		}
		seen[row.DeliveryID] = true
	}
	if len(seen) < 2 {
		t.Fatalf("duplicate perturbation did not produce multiple delivery instances")
	}
}

func TestHistoryIsNotSilentlyCapped(t *testing.T) {
	spec, a := testBase(t)
	// A heartbeat at 1ms produces more than the old 10,000-entry cap in a
	// short stepped run without availability gating.
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
	start := model.DefaultStartTimeNS + 4*3600*1e9
	r, err := New(context.Background(), Config{
		Domain: spec, Adapter: a, Seed: 13, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(start+11*1e9, false); err != nil {
		t.Fatal(err)
	}
	if len(r.History()) <= 10000 {
		t.Fatalf("history was silently capped: %d", len(r.History()))
	}
}

// TestRunClosedLoop: fault -> evidence -> effector -> effect -> recovery,
// with idempotency: a repeated command_id applies exactly one effect.
func TestRunClosedLoop(t *testing.T) {
	spec, a := testBase(t)
	start := model.DefaultStartTimeNS + 4*3600*1e9
	r, err := New(context.Background(), Config{
		Domain: spec, Adapter: a, Seed: 11, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start,
	})
	if err != nil {
		t.Fatal(err)
	}
	pond := "site-a/pond-1"
	if _, err := r.Advance(start+1*3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	doBefore := r.World.StateValue(pond, "dissolved_oxygen_true", r.World.Clock())
	// Fault: the aerator stops.
	if _, err := r.InjectFault(pond, "aerator_failure", 0, nil); err != nil {
		t.Fatal(err)
	}
	// The effect (with time constant) propagates; DO falls through the night.
	if _, err := r.Advance(start+2*3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	doAfterFault := r.World.StateValue(pond, "dissolved_oxygen_true", r.World.Clock())
	if doAfterFault >= doBefore {
		t.Fatalf("fault should depress DO: before %v after %v", doBefore, doAfterFault)
	}
	// Actuate: start the aerator (command_id = idempotency key).
	res, err := r.InvokeEffector("start_aerator", pond, "cmd-1", map[string]any{"pond_id": pond, "level": 1.0}, r.World.Clock())
	if err != nil {
		t.Fatal(err)
	}
	if !res.Accepted {
		t.Fatalf("effector refused: %s", res.Reason)
	}
	// The same command_id must apply exactly one effect: the second call
	// replays the original result and the effector log still holds one entry
	// (the command_id is the action; a retry is the same action).
	res2, err := r.InvokeEffector("start_aerator", pond, "cmd-1", map[string]any{"pond_id": pond, "level": 1.0}, r.World.Clock())
	if err != nil {
		t.Fatal(err)
	}
	if res2.CommandID != res.CommandID || res2.Mode != res.Mode || len(r.World.EffectorCalls()) != 1 {
		t.Fatalf("idempotency broken: res=%+v res2=%+v calls=%+v", res, res2, r.World.EffectorCalls())
	}
	// The effect recovers DO over its time constant.
	if _, err := r.Advance(start+5*3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	doRecovered := r.World.StateValue(pond, "dissolved_oxygen_true", r.World.Clock())
	if doRecovered <= doAfterFault+0.3 {
		t.Fatalf("effector should raise DO: fault %v recovered %v", doAfterFault, doRecovered)
	}
	if _, err := r.End(""); err != nil {
		t.Fatal(err)
	}
	// The effector log is the authority on actions.
	calls := r.World.EffectorCalls()
	if len(calls) != 1 || calls[0].CommandID != "cmd-1" || !calls[0].EffectApplied {
		t.Fatalf("effector log wrong: %+v", calls)
	}
}

// TestReplayDetectsTampering: a corrupted expected digest must fail verify.
func TestReplayDetectsTampering(t *testing.T) {
	spec, a := testBase(t)
	cfg := Config{
		Domain: spec, Adapter: a, Seed: 5, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: model.DefaultStartTimeNS + 4*3600*1e9,
	}
	art := buildArtifact(t, cfg)
	art.ExpectedTraceDigest = "sha256:" + strings.Repeat("0", 64)
	res, err := ReplayArtifact(context.Background(), art, spec, a, "")
	if err != nil {
		t.Fatal(err)
	}
	if res.Matches {
		t.Fatal("tampered digest must not match")
	}
}

func TestReplayRejectsInputDigestMismatch(t *testing.T) {
	spec, a := testBase(t)
	art := buildArtifact(t, Config{
		Domain: spec, Adapter: a, Seed: 6, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: model.DefaultStartTimeNS + 4*3600*1e9,
	})
	bad := *art
	bad.Domain = art.Domain
	bad.Domain.Digest = "sha256:" + strings.Repeat("0", 64)
	if _, err := ReplayArtifact(context.Background(), &bad, spec, a, ""); err == nil {
		t.Fatal("replay must reject a changed domain digest before execution")
	}
}

// TestArtifactRoundTrip validates the artifact against its schema and the
// command log is a faithful record of the run.
func TestArtifactRoundTrip(t *testing.T) {
	spec, a := testBase(t)
	cfg := Config{
		Domain: spec, Adapter: a, Seed: 9, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: model.DefaultStartTimeNS + 4*3600*1e9,
	}
	dir := t.TempDir()
	r, err := New(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.InjectFault("site-a/pond-1", "do_probe_fouling", 0, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(cfg.StartTimeNS+2*3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	art, err := r.End(dir)
	if err != nil {
		t.Fatal(err)
	}
	// The artifact validates against the committed schema.
	raw, err := json.Marshal(art)
	if err != nil {
		t.Fatal(err)
	}
	if err := model.ValidateRunArtifact(raw); err != nil {
		t.Fatalf("artifact fails its own schema: %v", err)
	}
	// Replay from the on-disk artifact reproduces the run.
	loaded, err := LoadArtifact(filepath.Join(dir, "run.json"))
	if err != nil {
		t.Fatal(err)
	}
	res, err := ReplayArtifact(context.Background(), loaded, spec, a, "")
	if err != nil {
		t.Fatal(err)
	}
	if !res.Matches {
		t.Fatalf("replay of on-disk artifact failed: %s vs %s", res.GotDigest, res.WantDigest)
	}
	// Files written: trace, ledger, history, run.json.
	for _, name := range []string{"trace.jsonl", "ledger.jsonl", "world_state_history.jsonl", "run.json"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("missing artifact file %s: %v", name, err)
		}
	}
}
