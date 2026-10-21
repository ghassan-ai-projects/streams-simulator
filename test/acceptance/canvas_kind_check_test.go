package acceptance

// The one-pass device checks through the real binary: a kind document checked
// from the page without writing a file, the binary's own default choices for
// the workspace kinds, and the instruments of the library by unit.

import (
	"encoding/json"
	"net/http"
	"strings"
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
	putKindFilesFrom(t, client, testsupport.TwinThermalKindsDir())
	kinds := []any{"fan", "indicator", "thermal_sensor"}
	suggested := client.action("kind.suggest", map[string]any{"kinds": kinds})
	choices, _ := suggested["choices"].(map[string]any)
	raw, err := json.Marshal(choices)
	if err != nil || choices["id"] == nil {
		t.Fatalf("kind.suggest = %v, %v", suggested, err)
	}
	putFile(t, client, "twins/suggested.twin.json", string(raw))
	twin := client.action("kind.twin", map[string]any{"kinds": kinds, "choices": "twins/suggested.twin.json", "out": testsupport.WorkspaceDomain("suggested"), "etag": ""})
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

const unfinishedPumpMonitor = `{"kind":"pump_monitor","metrics":[{"name":"pressure","data_type":"number","writable":false}],"commands":[]}`

func TestAnUnfinishedKindOfTheWorkspaceNeverBreaksTheTwinOfAnotherDevice(t *testing.T) {
	t.Parallel()
	client, _ := startDeviceBridge(t)
	putKindFiles(t, client)
	putFile(t, client, "kinds/pump_monitor.kind.json", unfinishedPumpMonitor)
	suggested := client.action("kind.suggest", map[string]any{"kinds": []any{"fan"}})
	choices, _ := suggested["choices"].(map[string]any)
	raw, err := json.Marshal(choices)
	if err != nil || choices["id"] != "fan-twin" || choices["title"] != "Twin of fan" {
		t.Fatalf("kind.suggest of the fan beside an unfinished pump monitor = %v, %v, want a twin of the fan alone", suggested, err)
	}
	putFile(t, client, "twins/fan.twin.json", string(raw))
	out := testsupport.WorkspaceDomain("fan-only")
	twin := client.action("kind.twin", map[string]any{"kinds": []any{"fan"}, "choices": "twins/fan.twin.json", "out": out, "etag": ""})
	if implements, _ := twin["implements"].([]any); len(implements) != 1 || implements[0] != "fan" {
		t.Fatalf("kind.twin = %v, want a twin of the fan alone", twin)
	}
	if valid := client.action("domain.validate", map[string]any{"path": out}); valid["valid"] != true {
		t.Fatalf("domain.validate of the fan twin beside an unfinished pump monitor = %v", valid)
	}
}

func TestAKindThatIsInTheTwinAndNotReadyIsNamedAndAKindThatIsAbsentIsNotFound(t *testing.T) {
	t.Parallel()
	client, _ := startDeviceBridge(t)
	putKindFiles(t, client)
	putFile(t, client, "kinds/pump_monitor.kind.json", unfinishedPumpMonitor)
	failure := refusedAction(t, client, "kind.suggest", map[string]any{"kinds": []any{"fan", "pump_monitor"}})
	if message, _ := failure["message"].(string); failure["code"] != "kind_failed" || !strings.Contains(message, "pump_monitor") {
		t.Fatalf("suggest with the unfinished kind in the twin = %v, want kind_failed naming pump_monitor", failure)
	}
	status, document, _ := client.call(http.MethodPost, "/api/actions/kind.suggest", map[string]any{"args": map[string]any{"kinds": []any{"valve_monitor"}}})
	if status != http.StatusNotFound || errorCode(document) != "not_found" {
		t.Fatalf("suggest for a kind the workspace does not hold = %d %v, want 404 not_found", status, document)
	}
}

func TestTheCanvasBridgeSaysWhichInstrumentsFitAReadingWithItsRange(t *testing.T) {
	t.Parallel()
	client, _ := startDeviceBridge(t)
	fitOf := func(args map[string]any, id string) map[string]any {
		entries, _ := client.action("component.list", args)["components"].([]any)
		for _, raw := range entries {
			if entry, _ := raw.(map[string]any); entry["id"] == id {
				fit, _ := entry["fit"].(map[string]any)
				return fit
			}
		}
		t.Fatalf("component.list %v lists no %s", args, id)
		return nil
	}
	pressure := fitOf(map[string]any{"unit": "Pa", "min": 0, "max": 500000}, "streamsim/pressure-transmitter")
	flow := fitOf(map[string]any{"unit": "L/min", "min": 0, "max": 2000}, "streamsim/mag-flow-meter")
	unstated := fitOf(map[string]any{"unit": "L/min"}, "streamsim/mag-flow-meter")
	if pressure["ok"] != true || flow["ok"] != true || unstated["ok"] != false || unstated["code"] != "range_unstated" || unstated["member"] != "eu_range" {
		t.Fatalf("fits = %v, %v, %v, want the kPa transmitter and the m3/h meter to fit converted ranges and an unstated range to name eu_range", pressure, flow, unstated)
	}
}
