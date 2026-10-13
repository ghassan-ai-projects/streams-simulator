package domain

import "testing"

func faultsInstrument(t *testing.T, faults []any) File {
	t.Helper()
	return instrumentVariant(t, "acme/faulty", "degC", func(doc map[string]any) { doc["faults"] = faults })
}

func faultOf(t *testing.T, doc map[string]any, id string) map[string]any {
	t.Helper()
	return named(t, member(t, doc, "kinds", 0, "faults"), "fault", "id", "kiln__temp__"+id)
}

func composeFaulty(t *testing.T, faults ...any) map[string]any {
	t.Helper()
	return composeOK(t, libraryWith(t, faultsInstrument(t, faults)), withEquipment(t, replacing("temp", "acme/faulty@1")))
}

func TestComposeExpandsAnOffsetThroughAHiddenBiasState(t *testing.T) {
	t.Parallel()
	doc := composeFaulty(t, map[string]any{"id": "offset", "form": "offset", "magnitude": 1.5, "is_negative_class": true})
	fault := faultOf(t, doc, "offset")
	near(t, "delta", member(t, fault, "affects", 0, "delta"), 1.5)
	if member(t, fault, "affects", 0, "state") != "temp__bias" || member(t, fault, "onset", "shape") != "step" || fault["is_negative_class"] != true {
		t.Errorf("fault = %v, want a step on temp__bias flagged negative class", fault)
	}
	detector := member(t, fault, "observability", "detector")
	if member(t, detector, "form") != "single_channel_snr" || member(t, detector, "channel") != "kiln.temp" {
		t.Errorf("detector = %v, want single_channel_snr on kiln.temp", detector)
	}
	near(t, "first_observable_snr", member(t, fault, "observability", "first_observable_snr"), 3)
	near(t, "unavoidable_snr", member(t, fault, "observability", "unavoidable_snr"), 10)
}

func TestComposeExpandsADriftAsARampToTheFullDelta(t *testing.T) {
	t.Parallel()
	doc := composeFaulty(t, map[string]any{"id": "drift", "form": "drift", "magnitude": 0.5, "duration_s": 7200})
	fault := faultOf(t, doc, "drift")
	near(t, "delta", member(t, fault, "affects", 0, "delta"), 1)
	near(t, "rate_per_hour", member(t, fault, "onset", "rate_per_hour"), 0.5)
	near(t, "onset magnitude", member(t, fault, "onset", "magnitude"), 1)
	if member(t, fault, "onset", "shape") != "ramp" {
		t.Errorf("onset = %v, want a ramp", fault["onset"])
	}
}

func TestComposeExpandsObservationFaultsWithoutABiasState(t *testing.T) {
	t.Parallel()
	doc := composeFaulty(t,
		map[string]any{"id": "stuck_low", "form": "stuck", "stuck_at": "low", "quality": "suspect"},
		map[string]any{"id": "stuck_high", "form": "stuck", "stuck_at": "high"},
		map[string]any{"id": "freeze", "form": "freeze", "quality": "stale"},
		map[string]any{"id": "dropout", "form": "dropout", "quality": "disconnected"},
		map[string]any{"id": "scale", "form": "scale", "factor": 1.2},
		map[string]any{"id": "noise", "form": "noise", "factor": 4})
	cases := []struct {
		id, form string
		value    float64
		factor   float64
		quality  string
	}{
		{"stuck_low", "stuck", 0, 0, "suspect"}, {"stuck_high", "stuck", 300, 0, ""}, {"freeze", "freeze", 0, 0, "stale"},
		{"dropout", "dropout", 0, 0, "disconnected"}, {"scale", "scale", 0, 1.2, ""}, {"noise", "noise", 0, 4, ""},
	}
	for _, tc := range cases {
		observation := member(t, faultOf(t, doc, tc.id), "observation")
		if member(t, observation, "channel") != "kiln.temp" || member(t, observation, "form") != tc.form {
			t.Errorf("%s: observation = %v", tc.id, observation)
		}
		checkObservation(t, observation.(map[string]any), tc.value, tc.factor, tc.quality)
		if member(t, faultOf(t, doc, tc.id), "onset", "shape") != "step" {
			t.Errorf("%s: onset is not a step", tc.id)
		}
	}
	for _, item := range member(t, doc, "kinds", 0, "state").([]any) {
		if item.(map[string]any)["name"] == "temp__bias" {
			t.Error("observation faults created a bias state")
		}
	}
	if _, present := channelOf(t, doc, "temp")["observation_bias"]; present {
		t.Error("the channel carries an observation bias although no fault acts through one")
	}
}

func checkObservation(t *testing.T, observation map[string]any, value, factor float64, quality string) {
	t.Helper()
	if form := observation["form"]; form == "stuck" {
		near(t, "stuck value", observation["value"], value)
	} else if _, present := observation["value"]; present {
		t.Errorf("%v: observation carries a value", form)
	}
	if factor != 0 {
		near(t, "factor", observation["factor"], factor)
	} else if _, present := observation["factor"]; present {
		t.Errorf("observation carries a factor: %v", observation)
	}
	if got, _ := observation["quality"].(string); got != quality {
		t.Errorf("quality = %q, want %q", got, quality)
	}
}

func TestComposeSticksAtTheEndsOfTheConvertedRange(t *testing.T) {
	t.Parallel()
	variant := instrumentVariant(t, "acme/probe-k", "K", func(doc map[string]any) {
		doc["faults"] = []any{map[string]any{"id": "low", "form": "stuck", "stuck_at": "low"}, map[string]any{"id": "high", "form": "stuck", "stuck_at": "high"}}
	})
	doc := composeOK(t, libraryWith(t, variant), withEquipment(t, replacing("temp", "acme/probe-k@1")))
	near(t, "low", member(t, faultOf(t, doc, "low"), "observation", "value"), 273.15)
	near(t, "high", member(t, faultOf(t, doc, "high"), "observation", "value"), 400)
}

func TestComposeMixesBiasAndObservationFaultsOnOnePoint(t *testing.T) {
	t.Parallel()
	doc := composeFaulty(t, map[string]any{"id": "offset", "form": "offset", "magnitude": 1}, map[string]any{"id": "freeze", "form": "freeze"})
	if member(t, channelOf(t, doc, "temp"), "observation_bias", "state") != "temp__bias" {
		t.Error("an offset fault beside a freeze fault lost its bias state")
	}
	faultOf(t, doc, "freeze")
	faultOf(t, doc, "offset")
}

func TestComposeRefusesABiasStateThatCollidesWithAComponentState(t *testing.T) {
	t.Parallel()
	clash := mutate(t, kilnJSON, func(doc map[string]any) {
		states := member(t, doc, "kind", "state").([]any)
		doc["kind"].(map[string]any)["state"] = append(states, map[string]any{"name": "temp__bias", "unit": "degC", "initial": 0})
	})
	files := baseFiles()
	files.equipment = []File{file("equipment/kiln.component.json", clash)}
	_, err := compose(t, loadLibrary(t, files.directory("lib")), withEquipment(t, kilnEntry(nil)))
	wantIssue(t, err, "composition-v0.1", "/equipment/0/use", `already declares a state "temp__bias"`)
}
