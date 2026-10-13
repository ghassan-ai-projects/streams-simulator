package domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const vocabularyJSON = `{
  "schema_version": "0.1",
  "unit_families": {
    "temperature": {"base": "degC", "units": {"degC": {"scale": 1}, "degF": {"scale": 0.5555555555555556, "offset": -17.77777777777778}, "K": {"scale": 1, "offset": -273.15}}},
    "power": {"base": "kW", "units": {"kW": {"scale": 1}, "W": {"scale": 0.001}}},
    "state": {"base": "on", "units": {"on": {"scale": 1}}}
  },
  "port_types": {"thermal_load": {"unit_family": "power"}, "heat_rejection": {"unit_family": "power"}},
  "tag_letters": {"TT": {"measures": ["temperature"]}, "JT": {"measures": ["power"]}, "YS": {"measures": ["state"]}}
}`

const kilnJSON = `{
  "schema_version": "0.1", "id": "acme/kiln", "version": 1, "title": "Kiln", "description": "library text that must not leak",
  "class": "static/furnace/kiln", "tags": ["thermal"],
  "kind": {
    "state": [
      {"name": "running", "unit": "on", "initial": 1, "min": 0, "max": 1},
      {"name": "setpoint", "unit": "degC", "initial": 100, "min": 0, "max": 200, "description": "state text"},
      {"name": "temp", "unit": "degC", "initial": 100},
      {"name": "load", "unit": "kW", "initial": 50, "min": 0},
      {"name": "heat", "unit": "kW", "initial": 40, "min": 0}
    ],
    "dynamics": [
      {"target": "load", "tier": "F0", "f0": {"baseline": 50, "drift_per_hour": 0, "seasonality": [{"amplitude": 10, "period_s": 86400, "phase_rad": 0}]}},
      {"target": "heat", "tier": "F0", "f0": {"baseline": 40, "drift_per_hour": 0, "seasonality": []}},
      {"target": "temp", "tier": "F1", "dt_ms": 5000, "f1": {"form": "first_order_lag", "time_constant_s": 300, "gain": 1,
        "inputs": [{"state": "setpoint", "coef": 1}, {"state": "load", "coef": 0.1}]}}
    ],
    "faults": [{"id": "fouling", "description": "fault text", "affects": [{"state": "temp", "delta": 5}],
      "observability": {"detector": {"form": "single_channel_snr", "channel": "temp"}, "first_observable_snr": 3, "unavoidable_snr": 10},
      "onset": {"magnitude": 1, "shape": "step"}}],
    "effectors": [{"name": "set_setpoint", "description": "effector text", "risk_class": "medium", "idempotency_window_s": 60,
      "ack": {"failure_modes": [{"mode": "ok", "probability": 1}], "latency_ms": {"mean": 100, "sigma": 10}},
      "args_schema": {"type": "object", "additionalProperties": false, "required": ["kiln_id", "setpoint"],
        "properties": {"kiln_id": {"type": "string"}, "setpoint": {"type": "number", "minimum": 0, "maximum": 200}}},
      "effect": {"dead_time_s": 0, "time_constant_s": 0, "state_deltas": [{"state": "setpoint", "assign_from_arg": "setpoint"}]}}]
  },
  "measurement_points": [
    {"name": "temp", "state": "temp", "unit_family": "temperature", "range": {"min": 0, "max": 300}, "tag": "TT", "default_instrument": "acme/probe@1"},
    {"name": "power", "state": "load", "unit_family": "power", "range": {"min": 0, "max": 100}, "tag": "JT", "default_instrument": "acme/meter@1"},
    {"name": "heat_flow", "state": "heat", "unit_family": "power", "range": {"min": 0, "max": 80}, "tag": "JT", "default_instrument": "acme/meter@1", "optional": true}
  ],
  "ports": [
    {"name": "load_in", "type": "thermal_load", "direction": "consumes", "state": "load"},
    {"name": "heat_out", "type": "heat_rejection", "direction": "provides", "state": "heat"}
  ],
  "default_preset": "std",
  "presets": {
    "std": {"title": "Standard", "values": {}},
    "big": {"title": "Big", "values": {"state/load/initial": 500, "dynamics/load/f0/baseline": 500,
      "dynamics/load/f0/seasonality/0/amplitude": 100, "measurement_points/power/range/max": 1000}}
  },
  "parameters": [
    {"name": "setpoint_c", "title": "Setpoint", "unit": "degC", "type": "number", "min": 50, "max": 150, "default": 100,
      "targets": ["state/setpoint/initial", "state/temp/initial"]},
    {"name": "insulation", "title": "Insulation", "type": "enum", "values": ["good", "poor"], "default": "good",
      "cases": {"good": {"dynamics/temp/f1/time_constant_s": 300}, "poor": {"dynamics/temp/f1/time_constant_s": 120}}},
    {"name": "stages", "title": "Stages", "type": "integer", "min": 0, "max": 4, "default": 0, "targets": ["state/load/min"]}
  ]
}`

