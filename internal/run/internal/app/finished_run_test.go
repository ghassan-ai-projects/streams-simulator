package app

import (
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// A finished run's command log, world and evidence are sealed: every command
// is refused by name and none of them is recorded.
func TestEveryCommandIsRefusedOnceTheRunIsFinished(t *testing.T) {
	t.Parallel()
	spec, a := testBase(t)
	r, err := New(t.Context(), Config{Domain: spec, Adapter: a, SinkName: model.SinkInproc})
	if err != nil {
		t.Fatal(err)
	}
	entity := r.World.EntityIDs()[0]
	if _, err := r.End(""); err != nil {
		t.Fatal(err)
	}
	logged := len(r.commandLog)
	now := r.World.Clock()
	commands := map[string]func() error{
		"Advance":        func() error { _, err := r.Advance(t.Context(), now+1e9, false); return err },
		"InjectFault":    func() error { _, err := r.InjectFault(entity, "x", now, nil); return err },
		"ClearFault":     func() error { return r.ClearFault("f-0", now) },
		"ApplyPerturb":   func() error { _, err := r.ApplyPerturb("drop", nil, now, 0); return err },
		"ClearPerturb":   func() error { return r.ClearPerturb("p-0") },
		"InvokeEffector": func() error { _, err := r.InvokeEffector("e", entity, "c", nil, now); return err },
		"AddEntity":      func() error { return r.AddEntity("late", now) },
		"RetireEntity":   func() error { return r.RetireEntity(entity, "x", now) },
		"EnvInject":      func() error { _, err := r.EnvInject("t", "pause", nil, now); return err },
		"SubmitVerdict":  func() error { return r.SubmitVerdict(&model.Verdict{}) },
	}
	for name, command := range commands {
		if err := command(); err == nil || !strings.HasPrefix(err.Error(), name+": run is finished") {
			t.Errorf("%s after End: err = %v, want %q", name, err, name+": run is finished")
		}
	}
	if len(r.commandLog) != logged {
		t.Fatalf("a refused command must not be recorded: %d -> %d", logged, len(r.commandLog))
	}
}

func TestArtifactCountsEveryInjectedFaultEvenOnceCleared(t *testing.T) {
	t.Parallel()
	spec, a := testBase(t)
	r, err := New(t.Context(), Config{Domain: spec, Adapter: a, SinkName: model.SinkInproc})
	if err != nil {
		t.Fatal(err)
	}
	entities := r.World.EntityIDs()
	fault := spec.Spec.Faults[0].ID
	onset := r.World.Clock() + 60e9
	first, err := r.InjectFault(entities[0], fault, onset, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.InjectFault(entities[1], fault, onset, nil); err != nil {
		t.Fatal(err)
	}
	if err := r.ClearFault(first, onset+1); err != nil {
		t.Fatal(err)
	}
	art, err := r.End("")
	if err != nil {
		t.Fatal(err)
	}
	if art.Counts.FaultsInjected != 2 {
		t.Fatalf("faults injected = %d, want 2 (one cleared, one still active)", art.Counts.FaultsInjected)
	}
}

// Operator goroutines read the verdict and evidence while the director is
// still advancing the run; under -race any unsynchronised access fails.
func TestEvidenceReadsDoNotRaceWithAnAdvancingRun(t *testing.T) {
	t.Parallel()
	spec, a := testBase(t)
	r, err := New(t.Context(), Config{Domain: spec, Adapter: a, SinkName: model.SinkInproc})
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for step := int64(1); step <= 20; step++ {
			if _, err := r.Advance(t.Context(), r.World.Clock()+60e9*step, false); err != nil {
				return
			}
		}
	}()
	for {
		select {
		case <-done:
			return
		default:
			_ = r.Ledger()
			_ = r.History()
			_ = r.Verdict()
			_ = r.Reproducible()
			_ = r.AppliedPerturbations()
			r.Unblind()
			_, _ = r.Unblinded()
		}
	}
}
