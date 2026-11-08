package acceptance

import (
	"path/filepath"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestOutcomeGradesAStoryRunWithTheReferencePolicyAboveDoingNothing(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("a three-hour conveyor story, replayed four times")
	}
	b := newBinary(t)
	dir, examples := t.TempDir(), filepath.Join(b.root, "examples")
	plantInputs := []string{"--domains-dir", examples, "--adapters-dir", testsupport.AdaptersDir()}
	scenario := filepath.Join(testsupport.BeltConveyorStoriesDir(), "scenarios", "s2-a.reference.scenario.json")
	b.run(append([]string{"run", "--scenario", scenario, "--sink", "file", "--out", dir}, plantInputs...)...).mustSucceed(t)
	graded := b.run(append([]string{"outcome", "--run", dir}, plantInputs...)...).mustSucceed(t).document(t)
	verdicts, _ := graded["verdicts"].([]any)
	first, _ := verdicts[0].(map[string]any)
	if graded["valid"] != true || first["baseline"] != "do_nothing" || first["verdict"] != "beats" {
		t.Fatalf("outcome %v, want a valid story whose reference policy beats doing nothing", graded)
	}
	t.Logf("verdicts: %v", verdicts)
}
