package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func observationFaultDoc(observation map[string]any, detectorChannel string) map[string]any {
	return map[string]any{"id": "probe_fault", "onset": map[string]any{"shape": "step", "magnitude": 1.0}, "observation": observation,
		"observability": map[string]any{"detector": map[string]any{"form": "single_channel_snr", "channel": detectorChannel}, "first_observable_snr": 3.0, "unavoidable_snr": 10.0}}
}

func TestAnObservationFaultNamesANumericChannelOfItsKindAndIsDetectedThere(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		fault   map[string]any
		pointer string
	}{
		{"a stuck flow meter", observationFaultDoc(map[string]any{"channel": "unit.flow", "form": "stuck", "value": 0.0, "quality": "out_of_range"}, "unit.flow"), ""},
		{"a channel of another kind", observationFaultDoc(map[string]any{"channel": "header.level", "form": "dropout"}, "header.level"), "/kinds/0/faults/0/observation/channel"},
		{"detected on another channel", observationFaultDoc(map[string]any{"channel": "unit.flow", "form": "dropout"}, "unit.temp"), "/kinds/0/faults/0/observability/detector"},
		{"a scale fault with no factor", observationFaultDoc(map[string]any{"channel": "unit.flow", "form": "scale"}, "unit.flow"), "/kinds/0/faults/0/observation/factor"},
		{"a noise fault on a noiseless channel", observationFaultDoc(map[string]any{"channel": "unit.flow", "form": "noise", "factor": 3.0}, "unit.flow"), "/kinds/0/faults/0/observation/form"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			violationIn(t, testsupport.CoupledDomain(), func(d map[string]any) {
				if tc.name == "a noise fault on a noiseless channel" {
					put(d, map[string]any{"model": "none", "sigma": 0.0}, "kinds", 0, "channels", 0, "noise")
				}
				put(d, tc.fault, "kinds", 0, "faults", 0)
			}, tc.pointer)
		})
	}
}
