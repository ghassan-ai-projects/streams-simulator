package app

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// FuzzScenarioCheck: a scenario file passes through contract validation,
// decoding, the static check and the dry run before anything is delivered;
// none of them may panic on adversarial bytes.
func FuzzScenarioCheck(f *testing.F) {
	spec, a := testBase(f)
	seeds := []string{
		`{"schema_version":"0.1","id":"s","domain":"aquaculture-pond","adapter":"native-jsonl","seed":1,"world":{"start_time_ns":0},"commands":[]}`,
		`{"schema_version":"0.1","id":"s","domain":"aquaculture-pond","adapter":"native-jsonl","seed":1,"world":{"start_time_ns":0,"profile":"x"},` +
			`"commands":[{"op":"effector.invoke","args":{"effector":"start_aerator","entity_id":"site-a/pond-1","command_id":"c","args":{"level":"high"},"at_ns":0}}]}`,
		`{"commands":[{"op":"clock.advance","args":{"to_ns":-1}}]}`,
		`not json`,
	}
	for _, seed := range seeds {
		f.Add([]byte(seed))
	}
	f.Fuzz(func(t *testing.T, raw []byte) {
		scenario, err := model.DecodeScenario(raw)
		if err != nil {
			return
		}
		_ = CheckScenario(t.Context(), ScenarioConfig(scenario, spec, a, model.SinkInproc, ""), scenario)
	})
}
