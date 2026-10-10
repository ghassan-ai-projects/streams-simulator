package app

import (
	"context"
	"reflect"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter"
	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/run"
)

// scoringEvidence is the only place a run is packed for the scorer, so each
// field is pinned to the run accessor it must come from, with values chosen
// so that swapping or dropping a field is visible.
func TestScoringEvidencePacksEveryFieldFromTheRun(t *testing.T) {
	t.Parallel()
	for name, unblind := range map[string]bool{"blind run": false, "unblinded run": true} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			checkScoringEvidence(t, unblind)
		})
	}
}

func checkScoringEvidence(t *testing.T, unblind bool) {
	t.Helper()
	spec, err := domain.Load(aquaculturePath)
	if err != nil {
		t.Fatal(err)
	}
	a, err := adapter.Load(nativeAdapter)
	if err != nil {
		t.Fatal(err)
	}
	start := model.DefaultStartTimeNS
	r, err := run.New(context.Background(), run.Config{
		Domain: spec, Adapter: a, Seed: 3, SinkName: model.SinkInproc,
		TimeMode: model.TimeStepped, StartTimeNS: start,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.ApplyPerturb("duplicate_burst", map[string]any{"rate": 1.0}, start, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Advance(context.Background(), start+1800*1e9, false); err != nil {
		t.Fatal(err)
	}
	if unblind {
		r.Unblind()
	}

	ev := scoringEvidence(r)
	if ev.RunID != r.ID || ev.Domain != r.Domain() {
		t.Fatalf("identity = %q %v", ev.RunID, ev.Domain)
	}
	if ev.Emitted == 0 || ev.Emitted != r.World.EmittedCount() {
		t.Fatalf("emitted = %d, world %d", ev.Emitted, r.World.EmittedCount())
	}
	if len(ev.Perturbations) == 0 || !reflect.DeepEqual(ev.Perturbations, r.AppliedPerturbations()) {
		t.Fatalf("perturbations = %v", ev.Perturbations)
	}
	if len(ev.Ledger) == 0 || len(ev.Ledger) != len(r.Ledger()) || len(ev.History) != len(r.History()) {
		t.Fatalf("ledger %d/%d, history %d/%d", len(ev.Ledger), len(r.Ledger()), len(ev.History), len(r.History()))
	}
	if ev.Unblinded != unblind || ev.Unblinded != r.UnblindedStamp() {
		t.Fatalf("unblinded = %v, want %v", ev.Unblinded, unblind)
	}
	if !ev.Reproducible || ev.Reproducible != r.Reproducible() {
		t.Fatalf("reproducible = %v, run says %v", ev.Reproducible, r.Reproducible())
	}
}
