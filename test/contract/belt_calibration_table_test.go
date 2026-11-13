package contract

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"sync"
	"testing"
)

var calibrationThresholds = map[string][]float64{
	"vibration": {0.05, 0.1, 0.2, 0.3, 0.5, 1},
	"winding":   {0.5, 1, 2, 3, 5},
	"slope":     {0.5, 1, 2, 4},
}

// The thresholds M3's drive-degradation spec starts from (M3-DESIGN section 7), with the leads they must keep.
var calibratedThresholds = []struct {
	story, feature   string
	threshold, leadH float64
}{
	{"s1-a", "vibration", 0.3, 2}, {"s1-b", "vibration", 0.3, 2}, {"s1-a", "winding", 2, 2}, {"s1-b", "winding", 2, 2},
	{"s1-a", "slope", 2, 2}, {"s1-b", "slope", 2, 2}, {"s3-a", "winding", 1, 0.5}, {"s3-a", "slope", 2, 0.5},
}

const (
	gearboxSlopeKH  = 2
	bearingFloorMMS = 0.05
)

type calibrationResult struct {
	story  string
	onsetS float64
	tripS  float64
	d3     residuals
	worst  map[string]float64
}

func (r residuals) of(feature string) []float64 {
	return map[string][]float64{"vibration": r.vibration, "winding": r.winding, "slope": r.slope}[feature]
}

func calibrationRow(t *testing.T, story beltStory, row map[string]any) calibrationResult {
	c := calibrate(t, story, row)
	result := calibrationResult{story: story.ID, tripS: c.tripS, d3: c.residualsOf("D-3"), worst: map[string]float64{}}
	if onset, present := row["onset_s"].(float64); present {
		result.onsetS = onset
	}
	for _, drive := range storyDrives {
		of := c.residualsOf(drive)
		for feature := range calibrationThresholds {
			result.worst[feature] = math.Max(math.Max(result.worst[feature], largest(of.of(feature))), 0)
		}
	}
	return result
}

// runCalibration runs every seed row of the stories as a parallel subtest and hands their results to judge once all
// have finished.
func runCalibration(t *testing.T, judge func([]calibrationResult), ids ...string) {
	var mu sync.Mutex
	var results []calibrationResult
	t.Cleanup(func() { judge(results) })
	for _, story := range loadStories(t).Stories {
		if !slices.Contains(ids, story.ID) {
			continue
		}
		for i, row := range story.Rows {
			t.Run(fmt.Sprintf("%s-row%d", story.ID, i+1), func(t *testing.T) {
				t.Parallel()
				result := calibrationRow(t, story, row)
				mu.Lock()
				results = append(results, result)
				mu.Unlock()
			})
		}
	}
}

// leadsH are the hours from each row's first crossing to D-3's do-nothing trip (S3: the seizure), and the rows that
// crossed before the fault began.
func leadsH(results []calibrationResult, story, feature string, threshold float64) (leads []float64, early int) {
	for _, r := range results {
		if r.story != story {
			continue
		}
		crossing := firstCrossingS(r.d3.of(feature), threshold)
		if crossing < r.onsetS {
			early++
		}
		leads = append(leads, (r.tripS-crossing)/3600)
	}
	slices.Sort(leads)
	return leads, early
}

// gearboxCall is the first minute, before D-3's trip, at which D-3's winding slope residual reaches slopeKH while its
// vibration residual is still below vibrationMMS: a gearbox, not a bearing.
func gearboxCall(r calibrationResult, slopeKH, vibrationMMS float64) float64 {
	for j, value := range r.d3.slope {
		at := float64((j + 1) * vibrationStepS)
		if at < r.tripS && value >= slopeKH && r.d3.vibration[j] < vibrationMMS {
			return at
		}
	}
	return math.Inf(1)
}

