package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func TestABiasWithoutANoiseScaleKeepsTheChannelNoise(t *testing.T) {
	t.Parallel()
	quiet := 0.25
	for _, tc := range []struct {
		name  string
		scale *float64
		want  float64
	}{{"absent", nil, 0.1}, {"explicit", &quiet, 0.025}} {
		w := constWorld(t, func(s *model.DomainSpec) {
			s.Kinds[0].State = append(s.Kinds[0].State, model.State{Name: "fouling", Initial: 1})
			s.Kinds[0].Channels[0].ObservationBias = &model.Bias{State: "fouling", Coef: 0, NoiseScaleAtFull: tc.scale}
		}, model.DefaultStartTimeNS)
		ent := w.entities["e-1"]
		if got := w.observationSigma(ent, &ent.kind.Channels[0], model.DefaultStartTimeNS); got != tc.want {
			t.Fatalf("%s noise scale: sigma at full bias = %v, want %v", tc.name, got, tc.want)
		}
	}
}
