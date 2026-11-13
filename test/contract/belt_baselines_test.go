package contract

import (
	"math"
	"slices"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const (
	ratedAmps     = 4.4
	magnetizingPU = 0.35
	b2Margin      = 1.1
)

// nominalDriveAmps is a drive's current at full feed with all three drives running, in closed form: the belt's running
// demand (empty belt, steady material, accelerating the feed, three drives' drag) shared by three, over the rating, through
// the motor's current curve I = I_rated·sqrt(i0² + (1 − i0²)·L²).
func nominalDriveAmps(regime beltRegime) float64 {
	demand := beltEmptyN + beltPerKgN*steadyMassKg + beltSpeed*nominalFeedKgS + 3*0.03*regime.rated
	load := demand / 3 / regime.rated
	return ratedAmps * math.Sqrt(magnetizingPU*magnetizingPU+(1-magnetizingPU*magnetizingPU)*load*load)
}

func TestB2AlarmsAtATenthAboveTheNominalDriveCurrent(t *testing.T) {
	t.Parallel()
	for _, regime := range beltRegimes {
		plant := beltPlant(t, regime.name)
		b := newBeltWorld(t, plant)
		b.advance(600)
		want := nominalDriveAmps(regime)
		policies := plant.Spec.GroundTruth.BaselinePolicies
		b2 := policies[slices.IndexFunc(policies, func(p model.BaselinePolicy) bool { return p.Name == "b2" })]
		t.Logf("plant %s: nominal drive current %.4f A (world %.4f A), b2 alarms above %.3f A", regime.name, want, b.state("D-1", "current"), b2.When.Threshold)
		if got := b.state("D-1", "current"); math.Abs(got-want) > 1e-3*want {
			t.Errorf("plant %s: drive current at full feed %.4f A, closed form %.4f A", regime.name, got, want)
		}
		if b2.When.State != "current" || math.Abs(b2.When.Threshold-b2Margin*want) > 1e-3 {
			t.Errorf("plant %s: b2 alarms on %s above %.4f, want current above %.4f A", regime.name, b2.When.State, b2.When.Threshold, b2Margin*want)
		}
	}
}
