package domain

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func firstBoard(birth map[string]any) map[string]any {
	return birth["boards"].([]any)[0].(map[string]any)
}

func TestACaptureEchoesEachBoardsEnvironmentAsTheBoardsClaim(t *testing.T) {
	t.Parallel()
	boards := []model.Board{{Name: "simulated-1", Members: []string{"P1"}, Environment: EnvironmentSimulated}, {Name: "uno-1", Members: []string{"V1"}, Environment: EnvironmentHardware}}
	birth, err := NodeBirthOf(NodeIdentity{Environment: EnvironmentMixed, ProtocolMax: 1, PriorLastSeq: -1}, birthKinds(), boards)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(stamped(t, birth)[0])
	if err != nil {
		t.Fatal(err)
	}
	captured, err := parse(string(raw))
	want := []CapturedBoard{{Board: "simulated-1", Environment: "simulated"}, {Board: "uno-1", Environment: "hardware"}}
	if err != nil || captured.Environment != "mixed" || !reflect.DeepEqual(captured.Boards, want) {
		t.Fatalf("ParseCapture = environment %q boards %+v, %v; want a mixed node echoing %+v", captured.Environment, captured.Boards, err, want)
	}
}

func TestACaptureWithoutBoardsEchoesAnEmptyListNotNothing(t *testing.T) {
	t.Parallel()
	birth, err := NodeBirthOf(NodeIdentity{Environment: EnvironmentHardware, ProtocolMax: 1, PriorLastSeq: -1}, birthKinds(), nil)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(stamped(t, birth)[0])
	captured, err := parse(string(raw))
	document, _ := json.Marshal(captured)
	if err != nil || !strings.Contains(string(document), `"boards":[]`) {
		t.Fatalf("ParseCapture = %s, %v, want boards as an empty list", document, err)
	}
}

func TestACaptureRefusesABirthWhoseEnvironmentContradictsItsBoards(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, pointer, reason string
		edit                  func(birth map[string]any)
	}{
		{"a hardware node with a simulated board", "/boards/0/environment", `board "b-1" is "simulated"`, func(b map[string]any) { firstBoard(b)["environment"] = "simulated" }},
		{"a simulated node with a hardware board", "/boards/0/environment", `needs every board simulated`, func(b map[string]any) { b["environment"] = "simulated" }},
		{"a mixed node with one environment of boards", "/boards", "none of its 1 boards is simulated", func(b map[string]any) { b["environment"] = "mixed" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := parse(birthLine(t, tc.edit))
			invalid := violation(t, err, ErrBirthRefused)
			issue := invalid.Issues[0]
			if len(invalid.Issues) != 1 || issue.Keyword != "environment_mismatch" || issue.Pointer != tc.pointer || !strings.Contains(issue.Message, tc.reason) {
				t.Fatalf("issues = %+v, want one environment_mismatch at %s saying %q", invalid.Issues, tc.pointer, tc.reason)
			}
		})
	}
}
