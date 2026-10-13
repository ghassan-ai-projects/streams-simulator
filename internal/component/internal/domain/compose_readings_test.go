package domain

import (
	"math"
	"testing"
)

func instrumentVariant(t *testing.T, id, unit string, change func(doc map[string]any)) File {
	t.Helper()
	raw := mutate(t, probeJSON, func(doc map[string]any) {
		doc["id"] = id
		doc["unit"] = unit
		change(doc)
	})
	return file("instruments/"+id[len("acme/"):]+".instrument.json", raw)
}

func replacing(point, ref string) map[string]any {
	return kilnEntry(map[string]any{"instruments": map[string]any{"replace": map[string]any{point: ref}}})
}

func channelOf(t *testing.T, doc map[string]any, name string) map[string]any {
	t.Helper()
	return named(t, member(t, doc, "kinds", 0, "channels"), "channel", "name", "kiln."+name)
}

func near(t *testing.T, label string, got any, want float64) {
	t.Helper()
	if value := number(t, got); math.Abs(value-want) > 1e-9 {
		t.Errorf("%s = %v, want %v", label, value, want)
	}
}

func TestComposeConvertsTheRangeAndTheReadingToTheInstrumentUnit(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name         string
		unit         string
		gain, offset float64
		min, max     float64
	}{
		{"kelvin", "K", 1, 273.15, 273.15, 573.15},
		{"fahrenheit", "degF", 1.8, 32, 32, 572},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			variant := instrumentVariant(t, "acme/probe-"+tc.name, tc.unit, func(doc map[string]any) { delete(doc, "range") })
			doc := composeOK(t, libraryWith(t, variant), withEquipment(t, replacing("temp", "acme/probe-"+tc.name+"@1")))
			channel := channelOf(t, doc, "temp")
			if channel["unit"] != tc.unit {
				t.Errorf("unit = %v, want %s", channel["unit"], tc.unit)
			}
			near(t, "observation_gain", channel["observation_gain"], tc.gain)
			near(t, "observation_offset", channel["observation_offset"], tc.offset)
			near(t, "range min", member(t, channel, "range", "min"), tc.min)
			near(t, "range max", member(t, channel, "range", "max"), tc.max)
		})
	}
}

func TestComposeWritesNoConversionForEqualUnits(t *testing.T) {
	t.Parallel()
	channel := channelOf(t, composeOK(t, baseLibrary(t), withEquipment(t, kilnEntry(nil))), "temp")
	for _, absent := range []string{"observation_gain", "observation_offset"} {
		if _, present := channel[absent]; present {
			t.Errorf("channel carries %s although the units are equal", absent)
		}
	}
}

func TestComposeIntersectsTheConvertedRangeWithTheInstrumentRange(t *testing.T) {
	t.Parallel()
	variant := instrumentVariant(t, "acme/probe-k", "K", func(doc map[string]any) { doc["range"] = map[string]any{"min": 280, "max": 500} })
	doc := composeOK(t, libraryWith(t, variant), withEquipment(t, replacing("temp", "acme/probe-k@1")))
	channel := channelOf(t, doc, "temp")
	near(t, "range min", member(t, channel, "range", "min"), 280)
	near(t, "range max", member(t, channel, "range", "max"), 500)
}

func TestComposeRefusesAnEmptyIntersection(t *testing.T) {
	t.Parallel()
	variant := instrumentVariant(t, "acme/probe-k", "K", func(doc map[string]any) { doc["range"] = map[string]any{"min": 700, "max": 800} })
	_, err := compose(t, libraryWith(t, variant), withEquipment(t, replacing("temp", "acme/probe-k@1")))
	wantIssue(t, err, "composition-v0.1", "/equipment/0/instruments/replace/temp", "is empty")
}

func TestComposeScalesNoiseAndFaultMagnitudesByTheSpan(t *testing.T) {
	t.Parallel()
	doc := composeOK(t, baseLibrary(t), withEquipment(t, kilnEntry(map[string]any{"preset": "big"})))
	power := channelOf(t, doc, "power")
	near(t, "power sigma", member(t, power, "noise", "sigma"), 0.01*1000)
	fault := named(t, member(t, doc, "kinds", 0, "faults"), "fault", "id", "kiln__power__offset")
	near(t, "power offset delta", member(t, fault, "affects", 0, "delta"), 0.05*1000)
	temp := channelOf(t, doc, "temp")
	near(t, "absolute sigma", member(t, temp, "noise", "sigma"), 0.05)
}

func TestComposeScalesBySpanOfTheConvertedPointRangeNotOfTheNarrowerInstrumentRange(t *testing.T) {
	t.Parallel()
	narrow := file("instruments/meter-w.instrument.json", mutate(t, meterJSON, func(doc map[string]any) {
		doc["id"], doc["unit"], doc["range"] = "acme/meter-w", "W", map[string]any{"min": 0, "max": 50000}
	}))
	doc := composeOK(t, libraryWith(t, narrow), withEquipment(t, replacing("power", "acme/meter-w@1")))
	power := channelOf(t, doc, "power")
	near(t, "range max", member(t, power, "range", "max"), 50000)
	near(t, "sigma", member(t, power, "noise", "sigma"), 0.01*100000)
	fault := named(t, member(t, doc, "kinds", 0, "faults"), "fault", "id", "kiln__power__offset")
	near(t, "offset delta", member(t, fault, "affects", 0, "delta"), 0.05*100000)
}

func TestComposeKeepsTheInstrumentChannelMembers(t *testing.T) {
	t.Parallel()
	channel := channelOf(t, composeOK(t, baseLibrary(t), withEquipment(t, kilnEntry(nil))), "temp")
	near(t, "resolution", channel["resolution"], 0.1)
	if channel["fidelity"] != "F1" || channel["absence"] != "signal" || member(t, channel, "cadence", "mode") != "periodic" {
		t.Errorf("channel = %v, want the instrument's fidelity, absence and cadence", channel)
	}
}
