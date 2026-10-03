package run

import (
	"context"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

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
	if _, err := r.Advance(context.Background(), start+1*3600*1e9, false); err != nil {
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
	if _, err := r.Advance(context.Background(), start+3600*1e9, false); err != nil {
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
	if _, err := r.Advance(context.Background(), start+11*1e9, false); err != nil {
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
	if _, err := r.Advance(context.Background(), start+1*3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	doBefore := r.World.StateValue(pond, "dissolved_oxygen_true", r.World.Clock())
	// Fault: the aerator stops.
	if _, err := r.InjectFault(pond, "aerator_failure", 0, nil); err != nil {
		t.Fatal(err)
	}
	// The effect (with time constant) propagates; DO falls through the night.
	if _, err := r.Advance(context.Background(), start+2*3600*1e9, false); err != nil {
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
	if _, err := r.Advance(context.Background(), start+5*3600*1e9, false); err != nil {
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
