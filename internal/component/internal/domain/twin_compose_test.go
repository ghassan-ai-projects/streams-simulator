package domain

import (
	"bytes"
	"slices"
	"testing"
)

func TestATwinDeclaresTheChannelsOfItsKindFileInOrderWithTheirUnitRangeAndResolution(t *testing.T) {
	t.Parallel()
	channels := member(t, heaterTwin(t), "kinds", 0, "channels").([]any)
	want := [][]any{
		{"temperature", "degC", "number", 0.5, 0.0, 200.0},
		{"power", "kW", "number", 1.0, 0.0, 100.0},
		{"setpoint", "degC", "counter", 1.0, 10.0, 90.0},
	}
	if len(channels) != len(want) {
		t.Fatalf("channels = %d, want %d", len(channels), len(want))
	}
	for i, row := range want {
		channel := channels[i]
		got := []any{member(t, channel, "name"), member(t, channel, "unit"), member(t, channel, "value_type"),
			number(t, member(t, channel, "resolution")), number(t, member(t, channel, "range", "min")), number(t, member(t, channel, "range", "max"))}
		if !slices.Equal(got, row) {
			t.Errorf("channel %d = %v, want %v (name, unit, value type, resolution, range)", i, got, row)
		}
	}
}

func TestATwinTakesNoiseCadenceTagAndFaultsFromTheInstrumentAndKeepsTheKindFileResolution(t *testing.T) {
	t.Parallel()
	kind := member(t, heaterTwin(t), "kinds", 0)
	channels := member(t, kind, "channels")
	temperature := named(t, channels, "channel", "name", "temperature")
	if period, sigma, resolution := number(t, member(t, temperature, "cadence", "period_s")), number(t, member(t, temperature, "noise", "sigma")),
		number(t, member(t, temperature, "resolution")); period != 10 || sigma != 0.05 || resolution != 0.5 {
		t.Errorf("temperature period %v s, sigma %v, resolution %v; want the probe's 10 s and 0.05 and the kind file's 0.5", period, sigma, resolution)
	}
	if got := member(t, temperature, "tag"); got != "TT" {
		t.Errorf("temperature tag = %v, want the probe's TT", got)
	}
	if got := member(t, temperature, "observation_bias", "state"); got != "temperature__bias" {
		t.Errorf("temperature bias state = %v, want temperature__bias", got)
	}
	named(t, member(t, kind, "state"), "state", "name", "temperature__bias")
	named(t, member(t, kind, "faults"), "fault", "id", "heater__temperature__offset")
	named(t, member(t, kind, "faults"), "fault", "id", "heater__temperature__drift")
}

func TestATwinScalesTheNoiseOfASpanInstrumentByTheRangeOfTheMetric(t *testing.T) {
	t.Parallel()
	power := named(t, member(t, heaterTwin(t), "kinds", 0, "channels"), "channel", "name", "power")
	if got := number(t, member(t, power, "noise", "sigma")); got != 1 {
		t.Fatalf("power noise sigma = %v, want 0.01 of the meter times the 100 kW span", got)
	}
}

func TestAMetricWithoutAnInstrumentIsAPlainChannelWithADropoutFault(t *testing.T) {
	t.Parallel()
	kind := member(t, heaterTwin(t), "kinds", 0)
	setpoint := named(t, member(t, kind, "channels"), "channel", "name", "setpoint")
	if got := member(t, setpoint, "noise", "model"); got != "none" {
		t.Errorf("setpoint noise model = %v, want none", got)
	}
	if got := number(t, member(t, setpoint, "cadence", "period_s")); got != plainPeriodS {
		t.Errorf("setpoint cadence = %v s, want %d", got, plainPeriodS)
	}
	fault := named(t, member(t, kind, "faults"), "fault", "id", "heater__setpoint__dropout")
	if got := member(t, fault, "observation", "form"); got != "dropout" {
		t.Errorf("setpoint fault form = %v, want dropout", got)
	}
}

func TestADirectResponseAssignsTheArgumentAndALeaseFallsBackToTheLowestValue(t *testing.T) {
	t.Parallel()
	effectors := member(t, heaterTwin(t), "kinds", 0, "effectors")
	set := named(t, effectors, "effector", "name", "set_setpoint")
	delta := member(t, set, "effect", "state_deltas", 0)
	if member(t, delta, "state") != "setpoint_value" || member(t, delta, "assign_from_arg") != "setpoint" {
		t.Errorf("set_setpoint delta = %v, want an assignment of setpoint to setpoint_value", delta)
	}
	lease := member(t, set, "lease")
	safe := member(t, lease, "safe_deltas", 0)
	if member(t, lease, "arg") != "lease_s" || member(t, lease, "unit") != "s" || member(t, lease, "fallback") != "safe_off" ||
		member(t, safe, "state") != "setpoint_value" || number(t, member(t, safe, "delta")) != 10 {
		t.Errorf("set_setpoint lease = %v, want lease_s in s falling back to 10 on setpoint_value", lease)
	}
	if member(t, set, "risk_class") != "medium" {
		t.Errorf("set_setpoint risk class = %v, want the kind file's medium", member(t, set, "risk_class"))
	}
}

