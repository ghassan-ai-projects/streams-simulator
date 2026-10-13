package domain

import "testing"

func TestComposeWritesNoUnitOnABooleanChannel(t *testing.T) {
	t.Parallel()
	variant := instrumentVariant(t, "acme/probe-flag", "degC", func(doc map[string]any) {
		doc["channel"] = map[string]any{"value_type": "boolean", "resolution": 1, "fidelity": "F1", "absence": "normal", "cadence": map[string]any{"mode": "periodic", "period_s": 5},
			"noise": map[string]any{"model": "none", "sigma": 0}}
		delete(doc, "faults")
	})
	doc := composeOK(t, libraryWith(t, variant), withEquipment(t, replacing("temp", "acme/probe-flag@1")))
	channel := channelOf(t, doc, "temp")
	if _, present := channel["unit"]; present || channel["value_type"] != "boolean" {
		t.Fatalf("channel = %v, want a boolean channel without a unit", channel)
	}
}
