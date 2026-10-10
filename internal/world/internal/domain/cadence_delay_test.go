package domain

import (
	"math"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/randutil"
)

func TestReportByExceptionEmitsTheFirstReadingAndThenOnlyBeyondTheDeadband(t *testing.T) {
	t.Parallel()
	ch := &model.Channel{Cadence: model.Cadence{Mode: "report_by_exception", Deadband: 0.5}}
	state := &channelRunState{}
	if !exceptionCadence(ch, state, 10) {
		t.Fatal("the first reading must always be sent")
	}
	if exceptionCadence(ch, state, 10.2) {
		t.Fatal("a change inside the deadband must not be sent")
	}
	if !exceptionCadence(ch, state, 10.5) {
		t.Fatal("a change of exactly the deadband must be sent")
	}
}

func TestEventDrivenCadenceFiresOnAResolutionStepAndNeverForStrings(t *testing.T) {
	t.Parallel()
	numeric := &model.Channel{ValueType: "number", Resolution: 1}
	state := &channelRunState{hasSent: true, lastTrigger: 5}
	if eventCadence(numeric, state, 5.4) {
		t.Fatal("a change below the resolution is not an event")
	}
	if !eventCadence(numeric, state, 6.5) {
		t.Fatal("a change of a resolution step is an event")
	}
	if state.lastTrigger != 6.5 {
		t.Fatalf("the trigger reference must follow the reading, got %v", state.lastTrigger)
	}
	if eventCadence(&model.Channel{ValueType: "string"}, &channelRunState{}, 1) {
		t.Fatal("a string channel has no numeric trigger")
	}
}

func TestCheckedEmissionPollsAtThePeriodOrEveryMinute(t *testing.T) {
	t.Parallel()
	const second = int64(1e9)
	if got := checkedEmission(&model.Channel{Cadence: model.Cadence{PeriodS: 5}}, 100*second); got != 105*second {
		t.Fatalf("with a period = %d, want %d", got, 105*second)
	}
	if got := checkedEmission(&model.Channel{}, 100*second); got != 160*second {
		t.Fatalf("without a period = %d, want %d", got, 160*second)
	}
}

func TestEmissionWithoutAPeriodNeverSchedulesAHumanOrPeriodicTick(t *testing.T) {
	t.Parallel()
	w := &World{}
	ch := &model.Channel{Name: "inspection"}
	if got := w.humanEmission("e-1", ch, 7); got != 0 {
		t.Fatalf("human-driven channel without a period = %d, want 0", got)
	}
	if got := w.periodicEmission("e-1", ch, 7); got != 0 {
		t.Fatalf("periodic channel without a period = %d, want 0", got)
	}
}

func TestLinkDelayModelsHonourTheirDeclaredShape(t *testing.T) {
	t.Parallel()
	rng := randutil.NewSplitMix64(7)
	if got := sampleLinkDelay(&model.LinkDelay{Model: "constant", MeanS: 0.25}, rng); got != 0.25 {
		t.Fatalf("constant delay = %v", got)
	}
	if got := sampleLinkDelay(&model.LinkDelay{Model: "unknown", MeanS: 1}, rng); got != 0 {
		t.Fatalf("an unknown model adds no delay, got %v", got)
	}
	for range 200 {
		if got := sampleLinkDelay(&model.LinkDelay{Model: "exponential", MeanS: 1, MaxS: 0.5}, rng); got < 0 || got > 0.5 {
			t.Fatalf("exponential delay %v escapes [0, max_s]", got)
		}
		if got := sampleLinkDelay(&model.LinkDelay{Model: "store_and_forward", MeanS: 2}, rng); got < 2 {
			t.Fatalf("store-and-forward delay %v is below the mean floor", got)
		}
	}
}

func TestLognormalLinkDelayHasTheDeclaredMean(t *testing.T) {
	t.Parallel()
	rng := randutil.NewSplitMix64(11)
	delay := &model.LinkDelay{Model: "lognormal", MeanS: 2, SigmaS: 0.5}
	const n = 20000
	sum := 0.0
	for range n {
		value := sampleLinkDelay(delay, rng)
		if value <= 0 {
			t.Fatalf("lognormal delay %v must be positive", value)
		}
		sum += value
	}
	if mean := sum / n; math.Abs(mean-2)/2 > 0.05 {
		t.Fatalf("sample mean %v, want 2 within 5%%", mean)
	}
	// A missing sigma falls back to mean/2 and still yields positive delays.
	if got := sampleLinkDelay(&model.LinkDelay{Model: "lognormal", MeanS: 2}, rng); got <= 0 {
		t.Fatalf("default-sigma delay = %v", got)
	}
}
