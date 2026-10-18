package acceptance

// The one-pass device checks through the real binary: a kind document checked
// from the page without writing a file, the binary's own default choices for
// the workspace kinds, and the instruments of the library by unit.

import (
	"encoding/json"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

const twoReadingsMissingResolution = `{"kind":"probe","metrics":[
 {"name":"temperature","data_type":"number","unit":"Cel","eu_range":{"min":0,"max":50},"writable":false},
 {"name":"humidity","data_type":"number","unit":"%","eu_range":{"min":0,"max":100},"writable":false}],
 "commands":[{"name":"halt","args_schema":{"type":"object","properties":{},"required":[],"additionalProperties":false},"risk_class":"low"}]}`

func TestTheCanvasBridgeChecksAKindDocumentWithoutWritingAFileAndNamesEveryMissingMember(t *testing.T) {
	t.Parallel()
	client, _ := startDeviceBridge(t)
	plain := client.action("kind.check", map[string]any{"content": twoReadingsMissingResolution})
	if kinds, _ := plain["kinds"].([]any); plain["valid"] != true || len(kinds) != 1 {
		t.Fatalf("the plant wire schema accepts the document: %v", plain)
	}
	failure := refusedAction(t, client, "kind.check", map[string]any{"content": twoReadingsMissingResolution, "twin": true})
	issues, _ := failure["issues"].([]any)
	got := map[string]bool{}
	for _, raw := range issues {
		issue, _ := raw.(map[string]any)
		pointer, _ := issue["pointer"].(string)
		keyword, _ := issue["keyword"].(string)
		member, _ := issue["member"].(string)
		got[pointer+" "+keyword+" "+member] = true
	}
	for _, want := range []string{"/metrics/0/resolution twin_member_required resolution", "/metrics/1/resolution twin_member_required resolution",
		"/commands/0/safe_stop twin_member_required safe_stop"} {
		if !got[want] {
			t.Errorf("issues %v lack %q", got, want)
		}
	}
	if failure["code"] != "kind_failed" || len(listed(t, client, "kinds")) != 0 {
		t.Fatalf("a refused check = %v and must leave the workspace without kind files", failure)
	}
}

func TestTheCanvasBridgeSuggestsChoicesThatTheTwinScaffoldAccepts(t *testing.T) {
	t.Parallel()
	client, workspace := startDeviceBridge(t)
	putKindFiles(t, client)
	suggested := client.action("kind.suggest", nil)
	choices, _ := suggested["choices"].(map[string]any)
	raw, err := json.Marshal(choices)
	if err != nil || choices["id"] == nil {
		t.Fatalf("kind.suggest = %v, %v", suggested, err)
	}
	putFile(t, client, "twins/suggested.twin.json", string(raw))
	twin := client.action("kind.twin", map[string]any{"choices": "twins/suggested.twin.json", "out": testsupport.WorkspaceDomain("suggested"), "etag": ""})
	if twin["id"] != choices["id"] || twin["composed"] == nil {
		t.Fatalf("kind.twin of the suggested choices = %v (workspace %s)", twin, workspace)
	}
}

func TestTheCanvasBridgeListsTheInstrumentsOfAUnitExactUnitFirst(t *testing.T) {
	t.Parallel()
	client, _ := startDeviceBridge(t)
	listing := client.action("component.list", map[string]any{"unit": "Cel"})
	entries, _ := listing["components"].([]any)
	if len(entries) == 0 {
		t.Fatalf("no instrument measures a temperature in Cel: %v", listing)
	}
	for _, raw := range entries {
		entry, _ := raw.(map[string]any)
		if entry["kind"] != "instrument" || entry["unit"] == nil || entry["range"] == nil || entry["match"] != "convertible" {
			t.Errorf("entry = %v, want an instrument with its unit, range and a convertible match", entry)
		}
	}
	if none := client.action("component.list", map[string]any{"unit": "furlong"}); len(none["components"].([]any)) != 0 {
		t.Fatalf("an unknown unit lists %v", none)
	}
}
