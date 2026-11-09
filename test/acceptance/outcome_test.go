package acceptance

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func storyScenario(story string) (string, bool) {
	dir := filepath.Join(testsupport.BeltConveyorStoriesDir(), "scenarios")
	if reference := filepath.Join(dir, story+".reference.scenario.json"); fileExists(reference) {
		return reference, true
	}
	return filepath.Join(dir, story+".do-nothing.scenario.json"), false
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func verdictAgainst(graded map[string]any, baseline string) map[string]any {
	verdicts, _ := graded["verdicts"].([]any)
	for _, item := range verdicts {
		if verdict, _ := item.(map[string]any); verdict["baseline"] == baseline {
			return verdict
		}
	}
	return nil
}

func policyCalls(graded map[string]any, mode string) int {
	runs, _ := graded["runs"].([]any)
	for _, item := range runs {
		if run, _ := item.(map[string]any); run["mode"] == mode {
			facts, _ := run["facts"].(map[string]any)
			calls, _ := facts["calls"].([]any)
			return len(calls)
		}
	}
	return -1
}

// baselineDid is what a baseline policy did on a story's first seed row: whether it made a call, and the story's own
// run (the reference policy, or nobody acting on a negative story) against it.
type baselineDid struct {
	fires   bool
	verdict string
}

const (
	beats = "beats"
	ties  = "ties"
)

var (
	quiet = baselineDid{false, ties}
	fires = baselineDid{true, beats}
)

// storyBaselines states what B1-B4 (PLAN.md section 2.2, O27) do on each story's first row. B1 (vibration) fires only
// where a bearing fails (S5-A too: a policy watches the vibration state, not the frozen reading), and stops the feed
// for good; B2 (current) on a lost drive and the surges above about 1.3x, but
// not on S3, whose currents pass 1.1x only in the seconds before the seizure; B3 (the safe limp feed from the start)
// fires everywhere and costs throughput; B4 on a lost drive. On S1-B B1 costs less than the reference policy, which
// limps at 18 % from 4 h after the onset, while B1 runs at full feed until its alarm 47 min before D-3 would trip, then
// empties the belt and stops D-3 in time: S1-B is easy by cost. On S6-A with the followers stopping no reaction saves
// the belt, so its reference ties doing nothing and every baseline (B3's throughput before the trip, 0.09 h on the
// first row, is inside the tie): it is easy by design. Every other story beats every baseline.
var storyBaselines = map[string][4]baselineDid{
	"s0-a":         {quiet, quiet, fires, quiet},
	"s1-a":         {fires, fires, fires, fires},
	"s1-b":         {{true, "loses"}, fires, fires, fires},
	"s2-a":         {{false, beats}, fires, fires, fires},
	"s2-b":         {{false, beats}, fires, fires, fires},
	"s3-a":         {{false, beats}, {false, beats}, fires, fires},
	"s4-a":         {quiet, fires, fires, quiet},
	"s4-b":         {quiet, fires, fires, quiet},
	"s5-a":         {fires, fires, fires, fires},
	"s6-a-stop":    {{false, ties}, {false, ties}, {true, ties}, {true, ties}},
	"s6-a-promote": {{false, beats}, fires, fires, fires},
}

// easyStories are the stories some baseline ties or beats; unavoidableStories, those no reaction can save, whose
// reference ties doing nothing.
var (
	easyStories        = map[string]bool{"s1-b": true, "s6-a-stop": true}
	unavoidableStories = map[string]bool{"s6-a-stop": true}
)

func againstNothing(story string, positive bool) string {
	if positive && !unavoidableStories[story] {
		return beats
	}
	return ties
}

func baselinesDid(graded map[string]any) [4]baselineDid {
	var did [4]baselineDid
	for i := range did {
		mode := fmt.Sprintf("policy:b%d", i+1)
		verdict, _ := verdictAgainst(graded, mode)["verdict"].(string)
		did[i] = baselineDid{policyCalls(graded, mode) > 0, verdict}
	}
	return did
}

func TestEveryStoryRunIsGradedAgainstDoingNothingAndEachBaseline(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("every conveyor story's first seed row run and graded through the binary")
	}
	b := newBinary(t)
	plantInputs := []string{"--domains-dir", filepath.Join(b.root, "examples"), "--adapters-dir", testsupport.AdaptersDir()}
	for story, want := range storyBaselines {
		t.Run(story, func(t *testing.T) {
			t.Parallel()
			scenario, positive := storyScenario(story)
			dir := t.TempDir()
			kinds := filepath.Join(b.root, "examples", "belt-conveyor", "kinds")
			b.run(append([]string{"run", "--scenario", scenario, "--sink", "file", "--out", dir, "--kinds-dir", kinds}, plantInputs...)...).mustSucceed(t)
			started := time.Now()
			graded := b.runWithin(30*time.Minute, append([]string{"outcome", "--run", dir}, plantInputs...)...).mustSucceed(t).document(t)
			t.Logf("%s: outcome took %s", story, time.Since(started).Round(time.Second))
			nothing, did := verdictAgainst(graded, "do_nothing"), baselinesDid(graded)
			t.Logf("%s: valid %v, easy %v; against do_nothing %v; b1-b4 (fires, verdict) %v; verdicts %v", story, graded["valid"], graded["easy"],
				nothing["verdict"], did, graded["verdicts"])
			if graded["valid"] != positive || graded["easy"] != easyStories[story] || nothing["verdict"] != againstNothing(story, positive) {
				t.Errorf("%s: valid %v, easy %v, %s doing nothing; want a story that is valid exactly when positive, easy only on s1-b and s6-a-stop, beating doing nothing on a positive story that can be saved and tying it otherwise",
					story, graded["valid"], graded["easy"], nothing["verdict"])
			}
			if did != want {
				t.Errorf("%s: baselines b1-b4 did %v, want %v", story, did, want)
			}
		})
	}
}
