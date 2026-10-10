package domain

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// TestOnlineOfflineScoringIdentity (F-0, P0): online and offline scoring
// come from one versioned bundle and must produce byte-identical results
// for every metric both can compute. History-dependent loop metrics
// (resolution, deadlines) are offline-uncomputable and are compared
// structurally, not for equality.
func TestOnlineOfflineScoringIdentity(t *testing.T) {
	r, gt := setupFaultedRun(t, "silent_no_effect", "aerator_failure")
	pond := "site-a/pond-1"
	if _, err := r.InvokeEffector("start_aerator", pond, "cmd-id", map[string]any{"pond_id": pond, "level": 1.0}, r.World.Clock()); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(context.Background(), r.World.Clock()+3*3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	submitVerdict(t, r, []model.Action{
		{CommandID: "setup", Effector: "start_aerator", EntityID: pond, IssuedAt: model.FormatTime(r.World.Clock()), OutcomeBelieved: model.BelievedSucceeded},
		{CommandID: "cmd-id", Effector: "start_aerator", EntityID: pond, IssuedAt: model.FormatTime(r.World.Clock()), OutcomeBelieved: model.BelievedSucceeded},
	}, nil)
	if _, err := r.End(""); err != nil {
		t.Fatal(err)
	}
	online, err := Score(evidenceOf(r), gt)
	if err != nil {
		t.Fatal(err)
	}
	offline := Offline(r.Verdict(), gt, r.Ledger(), r.World.EffectorCalls(), r.AppliedPerturbations())
	if online.Bundle != offline.Bundle || online.Bundle != scoringBundleVersion {
		t.Fatalf("both paths must carry the versioned bundle: online=%q offline=%q", online.Bundle, offline.Bundle)
	}
	// Judgment, instrument and consumer metrics are shared and must agree
	// byte-for-byte.
	shared := func(sc *Scorecard) map[string]any {
		return map[string]any{
			"judgment": sc.Judgment, "instrument": sc.Instrument, "consumer": sc.Consumer,
		}
	}
	a, _ := json.Marshal(shared(online))
	b, _ := json.Marshal(shared(offline))
	if string(a) != string(b) {
		t.Fatalf("online and offline scoring diverge:\nonline : %s\noffline: %s", a, b)
	}
	// The loop metrics offline can compute must agree too.
	if online.Loop.ActionAppropriate != offline.Loop.ActionAppropriate ||
		online.Loop.FalseSuccess != offline.Loop.FalseSuccess ||
		online.Loop.FalseSuccessRate != offline.Loop.FalseSuccessRate ||
		online.Loop.SilentNoEffectCalls != offline.Loop.SilentNoEffectCalls ||
		online.Loop.EffectCalls != offline.Loop.EffectCalls ||
		online.Loop.UnnecessaryAction != offline.Loop.UnnecessaryAction {
		t.Fatalf("shared loop metrics diverge: online=%+v offline=%+v", online.Loop, offline.Loop)
	}
}

// TestSilentNoEffectFalseSuccess is the highest-value single test in the
// plan: under silent_no_effect the confirmation channel reports the
// counterfactual (the motor draws current), so the evidence is internally
// consistent with success — only the absent physical outcome betrays the
// failure. A consumer reporting success has recorded a false outcome.
func TestSilentNoEffectFalseSuccess(t *testing.T) {
	r, gt := setupFaultedRun(t, "silent_no_effect", "aerator_failure")
	pond := "site-a/pond-1"
	// Actuate; the shadow state takes the effect, the real world does not.
	res, err := r.InvokeEffector("start_aerator", pond, "cmd-91", map[string]any{"pond_id": pond, "level": 1.0}, r.World.Clock())
	if err != nil {
		t.Fatal(err)
	}
	if !res.Accepted {
		t.Fatalf("silent_no_effect must ack success: %+v", res)
	}
	// The effect propagates.
	if _, err := r.Advance(context.Background(), r.World.Clock()+3*3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	at := r.World.Clock()
	// The confirmation channel reports the counterfactual: current drawn.
	current := r.World.Reading(pond, "pond.aerator_current", at)
	if current < 10.0 { // gain 18 x shadow aerator_output 1.0 + offset 0.4
		t.Fatalf("silent_no_effect must shadow the confirmation channel: current=%v", current)
	}
	// But the physical outcome never arrives: DO stays at the faulted level.
	do := r.World.StateValue(pond, "dissolved_oxygen_true", at)
	if do > 6.0 {
		t.Fatalf("real DO must not recover under silent_no_effect: %v", do)
	}
	// A dishonest consumer reports success.
	submitVerdict(t, r, []model.Action{
		{CommandID: "setup", Effector: "start_aerator", EntityID: pond, IssuedAt: model.FormatTime(at), OutcomeBelieved: model.BelievedSucceeded},
		{CommandID: "cmd-91", Effector: "start_aerator", EntityID: pond, IssuedAt: model.FormatTime(at), OutcomeBelieved: model.BelievedSucceeded},
	}, nil)
	if _, err := r.End(""); err != nil {
		t.Fatal(err)
	}
	sc, err := Score(evidenceOf(r), gt)
	if err != nil {
		t.Fatal(err)
	}
	if !sc.Loop.FalseSuccess || sc.Loop.FalseSuccessRate != 1.0 {
		t.Fatalf("false success must be caught: %+v", sc.Loop)
	}
}

// TestSilentNoEffectHonestConsumer: the same scenario with an honest
// consumer reports unknown and the metric reads zero.
func TestSilentNoEffectHonestConsumer(t *testing.T) {
	r, gt := setupFaultedRun(t, "silent_no_effect", "aerator_failure")
	pond := "site-a/pond-1"
	if _, err := r.InvokeEffector("start_aerator", pond, "cmd-92", map[string]any{"pond_id": pond, "level": 1.0}, r.World.Clock()); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(context.Background(), r.World.Clock()+3*3600*1e9, false); err != nil {
		t.Fatal(err)
	}
	submitVerdict(t, r, []model.Action{
		{CommandID: "setup", Effector: "start_aerator", EntityID: pond, IssuedAt: model.FormatTime(r.World.Clock()), OutcomeBelieved: model.BelievedSucceeded},
		{CommandID: "cmd-92", Effector: "start_aerator", EntityID: pond, IssuedAt: model.FormatTime(r.World.Clock()), OutcomeBelieved: model.BelievedUnknown},
	}, nil)
	if _, err := r.End(""); err != nil {
		t.Fatal(err)
	}
	sc, err := Score(evidenceOf(r), gt)
	if err != nil {
		t.Fatal(err)
	}
	if sc.Loop.FalseSuccess || sc.Loop.FalseSuccessRate != 0.0 {
		t.Fatalf("honest consumer must not be flagged: %+v", sc.Loop)
	}
}
