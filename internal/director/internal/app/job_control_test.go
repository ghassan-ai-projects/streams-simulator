package app

import (
	"errors"
	"math"
	"testing"

	rules "github.com/ghassan-ai-projects/streams-simulator/internal/director/internal/domain"
)

func requireToolCode(t *testing.T, err error, code rules.Code) {
	t.Helper()
	var tool *rules.ToolError
	if !errors.As(err, &tool) || tool.Code != code {
		t.Fatalf("err = %v, want a %s refusal", err, code)
	}
}

func TestPausingAJobParksItAtTheNextBoundaryUntilItIsResumed(t *testing.T) {
	t.Parallel()
	pacer := newGatedPacer()
	d := pacedBy(shippedDirector(t, t.Context()), pacer)
	started := startJob(t, d, pondJob())
	<-pacer.asked
	paused, err := d.PauseJob(started.WorldID)
	if err != nil || paused.State != rules.JobPaused || paused.ReachedNS != started.StartNS {
		t.Fatalf("PauseJob = %+v, %v, want a paused job that has not moved", paused, err)
	}
	pacer.release(5)
	frozen, _ := d.JobStatus(started.WorldID)
	if frozen.State != rules.JobPaused || frozen.ReachedNS != started.StartNS || len(pacer.gate) != 5 {
		t.Fatalf("status %+v with %d open gates, want the job frozen while paused", frozen, len(pacer.gate))
	}
	resumed, err := d.ResumeJob(started.WorldID)
	if err != nil || resumed.State != rules.JobRunning {
		t.Fatalf("ResumeJob = %+v, %v, want a running job", resumed, err)
	}
	if again := <-pacer.asked; again != 120e9 {
		t.Fatalf("after the resume the pacer was asked for %d, want the same first chunk (120 s)", again)
	}
	pacer.release(1000)
	if status := awaitJob(t, d, started.WorldID); status.State != rules.JobCompleted {
		t.Fatalf("status = %+v, want the resumed job completed", status)
	}
	requireVerifies(t, d, started.WorldID)
}

func TestResumingReanchorsThePacerSoSimulatedTimeDoesNotCatchUp(t *testing.T) {
	t.Parallel()
	pacer := newGatedPacer()
	d := pacedBy(shippedDirector(t, t.Context()), pacer)
	started := startJob(t, d, pondJob())
	<-pacer.asked
	if _, err := d.PauseJob(started.WorldID); err != nil {
		t.Fatal(err)
	}
	if _, err := d.StepJob(started.WorldID); err != nil {
		t.Fatal(err)
	}
	if _, err := d.ResumeJob(started.WorldID); err != nil {
		t.Fatal(err)
	}
	if got, want := pacer.reanchored(), []anchor{{120e9, 60}}; len(got) != 1 || got[0] != want[0] {
		t.Fatalf("re-anchors = %v, want one at the reached instant (120 s) with the configured rate", got)
	}
	pacer.release(1000)
	awaitJob(t, d, started.WorldID)
}

func TestSteppingAPausedJobReleasesExactlyOneChunkWithoutPacing(t *testing.T) {
	t.Parallel()
	pacer := newGatedPacer()
	d := pacedBy(shippedDirector(t, t.Context()), pacer)
	started := startJob(t, d, pondJob())
	<-pacer.asked
	if _, err := d.PauseJob(started.WorldID); err != nil {
		t.Fatal(err)
	}
	for step := int64(1); step <= 2; step++ {
		status, err := d.StepJob(started.WorldID)
		if err != nil || status.State != rules.JobPaused || status.ReachedNS != started.StartNS+step*120e9 {
			t.Fatalf("step %d = %+v, %v, want one more chunk reached and the job paused again", step, status, err)
		}
	}
	if len(pacer.asked) != 0 {
		t.Fatalf("a stepped chunk asked the pacer %d times, want none", len(pacer.asked))
	}
	if _, err := d.StopJob(started.WorldID); err != nil {
		t.Fatal(err)
	}
	awaitJob(t, d, started.WorldID)
}

