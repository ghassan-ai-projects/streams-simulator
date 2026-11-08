package acceptance

import (
	"os"
	"path/filepath"
	"testing"

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

func TestEveryStoryRunIsGradedAgainstDoingNothingAndB4(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("every conveyor story's first seed row run and graded through the binary")
	}
	b := newBinary(t)
	plantInputs := []string{"--domains-dir", filepath.Join(b.root, "examples"), "--adapters-dir", testsupport.AdaptersDir()}
	for _, story := range []string{"s0-a", "s1-a", "s1-b", "s2-a", "s2-b", "s3-a", "s4-a", "s4-b", "s5-a"} {
		t.Run(story, func(t *testing.T) {
			t.Parallel()
			scenario, positive := storyScenario(story)
			dir := t.TempDir()
			b.run(append([]string{"run", "--scenario", scenario, "--sink", "file", "--out", dir}, plantInputs...)...).mustSucceed(t)
			graded := b.run(append([]string{"outcome", "--run", dir}, plantInputs...)...).mustSucceed(t).document(t)
			nothing, b4 := verdictAgainst(graded, "do_nothing"), verdictAgainst(graded, "policy:b4")
			t.Logf("%s: valid %v; against do_nothing %v, against b4 %v; b4 calls %d", story, graded["valid"], nothing, b4, policyCalls(graded, "policy:b4"))
			switch {
			case positive && (graded["valid"] != true || nothing["verdict"] != "beats" || b4["verdict"] != "beats"):
				t.Errorf("%s: the reference policy should beat doing nothing and b4 on a valid story", story)
			case !positive && (graded["valid"] != false || nothing["verdict"] != "ties" || b4["verdict"] != "ties" || policyCalls(graded, "policy:b4") != 0):
				t.Errorf("%s: a negative story should leave doing nothing clean, b4 should never fire, and all should tie", story)
			}
		})
	}
}
