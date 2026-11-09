package contract

import (
	"math"
	"sort"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

const (
	windingStepS   = 10
	vibrationStepS = 60
	meanWindowS    = 900
	slopeWindowS   = 1800
)

// calibrationRun is a do-nothing story run as a host receives it: every drive's winding (every 10 s) and vibration (every
// 60 s) readings as the world emits them, with their noise and outages, and the first protective trip of D-3.
type calibrationRun struct {
	run     *storyRun
	samples map[string][]timedValue
	tripS   float64
}

type timedValue struct{ atS, value float64 }

func calibrate(t *testing.T, story beltStory, row map[string]any) *calibrationRun {
	t.Helper()
	scenario := story.scenario(row, policyNone)
	w, err := world.New(beltPlant(t, story.Plant), scenario.Seed, "belt", model.DefaultStartTimeNS, world.Options{})
	if err != nil {
		t.Fatal(err)
	}
	c := &calibrationRun{run: &storyRun{beltWorld: &beltWorld{t: t, w: w}}, samples: map[string][]timedValue{}, tripS: math.Inf(1)}
	w.SetEmitter(c.collect)
	for _, command := range scenario.Commands {
		if command.Op == model.OpClockAdvance {
			c.stepUntil(secondsOf(command.Args["to_ns"]))
			continue
		}
		c.run.execute(command)
	}
	return c
}

func (c *calibrationRun) collect(event model.SimEvent) {
	if event.Channel != "conveyor_drive.winding" && event.Channel != "conveyor_drive.vibration" || event.Value == nil {
		return
	}
	at, err := model.ParseTime(event.EventTime)
	if err != nil {
		c.run.t.Fatal(err)
	}
	key := event.EntityID + "/" + event.Channel
	c.samples[key] = append(c.samples[key], timedValue{float64(at-model.DefaultStartTimeNS) / 1e9, number(event.Value)})
}

func (c *calibrationRun) stepUntil(untilS float64) {
	for next := math.Floor(c.run.nowS/windingStepS)*windingStepS + windingStepS; next <= untilS; next += windingStepS {
		c.run.advance(next)
		c.watchTrip(next)
	}
	if c.run.nowS < untilS {
		c.run.advance(untilS)
	}
}

func (c *calibrationRun) watchTrip(atS float64) {
	for _, protection := range storyProtections {
		if math.IsInf(c.tripS, 1) && c.run.state("D-3", protection) == 0 {
			c.tripS = atS
		}
	}
}

// features are the three per-drive views a spec computes at the end of every minute: the 15 min means of vibration and
// winding and the least-squares winding slope over 30 min (K/h), each over a covered window.
type features struct{ vibration, winding, slope []float64 }

func (c *calibrationRun) features(drive string) features {
	minutes := int(c.run.nowS / vibrationStepS)
	vibration, winding := c.samples[drive+"/conveyor_drive.vibration"], c.samples[drive+"/conveyor_drive.winding"]
	f := features{make([]float64, minutes), make([]float64, minutes), make([]float64, minutes)}
	for j := range minutes {
		end := float64((j + 1) * vibrationStepS)
		f.vibration[j] = mean(covered(within(vibration, end-meanWindowS, end), meanWindowS/vibrationStepS))
		f.winding[j] = mean(covered(within(winding, end-meanWindowS, end), meanWindowS/windingStepS))
		f.slope[j] = slope(covered(within(winding, end-slopeWindowS, end), slopeWindowS/windingStepS))
	}
	return f
}

// within is the samples in (from, until]; it is empty until the run has lasted a whole window.
func within(samples []timedValue, from, until float64) []timedValue {
	if from < 0 {
		return nil
	}
	lo := sort.Search(len(samples), func(i int) bool { return samples[i].atS > from })
	hi := sort.Search(len(samples), func(i int) bool { return samples[i].atS > until })
	return samples[lo:hi]
}

// covered keeps a window only when it holds at least 90 % of the readings its cadence gives: a sensor outage (the
// channels' availability is 0.999 with 10 min repairs) during a load change would otherwise bias one drive's view.
func covered(samples []timedValue, expected int) []timedValue {
	if 10*len(samples) < 9*expected {
		return nil
	}
	return samples
}

func mean(samples []timedValue) float64 {
	if len(samples) == 0 {
		return math.NaN()
	}
	total := 0.0
	for _, sample := range samples {
		total += sample.value
	}
	return total / float64(len(samples))
}

func slope(samples []timedValue) float64 {
	if len(samples) < 2 {
		return math.NaN()
	}
	tMean, yMean := 0.0, mean(samples)
	for _, sample := range samples {
		tMean += sample.atS / float64(len(samples))
	}
	num, den := 0.0, 0.0
	for _, sample := range samples {
		num += (sample.atS - tMean) * (sample.value - yMean)
		den += (sample.atS - tMean) * (sample.atS - tMean)
	}
	return num / den * 3600
}

// residual is a drive's view minus the largest of its peers', per minute.
func residual(own []float64, peers ...[]float64) []float64 {
	out := make([]float64, len(own))
	for j := range own {
		largest := math.Inf(-1)
		for _, peer := range peers {
			largest = math.Max(largest, peer[j])
		}
		out[j] = own[j] - largest
	}
	return out
}

type residuals struct{ vibration, winding, slope []float64 }

func (c *calibrationRun) residualsOf(drive string) residuals {
	own, peers := c.features(drive), []features{}
	for _, other := range storyDrives {
		if other != drive {
			peers = append(peers, c.features(other))
		}
	}
	return residuals{residual(own.vibration, peers[0].vibration, peers[1].vibration), residual(own.winding, peers[0].winding, peers[1].winding),
		residual(own.slope, peers[0].slope, peers[1].slope)}
}

// firstCrossingS is the first minute at which a series reaches the threshold, in seconds from the start.
func firstCrossingS(series []float64, threshold float64) float64 {
	for j, value := range series {
		if value >= threshold {
			return float64((j + 1) * vibrationStepS)
		}
	}
	return math.Inf(1)
}

func largest(series []float64) float64 {
	top := math.Inf(-1)
	for _, value := range series {
		if !math.IsNaN(value) {
			top = math.Max(top, value)
		}
	}
	return top
}