const probeJSON = `{
  "schema_version": "0.1", "id": "acme/probe", "version": 1, "title": "Probe", "tag": "TT", "unit_family": "temperature", "unit": "degC",
  "range": {"min": -50, "max": 400},
  "channel": {"value_type": "number", "resolution": 0.1, "fidelity": "F1", "absence": "signal",
    "cadence": {"mode": "periodic", "period_s": 10}, "noise": {"model": "gaussian", "sigma": 0.05}},
  "faults": [{"id": "offset", "form": "offset", "magnitude": 2.0}, {"id": "drift", "form": "drift", "magnitude": 0.05, "duration_s": 86400}]
}`

const meterJSON = `{
  "schema_version": "0.1", "id": "acme/meter", "version": 1, "title": "Meter", "tag": "JT", "unit_family": "power", "unit": "kW",
  "channel": {"value_type": "number", "resolution": 0.1, "fidelity": "F1", "absence": "signal",
    "cadence": {"mode": "periodic", "period_s": 15}, "noise": {"model": "gaussian", "sigma": 0.01}},
  "faults": [{"id": "offset", "form": "offset", "magnitude": 0.05, "is_negative_class": true}],
  "scaling": "span"
}`

const compositionJSON = `{
  "schema_version": "0.1", "id": "test-plant", "version": "0.1.0", "title": "Test plant", "description": "A composition for tests.",
  "stresses": "a small plant composed from a test library for the compose rules",
  "axes": {"absence": "signal", "actuation": "advisory", "cadence": ["periodic"], "cardinality": "small", "consequence": ["cost"],
    "correlation": ["independent"], "fidelity": ["F0", "F1"], "lateness": "none", "rate": "low", "seasonality": ["diurnal"],
    "time_reference": "wall", "value_shape": ["scalar"]},
  "equipment": [{"use": "acme/kiln@1", "names": ["K-1"]}],
  "profiles": [{"name": "nominal", "duration_s": 3600},
    {"name": "sensor_pathology", "not_applicable": "the test plant weighs no fault"},
    {"name": "correlated_cascade", "not_applicable": "the test plant has no cascade"}],
  "ground_truth": {"min_per_label": 10, "min_scenarios": 100, "negative_class_fraction": 0.35, "pre_degraded_fraction": 0.0}
}`

func mutate(t *testing.T, raw string, change func(doc map[string]any)) string {
	t.Helper()
	var doc map[string]any
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&doc); err != nil {
		t.Fatal(err)
	}
	change(doc)
	out, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func file(path, raw string) File { return File{Path: path, Raw: []byte(raw)} }

type testFiles struct {
	vocabulary  string
	equipment   []File
	instruments []File
}

func baseFiles() testFiles {
	return testFiles{
		vocabulary:  vocabularyJSON,
		equipment:   []File{file("equipment/kiln.component.json", kilnJSON)},
		instruments: []File{file("instruments/probe.instrument.json", probeJSON), file("instruments/meter.instrument.json", meterJSON)},
	}
}

