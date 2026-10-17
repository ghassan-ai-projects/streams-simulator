package domain

import (
	"bytes"
	"encoding/json"
	"testing"
)

const heaterKindJSON = `{
  "kind": "heater",
  "metrics": [
    {"name": "temperature", "data_type": "number", "unit": "degC", "eu_range": {"min": 0, "max": 200}, "resolution": 0.5, "writable": false},
    {"name": "power", "data_type": "number", "unit": "kW", "eu_range": {"min": 0, "max": 100}, "resolution": 1, "writable": false},
    {"name": "setpoint", "data_type": "integer", "unit": "degC", "eu_range": {"min": 10, "max": 90}, "resolution": 1, "writable": true}
  ],
  "commands": [
    {"name": "set_setpoint", "risk_class": "medium", "safe_stop": false, "args_schema": {"type": "object", "additionalProperties": false,
      "required": ["setpoint", "lease_s"], "properties": {"setpoint": {"type": "integer", "minimum": 10, "maximum": 90},
      "lease_s": {"type": "integer", "minimum": 5, "maximum": 60}}}},
    {"name": "ping", "risk_class": "low", "safe_stop": false, "args_schema": {"type": "object", "additionalProperties": false, "required": [], "properties": {}}},
    {"name": "halt", "risk_class": "low", "safe_stop": true, "args_schema": {"type": "object", "additionalProperties": false, "required": [], "properties": {}}}
  ]
}`

const heaterChoicesJSON = `{
  "schema_version": "0.1", "id": "heater-twin", "version": "0.1.0", "title": "Heater twin", "description": "A heater for the scaffold rules.",
  "site": {"name": "lab", "area": "bench"}, "board": {"name": "board-1"},
  "kinds": [{
    "kind": "heater", "instance": "heater-01",
    "metrics": {
      "temperature": {"instrument": "acme/probe", "initial": 20},
      "power": {"instrument": "acme/meter", "initial": 10},
      "setpoint": {"instrument": "none"}
    },
    "commands": {
      "set_setpoint": {"response": "first_order", "sets": {"setpoint": "setpoint"},
        "drives": [{"from": "setpoint", "kind": "heater", "metric": "temperature", "gain": 1, "time_constant_s": 120}]},
      "ping": {"response": "none"}
    }
  }]
}`

const zoneKindJSON = `{
  "kind": "zone",
  "metrics": [{"name": "temp", "data_type": "number", "unit": "degC", "eu_range": {"min": 0, "max": 100}, "resolution": 1, "writable": false}],
  "commands": []
}`

const fanKindJSON = `{
  "kind": "fan",
  "metrics": [{"name": "duty", "data_type": "integer", "unit": "permille", "eu_range": {"min": 0, "max": 600}, "resolution": 1, "writable": true}],
  "commands": [
    {"name": "set_duty", "risk_class": "medium", "safe_stop": false, "args_schema": {"type": "object", "additionalProperties": false,
      "required": ["duty", "lease_ms"], "properties": {"duty": {"type": "integer", "minimum": 0, "maximum": 600},
      "lease_ms": {"type": "integer", "minimum": 5000, "maximum": 10000}}}},
    {"name": "stop_fan", "risk_class": "low", "safe_stop": true, "args_schema": {"type": "object", "additionalProperties": false, "required": [], "properties": {}}}
  ]
}`

const fanZoneChoicesJSON = `{
  "schema_version": "0.1", "id": "fan-zone-twin", "version": "0.1.0", "title": "Fan and zone twin",
  "kinds": [
    {"kind": "zone", "instance": "zone-1", "metrics": {"temp": {"instrument": "acme/probe", "initial": 30}}},
    {"kind": "fan", "instance": "fan-1", "metrics": {"duty": {"instrument": "none"}},
      "commands": {"set_duty": {"response": "first_order", "sets": {"duty": "duty"},
        "drives": [{"from": "duty", "kind": "zone", "metric": "temp", "gain": -0.02, "time_constant_s": 100}]}}}
  ]
}`

func kindDocuments(t *testing.T, raws ...string) map[string]map[string]any {
	t.Helper()
	documents := map[string]map[string]any{}
	for _, raw := range raws {
		var doc map[string]any
		decoder := json.NewDecoder(bytes.NewReader([]byte(raw)))
		decoder.UseNumber()
		if err := decoder.Decode(&doc); err != nil {
			t.Fatal(err)
		}
		documents[doc["kind"].(string)] = doc
	}
	return documents
}

func heaterKinds(t *testing.T, change func(doc map[string]any)) map[string]map[string]any {
	t.Helper()
	return kindDocuments(t, mutate(t, heaterKindJSON, change))
}

func twinDocument(t *testing.T, composed *Composed) map[string]any {
	t.Helper()
	var doc map[string]any
	decoder := json.NewDecoder(bytes.NewReader(composed.Bytes))
	decoder.UseNumber()
	if err := decoder.Decode(&doc); err != nil {
		t.Fatal(err)
	}
	return doc
}

func twinOK(t *testing.T, kinds map[string]map[string]any, choices string) (*Composed, map[string]any) {
	t.Helper()
	composed, err := ScaffoldTwin(baseLibrary(t), kinds, []byte(choices), "test.twin.json")
	if err != nil {
		t.Fatal(err)
	}
	return composed, twinDocument(t, composed)
}

func heaterTwin(t *testing.T) map[string]any {
	t.Helper()
	_, doc := twinOK(t, kindDocuments(t, heaterKindJSON), heaterChoicesJSON)
	return doc
}

func fanZoneTwin(t *testing.T) map[string]any {
	t.Helper()
	_, doc := twinOK(t, kindDocuments(t, zoneKindJSON, fanKindJSON), fanZoneChoicesJSON)
	return doc
}
