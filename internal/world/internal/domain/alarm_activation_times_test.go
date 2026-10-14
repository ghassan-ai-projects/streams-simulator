package domain

import (
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestAlarmActivationInstantsMatchTheIndependentScanAndAreOwnedCopies(t *testing.T) {
	t.Parallel()
	w := kpiWorld(t, func(*model.DomainSpec) {})
	if _, _, err := w.Advance(atSeconds(1200)); err != nil {
		t.Fatal(err)
	}
	var want []int64
	for _, at := range referenceAlarm(t, 1200).activations {
		want = append(want, atSeconds(at))
	}
	got := w.KPITruth().Entities[0].Alarms[0]
	if got.Activations != 3 || !slices.Equal(got.ActivationTimesNS, want) || got.Activations != len(got.ActivationTimesNS) {
		t.Fatalf("activation times = %v, count %d; want %v", got.ActivationTimesNS, got.Activations, want)
	}
	got.ActivationTimesNS[0] = -1
	if next := w.KPITruth().Entities[0].Alarms[0].ActivationTimesNS; !slices.Equal(next, want) {
		t.Fatalf("caller mutated world activation history: %v, want %v", next, want)
	}
}

func TestAnInitiallyStandingAlarmHasNoActivationInstant(t *testing.T) {
	t.Parallel()
	track := kpiTrack{activeSince: -1}
	track.sampleAlarm(atSeconds(0), true)
	track.sampleAlarm(atSeconds(10), true)
	got := track.alarmKPI(atSeconds(20), false)
	if got.Activations != 0 || got.ActivationTimesNS == nil || len(got.ActivationTimesNS) != 0 || !got.ActiveAtEnd {
		t.Fatalf("initially standing alarm = %+v; want an empty activation list and active at end", got)
	}
}
