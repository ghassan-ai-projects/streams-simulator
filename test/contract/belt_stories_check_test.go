package contract

import (
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

const (
	steadyMassKg     = 1250.0
	strandWithinS    = 60.0
	cascadeWithinS   = 300.0
	gearRiseEachHour = 0.5
	seizeLevel       = 0.95
)

var survivors = []string{"D-1", "D-2"}

type storyCheck struct {
	t     *testing.T
	story beltStory
	row   map[string]any
	runs  *storyRuns
}

// storyRuns shares runs between seed rows: a noiseless world draws from its seed only for effector acknowledgements,
// so rows without commands to the plant whose disturbances are equal (S0's) have one outcome.
type storyRuns struct {
	outcomes  sync.Map
	simulated atomic.Int64
}

type sharedRun struct {
	once    sync.Once
	outcome storyOutcome
}

func (c storyCheck) key(policy string) string {
	scenario := c.story.scenario(c.row, policy)
	key, _ := json.Marshal(scenario.Commands)
	if policy == policyReference {
		return fmt.Sprintf("%s/%s/%d/%s", c.story.ID, policy, scenario.Seed, key)
	}
	return c.story.ID + "/" + policy + "/" + string(key)
}

func (c storyCheck) field(name string) float64 { return rowNumbers(c.row, name)[0] }

func (c storyCheck) errorf(format string, args ...any) {
	c.t.Helper()
	c.t.Errorf("%s seed %.0f: "+format, append([]any{c.story.ID, c.row["seed"]}, args...)...)
}

func (c storyCheck) resolution() float64 {
	if c.story.Story == "S2" || c.story.Story == "S3" {
		return 1
	}
	return 10
}

func (c storyCheck) run(policy string) storyOutcome {
	anchor := math.Inf(1)
	if c.story.Story == "S2" {
		anchor = c.field("trip_s")
	}
	value, _ := c.runs.outcomes.LoadOrStore(c.key(policy), &sharedRun{})
	shared := value.(*sharedRun)
	shared.once.Do(func() {
		shared.outcome = runStory(c.t, c.story, c.row, policy, c.resolution(), anchor)
		c.runs.simulated.Add(int64(shared.outcome.simulatedS))
	})
	return shared.outcome
}

func (c storyCheck) quiet(o storyOutcome, policy string) {
	c.t.Helper()
	if when, what := o.firstOf(storyDrives, storyProtections); !math.IsInf(when, 1) {
		c.errorf("%s: %s tripped at %.0f s on a healthy belt", policy, what, when)
	}
	if !math.IsInf(o.at("B-1.stall_ok"), 1) || !math.IsInf(o.at("B-1.stranded"), 1) {
		c.errorf("%s: the belt stalled at %.0f s or was stranded at %.0f s", policy, o.at("B-1.stall_ok"), o.at("B-1.stranded"))
	}
}

func (c storyCheck) saved(o storyOutcome) {
	c.t.Helper()
	if when, what := o.firstOf(survivors, storyProtections); !math.IsInf(when, 1) {
		c.errorf("reference: survivor %s tripped at %.0f s", what, when)
	}
	if at := o.at("B-1.stranded"); !math.IsInf(at, 1) {
		c.errorf("reference: the belt was stranded at %.0f s", at)
	}
}

func (c storyCheck) strandsAtTheDownstreamStop(o storyOutcome) {
	c.t.Helper()
	stop, stranded := c.field("stop_s"), o.at("B-1.stranded")
	if when, what := o.firstOf(survivors, storyProtections); !math.IsInf(when, 1) {
		c.errorf("do nothing, plant a: survivor %s tripped at %.0f s", what, when)
	}
	if stranded < stop || stranded > stop+strandWithinS {
		c.errorf("do nothing, plant a: stranded at %.0f s, want within %.0f s of the downstream stop at %.0f s", stranded, strandWithinS, stop)
	}
}

func (c storyCheck) cascades(o storyOutcome, after float64) float64 {
	c.t.Helper()
	cascade, what := o.firstOf(survivors, storyProtections)
	if cascade <= after || cascade > after+cascadeWithinS || (what != "D-1.inverter_ok" && what != "D-2.inverter_ok") {
		c.errorf("do nothing, plant b: first survivor trip %s at %.0f s, want an inverter within %.0f s of D-3 at %.0f s", what, cascade, cascadeWithinS, after)
	}
	if stranded := o.at("B-1.stranded"); stranded < cascade || stranded > cascade+strandWithinS {
		c.errorf("do nothing, plant b: stranded at %.0f s, want within %.0f s of the cascade at %.0f s", stranded, strandWithinS, cascade)
	}
	return cascade
}

func (c storyCheck) slowBearing(policy string) {
	c.t.Helper()
	if policy == policyReference {
		reference := c.run(policyReference)
		c.saved(reference)
		if stopped, trip := reference.at("D-3.run_cmd"), c.d3Trip(); stopped > trip-c.story.StopLeadS || stopped >= trip {
			c.errorf("reference: D-3 stopped by command at %.0f s, not %.0f s before its do-nothing trip at %.0f s", stopped, c.story.StopLeadS, trip)
		}
		return
	}
	idle, onset := c.run(policyNone), c.field("onset_s")
	trip, what := idle.firstOf([]string{"D-3"}, storyProtections)
	c.t.Logf("%s seed %.0f: severity %.2f, D-3 %s %.2f h after onset", c.story.ID, c.row["seed"], c.field("severity"), what, (trip-onset)/3600)
	if trip < onset+6*3600 || trip > onset+9*3600 {
		c.errorf("do nothing: D-3 trips %.2f h after onset, want 6-9 h", (trip-onset)/3600)
	}
	if c.story.Plant == "a" {
		c.strandsAtTheDownstreamStop(idle)
	} else {
		cascade := c.cascades(idle, trip)
		c.t.Logf("%s seed %.0f: survivors' inverters trip %.0f s after D-3", c.story.ID, c.row["seed"], cascade-trip)
	}
}

func (c storyCheck) d3Trip() float64 {
	trip, _ := c.run(policyNone).firstOf([]string{"D-3"}, storyProtections)
	return trip
}

func (c storyCheck) suddenTrip(policy string) {
	c.t.Helper()
	if policy == policyReference {
		c.saved(c.run(policyReference))
		return
	}
	idle := c.run(policyNone)
	if c.story.Plant == "a" {
		c.strandsAtTheDownstreamStop(idle)
	} else {
		trip, regime := c.field("trip_s"), c.story.regime()
		m, v, feed := idle.anchor["B-1.m"], idle.anchor["B-1.v"], idle.anchor["B-1.feed"]
		load := (beltEmptyN + beltPerKgN*m + v*feed + 0.03*regime.rated*4) / 2 / regime.rated
		want := trip + 1 + tripAfter(inverterTauS, idle.anchor["D-1.theta_inv"], currentSquared(load))
		got := c.cascades(idle, trip)
		c.t.Logf("%s seed %.0f: cascade %.0f s after D-3, closed form %.1f s", c.story.ID, c.row["seed"], got-trip, want-trip)
		if math.Abs(got-want) > tripStepsS {
			c.errorf("do nothing: cascade at %.0f s, closed form %.1f s", got, want)
		}
	}
}

func (c storyCheck) seizure(policy string) {
	c.t.Helper()
	if policy == policyReference {
		reference := c.run(policyReference)
		c.saved(reference)
		if reference.seizedMassKg > 0.15*steadyMassKg || math.IsInf(reference.at("D-3.free_ok"), 1) {
			c.errorf("reference: %.0f kg on the belt when D-3 seized at %.0f s, want at most 15 %% of %.0f kg", reference.seizedMassKg, reference.at("D-3.free_ok"), steadyMassKg)
		}
		return
	}
	idle := c.run(policyNone)
	want := c.field("onset_s") + seizeLevel/(gearRiseEachHour*c.field("severity"))*3600
	seized, stall := idle.at("D-3.free_ok"), idle.at("B-1.stall_ok")
	c.t.Logf("%s seed %.0f: seizure %.2f h after onset (closed form %.2f h), stall %.0f s later, %.0f kg on the belt", c.story.ID, c.row["seed"],
		(seized-c.field("onset_s"))/3600, (want-c.field("onset_s"))/3600, stall-seized, idle.seizedMassKg)
	if math.Abs(seized-want) > 2 || stall < seized || stall > seized+boostS+1 || idle.seizedMassKg < 0.3*steadyMassKg {
		c.errorf("do nothing: seizure at %.0f s (closed form %.0f s), stall at %.0f s, %.0f kg on the belt", seized, want, stall, idle.seizedMassKg)
	}
	if stranded := idle.at("B-1.stranded"); stranded > stall+strandWithinS {
		c.errorf("do nothing: the stalled loaded belt was stranded at %.0f s", stranded)
	}
}

func (c storyCheck) tell(policy string) {
	switch c.story.Story {
	case "S0", "S4":
		c.quiet(c.run(policy), policy)
	case "S1", "S5":
		c.slowBearing(policy)
	case "S2":
		c.suddenTrip(policy)
	case "S3":
		c.seizure(policy)
	}
	c.scoreOutcome(policy)
}

// byCost puts the stories that simulate the most plant time first, so the longest runs start first.
func byCost(stories []beltStory) []beltStory {
	slices.SortStableFunc(stories, func(a, b beltStory) int {
		return compareFloat(b.HorizonS*float64(len(b.policies())), a.HorizonS*float64(len(a.policies())))
	})
	return stories
}

// runShared runs once, before the rows, every do-nothing run that several rows share.
func runShared(t *testing.T, story beltStory, runs *storyRuns) {
	t.Helper()
	seen := map[string]bool{}
	for _, row := range story.Rows {
		c := storyCheck{t: t, story: story, row: row, runs: runs}
		if key := c.key(policyNone); seen[key] {
			c.run(policyNone)
		} else {
			seen[key] = true
		}
	}
}

func TestEverySeedRowOfEveryStoryTellsItsStory(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("simulates every story for every seed row, with and without its reference policy")
	}
	testsupport.SkipUnderRace(t, "about 1000 noiseless plant hours")
	started, runs := time.Now(), &storyRuns{}
	t.Cleanup(func() {
		t.Logf("long tier: %.0f plant minutes in %s wall time", float64(runs.simulated.Load())/60, time.Since(started).Round(time.Second))
	})
	for _, story := range byCost(loadStories(t).Stories) {
		t.Run(story.ID, func(t *testing.T) {
			t.Parallel()
			runShared(t, story, runs)
			for i, row := range story.Rows {
				for _, policy := range story.policies() {
					t.Run(fmt.Sprintf("row-%d/%s", i+1, policy), func(t *testing.T) {
						t.Parallel()
						storyCheck{t: t, story: story, row: row, runs: runs}.tell(policy)
					})
				}
			}
		})
	}
}