func gearboxSummary(results []calibrationResult, story string) string {
	var calls []string
	for _, r := range results {
		if r.story == story {
			calls = append(calls, fmt.Sprintf("%.2f", (r.tripS-gearboxCall(r, gearboxSlopeKH, bearingFloorMMS))/3600))
		}
	}
	slices.Sort(calls)
	return fmt.Sprintf("%s gearbox call (slope >= %v K/h, vibration < %v mm/s), hours before the trip: %s", story, gearboxSlopeKH, bearingFloorMMS,
		strings.Join(calls, " "))
}

func leadSummary(results []calibrationResult, story, feature string, threshold float64) string {
	leads, early := leadsH(results, story, feature, threshold)
	return fmt.Sprintf("| %s | %s | %v | %.2f | %.2f | %.2f | %d |", story, feature, threshold, leads[0], leads[len(leads)/2], leads[len(leads)-1], early)
}

func worstOf(results []calibrationResult, story string) map[string]float64 {
	worst := map[string]float64{}
	for _, r := range results {
		if r.story == story {
			for feature, value := range r.worst {
				worst[feature] = math.Max(worst[feature], value)
			}
		}
	}
	return worst
}

func tripSummary(results []calibrationResult, story string) string {
	var lines []string
	for _, r := range results {
		if r.story == story {
			lines = append(lines, fmt.Sprintf("%.2f/%.2f", r.onsetS/3600, r.tripS/3600))
		}
	}
	slices.Sort(lines)
	return story + " onset/trip h: " + strings.Join(lines, " ")
}

func calibrationTable(results []calibrationResult, positive, negative []string) string {
	lines := []string{"| story | residual | threshold | lead min h | lead median h | lead max h | rows before onset |", "|---|---|---|---|---|---|---|"}
	for _, story := range positive {
		for _, feature := range []string{"vibration", "winding", "slope"} {
			for _, threshold := range calibrationThresholds[feature] {
				lines = append(lines, leadSummary(results, story, feature, threshold))
			}
		}
	}
	lines = append(lines, "", "| story | max vibration residual mm/s | max winding residual K | max slope residual K/h |", "|---|---|---|---|")
	for _, story := range negative {
		worst := worstOf(results, story)
		lines = append(lines, fmt.Sprintf("| %s | %.3f | %.3f | %.3f |", story, worst["vibration"], worst["winding"], worst["slope"]))
	}
	for _, story := range positive {
		lines = append(lines, tripSummary(results, story), gearboxSummary(results, story))
	}
	return strings.Join(lines, "\n")
}

func TestTheDriveDegradationResidualsWarnLongBeforeTheTripAndStayQuietOnSurgesAndANormalDay(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("every seed row of S0, S1, S3 and S4, noisy, to their horizons")
	}
	positive, negative := []string{"s1-a", "s1-b", "s3-a"}, []string{"s0-a", "s4-a", "s4-b"}
	runCalibration(t, func(results []calibrationResult) { judgeCalibration(t, results, positive, negative) }, slices.Concat(positive, negative)...)
}

func judgeCalibration(t *testing.T, results []calibrationResult, positive, negative []string) {
	if t.Failed() {
		return
	}
	t.Log("\n" + calibrationTable(results, positive, negative))
	for _, claim := range calibratedThresholds {
		if leads, early := leadsH(results, claim.story, claim.feature, claim.threshold); leads[0] < claim.leadH || early > 0 {
			t.Errorf("%s %s residual %v: least lead %.2f h (want %.1f h), %d rows before the fault", claim.story, claim.feature, claim.threshold, leads[0], claim.leadH, early)
		}
	}
	for _, r := range results {
		call := gearboxCall(r, gearboxSlopeKH, bearingFloorMMS)
		if (r.story == "s3-a") != (r.tripS-call >= 0.5*3600) {
			t.Errorf("%s: a gearbox call %.2f h before the trip, want one at least 0.5 h before a seizure and none on a bearing", r.story, (r.tripS-call)/3600)
		}
	}
	for _, story := range negative {
		for _, claim := range calibratedThresholds {
			if worst := worstOf(results, story)[claim.feature]; worst >= claim.threshold/2 {
				t.Errorf("%s: largest %s residual %.3f, want under half the threshold %v", story, claim.feature, worst, claim.threshold)
			}
		}
	}
}