func TestANoneResponseKeepsTheCommandAndChangesNothing(t *testing.T) {
	t.Parallel()
	ping := named(t, member(t, heaterTwin(t), "kinds", 0, "effectors"), "effector", "name", "ping")
	delta := member(t, ping, "effect", "state_deltas", 0)
	if member(t, delta, "state") != "temperature_value" || number(t, member(t, delta, "delta")) != 0 || member(t, ping, "lease") != nil {
		t.Fatalf("ping effect = %v, want a zero delta on the first state and no lease", member(t, ping, "effect"))
	}
}

func TestASafeStopDrivesEveryActuatedMetricToTheLowestValueOfItsRange(t *testing.T) {
	t.Parallel()
	halt := named(t, member(t, heaterTwin(t), "kinds", 0, "effectors"), "effector", "name", "halt")
	deltas := member(t, halt, "effect", "state_deltas").([]any)
	if len(deltas) != 1 || member(t, deltas[0], "state") != "setpoint_value" || number(t, member(t, deltas[0], "delta")) != -80 {
		t.Fatalf("halt deltas = %v, want setpoint_value lowered by the 80 of its range", deltas)
	}
}

func TestAFirstOrderResponseMakesTheReadingFollowItsRestValueAndTheDrivingMetric(t *testing.T) {
	t.Parallel()
	dynamics := member(t, heaterTwin(t), "kinds", 0, "dynamics", 0)
	f1 := member(t, dynamics, "f1")
	inputs := member(t, f1, "inputs").([]any)
	if member(t, dynamics, "target") != "temperature_value" || member(t, f1, "form") != "first_order_lag" || number(t, member(t, f1, "time_constant_s")) != 120 {
		t.Fatalf("dynamics = %v, want a 120 s first-order lag of temperature_value", dynamics)
	}
	if len(inputs) != 2 || member(t, inputs[0], "state") != "temperature_rest" || member(t, inputs[1], "state") != "setpoint_value" || member(t, inputs[1], "link") != nil {
		t.Fatalf("inputs = %v, want the rest state and the same-kind driver without a link", inputs)
	}
	if got := number(t, member(t, f1, "clamp", "max")); got != 200 {
		t.Errorf("clamp max = %v, want the 200 of the eu_range", got)
	}
}

func TestACommandOfOneKindDrivesAReadingOfAnotherOverALink(t *testing.T) {
	t.Parallel()
	doc := fanZoneTwin(t)
	linkType := member(t, doc, "link_types", 0)
	link := member(t, doc, "links", 0)
	if member(t, linkType, "name") != "fan_drives_zone" || member(t, linkType, "from") != "fan" || member(t, linkType, "to") != "zone" ||
		member(t, link, "from") != "fan-1" || member(t, link, "to") != "zone-1" || member(t, link, "type") != "fan_drives_zone" {
		t.Fatalf("link type %v and link %v, want fan-1 joined to zone-1 by fan_drives_zone", linkType, link)
	}
	input := member(t, doc, "kinds", 0, "dynamics", 0, "f1", "inputs", 1)
	if member(t, input, "state") != "duty_value" || number(t, member(t, input, "coef")) != -0.02 || member(t, input, "link", "type") != "fan_drives_zone" || member(t, input, "link", "aggregate") != "sum" {
		t.Fatalf("zone input = %v, want duty_value at -0.02 summed over fan_drives_zone", input)
	}
}

func TestATwinDeclaresItsSiteBoardInstancesAndTheKindsItImplements(t *testing.T) {
	t.Parallel()
	composed, doc := twinOK(t, kindDocuments(t, heaterKindJSON), heaterChoicesJSON)
	if !slices.Equal(composed.Implements, []string{"heater"}) || member(t, doc, "kinds", 0, "implements") != "heater" {
		t.Errorf("implements = %v and %v, want heater", composed.Implements, member(t, doc, "kinds", 0, "implements"))
	}
	if member(t, doc, "instances", 0, "names", 0) != "heater-01" || member(t, doc, "instances", 0, "area") != "bench" ||
		member(t, doc, "boards", 0, "name") != "board-1" || member(t, doc, "boards", 0, "members", 0) != "heater-01" || member(t, doc, "site", "name") != "lab" {
		t.Errorf("structure = instances %v, boards %v, site %v", member(t, doc, "instances"), member(t, doc, "boards"), member(t, doc, "site"))
	}
	if member(t, doc, "axes", "correlation", 0) != "coupled" || member(t, doc, "axes", "actuation") != "fast_physical" {
		t.Errorf("axes = %v, want a coupled plant with fast physical actuation", member(t, doc, "axes"))
	}
}

func TestAScaffoldedTwinIsTheSameBytesEveryTime(t *testing.T) {
	t.Parallel()
	first, _ := twinOK(t, kindDocuments(t, zoneKindJSON, fanKindJSON), fanZoneChoicesJSON)
	second, _ := twinOK(t, kindDocuments(t, zoneKindJSON, fanKindJSON), fanZoneChoicesJSON)
	if !bytes.Equal(first.Bytes, second.Bytes) || first.Summary.Digest != second.Summary.Digest {
		t.Fatal("two scaffolds of one input differ")
	}
}
