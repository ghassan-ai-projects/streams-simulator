package contract

import (
	"math"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

type beltLever struct {
	name     string
	prepare  func(b *beltWorld)
	effector string
	entity   string
	args     map[string]any
}

func emptyTheBelt(b *beltWorld) {
	b.invoke("bulk_feeder__set_feed", "F-1", map[string]any{"percent": 0})
	b.advance(600)
}

func stopTheBelt(b *beltWorld) {
	b.advance(60)
	b.invoke("belt_conveyor__stop_belt", "B-1", nil)
	b.advance(600)
}

var beltLevers = []beltLever{
	{"set_feed", func(*beltWorld) {}, "bulk_feeder__set_feed", "F-1", map[string]any{"percent": 50}},
	{"stop_drive", emptyTheBelt, "conveyor_drive__stop_drive", "D-3", nil},
	{"stop_belt", func(*beltWorld) {}, "belt_conveyor__stop_belt", "B-1", nil},
	{"start_belt", stopTheBelt, "belt_conveyor__start_belt", "B-1", nil},
}

func TestEveryLeverChangesTheSurvivorsLoadOrWhatTheBeltCarries(t *testing.T) {
	t.Parallel()
	for _, regime := range beltRegimes {
		for _, lever := range beltLevers {
			called, uncalled := newBeltWorld(t, beltPlant(t, regime.name)), newBeltWorld(t, beltPlant(t, regime.name))
			for _, b := range []*beltWorld{called, uncalled} {
				lever.prepare(b)
			}
			at := math.Max(called.nowS, 60)
			called.advance(at)
			if !called.invoke(lever.effector, lever.entity, lever.args) {
				t.Fatalf("plant %s: %s was refused", regime.name, lever.name)
			}
			for _, b := range []*beltWorld{called, uncalled} {
				b.advance(at + 120)
			}
			load := [2]float64{called.state("D-1", "load"), uncalled.state("D-1", "load")}
			mass := [2]float64{called.state("B-1", "m"), uncalled.state("B-1", "m")}
			t.Logf("plant %s, %s: D-1 load %.4f against %.4f, material %.1f kg against %.1f kg", regime.name, lever.name, load[0], load[1], mass[0], mass[1])
			if math.Abs(load[0]-load[1]) < 0.01 && math.Abs(mass[0]-mass[1]) < 1 {
				t.Errorf("plant %s: %s changed neither D-1's load nor the material on the belt", regime.name, lever.name)
			}
		}
	}
}

func TestAFrozenVibrationReadingDriftsAwayFromTheTrueValue(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("runs the belt for three hours")
	}
	testsupport.SkipUnderRace(t, "three hours of the belt")
	b := newBeltWorld(t, beltPlant(t, "a"))
	b.advance(3600)
	b.inject("D-1", "conveyor_drive__vibration__frozen", 1)
	held := b.w.Reading("D-1", "conveyor_drive.vibration", beltNS(b.nowS))
	b.advance(3 * 3600)
	reading, truth := b.w.Reading("D-1", "conveyor_drive.vibration", beltNS(b.nowS)), b.state("D-1", "vibration")
	if reading != held || math.Abs(truth-held) < 0.1 {
		t.Errorf("frozen at %.3f mm/s: reads %.3f two hours later while the drive vibrates at %.3f mm/s", held, reading, truth)
	}
	if peer := b.state("D-2", "vibration"); peer != truth {
		t.Errorf("the daily wiggle differs between drives: %.4f against %.4f mm/s", truth, peer)
	}
}
