package domain

import (
	"testing"
)

func setCommand(k map[string]any) map[string]any { return chosen(k, "commands", "set_setpoint") }

func firstDrive(k map[string]any) map[string]any {
	return setCommand(k)["drives"].([]any)[0].(map[string]any)
}

func withoutResponse(k map[string]any) {
	command := setCommand(k)
	command["response"] = "none"
	delete(command, "sets")
	delete(command, "drives")
}

func withSecondDrive(k map[string]any) {
	second := map[string]any{"from": "setpoint", "kind": "heater", "metric": "temperature", "gain": 2, "time_constant_s": 60}
	setCommand(k)["drives"] = append(setCommand(k)["drives"].([]any), second)
}

func TestAResponseThatDoesNotFitItsCommandIsRefusedAtTheMemberThatIsWrong(t *testing.T) {
	t.Parallel()
	const set, ping, drive = "/kinds/0/commands/set_setpoint", "/kinds/0/commands/ping", "/kinds/0/commands/set_setpoint/drives/0"
	cases := []struct {
		name     string
		change   func(k map[string]any)
		sentinel error
		pointer  string
	}{
		{"direct with drives", func(k map[string]any) { setCommand(k)["response"] = "direct" }, ErrResponseMismatch, set},
		{"first_order without drives", func(k map[string]any) { delete(setCommand(k), "drives") }, ErrResponseMismatch, set},
		{"none with sets", func(k map[string]any) { chosen(k, "commands", "ping")["sets"] = map[string]any{"setpoint": "setpoint"} }, ErrResponseMismatch, ping},
		{"an argument the command lacks", func(k map[string]any) { setCommand(k)["sets"] = map[string]any{"ghost": "setpoint"} }, ErrUnknownChoice, set + "/sets/ghost"},
		{"a metric the kind lacks", func(k map[string]any) { setCommand(k)["sets"] = map[string]any{"setpoint": "ghost"} }, ErrUnknownChoice, set + "/sets/setpoint"},
		{"a read-only metric", func(k map[string]any) { setCommand(k)["sets"] = map[string]any{"setpoint": "power"} }, ErrResponseMismatch, set + "/sets/setpoint"},
		{"a writable metric no command sets", withoutResponse, ErrResponseMismatch, "/kinds/0/metrics/setpoint"},
		{"a drive from a metric the command does not set", func(k map[string]any) { firstDrive(k)["from"] = "power" }, ErrResponseMismatch, drive + "/from"},
		{"a drive with no gain", func(k map[string]any) { firstDrive(k)["gain"] = 0 }, ErrResponseMismatch, drive + "/gain"},
		{"a drive into a metric a command sets", func(k map[string]any) { firstDrive(k)["metric"] = "setpoint" }, ErrResponseMismatch, drive + "/metric"},
		{"a drive into a kind that is not chosen", func(k map[string]any) { firstDrive(k)["kind"] = "ghost" }, ErrUnknownChoice, drive + "/kind"},
		{"a drive into a metric the kind lacks", func(k map[string]any) { firstDrive(k)["metric"] = "ghost" }, ErrUnknownChoice, drive + "/metric"},
		{"two drives that disagree on the time constant", withSecondDrive, ErrResponseMismatch, set + "/drives/1/time_constant_s"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			wantTwinRefusal(t, kindDocuments(t, heaterKindJSON), heaterChoices(t, tc.change), tc.sentinel, tc.pointer)
		})
	}
}

const valveKindJSON = `{
  "kind": "valve",
  "metrics": [{"name": "flow", "data_type": "number", "unit": "kW", "eu_range": {"min": 0, "max": 100}, "resolution": 1, "writable": false}],
  "commands": [{"name": "halt", "risk_class": "low", "safe_stop": true, "args_schema": {"type": "object", "additionalProperties": false, "required": [], "properties": {}}}]
}`

const valveChoicesJSON = `{
  "schema_version": "0.1", "id": "valve-twin", "version": "0.1.0", "title": "Valve twin",
  "kinds": [{"kind": "valve", "instance": "valve-1", "metrics": {"flow": {"instrument": "none"}}}]
}`

func TestASafeStopOfAKindWhoseCommandsSetNothingIsRefused(t *testing.T) {
	t.Parallel()
	wantTwinRefusal(t, kindDocuments(t, valveKindJSON), valveChoicesJSON, ErrResponseMismatch, "/kinds/0/commands/halt")
}

func TestKindsThatShareAnInstanceOrAMetricNameAreRefused(t *testing.T) {
	t.Parallel()
	zoneCopy := mutate(t, zoneKindJSON, func(doc map[string]any) { doc["kind"] = "zone_two" })
	cases := []struct {
		name    string
		kinds   []string
		change  func(doc map[string]any)
		pointer string
	}{
		{"one instance name", []string{zoneKindJSON, fanKindJSON}, func(doc map[string]any) { doc["kinds"].([]any)[0].(map[string]any)["instance"] = "fan-1" }, "/kinds/1/instance"},
		{"one metric name", []string{zoneKindJSON, zoneCopy, fanKindJSON}, addZoneTwo, "/kinds/1/kind"},
		{"one kind chosen twice", []string{zoneKindJSON, fanKindJSON}, chooseZoneTwice, "/kinds/1/kind"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			choices := mutate(t, fanZoneChoicesJSON, tc.change)
			wantTwinRefusal(t, kindDocuments(t, tc.kinds...), choices, ErrNameCollision, tc.pointer)
		})
	}
}

func addZoneTwo(doc map[string]any) {
	kinds := doc["kinds"].([]any)
	second := map[string]any{"kind": "zone_two", "instance": "zone-2", "metrics": map[string]any{"temp": map[string]any{"instrument": "none"}}}
	doc["kinds"] = []any{kinds[0], second, kinds[1]}
}

func chooseZoneTwice(doc map[string]any) {
	kinds := doc["kinds"].([]any)
	again := map[string]any{"kind": "zone", "instance": "zone-2", "metrics": map[string]any{"temp": map[string]any{"instrument": "none"}}}
	doc["kinds"] = []any{kinds[0], again, kinds[1]}
}
