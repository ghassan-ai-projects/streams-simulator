package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const gaugeKindJSON = `{
  "kind": "gauge",
  "metrics": [
    {"name": "ready", "data_type": "boolean", "resolution": 1, "writable": false},
    {"name": "level", "data_type": "number", "unit": "kW", "resolution": 1, "writable": false},
    {"name": "fill", "data_type": "number", "unit": "kW", "eu_range": {"min": 0, "max": 10}, "resolution": 1, "writable": false}
  ],
  "commands": []
}`

const gaugeChoicesJSON = `{
  "schema_version": "0.1", "id": "gauge-twin", "version": "0.1.0", "title": "Gauge twin",
  "kinds": [{"kind": "gauge", "instance": "g-1", "metrics": {
    "ready": {"instrument": "none"}, "level": {"instrument": "none"}, "fill": {"instrument": "none"}}}]
}`

const tankKindJSON = `{
  "kind": "tank",
  "metrics": [{"name": "level", "data_type": "number", "unit": "kW", "resolution": 1, "writable": false}],
  "commands": []
}`

const tankFanChoicesJSON = `{
  "schema_version": "0.1", "id": "tank-fan-twin", "version": "0.1.0", "title": "Tank and fan twin",
  "kinds": [
    {"kind": "tank", "instance": "tank-1", "metrics": {"level": {"instrument": "none", "initial": 3}}},
    {"kind": "fan", "instance": "fan-1", "metrics": {"duty": {"instrument": "none"}},
      "commands": {"set_duty": {"response": "first_order", "sets": {"duty": "duty"},
        "drives": [{"from": "duty", "kind": "tank", "metric": "level", "gain": 0.5, "time_constant_s": 50}]}}}
  ]
}`

const quietInstrumentJSON = `{
  "schema_version": "0.1", "id": "acme/quiet", "version": 1, "title": "Quiet", "tag": "TT", "unit_family": "temperature", "unit": "degC",
  "range": {"min": -50, "max": 400},
  "channel": {"value_type": "number", "resolution": 0.1, "fidelity": "F1", "absence": "signal",
    "cadence": {"mode": "periodic", "period_s": 10}, "noise": {"model": "gaussian", "sigma": 0.05}}
}`

const quietChoicesJSON = `{
  "schema_version": "0.1", "id": "quiet-twin", "version": "0.1.0", "title": "Quiet twin",
  "kinds": [{"kind": "zone", "instance": "zone-1", "metrics": {"temp": {"instrument": "acme/quiet"}}}]
}`

func TestAKindOfBooleanAndUnrangedReadingsGetsPlainChannelsAndDefaultValues(t *testing.T) {
	t.Parallel()
	doc := twinDocumentOf(t, kindDocuments(t, gaugeKindJSON), gaugeChoicesJSON)
	kind := member(t, doc, "kinds", 0)
	ready := named(t, member(t, kind, "channels"), "channel", "name", "ready")
	if member(t, ready, "value_type") != "boolean" || member(t, ready, "unit") != nil || member(t, ready, "range") != nil {
		t.Errorf("ready channel = %v, want a boolean channel without unit or range", ready)
	}
	states := member(t, kind, "state")
	if got := number(t, member(t, named(t, states, "state", "name", "fill_value"), "initial")); got != 5 {
		t.Errorf("fill initial = %v, want the middle of its range", got)
	}
	if got := number(t, member(t, named(t, states, "state", "name", "level_value"), "initial")); got != 0 {
		t.Errorf("level initial = %v, want 0 for a reading without a range", got)
	}
	if member(t, kind, "dynamics") != nil || member(t, kind, "effectors") != nil {
		t.Errorf("kind = %v, want no dynamics and no effectors", kind)
	}
	axes := member(t, doc, "axes")
	if member(t, axes, "correlation", 0) != "independent" || member(t, axes, "actuation") != "observe_only" || member(t, axes, "value_shape", 0) != "boolean" {
		t.Errorf("axes = %v, want independent readings, observe only, boolean first", axes)
	}
}

