package domain

import (
	"encoding/json"
	"errors"
	"maps"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/jsonschema"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	embedded "github.com/ghassan-ai-projects/streams-simulator/internal/schemas"
)

func TestEveryRequestOfTheContractDecodesToItsControlEntry(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, line, op, command string
		scope                   Scope
		enabled                 any
	}{
		{"hold a board", `{"op":"hold","scope":{"kind":"board","id":"uno-07"}}`, OpHold, "", Scope{ScopeBoard, "uno-07"}, nil},
		{"release everything", `{"op":"release","scope":{"kind":"all"}}`, OpRelease, "", Scope{ScopeAll, ""}, nil},
		{"stop a device", `{"op":"estop","scope":{"kind":"device","id":"fan-01"}}`, OpEStop, "", Scope{ScopeDevice, "fan-01"}, nil},
		{"approval mode on", `{"op":"approve_mode","scope":{"kind":"all"},"enabled":true}`, OpApproveMode, "", Scope{ScopeAll, ""}, true},
		{"approval mode off", `{"op":"approve_mode","scope":{"kind":"board","id":"b"},"enabled":false}`, OpApproveMode, "", Scope{ScopeBoard, "b"}, false},
		{"approve", `{"op":"approve","command_id":"cmd-1"}`, OpApprove, "cmd-1", Scope{}, nil},
		{"deny", `{"op":"deny","command_id":"cmd-1"}`, OpDeny, "cmd-1", Scope{}, nil},
		{"state", `{"op":"state"}`, OpState, "", Scope{}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			entry := controlOf(t, tc.line)
			if entry.Event() != EventControl || entry.Op() != tc.op || entry.Text("command_id") != tc.command || entry["enabled"] != tc.enabled {
				t.Fatalf("entry = %v, want a control of %s for %q with enabled %v", entry, tc.op, tc.command, tc.enabled)
			}
			if tc.scope.Kind != "" && scopeOf(entry["scope"]) != tc.scope {
				t.Fatalf("scope = %v, want %v", entry["scope"], tc.scope)
			}
		})
	}
}

func TestARequestThatBreaksTheContractIsRefusedAsMalformedAndNamesTheContract(t *testing.T) {
	t.Parallel()
	for name, line := range map[string]string{
		"an unknown operation":          `{"op":"explode"}`,
		"no operation":                  `{"scope":{"kind":"all"}}`,
		"an unknown member":             `{"op":"state","force":true}`,
		"a hold without a scope":        `{"op":"hold"}`,
		"a scope of all with an id":     `{"op":"hold","scope":{"kind":"all","id":"b"}}`,
		"a board without an id":         `{"op":"hold","scope":{"kind":"board"}}`,
		"an id that is no wire id":      `{"op":"hold","scope":{"kind":"device","id":"fan 01"}}`,
		"approval mode without enabled": `{"op":"approve_mode","scope":{"kind":"all"}}`,
		"an approval without a command": `{"op":"approve"}`,
		"a hold carrying a command":     `{"op":"hold","scope":{"kind":"all"},"command_id":"c-1"}`,
		"a state carrying a scope":      `{"op":"state","scope":{"kind":"all"}}`,
		"not JSON":                      `hold everything`,
		"two documents":                 `{"op":"state"} {"op":"state"}`,
	} {
		_, err := DecodeControl([]byte(line))
		invalid, named := model.AsValidationError(err)
		if !errors.Is(err, ErrControlMalformed) || !named || invalid.Contract != model.ContractNodeControl {
			t.Errorf("%s: DecodeControl = %v, want ErrControlMalformed naming %s", name, err, model.ContractNodeControl)
		}
	}
}

func replySchema(t *testing.T) *jsonschema.Schema {
	t.Helper()
	var doc map[string]any
	if err := json.Unmarshal(embedded.NodeControl(), &doc); err != nil {
		t.Fatal(err)
	}
	defs, _ := doc["$defs"].(map[string]any)
	reply, _ := defs["reply"].(map[string]any)
	root := maps.Clone(reply)
	root["$defs"] = defs
	raw, err := json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	schema, err := jsonschema.CompileJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	return schema
}

func TestEveryAnswerTheNodeGivesSatisfiesTheReplyContract(t *testing.T) {
	t.Parallel()
	schema := replySchema(t)
	g := plantGate()
	apply(t, g, approveModeLine(ScopeDevice, "pump-2", true))
	judge(g, "c-1", "pump-2")
	lines := []string{holdLine(), scopeLine(OpEStop, ScopeBoard, "b-1"), scopeLine(OpRelease, ScopeBoard, "b-1"),
		approveModeLine(ScopeAll, "", true), `{"op":"approve","command_id":"c-1"}`, `{"op":"deny","command_id":"c-9"}`,
		scopeLine(OpHold, ScopeBoard, "b-9"), `{"op":"state"}`}
	replies := []map[string]any{RefusedReply(ReasonMalformed, g.View()), RefusedReply(ReasonNodeBusy, g.View()), StateReply(g.View())}
	for _, line := range lines {
		replies = append(replies, apply(t, g, line).Reply)
	}
	for _, reply := range replies {
		var doc any
		if err := model.DecodeBytes(ControlReply(reply), &doc); err != nil {
			t.Fatal(err)
		}
		if errs := schema.Validate(doc); len(errs) > 0 {
			t.Errorf("reply %s: %s", ControlReply(reply), jsonschema.FormatErrors(errs))
		}
	}
}

func holdLine() string { return scopeLine(OpHold, ScopeDevice, "valve-1") }

func TestAReplyCarriesNothingOfTheLog(t *testing.T) {
	t.Parallel()
	reply := apply(t, plantGate(), `{"op":"state"}`).Reply
	var decoded map[string]any
	if err := json.Unmarshal(ControlReply(reply), &decoded); err != nil {
		t.Fatal(err)
	}
	for _, member := range []string{"event", "clock", "node_ns"} {
		if _, present := decoded[member]; present {
			t.Errorf("reply holds %s, a member of the log", member)
		}
	}
	if _, kept := reply["event"]; !kept {
		t.Fatal("ControlReply removed a member from the entry it was given")
	}
}
