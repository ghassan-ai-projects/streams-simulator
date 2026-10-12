package model

import (
	"errors"
	"strings"
	"testing"
)

const scenarioHead = `{"schema_version":"0.1","id":"s","domain":"aquaculture-pond","adapter":"native-jsonl","seed":1,"world":{"start_time_ns":1},"commands":[`

// A malformed command is reported at the argument it concerns, named by its
// op, not as an anonymous mismatch of the command alternatives.
func TestAMalformedCommandIsReportedAtItsArgument(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, command, pointer, message string
	}{
		{"missing member", `{"op":"fault.clear","args":{"fault_id":"f-0"}}`, "/commands/0/args", `missing required property "at_ns"`},
		{"unknown member", `{"op":"clock.advance","args":{"to_ns":5,"await_consumer":true}}`, "/commands/0/args", `additional property "await_consumer" not allowed`},
		{"wrong type", `{"op":"entity.add","args":{"entity_id":"e","at_ns":"soon"}}`, "/commands/0/args/at_ns", ""},
		{"unknown op", `{"op":"world.create","args":{}}`, "/commands/0/op", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateScenario([]byte(scenarioHead + tc.command + `]}`))
			var invalid *ValidationError
			if !errors.As(err, &invalid) {
				t.Fatalf("ValidateScenario = %v, want a violation", err)
			}
			for _, issue := range invalid.Issues {
				if issue.Pointer == tc.pointer && strings.Contains(issue.Message, tc.message) && !strings.Contains(issue.Message, "oneOf") {
					return
				}
			}
			t.Fatalf("issues = %+v, want one at %s containing %q", invalid.Issues, tc.pointer, tc.message)
		})
	}
}

func TestEveryCommandOpIsAccepted(t *testing.T) {
	t.Parallel()
	commands := `{"op":"clock.advance","args":{"to_ns":5}},` +
		`{"op":"fault.inject","args":{"entity_id":"e","fault":"f","onset_ns":1,"params":null}},` +
		`{"op":"fault.clear","args":{"fault_id":"f-0","at_ns":2}},` +
		`{"op":"perturb.apply","args":{"perturbation":"drop","params":{"rate":0.1},"from_ns":1,"until_ns":0}},` +
		`{"op":"perturb.clear","args":{"perturb_id":"p-0"}},` +
		`{"op":"effector.invoke","args":{"effector":"a","entity_id":"e","command_id":"c","args":{},"at_ns":3}},` +
		`{"op":"entity.add","args":{"entity_id":"e2","at_ns":4}},` +
		`{"op":"entity.retire","args":{"entity_id":"e2","reason":"done","at_ns":5}}`
	if err := ValidateScenario([]byte(scenarioHead + commands + `]}`)); err != nil {
		t.Fatalf("ValidateScenario = %v", err)
	}
}

// A document the schema admits but a scenario cannot hold is a located
// violation of the scenario contract, not a decoding failure.
func TestAValueTooLargeForAScenarioIsALocatedViolation(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"/seed":                `{"schema_version":"0.1","id":"s","domain":"aquaculture-pond","adapter":"native-jsonl","seed":1e30,"world":{"start_time_ns":1},"commands":[]}`,
		"/world/start_time_ns": `{"schema_version":"0.1","id":"s","domain":"aquaculture-pond","adapter":"native-jsonl","seed":1,"world":{"start_time_ns":1.7e18},"commands":[]}`,
	}
	for pointer, raw := range cases {
		t.Run(pointer, func(t *testing.T) {
			t.Parallel()
			_, err := DecodeScenario([]byte(raw))
			var invalid *ValidationError
			if !errors.As(err, &invalid) || invalid.Contract != ContractScenario || invalid.Issues[0].Pointer != pointer {
				t.Fatalf("DecodeScenario = %+v, want a scenario violation at %s", err, pointer)
			}
		})
	}
	scenario, err := DecodeScenario([]byte(scenarioHead + `]}`))
	if err != nil || scenario.ID != "s" || scenario.World.StartTimeNS != 1 {
		t.Fatalf("DecodeScenario = %+v, %v", scenario, err)
	}
}

// Ids are bounded and free of control characters, so a scenario cannot
// smuggle a NUL or an unbounded string into a run.
func TestScenarioIdsAreBoundedPrintableText(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"control character": `{"op":"entity.add","args":{"entity_id":"pod\u0000x","at_ns":1}}`,
		"too long":          `{"op":"entity.add","args":{"entity_id":"` + strings.Repeat("x", 257) + `","at_ns":1}}`,
	}
	for name, command := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var invalid *ValidationError
			if err := ValidateScenario([]byte(scenarioHead + command + `]}`)); !errors.As(err, &invalid) ||
				invalid.Issues[0].Pointer != "/commands/0/args/entity_id" {
				t.Fatalf("ValidateScenario = %v, want the entity id refused", err)
			}
		})
	}
	if err := ValidateScenario([]byte(scenarioHead + `{"op":"entity.add","args":{"entity_id":"` + strings.Repeat("x", 256) + `","at_ns":1}}]}`)); err != nil {
		t.Fatalf("a 256-character id refused: %v", err)
	}
}
