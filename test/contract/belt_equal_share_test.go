package contract

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// The digests were taken from the plants before drives shared by weight (M3 row 3): with every weight 1 a run must emit
// the same readings and record the same outcome, bit for bit.
var equalShareRuns = map[string]struct{ trace, outcome string }{
	"s1-b.reference": {
		"sha256:bb7b1d48acf9cd094c25fbd30c20c944f40995c4fe5bdaf1109332b294169bd9",
		"sha256:3d18fa80c418bd0e035d038cc7ce228ae91a5ae042bfd5ca80c67e172653ba5d"},
	"s2-b.reference": {
		"sha256:e52f0836b349af90f9834a2006c383c779f69411851a04491990c63b0d3bd045",
		"sha256:d517abf528bb4b4437f1932104672c18889a5df573239a86e450d64041450218"},
	"s3-a.reference": {
		"sha256:7561f246ba749d365ff8c1788012de7dc9f9b0a18501054eaf48eb6faf9065a5",
		"sha256:06855279b05303bf820c87aa75edd47cf0443f6c1a7f2b6f5f478d787470d2ef"},
	"s4-a.do-nothing": {
		"sha256:052019fe1eb6c13795317f5078d1848fb94dff4dff2eb3eaa8721c3c09975c7c",
		"sha256:5611da863797a93255ef4a2d88e806bbfb17ea077fb0c1d38469a4379fc35c78"},
}

func outcomeDigest(t *testing.T, facts *model.OutcomeFacts) string {
	t.Helper()
	raw, err := json.Marshal(facts)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func TestWithEqualWeightsTheStoriesRunBitForBitAsBeforeWeightedSharing(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("four story runs of hours of world time")
	}
	for name, want := range equalShareRuns {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			art := runScenarioArtifact(t, storyScenario(t, name))
			got := outcomeDigest(t, outcomeOf(t, art, model.OutcomeAsRun))
			if art.ExpectedTraceDigest != want.trace || got != want.outcome {
				t.Errorf("trace %s outcome %s, want %s and %s", art.ExpectedTraceDigest, got, want.trace, want.outcome)
			}
		})
	}
}
