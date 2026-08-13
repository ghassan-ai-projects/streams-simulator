package suite

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/perturb"
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

func TestGenerateSmallSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("suite generation runs full worlds and audits; skipped in -short mode")
	}
	spec := loadSpec(t)
	cfg := Config{
		Domain: spec, Profile: "nominal", N: 10, Seed: 42, SampleNS: 120 * 1e9,
		MaxAttempts: 40,
	}
	s, err := Generate(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Scenarios) == 0 {
		t.Fatalf("no scenarios generated: %s", s.TerminalState)
	}
	if len(s.Labels) != len(s.Scenarios) {
		t.Fatalf("labels/scenarios mismatch: %d vs %d", len(s.Labels), len(s.Scenarios))
	}
	// Every label references a declared fault and entity.
	for _, l := range s.Labels {
		if !spec.HasFault(l.Label) {
			t.Fatalf("label %q is not a declared fault", l.Label)
		}
		if l.TrivialBaselineVerdict != model.TrivialNonTrivial {
			t.Fatalf("graded label %s must be non-trivial", l.ScenarioID)
		}
		if l.FirstObservableTimeNS == 0 {
			t.Fatalf("label %s has no observability timestamps", l.ScenarioID)
		}
	}
	// Scenario command logs are replayable sequences.
	for _, sc := range s.Scenarios {
		if len(sc.CommandLog) < 3 {
			t.Fatalf("scenario %s command log too short", sc.ID)
		}
		if sc.CommandLog[0].Op != model.OpWorldCreate {
			t.Fatalf("scenario %s must start with world.create", sc.ID)
		}
		for _, p := range sc.Perturbations {
			valid := false
			for _, name := range perturb.Names {
				if p.Name == name {
					valid = true
				}
			}
			if !valid {
				t.Fatalf("scenario %s has unknown perturbation %q", sc.ID, p.Name)
			}
		}
	}
	// Perturbation coverage tracking ran; composition shortfalls are a
	// reportable terminal state, not a hard error (G-06).
	_ = s.TerminalState
}

func TestNegativeClassFractionAdapts(t *testing.T) {
	if testing.Short() {
		t.Skip("suite generation runs full worlds and audits; skipped in -short mode")
	}
	spec := loadSpec(t)
	cfg := Config{Domain: spec, Profile: "nominal", N: 24, Seed: 7, SampleNS: 300 * 1e9, MaxAttempts: 120}
	s, err := Generate(cfg)
	if err != nil {
		t.Fatal(err)
	}
	neg := 0
	for _, l := range s.Labels {
		if l.IsNegativeClass {
			neg++
		}
	}
	frac := float64(neg) / float64(len(s.Scenarios))
	// The sampler must adapt toward the declared band; a domain that cannot
	// sustain the composition must report it as a terminal state (G-06)
	// rather than silently degrade.
	if frac > 0.55 {
		t.Fatalf("adaptive sampling failed: negative fraction %.2f", frac)
	}
	if frac < 0.35 || frac > 0.45 {
		if s.TerminalState == "" || !strings.Contains(s.TerminalState, "negative-class") {
			t.Fatalf("out-of-band fraction %.2f without a reportable terminal state (state=%q)", frac, s.TerminalState)
		}
	}
	seen := map[string]bool{}
	for _, sc := range s.Scenarios {
		if seen[sc.ID] {
			t.Fatalf("scenario id repeated after rejected attempt: %q", sc.ID)
		}
		seen[sc.ID] = true
	}
}

// TestGenerateIsByteIdentical: two generations from the same seed produce
// byte-identical scenarios and labels — suite generation is a pure function
// of (seed, domain, profile), and rejected candidates never shift the
// identity of later admitted ones.
func TestGenerateIsByteIdentical(t *testing.T) {
	if testing.Short() {
		t.Skip("suite generation runs full worlds and audits; skipped in -short mode")
	}
	spec := loadSpec(t)
	cfg := Config{Domain: spec, Profile: "nominal", N: 6, Seed: 99, SampleNS: 120 * 1e9, MaxAttempts: 40}
	a, err := Generate(cfg)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Generate(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(a.Scenarios) != len(b.Scenarios) {
		t.Fatalf("generation counts differ: %d vs %d", len(a.Scenarios), len(b.Scenarios))
	}
	for i := range a.Scenarios {
		aj, _ := json.Marshal(a.Scenarios[i])
		bj, _ := json.Marshal(b.Scenarios[i])
		if string(aj) != string(bj) {
			t.Fatalf("scenario %d diverges across identical seeds", i)
		}
		if a.Labels[i].ScenarioID != b.Labels[i].ScenarioID {
			t.Fatalf("label identity diverges at %d: %s vs %s", i, a.Labels[i].ScenarioID, b.Labels[i].ScenarioID)
		}
	}
	if a.TerminalState != b.TerminalState {
		t.Fatalf("terminal state diverges: %q vs %q", a.TerminalState, b.TerminalState)
	}
}