func twinDocumentOf(t *testing.T, kinds map[string]map[string]any, choices string) map[string]any {
	t.Helper()
	_, doc := twinOK(t, kinds, choices)
	return doc
}

func TestAReadingWithoutARangeFollowsItsDriveWithoutAClamp(t *testing.T) {
	t.Parallel()
	doc := twinDocumentOf(t, kindDocuments(t, tankKindJSON, fanKindJSON), tankFanChoicesJSON)
	dynamics := member(t, doc, "kinds", 0, "dynamics", 0)
	level := named(t, member(t, doc, "kinds", 0, "state"), "state", "name", "level_value")
	if member(t, dynamics, "f1", "clamp") != nil || member(t, level, "min") != nil || member(t, level, "max") != nil {
		t.Fatalf("dynamics %v and state %v, want neither clamp nor bounds", dynamics, level)
	}
	if member(t, doc, "link_types", 0, "name") != "fan_drives_tank" {
		t.Fatalf("link types = %v, want fan_drives_tank", member(t, doc, "link_types"))
	}
}

func TestAnInstrumentWithoutFaultsStillLeavesTheKindAFaultToInject(t *testing.T) {
	t.Parallel()
	library := libraryWith(t, file("instruments/quiet.instrument.json", quietInstrumentJSON))
	composed, err := ScaffoldTwin(library, kindDocuments(t, zoneKindJSON), []byte(quietChoicesJSON), "quiet.twin.json")
	if err != nil {
		t.Fatal(err)
	}
	named(t, member(t, twinDocument(t, composed), "kinds", 0, "faults"), "fault", "id", "zone__temp__dropout")
}

func TestATwinThatThePhysicsOfTheDomainRefusesIsReportedAsADomainViolation(t *testing.T) {
	t.Parallel()
	choices := heaterChoices(t, func(k map[string]any) { firstDrive(k)["time_constant_s"] = 0.1 })
	_, err := ScaffoldTwin(baseLibrary(t), kindDocuments(t, heaterKindJSON), []byte(choices), "test.twin.json")
	invalid := asViolation(t, err, model.ContractDomainSpec)
	if len(invalid.Issues) == 0 || invalid.Issues[0].Pointer != "/kinds/0/dynamics/0/dt_ms" {
		t.Fatalf("ScaffoldTwin = %v, want a domain violation at /kinds/0/dynamics/0/dt_ms", err)
	}
}

func TestMoreKindFilesThatATwinCannotReproduceAreRefused(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		change func(doc map[string]any)
	}{
		{"an inverted range", func(doc map[string]any) { fileMetric(doc)["eu_range"] = map[string]any{"min": 5, "max": 5} }},
		{"a command name with capitals", func(doc map[string]any) { fileCommand(doc)["name"] = "SetSetpoint" }},
		{"a metric that is no object", func(doc map[string]any) { doc["metrics"] = []any{1} }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			wantTwinRefusal(t, heaterKinds(t, tc.change), heaterChoicesJSON, ErrKindNotReproducible, "/kinds/0/kind")
		})
	}
}

func TestAKindNameWithCapitalsIsRefused(t *testing.T) {
	t.Parallel()
	kinds := heaterKinds(t, func(doc map[string]any) { doc["kind"] = "Heater" })
	choices := heaterChoices(t, func(k map[string]any) { k["kind"] = "Heater" })
	wantTwinRefusal(t, kinds, choices, ErrKindNotReproducible, "/kinds/0/kind")
}

func TestACommandCannotSetAMetricThatStatesNoRange(t *testing.T) {
	t.Parallel()
	kinds := heaterKinds(t, func(doc map[string]any) { delete(doc["metrics"].([]any)[2].(map[string]any), "eu_range") })
	wantTwinRefusal(t, kinds, heaterChoicesJSON, ErrResponseMismatch, "/kinds/0/commands/set_setpoint/sets/setpoint")
}
