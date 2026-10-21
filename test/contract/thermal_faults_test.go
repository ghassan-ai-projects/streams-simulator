package contract

import (
	"math"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/world"
)

const arduinoAmbientC, arduinoHeatLoadC, arduinoFaultOnsetS = 22.0, 12.0, 60.0

func faultRiseCurve(seconds, riseC float64) float64 {
	if seconds <= arduinoFaultOnsetS {
		return arduinoRestC
	}
	return arduinoRestC + riseC*(1-math.Exp(-(seconds-arduinoFaultOnsetS)/arduinoTauS))
}

func faultedArduinoWorld(t *testing.T, entity, fault string) *world.World {
	t.Helper()
	w := arduinoWorld(t, 11, world.Options{Noiseless: true, EmitDisabled: true})
	if _, err := w.InjectFault(entity, fault, arduinoAt(arduinoFaultOnsetS), nil); err != nil {
		t.Fatal(err)
	}
	return w
}

func TestTheZoneFaultsRaiseTheZoneAlongTheClosedFormOfItsFirstOrderLag(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		fault                string
		riseC, ambient, heat float64
	}{
		{"ambient_warm_spell", 4, arduinoAmbientC + 4, arduinoHeatLoadC},
		{"zone_heat_load", 8, arduinoAmbientC, arduinoHeatLoadC + 8},
	} {
		t.Run(tc.fault, func(t *testing.T) {
			t.Parallel()
			w := faultedArduinoWorld(t, arduinoZone, tc.fault)
			for _, seconds := range []float64{30, 60, 120, 240, 600, 1200, 3600} {
				want, got := faultRiseCurve(seconds, tc.riseC), arduinoState(w, arduinoZone, "zone_temperature", seconds)
				if math.Abs(got-want) > 0.1 {
					t.Errorf("zone temperature at %v s = %.4f, closed form %.4f; the held-input step allows 0.1", seconds, got, want)
				}
			}
			ambient, heat := arduinoState(w, arduinoZone, "ambient_temp", 3600), arduinoState(w, arduinoZone, "heat_load", 3600)
			if math.Abs(ambient-tc.ambient) > 1e-9 || math.Abs(heat-tc.heat) > 1e-9 {
				t.Errorf("at 3600 s ambient = %v and heat load = %v, want %v and %v", ambient, heat, tc.ambient, tc.heat)
			}
		})
	}
}

func TestTheZoneHeatLoadIsAFaultAndTheWarmSpellIsNot(t *testing.T) {
	t.Parallel()
	spec := arduinoSpec(t)
	for fault, negative := range map[string]bool{"ambient_warm_spell": true, "zone_heat_load": false} {
		declared := spec.Fault(fault)
		if declared == nil || declared.IsNegativeClass != negative {
			t.Errorf("fault %s = %+v, want it declared with is_negative_class %v", fault, declared, negative)
		}
	}
}