func under(path string, files []File) []File {
	out := make([]File, len(files))
	for i, entry := range files {
		out[i] = File{Path: path + "/" + entry.Path, Raw: entry.Raw}
	}
	return out
}

func (f testFiles) directory(path string) Directory {
	directory := Directory{Path: path, Equipment: under(path, f.equipment), Instruments: under(path, f.instruments)}
	if f.vocabulary != "" {
		directory.Vocabulary = &File{Path: path + "/vocabulary.json", Raw: []byte(f.vocabulary)}
	}
	return directory
}

func loadLibrary(t *testing.T, directories ...Directory) *Library {
	t.Helper()
	library, err := LoadLibrary(directories)
	if err != nil {
		t.Fatal(err)
	}
	return library
}

func baseLibrary(t *testing.T) *Library {
	t.Helper()
	return loadLibrary(t, baseFiles().directory("lib"))
}

func libraryWith(t *testing.T, extra ...File) *Library {
	t.Helper()
	files := baseFiles()
	files.instruments = append(files.instruments, extra...)
	return loadLibrary(t, files.directory("lib"))
}

func asViolation(t *testing.T, err error, contract model.Contract) *model.ValidationError {
	t.Helper()
	var invalid *model.ValidationError
	if !errors.As(err, &invalid) {
		t.Fatalf("error = %v, want a validation error of %s", err, contract)
	}
	if invalid.Contract != contract {
		t.Fatalf("contract = %s, want %s: %v", invalid.Contract, contract, err)
	}
	return invalid
}

func wantViolation(t *testing.T, err error, contract model.Contract) {
	t.Helper()
	_ = asViolation(t, err, contract)
}

func wantIssue(t *testing.T, err error, contract model.Contract, pointerWant, messagePart string) {
	t.Helper()
	invalid := asViolation(t, err, contract)
	for _, issue := range invalid.Issues {
		if issue.Pointer == pointerWant && strings.Contains(issue.Message, messagePart) {
			return
		}
	}
	t.Fatalf("issues = %+v, want one at %q containing %q", invalid.Issues, pointerWant, messagePart)
}

func compose(t *testing.T, library *Library, raw string) (map[string]any, error) {
	t.Helper()
	out, err := Compose(library, []byte(raw), "test.composition.json")
	if err != nil {
		return nil, err
	}
	var doc map[string]any
	decoder := json.NewDecoder(bytes.NewReader(out))
	decoder.UseNumber()
	if err := decoder.Decode(&doc); err != nil {
		t.Fatal(err)
	}
	return doc, nil
}

func composeOK(t *testing.T, library *Library, raw string) map[string]any {
	t.Helper()
	doc, err := compose(t, library, raw)
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func withEquipment(t *testing.T, entries ...map[string]any) string {
	t.Helper()
	return mutate(t, compositionJSON, func(doc map[string]any) {
		list := make([]any, len(entries))
		for i, entry := range entries {
			list[i] = entry
		}
		doc["equipment"] = list
	})
}

func kilnEntry(extra map[string]any) map[string]any {
	entry := map[string]any{"use": "acme/kiln@1", "names": []any{"K-1"}}
	for key, value := range extra {
		entry[key] = value
	}
	return entry
}

func member(t *testing.T, doc any, path ...any) any {
	t.Helper()
	current := doc
	for _, step := range path {
		switch typed := step.(type) {
		case string:
			object, _ := current.(map[string]any)
			current = object[typed]
		case int:
			list, _ := current.([]any)
			if typed >= len(list) {
				t.Fatalf("no item %d in %v", typed, path)
			}
			current = list[typed]
		}
	}
	return current
}

func named(t *testing.T, doc any, section, key, value string) map[string]any {
	t.Helper()
	list, _ := doc.([]any)
	for _, item := range list {
		if entry, _ := item.(map[string]any); entry[key] == value {
			return entry
		}
	}
	t.Fatalf("no %s with %s %q in %v", section, key, value, list)
	return nil
}

func number(t *testing.T, value any) float64 {
	t.Helper()
	parsed, err := value.(json.Number).Float64()
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}
