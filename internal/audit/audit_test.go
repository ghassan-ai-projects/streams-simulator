package audit

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/truth"
)

const aquaculturePath = "../../docs/examples/aquaculture-pond.domain.json"

func loadSpec(t *testing.T) *domain.Compiled {
	t.Helper()
	spec, err := domain.Load(aquaculturePath)
	if err != nil {
		t.Fatal(err)
	}
	return spec
}

func ids(spec *domain.Compiled) []string {
	n := spec.Spec.Entities.Count.Default
	var out []string
	for i := 1; i <= n; i++ {
		out = append(out, "site-a/pond-"+string(rune('0'+i)))
	}
	return out
}

func runningAerator() []truth.SetupCall {
	return []truth.SetupCall{{
		Effector: "start_aerator", EntityID: "site-a/pond-1", CommandID: "setup",
		Args: map[string]any{"pond_id": "site-a/pond-1", "level": 1.0}, AtNS: model.DefaultStartTimeNS + 4*3600*1e9,
	}}
}

// TestLoudFaultIsTrivial: aerator_failure against a running aerator is a
// step on a dedicated confirmation channel — a fixed threshold with
// hindsight must solve it. That is the point of the audit: the label the
// threshold rule already gets right must not enter the graded suite.
func TestLoudFaultIsTrivial(t *testing.T) {
	spec := loadSpec(t)
	start := model.DefaultStartTimeNS + 4*3600*1e9
	onset := start + 2*3600*1e9
	panel := NewPanel(spec, 42, 60*1e9)
	v, err := panel.Audit("site-a/pond-1", "aerator_failure", onset, start, ids(spec), 12*3600*1e9, runningAerator(), nil)
	if err != nil {
		t.Fatal(err)
	}
	// The aerator current channel drops 18 A against sigma 0.2: any
	// hindsight-fitted threshold separates it.
	if !v.Trivial {
		t.Fatalf("aerator_failure should be trivial, scores: %+v", v.Scores)
	}
	if v.Best != "fixed_threshold" {
		t.Fatalf("expected fixed_threshold to win, got %s (%v)", v.Best, v.Scores)
	}
}

// TestProbeFoulingIsNonTrivial: the lethal sensor fault has no single
// channel signature — the probe reads a plausible value within the normal
// diurnal range. The audit must keep it out of the trivial bucket.
func TestProbeFoulingIsNonTrivial(t *testing.T) {
	spec := loadSpec(t)
	start := model.DefaultStartTimeNS + 4*3600*1e9
	onset := start + 2*3600*1e9
	panel := NewPanel(spec, 42, 60*1e9)
	v, err := panel.Audit("site-a/pond-1", "do_probe_fouling", onset, start, ids(spec), 8*3600*1e9, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if v.Trivial {
		t.Fatalf("do_probe_fouling must be non-trivial at an 8h horizon, scores: %+v", v.Scores)
	}
}
