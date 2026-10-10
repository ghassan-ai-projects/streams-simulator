package domain

import (
	"math"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// Epoch zero is a legal start time: an onset observable at t=0 is "at time 0",
// not "never observable".
func TestAnOnsetObservableAtTimeZeroIsObservable(t *testing.T) {
	t.Parallel()
	spec := loadSpec(t)
	setup := []model.SetupCall{{
		Effector: "start_aerator", EntityID: "site-a/pond-1", CommandID: "setup",
		Args: map[string]any{"pond_id": "site-a/pond-1", "level": 1.0}, AtNS: -7200e9,
	}}
	solver := NewSolver(spec, 42, 60*1e9, 24*3600*1e9)
	res, err := solver.solve("site-a/pond-1", "aerator_failure", 0, 0, entityIDs(spec), setup)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Observable || res.FirstObservableNS != 0 {
		t.Fatalf("an immediately observable onset at t=0 must be observable at t=0: %+v", res)
	}
	if res.UnavoidableNS != 0 {
		t.Fatalf("the unavoidable threshold is crossed at the same instant: %+v", res)
	}
}

// A detector that cannot be evaluated must not turn every fault into an
// instantly observable one.
func TestAnUnknownDetectorFormIsRefusedNotTreatedAsObservable(t *testing.T) {
	t.Parallel()
	spec := loadSpec(t)
	for i := range spec.Spec.Faults {
		if spec.Spec.Faults[i].ID == "aerator_failure" {
			spec.Spec.Faults[i].Observability.Detector.Form = "no_such_form"
		}
	}
	solver := NewSolver(spec, 42, 60*1e9, 3600*1e9)
	_, err := solver.solve("site-a/pond-1", "aerator_failure", 3600*1e9, 0, entityIDs(spec), nil)
	if err == nil || !strings.Contains(err.Error(), `unknown detector form "no_such_form"`) {
		t.Fatalf("err = %v", err)
	}
}

// A peer residual pools the variance of the siblings the world holds, even
// when the caller leaves the entity list to the world's defaults.
func TestPeerResidualSigmaCountsTheEntitiesTheWorldHolds(t *testing.T) {
	t.Parallel()
	spec := loadSpec(t)
	for i := range spec.Spec.Faults {
		if spec.Spec.Faults[i].ID == "aerator_failure" {
			spec.Spec.Faults[i].Observability.Detector = model.Detector{Form: model.DetectorPeerResidual, Channel: "pond.aerator_current"}
		}
	}
	solver := NewSolver(spec, 42, 60*1e9, 3600*1e9)
	res, err := solver.solve("site-a/pond-1", "aerator_failure", 1800*1e9, 0, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	peers := float64(spec.Spec.Entities.Count.Default)
	want := 0.2 * math.Sqrt(1+1/(peers-1))
	if math.Abs(res.EffectiveSigma-want) > 1e-9 {
		t.Fatalf("effective sigma = %v, want %v for %v peers", res.EffectiveSigma, want, peers)
	}
}

// With no noise floor every threshold is zero, and zero deviation is still
// not a detection: a fault that changes nothing is not observable.
func TestAFaultWithNoEffectIsNotObservableEvenWithoutNoise(t *testing.T) {
	t.Parallel()
	spec := loadSpec(t)
	for i := range spec.Spec.Channels {
		if spec.Spec.Channels[i].Name == "pond.aerator_current" {
			spec.Spec.Channels[i].Noise.Sigma = 0
		}
	}
	solver := NewSolver(spec, 42, 60*1e9, 3600*1e9)
	start := model.DefaultStartTimeNS
	// The aerator is off, so its failure changes nothing the detector sees.
	res, err := solver.solve("site-a/pond-1", "aerator_failure", start+600e9, start, entityIDs(spec), nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.Observable || res.EffectiveSigma != 0 {
		t.Fatalf("a no-effect fault must not be observable: %+v", res)
	}
}