func TestSteppingARunningJobIsRefusedAsNotPaused(t *testing.T) {
	t.Parallel()
	pacer := newGatedPacer()
	d := pacedBy(shippedDirector(t, t.Context()), pacer)
	started := startJob(t, d, pondJob())
	<-pacer.asked
	_, err := d.StepJob(started.WorldID)
	requireToolCode(t, err, rules.CodeJobNotPaused)
	pacer.release(1000)
	awaitJob(t, d, started.WorldID)
	_, err = d.StepJob(started.WorldID)
	requireToolCode(t, err, rules.CodeJobNotRunning)
}

func TestChangingTheRateReanchorsAtTheReachedInstantAndKeepsTheJobRunning(t *testing.T) {
	t.Parallel()
	pacer := newGatedPacer()
	d := pacedBy(shippedDirector(t, t.Context()), pacer)
	started := startJob(t, d, pondJob())
	pacer.release(2)
	for range 3 {
		<-pacer.asked
	}
	status, err := d.SetJobRate(started.WorldID, 0)
	if err != nil || status.Rate != 0 || status.State != rules.JobRunning {
		t.Fatalf("SetJobRate = %+v, %v, want the new rate and a running job", status, err)
	}
	if got := pacer.reanchored(); len(got) != 1 || got[0] != (anchor{240e9, 0}) {
		t.Fatalf("re-anchors = %v, want one at 240 s with rate 0", got)
	}
	pacer.release(1000)
	awaitJob(t, d, started.WorldID)
}

func TestARateThatIsNotAFiniteNonNegativeNumberIsRefused(t *testing.T) {
	t.Parallel()
	pacer := newGatedPacer()
	d := pacedBy(shippedDirector(t, t.Context()), pacer)
	started := startJob(t, d, pondJob())
	<-pacer.asked
	for _, rate := range []float64{-1, math.NaN(), math.Inf(1)} {
		_, err := d.SetJobRate(started.WorldID, rate)
		requireToolCode(t, err, rules.CodeInvalidArgs)
	}
	if status, _ := d.JobStatus(started.WorldID); status.Rate != 60 || len(pacer.reanchored()) != 0 {
		t.Fatalf("status %+v after refused rates, want the rate and pacing untouched", status)
	}
	pacer.release(1000)
	awaitJob(t, d, started.WorldID)
}

func TestControlsNameUnknownWorldsAndJobsThatAreNotRunning(t *testing.T) {
	t.Parallel()
	d := pacedBy(shippedDirector(t, t.Context()), freePacer{})
	for name, call := range map[string]func(string) error{
		"pause":  func(id string) error { _, err := d.PauseJob(id); return err },
		"resume": func(id string) error { _, err := d.ResumeJob(id); return err },
		"step":   func(id string) error { _, err := d.StepJob(id); return err },
		"rate":   func(id string) error { _, err := d.SetJobRate(id, 1); return err },
	} {
		requireToolCode(t, call("w-99"), rules.CodeWorldNotFound)
		started := startJob(t, d, pondJob())
		awaitJob(t, d, started.WorldID)
		if err := call(started.WorldID); name != "step" {
			requireToolCode(t, err, rules.CodeJobNotRunning)
		}
	}
}

func TestPausingAndResumingAreIdempotentAndAPausedJobBlocksASecondStart(t *testing.T) {
	t.Parallel()
	pacer := newGatedPacer()
	d := pacedBy(shippedDirector(t, t.Context()), pacer)
	started := startJob(t, d, pondJob())
	<-pacer.asked
	for range 2 {
		if status, err := d.PauseJob(started.WorldID); err != nil || status.State != rules.JobPaused {
			t.Fatalf("PauseJob = %+v, %v", status, err)
		}
	}
	_, err := d.StartJob(pondJob())
	requireToolCode(t, err, rules.CodeJobActive)
	for range 2 {
		if status, err := d.ResumeJob(started.WorldID); err != nil || status.State != rules.JobRunning {
			t.Fatalf("ResumeJob = %+v, %v", status, err)
		}
	}
	pacer.release(1000)
	awaitJob(t, d, started.WorldID)
}
